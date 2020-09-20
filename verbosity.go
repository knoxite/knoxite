/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *     Copyright (c) 2020, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

type Verbosity int

const (
	LogLevelWarning = iota
	LogLevelInfo
	LogLevelDebug
	LogLevelFatal
)

func (v Verbosity) String() string {
	return [...]string{"Warning", "Info", "Debug", "Fatal"}[v]
}
