.PHONY: dev-up dev-down migrate e2e bench

dev-up:
	@echo "==> dev-up: no dev stack (kind/compose) defined yet - deferred to T-004"

dev-down:
	@echo "==> dev-down: no dev stack (kind/compose) defined yet - deferred to T-004"

migrate:
	@echo "==> migrate SVC=$(SVC): no migrations exist yet - deferred to T-004"

e2e:
	@echo "==> e2e: no services implemented yet - nothing to run"

bench:
	@echo "==> bench: no benchmarks implemented yet - nothing to run"
