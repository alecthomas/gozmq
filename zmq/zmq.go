package zmq

// #include <zmq.h>
// #include <stdlib.h>
import "C"

import (
  "fmt"
  "os"
  "unsafe"
)


type Error os.Errno

type SocketType int
type IntSocketOption int
type StringSocketOption int
type SendRecvOption int

const (
  // Socket types
  ZMQ_PAIR = SocketType(C.ZMQ_PAIR)
  ZMQ_PUB = SocketType(C.ZMQ_PUB)
  ZMQ_SUB = SocketType(C.ZMQ_SUB)
  ZMQ_REQ = SocketType(C.ZMQ_REQ)
  ZMQ_REP = SocketType(C.ZMQ_REP)
  ZMQ_XREQ = SocketType(C.ZMQ_XREQ)
  ZMQ_XREP = SocketType(C.ZMQ_XREP)
  ZMQ_PULL = SocketType(C.ZMQ_PULL)
  ZMQ_PUSH = SocketType(C.ZMQ_PUSH)

  // Socket options
  ZMQ_HWM = IntSocketOption(C.ZMQ_HWM)
  ZMQ_SWAP = IntSocketOption(C.ZMQ_SWAP)
  ZMQ_AFFINITY = IntSocketOption(C.ZMQ_AFFINITY)
  ZMQ_IDENTITY = StringSocketOption(C.ZMQ_IDENTITY)
  ZMQ_SUBSCRIBE = StringSocketOption(C.ZMQ_SUBSCRIBE)
  ZMQ_UNSUBSCRIBE = StringSocketOption(C.ZMQ_UNSUBSCRIBE)
  ZMQ_RATE = IntSocketOption(C.ZMQ_RATE)
  ZMQ_RECOVERY_IVL = IntSocketOption(C.ZMQ_RECOVERY_IVL)
  ZMQ_MCAST_LOOP = IntSocketOption(C.ZMQ_MCAST_LOOP)
  ZMQ_SNDBUF = IntSocketOption(C.ZMQ_SNDBUF)
  ZMQ_RCVBUF = IntSocketOption(C.ZMQ_RCVBUF)
  ZMQ_RCVMORE = IntSocketOption(C.ZMQ_RCVMORE)

  // Send/recv options
  ZMQ_NOBLOCK = SendRecvOption(C.ZMQ_NOBLOCK)
  ZMQ_SNDMORE = SendRecvOption(C.ZMQ_SNDMORE)
)



// Misc functions

// TODO int zmq_poll (zmq_pollitem_t *items, int nitems, long timeout);
// TODO int zmq_device (int device, void * insocket, void* outsocket);


// void zmq_version (int *major, int *minor, int *patch);
func Version() (int, int, int) {
  var major, minor, patch C.int
  C.zmq_version(&major, &minor, &patch)
  return int(major), int(minor), int(patch)
}

// int zmq_errno ();
func Errno() Error {
  return Error(C.zmq_errno())
}

// const char *zmq_strerror (int errnum);
func (e Error) String() string {
  return C.GoString(C.zmq_strerror(C.int(e)))
}



// Context methods
type ZmqContext struct {
  c unsafe.Pointer
}


// void *zmq_init (int io_threads);
func New() *ZmqContext {
  c := new(ZmqContext)
  // TODO Pass something useful here. Number of cores?
  c.c = C.zmq_init(1)
  return c
}

// int zmq_term (void *context);
func (c *ZmqContext) destroy() {
  fmt.Printf("destroying 0mq context\n")
  C.zmq_term(c.c)
}



// Socket methods
type ZmqSocket struct {
  s unsafe.Pointer
  m C.zmq_msg_t
}

// void *zmq_socket (void *context, int type);
func (c *ZmqContext) Socket(type_ SocketType) (s *ZmqSocket, error Error) {
  s = new(ZmqSocket)
  s.s = C.zmq_socket(c.c, C.int(type_))
  error = Error(C.zmq_msg_init(&s.m))
  return s
}

// int zmq_close (void *s);
func (s *ZmqSocket) Close() Error {
  return Error(C.zmq_close(s.s))
}

func (s *ZmqSocket) destroy() {
  if error := s.Close(); error != 0 {
    panic("Error while destroying ZmqSocket: " + error.String() + "\n")
  }
}

