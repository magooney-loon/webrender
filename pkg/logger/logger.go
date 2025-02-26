package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Logger represents a simple logger with leveled logging capabilities
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	useJSON bool
	level   string
	fields  map[string]interface{}
}

// Option defines a function type for configuring the logger
type Option func(*Logger)

// WithJSON sets whether to output logs in JSON format
func WithJSON(enabled bool) Option {
	return func(l *Logger) {
		l.useJSON = enabled
	}
}

// WithLevel sets the minimum log level
func WithLevel(level string) Option {
	return func(l *Logger) {
		l.level = strings.ToUpper(level)
	}
}

// WithOutput sets the output writer for logs
func WithOutput(out io.Writer) Option {
	return func(l *Logger) {
		l.out = out
	}
}

// WithFields adds default fields to all log entries
func WithFields(fields map[string]interface{}) Option {
	return func(l *Logger) {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
}

// New creates a new logger with the given options
func New(options ...Option) *Logger {
	l := &Logger{
		out:     os.Stdout,
		level:   "INFO",
		fields:  make(map[string]interface{}),
		useJSON: false,
	}

	for _, opt := range options {
		opt(l)
	}

	return l
}

// shouldLog determines if the given level should be logged
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"DEBUG": 0,
		"INFO":  1,
		"WARN":  2,
		"ERROR": 3,
		"FATAL": 4,
	}

	currentLevel, ok := levels[l.level]
	if !ok {
		currentLevel = levels["INFO"] // Default to INFO if level is invalid
	}

	levelValue, ok := levels[level]
	if !ok {
		return false // Don't log unknown levels
	}

	return levelValue >= currentLevel
}

// log logs a message at the given level with additional fields
func (l *Logger) log(level string, msg string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	data := make(map[string]interface{})

	// Merge base fields
	for k, v := range l.fields {
		data[k] = v
	}

	// Merge additional fields
	for k, v := range fields {
		data[k] = v
	}

	data["level"] = level
	data["msg"] = msg
	data["timestamp"] = now.Format(time.RFC3339)

	if l.useJSON {
		if err := json.NewEncoder(l.out).Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode log entry: %v\n", err)
		}
		return
	}

	// Simple formatted output
	fmt.Fprintf(l.out, "[%s] %-5s %s",
		now.Format("2006-01-02 15:04:05"),
		level,
		msg,
	)

	if len(fields) > 0 {
		fmt.Fprint(l.out, " ")
		first := true
		for k, v := range fields {
			if !first {
				fmt.Fprint(l.out, " ")
			}
			fmt.Fprintf(l.out, "%s=%v", k, v)
			first = false
		}
	}
	fmt.Fprintln(l.out)
}

// With returns a new logger with additional default fields
func (l *Logger) With(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		out:     l.out,
		useJSON: l.useJSON,
		level:   l.level,
		fields:  make(map[string]interface{}, len(l.fields)+len(fields)),
	}

	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log("DEBUG", msg, fields)
}

// Info logs a message at INFO level
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log("INFO", msg, fields)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log("WARN", msg, fields)
}

// Error logs a message at ERROR level
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log("ERROR", msg, fields)
}

// Fatal logs a message at FATAL level and exits the program
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.log("FATAL", msg, fields)
	os.Exit(1)
}
