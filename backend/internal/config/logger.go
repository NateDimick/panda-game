package config

import (
	"context"
	"errors"
	"log/slog"
	"os"
)

func SetLogger() {
	options := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	f, _ := os.Open("panda-server.log")
	stdoutLogger := slog.NewJSONHandler(os.Stdout, &options)
	fileLogger := slog.NewJSONHandler(f, &options)
	handler := &TeeHandler{
		stdout: stdoutLogger,
		file:   fileLogger,
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)
}

// tees output to a file and to stdout
type TeeHandler struct {
	stdout slog.Handler
	file   slog.Handler
}

func (t *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.stdout.Enabled(ctx, level) && t.file.Enabled(ctx, level)
}

func (t *TeeHandler) Handle(ctx context.Context, record slog.Record) error {
	e1 := t.file.Handle(ctx, record)
	e2 := t.stdout.Handle(ctx, record)
	return errors.Join(e1, e2)
}

func (t *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	t2 := new(TeeHandler)
	t2.file = t.file.WithAttrs(attrs)
	t2.stdout = t.stdout.WithAttrs(attrs)
	return t2
}

func (t *TeeHandler) WithGroup(name string) slog.Handler {
	t2 := new(TeeHandler)
	t2.file = t.file.WithGroup(name)
	t2.stdout = t.stdout.WithGroup(name)
	return t2
}
