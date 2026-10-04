BINARY    := herdr-wtm
BUILD_DIR := bin

.PHONY: build test vet fmt lint dead tidy snapshot clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/herdr-wtm

test:
	go test ./... -race -count=1

vet:
	go vet ./...

# fmt fails rather than rewrites: a formatting fix belongs in the commit that
# caused it, not in the one that ran the linter.
fmt:
	@test -z "$$(gofmt -l cmd internal scripts)" || \
		{ echo "gofmt needed:"; gofmt -l cmd internal scripts; exit 1; }

# dead finds functions no path reaches, tests included; staticcheck only sees
# what is unused inside one package.
dead:
	@out=$$(go tool deadcode -test ./...) || exit 1; \
		test -z "$$out" || { echo "unreachable code:"; echo "$$out"; exit 1; }

lint: fmt vet dead
	go tool staticcheck ./...

tidy:
	go mod tidy

# Builds every release archive without publishing (needs goreleaser).
snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf $(BUILD_DIR) dist
