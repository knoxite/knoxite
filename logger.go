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
)

type Logger struct {
	VerbosityLevel Verbosity
}

func (l Logger) printV(verbosity Verbosity, v ...interface{}) {
	fmt.Print(verbosity.String() + ": ")
	fmt.Println(v...)
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
