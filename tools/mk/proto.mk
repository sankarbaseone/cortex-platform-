.PHONY: gen

gen:
	@if [ -n "$$(find proto -name '*.proto' 2>/dev/null)" ]; then buf generate; else echo "==> gen: no .proto files yet - skipping"; fi
