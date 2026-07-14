.PHONY: bootstrap
bootstrap:
	@echo "==> mise install (Go/Rust/Python/Node per .tool-versions)"
	mise install
	@echo "==> buf"
	command -v buf >/dev/null 2>&1 || mise use -g -y buf@1.42.0
	@echo "==> protoc-gen-go / protoc-gen-go-grpc"
	command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	@echo "==> golangci-lint"
	command -v golangci-lint >/dev/null 2>&1 || mise use -g -y golangci-lint@latest
	@echo "==> uv"
	command -v uv >/dev/null 2>&1 || curl -LsSf https://astral.sh/uv/install.sh | sh
	@echo "==> pnpm (web/)"
	@if [ -f web/app/package.json ]; then pnpm -C web/app install; else echo "    no web/app package.json yet - skipping"; fi
