package logger

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level LogLevel
}

func New(level LogLevel) *Logger {
	return &Logger{
		level: level,
	}
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.level <= LevelDebug {
		l.logf("DEBUG", msg, fields...)
	}
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.level <= LevelInfo {
		l.logf("INFO", msg, fields...)
	}
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	if l.level <= LevelWarn {
		l.logf("WARN", msg, fields...)
	}
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	if l.level <= LevelError {
		l.logf("ERROR", msg, fields...)
	}
}

func (l *Logger) logf(level, msg string, fields ...interface{}) {
	logMsg := fmt.Sprintf("[%s] %s", level, msg)
	if len(fields) > 0 {
		logMsg += fmt.Sprintf(" %v", fields)
	}
	log.New(os.Stdout, "", log.LstdFlags).Println(logMsg)
}

func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
