IMG = forrest-tao/device-plugin-demo:v1

.PHONY: build
build:
	CGO_ENABLE=0 GOOS=linux go build -o bin/device-plugin-demo cmd/main.go

.PHONY:build-image
build-image:
	docker build -t ${IMG} .