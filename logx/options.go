package logx

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Options defines the configuration options for a logger instance.
type Options struct {
	level             slog.Level
	addSource         bool
	output            io.Writer
	contextExtractors map[string]ContextExtractor
}

// ContextExtractor is a function type that extracts string values from context.
type ContextExtractor func(ctx context.Context) string

// OptionsFunc is a function type for modifying Options.
type OptionsFunc func(*Options)

// defaultOptions returns a new Options instance with default values.
func defaultOptions() *Options {
	return &Options{
		level:             slog.LevelInfo,
		addSource:         true,
		output:            os.Stdout,
		contextExtractors: make(map[string]ContextExtractor),
	}
}

// WithLevel sets the logging level for the logger.
func WithLevel(level slog.Level) OptionsFunc {
	return func(o *Options) {
		o.level = level
	}
}

// WithAddSource enables or disables source code location in log entries.
func WithAddSource(addSource bool) OptionsFunc {
	return func(o *Options) {
		o.addSource = addSource
	}
}

// WithOutput sets the output writer for the logger.
// If nil, logs will be written to stdout.
func WithOutput(w io.Writer) OptionsFunc {
	return func(o *Options) {
		if w == nil {
			o.output = os.Stdout
			return
		}
		o.output = w
	}
}

// WithContextExtractor adds a context extractor function for the specified key.
// The extractor will be called for each log entry to extract values from the context.
func WithContextExtractor(key string, extractor ContextExtractor) OptionsFunc {
	return func(o *Options) {
		o.contextExtractors[key] = extractor
	}
}

// NewOptions creates a new Options instance with the provided option functions applied.
func NewOptions(options ...OptionsFunc) *Options {
	opts := defaultOptions()
	for _, option := range options {
		option(opts)
	}
	return opts
}
