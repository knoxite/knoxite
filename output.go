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

type Renderers []Renderer

func (r Renderers) Render(p Progress) error {
	for _, renderer := range r {
		err := renderer.Render(p)
		if err != nil {
			return err
		}
	}
	return nil
}

type Output interface {
	Render(p chan Progress, cancel shutdown.Notifier) error
}

type DefaultOutput struct {
	Renderers Renderers
}

func (o DefaultOutput) Render(p chan Progress, cancel shutdown.Notifier) error {
	for v := range p {
		select {
		case n := <-cancel:
			fmt.Println("Aborting rendering...")
			close(n)
			return nil
		default:
			err := o.Renderers.Render(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
