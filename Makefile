build:
	@go build -o bin/godfs

run: build
	@./bin/godfs

test:
	@go test -v ./...
