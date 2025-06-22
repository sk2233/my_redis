package main

import (
	"fmt"
	"runtime"
	"time"
)

var (
	LogLevel = LogInfo
)

func getLocation() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", file, line)
}

func getTime() string {
	return time.Now().Format(TimeLayout)
}

func Info(format string, args ...any) {
	if LogLevel > LogInfo {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s [Info][%s] %s%s\n", ColorGreen, getTime(), getLocation(), msg, ColorReset)
}

func Warn(format string, args ...any) {
	if LogLevel > LogWarn {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s [Warn][%s] %s%s\n", ColorYellow, getTime(), getLocation(), msg, ColorReset)
}

func Error(format string, args ...any) {
	if LogLevel > LogError {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s [Error][%s] %s%s\n", ColorRed, getTime(), getLocation(), msg, ColorReset)
}
