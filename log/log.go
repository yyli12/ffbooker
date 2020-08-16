package log

import (
	"fmt"
	"time"
)

var (
	red    = color("\033[1;31m%s\033[0m")
	yellow = color("\033[1;33m%s\033[0m")
	teal   = color("\033[1;36m%s\033[0m")
)

var (
	Info  = print(teal)
	Warn  = print(yellow)
	Error = print(red)
)

func color(colorString string) func(...interface{}) string {
	return func(args ...interface{}) string {
		if fmtString, ok := args[0].(string); len(args) > 0 && ok {
			moreArgs := []interface{}{time.Now().Format("2006-01-02 15:04:05 -0700")}
			moreArgs = append(moreArgs, args[1:]...)
			return fmt.Sprintf(colorString, fmt.Sprintf("[%s] "+fmtString, moreArgs...))
		}
		return fmt.Sprintf(colorString, fmt.Sprint(args))
	}
}

func print(colorFn func(...interface{}) string) func(args ...interface{}) {
	return func(args ...interface{}) {
		fmt.Println(colorFn(args...))
	}
}
