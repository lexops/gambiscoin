BINARY_NAME=gambiscoin

build:
		go build -o bin/$(BINARY_NAME) cmd/gambiscoin/main.go

run: build
		./bin/$(BINARY_NAME)

node1:
		./bin/$(BINARY_NAME) 3001

node2:
		./bin/$(BINARY_NAME) 3002

node3:
		./bin/$(BINARY_NAME) 3003

node4:
		./bin/$(BINARY_NAME) 3004
