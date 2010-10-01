package main


import (
  "fmt"
  "zmq"
)


func main() {
  context := zmq.Context()
  socket := context.Socket(zmq.REP)
  socket.Bind("tcp://127.0.0.1:5000")
  
  for {
	msg, _ := socket.Recv(0)
	fmt.Printf("Received message '%s'\n", string(msg))
	socket.Send([]byte(msg), 0)
  }
}
