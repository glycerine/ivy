import { describe, expect, test } from "./testHarness.js";
import fs from "node:fs";
import path from "node:path";
import {
  artifactPathForImportPath,
  buildPackages,
  buildStandardLibraryPackage,
  BuildArtifactStore,
  collectSourceImportPaths,
  createNodeSourcePackageProvider,
  createStandardLibrarySourcePackageProvider,
  inspectPackageJavaScript,
  parseGoJuniorPackageArchive,
  resolveArtifactRoot
} from "../src/index.js";

class MemoryArtifactStore implements BuildArtifactStore {
  public readonly writes = new Map<string, string>();
  public writeCount = 0;

  public read(path: string): string | undefined {
    return this.writes.get(path);
  }

  public writeAtomic(path: string, source: string): void {
    this.writeCount++;
    this.writes.set(path, source);
  }
}

function artifactJSON(source: string | undefined): Record<string, unknown> {
  const archive = parseGoJuniorPackageArchive(source ?? "");
  if (archive) return archive.pkgdef as unknown as Record<string, unknown>;
  const match = /export const gojrPackageArtifact = ([\s\S]*);\s*$/.exec(source ?? "");
  if (!match?.[1]) throw new Error("missing gojrPackageArtifact envelope");
  return JSON.parse(match[1]) as Record<string, unknown>;
}

