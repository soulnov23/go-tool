package file

import (
	"bufio"
	"os"
	"slices"
	"sync"
)

func ReadAll(filepath string) ([]string, error) {
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
