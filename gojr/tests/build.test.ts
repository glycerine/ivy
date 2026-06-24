import { describe, expect, test } from "./testHarness.js";
import fs from "node:fs";
import { arch as nodeArch, platform as nodePlatform } from "node:os";
import path from "node:path";
import {
  artifactPathForImportPath,
  buildPackages,
  buildStandardLibraryPackage,
  BuildArtifactStore,
  collectSourceImportPaths,
  createNodeSourcePackageProvider,
  createStandardLibrarySourcePackageProvider,
  evaluatePackageSourceFiles,
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

function arMemberHeaderSize(source: string | undefined, name: string): number {
  const bytes = Buffer.from(source ?? "", "utf8");
  if (bytes.subarray(0, 8).toString("ascii") !== "!<arch>\n") {
    throw new Error("missing ar header");
  }
  let offset = 8;
  while (offset < bytes.length) {
    if (offset + 60 > bytes.length) throw new Error("truncated ar member header");
    const header = bytes.subarray(offset, offset + 60).toString("ascii");
    const memberName = header.slice(0, 16).trim().replace(/\/$/, "");
    const size = Number.parseInt(header.slice(48, 58).trim(), 10);
    if (!Number.isFinite(size) || size < 0) throw new Error("bad ar member size");
    if (memberName === name) return size;
    offset += 60 + size + (size % 2);
  }
  throw new Error(`missing ar member ${name}`);
}

interface ExecutablePackageArtifactModule {
  gojrCompiledFunctionBodies: Record<string, unknown>;
  instantiateGoJrPackage(
    runtime: { evaluatePackageSourceFiles: typeof evaluatePackageSourceFiles },
    options?: Record<string, unknown>
  ): Promise<{
    diagnostics: unknown[];
    output: string[];
    compiledFunctions: Record<string, (...args: unknown[]) => Promise<unknown>>;
  }>;
}

async function importArtifactJavaScript(source: string): Promise<ExecutablePackageArtifactModule> {
  const url = `data:text/javascript;base64,${Buffer.from(source, "utf8").toString("base64")}`;
  return await import(url) as ExecutablePackageArtifactModule;
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

function testHostGOOS(): string {
  switch (nodePlatform()) {
    case "win32": return "windows";
    case "sunos": return "solaris";
    default: return nodePlatform();
  }
}

function testHostGOARCH(): string {
  switch (nodeArch()) {
    case "x64": return "amd64";
    case "ia32": return "386";
    case "mipsel": return "mipsle";
    default: return nodeArch();
  }
}

describe("Go-junior package build artifacts", () => {
  test("maps import paths to gojr_js package artifacts under a package-cache parent", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/demo/math",
      packageCacheParent: "/tmp/gopath/pkg",
      packageSources: {
        fmt: [{
          filename: "/workspace/fmt/print.go",
          source: "package fmt\n\nfunc Printf(format string, args ...any) (n int, err error) { return 0, nil }\n"
        }]
      },
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
    expect(result.artifacts).toHaveLength(2);
    expect(result.artifacts[0]?.artifactPath).toBe("/tmp/gopath/pkg/gojr_js/fmt.a");
    expect(result.artifacts[0]?.dependencies).toEqual([]);
    expect(result.artifacts[1]?.artifactPath).toBe("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a");
    expect(result.artifacts[1]?.dependencies).toEqual(["fmt"]);
    expect(result.artifacts[1]?.dependencyCacheKeys).toEqual([`fmt:${result.artifacts[0]?.cacheKey}`]);
    expect(result.artifacts[1]?.exports).toEqual([
      { name: "Add", kind: "func", typeText: "func(a int, b int) int" },
      { name: "Answer", kind: "const", typeText: "untyped int" },
      { name: "Box", kind: "type", typeText: "Box", underlyingTypeText: "struct{Value T}" },
      { name: "Count", kind: "var", typeText: "int" },
      { name: "Identity", kind: "func", typeText: "func[T any](value T) T" },
      { name: "Point", kind: "type", typeText: "Point", underlyingTypeText: "struct{X int}" }
    ]);
    expect(store.writes.has("/tmp/gopath/pkg/gojr_js/fmt.a")).toBe(true);
    expect(store.writes.has("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a")).toBe(true);
    expect((store.writes.get("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a") ?? "").slice(8, 24).trim()).toBe("__.PKGDEF");
    const archive = parseGoJuniorPackageArchive(store.writes.get("/tmp/gopath/pkg/gojr_js/example.com/demo/math.a") ?? "");
    expect(archive?.members.map((member) => member.name)).toEqual(["__.PKGDEF", "_gojr.js"]);
    const pkgdefMember = archive?.members.find((member) => member.name === "__.PKGDEF")?.data ?? "";
    const javascriptMember = archive?.members.find((member) => member.name === "_gojr.js")?.data ?? "";
    expect(javascriptMember).not.toBe(pkgdefMember);
    expect(javascriptMember).toContain("export const gojrPackageSources =");
    expect(javascriptMember).toContain("export async function instantiateGoJrPackage");
    expect(javascriptMember).toContain("export const gojrCompiledFunctionBodies =");
    expect(javascriptMember).toContain("async function gojr$Add");
    expect(javascriptMember).toContain("func Add(a, b int) int");
  });

  test("writes ar member sizes as UTF-8 byte counts for Unicode package source", () => {
    const store = new MemoryArtifactStore();
    const source = "package unicodepkg\n\n// café λ 世界\nfunc Message() string { return \"hello, λ\" }\n";
    const result = buildPackages({
      importPath: "example.com/unicodepkg",
      artifactRoot: "/tmp/gojr-unicode",
      files: [{ filename: "unicode.go", source }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const artifactSource = store.writes.get("/tmp/gojr-unicode/example.com/unicodepkg.a");
    const archive = parseGoJuniorPackageArchive(artifactSource ?? "");
    const javascriptMember = archive?.javascript ?? "";
    expect(javascriptMember).toContain("func Message() string");
    expect(javascriptMember).toContain("hello, λ");
    expect(javascriptMember).toContain("instantiateGoJrPackage");
    expect(javascriptMember).toContain("async function gojr$Message");
    expect(javascriptMember.trimEnd().endsWith("};")).toBe(true);
    expect(arMemberHeaderSize(artifactSource, "_gojr.js")).toBe(Buffer.byteLength(javascriptMember, "utf8"));
    expect(Buffer.byteLength(javascriptMember, "utf8")).toBeGreaterThan(javascriptMember.length);
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
    expect(result.source).toContain("export const gojrPackageSources =");
    expect(result.source).toContain("export const gojrCompiledFunctionBodies =");
    expect(result.source).toContain("export async function instantiateGoJrPackage");
    expect(result.source).toContain("\"importPath\": \"example.com/inspect\"");
    expect(result.source).toContain("\"name\": \"Answer\"");
    expect(result.source).toContain("async function gojr$Answer");
    expect(result.source).toContain("func Answer() int { return 42 }");
  });

  test("reads a package archive back and executes its JavaScript function body", async () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/hello",
      artifactRoot: "/tmp/gojr-hello",
      files: [{
        filename: "hello.go",
        source: "package hello\n\nfunc Hello() { print(\"hello gorj!\\n\") }\n"
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    const artifact = store.writes.get("/tmp/gojr-hello/example.com/hello.a");
    const archive = parseGoJuniorPackageArchive(artifact ?? "");
    expect(archive?.pkgdef.importPath).toBe("example.com/hello");
    expect(archive?.javascript).toContain("async function gojr$Hello");

    const artifactModule = await importArtifactJavaScript(archive?.javascript ?? "");
    expect(Object.keys(artifactModule.gojrCompiledFunctionBodies)).toEqual(["Hello"]);
    const instantiated = await artifactModule.instantiateGoJrPackage({ evaluatePackageSourceFiles });
    expect(instantiated.diagnostics).toEqual([]);
    await instantiated.compiledFunctions.Hello?.();
    expect(instantiated.output.join("")).toBe("hello gorj!\n");
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

  test("builds ambient standard runtime imports as cached dependency artifacts", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-stdlib",
      sourcePackageProvider: {
        load(importPath) {
          if (importPath !== "runtime") return undefined;
          return [{
            filename: "/usr/local/go/src/runtime/alg.go",
            source: "package runtime\nconst _ = 1 / 0\n"
          }];
        },
        isStandardLibraryPackage(importPath) {
          return importPath === "runtime";
        }
      },
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "runtime"

var OS = runtime.GOOS
var Arch = runtime.GOARCH

func Capture(buf []byte) int {
  pc, file, line, ok := runtime.Caller(0)
  if ok && pc > 0 && line > 0 && file != "" {
    return runtime.Stack(buf, false) + runtime.Callers(0, []uintptr{pc})
  }
  frames := runtime.CallersFrames([]uintptr{})
  _, _ = frames.Next()
  marker := ""
  runtime.AddCleanup(&marker, func(name string) {}, "marker").Stop()
  return runtime.Stack(buf, false)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(result.artifacts.map((artifact) => artifact.importPath)).toEqual(["runtime", "example.com/app"]);
    expect(result.artifacts[0]?.artifactPath).toBe("/tmp/gojr-stdlib/runtime.a");
    expect(result.artifacts[0]?.dependencies).toEqual([]);
    const runtimeExports = result.artifacts[0]?.exports.map((item) => item.name) ?? [];
    for (const name of [
      "AddCleanup",
      "Caller",
      "Callers",
      "CallersFrames",
      "Cleanup",
      "Error",
      "Frame",
      "Frames",
      "FuncForPC",
      "Func",
      "GOARCH",
      "GOOS",
      "Stack"
    ]) {
      expect(runtimeExports).toContain(name);
    }
    expect(result.artifacts[1]?.dependencies).toEqual(["runtime"]);
    expect(result.artifacts[1]?.dependencyCacheKeys).toEqual([`runtime:${result.artifacts[0]?.cacheKey}`]);
    expect([...store.writes.keys()]).toEqual([
      "/tmp/gojr-stdlib/runtime.a",
      "/tmp/gojr-stdlib/example.com/app.a"
    ]);
  });

  test("builds source packages without auto-importing fmt into package scope", () => {
    const store = new MemoryArtifactStore();
    const result = buildPackages({
      importPath: "fmt",
      artifactRoot: "/tmp/gojr-no-auto-fmt",
      files: [{
        filename: "/usr/local/go/src/fmt/format.go",
        source: `package fmt

type fmt struct {
  buf []byte
}

func (f *fmt) write(s string) {
  f.buf = append(f.buf, s...)
}
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(result.artifacts.map((artifact) => artifact.importPath)).toEqual(["fmt"]);
  });

  test("builds provider-supplied ordinary stdlib packages before the root", () => {
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

  test("builds runtime/pprof from Go-junior stub source before consulting providers", () => {
    const store = new MemoryArtifactStore();
    const loaded: string[] = [];
    const result = buildPackages({
      importPath: "example.com/app",
      artifactRoot: "/tmp/gojr-pprof-stub",
      sourcePackageProvider: {
        load(importPath) {
          loaded.push(importPath);
          if (importPath === "context") {
            return [{
              filename: "/workspace/context/context.go",
              source: "package context\n\ntype Context interface{}\n"
            }];
          }
          if (importPath === "io") {
            return [{
              filename: "/workspace/io/io.go",
              source: "package io\n\ntype Writer interface { Write([]byte) (int, error) }\n"
            }];
          }
          if (importPath === "runtime/pprof") {
            return [{
              filename: "/usr/local/go/src/runtime/pprof/pprof.go",
              source: "package pprof\n\nconst _ = 1 / 0\n"
            }];
          }
          return undefined;
        },
        isStandardLibraryPackage(importPath) {
          return importPath === "context" || importPath === "io" || importPath === "runtime/pprof";
        }
      },
      files: [{
        filename: "/workspace/example.com/app/app.go",
        source: `package app

import "runtime/pprof"

func F() { _ = pprof.Lookup("heap") }
`
      }]
    }, store);

    expect(result.diagnostics).toEqual([]);
    expect(result.ok).toBe(true);
    expect(loaded).toEqual(["context", "io"]);
    expect(result.artifacts.map((artifact) => artifact.importPath)).toEqual([
      "context",
      "io",
      "runtime/pprof",
      "example.com/app"
    ]);
    const pprof = artifactJSON(store.writes.get("/tmp/gojr-pprof-stub/runtime/pprof.a"));
    expect(pprof.importPath).toBe("runtime/pprof");
    expect(pprof.standardLibrary).toBe(true);
    expect(pprof.sources).toEqual([{
      filename: "gojr:stub/runtime/pprof/pprof.go",
      hash: expect.any(String)
    }]);
    expect(result.artifacts.find((artifact) => artifact.importPath === "runtime/pprof")?.exports.map((item) => item.name)).toContain("StartCPUProfile");
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

  test("node standard-library provider uses native host build tags, not the GoJr output target", () => {
    const goroot = fs.mkdtempSync(path.join("/tmp", "gojr-goroot-"));
    const depDir = path.join(goroot, "src", "internal", "bytealg");
    fs.mkdirSync(depDir, { recursive: true });
    fs.writeFileSync(path.join(depDir, "bytealg.go"), "package bytealg\n");
    const hostGOOS = testHostGOOS();
    const hostGOARCH = testHostGOARCH();
    fs.writeFileSync(
      path.join(depDir, "native.go"),
      `//go:build ${hostGOOS} && ${hostGOARCH}\n\npackage bytealg\n\nfunc NativeTag() int { return 1 }\n`
    );
    fs.writeFileSync(
      path.join(depDir, "unix.go"),
      "//go:build unix\n\npackage bytealg\n\nfunc UnixTag() int { return 3 }\n"
    );
    fs.writeFileSync(
      path.join(depDir, "target_gojr_js.go"),
      "package bytealg\n\nfunc TargetTag() int { return 2 }\n"
    );
    fs.writeFileSync(
      path.join(depDir, `target_${["w", "a", "s", "m"].join("")}.go`),
      "package bytealg\n\nfunc OtherTargetTag() int { return 4 }\n"
    );

    const previousGOROOT = process.env.GOROOT;
    process.env.GOROOT = goroot;
    try {
      const provider = createNodeSourcePackageProvider([]);
      const files = provider?.load("internal/bytealg") ?? [];
      expect(files.map((file) => path.basename(file.filename)).sort()).toEqual([
        "bytealg.go",
        "native.go",
        "unix.go"
      ]);
    } finally {
      if (previousGOROOT === undefined) {
        delete process.env.GOROOT;
      } else {
        process.env.GOROOT = previousGOROOT;
      }
    }
  });

  test("package names come from parsed package clauses, not package comments", async () => {
    const result = await evaluatePackageSourceFiles([
      {
        filename: "/tmp/os/signal/doc.go",
        source: `/*
Package signal documents behavior for package on Unix systems.
*/
package signal

var X = 1
`
      },
      {
        filename: "/tmp/os/signal/signal.go",
        source: `package signal

var Y = 2
`
      }
    ], { importPath: "os/signal" });

    expect(result.diagnostics).toEqual([]);
    expect(result.package?.X).toBe(1n);
    expect(result.package?.Y).toBe(2n);
  });

  test("node source provider resolves packages from module roots", () => {
    const root = fs.mkdtempSync(path.join("/tmp", "gojr-module-root-"));
    const depDir = path.join(root, "sub", "pkg");
    fs.mkdirSync(depDir, { recursive: true });
    fs.writeFileSync(path.join(root, "go.mod"), "module example.com/mod/v2\n\ngo 1.27\n");
    fs.writeFileSync(path.join(depDir, "pkg.go"), "package pkg\n\nfunc One() int { return 1 }\n");

    const provider = createNodeSourcePackageProvider([root]);
    const files = provider?.load("example.com/mod/v2/sub/pkg") ?? [];

    expect(files.map((file) => path.relative(root, file.filename))).toEqual([path.join("sub", "pkg", "pkg.go")]);
    expect(provider?.load("example.com/other/sub/pkg")).toBe(undefined);
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
