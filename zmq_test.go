/*
  Copyright 2010 Alec Thomas

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/
package gozmq

import (
	"log"
	"runtime"
	"testing"
	"time"
)

const ADDRESS1 = "tcp://127.0.0.1:23456"
const ADDRESS2 = "tcp://127.0.0.1:23457"
const ADDRESS3 = "tcp://127.0.0.1:23458"

// Addresses for the device test. These cannot be reused since the device 
// will keep running after the test terminates
const ADDR_DEV_IN = "tcp://127.0.0.1:24111"
const ADDR_DEV_OUT = "tcp://127.0.0.1:24112"

// a process local address
const ADDRESS_INPROC = "inproc://test"

const SERVER_READY = "SERVER READY"

func runServer(t *testing.T, c Context, callback func(s Socket)) chan bool {
	finished := make(chan bool)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		s, _ := c.NewSocket(REP)
		defer s.Close()
		if rc := s.Bind(ADDRESS1); rc != nil {
			t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.Error())
		}
		callback(s)
		finished <- true
	}()
	return finished
}

func runPollServer(t *testing.T) (done, bound chan bool) {
	done = make(chan bool)
	bound = make(chan bool)
	go func() {
		te := NewTestEnv(t)
		defer te.Close()
		s1 := te.NewBoundSocket(REP, ADDRESS1)
		s2 := te.NewBoundSocket(REP, ADDRESS2)
		s3 := te.NewBoundSocket(REP, ADDRESS3)

		pi := PollItems{
			PollItem{Socket: s1, Events: POLLIN},
			PollItem{Socket: s2, Events: POLLIN},
			PollItem{Socket: s3, Events: POLLIN},
		}
		bound <- true

		sent := 0
		for {
			_, err := Poll(pi, -1)
			if err != nil {
				done <- false
				return
			}

			switch {
			case pi[0].REvents&POLLIN != 0:
				pi[0].Socket.Recv(0) // eat the incoming message
				pi[0].Socket.Send(nil, 0)
				sent++
			case pi[1].REvents&POLLIN != 0:
				pi[1].Socket.Recv(0) // eat the incoming message
				pi[1].Socket.Send(nil, 0)
				sent++
			case pi[2].REvents&POLLIN != 0:
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
	c, _ := NewContext()
	c.Close()
	c, _ = NewContext()
	c.Close()
}

func TestBindToLoopBack(t *testing.T) {
	c, _ := NewContext()
	defer c.Close()
	s, _ := c.NewSocket(REP)
	defer s.Close()
	if rc := s.Bind(ADDRESS1); rc != nil {
		t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.Error())
	}
}

func TestSetSockOptString(t *testing.T) {
	c, _ := NewContext()
	defer c.Close()
	s, _ := c.NewSocket(SUB)
	defer s.Close()
	if rc := s.Bind(ADDRESS1); rc != nil {
		t.Errorf("Failed to bind to %s; %s", ADDRESS1, rc.Error())
	}
	if rc := s.SetSockOptString(SUBSCRIBE, "TEST"); rc != nil {
		t.Errorf("Failed to subscribe; %v", rc)
	}
}

func TestMultipart(t *testing.T) {
	c, _ := NewContext()
	defer c.Close()
	finished := runServer(t, c, func(s Socket) {
		parts, rc := s.RecvMultipart(0)
		if rc != nil {
			t.Errorf("Failed to receive multipart message; %s", rc.Error())
		}
		if len(parts) != 2 {
			t.Errorf("Invalid multipart message, not enough parts; %d", len(parts))
		}
		if string(parts[0]) != "part1" || string(parts[1]) != "part2" {
			t.Errorf("Invalid multipart message.")
		}
	})

	s, _ := c.NewSocket(REQ)
	defer s.Close()
	if rc := s.Connect(ADDRESS1); rc != nil {
		t.Errorf("Failed to connect to %s; %s", ADDRESS1, rc.Error())
	}
	if rc := s.SendMultipart([][]byte{[]byte("part1"), []byte("part2")}, 0); rc != nil {
		t.Errorf("Failed to send multipart message; %s", rc.Error())
	}
	<-finished
}

func TestPoll(t *testing.T) {
	te := NewTestEnv(t)
	defer te.Close()
	finished, bound := runPollServer(t)

	// wait for sockets to bind
	<-bound

	for _, addr := range []string{ADDRESS2, ADDRESS3, ADDRESS1} {
		sock := te.NewConnectedSocket(REQ, addr)
		te.Send(sock, []byte("request data"), 0)
		te.Recv(sock, 0)
	}

	<-finished

}

func TestDevice(t *testing.T) {
	go func() {
		// the device will never exit so this goroutine will never terminate
		te := NewTestEnv(t)
		defer te.Close()
		in := te.NewBoundSocket(PULL, ADDR_DEV_IN)
		out := te.NewBoundSocket(PUSH, ADDR_DEV_OUT)
		err := Device(STREAMER, in, out)

		// Should never get to here
		t.Error("Device() failed: ", err)
	}()

	te := NewTestEnv(t)
	defer te.Close()
	out := te.NewConnectedSocket(PUSH, ADDR_DEV_IN)
	in := te.NewConnectedSocket(PULL, ADDR_DEV_OUT)

	time.Sleep(1e8)

	te.Send(out, nil, 0)
	te.Recv(in, 0)
}

func TestZmqErrorStr(t *testing.T) {
	var e error = EFSM
	es := e.Error()
	if es != "Operation cannot be accomplished in current state" {
		t.Errorf("EFSM.String() returned unexpected result: %s", e)
	}
}

// expensive test - send a huge amount of data. should be enough to
// trash a current machine if Send or Recv are leaking.
/*
func TestMessageMemory(t *testing.T) {
	// primarily to see if Send or Recv are leaking memory

	const MSG_SIZE = 1e6
	const MSG_COUNT = 100 * 1000

	te := NewTestEnv(nil)
	defer te.Close()

	data := make([]byte, MSG_SIZE)

	out := te.NewBoundSocket(PUSH, ADDRESS1)
	in := te.NewConnectedSocket(PULL, ADDRESS1)

	for i := 0; i < MSG_COUNT; i++ {
		te.Send(out, data, 0)
		d2 := te.Recv(in, 0)
		if len(d2) != MSG_SIZE {
			t.Errorf("Bad message size received")
		}
	}
}
*/

