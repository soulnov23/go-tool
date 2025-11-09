package utils

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ColorRed    = "\033[1;31m"
	ColorGreen  = "\033[1;32m"
	ColorYellow = "\033[1;33m"
	ColorPurple = "\033[1;35m"
	ColorWhite  = "\033[1;37m"
	ColorReset  = "\033[m"
)

var callers sync.Map

func Debug(args ...any) {
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorGreen, time.Now().Format(time.DateTime), caller(2), fmt.Sprint(args...), ColorReset)
}

func Debugf(formatter string, args ...any) {
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorGreen, time.Now().Format(time.DateTime), caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Info(args ...any) {
	fmt.Printf("%s%s INFO %s %s%s\n", ColorWhite, time.Now().Format(time.DateTime), caller(2), fmt.Sprint(args...), ColorReset)
}

func Infof(formatter string, args ...any) {
	fmt.Printf("%s%s INFO %s %s%s\n", ColorWhite, time.Now().Format(time.DateTime), caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Warn(args ...any) {
	fmt.Printf("%s%s WARN %s %s%s\n", ColorYellow, time.Now().Format(time.DateTime), caller(2), fmt.Sprint(args...), ColorReset)
}

func Warnf(formatter string, args ...any) {
	fmt.Printf("%s%s WARN %s %s%s\n", ColorYellow, time.Now().Format(time.DateTime), caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Error(args ...any) {
	fmt.Printf("%s%s ERROR %s %s%s\n", ColorRed, time.Now().Format(time.DateTime), caller(2), fmt.Sprint(args...), ColorReset)
}

func Errorf(formatter string, args ...any) {
	fmt.Printf("%s%s ERROR %s %s%s\n", ColorRed, time.Now().Format(time.DateTime), caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Fatal(args ...any) {
	fmt.Printf("%s%s FATAL %s %s%s\n", ColorPurple, time.Now().Format(time.DateTime), caller(2), fmt.Sprint(args...), ColorReset)
	os.Exit(1)
}

func Fatalf(formatter string, args ...any) {
	fmt.Printf("%s%s FATAL %s %s%s\n", ColorPurple, time.Now().Format(time.DateTime), caller(2), fmt.Sprintf(formatter, args...), ColorReset)
	os.Exit(1)
}

// 返回调用者的"package/frame.File:frame.Line"
func caller(skip int) string {
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
