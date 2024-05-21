all: mod server client

client:
	go build -o ./bin/client ./cmd/client

server:
	go build -o ./bin/server ./cmd/server

clean:
	rm -rf ./bin ./logs

mod:
	go mod tidy