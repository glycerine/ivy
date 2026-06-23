import { describe, expect, test } from "./testHarness.js";
import {
  compilePackageSourceFiles,
  compileSource,
  compileSourceFiles
} from "../src/index.js";

describe("Go-junior static compile API", () => {
  test("typechecks formula snippets without evaluating them", () => {
    const result = compileSource("sheet.A1 + 2", {
      sheet: {
        A1: 40n
      }
    });

    expect(result.diagnostics).toEqual([]);
  });

  test("reports type errors without running source", () => {
    const result = compileSource(`
var a = 10
var s = "hi"
fmt.Printf("this should not run")
return a + s
`);

    expect(result.diagnostics.map((diagnostic) => diagnostic.code)).toEqual(["GOJR_TYPE001"]);
    expect(result.diagnostics[0]?.message).toContain("mismatched types int64 and string");
  });

  test("typechecks package source files", () => {
    const result = compileSourceFiles([{
      filename: "calc.go",
      source: `package calc

func Add(a, b int) int {
  return a + b
}
`
    }]);

    expect(result.diagnostics).toEqual([]);
  });

  test("typechecks packages against imported package metadata without evaluation", () => {
    const lib = compilePackageSourceFiles([{
      filename: "lib.go",
      source: `package lib

func One() int { return 1 }
`
    }], { importPath: "example.com/lib" });
    expect(lib.diagnostics).toEqual([]);

    const app = compilePackageSourceFiles([{
      filename: "app.go",
      source: `package app

import lib "example.com/lib"

func Two() int {
  return lib.One() + 1
}
`
    }], {
      importPath: "example.com/app",
      packageInfos: {
        "example.com/lib": lib.packageInfo
      }
    });
    expect(app.diagnostics).toEqual([]);

    const bad = compileSourceFiles([{
      filename: "formula.go",
      source: `import app "example.com/app"

return app.Nope()
`
    }], {
      packageInfos: {
        "example.com/app": app.packageInfo
      }
    });
    expect(bad.diagnostics.map((diagnostic) => diagnostic.code)).toEqual(["GOJR_TYPE001", "GOJR_TYPE001"]);
    expect(bad.diagnostics.map((diagnostic) => diagnostic.message).join("\n")).toContain("Nope");
  });
});
