package zmq

import (
	"runtime"
	"testing"
)

const ADDRESS1 = "tcp://127.0.0.1:23456"
const ADDRESS2 = "tcp://127.0.0.1:23457"
const ADDRESS3 = "tcp://127.0.0.1:23458"
const SERVER_READY = "SERVER READY"

func runServer(t *testing.T, c ZmqContext, callback func(s ZmqSocket)) chan bool {
	finished := make(chan bool)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		s := c.Socket(REP)
		defer s.Close()
		if rc := s.Bind(ADDRESS1); rc != nil {
			t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.String())
		}
		callback(s)
		finished <- true
	}()
	return finished
}

func runPollServer(t *testing.T, c ZmqContext) (done, bound chan bool) {
	done = make(chan bool)
	bound = make(chan bool)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		s1 := c.Socket(REP)
		defer s1.Close()
		if rc := s1.Bind(ADDRESS1); rc != nil {
			t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.String())
		}

		s2 := c.Socket(REP)
		defer s2.Close()
		if rc := s2.Bind(ADDRESS2); rc != nil {
			t.Errorf("Failed to bind to %s; %s", ADDRESS2, rc.String())
		}

		s3 := c.Socket(REP)
		defer s3.Close()

		if rc := s3.Bind(ADDRESS3); rc != nil {
			t.Errorf("Failed to bind to %s; %s", ADDRESS3, rc.String())
		}

		pi := PollItems{ 	PollItem{ Socket: s1, Events: POLLIN },
							PollItem{ Socket: s2, Events: POLLIN },
							PollItem{ Socket: s3, Events: POLLIN } }
		bound <- true
		
		sent := 0				
		for {
			_, err := Poll(pi, -1)
			if err != nil {
				done <- false
				return
			}
			
			switch {
				case pi[0].REvents & POLLIN != 0:
					pi[0].Socket.Recv(0) // eat the incoming message
					pi[0].Socket.Send(nil, 0)
					sent++
				case pi[1].REvents & POLLIN != 0:
					pi[1].Socket.Recv(0) // eat the incoming message
					pi[1].Socket.Send(nil, 0)
					sent++
				case pi[2].REvents & POLLIN != 0:
					pi[2].Socket.Recv(0) // eat the incoming message
					pi[2].Socket.Send(nil, 0)
					sent++
			}

			if sent == 3 {
				break
			}
		}

		done <- true
	}()
	return
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
	if rc := s.Bind(ADDRESS1); rc != nil {
		t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.String())
	}
}

func TestSetSockOptString(t *testing.T) {
	c := Context()
	defer c.Close()
	s := c.Socket(SUB)
	defer s.Close()
	if rc := s.Bind(ADDRESS1); rc != nil {
		t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.String())
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
	if rc := s.Connect(ADDRESS1); rc != nil {
		t.Errorf("Failed to connect to %s; %s", ADDRESS1, rc.String())
	}
	if rc := s.SendMultipart([][]byte{[]byte("part1"), []byte("part2")}, 0); rc != nil {
		t.Errorf("Failed to send multipart message; %s", rc.String())
	}
	<-finished
}

func TestPoll(t *testing.T) {
	c := Context()
	defer c.Close()
	finished, bound := runPollServer(t, c)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	// wait for sockets to bind
	<-bound
	
	for _, addr := range []string{ADDRESS2, ADDRESS3, ADDRESS1} {
		s := c.Socket(REQ)
		defer s.Close()
		
		if rc := s.Connect(addr); rc != nil {
			t.Errorf("Failed to connect to %s; %s", addr, rc.String())
		}
		if rc := s.Send([]byte("request data"), 0); rc != nil {
			t.Errorf("Failed to send message: %v", rc)
		}
		if _, rc := s.Recv(0); rc != nil {
			t.Errorf("Failed to recv message: %v", rc)
		}
	}

	<-finished
	
}
// TODO Test various socket types. UDP, TCP, etc.
// TODO Test NOBLOCK mode.
// TODO Test getting/setting socket options. Probably sufficient to do just one
// int and one string test.


// TODO Test that closing a context underneath a socket behaves "reasonably" (ie. doesnt' crash).
