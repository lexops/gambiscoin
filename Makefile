BINARY_NAME=gambiscoin

build:
		go build -o dist/$(BINARY_NAME) api/api.go

run: build
		./dist/$(BINARY_NAME)

node_1:
		./dist/$(BINARY_NAME) 3001

node_2:
		./dist/$(BINARY_NAME) 3002

node_3:
		./dist/$(BINARY_NAME) 3003

node_4:
		./dist/$(BINARY_NAME) 3004
