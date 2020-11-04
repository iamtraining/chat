package logger

import (
	"fmt"
	"io"
)

type Logger interface {
	Log(...interface{})
}

type nilLogger struct{}

type logger struct {
	writer io.Writer
}

func New(w io.Writer) Logger {
	return &logger{writer: w}
}

func (l *logger) Log(args ...interface{}) {
	fmt.Fprint(l.writer, args...)
	fmt.Fprintln(l.writer)
}

func (l *nilLogger) Log(args ...interface{}) {}

func Silent() Logger {
	return &nilLogger{}
}
