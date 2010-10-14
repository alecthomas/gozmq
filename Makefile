include $(GOROOT)/src/Make.inc

TARG=zmq

CGOFILES=zmq.go
CGO_CFLAGS=-I/Users/aat/Homebrew/include
CGO_LDFLAGS=-lzmq -L/Users/aat/Homebrew/lib

include $(GOROOT)/src/Make.pkg
