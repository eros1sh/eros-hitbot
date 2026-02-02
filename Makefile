.PHONY: build run clean test

build:
	go build -o eroshit.exe ./cmd/eroshit

run: build
	./eroshit.exe -domain example.com -pages 5 -duration 2 -hpm 10

clean:
	rm -f eroshit eroshit.exe

test:
	go test ./...
