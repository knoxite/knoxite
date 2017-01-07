/*
 * knoxite
 *     Copyright (c) 2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "testing"

func TestBackendURLError(t *testing.T) {
	// Go 1.6 & up only
	// _, err := BackendFromURL("http://a b/")

	_, err := BackendFromURL("http://%test")
	if err == nil {
		t.Errorf("Expected an error, got %v", err)
	}
}
