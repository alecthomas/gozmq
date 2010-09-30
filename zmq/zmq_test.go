package zmq

import (
  "os"
  "testing"
  "zmq"
)

func TestVersion(t *testing.T) {
  major, minor, patch := zmq.Version()
  if major == 0 && minor == 0 && patch == 0 {
    t.Errorf("expected non-zero versions")
  }
}

func TestErrno(t *testing.T) {
  errno := zmq.Errno()
  if errno != os.Errno(0) {
    t.Errorf("expected zmq.Errno() to be zero")
  }
}

func TestStrError(t *testing.T) {
  error := zmq.Error(1)
  msg := error.String()
  if msg != "Operation not permitted" {
    t.Errorf("expected error.String(1) to return 'Operation not permitted', got '%s' instead", msg)
  }
}
