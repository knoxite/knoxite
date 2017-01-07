/*
 * knoxite
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import (
	"errors"
	"testing"
	"time"
)

func TestProgressTransferSpeed(t *testing.T) {
	p := Progress{
		Timer: time.Now(),
		CurrentItemStats: Stats{
			Transferred: 1024,
		},
	}

	time.Sleep(1 * time.Second)
	s := p.TransferSpeed()
	if s < 1000 || s > 1024 {
		t.Errorf("Expected %d, got %d", 1024, s)
	}
}

func TestProgressError(t *testing.T) {
	p := newProgressError(errors.New("TestError"))
	if p.Error == nil {
		t.Errorf("Expected error, got %s", p.Error)
	}
}
