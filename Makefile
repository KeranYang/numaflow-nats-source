.PHONY: swagger-gen
swagger-gen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger generate spec -o ./schema/helper/config-swagger.json --scan-models
	swagger mixin ./schema/helper/base.json ./schema/helper/config-swagger.json -o ./schema/config.json

.PHONY: build
build: swagger-gen
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/nats-source main.go

.PHONY: image
image: build
	docker build -t "quay.io/numaio/numaflow-source/nats-source:v0.5.2" --target nats-source .

clean:
	-rm -rf ./dist

