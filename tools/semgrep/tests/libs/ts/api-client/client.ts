// Semgrep test fixture only (see tools/semgrep/rules/forbidden-ts-deps.yaml).
// Lives under libs/ts/api-client/** to exercise the rule's `paths.exclude`.

// ok: ts-no-axios-outside-api-client
import axios from "axios";
