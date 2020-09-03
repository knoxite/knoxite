/*
 * knoxite
 *     Copyright (c) 2020, Matthias Hartmann <mahartma@mahartma.com>
 *
 *   For license see LICENSE
 */

package renderers

import (
	"fmt"

	"github.com/knoxite/knoxite"
	"github.com/muesli/goprogressbar"
)

type VerifyRenderer struct {
	DefaultRenderer knoxite.DefaultRenderer
	ProgressBar     goprogressbar.ProgressBar
	LastPath        string
	Errors          *[]error
	DisposeMessage  string
}

func (r *VerifyRenderer) Init() error {
	r.ProgressBar = goprogressbar.ProgressBar{Total: 1000, Width: 40}
	return nil
}

func (r *VerifyRenderer) Render(p knoxite.Progress) error {
	if err := r.DefaultRenderer.Render(p); err != nil {
		return err
	}

	if p.Error != nil {
		fmt.Println()
		errors := append(*r.Errors, p.Error)
		r.Errors = &errors
		return p.Error
	}

	if p.Path != r.LastPath {
		// We have just started restoring a new item
		if len(r.LastPath) > 0 {
			fmt.Println()
		}
		r.LastPath = p.Path
		r.ProgressBar.Text = p.Path
	}

	r.ProgressBar.Total = int64(p.CurrentItemStats.Size)
	r.ProgressBar.Current = int64(p.CurrentItemStats.Transferred)
	r.ProgressBar.PrependText = fmt.Sprintf("%s / %s",
		knoxite.SizeToString(uint64(r.ProgressBar.Current)),
		knoxite.SizeToString(uint64(r.ProgressBar.Total)))

	r.ProgressBar.LazyPrint()

	return nil
}

func (r *VerifyRenderer) Dispose() error {
	fmt.Println(r.DisposeMessage)
	return nil
}
