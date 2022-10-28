package logger

import (
	"fmt"
)

var (
	LoggingLevel int = 1
)

func Debug(message string) {
	if LoggingLevel <= 0 {
		fmt.Println(message)
	}
}

func Debug1(message string) {
	if LoggingLevel <= -1 {
		fmt.Println(message)
	}
}