const nodeSourceHost = {
  readDir(dir: string) {
    return fs.readdirSync(dir, { withFileTypes: true }).map((entry) => ({
      name: entry.name,
      isFile: entry.isFile()
    }));
  },
  readFile(filename: string) {
    return fs.readFileSync(filename, "utf8");
  }
};

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
type Box[T any] struct { Value T }
func Add(a, b int) int { fmt.Printf("%v", a); return a + b }
func Identity[T any](value T) T { return value }
func hidden() {}
`
        }
      ]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(result.artifacts).toHaveLength(1);
    expect(result.artifacts[0]?.artifactPath).toBe("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a");
    expect(result.artifacts[0]?.dependencies).toEqual(["fmt"]);
    expect(result.artifacts[0]?.exports).toEqual([
      { name: "Add", kind: "func", typeText: "func(a int, b int) int" },
      { name: "Answer", kind: "const", typeText: "untyped int" },
      { name: "Box", kind: "type", typeText: "Box", underlyingTypeText: "struct{Value T}" },
      { name: "Count", kind: "var", typeText: "int" },
      { name: "Identity", kind: "func", typeText: "func[T any](value T) T" },
      { name: "Point", kind: "type", typeText: "Point", underlyingTypeText: "struct{X int}" }
    ]);
    expect(store.writes.has("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a")).toBe(true);
    expect(parseGoJuniorPackageArchive(store.writes.get("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a") ?? "")?.members[0]?.name).toBe("__.PKGDEF");
  });

  test("uses an exact artifact root override directly", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/demo",
      artifactRoot: "/tmp/custom-cache",
      files: [{ filename: "demo.go", source: "package demo\nfunc F() {}\n" }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.artifacts[0]?.artifactPath).toBe("/tmp/custom-cache/example.com/demo.a");
    expect([...store.writes.keys()]).toEqual(["/tmp/custom-cache/example.com/demo.a"]);
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
    expect(artifactPathForImportPath("/home/me/go/pkg/gojr_js", "github.com/u/p")).toBe("/home/me/go/pkg/gojr_js/github.com/u/p.a");
  });

  test("builds leaf standard-library packages into the GOPATH-style gojr_js cache", () => {
    const goSourceRoot = "/usr/local/go1.27rc1/src";
    expect(fs.existsSync(path.join(goSourceRoot, "cmp"))).toBe(true);
    expect(fs.existsSync(path.join(goSourceRoot, "unsafe"))).toBe(true);

    const store = new MemoryArtifactStore();
    const provider = createStandardLibrarySourcePackageProvider({
      sourceRoot: goSourceRoot,
      host: nodeSourceHost
    });

    const cmpFirst = buildStandardLibraryPackage({
      importPath: "cmp",
      packageCacheParent: "/tmp/gopath/pkg",
      standardLibrary: {
        sourceRoot: goSourceRoot,
        host: nodeSourceHost
      },
      sourcePackageProvider: provider
    }, store);
    const unsafeFirst = buildStandardLibraryPackage({
      importPath: "unsafe",
      packageCacheParent: "/tmp/gopath/pkg",
      standardLibrary: {
        sourceRoot: goSourceRoot,
        host: nodeSourceHost
      },
      sourcePackageProvider: provider
    }, store);
    const cmpSecond = buildStandardLibraryPackage({
      importPath: "cmp",
      packageCacheParent: "/tmp/gopath/pkg",
      standardLibrary: {
        sourceRoot: goSourceRoot,
        host: nodeSourceHost
      },
      sourcePackageProvider: provider
    }, store);
    const unsafeSecond = buildStandardLibraryPackage({
      importPath: "unsafe",
      packageCacheParent: "/tmp/gopath/pkg",
      standardLibrary: {
        sourceRoot: goSourceRoot,
        host: nodeSourceHost
      },
      sourcePackageProvider: provider
    }, store);

    expect(cmpFirst.diagnostics).toEqual([]);
    expect(cmpFirst.ok).toBe(true);
    expect(cmpFirst.artifacts).toHaveLength(1);
    expect(cmpFirst.artifacts[0]?.artifactPath).toBe("/tmp/gopath/pkg/gojr_js/cmp.a");
    expect(cmpFirst.artifacts[0]?.dependencies).toEqual([]);
    expect(cmpFirst.artifacts[0]?.exports.map((item) => item.name)).toEqual(["Compare", "Less", "Or", "Ordered"]);
    expect(unsafeFirst.diagnostics).toEqual([]);
    expect(unsafeFirst.ok).toBe(true);
    expect(unsafeFirst.artifacts[0]?.artifactPath).toBe("/tmp/gopath/pkg/gojr_js/unsafe.a");
    expect(unsafeFirst.artifacts[0]?.exports.map((item) => item.name)).toContain("Pointer");
    expect(unsafeFirst.artifacts[0]?.exports.map((item) => item.name)).toContain("Sizeof");
    expect(cmpSecond.artifacts[0]?.action).toBe("skipped");
    expect(unsafeSecond.artifacts[0]?.action).toBe("skipped");
    expect(store.writeCount).toBe(2);

    const artifact = artifactJSON(store.writes.get("/tmp/gopath/pkg/gojr_js/cmp.a"));
    expect(artifact.goos).toBe("gojr");
    expect(artifact.goarch).toBe("js");
    expect(artifact.importPath).toBe("cmp");
    expect(artifact.standardLibrary).toBe(true);
    expect(artifact.buildTags).toEqual(expect.any(Array));
    expect(artifact.sources).toEqual([
      {
        filename: "/usr/local/go1.27rc1/src/cmp/cmp.go",
        hash: expect.any(String)
      }
    ]);

    const unsafeArtifact = artifactJSON(store.writes.get("/tmp/gopath/pkg/gojr_js/unsafe.a"));
    expect(unsafeArtifact.goos).toBe("gojr");
    expect(unsafeArtifact.goarch).toBe("js");
    expect(unsafeArtifact.importPath).toBe("unsafe");
    expect(unsafeArtifact.standardLibrary).toBe(true);
    expect(unsafeArtifact.sources).toEqual([
      {
        filename: "/usr/local/go1.27rc1/src/unsafe/unsafe.go",
        hash: expect.any(String)
      }
    ]);
  });

  test("inspects package artifact JavaScript without a cache store", () => {
    const result = inspectPackageJavaScript({
      importPath: "example.com/inspect",
      artifactRoot: "/tmp/gojr-inspect",
      files: [{ filename: "inspect.go", source: "package inspect\n\nfunc Answer() int { return 42 }\n" }]
    });

    expect(result.ok).toBe(true);
    expect(result.built).toEqual(["/tmp/gojr-inspect/example.com/inspect.a"]);
    expect(result.source).toContain("export const gojrPackageArtifact =");
    expect(result.source).toContain("\"importPath\": \"example.com/inspect\"");
    expect(result.source).toContain("\"name\": \"Answer\"");
  });

  test("skips fresh package artifacts with the same cache key", () => {
    const store = new MemoryArtifactStore();
    const request = {
      importPath: "example.com/cache",
      artifactRoot: "/tmp/gojr-cache",
      files: [{ filename: "cache.go", source: "package cache\nfunc F() int { return 1 }\n" }]
    };

    const first = buildPackages(request, store);
    const second = buildPackages(request, store);

    expect(first.ok).toBe(true);
    expect(second.ok).toBe(true);
    expect(first.artifacts[0]?.action).toBe("built");
    expect(second.artifacts[0]?.action).toBe("skipped");
    expect(first.built).toEqual(["/tmp/gojr-cache/example.com/cache.a"]);
    expect(second.built).toEqual([]);
    expect(second.skipped).toEqual(["/tmp/gojr-cache/example.com/cache.a"]);
    expect(store.writeCount).toBe(1);
    expect(first.artifacts[0]?.cacheKey).toBe(second.artifacts[0]?.cacheKey);
  });

  test("rebuilds package artifacts when cache key inputs change", () => {
    const store = new MemoryArtifactStore();
    const baseRequest = {
      importPath: "example.com/cache",
      artifactRoot: "/tmp/gojr-cache",
      files: [{ filename: "cache.go", source: "package cache\nfunc F() int { return 1 }\n" }]
    };

    const first = buildPackages(baseRequest, store);
    const hostChanged = buildPackages({ ...baseRequest, hostSpecVersion: "host-v2" }, store);
    const compilerChanged = buildPackages({ ...baseRequest, compilerVersion: "compiler-v2" }, store);

    expect(first.artifacts[0]?.action).toBe("built");
    expect(hostChanged.artifacts[0]?.action).toBe("built");
    expect(compilerChanged.artifacts[0]?.action).toBe("built");
    expect(hostChanged.artifacts[0]?.cacheKey).not.toBe(first.artifacts[0]?.cacheKey);
    expect(compilerChanged.artifacts[0]?.cacheKey).not.toBe(hostChanged.artifacts[0]?.cacheKey);
    expect(store.writeCount).toBe(3);
  });

  test("rebuilds package artifacts when source changes or existing cache is invalid", () => {
    const store = new MemoryArtifactStore();
    const request = {
      importPath: "example.com/cache",
      artifactRoot: "/tmp/gojr-cache",
      files: [{ filename: "cache.go", source: "package cache\nfunc F() int { return 1 }\n" }]
    };

    const first = buildPackages(request, store);
    const sourceChanged = buildPackages({
      ...request,
      files: [{ filename: "cache.go", source: "package cache\nfunc F() int { return 2 }\n" }]
    }, store);
    store.writes.set("/tmp/gojr-cache/example.com/cache.a", "not a gojr artifact");
    const invalidCache = buildPackages(request, store);

    expect(first.artifacts[0]?.action).toBe("built");
    expect(sourceChanged.artifacts[0]?.action).toBe("built");
    expect(sourceChanged.artifacts[0]?.cacheKey).not.toBe(first.artifacts[0]?.cacheKey);
    expect(invalidCache.artifacts[0]?.action).toBe("built");
    expect(store.writeCount).toBe(3);
  });

  test("builds source package dependencies before the root and keys roots by dependency cache keys", () => {
    const store = new MemoryArtifactStore();
    const request = {
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-graph",
      packageSources: {
        "example.com/lib": [{
          filename: "lib.go",
          source: "package lib\n\nfunc One() int { return 1 }\n"
        }]
      },
      files: [{
        filename: "app.go",
        source: `package app

