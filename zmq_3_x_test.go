// +build zmq_3_x

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
	"testing"
	"time"
)

const ADDR_PROXY_IN = "tcp://127.0.0.1:24114"
const ADDR_PROXY_OUT = "tcp://127.0.0.1:24115"
const ADDR_PROXY_CAP = "tcp://127.0.0.1:24116"

func TestProxy(t *testing.T) {
	go func() {
		// the proxy will never exit so this goroutine will never terminate
		te := NewTestEnv(t)
		defer te.Close()
		in := te.NewBoundSocket(ROUTER, ADDR_PROXY_IN)
		out := te.NewBoundSocket(DEALER, ADDR_PROXY_OUT)
		capture := te.NewBoundSocket(PUSH, ADDR_PROXY_CAP)
		err := Proxy(in, out, capture)

		// Should never get to here
		t.Error("Proxy() failed: ", err)
	}()

	te := NewTestEnv(t)
	defer te.Close()
	in := te.NewConnectedSocket(REQ, ADDR_PROXY_IN)
	out := te.NewConnectedSocket(REP, ADDR_PROXY_OUT)
	capture := te.NewConnectedSocket(PULL, ADDR_PROXY_CAP)
	time.Sleep(1e8)
	te.Send(in, nil, 0)
	te.Recv(out, 0)
	te.Recv(capture, 0)
}
