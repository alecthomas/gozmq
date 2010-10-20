package zmq

import (
	"runtime"
	"testing"
)


const ADDRESS = "tcp://127.0.0.1:23456"
const SERVER_READY = "SERVER READY"

func runServer(t *testing.T, c ZmqContext, callback func(s ZmqSocket)) chan bool {
	finished := make(chan bool)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		s := c.Socket(REP)
		defer s.Close()
		if rc := s.Bind(ADDRESS); rc != nil {
			t.Errorf("Failed to bind to %s; %s", ADDRESS, rc.String())
		}
		callback(s)
		finished <- true
	}()
	return finished
}

func TestVersion(t *testing.T) {
	major, minor, patch := Version()
	// Require at least 2.0.9
	if major > 2 && minor >= 0 && patch >= 9 {
		t.Errorf("expected at least 0mq version 2.0.9")
	}
}

func TestCreateDestroyContext(t *testing.T) {
	c := Context()
	c.Close()
	c = Context()
	c.Close()
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

func TestSetSockOptString(t *testing.T) {
	c := Context()
	defer c.Close()
	s := c.Socket(SUB)
	defer s.Close()
	if rc := s.Bind(ADDRESS); rc != nil {
		t.Errorf("Failed to bind to %s; %s", ADDRESS, rc.String())
	}
	if rc := s.SetSockOptString(SUBSCRIBE, "TEST"); rc != nil {
		t.Errorf("Failed to subscribe; %v", rc)
	}
}

func TestMultipart(t *testing.T) {
	c := Context()
	defer c.Close()
	finished := runServer(t, c, func(s ZmqSocket) {
		parts, rc := s.RecvMultipart(0)
		if rc != nil {
			t.Errorf("Failed to receive multipart message; %s", rc.String())
		}
		if len(parts) != 2 {
			t.Errorf("Invalid multipart message, not enough parts; %d", len(parts))
		}
		if string(parts[0]) != "part1" || string(parts[1]) != "part2" {
			t.Errorf("Invalid multipart message.")
		}
	})

	s := c.Socket(REQ)
	defer s.Close()
	if rc := s.Connect(ADDRESS); rc != nil {
		t.Errorf("Failed to connect to %s; %s", ADDRESS, rc.String())
	}
	if rc := s.SendMultipart([][]byte{[]byte("part1"), []byte("part2")}, 0); rc != nil {
		t.Errorf("Failed to send multipart message; %s", rc.String())
	}
	<-finished
}

// TODO Test various socket types. UDP, TCP, etc.
// TODO Test NOBLOCK mode.
// TODO Test getting/setting socket options. Probably sufficient to do just one
// int and one string test.


// TODO Test that closing a context underneath a socket behaves "reasonably" (ie. doesnt' crash).
