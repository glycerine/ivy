import { describe, expect, test } from "./testHarness.js";
import {
  benchmarkGoJuniorOnNode,
  benchmarkGoJuniorOnNodeHostPayload,
  formatGoJuniorBenchmarkReport,
  runGoJuniorBenchmark
} from "../src/index.js";

describe("Go-junior benchmark harness", () => {
  test("measures front-end and package artifact phases for built-in cases", async () => {
    const report = await runGoJuniorBenchmark(undefined, {
      caseNames: ["tiny-function"],
      iterations: 1,
      phaseSet: "front-end-and-build"
    });

    expect(report.ok).toBe(true);
    expect(report.cases).toHaveLength(1);
    expect(report.cases[0]?.name).toBe("tiny-function");
    expect(report.cases[0]?.phases.map((phase) => phase.name)).toEqual([
      "parse",
      "ast-lower",
      "typecheck",
      "package-build"
    ]);
    expect(report.cases[0]?.metrics.artifactBytesMax).toBeGreaterThan(0);
    expect(report.cases[0]?.metrics.pkgdefBytesMax).toBeGreaterThan(0);
    expect(formatGoJuniorBenchmarkReport(report)).toContain("gojr bench: iterations=1");
  });

  test("node benchmark host payload includes formatted text for native wrapper output", async () => {
    const report = await benchmarkGoJuniorOnNode({
      caseNames: ["loop-and-branch"],
      iterations: 1,
      phaseSet: "front-end"
    });
    const payload = benchmarkGoJuniorOnNodeHostPayload(report);

    expect(payload.ok).toBe(true);
    expect(payload.diagnostics).toEqual([]);
    expect(payload.output).toContain("case loop-and-branch");
    expect(payload.output).toContain("typecheck:");
  });

  test("warm cache mode records cacheable package build timing", async () => {
    const report = await runGoJuniorBenchmark(undefined, {
      caseNames: ["tiny-function"],
      iterations: 2,
      warmupIterations: 1,
      cacheMode: "warm",
      phaseSet: "package-build"
    });
    const phase = report.cases[0]?.phases.find((item) => item.name === "package-build");

    expect(report.ok).toBe(true);
    expect(phase?.iterations).toBe(2);
    expect(report.cacheMode).toBe("warm");
  });
});
