package main

import (
	"fmt"
	"time"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Critical
)

type Logger struct {
	LogLevel LogLevel
}

func NewLogger() *Logger {
	return &Logger{LogLevel: Warn}
}

func (l *Logger) Debug(msg string) {
	if l.LogLevel <= Debug {
		now := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] [DEBUG] %s\n", now, msg)
	}
}

func (l *Logger) Info(msg string) {
	if l.LogLevel <= Info {
		now := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] [INFO] %s\n", now, msg)
	}
}

func (l *Logger) Warn(msg string) {
	if l.LogLevel <= Warn {
		now := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] [WARN] %s\n", now, msg)
	}
}

func (l *Logger) Error(msg string) {
	if l.LogLevel <= Error {
		now := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] [ERROR] %s\n", now, msg)
	}
}

func (l *Logger) Critical(msg string) {
	if l.LogLevel <= Critical {
		now := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] [CRITICAL] %s\n", now, msg)
	}
}
