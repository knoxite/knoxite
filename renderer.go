/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package knoxite

// The renderer is the leave in the composite constellation with the output
// type. You can also use an output to group renderers and use it as a renderer.
type Renderer interface {
	Init() error
	Render(p Progress) error
	Dispose() error
}

type DefaultRenderer struct {
}

func (r *DefaultRenderer) Init() error {
	// we want to leave this extension point open for later renderers
	return nil
}

func (r DefaultRenderer) Render(p Progress) error {
	// we want to leave this extension point open for later renderers
	return nil
}

func (r *DefaultRenderer) Dispose() error {
	return nil
}
