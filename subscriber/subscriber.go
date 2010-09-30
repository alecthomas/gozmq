package main


import (
  "fmt"
  "zmq"
)


func main() {
  c := zmq.New()
  s := c.Socket(zmq.ZMQ_PAIR)
  // Subscribe to all channels
  // s.SetSockOptString(zmq.ZMQ_SUBSCRIBE, "")
  s.Connect("tcp://127.0.0.1:2000")

  for {
    fmt.Printf("Recving...\n")
    msg, rc := s.Recv(0)
    if rc != 0 {
      panic("Recv failed: " + rc.String())
    }
    fmt.Printf("Message received: %s\n", msg)
  }
}
