all: sender receiver


sender: sender.6
	6l -o $@ $<

receiver: receiver.6
	6l -o $@ $<


%.6: %.go
	6g -o $@ $<


clean:
	rm -f sender.6 sender receiver.6 receiver