// int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
func (s *ZmqSocket) SetSockOptInt(option IntSocketOption, value int64) Error {
  return Error(C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(&value))))
}

func (s *ZmqSocket) SetSockOptString(option StringSocketOption, value string) Error {
  v := C.CString(value)
  defer C.free(unsafe.Pointer(v))
  return Error(C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(v), C.size_t(len(value))))
}

// int zmq_getsockopt (void *s, int option, void *optval, size_t *optvallen);
func (s *ZmqSocket) GetSockOptInt(option IntSocketOption) (value int64, error Error) {
  size := C.size_t(8)
  error = Error(C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&value), &size))
  return
}

func (s *ZmqSocket) GetSockOptString(option StringSocketOption) (value string, error Error) {
  var buffer [1024]byte
  var size C.size_t = 1024
  error = Error(C.zmq_getsockopt(s.s, C.int(option), unsafe.Pointer(&buffer), &size))
  value = string(buffer[:size])
  return
}

// int zmq_bind (void *s, const char *addr);
func (s *ZmqSocket) Bind(address string) Error {
  a := C.CString(address)
  defer C.free(unsafe.Pointer(a))
  return Error(C.zmq_bind(s.s, a))
}

// int zmq_connect (void *s, const char *addr);
func (s *ZmqSocket) Connect(address string) Error {
  a := C.CString(address)
  defer C.free(unsafe.Pointer(a))
  return Error(C.zmq_connect(s.s, a))
}

// TODO int zmq_send (void *s, zmq_msg_t *msg, int flags);
func (s *ZmqSocket) Send(m *ZmqMessage, flags SendRecvOption) Error {
  return Error(C.zmq_send(s.s, &m.m, C.int(flags)))
}

// int zmq_recv (void *s, zmq_msg_t *msg, int flags);
func (s *ZmqSocket) Recv(flags SendRecvOption) (data *[]byte, error Error) {
  m := new(C.zmq_msg_t)
  error = Error(C.zmq_recv(s.s, m, C.int(flags)))
  if error != 0 {
    data = nil
    return
  }
  defer C.zmq_msg_close(m)
  size := int(C.zmq_msg_size(m))
  data = new([10]byte)
  //(C.zmq_msg_data(m))
  return
}



// Message methods
type ZmqMessage struct {
  m C.zmq_msg_t
  init bool
  data *[]byte
}

// int zmq_msg_init (zmq_msg_t *msg);
func EmptyMessage() (m *ZmqMessage, error Error) {
  m = new(ZmqMessage)
  error = Error(C.zmq_msg_init(&m.m))
  m.init = true
  return 
}

// int zmq_msg_init_size (zmq_msg_t *msg, size_t size);
func SizedMessage(size int) (m *ZmqMessage, error Error) {
  m = new(ZmqMessage)
  error = Error(C.zmq_msg_init_size(&m.m, C.size_t(size)))
  m.init = true
  return 
}

// int zmq_msg_init_data (zmq_msg_t *msg, void *data, size_t size, zmq_free_fn *ffn, void *hint);
func Message(data []byte) (m *ZmqMessage, error Error) {
  m = new(ZmqMessage)
  m.init = true
  error = Error(C.zmq_msg_init_data(&m.m, unsafe.Pointer(&data), C.size_t(len(data)), nil, nil))
  m.data = &data
  return 
}

// int zmq_msg_close (zmq_msg_t *msg);
func (m *ZmqMessage) Close() (error Error) {
  error = Error(C.zmq_msg_close(&m.m))
  m.data = nil
  return
}

func (m *ZmqMessage) destroy() {
  if rc := m.Close(); rc != 0 {
    panic("Failed to destroy message: " + rc.String())
  }
}

// size_t zmq_msg_size (zmq_msg_t *msg);
func (m *ZmqMessage) Size() int {
  return int(C.zmq_msg_size(&m.m))
}

// void *zmq_msg_data (zmq_msg_t *msg);
func (m *ZmqMessage) Data() *[]byte {
  return (*[]byte)(C.zmq_msg_data(&m.m))
}

// TODO int zmq_msg_copy (zmq_msg_t *dest, zmq_msg_t *src);
// TODO int zmq_msg_move (zmq_msg_t *dest, zmq_msg_t *src);
