package writer

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lestrrat-go/strftime"
)

const (
	CompressSuffix = ".gz"
)

// ensure we always implement io.WriteCloser.
var _ io.WriteCloser = (*rollWriter)(nil)

// rollWriter is a file log writer which support rolling by size or datetime.
// It implements io.WriteCloser.
type rollWriter struct {
	filePath string
	opts     *Options

	pattern  *strftime.Strftime
	currDir  string
	currPath string
	currSize int64
	currFile atomic.Value
	currTime string
	currNum  int
	openTime int64

	mu         sync.Mutex
	notifyOnce sync.Once
	notifyCh   chan bool
	closeOnce  sync.Once
	closeCh    chan *os.File
}

// New creates a new rollWriter.
func New(filePath string, opt ...Option) (*rollWriter, error) {
	opts := &Options{
		MaxSize:           0,     // default no rolling by file size
		MaxAge:            0,     // default no scavenging on expired logs
		MaxBackups:        0,     // default no scavenging on redundant logs
		Compress:          false, // default no compressing
		CloseFileDelay:    20,    // default 20ms delay before closing files
		CloseFileChanSize: 100,   // default 100 buffer size for close channel
	}

	// opt has the highest priority and should overwrite the original one.
	for _, o := range opt {
		o(opts)
	}

	if filePath == "" {
		return nil, errors.New("file path is empty")
	}

	pattern, err := strftime.New(filePath + opts.TimeFormat)
	if err != nil {
		return nil, errors.New("get file pattern: " + err.Error())
	}

	w := &rollWriter{
		filePath: filePath,
		opts:     opts,
		pattern:  pattern,
		currDir:  filepath.Dir(filePath),
	}

	if err := os.MkdirAll(w.currDir, 0o755); err != nil {
		return nil, err
	}

	return w, nil
}

// Write writes logs. It implements io.Writer.
func (w *rollWriter) Write(v []byte) (n int, err error) {
	now := time.Now()
	// reopen file every 10 seconds.
	if w.getCurrFile() == nil || now.Unix()-atomic.LoadInt64(&w.openTime) > 10 {
		if err := w.reopenWithCheck(now, false, int64(len(v))); err != nil {
			return 0, err
		}
	}

	// return when failed to open the file.
	if w.getCurrFile() == nil {
		return 0, errors.New("get current file: open file fail")
	}

	// check if we need to roll the log file due to time change or size limit
	needRolling := w.currTime != w.pattern.FormatString(now)
	needRolling = needRolling || (w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize)+int64(len(v)) >= w.opts.MaxSize)

	if needRolling {
		if err := w.reopenWithCheck(now, true, int64(len(v))); err != nil {
			return 0, err
		}
	}

	// write logs to file.
	n, err = w.getCurrFile().Write(v)
	atomic.AddInt64(&w.currSize, int64(n))

	return n, err
}

// reopenWithCheck handles reopening log file with proper locking and checks
func (w *rollWriter) reopenWithCheck(now time.Time, isRolling bool, bytes int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// recheck condition after acquiring the lock
	if !isRolling && (w.getCurrFile() == nil || now.Unix()-atomic.LoadInt64(&w.openTime) > 10) {
		formatString := w.pattern.FormatString(now)
		lastFileNum := w.getLastFileNum(filepath.Base(formatString))
		if lastFileNum == -1 {
			lastFileNum = 0
		}
		currPath := w.fileNameWithTimeAndNum(formatString, lastFileNum)
		if w.currPath != currPath {
			w.notify()
		}
		if w.doReopenFile(currPath) == nil {
			w.currPath = currPath
			w.currNum = lastFileNum
			w.currTime = formatString
			atomic.StoreInt64(&w.openTime, now.Unix())
		}
		return nil
	}

	// Check if time changed
	if isRolling && (w.currTime != w.pattern.FormatString(now)) {
		formatString := w.pattern.FormatString(now)
		lastFileNum := w.getLastFileNum(filepath.Base(formatString))
		currPath := w.fileNameWithTimeAndNum(formatString, lastFileNum+1)
		if w.currPath != currPath {
			w.notify()
		}
		if w.doReopenFile(currPath) == nil {
			w.currPath = currPath
			w.currNum = lastFileNum
			w.currTime = formatString
			atomic.StoreInt64(&w.openTime, now.Unix())
		}
		return nil
	}

	// Check for size-based rolling
	if isRolling && w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize)+bytes >= w.opts.MaxSize {
		formatString := w.pattern.FormatString(now)
		lastFileNum := w.getLastFileNum(filepath.Base(formatString))
		currPath := w.fileNameWithTimeAndNum(formatString, lastFileNum+1)
		if w.currPath != currPath {
			w.notify()
		}
		if w.doReopenFile(currPath) == nil {
			w.currPath = currPath
			w.currNum = lastFileNum
			w.currTime = formatString
			atomic.StoreInt64(&w.openTime, now.Unix())
		}
		return nil
	}

	return nil
}

