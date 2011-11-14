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

//
// This package implements Go bindings for the 0mq C API.
//
// It does not attempt to expose zmq_msg_t at all. Instead, Recv() and Send()
// both operate on byte slices, allocating and freeing the memory
// automatically. Currently this requires copying to/from C malloced memory,
// but a future implementation may be able to avoid this to a certain extent.
//
package gozmq

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lzmq
#include <zmq.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"errors"
	"os"
	"unsafe"
)

type Context interface {
	NewSocket(t SocketType) (Socket, error)
	Close()
}

type Socket interface {
	Bind(address string) error
	Connect(address string) error
	Send(data []byte, flags SendRecvOption) error
	Recv(flags SendRecvOption) (data []byte, err error)
	RecvMultipart(flags SendRecvOption) (parts [][]byte, err error)
	SendMultipart(parts [][]byte, flags SendRecvOption) (err error)
	Close() error

	SetSockOptInt64(option Int64SocketOption, value int64) error
	SetSockOptUInt64(option UInt64SocketOption, value uint64) error
	SetSockOptString(option StringSocketOption, value string) error
	GetSockOptInt64(option Int64SocketOption) (value int64, err error)
	GetSockOptUInt64(option UInt64SocketOption) (value uint64, err error)
	GetSockOptString(option StringSocketOption) (value string, err error)

	// Package local function makes this interface unimplementable outside
	// of this package which removes some of the point of using an interface
	apiSocket() unsafe.Pointer
}

type SocketType int
type Int64SocketOption int
type UInt64SocketOption int
type StringSocketOption int
type SendRecvOption int
type zmqErrno os.Errno

const (
	// NewSocket types
	PAIR = SocketType(C.ZMQ_PAIR)
	PUB  = SocketType(C.ZMQ_PUB)
	SUB  = SocketType(C.ZMQ_SUB)
	REQ  = SocketType(C.ZMQ_REQ)
	REP  = SocketType(C.ZMQ_REP)
	XREQ = SocketType(C.ZMQ_XREQ)
	XREP = SocketType(C.ZMQ_XREP)
	PULL = SocketType(C.ZMQ_PULL)
	PUSH = SocketType(C.ZMQ_PUSH)

	// NewSocket options
	HWM          = UInt64SocketOption(C.ZMQ_HWM)
	SWAP         = Int64SocketOption(C.ZMQ_SWAP)
	AFFINITY     = UInt64SocketOption(C.ZMQ_AFFINITY)
	IDENTITY     = StringSocketOption(C.ZMQ_IDENTITY)
	SUBSCRIBE    = StringSocketOption(C.ZMQ_SUBSCRIBE)
	UNSUBSCRIBE  = StringSocketOption(C.ZMQ_UNSUBSCRIBE)
	RATE         = Int64SocketOption(C.ZMQ_RATE)
	RECOVERY_IVL = Int64SocketOption(C.ZMQ_RECOVERY_IVL)
	MCAST_LOOP   = Int64SocketOption(C.ZMQ_MCAST_LOOP)
	SNDBUF       = UInt64SocketOption(C.ZMQ_SNDBUF)
	RCVBUF       = UInt64SocketOption(C.ZMQ_RCVBUF)

	RCVMORE = UInt64SocketOption(C.ZMQ_RCVMORE)

	// Send/recv options
	NOBLOCK = SendRecvOption(C.ZMQ_NOBLOCK)
	SNDMORE = SendRecvOption(C.ZMQ_SNDMORE)

	// Additional ZMQ errors
	EFSM           = zmqErrno(C.EFSM)
	ENOCOMPATPROTO = zmqErrno(C.ENOCOMPATPROTO)
	ETERM          = zmqErrno(C.ETERM)
	EMTHREAD       = zmqErrno(C.EMTHREAD)
)

type PollEvents C.short

const (
	POLLIN  = PollEvents(C.ZMQ_POLLIN)
	POLLOUT = PollEvents(C.ZMQ_POLLOUT)
	POLLERR = PollEvents(C.ZMQ_POLLERR)
)

type DeviceType int

const (
	STREAMER  = DeviceType(C.ZMQ_STREAMER)
	FORWARDER = DeviceType(C.ZMQ_FORWARDER)
	QUEUE     = DeviceType(C.ZMQ_QUEUE)
)

// void zmq_version (int *major, int *minor, int *patch);
func Version() (int, int, int) {
	var major, minor, patch C.int
	C.zmq_version(&major, &minor, &patch)
	return int(major), int(minor), int(patch)
}

