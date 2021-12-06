/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

// log levels for logging.
type LogLevel int

const (
	LogLevelFatal = iota
	LogLevelWarning
	LogLevelPrint
	LogLevelInfo
	LogLevelDebug
)

func (l LogLevel) String() string {
	return [...]string{"Fatal", "Warning", "Print", "Info", "Debug"}[l]
}
