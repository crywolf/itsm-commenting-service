swagger = docker run --rm -it -e GOPATH=$$HOME/go:/go -v $$HOME:$$HOME -w $$(pwd) quay.io/goswagger/swagger:v0.27.0
PORT ?= 3001 # HTTP port for docs server

COMMIT=$(shell git rev-parse HEAD)
COMMIT_SHORT=$(shell git rev-parse --short HEAD)
BRANDNAME?=itsm-commenting-service
IMG_REPO?=
IMG_TAG?=${COMMIT_SHORT}
IMG_TAG_VERSION?=latest
CGO?=0
GOPROXY?=$(shell go env GOPROXY)
GOPRIVATE?='github.com/KompiTech/*'

PKG_NAME?=${BRANDNAME}
IMAGE?=${BRANDNAME}
CMD_PATH?=cmd/httpserver/main.go
BUILD_DIR?=build

test:
	go test -v ./pkg/...

e2e-test:
	go test -v ./e2e_tests/.

test-all: test e2e-test

run:
	go run ./cmd/httpserver/main.go

docs:
	go run ./cmd/docserver/main.go --port $(PORT)

swagger:
	$(swagger) generate spec -o ./swagger.yaml --scan-models

build-linux:
	env GO111MODULE=on GOOS=linux GOPROXY=${GOPROXY} GOARCH=amd64 CGO_ENABLED=${CGO} go build -o ${BUILD_DIR}/${PKG_NAME}.linux ${CMD_PATH}

clean:
	rm -rf ./${BUILD_DIR}/

image: clean
	DOCKER_BUILDKIT=1 docker build --ssh default --build-arg GOPRIVATE=${GOPRIVATE} --build-arg GOPROXY="${GOPROXY}" --build-arg BRAND=${BRANDNAME} -t ${IMG_REPO}${IMAGE}:${IMG_TAG} -t ${IMG_REPO}${IMAGE}:${IMG_TAG_VERSION} --progress=plain .

image-publish: image publish

publish:
	docker push ${IMG_REPO}${IMAGE}:${IMG_TAG}
	docker push ${IMG_REPO}${IMAGE}:${IMG_TAG_VERSION}

list-updates:
	go list -u -m -json all 2>/dev/null | jq 'select(. | has("Update")) | select(. | any(.; .Indirect != true))' | jq -r '(.Update.Path + "@" + .Update.Version)'
