package log

import (
	"context"
	"io"

	"golang.org/x/exp/slog"
)

const (
	LevelDebug = int(slog.LevelDebug) // LevelDebug debug level
	LevelInfo  = int(slog.LevelInfo)  // LevelInfo info level
	LevelWarn  = int(slog.LevelWarn)  // LevelWarn warning level
	LevelErr   = int(slog.LevelError) // LevelErr error level

	JSONOutput = 1 // JSONOutput Log output will be json format
	TextOutput = 2 // TextOutput log output will be text format
)

// NewContext returns a context with an injected logger.
func NewContext(ctx context.Context, level, format int, source bool, w io.Writer) context.Context {
	l := slog.LevelVar{}
	l.Set(slog.Level(level))

	opts := slog.HandlerOptions{
		AddSource: source,
		Level:     &l,
	}
	if format == JSONOutput {
		return slog.NewContext(ctx, slog.New(opts.NewJSONHandler(w)))
	}
	return slog.NewContext(ctx, slog.New(opts.NewTextHandler(w)))
}

// CopyFromContext is a helper function that extracts returns a new context from dest, adding
// the log included in orig.
func CopyFromContext(orig, dest context.Context) context.Context {
	return slog.NewContext(dest, slog.FromContext(orig))
}

// With changes the context logger with a new logger that will include  the extra attributes
// from args parameters.
func With(ctx context.Context, args ...any) context.Context {
	return slog.NewContext(ctx, slog.FromContext(ctx).With(args...))
}

// Debug logs a debug message  using context logger
func Debug(ctx context.Context, msg string, args ...any) {
	slog.FromContext(ctx).Info(msg, args...)
}

// Info logs an info using context logger
func Info(ctx context.Context, msg string, args ...any) {
	slog.FromContext(ctx).Info(msg, args...)
}

// Warn logs a warning using context logger
func Warn(ctx context.Context, msg string, args ...any) {
	slog.FromContext(ctx).Warn(msg, args...)
}

// Error logs an error using context logger
func Error(ctx context.Context, msg string, err error, args ...any) {
	slog.FromContext(ctx).Error(msg, err, args...)
}
