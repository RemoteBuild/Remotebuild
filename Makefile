build:
	go mod download
	go build -o main

debug: build 
	./main server start -l debug

run: build 
	./main server start

upgrade:
	go mod download
	go get -u -v
	go mod tidy
	go mod verify

test:
	go test ./...

default: build
