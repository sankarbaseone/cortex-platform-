.PHONY: security semgrep-test

security:
	@echo "==> semgrep (tools/semgrep/rules)"
	@command -v semgrep >/dev/null 2>&1 && semgrep --config tools/semgrep/rules --error . || echo "    semgrep not installed locally - skipping (CI installs it)"
	@echo "==> gitleaks"
	@command -v gitleaks >/dev/null 2>&1 && gitleaks detect --no-banner --source . || echo "    gitleaks not installed locally - skipping (CI installs it)"
	@echo "==> osv-scanner"
	@command -v osv-scanner >/dev/null 2>&1 && osv-scanner -r . || echo "    osv-scanner not installed locally - skipping (CI installs it)"

semgrep-test:
	@command -v semgrep >/dev/null 2>&1 && semgrep --test --config tools/semgrep/rules tools/semgrep/tests || echo "    semgrep not installed locally - skipping (CI installs it)"
