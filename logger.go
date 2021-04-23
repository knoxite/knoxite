/*
 * knoxite
 *     Copyright (c) 2021, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package knoxite

// Logger is used for levelled logging and verbose flag.
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
}

var (
	log Logger = NopLogger{}
)

func SetLogger(l Logger) {
	log = l
}

// NopLogger will be used by default if no logger has been set via SetLogger().
type NopLogger struct {
}

func (nl NopLogger) Warn(v ...interface{}) {}

func (nl NopLogger) Warnf(format string, v ...interface{}) {}

func (nl NopLogger) Info(v ...interface{}) {}

func (nl NopLogger) Infof(format string, v ...interface{}) {}

func (nl NopLogger) Debug(v ...interface{}) {}

func (nl NopLogger) Debugf(format string, v ...interface{}) {}

func (nl NopLogger) Fatal(v ...interface{}) {}

func (nl NopLogger) Fatalf(format string, v ...interface{}) {}
