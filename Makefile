GO_FILES := $(shell find . -name '*.go' -not -path './.git/*')

.PHONY: fmt fmt-check test test-integration vet verify

fmt:
	gofmt -w $(GO_FILES)

fmt-check:
	@test -z "$$(gofmt -l $(GO_FILES))"

test:
	go test ./...

test-integration:
	go test -count=1 -run Integration ./...

vet:
	go vet ./...

verify: fmt-check test vet
