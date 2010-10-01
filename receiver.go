package main


import (
  "fmt"
  "zmq"
)


func main() {
  context := zmq.Context()
  socket := context.Socket(zmq.REQ)
  socket.Connect("tcp://127.0.0.1:5000")
 
  for i := 0; i < 10; i++ {
	msg := fmt.Sprintf("msg %d", i)
	fmt.Printf("Sending %s\n", msg)
	socket.Send([]byte(msg), 0)
	fmt.Printf("Sent %s\n", msg)
	fmt.Printf("Waiting for a message.\n")
	msg2, _ := socket.Recv(0)
	fmt.Printf("Received message '%s'\n", string(msg2))
  }
}
