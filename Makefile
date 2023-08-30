.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/simple-source-example main.go

.PHONY: image
image: build
	docker build -t "quay.io/numaio/numaflow-go/keran-test-nats-source:secret0.5.1" --target simple-source .

clean:
	-rm -rf ./dist
