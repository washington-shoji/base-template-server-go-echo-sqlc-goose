package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
	FatalLevel LogLevel = "FATAL"
)

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

type Logger struct {
	*zap.Logger
	component string
	traceID   string
}

var (
	globalLogger *zap.Logger
	appEnv       string
)

func Initialize(environment, logLevelString string) error {
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

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func Sync() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Sync()
}

func WithContext(ctx context.Context, component string) *Logger {
	traceID := RequestIDFromContext(ctx)
	if traceID == "" {
		traceID = uuid.New().String()
	}
	return &Logger{Logger: globalLogger, component: component, traceID: traceID}
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	// Backward-compatible string key used by older test helpers.
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

func (l *Logger) Debug(msg string, metadata map[string]interface{}) {
	l.log(DebugLevel, msg, nil, metadata)
}
func (l *Logger) Info(msg string, metadata map[string]interface{}) {
	l.log(InfoLevel, msg, nil, metadata)
}
func (l *Logger) Warn(msg string, metadata map[string]interface{}) {
	l.log(WarnLevel, msg, nil, metadata)
}
func (l *Logger) Error(msg string, err error, metadata map[string]interface{}) {
	l.log(ErrorLevel, msg, err, metadata)
}
func (l *Logger) Fatal(msg string, err error, metadata map[string]interface{}) {
	l.log(FatalLevel, msg, err, metadata)
	os.Exit(1)
}

func (l *Logger) log(level LogLevel, msg string, err error, metadata map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Level:       level,
		Message:     msg,
		TraceID:     l.traceID,
		Service:     "app",
		Environment: appEnv,
		Component:   l.component,
		Metadata:    metadata,
	}
	if err != nil {
		entry.Error = err.Error()
		if level == ErrorLevel || level == FatalLevel {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			entry.Stack = string(buf[:n])
		}
	}
	jsonData, marshalErr := json.Marshal(entry)
	if marshalErr != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", marshalErr)
		return
	}
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
}

// Middleware injects Echo's request ID into the request context and logs HTTP traffic.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = c.Request().Header.Get(echo.HeaderXRequestID)
			}
			if reqID == "" {
				reqID = uuid.New().String()
			}
			ctx := ContextWithRequestID(c.Request().Context(), reqID)
			c.SetRequest(c.Request().WithContext(ctx))

			log := WithContext(ctx, "http")
			log.Info("Incoming request", map[string]interface{}{
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
				"remote_ip":  c.Request().RemoteAddr,
				"user_agent": c.Request().UserAgent(),
				"request_id": reqID,
			})

			err := next(c)
			if err != nil {
				log.Error("Request failed", err, map[string]interface{}{"status_code": c.Response().Status})
			} else {
				log.Info("Request completed", map[string]interface{}{"status_code": c.Response().Status})
			}
			return err
		}
	}
}
