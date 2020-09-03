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

// Renderers is a type to easily operate on the array of renderers.
type Renderers []Renderer

func (r Renderers) Init() error {
	for _, renderer := range r {
		if err := renderer.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderers) Render(p Progress) error {
	for _, renderer := range r {
		if err := renderer.Render(p); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderers) Close() error {
	fmt.Println()
	for _, renderer := range r {
		if err := renderer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// The composite with renderers as leafs is about to be decorated
// for the different UIs.
type Output interface {
	Init() error
	Render(p chan Progress, cancel shutdown.Notifier) error
	Close() error
}
