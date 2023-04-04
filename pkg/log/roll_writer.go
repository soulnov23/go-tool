package log

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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

// NewRollWriter creates a new rollWriter.
func NewRollWriter(filePath string, opt ...Option) (*rollWriter, error) {
	opts := &Options{
		MaxSize:    0,     // default no rolling by file size
		MaxAge:     0,     // default no scavenging on expired logs
		MaxBackups: 0,     // default no scavenging on redundant logs
		Compress:   false, // default no compressing
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

	if err := os.MkdirAll(w.currDir, 0755); err != nil {
		return nil, err
	}

	return w, nil
}

// Write writes logs. It implements io.Writer.
func (w *rollWriter) Write(v []byte) (n int, err error) {
	now := time.Now()
	// reopen file every 10 seconds.
	if w.getCurrFile() == nil || now.Unix()-atomic.LoadInt64(&w.openTime) > 10 {
		w.mu.Lock()
		if w.getCurrFile() == nil || now.Unix()-atomic.LoadInt64(&w.openTime) > 10 {
			formatString := w.pattern.FormatString(now)
			lastFileNum := w.getLastFileNum(filepath.Base(formatString))
			if lastFileNum == -1 {
				lastFileNum = 0
			}
			currPath := w.fileNameWithTimeAndNum(w.pattern, lastFileNum)
			if w.currPath != currPath {
				w.notify()
			}
			if w.doReopenFile(currPath) == nil {
				w.currPath = currPath
				w.currNum = lastFileNum
				w.currTime = formatString
				atomic.StoreInt64(&w.openTime, now.Unix())
			}
		}
		w.mu.Unlock()
	}

	// return when failed to open the file.
	if w.getCurrFile() == nil {
		return 0, errors.New("get current file: open file fail")
	}

	// rolling on full
	if w.currTime != w.pattern.FormatString(now) || (w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize)+int64(len(v)) >= w.opts.MaxSize) {
		w.mu.Lock()
		if w.currTime != w.pattern.FormatString(now) || (w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize)+int64(len(v)) >= w.opts.MaxSize) {
			formatString := w.pattern.FormatString(now)
			lastFileNum := w.getLastFileNum(filepath.Base(formatString))
			currPath := w.fileNameWithTimeAndNum(w.pattern, lastFileNum+1)
			if w.currPath != currPath {
				w.notify()
			}
			if w.doReopenFile(currPath) == nil {
				w.currPath = currPath
				w.currNum = lastFileNum
				w.currTime = formatString
				atomic.StoreInt64(&w.openTime, now.Unix())
			}
		}
		w.mu.Unlock()
	}

	// write logs to file.
	n, err = w.getCurrFile().Write(v)
	atomic.AddInt64(&w.currSize, int64(n))

	return n, err
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
	of, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
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
		w.closeCh = make(chan *os.File, 100)
		go w.runCloseFiles()
	})
	w.closeCh <- file
}

// runCloseFiles delay closing file in a new goroutine.
func (w *rollWriter) runCloseFiles() {
	for f := range w.closeCh {
		// delay 20ms
		time.Sleep(20 * time.Millisecond)
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
	var files []string
	number := -1
	fileInfo, err := os.ReadDir(w.currDir)
	if err != nil {
		return number
	}
	for _, file := range fileInfo {
		if strings.HasPrefix(file.Name(), fileName) {
			files = append(files, file.Name())
		}
	}
	for _, file := range files {
		ext := filepath.Ext(file)
		if len(ext) > 0 {
			extNum, err := strconv.Atoi(strings.TrimLeft(ext, "."))
			if err == nil && extNum > number {
				number = extNum
			}
		}
	}
	return number
}

func (w *rollWriter) fileNameWithTimeAndNum(pattern *strftime.Strftime, fileNum int) string {
	return fmt.Sprintf("%s.%d", pattern.FormatString(time.Now()), fileNum)
}

// getOldLogFiles returns the log file list ordered by modified time.
func (w *rollWriter) getOldLogFiles() ([]logInfo, error) {
	files, err := os.ReadDir(w.currDir)
	if err != nil {
		return nil, errors.New("can't read log file " + w.currDir + " directory: " + err.Error())
	}
	logFiles := []logInfo{}
	filename := filepath.Base(w.filePath)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if modTime, err := w.matchLogFile(f.Name(), filename); err == nil {
			logFiles = append(logFiles, logInfo{modTime, f})
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
		runtime.Caller(0)
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
		return errors.New("failed to open file " + src + ": " + err.Error())
	}
	defer f.Close()

	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.New("failed to open compressed file " + dst + ": " + err.Error())
	}
	defer gzf.Close()

	gz := gzip.NewWriter(gzf)
	defer func() {
		gz.Close()
		if err != nil {
			os.Remove(dst)
			err = errors.New("failed to compress file: " + err.Error())
		} else {
			os.Remove(src)
		}
	}()

	if _, err := io.Copy(gz, f); err != nil {
		return err
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