func doBenchmarkSendReceive(b *testing.B, size int, addr string) {
	// since this is a benchmark it should probably call
	// this package's api functions directly rather than 
	// using the testEnv wrappers
	b.StopTimer()
	data := make([]byte, size)

	te := NewTestEnv(nil)
	defer te.Close()
	b.StartTimer()

	out := te.NewBoundSocket(PUSH, ADDRESS1)
	in := te.NewConnectedSocket(PULL, ADDRESS1)

	for i := 0; i < b.N; i++ {
		te.Send(out, data, 0)
		d2 := te.Recv(in, 0)
		if len(d2) != size {
			panic("Bad message size received")
		}
	}
}

func BenchmarkSendReceive1Btcp(b *testing.B) {
	doBenchmarkSendReceive(b, 1, ADDRESS1)
}

func BenchmarkSendReceive1KBtcp(b *testing.B) {
	doBenchmarkSendReceive(b, 1e3, ADDRESS1)
}

func BenchmarkSendReceive1MBtcp(b *testing.B) {
	doBenchmarkSendReceive(b, 1e6, ADDRESS1)
}

func BenchmarkSendReceive1Binproc(b *testing.B) {
	doBenchmarkSendReceive(b, 1, ADDRESS_INPROC)
}

func BenchmarkSendReceive1KBinproc(b *testing.B) {
	doBenchmarkSendReceive(b, 1e3, ADDRESS_INPROC)
}

func BenchmarkSendReceive1MBinproc(b *testing.B) {
	doBenchmarkSendReceive(b, 1e6, ADDRESS_INPROC)
}

// A helper to make tests less verbose
type testEnv struct {
	context Context
	sockets []Socket
	t       *testing.T
}

func NewTestEnv(t *testing.T) *testEnv {
	// Encapsulate everything, including (unnecessarily) the context
	// in the same thread.
	runtime.LockOSThread()
	c, err := NewContext()
	if err != nil {
		t.Errorf("failed to create context in testEnv: %v", err)
		t.FailNow()
	}
	return &testEnv{context: c, t: t}
}

func (te *testEnv) NewSocket(t SocketType) Socket {
	s, err := te.context.NewSocket(t)
	if err != nil {
		log.Panicf("Failed to Create socket of type %v: %v", t, err)
	}
	return s
}

func (te *testEnv) NewBoundSocket(t SocketType, bindAddr string) Socket {
	s := te.NewSocket(t)
	if err := s.Bind(bindAddr); err != nil {
		log.Panicf("Failed to connect to %v: %v", bindAddr, err)
	}
	te.pushSocket(s)
	return s
}

func (te *testEnv) NewConnectedSocket(t SocketType, connectAddr string) Socket {
	s := te.NewSocket(t)
	if err := s.Connect(connectAddr); err != nil {
		log.Panicf("Failed to connect to %v: %v", connectAddr, err)
	}
	te.pushSocket(s)
	return s
}

func (te *testEnv) pushSocket(s Socket) {
	te.sockets = append(te.sockets, s)
}

func (te *testEnv) Close() {

	if err := recover(); err != nil {
		te.t.Errorf("failed in testEnv: %v", err)
	}

	for _, s := range te.sockets {
		s.Close()
	}

	if te.context != nil {
		te.context.Close()
	}
	runtime.UnlockOSThread()
}

func (te *testEnv) Send(sock Socket, data []byte, flags SendRecvOption) {
	if err := sock.Send(data, flags); err != nil {
		te.t.Errorf("Send failed")
	}
}

func (te *testEnv) Recv(sock Socket, flags SendRecvOption) []byte {
	data, err := sock.Recv(flags)
	if err != nil {
		te.t.Errorf("Receive failed")
	}
	return data
}

// TODO Test various socket types. UDP, TCP, etc.
// TODO Test NOBLOCK mode.
// TODO Test getting/setting socket options. Probably sufficient to do just one
// int and one string test.

// TODO Test that closing a context underneath a socket behaves "reasonably" (ie. doesnt' crash).