// Close closes the current log file. It implements io.Closer.
func (w *rollWriter) Close() error {
	if w.getCurrFile() == nil {
		return nil
	}
	err := w.getCurrFile().Close()
	w.setCurrFile(nil)

	if w.notifyCh != nil {
		close(w.notifyCh)
		w.notifyCh = nil
	}

	if w.closeCh != nil {
		close(w.closeCh)
		w.closeCh = nil
	}

	return err
}

// getCurrFile returns the current log file.
func (w *rollWriter) getCurrFile() *os.File {
	if file, ok := w.currFile.Load().(*os.File); ok {
		return file
	}
	return nil
}

// setCurrFile sets the current log file.
func (w *rollWriter) setCurrFile(file *os.File) {
	w.currFile.Store(file)
}

// doReopenFile reopen the file.
func (w *rollWriter) doReopenFile(path string) error {
	lastFile := w.getCurrFile()
	of, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
	if err == nil {
		w.setCurrFile(of)
		if lastFile != nil {
			// delay closing until not used.
			w.delayCloseFile(lastFile)
		}
		st, _ := os.Stat(path)
		if st != nil {
			atomic.StoreInt64(&w.currSize, st.Size())
		}
	}
	return err
}

// notify runs scavengers.
func (w *rollWriter) notify() {
	w.notifyOnce.Do(func() {
		w.notifyCh = make(chan bool, 1)
		go w.runCleanFiles()
	})
	select {
	case w.notifyCh <- true:
	default:
	}
}

// runCleanFiles cleans redundant or expired (compressed) logs in a new goroutine.
func (w *rollWriter) runCleanFiles() {
	for range w.notifyCh {
		if w.opts.MaxBackups == 0 && w.opts.MaxAge == 0 && !w.opts.Compress {
			continue
		}
		w.cleanFiles()
	}
}

// delayCloseFile delay closing file
func (w *rollWriter) delayCloseFile(file *os.File) {
	w.closeOnce.Do(func() {
		w.closeCh = make(chan *os.File, w.opts.CloseFileChanSize)
		go w.runCloseFiles()
	})
	select {
	case w.closeCh <- file:
		// File added to close channel
	default:
		// Channel is full, close immediately
		file.Close()
	}
}

// runCloseFiles delay closing file in a new goroutine.
func (w *rollWriter) runCloseFiles() {
	for f := range w.closeCh {
		if w.opts.CloseFileDelay > 0 {
			time.Sleep(time.Duration(w.opts.CloseFileDelay) * time.Millisecond)
		}
		f.Close()
	}
}

// cleanFiles cleans redundant or expired (compressed) logs.
func (w *rollWriter) cleanFiles() {
	// get the file list of current log.
	files, err := w.getOldLogFiles()
	if err != nil || len(files) == 0 {
		return
	}

	// find the oldest files to scavenge.
	var compress, remove []logInfo
	files = filterByMaxBackups(files, &remove, w.opts.MaxBackups)

	// find the expired files by last modified time.
	files = filterByMaxAge(files, &remove, w.opts.MaxAge)

	// find files to compress by file extension .gz.
	filterByCompressExt(files, &compress, w.opts.Compress)

	// delete expired or redundant files.
	w.removeFiles(remove)

	// compress log files.
	w.compressFiles(compress)
}

// getLastFileNum returns the log file list ordered by modified time.
func (w *rollWriter) getLastFileNum(fileName string) int {
	number := -1
	fileInfo, err := os.ReadDir(w.currDir)
	if err != nil {
		return number
	}

	pattern := "^" + regexp.QuoteMeta(fileName) + "\\.(\\d+)$"
	re := regexp.MustCompile(pattern)

	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(file.Name())
		if len(matches) == 2 {
			// Extract the number from the matched group
			if extNum, err := strconv.Atoi(matches[1]); err == nil && extNum > number {
				number = extNum
			}
		}
	}
	return number
}

// fileNameWithTimeAndNum generates a filename with the time and sequence number
func (w *rollWriter) fileNameWithTimeAndNum(formatTimeStr string, fileNum int) string {
	return fmt.Sprintf("%s.%d", formatTimeStr, fileNum)
}

