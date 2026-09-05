package logger

import "context"

type noOp struct{}

func NoOp() Logger {
	return &noOp{}
}

func (n *noOp) Debug(m string, args ...any)                             {}
func (n *noOp) Info(m string, args ...any)                              {}
func (n *noOp) Warn(m string, args ...any)                              {}
func (n *noOp) Error(m string, args ...any)                             {}
func (n *noOp) DebugContext(ctx context.Context, m string, args ...any) {}
func (n *noOp) InfoContext(ctx context.Context, m string, args ...any)  {}
func (n *noOp) WarnContext(ctx context.Context, m string, args ...any)  {}
func (n *noOp) ErrorContext(ctx context.Context, m string, args ...any) {}
func (n *noOp) With(args ...any) Logger                                 { return n }
