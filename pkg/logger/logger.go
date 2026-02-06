// Package logger provides a structured logging wrapper around zap.
// It supports JSON/Console formats, log rotation, context-aware logging,
// and performance mode with async logging.
package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// contextKey is used for storing logger fields in context
type contextKey struct{}

var (
	// defaultLogger is the package-level default logger
	defaultLogger *Logger
	initOnce      sync.Once
)

// Config holds logger configuration
type Config struct {
	// Level is the minimum log level: debug, info, warn, error, fatal
	Level string `json:"level" yaml:"level"`
	// Format is the output format: json or console
	Format string `json:"format" yaml:"format"`
	// Output is the log file path. Use "stdout" or "stderr" for console output
	Output string `json:"output" yaml:"output"`
	// MaxSize is the maximum size in megabytes before log rotation
	MaxSize int `json:"max_size" yaml:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `json:"max_backups" yaml:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `json:"max_age" yaml:"max_age"`
	// Compress determines if rotated logs should be gzipped
	Compress bool `json:"compress" yaml:"compress"`
	// Async enables async logging for better performance
	Async bool `json:"async" yaml:"async"`
	// AsyncBufferSize is the size of the async log buffer
	AsyncBufferSize int `json:"async_buffer_size" yaml:"async_buffer_size"`
	// Development mode enables stack traces and more verbose output
	Development bool `json:"development" yaml:"development"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Level:           "info",
		Format:          "console",
		Output:          "stdout",
		MaxSize:         100,
		MaxBackups:      5,
		MaxAge:          30,
		Compress:        true,
		Async:           false,
		AsyncBufferSize: 1000,
		Development:     false,
	}
}

// Logger is a structured logger wrapper around zap
type Logger struct {
	zap    *zap.Logger
	sugar  *zap.SugaredLogger
	async  bool
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// New creates a new Logger with the given configuration
func New(cfg Config) (*Logger, error) {
	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create encoder config
	ec := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Development {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
		ec.EncodeCaller = zapcore.FullCallerEncoder
	}

	// Create encoder based on format
	var encoder zapcore.Encoder
	switch strings.ToLower(cfg.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(ec)
	case "console":
		encoder = zapcore.NewConsoleEncoder(ec)
	default:
		return nil, fmt.Errorf("invalid format: %s (must be 'json' or 'console')", cfg.Format)
	}

	// Create write syncer
	ws, cleanup, err := newWriteSyncer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create write syncer: %w", err)
	}

	// Create core
	core := zapcore.NewCore(encoder, ws, level)

	// Add async wrapper if enabled
	l := &Logger{
		async:  cfg.Async,
		stopCh: make(chan struct{}),
	}

	if cfg.Async {
		core = &asyncCore{
			Core:        core,
			bufferSize:  cfg.AsyncBufferSize,
			stopCh:      l.stopCh,
			wg:          &l.wg,
		}
	}

	// Create zap logger
	zapOpts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	if cfg.Development {
		zapOpts = append(zapOpts, zap.Development())
	}
	if cleanup != nil {
		zapOpts = append(zapOpts, zap.Hooks(cleanup))
	}

	l.zap = zap.New(core, zapOpts...)
	l.sugar = l.zap.Sugar()

	return l, nil
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	l, err := New(DefaultConfig())
	if err != nil {
		// Fallback to a basic logger that can't fail
		z, _ := zap.NewProduction()
		return &Logger{zap: z, sugar: z.Sugar()}
	}
	return l
}

// SetDefault sets the package-level default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Default returns the package-level default logger
func Default() *Logger {
	initOnce.Do(func() {
		if defaultLogger == nil {
			defaultLogger = NewDefault()
		}
	})
	return defaultLogger
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	if l.async {
		close(l.stopCh)
		l.wg.Wait()
	}
	return l.zap.Sync()
}

// With creates a new logger with the given fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zap:   l.zap.With(fields...),
		sugar: l.sugar.With(fieldsToArgs(fields)...),
	}
}

// WithContext returns a context with the given fields attached
func (l *Logger) WithContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, contextKey{}, fields)
}

// WithRequestID returns a context with request ID attached
func (l *Logger) WithRequestID(ctx context.Context, requestID string) context.Context {
	return l.WithContext(ctx, zap.String("request_id", requestID))
}

// WithSessionID returns a context with session ID attached
func (l *Logger) WithSessionID(ctx context.Context, sessionID string) context.Context {
	return l.WithContext(ctx, zap.String("session_id", sessionID))
}

// getContextFields extracts fields from context
func getContextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}
	if fields, ok := ctx.Value(contextKey{}).([]zap.Field); ok {
		return fields
	}
	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// DebugContext logs a debug message with context fields
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(getContextFields(ctx), fields...)
	l.zap.Debug(msg, allFields...)
}

// InfoContext logs an info message with context fields
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(getContextFields(ctx), fields...)
	l.zap.Info(msg, allFields...)
}

// WarnContext logs a warning message with context fields
func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(getContextFields(ctx), fields...)
	l.zap.Warn(msg, allFields...)
}

// ErrorContext logs an error message with context fields
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(getContextFields(ctx), fields...)
	l.zap.Error(msg, allFields...)
}

// FatalContext logs a fatal message with context fields and exits
func (l *Logger) FatalContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(getContextFields(ctx), fields...)
	l.zap.Fatal(msg, allFields...)
}

// Sugar methods for convenience

// Debugf logs a formatted debug message
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// Package-level functions that use the default logger

// Debug uses the default logger
func Debug(msg string, fields ...zap.Field) { Default().Debug(msg, fields...) }

// Info uses the default logger
func Info(msg string, fields ...zap.Field) { Default().Info(msg, fields...) }

// Warn uses the default logger
func Warn(msg string, fields ...zap.Field) { Default().Warn(msg, fields...) }

// Error uses the default logger
func Error(msg string, fields ...zap.Field) { Default().Error(msg, fields...) }

// Fatal uses the default logger
func Fatal(msg string, fields ...zap.Field) { Default().Fatal(msg, fields...) }

// DebugContext uses the default logger
func DebugContext(ctx context.Context, msg string, fields ...zap.Field) { Default().DebugContext(ctx, msg, fields...) }

// InfoContext uses the default logger
func InfoContext(ctx context.Context, msg string, fields ...zap.Field) { Default().InfoContext(ctx, msg, fields...) }

// WarnContext uses the default logger
func WarnContext(ctx context.Context, msg string, fields ...zap.Field) { Default().WarnContext(ctx, msg, fields...) }

// ErrorContext uses the default logger
func ErrorContext(ctx context.Context, msg string, fields ...zap.Field) { Default().ErrorContext(ctx, msg, fields...) }

// FatalContext uses the default logger
func FatalContext(ctx context.Context, msg string, fields ...zap.Field) { Default().FatalContext(ctx, msg, fields...) }

// parseLevel parses a log level string
func parseLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown level: %s", level)
	}
}

// newWriteSyncer creates a write syncer based on output configuration
func newWriteSyncer(cfg Config) (zapcore.WriteSyncer, func(zapcore.Entry) error, error) {
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		return zapcore.AddSync(os.Stdout), nil, nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil, nil
	default:
		// Ensure log directory exists
		dir := filepath.Dir(cfg.Output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Create lumberjack logger for rotation
		lj := &lumberjack.Logger{
			Filename:   cfg.Output,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}

		cleanup := func(zapcore.Entry) error {
			return lj.Close()
		}

		return zapcore.AddSync(lj), cleanup, nil
	}
}

// fieldsToArgs converts zap.Fields to sugar args
func fieldsToArgs(fields []zap.Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Interface)
	}
	return args
}

// asyncCore wraps a zapcore.Core to provide async logging
type asyncCore struct {
	zapcore.Core
	bufferSize int
	entries    chan zapcore.Entry
	fields     chan []zapcore.Field
	stopCh     chan struct{}
	wg         *sync.WaitGroup
	initOnce   sync.Once
}

func (c *asyncCore) init() {
	c.initOnce.Do(func() {
		c.entries = make(chan zapcore.Entry, c.bufferSize)
		c.fields = make(chan []zapcore.Field, c.bufferSize)
		c.wg.Add(1)
		go c.process()
	})
}

func (c *asyncCore) process() {
	defer c.wg.Done()
	for {
		select {
		case entry := <-c.entries:
			fields := <-c.fields
			if ce := c.Core.Check(entry, nil); ce != nil {
				ce.Write(fields...)
			}
		case <-c.stopCh:
			// Drain remaining entries
			for {
				select {
				case entry := <-c.entries:
					fields := <-c.fields
					if ce := c.Core.Check(entry, nil); ce != nil {
						ce.Write(fields...)
					}
				default:
					return
				}
			}
		}
	}
}

func (c *asyncCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	c.init()
	select {
	case c.entries <- entry:
		c.fields <- fields
		return nil
	default:
		// Buffer full, fall back to sync write
		return c.Core.Write(entry, fields)
	}
}

func (c *asyncCore) Sync() error {
	// Drain entries before syncing
	for {
		select {
		case entry := <-c.entries:
			fields := <-c.fields
			if ce := c.Core.Check(entry, nil); ce != nil {
				ce.Write(fields...)
			}
		default:
			return c.Core.Sync()
		}
	}
}
