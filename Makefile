VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: run build test lint validate release-snapshot

run:
	go run ./cmd/vimdojo

build:
	go build -ldflags "$(LDFLAGS)" -o ./bin/vimdojo ./cmd/vimdojo

test:
	go test ./...

lint:
	go vet ./...
	gofmt -l .

validate:
	go run ./cmd/validate

# Compila los binarios de todas las plataformas con goreleaser sin
# publicar nada (ni tags, ni GitHub release); útil para probar el
# release localmente. Requiere goreleaser instalado y un repo git con al
# menos un commit.
release-snapshot:
	goreleaser release --snapshot --clean
