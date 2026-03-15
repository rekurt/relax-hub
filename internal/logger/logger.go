package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	sugar *zap.SugaredLogger
	level LogLevel
}

// New creates a Logger with the given level using console format.
// For production use, prefer NewWithFormat to get JSON output.
func New(level LogLevel) *Logger {
	return NewWithFormat(level, "console")
}

// NewWithFormat creates a Logger with the given level and output format ("json" or "console").
func NewWithFormat(level LogLevel, format string) *Logger {
	var cfg zap.Config
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(toZapLevel(level))
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		// Fallback to nop logger — should never happen with valid config
		z = zap.NewNop()
	}

	return &Logger{
		sugar: z.Sugar(),
		level: level,
	}
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.sugar.Debugw(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.sugar.Infow(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.sugar.Warnw(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.sugar.Errorw(msg, fields...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() {
	_ = l.sugar.Sync()
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

func toZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
