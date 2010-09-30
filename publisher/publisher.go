package main


import (
  "fmt"
  "zmq"
)  


func main() {
  c := zmq.New()
  s := c.Socket(zmq.ZMQ_PAIR)

  s.Bind("tcp://127.0.0.1:2000")
  m, _ := zmq.Message([]byte("hello world"))
  fmt.Printf("message: %s\n", m.Data())
  s.Send(m, 0)
}
