package main


import (
  "zmq"
)  


func main() {
  c := zmq.New()
  s := c.Socket(zmq.ZMQ_PAIR)

  s.Bind("tcp://127.0.0.1:2000")
  s.Send(([]byte)("hello world"), zmq.ZMQ_NOBLOCK)
}
