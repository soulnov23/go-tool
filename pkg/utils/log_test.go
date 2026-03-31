package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestColor(t *testing.T) {
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorBlack, time.Now().Format(time.DateTime), caller(2), "black", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorRed, time.Now().Format(time.DateTime), caller(2), "red", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorGreen, time.Now().Format(time.DateTime), caller(2), "green", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorYellow, time.Now().Format(time.DateTime), caller(2), "yellow", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorBlue, time.Now().Format(time.DateTime), caller(2), "blue", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorPurple, time.Now().Format(time.DateTime), caller(2), "purple", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorCyan, time.Now().Format(time.DateTime), caller(2), "cyan", ColorReset)
	fmt.Printf("%s%s DEBUG %s %s%s\n", ColorWhite, time.Now().Format(time.DateTime), caller(2), "white", ColorReset)
}
