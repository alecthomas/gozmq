/*
 * This package implements Go bindings for the 0mq C API.
 *
 * It does not attempt to expose zmq_msg_t at all. Instead, Recv() and Send()
 * both operate on byte slices, allocating and freeing the memory
 * automatically. Currently this requires copying to/from C malloced memory,
 * but a future implementation may be able to avoid this to a certain extent.
 *
 * Multi-part messages are not supported at all.
 *
 * A note about memory management: it's not entirely clear from the 0mq
 * documentation how memory for zmq_msg_t and packet data is managed once 0mq
 * takes ownership. This module operates under the following (educated)
 * assumptions:
 *
 * - For zmq_msg_t structures references are not held beyond the duration of
 *   any function call.
 * - Packet data is reference counted. The count is incremented when a packet
 *   is queued for delivery to a destination (the inference being that for
 *   delivery to N destinations, the reference count will be incremented N
 *   times) and decremented once the packet has either been delivered or
 *   errored.
 */
package zmq

/*
#include <stdint.h>
#include <zmq.h>
#include <stdlib.h>
#include <string.h>
// FIXME Could not for the life of me figure out how to get the size of a C
// structure from Go. unsafe.Sizeof() didn't work, C.sizeof didn't work, and so
// on.
zmq_msg_t *alloc_zmq_msg_t() {
  zmq_msg_t *msg = (zmq_msg_t*)malloc(sizeof(zmq_msg_t));
  return msg;
}
// Callback for zmq_msg_init_data.
void free_zmq_msg_t_data(void *data, void *hint) {
  free(data);
}
// FIXME This works around always getting the error "must call
// C.free_zmq_msg_t_data" when attempting to reference a C function pointer.
// What the?!
zmq_free_fn *free_zmq_msg_t_data_ptr = free_zmq_msg_t_data;
*/
import "C"

import (
	"os"
	//  "runtime"
	"unsafe"
)

type ZmqContext interface {
	Socket(t SocketType) ZmqSocket
	Close()
}

type ZmqSocket interface {
	Bind(address string) os.Error
	Connect(address string) os.Error
	Send(data []byte, flags SendRecvOption) os.Error
	Recv(flags SendRecvOption) (data []byte, error os.Error)
	Close() os.Error

	SetSockOptInt64(option Int64SocketOption, value int64) os.Error
	SetSockOptUInt64(option UInt64SocketOption, value uint64) os.Error
	SetSockOptString(option StringSocketOption, value string) os.Error
	GetSockOptInt64(option Int64SocketOption) (value int64, error os.Error)
	GetSockOptUInt64(option UInt64SocketOption) (value uint64, error os.Error)
	GetSockOptString(option StringSocketOption) (value string, error os.Error)
}

type SocketType int
type Int64SocketOption int
type UInt64SocketOption int
type StringSocketOption int
type SendRecvOption int

const (
	// Socket types
	PAIR = SocketType(C.ZMQ_PAIR)
	PUB  = SocketType(C.ZMQ_PUB)
	SUB  = SocketType(C.ZMQ_SUB)
	REQ  = SocketType(C.ZMQ_REQ)
	REP  = SocketType(C.ZMQ_REP)
	XREQ = SocketType(C.ZMQ_XREQ)
	XREP = SocketType(C.ZMQ_XREP)
	PULL = SocketType(C.ZMQ_PULL)
	PUSH = SocketType(C.ZMQ_PUSH)

	// Socket options
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
	// Not documented. Probably? related to SNDMORE.
	// RCVMORE = UInt64SocketOption(C.ZMQ_RCVMORE)

	// Send/recv options
	NOBLOCK = SendRecvOption(C.ZMQ_NOBLOCK)
	// This is not supported (yet).
	//SNDMORE = SendRecvOption(C.ZMQ_SNDMORE)
)


/*
 * Misc functions
 */
// TODO int zmq_poll (zmq_pollitem_t *items, int nitems, long timeout);
// TODO int zmq_device (int device, void * insocket, void* outsocket);


// void zmq_version (int *major, int *minor, int *patch);
func Version() (int, int, int) {
	var major, minor, patch C.int
	C.zmq_version(&major, &minor, &patch)
	return int(major), int(minor), int(patch)
}

// int zmq_errno ();
func errno() os.Error {
	return os.Errno(C.zmq_errno())
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
func Context() ZmqContext {
	// TODO Pass something useful here. Number of cores?
	return &zmqContext{C.zmq_init(1)}
}

// int zmq_term (void *context);
func (c *zmqContext) destroy() {
	c.Close()
}

func (c *zmqContext) Close() {
	C.zmq_term(c.c)
}

// Create a new socket.
// void *zmq_socket (void *context, int type);
func (c *zmqContext) Socket(t SocketType) ZmqSocket {
	return &zmqSocket{c: c, s: C.zmq_socket(c.c, C.int(t))}
}

type zmqSocket struct {
	// XXX Ensure the zmq context doesn't get destroyed underneath us.
	c *zmqContext
	s unsafe.Pointer
}

// Shutdown the socket.
// int zmq_close (void *s);
func (s *zmqSocket) Close() os.Error {
	if C.zmq_close(s.s) != 0 {
		return errno()
	}
	s.c = nil
	return nil
}

func (s *zmqSocket) destroy() {
	if error := s.Close(); error != nil {
		panic("Error while destroying zmqSocket: " + error.String() + "\n")
	}
}

// Set an int64 option on the socket.
// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *zmqSocket) SetSockOptInt64(option Int64SocketOption, value int64) os.Error {
	if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(&value))) != 0 {
		return errno()
	}
	return nil
}

