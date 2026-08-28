.PHONY: server client gui clean

all:
	@echo "Use the following commands"
	@echo "make server"
	@echo "make client"
	@echo "make gui"
	@echo "make clean"

server:
	go run ./tcp/server

client:
	go run .

gui:
	go run . gui

clean:
	rm -f server tap-gui
	rm -f *.log
	go clean
