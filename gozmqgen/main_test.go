package main

import (
	"testing"
)

// This test depends on a remote HTTP service that may not even be available.
func TestLoadManual(t *testing.T) {
	if m, err := LoadManual("3.2", "zmq-setsockopt"); err != nil {
		t.Errorf(err.Error())
	} else if m == nil {
		t.Errorf("no error yet nil reader.")
	}
}
