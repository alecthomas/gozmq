Go Bindings for 0mq (zeromq)
############################
This package implements `Go <http://golang.org>`_ bindings for the `0mq
<http://zeromq.org>`_ C API.

Note that this is *not* the same as `this implementation
<http://github.com/boggle/gozero>`_.

Differences from the C API
==========================
The API implemented by this package does not attempt to expose ``zmq_msg_t`` at
all. Instead, ``Recv()`` and ``Send()`` both operate on byte slices, allocating
and freeing the memory automatically. Currently this requires copying to/from C
malloced memory, but a future implementation may be able to avoid this to a
certain extent.

All major features are supported: contexts, sockets, devices, and polls.

Example
-------
Here are direct translations of some of the examples from `this blog post
<http://nichol.as/zeromq-an-introduction>`_.

A simple echo server::

  package main

  import zmq "github.com/alecthomas/gozmq"

  func main() {
    context := zmq.NewContext()
    socket := context.NewSocket(zmq.REP)
    socket.Bind("tcp://127.0.0.1:5000")
    socket.Bind("tcp://127.0.0.1:6000")

    for {
      msg, _ := socket.Recv(0)
      println("Got", string(msg))
      socket.Send(msg, 0)
    }
  }

A simple client for the above server::

  package main

  import "fmt"
  import zmq "github.com/alecthomas/gozmq"

  func main() {
    context := zmq.NewContext()
    socket := context.NewSocket(zmq.REQ)
    socket.Connect("tcp://127.0.0.1:5000")
    socket.Connect("tcp://127.0.0.1:6000")

    for i := 0; i < 10; i++ {
      msg := fmt.Sprintf("msg %d", i)
      socket.Send([]byte(msg), 0)
      println("Sending", msg)
      socket.Recv(0)
    }
  }

Caveats
=======

Thread safety
-------------
The `0mq API <http://api.zeromq.org>`_ warns:

  Each ØMQ socket belonging to a particular context may only be used by **the
  thread that created it** using ``zmq_socket()``.

This is a bit of an onerous restriction.

The only way to guarantee this in Go, as far as I can tell, is to start a
goroutine, call ``runtime.LockOSThread()``, create a socket then call all socket
methods in this single goroutine. For the gozmq API to abstract this would
require proxying all method calls into this goroutine via channels, vastly
complicating the implementation.

For now I would suggest creation and all subsequent access to each socket be
performed inside a single goroutine pinned with ``runtime.LockOSThread()``::

  context := zmq.NewContext()
  go func () {
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()

    socket := context.NewSocket(zmq.REQ)
    defer socket.Close()
    ... do stuff with socket
  }()

Memory management
-----------------
It's not entirely clear from the 0mq documentation how memory for ``zmq_msg_t``
and packet data is managed once 0mq takes ownership. After digging into the
source a little, this package operates under the following (educated)
assumptions:

- References to ``zmq_msg_t`` structures are not held by the C API beyond the
  duration of any function call.
- Packet data is reference counted internally by the C API. The count is
  incremented when a packet is queued for delivery to a destination (the
  inference being that for delivery to N destinations, the reference count will
  be incremented N times) and decremented once the packet has either been
  delivered or errored.
