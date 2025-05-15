package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
	FatalLevel LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Component   string                 `json:"component"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Logger wraps zap logger with additional context
type Logger struct {
	*zap.Logger
	component string
	traceID   string
}

var (
	globalLogger *zap.Logger
	appEnv       string
)

// Initialize sets up the logger with the specified environment and log level string.
func Initialize(environment string, logLevelString string) error {
	appEnv = environment

	var level zapcore.Level
	switch strings.ToUpper(logLevelString) {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "FATAL":
		level = zapcore.FatalLevel
	default:
		fmt.Printf("Warning: Invalid LOG_LEVEL '%s', defaulting to DEBUG\n", logLevelString)
		level = zapcore.DebugLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	globalLogger = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return nil
}

// WithContext creates a new logger with request context
func WithContext(ctx context.Context, component string) *Logger {
	// Get trace ID from context or generate new one
	var traceID string
	if id := ctx.Value("request_id"); id != nil {
		traceID = id.(string)
	} else {
		traceID = uuid.New().String()
	}

	return &Logger{
		Logger:    globalLogger,
		component: component,
		traceID:   traceID,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, metadata map[string]interface{}) {
	l.log(DebugLevel, msg, nil, metadata)
}

// Info logs an info message
func (l *Logger) Info(msg string, metadata map[string]interface{}) {
	l.log(InfoLevel, msg, nil, metadata)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, metadata map[string]interface{}) {
	l.log(WarnLevel, msg, nil, metadata)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, metadata map[string]interface{}) {
	l.log(ErrorLevel, msg, err, metadata)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, metadata map[string]interface{}) {
	l.log(FatalLevel, msg, err, metadata)
	os.Exit(1)
}

// log handles the actual logging
func (l *Logger) log(level LogLevel, msg string, err error, metadata map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       level,
		Message:     msg,
		TraceID:     l.traceID,
		Service:     "todo-service",
		Environment: appEnv,
		Component:   l.component,
		Metadata:    metadata,
	}

	if err != nil {
		entry.Error = err.Error()
		if level >= ErrorLevel {
			// Capture stack trace for errors
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			entry.Stack = string(buf[:n])
		}
	}

	// Convert to JSON and log
	if jsonData, err := json.Marshal(entry); err == nil {
		switch level {
		case DebugLevel:
			l.Logger.Debug(string(jsonData))
		case InfoLevel:
			l.Logger.Info(string(jsonData))
		case WarnLevel:
			l.Logger.Warn(string(jsonData))
		case ErrorLevel:
			l.Logger.Error(string(jsonData))
		case FatalLevel:
			l.Logger.Fatal(string(jsonData))
		}
	} else {
		// Fallback if JSON marshaling fails
		fmt.Printf("Failed to marshal log entry: %v\n", err)
	}
}
