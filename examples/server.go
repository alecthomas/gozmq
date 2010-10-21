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
