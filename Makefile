.DEFAULT_GOAL := help

include tools/mk/toolchain.mk
include tools/mk/go.mk
include tools/mk/proto.mk
include tools/mk/dev.mk
include tools/mk/release.mk

.PHONY: help
help:
	@echo "Targets: bootstrap build gen lint test itest e2e bench dev-up dev-down migrate image chart-lint sbom sign"
