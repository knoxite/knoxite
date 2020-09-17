/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"fmt"
	"io"
	"os"
)

type Logger struct {
	VerbosityLevel Verbosity
	w              io.Writer
}

func NewLogger(v Verbosity) *Logger {
	return &Logger{
		VerbosityLevel: v,
		w:              os.Stdout,
	}
}

func (l *Logger) WithWriter(w io.Writer) *Logger {
	l.w = w
	return l
}

func (l Logger) printV(verbosity Verbosity, v ...interface{}) {
	_, _ = l.w.Write([]byte(verbosity.String() + ": "))
	_, _ = l.w.Write([]byte(fmt.Sprint(v...)))
	_, _ = l.w.Write([]byte("\n"))
}

func (l Logger) Warn(v ...interface{}) {
	l.log(LogLevelWarning, v...)
}

func (l Logger) Info(v ...interface{}) {
	l.log(LogLevelInfo, v...)
}

func (l Logger) Debug(v ...interface{}) {
	l.log(LogLevelDebug, v...)
}

func (l Logger) log(verbosity Verbosity, v ...interface{}) {
	if verbosity <= l.VerbosityLevel {
		l.printV(verbosity, v...)
	}
}
