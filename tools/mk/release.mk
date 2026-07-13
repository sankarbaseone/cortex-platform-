.PHONY: image chart-lint sbom sign

image:
	@echo "==> image SVC=$(SVC): no Dockerfiles exist yet - deferred to later tasks"

chart-lint:
	@echo "==> chart-lint: no charts exist yet - deferred to later tasks"

sbom:
	@echo "==> sbom: no build artifacts exist yet - nothing to scan"

sign:
	@echo "==> sign: no build artifacts exist yet - nothing to sign"
