.PHONY: gen breaking

gen:
	@if [ -n "$$(find proto -name '*.proto' 2>/dev/null)" ]; then buf generate; else echo "==> gen: no .proto files yet - skipping"; fi

# RFC-011 J.3 buf breaking-change check, against the PR base ref in CI
# (BUF_AGAINST_REF) or the local default branch otherwise.
breaking:
	@if [ -n "$$(find proto -name '*.proto' 2>/dev/null)" ]; then \
		buf breaking --against ".git#branch=$${BUF_AGAINST_REF:-develop}"; \
	else \
		echo "==> breaking: no .proto files yet - skipping"; \
	fi
