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
	ColorBlue   = "\033[1;34m"
	ColorPink   = "\033[1;35m"
	ColorReset  = "\033[m"

	DefaultTimeFormat = "2006-01-02 15:04:05.000"
)

func Info(formatter string, args ...interface{}) {
	fmt.Printf("%s %sINFO%s %s %s\n", time.Now().Format(DefaultTimeFormat), ColorYellow, ColorReset, utils.GetCaller(2), fmt.Sprintf(formatter, args...))
}

func Debug(formatter string, args ...interface{}) {
	fmt.Printf("%s %sDEBUG%s %s %s\n", time.Now().Format(DefaultTimeFormat), ColorBlue, ColorReset, utils.GetCaller(2), fmt.Sprintf(formatter, args...))
}

func Error(formatter string, args ...interface{}) {
	fmt.Printf("%s %sERROR%s %s %s\n", time.Now().Format(DefaultTimeFormat), ColorRed, ColorReset, utils.GetCaller(2), fmt.Sprintf(formatter, args...))
}
