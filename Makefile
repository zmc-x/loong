
server:
	go build -o ./bin/server ./cmd/server

clean:
	rm -rf ./bin ./logs