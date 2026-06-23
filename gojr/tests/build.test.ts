import { describe, expect, test } from "./testHarness.js";
import { artifactPathForImportPath, buildPackages, BuildArtifactStore, resolveArtifactRoot } from "../src/index.js";

class MemoryArtifactStore implements BuildArtifactStore {
  public readonly writes = new Map<string, string>();

  public writeAtomic(path: string, source: string): void {
    this.writes.set(path, source);
  }
}

describe("Go-junior package build artifacts", () => {
  test("maps import paths to gojr_js package artifacts under a package-cache parent", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/demo/math",
      packageCacheParent: "/tmp/gopath/pkg",
      files: [
        {
          filename: "/tmp/demo/math/math.go",
          source: `package math

import "fmt"

const Answer = 42
var Count int
type Point struct { X int }
func Add(a, b int) int { fmt.Printf("%v", a); return a + b }
func hidden() {}
`
        }
      ]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(result.artifacts).toHaveLength(1);
    expect(result.artifacts[0]?.artifactPath).toBe("/tmp/gopath/pkg/gojr_js/example.com/demo/math.js");
    expect(result.artifacts[0]?.dependencies).toEqual(["fmt"]);
    expect(result.artifacts[0]?.exports).toEqual([
      { name: "Add", kind: "func" },
      { name: "Answer", kind: "const" },
      { name: "Count", kind: "var" },
      { name: "Point", kind: "type" }
    ]);
    expect(store.writes.has("/tmp/gopath/pkg/gojr_js/example.com/demo/math.js")).toBe(true);
    expect(store.writes.get("/tmp/gopath/pkg/gojr_js/example.com/demo/math.js")).toContain("gojrPackageArtifact");
  });

  test("uses an exact artifact root override directly", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/demo",
      artifactRoot: "/tmp/custom-cache",
      files: [{ filename: "demo.go", source: "package demo\nfunc F() {}\n" }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.artifacts[0]?.artifactPath).toBe("/tmp/custom-cache/example.com/demo.js");
    expect([...store.writes.keys()]).toEqual(["/tmp/custom-cache/example.com/demo.js"]);
  });

  test("rejects mixed package names before writing artifacts", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/bad",
      packageCacheParent: "/tmp/gopath/pkg",
      files: [
        { filename: "a.go", source: "package one\nfunc A() {}\n" },
        { filename: "b.go", source: "package two\nfunc B() {}\n" }
      ]
    }, store);

    expect(result.ok).toBe(false);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_BUILD001");
    expect(result.diagnostics[0]?.filename).toBe("b.go");
    expect(result.diagnostics[0]?.message).toContain("package two does not match package one");
    expect([...store.writes.keys()]).toEqual([]);
  });

  test("documents artifact root helpers", () => {
    expect(resolveArtifactRoot({ packageCacheParent: "/home/me/go/pkg" })).toBe("/home/me/go/pkg/gojr_js");
    expect(resolveArtifactRoot({ artifactRoot: "/tmp/gojr_js" })).toBe("/tmp/gojr_js");
    expect(artifactPathForImportPath("/home/me/go/pkg/gojr_js", "github.com/u/p")).toBe("/home/me/go/pkg/gojr_js/github.com/u/p.js");
  });
});
