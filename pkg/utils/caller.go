package utils

import (
	"runtime"
	"strconv"
	"strings"
)

// 返回调用者的"package/file:line"
func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown_caller"
	}
	strLine := strconv.Itoa(line)
	fullCaller := strings.Join([]string{file, strLine}, ":")
	// 返回最后一个分隔符
	idx := strings.LastIndexByte(file, '/')
	if idx == -1 {
		return fullCaller
	}
	// 返回倒数第二个分隔符
	idx = strings.LastIndexByte(file[:idx], '/')
	if idx == -1 {
		return fullCaller
	}
	// 返回倒数第二个分隔符之后的所有内容
	return strings.Join([]string{file[idx+1:], strLine}, ":")
}
