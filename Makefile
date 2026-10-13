.PHONY: install run-server stop-server run-client run-client-gui lint clean

all:
	@echo "Use the following commands"
	@echo "make run-server"
	@echo "make stop-server"
	@echo "make run-client"
	@echo "make run-client-gui"
	@echo "make lint"
	@echo "make clean"

install:
	go mod tidy

run-server:
	go run ./tcp/server

stop-server:
	fuser -k 8080/tcp

run-client:
	go run ./tcp/client

run-client-gui:
	go run . gui

lint:
	go vet ./...

clean:
	rm -f server tap-gui
	rm -f *.log
	go clean
