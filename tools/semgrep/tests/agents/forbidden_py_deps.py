# Semgrep test fixture only (see tools/semgrep/rules/forbidden-py-deps.yaml).
# Lives under agents/ to exercise the rule's `paths.include`.

# ruleid: py-no-pandas-in-services
import pandas as pd

# ok: py-no-pandas-in-services
import json
