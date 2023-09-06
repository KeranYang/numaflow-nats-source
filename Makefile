.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/nats-source main.go

.PHONY: image
image: build
	docker build -t "quay.io/numaio/numaflow-source/nats-source:v0.5.2" --target nats-source .

clean:
	-rm -rf ./dist