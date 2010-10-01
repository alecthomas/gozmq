package zmq

import (
  "testing"
  "zmq"
)

func TestVersion(t *testing.T) {
  major, minor, patch := zmq.Version()
  if major == 0 && minor == 0 && patch == 0 {
    t.Errorf("expected non-zero versions")
  }
}

func TestCreateContext(t *testing.T) {
/*  c := zmq.Context()
  c.Socket()*/
}
