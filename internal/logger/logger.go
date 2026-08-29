package logger

import (
	"fmt"
	"log"
)

type Logger struct {
	logLevel string
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(m string, kv ...any) {
	m = fmt.Sprintf(m, kv...)
	log.Println("INFO: ", m)
}

func (l *Logger) Error(m string, kv ...any) {
	m = fmt.Sprintf(m, kv...)
	log.Println("ERROR: ", m)
}
