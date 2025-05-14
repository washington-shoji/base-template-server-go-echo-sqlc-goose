package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
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

var (
	logger *zap.Logger
	env    string
)

// Initialize sets up the logger with the specified environment
func Initialize(environment string) error {
	env = environment

	// Configure encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Set log level based on environment
	var level zapcore.Level
	if environment == "production" {
		level = zapcore.InfoLevel
	} else {
		level = zapcore.DebugLevel
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// WithContext creates a new logger with context values
func WithContext(ctx context.Context, component string) *Logger {
	traceID := getTraceID(ctx)
	return &Logger{
		component: component,
		traceID:   traceID,
	}
}

// Logger provides methods for structured logging
type Logger struct {
	component string
	traceID   string
}

// getTraceID extracts or generates a trace ID from context
func getTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return uuid.New().String()
}

// log creates a structured log entry
func (l *Logger) log(level LogLevel, message string, err error, metadata map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       level,
		Message:     message,
		TraceID:     l.traceID,
		Service:     "go-echo-server-template",
		Environment: env,
		Component:   l.component,
		Metadata:    metadata,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if level == ErrorLevel || level == FatalLevel {
		// Get stack trace
		const depth = 32
		var pcs [depth]uintptr
		n := runtime.Callers(3, pcs[:])
		frames := runtime.CallersFrames(pcs[:n])
		stack := ""
		for {
			frame, more := frames.Next()
			stack += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
			if !more {
				break
			}
		}
		entry.Stack = stack
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(entry)
	fmt.Println(string(jsonData))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, metadata map[string]interface{}) {
	l.log(DebugLevel, message, nil, metadata)
}

// Info logs an info message
func (l *Logger) Info(message string, metadata map[string]interface{}) {
	l.log(InfoLevel, message, nil, metadata)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, metadata map[string]interface{}) {
	l.log(WarnLevel, message, nil, metadata)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, metadata map[string]interface{}) {
	l.log(ErrorLevel, message, err, metadata)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, metadata map[string]interface{}) {
	l.log(FatalLevel, message, err, metadata)
	os.Exit(1)
}