// Set a uint64 option on the socket.
// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *zmqSocket) SetSockOptUInt64(option UInt64SocketOption, value uint64) os.Error {
	if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(&value))) != 0 {
		return errno()
	}
	return nil
}

// Set a string option on the socket.
// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *zmqSocket) SetSockOptString(option StringSocketOption, value string) os.Error {
	v := C.CString(value)
	defer C.free(unsafe.Pointer(v))
	if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(v), C.size_t(len(value))) != 0 {
		return errno()
	}
	return nil
}

// Get an int64 option from the socket.
// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *zmqSocket) GetSockOptInt64(option Int64SocketOption) (value int64, error os.Error) {
	size := C.size_t(unsafe.Sizeof(value))
	if C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&value), &size) != 0 {
		error = errno()
		return
	}
	error = nil
	return
}

// Get a uint64 option from the socket.
// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *zmqSocket) GetSockOptUInt64(option UInt64SocketOption) (value uint64, error os.Error) {
	size := C.size_t(unsafe.Sizeof(value))
	if C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&value), &size) != 0 {
		error = errno()
		return
	}
	error = nil
	return
}

// Get a string option from the socket.
// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *zmqSocket) GetSockOptString(option StringSocketOption) (value string, error os.Error) {
	var buffer [1024]byte
	var size C.size_t = 1024
	if C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&buffer), &size) != 0 {
		error = errno()
		return
	}
	value = string(buffer[:size])
	error = nil
	return
}

// Bind the socket to a listening address.
// int zmq_bind (void *s, const char *addr);
func (s *zmqSocket) Bind(address string) os.Error {
	a := C.CString(address)
	defer C.free(unsafe.Pointer(a))
	if C.zmq_bind(s.s, a) != 0 {
		return errno()
	}
	return nil
}

// Connect the socket to an address.
// int zmq_connect (void *s, const char *addr);
func (s *zmqSocket) Connect(address string) os.Error {
	a := C.CString(address)
	defer C.free(unsafe.Pointer(a))
	if C.zmq_connect(s.s, a) != 0 {
		return errno()
	}
	return nil
}

// Send a message to the socket.
// int zmq_send (void *s, zmq_msg_t *msg, int flags);
func (s *zmqSocket) Send(data []byte, flags SendRecvOption) os.Error {
	var m C.zmq_msg_t
	// Copy data array into C-allocated buffer.
	size := C.size_t(len(data))
	d := unsafe.Pointer(C.malloc(size))
	// FIXME Ideally this wouldn't require a copy.
	C.memcpy(d, unsafe.Pointer(&data[0]), size) // XXX I hope this works...(seems to)
	// The semantics around this failing are not clear. Will d be freed? Who knows.
	if C.zmq_msg_init_data(&m, d, size, C.free_zmq_msg_t_data_ptr, nil) != 0 {
		C.free(d) // We didn't use defer for this because if init succeeds we do not want to free the data buffer.
		return errno()
	}
	if C.zmq_send(s.s, &m, C.int(flags)) != 0 {
		return errno()
	}
	return nil
}

// Receive a message from the socket.
// int zmq_recv (void *s, zmq_msg_t *msg, int flags);
func (s *zmqSocket) Recv(flags SendRecvOption) (data []byte, error os.Error) {
	// Allocate and initialise a new zmq_msg_t
	m := C.alloc_zmq_msg_t()
	defer C.free_zmq_msg_t_data(unsafe.Pointer(m), nil)
	if C.zmq_msg_init(m) != 0 {
		error = errno()
		data = nil
		return
	}
	defer C.zmq_msg_close(m)
	// Receive into message
	if C.zmq_recv(s.s, m, C.int(flags)) != 0 {
		error = errno()
		data = nil
		return
	}
	// Copy message data into a byte array
	// FIXME Ideally this wouldn't require a copy.
	size := C.zmq_msg_size(m)
	error = nil
	data = make([]byte, int(size))
	C.memcpy(unsafe.Pointer(&data[0]), C.zmq_msg_data(m), size)
	return
}


// XXX For now, this library abstracts zmq_msg_t out of the API.
// int zmq_msg_init (zmq_msg_t *msg);
// int zmq_msg_init_size (zmq_msg_t *msg, size_t size);
// int zmq_msg_close (zmq_msg_t *msg);
// size_t zmq_msg_size (zmq_msg_t *msg);
// void *zmq_msg_data (zmq_msg_t *msg);
// int zmq_msg_copy (zmq_msg_t *dest, zmq_msg_t *src);
// int zmq_msg_move (zmq_msg_t *dest, zmq_msg_t *src);
