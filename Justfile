default:
    @just --list

build:
    go build -o bin/twinmind-mcp ./cmd/twinmind-mcp
    go build -o bin/twinmind ./cmd/twinmind-cli

build-release version:
    go build -ldflags "-s -w -X main.Version={{version}}" -o bin/twinmind-mcp ./cmd/twinmind-mcp
    go build -ldflags "-s -w -X main.Version={{version}}" -o bin/twinmind ./cmd/twinmind-cli

test:
    go test ./... -v

test-race:
    go test -race ./... -v

lint:
    golangci-lint run ./...

fmt:
    gofmt -w .

tidy:
    go mod tidy

snapshot:
    goreleaser release --snapshot --clean

clean:
    rm -rf bin/ dist/

run-mcp:
    go run ./cmd/twinmind-mcp

run-cli *args:
    go run ./cmd/twinmind-cli {{args}}

install path="/usr/local/bin/": build
    sudo cp ./bin/twinmind ./bin/twinmind-mcp {{path}}
