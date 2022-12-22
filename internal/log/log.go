package log

import (
	"context"
	"io"

	"golang.org/x/exp/slog"
)

const (
	LevelDebug = int(slog.LevelDebug)
	LevelInfo  = int(slog.LevelInfo)
	LevelWarn  = int(slog.LevelWarn)
	LevelErr   = int(slog.LevelError)

	JSONOutput = 1
	TextOutput = 2
)

func NewContext(ctx context.Context, level int, format int, source bool, w io.Writer) context.Context {
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

func With(ctx context.Context, args ...any) context.Context {
	return slog.NewContext(ctx, slog.FromContext(ctx).With(args...))
}

func Debug(ctx context.Context, msg string, args ...any) {
	slog.FromContext(ctx).Info(msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slog.FromContext(ctx).Info(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	slog.FromContext(ctx).Warn(msg, args...)
}

func Error(ctx context.Context, msg string, err error, args ...any) {
	slog.FromContext(ctx).Error(msg, err, args...)
}
