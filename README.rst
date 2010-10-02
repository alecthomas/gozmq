`Go <http://golang.org>`_ Bindings for `0mq <http://zeromq.org>`_
=================================================================
This package implements Go bindings for the 0mq C API.

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
