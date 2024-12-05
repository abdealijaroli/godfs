build:
	@go build -o bin/godfs

run: build
	@./bin/godfs

test:
	@go test -v ./...

.PHONY: serve dial

serve:
	go run *.go -type serve

dial:
	go run *.go -type dial