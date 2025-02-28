package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	types "github.com/magooney-loon/webserver/types/server"
)

var (
	loggerInstance *logger
	loggerOnce     sync.Once
)

type logger struct {
	config *types.LoggerConfig
	fields types.Fields
	mu     sync.RWMutex
}

func NewLogger(options ...types.LoggerOption) types.Logger {
	loggerOnce.Do(func() {
		cfg := &types.LoggerConfig{
			Level:       types.InfoLevel,
			ServiceName: "webserver",
			OutputPaths: []string{"stdout"},
		}

		for _, opt := range options {
			opt(cfg)
		}

		loggerInstance = &logger{
			config: cfg,
			fields: make(types.Fields),
		}
	})

	return loggerInstance
}

func (l *logger) With(fields types.Fields) types.Logger {
	newLogger := &logger{
		config: l.config,
		fields: make(types.Fields, len(l.fields)+len(fields)),
	}

	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

func (l *logger) WithContext(ctx context.Context) types.Logger {
	if ctx == nil {
		return l
	}
	return l.With(types.Fields{"trace_id": ctx.Value("trace_id")})
}

func (l *logger) log(level types.LogLevel, msg string, fields ...types.Fields) {
	if level < l.config.Level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().UTC().Format("01-02T15:04:05.000Z")
	logLine := fmt.Sprintf("<%s> [%s:%d] = %s //%s//",
		l.config.ServiceName,
		level.String(),
		level,
		msg,
		timestamp,
	)

	// Add context fields if they exist
	if len(l.fields) > 0 {
		ctxData, _ := json.Marshal(l.fields)
		logLine += fmt.Sprintf(" ctx=%s", string(ctxData))
	}

	fmt.Fprintf(os.Stdout, "%s\n", logLine)

	if level == types.FatalLevel {
		os.Exit(1)
	}
}

func (l *logger) Debug(msg string, fields ...types.Fields) { l.log(types.DebugLevel, msg, fields...) }
func (l *logger) Info(msg string, fields ...types.Fields)  { l.log(types.InfoLevel, msg, fields...) }
func (l *logger) Warn(msg string, fields ...types.Fields)  { l.log(types.WarnLevel, msg, fields...) }
func (l *logger) Error(msg string, fields ...types.Fields) { l.log(types.ErrorLevel, msg, fields...) }
func (l *logger) Fatal(msg string, fields ...types.Fields) { l.log(types.FatalLevel, msg, fields...) }
