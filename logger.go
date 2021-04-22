/*
 * knoxite
 *     Copyright (c) 2021, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package knoxite

// Logger is used for levelled logging and verbose flag
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
