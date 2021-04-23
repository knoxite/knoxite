/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

// verbosity levels for logging.
type Verbosity int

const (
	LogLevelFatal = iota
	LogLevelWarning
	LogLevelInfo
	LogLevelDebug
)

func (v Verbosity) String() string {
	return [...]string{"Fatal", "Warning", "Info", "Debug"}[v]
}
