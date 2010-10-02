package main

import "fmt"
import "zmq"

func main() {
	context := zmq.Context()
	socket := context.Socket(zmq.REQ)
	socket.Connect("tcp://127.0.0.1:5000")
	socket.Connect("tcp://127.0.0.1:6000")

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("msg %d", i)
		socket.Send([]byte(msg), 0)
		println("Sending", msg)
		socket.Recv(0)
	}
}
