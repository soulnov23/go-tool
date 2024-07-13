package file

import (
	"bufio"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/soulnov23/go-tool/pkg/utils"
)

func ReadAll(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return utils.BytesToString(data), nil
}

func ReadLines(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// scanner每次Scan返回的buf，如果使用BytesToString引用buf的地址，当Scan超过startBufSize后，前面引用buf地址的数据就被覆盖了
func Deduplicate(filepath string, sorted bool) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	var uniq sync.Map
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if _, loaded := uniq.LoadOrStore(scanner.Text(), true); !loaded {
			lines = append(lines, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if sorted {
		slices.Sort(lines)
	}
	tmpName := filepath + ".tmp"
	tmp, err := os.Create(tmpName)
	if err != nil {
		return err
	}
	defer tmp.Close()
	defer os.Remove(tmpName)
	for _, line := range lines {
		if _, err := tmp.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return os.Rename(tmpName, filepath)
}
