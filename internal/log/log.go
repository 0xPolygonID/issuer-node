// Package log is a wrapper around go >1.21 slog package.
package log

import (
	"context"
	"io"
	"log/slog"
)

// Log configuration constants
const (
	LevelDebug = int(slog.LevelDebug) // debug level
	LevelInfo  = int(slog.LevelInfo)  // info level
	LevelWarn  = int(slog.LevelWarn)  //  warning level
	LevelErr   = int(slog.LevelError) //  error level

	OutputJSON = 1 // Log output will be json format
	OutputText = 2 //  log output will be text format
)

// Config configures the default logger.
func Config(level, format int, w io.Writer) {
	var handler slog.Handler

	l := slog.LevelVar{}
	l.Set(slog.Level(level))

	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     &l,
	}
	handler = slog.NewTextHandler(w, &opts)
	if format == OutputJSON {
		handler = slog.NewJSONHandler(w, &opts)
	}
	slog.SetDefault(slog.New(handler))
}

// With changes the default logger to include the extra attributes
// from args parameters.
func With(args ...any) {
	slog.With(args...)
}

// Debug logs a debug message  using context logger
func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

// Info logs an info using context logger
func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

// Warn logs a warning using context logger
func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

// Error logs an error using context logger
func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}
