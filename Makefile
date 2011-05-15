include $(GOROOT)/src/Make.inc

TARG=github.com/alecthomas/gozmq

CGOFILES=zmq.go
CGO_CFLAGS=-I/usr/local/include
CGO_LDFLAGS=-lzmq -L/usr/local/lib

include $(GOROOT)/src/Make.pkg
