package main

import "zmq"

func main() {
	context := zmq.Context()
	socket := context.Socket(zmq.REP)
	socket.Bind("tcp://127.0.0.1:5000")
	socket.Bind("tcp://127.0.0.1:6000")

	for {
		msg, _ := socket.Recv(0)
		println("Got", string(msg))
		socket.Send(msg, 0)
	}
}