func (e zmqErrno) String() string {
	return C.GoString(C.zmq_strerror(C.int(e)))
}

func (e zmqErrno) Error() string {
	return C.GoString(C.zmq_strerror(C.int(e)))
}

// int zmq_errno ();
func errno() error {
	return zmqErrno(C.zmq_errno())
}

/*
 * A context handles socket creation and asynchronous message delivery.
 * There should generally be one context per application.
 */
type zmqContext struct {
	c unsafe.Pointer
}

// Create a new context.
// void *zmq_init (int io_threads);
func NewContext() (Context, error) {
	// TODO Pass something useful here. Number of cores?
	// C.NULL is correct but causes a runtime failure on darwin at present
	if c := C.zmq_init(1); c != nil /*C.NULL*/ {
		return &zmqContext{c}, nil
	}
	return nil, errno()
}

// int zmq_term (void *context);
func (c *zmqContext) destroy() {
	// Will this get called without being added by runtime.SetFinalizer()?
	c.Close()
}

func (c *zmqContext) Close() {
	C.zmq_term(c.c)
}

// Create a new socket.
// void *zmq_socket (void *context, int type);
func (c *zmqContext) NewSocket(t SocketType) (Socket, error) {
	// C.NULL is correct but causes a runtime failure on darwin at present
	if s := C.zmq_socket(c.c, C.int(t)); s != nil /*C.NULL*/ {
		return &zmqSocket{c: c, s: s}, nil
	}
	return nil, errno()
}

type zmqSocket struct {
	// XXX Ensure the zmq context doesn't get destroyed underneath us.
	c *zmqContext
	s unsafe.Pointer
}

// Shutdown the socket.
// int zmq_close (void *s);
func (s *zmqSocket) Close() error {
	if C.zmq_close(s.s) != 0 {
		return errno()
	}
	s.c = nil
	return nil
}

func (s *zmqSocket) destroy() {
	// Will this get called without being added by runtime.SetFinalizer()?
	if err := s.Close(); err != nil {
		panic("Error while destroying zmqSocket: " + err.Error() + "\n")
	}
}

// Set an int64 option on the socket.
// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *zmqSocket) SetSockOptInt64(option Int64SocketOption, value int64) error {
	if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(&value))) != 0 {
		return errno()
	}
	return nil
}

// Set a uint64 option on the socket.
// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *zmqSocket) SetSockOptUInt64(option UInt64SocketOption, value uint64) error {
	if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(&value))) != 0 {
		return errno()
	}
	return nil
}

// Set a string option on the socket.
// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *zmqSocket) SetSockOptString(option StringSocketOption, value string) error {
	v := C.CString(value)
	defer C.free(unsafe.Pointer(v))
	if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(v), C.size_t(len(value))) != 0 {
		return errno()
	}
	return nil
}

// Get an int64 option from the socket.
// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *zmqSocket) GetSockOptInt64(option Int64SocketOption) (value int64, err error) {
	size := C.size_t(unsafe.Sizeof(value))
	if C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&value), &size) != 0 {
		err = errno()
		return
	}
	return
}

// Get a uint64 option from the socket.
// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *zmqSocket) GetSockOptUInt64(option UInt64SocketOption) (value uint64, err error) {
	size := C.size_t(unsafe.Sizeof(value))
	if C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&value), &size) != 0 {
		err = errno()
		return
	}
	return
}

// Get a string option from the socket.
// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *zmqSocket) GetSockOptString(option StringSocketOption) (value string, err error) {
	var buffer [1024]byte
	var size C.size_t = 1024
	if C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&buffer), &size) != 0 {
		err = errno()
		return
	}
	value = string(buffer[:size])
	return
}

// Bind the socket to a listening address.
// int zmq_bind (void *s, const char *addr);
func (s *zmqSocket) Bind(address string) error {
	a := C.CString(address)
	defer C.free(unsafe.Pointer(a))
	if C.zmq_bind(s.s, a) != 0 {
		return errno()
	}
	return nil
}

// Connect the socket to an address.
// int zmq_connect (void *s, const char *addr);
func (s *zmqSocket) Connect(address string) error {
	a := C.CString(address)
	defer C.free(unsafe.Pointer(a))
	if C.zmq_connect(s.s, a) != 0 {
		return errno()
	}
	return nil
}

