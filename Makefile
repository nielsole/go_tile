CGO_ENABLED ?= 0

build:
	CGO_ENABLED=$(CGO_ENABLED) go build .

clean:
	CGO_ENABLED=$(CGO_ENABLED) go clean .

run:
	CGO_ENABLED=$(CGO_ENABLED) go run .

test:
	CGO_ENABLED=$(CGO_ENABLED) go test .

.PHONY: \
	build \
	clean \
	run \
	test
