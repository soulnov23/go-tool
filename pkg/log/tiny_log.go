package log

import (
	"fmt"
	"time"

	"github.com/soulnov23/go-tool/pkg/utils"
)

const (
	ColorRed    = "\033[1;31m"
	ColorGreen  = "\033[1;32m"
	ColorYellow = "\033[1;33m"
	ColorPurple = "\033[1;35m"
	ColorWhite  = "\033[1;37m"
	ColorReset  = "\033[m"

	DefaultTimeFormat = "2006-01-02 15:04:05.000"
)

func Debug(formatter string, args ...any) {
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorGreen, time.Now().Format(DefaultTimeFormat), utils.GetCaller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Info(formatter string, args ...any) {
	fmt.Printf("%s%s INFO %s %s%s\n", ColorWhite, time.Now().Format(DefaultTimeFormat), utils.GetCaller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Warn(formatter string, args ...any) {
	fmt.Printf("%s%s WARN %s %s%s\n", ColorYellow, time.Now().Format(DefaultTimeFormat), utils.GetCaller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Error(formatter string, args ...any) {
	fmt.Printf("%s%s ERROR %s %s%s\n", ColorRed, time.Now().Format(DefaultTimeFormat), utils.GetCaller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func Fatal(formatter string, args ...any) {
	fmt.Printf("%s%s FATAL %s %s%s\n", ColorPurple, time.Now().Format(DefaultTimeFormat), utils.GetCaller(2), fmt.Sprintf(formatter, args...), ColorReset)
}
