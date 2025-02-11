.PHONY: build run certs node1 node2 node3 node4 node5 run-all

build:
	@go build -o main main.go

certs:
	@echo "Generating certificates..."
	@.\certs\gen-certs.bat

clean-certs:
	@echo "Cleaning old certificates..."
	@if exist "certs\*.key" del /F /Q certs\*.key
	@if exist "certs\*.crt" del /F /Q certs\*.crt
	@if exist "certs\*.srl" del /F /Q certs\*.srl

run: build
	./main

node1: 
	go run main.go -port 8443

node2: 
	go run main.go -port 8444

node3: 
	go run main.go -port 8445

node4: 
	go run main.go -port 8446

node5: 
	go run main.go -port 8447

# Run all nodes in separate terminals
run-all:
	@echo "Starting all nodes..."
	@start make node1
	@start make node2
	@start make node3
	@start make node4
	@start make node5

dev: clean
	@echo "Starting development environment..."
	@mkdir -p storage
	@go run main.go -port 8001 & \
    go run main.go -port 8002 & \
    go run main.go -port 8003 & \
    go run main.go -port 8004 & \
    go run main.go -port 8005

clean:
	@echo "Cleaning up..."
	@rm -rf storage/*
	@pkill -f "go run main.go" || true

monitor:
	@watch -n 1 "ls -l storage/"

test:
	@go test -v ./...

serve:
	go run *.go -type serve

dial:
	go run *.go -type dial


