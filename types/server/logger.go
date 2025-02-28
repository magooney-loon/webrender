package types

import "context"

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Fields map[string]interface{}

type Logger interface {
	With(fields Fields) Logger
	WithContext(ctx context.Context) Logger

	Debug(msg string, fields ...Fields)
	Info(msg string, fields ...Fields)
	Warn(msg string, fields ...Fields)
	Error(msg string, fields ...Fields)
	Fatal(msg string, fields ...Fields)
}

type LoggerOption func(*LoggerConfig)

type LoggerConfig struct {
	Level       LogLevel
	ServiceName string
	Environment string
	OutputPaths []string
	Development bool
}
