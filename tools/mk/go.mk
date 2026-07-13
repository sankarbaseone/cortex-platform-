.PHONY: build lint test itest

build:
	@if [ -n "$$(find . -path ./docs -prune -o -name '*.go' -print 2>/dev/null)" ]; then go build ./...; else echo "==> build: no Go source outside docs/ yet - skipping"; fi

lint:
	@if [ -n "$$(find . -path ./docs -prune -o -name '*.go' -print 2>/dev/null)" ]; then golangci-lint run ./...; else echo "==> golangci-lint: no Go source outside docs/ yet - skipping"; fi
	@if [ -n "$$(find proto -name '*.proto' 2>/dev/null)" ]; then buf lint; else echo "==> buf lint: no .proto files yet - skipping"; fi

test:
	@if [ -n "$$(find . -path ./docs -prune -o -name '*.go' -print 2>/dev/null)" ]; then go test ./...; else echo "==> test: no Go source outside docs/ yet - skipping"; fi

itest:
	@echo "==> itest: no adapters implemented yet - nothing to run"