// Send a message to the socket.
// int zmq_send (void *s, zmq_msg_t *msg, int flags);
func (s *zmqSocket) Send(data []byte, flags SendRecvOption) error {
	var m C.zmq_msg_t
	// Copy data array into C-allocated buffer.
	size := C.size_t(len(data))

	if C.zmq_msg_init_size(&m, size) != 0 {
		return errno()
	}

	if size > 0 {
		// FIXME Ideally this wouldn't require a copy.
		C.memcpy(unsafe.Pointer(C.zmq_msg_data(&m)), unsafe.Pointer(&data[0]), size) // XXX I hope this works...(seems to)
	}

	if C.zmq_send(s.s, &m, C.int(flags)) != 0 {
		// zmq_send did not take ownership, free message
		C.zmq_msg_close(&m)
		return errno()
	}
	return nil
}

// Receive a message from the socket.
// int zmq_recv (void *s, zmq_msg_t *msg, int flags);
func (s *zmqSocket) Recv(flags SendRecvOption) (data []byte, err error) {
	// Allocate and initialise a new zmq_msg_t
	var m C.zmq_msg_t
	if C.zmq_msg_init(&m) != 0 {
		err = errno()
		return
	}
	defer C.zmq_msg_close(&m)
	// Receive into message
	if C.zmq_recv(s.s, &m, C.int(flags)) != 0 {
		err = errno()
		return
	}
	// Copy message data into a byte array
	// FIXME Ideally this wouldn't require a copy.
	size := C.zmq_msg_size(&m)
	if size > 0 {
		data = make([]byte, int(size))
		C.memcpy(unsafe.Pointer(&data[0]), C.zmq_msg_data(&m), size)
	} else {
		data = nil
	}
	return
}

// Send a multipart message.
func (s *zmqSocket) SendMultipart(parts [][]byte, flags SendRecvOption) (err error) {
	for i := 0; i < len(parts)-1; i++ {
		if err = s.Send(parts[i], SNDMORE|flags); err != nil {
			return
		}
	}
	err = s.Send(parts[(len(parts)-1)], flags)
	return
}

// Receive a multipart message.
func (s *zmqSocket) RecvMultipart(flags SendRecvOption) (parts [][]byte, err error) {
	parts = make([][]byte, 0)
	for {
		var data []byte
		var more uint64

		data, err = s.Recv(flags)
		if err != nil {
			return
		}
		parts = append(parts, data)
		more, err = s.GetSockOptUInt64(RCVMORE)
		if err != nil {
			return
		}
		if more == 0 {
			break
		}
	}
	return
}

// return the 
func (s *zmqSocket) apiSocket() unsafe.Pointer {
	return s.s
}

// Item to poll for read/write events on, either a Socket or a file descriptor
type PollItem struct {
	Socket  Socket     // socket to poll for events on 
	Fd      int        // fd to poll for events on as returned from os.File.Fd() 
	Events  PollEvents // event set to poll for
	REvents PollEvents // events that were present
}

// a set of items to poll for events on
type PollItems []PollItem

// Poll ZmqSockets and file descriptors for I/O readiness. Timeout is in
// microseconds.
func Poll(items []PollItem, timeout int64) (count int, err error) {
	zitems := make([]C.zmq_pollitem_t, len(items))
	for i, pi := range items {
		zitems[i].socket = pi.Socket.apiSocket()
		zitems[i].fd = C.int(pi.Fd)
		zitems[i].events = C.short(pi.Events)
	}
	rc := int(C.zmq_poll(&zitems[0], C.int(len(zitems)), C.long(timeout)))
	if rc == -1 {
		return 0, errno()
	}

	for i, zi := range zitems {
		items[i].REvents = PollEvents(zi.revents)
	}

	return rc, nil
}

// run a zmq_device passing messages between in and out
func Device(t DeviceType, in, out Socket) error {
	if C.zmq_device(C.int(t), in.apiSocket(), out.apiSocket()) != 0 {
		return errno()
	}
	return errors.New("zmq_device() returned unexpectedly.")
}

// XXX For now, this library abstracts zmq_msg_t out of the API.
// int zmq_msg_init (zmq_msg_t *msg);
// int zmq_msg_init_size (zmq_msg_t *msg, size_t size);
// int zmq_msg_close (zmq_msg_t *msg);
// size_t zmq_msg_size (zmq_msg_t *msg);
// void *zmq_msg_data (zmq_msg_t *msg);
// int zmq_msg_copy (zmq_msg_t *dest, zmq_msg_t *src);
// int zmq_msg_move (zmq_msg_t *dest, zmq_msg_t *src);