import lib "example.com/lib"

func Two() int { return lib.One() + 1 }
`
      }]
    };

    const first = buildPackages(request, store);
    const second = buildPackages(request, store);
    const changedDependency = buildPackages({
      ...request,
      packageSources: {
        "example.com/lib": [{
          filename: "lib.go",
          source: "package lib\n\nfunc One() int { return 2 }\n"
        }]
      }
    }, store);

    expect(first.diagnostics).toEqual([]);
    expect(first.ok).toBe(true);
    expect(first.artifacts.map((artifact) => artifact.importPath)).toEqual(["example.com/lib", "example.com/app"]);
    expect(first.built).toEqual([
      "/tmp/gojr-graph/example.com/lib.a",
      "/tmp/gojr-graph/example.com/app.a"
    ]);
    expect(first.artifacts[1]?.dependencies).toEqual(["example.com/lib"]);
    expect(first.artifacts[1]?.dependencyCacheKeys).toEqual([`example.com/lib:${first.artifacts[0]?.cacheKey}`]);
    expect(first.artifacts[1]?.exports).toEqual([
      { name: "Two", kind: "func", typeText: "func() int" }
    ]);

    expect(second.artifacts.map((artifact) => artifact.action)).toEqual(["skipped", "skipped"]);
    expect(second.built).toEqual([]);
    expect(second.skipped).toEqual([
      "/tmp/gojr-graph/example.com/lib.a",
      "/tmp/gojr-graph/example.com/app.a"
    ]);

    expect(changedDependency.artifacts.map((artifact) => artifact.action)).toEqual(["built", "built"]);
    expect(changedDependency.artifacts[0]?.cacheKey).not.toBe(first.artifacts[0]?.cacheKey);
    expect(changedDependency.artifacts[1]?.cacheKey).not.toBe(first.artifacts[1]?.cacheKey);
    expect(changedDependency.artifacts[1]?.dependencyCacheKeys).toEqual([`example.com/lib:${changedDependency.artifacts[0]?.cacheKey}`]);
  });

  test("rejects unknown source package imports before writing artifacts", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-graph",
      files: [{
        filename: "app.go",
        source: `package app

