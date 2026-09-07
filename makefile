.PHONY: install server stop-server client gui lint clean

all:
	@echo "Use the following commands"
	@echo "make server"
	@echo "make stop-server"
	@echo "make client"
	@echo "make gui"
	@echo "make lint"
	@echo "make clean"

install:
	go mod tidy

server:
	go run ./tcp/server

stop-server:
	fuser -k 8080/tcp

client:
	go run ./tcp/client

gui:
	go run . gui

lint:
	go vet ./...

clean:
	rm -f server tap-gui
	rm -f *.log
	go clean
