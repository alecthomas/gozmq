include $(GOROOT)/src/Make.inc

all: client server

client: client.$O
	$(LD) -o $@ $<

server: server.$O
	$(LD) -o $@ $<

%.$O: %.go
	$(GC) -o $@ $<

CLEANFILES+=client server

include $(GOROOT)/src/Make.common
