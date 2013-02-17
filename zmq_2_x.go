// +build !zmq_3_x

/*
  Copyright 2010-2012 Alec Thomas

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

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lzmq
#include <zmq.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import "unsafe"

const (
	RECOVERY_IVL_MSEC = Int64SocketOption(C.ZMQ_RECOVERY_IVL_MSEC)
	SWAP              = Int64SocketOption(C.ZMQ_SWAP)
	MCAST_LOOP        = Int64SocketOption(C.ZMQ_MCAST_LOOP)
	HWM               = UInt64SocketOption(C.ZMQ_HWM)
	NOBLOCK           = SendRecvOption(C.ZMQ_NOBLOCK)
)

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