import "example.com/missing"

func F() {}
`
      }]
    }, store);

    expect(result.ok).toBe(false);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_BUILD001");
    expect(result.diagnostics[0]?.message).toContain("example.com/missing");
    expect([...store.writes.keys()]).toEqual([]);
  });

  test("rejects source package import cycles before writing artifacts", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/a",
      artifactRoot: "/tmp/gojr-graph",
      packageSources: {
        "example.com/b": [{
          filename: "b.go",
          source: `package b

import "example.com/a"

func B() {}
`
        }]
      },
      files: [{
        filename: "a.go",
        source: `package a

import "example.com/b"

func A() {}
`
      }]
    }, store);

    expect(result.ok).toBe(false);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_BUILD001");
    expect(result.diagnostics[0]?.message).toContain("package import cycle detected");
    expect([...store.writes.keys()]).toEqual([]);
  });

  test("loads source package dependencies from a build provider", () => {
    const store = new MemoryArtifactStore();
    const loaded: string[] = [];
    const result = buildPackages({
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-provider",
      sourcePackageProvider: {
        load(importPath) {
          loaded.push(importPath);
          if (importPath !== "example.com/lib") return undefined;
          return [{
            filename: "/workspace/example.com/lib/lib.go",
            source: "package lib\n\nfunc One() int { return 1 }\n"
          }];
        }
      },
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import lib "example.com/lib"

func Two() int { return lib.One() + 1 }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(loaded).toEqual(["example.com/lib"]);
    expect(result.artifacts.map((artifact) => artifact.importPath)).toEqual(["example.com/lib", "example.com/app"]);
    expect([...store.writes.keys()]).toEqual([
      "/tmp/gojr-provider/example.com/lib.a",
      "/tmp/gojr-provider/example.com/app.a"
    ]);
  });

  test("builds provider-supplied ambient packages before the root", () => {
    const store = new MemoryArtifactStore();
    const loaded: string[] = [];
    const result = buildPackages({
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-provider",
      sourcePackageProvider: {
        load(importPath) {
          loaded.push(importPath);
          if (importPath !== "fmt") return undefined;
          return [{
            filename: "/workspace/fmt/print.go",
            source: "package fmt\n\nfunc Println(s string) {}\n"
          }];
        }
      },
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "fmt"

func F() { fmt.Println("hi") }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(loaded).toEqual(["fmt"]);
    expect(result.artifacts.map((artifact) => artifact.importPath)).toEqual(["fmt", "example.com/app"]);
    expect([...store.writes.keys()]).toEqual([
      "/tmp/gojr-provider/fmt.a",
      "/tmp/gojr-provider/example.com/app.a"
    ]);
  });

  test("node source provider filters files excluded by build constraints", () => {
    const root = fs.mkdtempSync(path.join("/tmp", "gojr-srcroot-"));
    const depDir = path.join(root, "example.com", "dep");
    fs.mkdirSync(depDir, { recursive: true });
    fs.writeFileSync(path.join(depDir, "dep.go"), "package dep\n\nfunc One() int { return 1 }\n");
    fs.writeFileSync(path.join(depDir, "tools.go"), "// +build tools\n\npackage dep\n\nimport _ \"example.com/missing\"\n");

    const provider = createNodeSourcePackageProvider([root]);
    const files = provider?.load("example.com/dep") ?? [];

    expect(files.map((file) => path.basename(file.filename))).toEqual(["dep.go"]);
  });

  test("reports source package provider failures as build diagnostics", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-provider",
      sourcePackageProvider: {
        load() {
          throw new Error("permission denied");
        }
      },
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "example.com/lib"

func F() {}
`
      }]
    }, store);

    expect(result.ok).toBe(false);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_BUILD001");
    expect(result.diagnostics[0]?.message).toContain("could not load package example.com/lib");
    expect(result.diagnostics[0]?.message).toContain("permission denied");
    expect([...store.writes.keys()]).toEqual([]);
  });

  test("collects import paths from formula and package source files", () => {
    const result = collectSourceImportPaths([{
      filename: "formula.go",
      source: `import (
  "fmt"
  app "example.com/app"
)

return app.Two()
`
    }, {
      filename: "pkg.go",
      source: `package pkg

import "example.com/lib"

func F() {}
`
    }]);

    expect(result.diagnostics).toEqual([]);
    expect(result.imports).toEqual(["example.com/app", "example.com/lib", "fmt"]);
  });
});
