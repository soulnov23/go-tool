package runtime

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	callers sync.Map
)

// 返回调用者的"package/frame.File:frame.Line"
func Caller(skip int) string {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(skip+1, rpc[:])
	if n < 1 {
		return "unknown"
	}
	var frame runtime.Frame
	if f, ok := callers.Load(rpc[0]); ok {
		frame = f.(runtime.Frame)
	} else {
		frame, _ = runtime.CallersFrames(rpc).Next()
		callers.Store(rpc[0], frame)
	}
	strLine := strconv.Itoa(frame.Line)
	fullCaller := strings.Join([]string{frame.File, strLine}, ":")
	// 返回最后一个分隔符
	idx := strings.LastIndexByte(frame.File, '/')
	if idx == -1 {
		return fullCaller
	}
	// 返回倒数第二个分隔符
	idx = strings.LastIndexByte(frame.File[:idx], '/')
	if idx == -1 {
		return fullCaller
	}
	// 返回倒数第二个分隔符之后的所有内容
	return strings.Join([]string{frame.File[idx+1:], strLine}, ":")
}
