.PHONY: server client gui

all:
	@echo "Use the following commands"
	@echo "make server"
	@echo "make client"
	@echo "make gui"

server:
	go run ./tcp/server

client:
	go run .

gui:
	go run . gui
