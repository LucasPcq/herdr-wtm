BINARY    := herdr-wtm
BUILD_DIR := bin

.PHONY: build test vet fmt lint dead tidy snapshot release-notes demos clean

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

# Re-records the README GIFs from docs/demos/*.tape (needs vhs and tmux), in an
# isolated HOME and herdr session; `make demos TAPE=agent` records one.
demos:
	docs/demos/record.sh $(TAPE)

# Builds every release archive without publishing (needs goreleaser).
snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -rf $(BUILD_DIR) dist

# release-notes prints the CHANGELOG section of VERSION (0.2.0, no v): the
# release workflow publishes it as the GitHub release notes, where a relative
# link would resolve under /releases/tag/, so docs links are pinned to the tag.
NOTES_REF = $(if $(filter Unreleased,$(VERSION)),main,v$(VERSION))

release-notes:
	@test -n "$(VERSION)" || { echo "usage: make release-notes VERSION=x.y.z"; exit 1; }
	@awk -v v="$(VERSION)" 'index($$0, "## [" v "]") == 1 { on = 1; next } on && /^## \[/ { exit } on && /^\[[^]]+\]: / { exit } on' CHANGELOG.md | \
		sed -e '/./,$$!d' -e 's|](docs/|](https://github.com/LucasPcq/herdr-wtm/blob/$(NOTES_REF)/docs/|g'
	@grep -q "^## \[$(VERSION)\]" CHANGELOG.md || { echo "no CHANGELOG section for $(VERSION)" >&2; exit 1; }
