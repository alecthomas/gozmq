Go Bindings for 0mq (zeromq)
============================
This package implements `Go <http://golang.org>`_ bindings for the `0mq
<http://zeromq.org>`_ C API.

It does not attempt to expose zmq_msg_t at all. Instead, Recv() and Send()
both operate on byte slices, allocating and freeing the memory
automatically. Currently this requires copying to/from C malloced memory,
but a future implementation may be able to avoid this to a certain extent.

Multi-part messages are not supported at all.

A note about memory management: it's not entirely clear from the 0mq
documentation how memory for zmq_msg_t and packet data is managed once 0mq
takes ownership. This module operates under the following (educated)
assumptions:

- For zmq_msg_t structures references are not held beyond the duration of
  any function call.
- Packet data is reference counted. The count is incremented when a packet
  is queued for delivery to a destination (the inference being that for
  delivery to N destinations, the reference count will be incremented N
  times) and decremented once the packet has either been delivered or
  errored.

Warning about thread safety
---------------------------
The `0mq API <http://api.zeromq.org>`_ warns:

  Each ØMQ socket belonging to a particular context may only be used by **the
  thread that created it** using zmq_socket().

This is a bit of an onerous restriction.

The only way to guarantee this in Go, as far as I can tell, is to start a
goroutine, call runtime.LockOSThread(), create a socket then call all socket
methods in this single goroutine. For the gozmq API to abstract this would
require proxying all method calls into this goroutine via channels, vastly
complicating the implementation.

For now I would suggest creation and all subsequent access to each socket be
performed inside a single goroutine pinned with runtime.LockOSThread()::

  context := zmq.Context()
  go func () {
    runtime.LockOSThread()

    socket := context.Socket(zmq.REQ)
    defer socket.Close()
    ... do stuff with socket
    runtime.UnlockOSThread()
  }()
