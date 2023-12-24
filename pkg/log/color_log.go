package log

import (
	"fmt"
	"time"

	"github.com/soulnov23/go-tool/pkg/utils"
)

const (
	colorRed    = "\033[1;31m"
	colorGreen  = "\033[1;32m"
	colorYellow = "\033[1;33m"
	colorPurple = "\033[1;35m"
	colorWhite  = "\033[1;37m"
	colorReset  = "\033[m"
)

func ColorDebugf(formatter string, args ...any) {
	fmt.Printf("%s%s DEBUG %s %s%s\n", colorGreen, time.Now().Format(time.DateTime), utils.Caller(2), fmt.Sprintf(formatter, args...), colorReset)
}

func ColorInfof(formatter string, args ...any) {
	fmt.Printf("%s%s INFO %s %s%s\n", colorWhite, time.Now().Format(time.DateTime), utils.Caller(2), fmt.Sprintf(formatter, args...), colorReset)
}

func ColorWarnf(formatter string, args ...any) {
	fmt.Printf("%s%s WARN %s %s%s\n", colorYellow, time.Now().Format(time.DateTime), utils.Caller(2), fmt.Sprintf(formatter, args...), colorReset)
}

func ColorErrorf(formatter string, args ...any) {
	fmt.Printf("%s%s ERROR %s %s%s\n", colorRed, time.Now().Format(time.DateTime), utils.Caller(2), fmt.Sprintf(formatter, args...), colorReset)
}

func ColorFatalf(formatter string, args ...any) {
	fmt.Printf("%s%s FATAL %s %s%s\n", colorPurple, time.Now().Format(time.DateTime), utils.Caller(2), fmt.Sprintf(formatter, args...), colorReset)
}
