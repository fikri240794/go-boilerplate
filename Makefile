clean:
	@find . -name *mock* -delete
	@find . -name wire_gen.go -delete
	@rm -rf ./transports/http/docs/swagger

generate:
	go generate ./...

test: clean generate
	go test -v -cover -covermode=atomic ./...

http: clean generate
	go run main.go http

grpc: clean generate
	go run main.go grpc

event-consumer: clean generate
	go run main.go event-consumer

app: clean generate
	go run main.go app

build: clean generate
	go build