// getOldLogFiles returns the log file list ordered by modified time.
func (w *rollWriter) getOldLogFiles() ([]logInfo, error) {
	files, err := os.ReadDir(w.currDir)
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory %s: %w", w.currDir, err)
	}

	// Pre-allocate reasonable size to avoid reallocations
	logFiles := make([]logInfo, 0, len(files)/2)
	filename := filepath.Base(w.filePath)
	currPathBase := filepath.Base(w.currPath)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fname := f.Name()

		// Quick check to avoid expensive operations
		if !strings.HasPrefix(fname, filename) {
			continue
		}

		// Skip current log file
		if fname == currPathBase {
			continue
		}

		if st, err := os.Stat(filepath.Join(w.currDir, fname)); err == nil {
			logFiles = append(logFiles, logInfo{st.ModTime(), f})
		}
	}

	sort.Sort(byFormatTime(logFiles))
	return logFiles, nil
}

// matchLogFile checks whether current log file matches all relative log files, if matched, returns
// the modified time.
func (w *rollWriter) matchLogFile(filename, filePrefix string) (time.Time, error) {
	// exclude current log file.
	// a.log
	// a.log.20200712
	if filepath.Base(w.currPath) == filename {
		return time.Time{}, errors.New("ignore current logfile")
	}

	// match all log files with current log file.
	// a.log -> a.log.20200712-1232/a.log.20200712-1232.gz
	// a.log.20200712 -> a.log.20200712.20200712-1232/a.log.20200712.20200712-1232.gz
	if !strings.HasPrefix(filename, filePrefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}

	if st, _ := os.Stat(filepath.Join(w.currDir, filename)); st != nil {
		return st.ModTime(), nil
	}
	return time.Time{}, errors.New("file stat fail")
}

// removeFiles deletes expired or redundant log files.
func (w *rollWriter) removeFiles(remove []logInfo) {
	// clean expired or redundant files.
	for _, f := range remove {
		os.Remove(filepath.Join(w.currDir, f.Name()))
	}
}

// compressFiles compresses demanded log files.
func (w *rollWriter) compressFiles(compress []logInfo) {
	// compress log files.
	for _, f := range compress {
		fn := filepath.Join(w.currDir, f.Name())
		compressFile(fn, fn+CompressSuffix)
	}
}

// filterByMaxBackups filters redundant files that exceeded the limit.
func filterByMaxBackups(files []logInfo, remove *[]logInfo, maxBackups int) []logInfo {
	if maxBackups == 0 || len(files) < maxBackups {
		return files
	}
	var remaining []logInfo
	preserved := make(map[string]bool)
	for _, f := range files {
		fn := strings.TrimSuffix(f.Name(), CompressSuffix)
		preserved[fn] = true

		if len(preserved) > maxBackups {
			*remove = append(*remove, f)
		} else {
			remaining = append(remaining, f)
		}
	}
	return remaining
}

// filterByMaxAge filters expired files.
func filterByMaxAge(files []logInfo, remove *[]logInfo, maxAge int) []logInfo {
	if maxAge <= 0 {
		return files
	}
	var remaining []logInfo
	diff := time.Duration(int64(24*time.Hour) * int64(maxAge))
	cutoff := time.Now().Add(-1 * diff)
	for _, f := range files {
		if f.timestamp.Before(cutoff) {
			*remove = append(*remove, f)
		} else {
			remaining = append(remaining, f)
		}
	}
	return remaining
}

// filterByCompressExt filters all compressed files.
func filterByCompressExt(files []logInfo, compress *[]logInfo, needCompress bool) {
	if !needCompress {
		return
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), CompressSuffix) {
			*compress = append(*compress, f)
		}
	}
}

// compressFile compresses file src to dst, and removes src on success.
func compressFile(src, dst string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", src, err)
	}
	defer f.Close()

	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open compressed file %s: %w", dst, err)
	}
	defer func() {
		gzErr := gzf.Close()
		if err == nil && gzErr != nil {
			err = fmt.Errorf("failed to close compressed file %s: %w", dst, gzErr)
		}
	}()

	gz := gzip.NewWriter(gzf)
	defer func() {
		closeErr := gz.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close gzip writer: %w", closeErr)
			os.Remove(dst)
		} else if err != nil {
			os.Remove(dst)
			err = fmt.Errorf("failed to compress file: %w", err)
		} else {
			os.Remove(src)
		}
	}()

	if _, err := io.Copy(gz, f); err != nil {
		return fmt.Errorf("failed to write compressed data: %w", err)
	}
	return nil
}

// logInfo is an assistant struct which is used to return file name and last modified time.
type logInfo struct {
	timestamp time.Time
	fs.DirEntry
}

// byFormatTime sorts by time descending order.
type byFormatTime []logInfo

// Less checks whether the time of b[j] is early than the time of b[i].
func (b byFormatTime) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

// Swap swaps b[i] and b[j].
func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Len returns the length of list b.
func (b byFormatTime) Len() int {
	return len(b)
}
