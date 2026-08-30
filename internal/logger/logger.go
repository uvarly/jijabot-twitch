package logger

import (
	"io"
	"log/slog"
	"os"
)

const (
	FormatJSON format = iota
	FormatText
)

type Logger interface {
	Debug(message string, args ...any)
	Info(message string, args ...any)
	Warn(message string, args ...any)
	Error(message string, args ...any)

	// Usage: logger.With("component", "<component name>")
	With(args ...any) Logger
}

type format int

type settings struct {
	level     slog.Level
	output    io.Writer
	format    format
	addSource bool
}

type Option func(*settings)

func WithLevel(level slog.Level) Option {
	return func(s *settings) {
		s.level = level
	}
}

func WithOutput(output io.Writer) Option {
	return func(s *settings) {
		s.output = output
	}
}

func WithTextFormat() Option {
	return func(s *settings) {
		s.format = FormatText
	}
}

func WithSource() Option {
	return func(s *settings) {
		s.addSource = true
	}
}

type slogLogger struct {
	l *slog.Logger
}

func NewSlogLogger(options ...Option) *slogLogger {
	var slogHandler slog.Handler

	settings := &settings{
		level:  slog.LevelInfo,
		output: os.Stdout,
		format: FormatJSON,
		// addSource: false,
	}

	for _, o := range options {
		o(settings)
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     settings.level,
		AddSource: settings.addSource,
	}

	if settings.format == FormatJSON {
		slogHandler = slog.NewJSONHandler(settings.output, handlerOptions)
	} else {
		slogHandler = slog.NewTextHandler(settings.output, handlerOptions)
	}

	return &slogLogger{
		l: slog.New(slogHandler),
	}
}

func (sl *slogLogger) Debug(m string, args ...any) { sl.l.Debug(m, args...) }

func (sl *slogLogger) Info(m string, args ...any) { sl.l.Info(m, args...) }

func (sl *slogLogger) Warn(m string, args ...any) { sl.l.Warn(m, args...) }

func (sl *slogLogger) Error(m string, args ...any) { sl.l.Error(m, args...) }

func (sl *slogLogger) With(args ...any) Logger {
	return &slogLogger{l: sl.l.With(args...)}
}
