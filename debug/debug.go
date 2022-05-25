package debug

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Debugging interface {
	Debugf(ctx context.Context, format string, v ...interface{})
	Printf(header string, format string, v ...interface{})
}

var debugger Debugging

func SetDebugger(d Debugging) {
	debugger = d
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	if debugger == nil {
		return
	}

	debugger.Debugf(ctx, format, v...)
}

func Printf(header string, format string, v ...interface{}) {
	if debugger == nil {
		return
	}

	debugger.Printf(header, format, v...)
}

type Debugger struct {
	*log.Logger
}

func FromPath(path string) Debugging {
	file, err := os.Create(fmt.Sprintf("%s", path))
	if err != nil {
		panic(err)
	}

	return NewDebugger(log.New(io.MultiWriter(os.Stdout, file), "", log.Ltime))
}

func NewDebugger(logger *log.Logger) Debugging {
	return &Debugger{
		Logger: logger,
	}
}

func (l *Debugger) Debugf(ctx context.Context, format string, v ...interface{}) {
	_ = l.Logger.Output(2, fmt.Sprintf("%s | %s | %s", GetCtxID(ctx), GetCtxElapsedTime(ctx), fmt.Sprintf(format, v...)))
}

func (l *Debugger) Printf(header string, format string, v ...interface{}) {
	_ = l.Logger.Output(2, fmt.Sprintf("%s | %s", header, fmt.Sprintf(format, v...)))
}

type NoopDebugger struct{}

func (l *NoopDebugger) Debugf(ctx context.Context, format string, v ...interface{}) {
	return
}

func (l *NoopDebugger) Printf(header string, format string, v ...interface{}) {
	return
}

func Padding(s string, padSize int) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, fmt.Sprintf("%s%s", strings.Repeat(" ", padSize), line))
	}
	return strings.Join(lines, "\n")
}
