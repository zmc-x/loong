all: server

client:
	go build -o ./bin/client ./cmd/client

server:
	go build -o ./bin/server ./cmd/server

clean:
	rm -rf ./bin ./logs