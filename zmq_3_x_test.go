// +build zmq_3_x zmq_4_x

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
	"errors"
	"testing"
	"time"
)

const ADDR_PROXY_IN = "tcp://127.0.0.1:24114"
const ADDR_PROXY_OUT = "tcp://127.0.0.1:24115"
const ADDR_PROXY_CAP = "tcp://127.0.0.1:24116"

func TestProxy(t *testing.T) {
	te1, te2 := NewTestEnv(t), NewTestEnv(t)
	exitOk := make(chan bool, 1)
	go func() {
		in := te1.NewBoundSocket(ROUTER, ADDR_PROXY_IN)
		out := te1.NewBoundSocket(DEALER, ADDR_PROXY_OUT)
		capture := te1.NewBoundSocket(PUSH, ADDR_PROXY_CAP)
		err := Proxy(in, out, capture)

		select {
		case <-exitOk:
		default:
			t.Error("Proxy() failed: ", err)
		}
	}()

	in := te2.NewConnectedSocket(REQ, ADDR_PROXY_IN)
	out := te2.NewConnectedSocket(REP, ADDR_PROXY_OUT)
	capture := te2.NewConnectedSocket(PULL, ADDR_PROXY_CAP)
	time.Sleep(1e8)
	te2.Send(in, nil, 0)
	te2.Recv(out, 0)
	te2.Recv(capture, 0)

	te2.Close()
	exitOk <- true
	te1.Close()
}

func TestProxyNoCapture(t *testing.T) {
	te1, te2 := NewTestEnv(t), NewTestEnv(t)
	exitOk := make(chan bool, 1)
	go func() {
		in := te1.NewBoundSocket(ROUTER, ADDR_PROXY_IN)
		out := te1.NewBoundSocket(DEALER, ADDR_PROXY_OUT)
		err := Proxy(in, out, nil)

		select {
		case <-exitOk:
		default:
			t.Error("Proxy() failed: ", err)
		}
	}()

	in := te2.NewConnectedSocket(REQ, ADDR_PROXY_IN)
	out := te2.NewConnectedSocket(REP, ADDR_PROXY_OUT)
	time.Sleep(1e8)
	te2.Send(in, nil, 0)
	te2.Recv(out, 0)

	te2.Close()
	exitOk <- true
	te1.Close()
}

func TestSocket_SetSockOptStringNil(t *testing.T) {
	failed := make(chan bool, 2)
	c, _ := NewContext()
	defer c.Close()
	go func() {
		srv, _ := c.NewSocket(REP)
		defer srv.Close()
		srv.SetSockOptString(TCP_ACCEPT_FILTER, "127.0.0.1")
		srv.SetSockOptString(TCP_ACCEPT_FILTER, "192.0.2.1")
		srv.Bind(ADDRESS1) // 127.0.0.1 and 192.0.2.1 are allowed here.
		// The test will fail if the following line is removed:
		srv.SetSockOptStringNil(TCP_ACCEPT_FILTER)
		srv.SetSockOptString(TCP_ACCEPT_FILTER, "192.0.2.2")
		srv.Bind(ADDRESS2) // Only 192.0.2.1 is allowed here.
		for {
			if _, err := srv.Recv(0); err != nil {
				break
			}
			srv.Send(nil, 0)
		}
	}()
	go func() {
		s2, _ := c.NewSocket(REQ)
		defer s2.Close()
		s2.SetSockOptInt(LINGER, 0)
		s2.Connect(ADDRESS2)
		s2.Send(nil, 0)
		if _, err := s2.Recv(0); err == nil {
			// 127.0.0.1 is supposed to be ignored by ADDRESS2:
			t.Error("SetSockOptStringNil did not clear TCP_ACCEPT_FILTER.")
		}
		failed <- true
	}()
	s1, _ := c.NewSocket(REQ)
	defer s1.Close()
	s1.Connect(ADDRESS1)
	s1.Send(nil, 0)
	s1.Recv(0)
	select {
	case <-failed:
	case <-time.After(50 * time.Millisecond):
	}
}

const (
	TESTMONITOR_ADDR_SINK   = "tcp://127.0.0.1:24117"
	TESTMONITOR_ADDR_EVENTS = "inproc://TestMonitorEvents"
)

func TestMonitor(t *testing.T) {
	te := NewTestEnv(t)
	defer te.Close()

	// Prepare the sink socket.
	out := te.NewSocket(PULL)
	err := out.Bind(TESTMONITOR_ADDR_SINK)
	if err != nil {
		t.Fatal(err)
	}

	// Prepare the source socket, do not connect yet.
	in := te.NewSocket(PUSH)
	defer in.Close()

	// Attach the monitor.
	err = in.Monitor(TESTMONITOR_ADDR_EVENTS,
		EVENT_CONNECTED|EVENT_DISCONNECTED)
	if err != nil {
		out.Close()
		t.Fatal(err)
	}

	monitor := te.NewConnectedSocket(PAIR, TESTMONITOR_ADDR_EVENTS)

	// Connect the client to the server, wait for EVENT_CONNECTED.
	err = in.Connect(TESTMONITOR_ADDR_SINK)
	if err != nil {
		out.Close()
		t.Fatal(err)
	}

	err = waitForEvent(t, monitor)
	if err != nil {
		out.Close()
		t.Fatal(err)
	}

	// Close the sink socket, wait for EVENT_DISCONNECTED.
	err = out.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = waitForEvent(t, monitor)
	if err != nil {
		t.Fatal(err)
	}
}

func waitForEvent(t *testing.T, monitor *Socket) error {
	exit := make(chan error, 1)

	// This goroutine will return either when an event is received
	// or the context is closed.
	go func() {
		// RecvMultipart should work for both zeromq3-x and libzmq API.
		_, ex := monitor.RecvMultipart(0)
		exit <- ex
	}()

	timeout := time.After(time.Second)

	select {
	case err := <-exit:
		return err
	case <-timeout:
		return errors.New("Test timed out")
	}

	return nil
}
