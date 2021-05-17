swagger = docker run --rm -it -e GOPATH=$$HOME/go:/go -v $$HOME:$$HOME -w $$(pwd) quay.io/goswagger/swagger
PORT ?= 3001 # HTTP port for docs server

test:
	go test -v ./...

run:
	go run ./cmd/httpserver/main.go

docs:
	go run ./cmd/docserver/main.go --port $(PORT)

swagger:
	$(swagger) generate spec -o ./swagger.yaml --scan-models
