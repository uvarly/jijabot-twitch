package toolbox

import (
	"jijabot/internal/config"
	"jijabot/internal/logger"
)

type Logger interface {
	Info(msg string, kv ...any)
	Error(msg string, kv ...any)
}

type ToolBox struct {
	Config config.Config
	Logger Logger
}

func NewToolbox(c config.Config) (*ToolBox, error) {
	var (
		l = logger.NewLogger()
	)

	return &ToolBox{
		Config: c,
		Logger: l,
	}, nil
}
