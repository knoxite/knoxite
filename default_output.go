/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"fmt"

	shutdown "github.com/klauspost/shutdown2"
)

// This is knoxites default output which is to be decorated for different UIs.
type DefaultOutput struct {
	Renderers Renderers
}

func (o *DefaultOutput) Init() error {
	return o.Renderers.Init()
}

func (o DefaultOutput) Render(p chan Progress, cancel shutdown.Notifier) error {
	for v := range p {
		select {
		case n := <-cancel:
			fmt.Println("Aborting rendering...")
			close(n)
			return nil
		default:
			if err := o.Renderers.Render(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *DefaultOutput) Dispose() error {
	return o.Renderers.Dispose()
}
