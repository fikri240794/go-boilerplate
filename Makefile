clean:
	@find . -name *mock*.go -delete
	@find . -name wire_gen.go -delete
	@rm -rf ./transports/http/docs/swagger

generate:
	go generate ./... && go run github.com/vektra/mockery/v3

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