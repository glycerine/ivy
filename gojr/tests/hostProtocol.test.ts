import { describe, expect, test } from "./testHarness.js";
import {
  evaluationResultToHostPayload,
  formatDiagnostic,
  runtimeOptionsFromEnvironment,
  spreadsheetFixtureResultToHostPayload
} from "../src/index.js";

describe("Go-junior host protocol", () => {
  test("formats diagnostics with filenames, spans, and stacks in one shared path", () => {
    expect(formatDiagnostic({
      filename: "fallback.go",
      severity: "error",
      code: "GOJR_TEST001",
      message: "badness",
      stack: "stack line",
      span: {
        filename: "actual.go",
        line: 7,
        column: 3
      }
    })).toBe("actual.go:7:3: error GOJR_TEST001: badness\nstack line");
  });

  test("formats evaluation results for native and browser hosts", () => {
    const payload = evaluationResultToHostPayload({
      diagnostics: [],
      output: ["body\n"],
      values: [1n, "two"],
      observedDeps: [{ kind: "cell", sheet: "sheet", cell: "A1" }]
    }, ["pkg\n"]);

    expect(payload).toEqual({
      ok: true,
      incomplete: false,
      diagnostics: [],
      output: "pkg\nbody\n",
      value: "1, \"two\"",
      valueIsNil: false,
      observedDeps: [{ kind: "cell", sheet: "sheet", cell: "A1" }]
    });
  });

  test("formats spreadsheet fixture reports and package diagnostic failures", () => {
    const fixture = spreadsheetFixtureResultToHostPayload({
      ok: true,
      unstable: false,
      diagnostics: [],
      evaluated: [{ sheet: "sheet", cell: "B1" }],
      sheets: {
        sheet: {
          A1: "hi",
          B1: 3n
        }
      },
      observedDeps: {
        "sheet!B1": [{ kind: "cell", sheet: "sheet", cell: "A1" }]
      }
    });

    expect(fixture.sheets).toEqual({
      sheet: {
        A1: "\"hi\"",
        B1: "3"
      }
    });

    const failed = spreadsheetFixtureResultToHostPayload({
      ok: false,
      unstable: false,
      diagnostics: [],
      evaluated: [],
      sheets: {},
      observedDeps: {},
      packageDiagnostics: [{
        filename: "pkg.go",
        severity: "error",
        code: "GOJR_BUILD001",
        message: "missing import"
      }]
    });

    expect(failed).toMatchObject({
      ok: false,
      diagnostics: ["pkg.go: error GOJR_BUILD001: missing import"],
      sheets: {}
    });
  });

  test("derives deterministic runtime options from host environment without Node coupling", () => {
    expect(runtimeOptionsFromEnvironment(undefined, { sheet: { A1: 1n } })).toEqual({ sheet: { A1: 1n } });
    expect(runtimeOptionsFromEnvironment({ GOJR_RANDOM_SEED: "seed" })).toEqual({ randomSeed: "seed" });
    expect(runtimeOptionsFromEnvironment({ GOJR_RANDOM_SEED: "" })).toEqual({});
  });
});
