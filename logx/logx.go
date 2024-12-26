// Package logx provides a flexible, module-based logging system built on top of slog.
// It supports context extraction, multiple log levels, and concurrent safe operations.
// The package is designed to be used in large applications where different modules
// need their own logging configurations while maintaining a consistent logging interface.
package logx

import (
	"context"
	"log/slog"
	"sync"
)

// Logx represents a logger instance for a specific module.
// It extends slog.Logger with module-specific functionality and context extraction capabilities.
// Each Logx instance is associated with a unique module name and can have its own
// set of context extractors for pulling values from context.Context.
type Logx struct {
	*slog.Logger
	moduleName        string
	contextExtractors map[string]ContextExtractor
}

// moduleLoggers is a thread-safe map storing all registered logger instances.
// The key is the module name and the value is the corresponding Logx instance.
var moduleLoggers sync.Map

// defaultModuleName is the name used when no specific module name is provided.
const defaultModuleName = "default"

// getModuleName returns the module name or default if empty.
// This ensures we always have a valid module name for logging.
func getModuleName(name string) string {
	if name == "" {
		return defaultModuleName
	}
	return name
}

// Register creates and registers a new logger instance for the specified module.
// Parameters:
//   - moduleName: The name of the module. If empty, uses default module name.
//   - options: Configuration options for the logger. If nil, uses default options.
//
// Returns:
//   - *Logx: The new or existing logger instance for the module.
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
// This is an internal function used by Register to initialize a new logger.
// Parameters:
//   - moduleName: The name of the module.
//   - options: Configuration options for the logger.
func newLogx(moduleName string, options *Options) *Logx {
	logger := &Logx{
		moduleName: moduleName,
	}
	configLogger(logger, options)
	return logger
}

// GetLogger retrieves the logger instance for the specified module.
// This is useful when you want to check if a logger exists without creating one.
// Parameters:
//   - moduleName: The name of the module to look up.
//
// Returns:
//   - *Logx: The logger instance if found, nil otherwise
//   - bool: true if logger was found, false otherwise
func GetLogger(moduleName string) (*Logx, bool) {
	moduleName = getModuleName(moduleName)
	if logx, ok := moduleLoggers.Load(moduleName); ok {
		return logx.(*Logx), true
	}
	return nil, false
}

// Default returns the default logger instance.
// If no default logger exists, it creates one with default options.
// This method is useful when you don't need module-specific logging.
func Default() *Logx {
	logger, _ := GetLogger(defaultModuleName)
	if logger == nil {
		return Register("", nil) // will use default options
	}
	return logger
}

// InfoContext logs a message at Info level with context-extracted values.
// It automatically extracts values from context using registered extractors
// and adds them to the log entry.
// Parameters:
//   - ctx: Context containing values to extract
//   - msg: The message to log
//   - args: Additional key-value pairs to include in the log
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

// configLogger configures a logger instance with the provided options.
// It sets up the JSON handler, source addition, log level, and context extractors.
// If options is nil, default options are used.
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

// setValueFromContext extracts values from context using registered extractors
// and appends them to the provided arguments list.
// This is an internal function used by all logging methods.
// Parameters:
//   - ctx: The context to extract values from
//   - ce: Map of context extractors
//   - args: Existing arguments to append to
//
// Returns:
//   - []any: Combined slice of existing args and extracted values
func setValueFromContext(ctx context.Context, ce map[string]ContextExtractor, args ...any) []any {
	for key, extractor := range ce {
		v := extractor(ctx)
		args = append(args, slog.String(key, v))
	}
	return args
}
