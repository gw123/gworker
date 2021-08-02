package amqpool

import "log"

// Default Loggers
var (
	DefaultLogger = &standardLogger{}
	NopLogger     = &nopLogger{}
)

// Logger is interface for logging
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

type standardLogger struct{}

const (
	levelDebug = "DEBUG"
	levelInfo  = "INFO"
	levelWarn  = "WARN"
	levelError = "ERROR"
	levelFatal = "FATAL"
)

func (s *standardLogger) Print(v ...interface{}) {
	log.Print(v...)
}

func (s *standardLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (s *standardLogger) Debug(v ...interface{}) {
	log.Print(s.prependLevel(levelDebug, v)...)
}

func (s *standardLogger) Debugf(format string, v ...interface{}) {
	log.Printf(s.levelFormat(levelDebug, format), v...)
}

func (s *standardLogger) Info(v ...interface{}) {
	log.Print(s.prependLevel(levelInfo, v)...)
}

func (s *standardLogger) Infof(format string, v ...interface{}) {
	log.Printf(s.levelFormat(levelInfo, format), v...)
}

func (s *standardLogger) Warn(v ...interface{}) {
	log.Print(s.prependLevel(levelWarn, v)...)
}

func (s *standardLogger) Warnf(format string, v ...interface{}) {
	log.Printf(s.levelFormat(levelWarn, format), v...)
}

func (s *standardLogger) Error(v ...interface{}) {
	log.Print(s.prependLevel(levelError, v)...)
}

func (s *standardLogger) Errorf(format string, v ...interface{}) {
	log.Printf(s.levelFormat(levelError, format), v...)
}

func (s *standardLogger) Fatal(v ...interface{}) {
	log.Fatal(s.prependLevel(levelFatal, v)...)
}

func (s *standardLogger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(s.levelFormat(levelFatal, format), v...)
}

func (s *standardLogger) prependLevel(level string, v []interface{}) []interface{} {
	return append([]interface{}{"[" + level + "] "}, v...)
}

func (s *standardLogger) levelFormat(level, format string) string {
	return "[" + level + "] " + format
}

type nopLogger struct{}

func (n *nopLogger) Print(v ...interface{}) {
}

func (n *nopLogger) Printf(format string, v ...interface{}) {
}

func (n *nopLogger) Debug(v ...interface{}) {
}

func (n *nopLogger) Debugf(format string, v ...interface{}) {
}

func (n *nopLogger) Info(v ...interface{}) {
}

func (n *nopLogger) Infof(format string, v ...interface{}) {
}

func (n *nopLogger) Warn(v ...interface{}) {
}

func (n *nopLogger) Warnf(format string, v ...interface{}) {
}

func (n *nopLogger) Error(v ...interface{}) {
}

func (n *nopLogger) Errorf(format string, v ...interface{}) {
}

func (n *nopLogger) Fatal(v ...interface{}) {
}

func (n *nopLogger) Fatalf(format string, v ...interface{}) {
}
