COMMIT_HASH = $(shell git rev-parse HEAD)

.PHONY: build
build:
	go build -ldflags "-s -w -X github.com/isokolovskii/commitizen/internal/version.commit=$(COMMIT_HASH)" -o commitizen

.PHONY: build-with-coverage
build-with-coverage:
	go build -cover -ldflags "-s -w -X github.com/isokolovskii/commitizen/internal/version.commit=$(COMMIT_HASH)" -o commitizen

install: build
ifeq ($(shell go env GOOS),windows)
	copy commitizen $(shell go env GOPATH)\bin\commitizen.exe
else
	cp commitizen $$(go env GOPATH)/bin
endif

.PHONY: test
test:
	go test -cpu 24 -race -count=1 -timeout=30s ./...

.PHONY: bench
bench:
	go test -cpu 24 -race -run=Bench -bench=. ./...

GOLANGCI_LINT_BIN := $(shell go env GOPATH)/bin/golangci-lint

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run

$(GOLANGCI_LINT_BIN):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.6.2

.ONESHELL:
version:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Release can only be created from the main branch"; \
		exit 1; \
	fi
	@if ! command -v git-cliff &> /dev/null; then \
		echo "git-cliff is not installed. Please install it from https://git-cliff.org/docs/installation/"; \
		exit 1; \
	fi

	@version=$$(git cliff --bumped-version 2>/dev/null); \
	echo "Bumping to version $$version"; \
	sed -i '' "s/version = .*/version = \"$$version\"/" internal/version/version.go; \
	sed -i '' "s/go install github.com\/isokolovskii\/commitizen@.*/go install github.com\/isokolovskii\/commitizen@$$version/" README.md; \
	git cliff --bump -o CHANGELOG.md; \
	git add internal/version/version.go README.md CHANGELOG.md; \
	git commit -m "chore(release): $$version"; \
	git tag "$$version"; \
	git push origin main; \
	git push origin $$version;
