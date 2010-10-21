include $(GOROOT)/src/Make.inc

TARG=github.com/alecthomas/gozmq

CGOFILES=zmq.go
CGO_LDFLAGS=-lzmq

include $(GOROOT)/src/Make.pkg
