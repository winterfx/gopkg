// Package logx provides a flexible, module-based logging system built on top of slog.
// It supports context extraction, multiple log levels, and concurrent safe operations.
package logx

import (
	"context"
	"log/slog"
	"sync"
)

// Logx represents a logger instance for a specific module.
type Logx struct {
	*slog.Logger
	moduleName        string
	contextExtractors map[string]ContextExtractor
}

var moduleLoggers sync.Map

const defaultModuleName = "default"

func getModuleName(name string) string {
	if name == "" {
		return defaultModuleName
	}
	return name
}

// Register creates and registers a new logger instance for the specified module.
// If a logger for the module already exists, it returns the existing instance.
// If moduleName is empty, it uses the default module name.
func Register(moduleName string, options *Options) *Logx {
	moduleName = getModuleName(moduleName)
	logx, ok := moduleLoggers.Load(moduleName)
	if !ok {
		logx = newLogx(moduleName, options)
		moduleLoggers.Store(moduleName, logx)
	}
	return logx.(*Logx)
}

// newLogx creates a new logger instance for the specified module with given options.
func newLogx(moduleName string, options *Options) *Logx {
	logger := &Logx{
		moduleName: moduleName,
	}
	configLogger(logger, options)
	return logger
}

// GetLogger retrieves the logger instance for the specified module.
// Returns the logger instance and true if found, nil and false otherwise.
func GetLogger(moduleName string) (*Logx, bool) {
	moduleName = getModuleName(moduleName)
	if logx, ok := moduleLoggers.Load(moduleName); ok {
		return logx.(*Logx), true
	}
	return nil, false
}

// Default returns the default logger instance.
// If the default logger doesn't exist, it creates one with default options.
func Default() *Logx {
	logger, _ := GetLogger(defaultModuleName)
	if logger == nil {
		return Register("", nil) // will use default options
	}
	return logger
}

// InfoContext logs a message at Info level with context-extracted values.
func (c *Logx) InfoContext(ctx context.Context, msg string, args ...any) {
	args = setValueFromContext(ctx, c.contextExtractors, args...)
	c.Logger.InfoContext(ctx, msg, args...)
}

// DebugContext logs a message at Debug level with context-extracted values.
func (c *Logx) DebugContext(ctx context.Context, msg string, args ...any) {
	args = setValueFromContext(ctx, c.contextExtractors, args...)
	c.Logger.DebugContext(ctx, msg, args...)
}

// ErrorContext logs a message at Error level with context-extracted values.
func (c *Logx) ErrorContext(ctx context.Context, msg string, args ...any) {
	args = setValueFromContext(ctx, c.contextExtractors, args...)
	c.Logger.ErrorContext(ctx, msg, args...)
}

// WarnContext logs a message at Warn level with context-extracted values.
func (c *Logx) WarnContext(ctx context.Context, msg string, args ...any) {
	args = setValueFromContext(ctx, c.contextExtractors, args...)
	c.Logger.WarnContext(ctx, msg, args...)
}

func configLogger(logger *Logx, options *Options) {
	if options == nil {
		options = defaultOptions()
	}
	slogOpts := &slog.HandlerOptions{
		AddSource: options.addSource,
		Level:     options.level,
	}

	handler := slog.NewJSONHandler(options.output, slogOpts)
	logger.Logger = slog.New(handler)
	logger.contextExtractors = options.contextExtractors
	logger.Logger = logger.Logger.With(slog.String("module", logger.moduleName))
}

func setValueFromContext(ctx context.Context, ce map[string]ContextExtractor, args ...any) []any {
	for key, extractor := range ce {
		v := extractor(ctx)
		args = append(args, slog.String(key, v))
	}
	return args
}
