/*
 * knoxite
 *     Copyright (c) 2020-2021, Matthias Hartmann <mahartma@mahartma.com>
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/knoxite/knoxite"
)

type Logger struct {
	VerbosityLevel knoxite.Verbosity
	w              io.Writer
}

func NewLogger(v knoxite.Verbosity) *Logger {
	return &Logger{
		VerbosityLevel: v,
		w:              os.Stdout,
	}
}

func (l *Logger) WithWriter(w io.Writer) *Logger {
	l.w = w
	return l
}

func (l Logger) Warn(v ...interface{}) {
	l.log(knoxite.LogLevelWarning, v...)
}

func (l Logger) Warnf(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelWarning, format, v...)
}

func (l Logger) Info(v ...interface{}) {
	l.log(knoxite.LogLevelInfo, v...)
}

func (l Logger) Infof(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelInfo, format, v...)
}

func (l Logger) Debug(v ...interface{}) {
	l.log(knoxite.LogLevelDebug, v...)
}

func (l Logger) Debugf(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelDebug, format, v...)
}

func (l Logger) Fatal(v ...interface{}) {
	l.log(knoxite.LogLevelFatal, v...)
	os.Exit(1)
}

func (l Logger) Fatalf(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelFatal, format, v...)
	os.Exit(1)
}

func (l Logger) log(verbosity knoxite.Verbosity, v ...interface{}) {
	if verbosity <= l.VerbosityLevel {
		l.printV(verbosity, v...)
	}
}

func (l Logger) logf(verbosity knoxite.Verbosity, format string, v ...interface{}) {
	if verbosity <= l.VerbosityLevel {
		l.printV(verbosity, fmt.Sprintf(format, v...))
	}
}

func (l Logger) printV(verbosity knoxite.Verbosity, v ...interface{}) {
	_, _ = l.w.Write([]byte(verbosity.String() + ": "))
	_, _ = l.w.Write([]byte(fmt.Sprint(v...)))
	_, _ = l.w.Write([]byte("\n"))
}
