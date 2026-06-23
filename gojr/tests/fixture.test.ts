import { describe, expect, test } from "./testHarness.js";
import {
  cellDependency,
  collectSpreadsheetFixtureFormulaSourceFiles,
  evaluatePackageSourceFiles,
  parseSpreadsheetFixtureJson,
  rangeDependency,
  runSpreadsheetFixture
} from "../src/index.js";

describe("Go-junior spreadsheet fixture integration", () => {
  test("evaluates Go-junior formula cells through the spreadsheet graph", async () => {
    const result = await runSpreadsheetFixture({
      cells: {
        A1: 40n,
        B1: { formula: "sheet.A1 + 2" },
        C1: { formula: "sheet.B1 * 2" }
      }
    });

    expect(result.ok).toBe(true);
    expect(result.sheets.sheet?.B1).toBe(42n);
    expect(result.sheets.sheet?.C1).toBe(84n);
    expect(result.observedDeps["sheet!B1"]).toEqual([cellDependency({ sheet: "sheet", cell: "A1" })]);
    expect(result.observedDeps["sheet!C1"]).toEqual([cellDependency({ sheet: "sheet", cell: "B1" })]);
  });

  test("captures dynamic dependencies from the branch that actually ran", async () => {
    const trueBranch = await runSpreadsheetFixture({
      cells: {
        A1: true,
        B1: 10n,
        C1: 100n,
        D1: { formula: "if sheet.A1 { return sheet.B1 }\nreturn sheet.C1" }
      }
    });

    expect(trueBranch.sheets.sheet?.D1).toBe(10n);
    expect(trueBranch.observedDeps["sheet!D1"]).toEqual([
      cellDependency({ sheet: "sheet", cell: "A1" }),
      cellDependency({ sheet: "sheet", cell: "B1" })
    ]);

    const falseBranch = await runSpreadsheetFixture({
      cells: {
        A1: false,
        B1: 10n,
        C1: 100n,
        D1: { formula: "if sheet.A1 { return sheet.B1 }\nreturn sheet.C1" }
      }
    });

    expect(falseBranch.sheets.sheet?.D1).toBe(100n);
    expect(falseBranch.observedDeps["sheet!D1"]).toEqual([
      cellDependency({ sheet: "sheet", cell: "A1" }),
      cellDependency({ sheet: "sheet", cell: "C1" })
    ]);
  });

  test("captures range and cross-sheet dependencies", async () => {
    const result = await runSpreadsheetFixture({
      currentSheetName: "Report",
      sheets: {
        Data: {
          A1: 1n,
          A2: 2n
        },
        Report: {
          A1: { formula: "rows := Data.A1:A2\nreturn rows[0][0] + rows[1][0]" }
        }
      }
    });

    expect(result.ok).toBe(true);
    expect(result.sheets.Report?.A1).toBe(3n);
    expect(result.observedDeps["Report!A1"]).toEqual([rangeDependency("Data", "A1", "A2")]);
  });

  test("lets later formulas index range values returned by earlier formulas", async () => {
    const result = await runSpreadsheetFixture({
      cells: {
        A1: 1n,
        A2: 2n,
        B1: { formula: "return sheet.A1:A2" },
        C1: { formula: "return sheet.B1[0][0] + sheet.B1[1][0]" }
      }
    });

    expect(result.ok).toBe(true);
    expect(result.sheets.sheet?.B1).toEqual([[1n], [2n]]);
    expect(result.sheets.sheet?.C1).toBe(3n);
    expect(result.observedDeps["sheet!B1"]).toEqual([rangeDependency("sheet", "A1", "A2")]);
    expect(result.observedDeps["sheet!C1"]).toEqual([cellDependency({ sheet: "sheet", cell: "B1" })]);
  });

  test("evaluates formula cells that import source package metadata", async () => {
    const pkg = await evaluatePackageSourceFiles([{
      filename: "mathx.go",
      source: `package mathx

func Add(a, b int) int {
  return a + b
}
`
    }], { importPath: "example.com/mathx" });
    expect(pkg.diagnostics).toEqual([]);

    const fixture = {
      cells: {
        A1: 40n,
        B1: { formula: `import mathx "example.com/mathx"\nreturn mathx.Add(sheet.A1, 2)` }
      }
    };
    expect(collectSpreadsheetFixtureFormulaSourceFiles(fixture)).toEqual([{
      filename: "sheet!B1.gojr",
      source: `import mathx "example.com/mathx"\nreturn mathx.Add(sheet.A1, 2)`
    }]);

    const result = await runSpreadsheetFixture(fixture, {
      packages: {
        "example.com/mathx": pkg.package ?? {}
      },
      packageInfos: {
        "example.com/mathx": pkg.packageInfo!
      }
    });

    expect(result.ok).toBe(true);
    expect(result.sheets.sheet?.B1).toBe(42n);
    expect(result.observedDeps["sheet!B1"]).toEqual([cellDependency({ sheet: "sheet", cell: "A1" })]);
  });

  test("parses fixture JSON with integer cells preserved as int64 values", async () => {
    const fixture = parseSpreadsheetFixtureJson(`{
      "cells": {
        "A1": 40,
        "B1": { "formula": "sheet.A1 + 2" }
      }
    }`);
    const result = await runSpreadsheetFixture(fixture);

    expect(result.ok).toBe(true);
    expect(result.sheets.sheet?.A1).toBe(40n);
    expect(result.sheets.sheet?.B1).toBe(42n);
  });
});
