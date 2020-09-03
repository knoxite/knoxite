/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package knoxite

type Renderer interface {
	Render(p Progress) error
}

type DefaultRenderer struct {
}

func (r DefaultRenderer) Render(p Progress) error {
	// we want to leave this extension point open for later renderers
	return nil
}
