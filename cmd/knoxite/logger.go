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
	LogLevel knoxite.LogLevel
	w        io.Writer
}

func NewLogger(l knoxite.LogLevel) *Logger {
	return &Logger{
		LogLevel: l,
		w:        os.Stdout,
	}
}

func (l *Logger) WithWriter(w io.Writer) *Logger {
	l.w = w
	return l
}

func (l Logger) Fatal(v ...interface{}) {
	l.log(knoxite.LogLevelFatal, v...)
}

func (l Logger) Fatalf(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelFatal, format, v...)
}

func (l Logger) Warn(v ...interface{}) {
	l.log(knoxite.LogLevelWarning, v...)
}

func (l Logger) Warnf(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelWarning, format, v...)
}

func (l Logger) Print(v ...interface{}) {
	l.log(knoxite.LogLevelPrint, v...)
}

func (l Logger) Printf(format string, v ...interface{}) {
	l.logf(knoxite.LogLevelPrint, format, v...)
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

func (l Logger) log(logLevel knoxite.LogLevel, v ...interface{}) {
	if logLevel <= l.LogLevel {
		l.printV(logLevel, v...)
	}
}

func (l Logger) logf(logLevel knoxite.LogLevel, format string, v ...interface{}) {
	if logLevel <= l.LogLevel {
		l.printV(logLevel, fmt.Sprintf(format, v...))
	}
}

func (l Logger) printV(logLevel knoxite.LogLevel, v ...interface{}) {
	if logLevel != knoxite.LogLevelPrint {
		_, _ = l.w.Write([]byte(logLevel.String() + ": "))
	}
	_, _ = l.w.Write([]byte(fmt.Sprint(v...)))
	_, _ = l.w.Write([]byte("\n"))
}
