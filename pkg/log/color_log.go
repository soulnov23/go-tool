package log

import (
	"fmt"
	"time"

	run "github.com/soulnov23/go-tool/pkg/runtime"
)

const (
	ColorRed    = "\033[1;31m"
	ColorGreen  = "\033[1;32m"
	ColorYellow = "\033[1;33m"
	ColorPurple = "\033[1;35m"
	ColorWhite  = "\033[1;37m"
	ColorReset  = "\033[m"
)

func ColorDebugf(formatter string, args ...any) {
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorGreen, time.Now().Format(time.DateTime), run.Caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func ColorInfof(formatter string, args ...any) {
	fmt.Printf("%s%s INFO %s %s%s\n", ColorWhite, time.Now().Format(time.DateTime), run.Caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func ColorWarnf(formatter string, args ...any) {
	fmt.Printf("%s%s WARN %s %s%s\n", ColorYellow, time.Now().Format(time.DateTime), run.Caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func ColorErrorf(formatter string, args ...any) {
	fmt.Printf("%s%s ERROR %s %s%s\n", ColorRed, time.Now().Format(time.DateTime), run.Caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}

func ColorFatalf(formatter string, args ...any) {
	fmt.Printf("%s%s FATAL %s %s%s\n", ColorPurple, time.Now().Format(time.DateTime), run.Caller(2), fmt.Sprintf(formatter, args...), ColorReset)
}
