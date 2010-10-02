package zmq

import (
  "testing"
)


const ADDRESS = "tcp://127.0.0.1:23456"
const SERVER_READY = "SERVER READY"


// Start a goroutine listening to a local socket
func runZmqServer(a string, t SocketType, shutdown chan bool, out chan string) {
  c := Context()
  defer c.Close()
  s := c.Socket(t)
  defer s.Close()
  if rc := s.Bind(a); rc != nil {
    panic("Failed to bind to " + a + "; " + rc.String())
  }
  out <- SERVER_READY
  for {
    data, rc := s.Recv(0)
    if rc != nil {
      panic("Failed to receive packet " + rc.String())
    }
    out <- string(data)
  }
}

func TestVersion(t *testing.T) {
  major, minor, patch := Version()
  // Require at least 2.0.9
  if major > 2 && minor >= 0 && patch >= 9 {
    t.Errorf("expected at least 0mq version 2.0.9")
  }
}

func TestBindToLoopBack(t *testing.T) {
  c := Context()
  defer c.Close()
  s := c.Socket(REP)
  defer s.Close()
  if rc := s.Bind(ADDRESS); rc != nil {
    t.Errorf("Failed to bind to %s; %s", ADDRESS, rc.String())
  }
}

func TestSend(t *testing.T) {
  server := make(chan string)
  shutdown := make(chan bool, 1)
  go runZmqServer(ADDRESS, REP, shutdown, server)
  ready := <-server
  if ready != SERVER_READY {
  }
  shutdown <- true
}

// TODO Test various socket types. UDP, TCP, etc.
// TODO Test NOBLOCK mode.
// TODO Test getting/setting socket options. Probably sufficient to do just one
// int and one string test.
