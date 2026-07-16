// Semgrep test fixture only (see tools/semgrep/rules/forbidden-ts-deps.yaml).

// ruleid: ts-no-redux
import { createStore } from "redux";

// ruleid: ts-no-axios-outside-api-client
import axios from "axios";

// ok: ts-no-redux
import { useQuery } from "@tanstack/react-query";
