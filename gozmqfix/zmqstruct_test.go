// Copyright 2013 Joshua Tacoma.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(zmqstructTests, zmqstruct)
}

var zmqstructTests = []testCase{
	{
		Name: "zmqstruct.0",
		In: `package main

import zmq "github.com/alecthomas/gozmq"

type M struct {
	c  zmq.Context
	s  zmq.Socket
	ss []zmq.Socket
	m  map[zmq.Context]zmq.Socket
}

type S0 zmq.Socket

func newM(c zmq.Context, s zmq.Socket, ss ...zmq.Socket) *M {
	if s == nil {
		s = c.NewSocket(zmq.PUB)
	}
	return &M{
		c:  c,
		s:  s,
		ss: ss,
	}
}

var GlobalM = newM(c.NewContext(), nil.(zmq.Socket))

type Socket zmq.Socket

var S Socket = Socket(M.s)
`,
		Out: `package main

import zmq "github.com/alecthomas/gozmq"

type M struct {
	c  *zmq.Context
	s  *zmq.Socket
	ss []*zmq.Socket
	m  map[*zmq.Context]*zmq.Socket
}

type S0 *zmq.Socket

func newM(c *zmq.Context, s *zmq.Socket, ss ...*zmq.Socket) *M {
	if s == nil {
		s = c.NewSocket(zmq.PUB)
	}
	return &M{
		c:  c,
		s:  s,
		ss: ss,
	}
}

var GlobalM = newM(c.NewContext(), nil.(*zmq.Socket))

type Socket *zmq.Socket

var S Socket = Socket(M.s)
`,
	},
	{
		Name: "zmqstruct.1",
		In: `package main

import "github.com/alecthomas/gozmq"

type Socket *gozmq.Socket
`,
		Out: `package main

import "github.com/alecthomas/gozmq"

type Socket *gozmq.Socket
`,
	},
}
