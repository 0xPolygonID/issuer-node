package log

import (
	"context"
	"io"

	"golang.org/x/exp/slog"
)

type contextKey struct{}

// Log configuration constants
const (
	LevelDebug = int(slog.LevelDebug) // debug level
	LevelInfo  = int(slog.LevelInfo)  // info level
	LevelWarn  = int(slog.LevelWarn)  //  warning level
	LevelErr   = int(slog.LevelError) //  error level

	OutputJSON = 1 // Log output will be json format
	OutputText = 2 //  log output will be text format
)

// NewContext returns a context with an injected logger.
func NewContext(ctx context.Context, level, format int, w io.Writer) context.Context {
	l := slog.LevelVar{}
	l.Set(slog.Level(level))

	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     &l,
	}
	if format == OutputJSON {
		return newContext(ctx, slog.New(opts.NewJSONHandler(w)))
	}
	return newContext(ctx, slog.New(opts.NewTextHandler(w)))
}

// CopyFromContext is a helper function that extracts returns a new context from dest, adding
// the log included in orig.
func CopyFromContext(orig, dest context.Context) context.Context {
	return newContext(dest, fromContext(orig))
}

// With changes the context logger with a new logger that will include  the extra attributes
// from args parameters.
func With(ctx context.Context, args ...any) context.Context {
	return newContext(ctx, fromContext(ctx).With(args...))
}

// Debug logs a debug message  using context logger
func Debug(ctx context.Context, msg string, args ...any) {
	fromContext(ctx).Info(msg, args...)
}

// Info logs an info using context logger
func Info(ctx context.Context, msg string, args ...any) {
	fromContext(ctx).Info(msg, args...)
}

// Warn logs a warning using context logger
func Warn(ctx context.Context, msg string, args ...any) {
	fromContext(ctx).Warn(msg, args...)
}

// Error logs an error using context logger
func Error(ctx context.Context, msg string, err error, args ...any) {
	fromContext(ctx).Error(msg, err, args...)
}

func newContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func fromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
