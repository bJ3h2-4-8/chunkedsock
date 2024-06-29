package logging

import (
	"fmt"
	"log/slog"
)

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

var DefaultLogger Logger = &slogbridge{*slog.Default()}

type slogbridge struct {
	slog.Logger
}

func (s *slogbridge) Infof(format string, args ...any) {
	s.Info(fmt.Sprintf(format, args...))
}

func (s *slogbridge) Errorf(format string, args ...any) {
	s.Error(fmt.Sprintf(format, args...))
}

func (s *slogbridge) Fatalf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}
