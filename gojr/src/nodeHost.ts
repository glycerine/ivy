import { existsSync, mkdirSync, readFileSync, readdirSync, renameSync, rmSync, statSync, writeFileSync } from "node:fs";
import { basename, delimiter, dirname, isAbsolute, join, relative, resolve, sep } from "node:path";
import { arch as nodeArch, homedir, platform as nodePlatform } from "node:os";
import {
  buildTagSetForContext,
  buildPackages,
  collectSourceImportPaths,
  GOJR_GOARCH,
  GOJR_GOOS,
  goSourceFileMatchesBuildContext,
  inspectPackageJavaScript,
  parseGoJuniorPackageArchive,
  resolveArtifactRoot,
  type BuildArtifactStore,
  type BuildExport,
  type BuildPackageReport,
  type BuildPackageRequest,
  type BuildSourcePackageProvider,
  type InspectPackageJavaScriptReport
} from "./build.js";
import { hasErrorDiagnostics, REPL_FILENAME, type Diagnostic, type SourceFile } from "./diagnostics.js";
import { formatBuildProgressEvent, formatEvaluationProgressEvent } from "./hostProtocol.js";
import {
  compilePackageSourceFiles,
  compileSourceFiles,
  type CompileResult
} from "./compile.js";
import {
  collectSpreadsheetFixtureFormulaSourceFiles,
  parseSpreadsheetFixtureJson,
  runSpreadsheetFixture,
  type SpreadsheetFixtureRunResult
} from "./fixture.js";
import { isHostResolvedSourceImport } from "./intrinsicPackages.js";
import { stubSourcePackageFiles } from "./stubPackages.js";
import { parseSheetJson, parseSheetsJson } from "./jsonInput.js";
import {
  evaluateSource,
  evaluateSourceFiles,
  evaluateSourcePackageGraph,
  runMainSourcePackageFiles,
  testSourceFiles,
  type EvaluationContext,
  type EvaluationOptions,
  type EvaluationResult,
  type RuntimeObject,
  type SourcePackageSpec
} from "./runtime.js";
import type { Package as GoTypesPackage } from "./go/types/index.js";

export interface NodeBuildPackageRequest extends BuildPackageRequest {
  sourceRoots?: string[];
  progress?: boolean;
}

export interface NodeSourcePackageRequest {
  importPath?: string;
  packageName?: string;
  source?: string;
  files?: SourceFile[];
  packages?: SourcePackageSpec[];
  sourceRoots?: string[];
  artifactRoot?: string;
  packageCacheParent?: string;
  compilerVersion?: string;
  progress?: boolean;
  argv?: string[];
  sheetJSON?: string;
  sheetsJSON?: string;
  testVerbose?: boolean;
}

export type NodeEvaluationWithPackagesResult = EvaluationResult & {
  packageOutput?: string[];
};

export interface NodeLoadedSourcePackages {
  packages: Record<string, RuntimeObject>;
  packageInfos: Record<string, GoTypesPackage>;
  packageContexts: Record<string, EvaluationContext>;
  diagnostics: Diagnostic[];
  output: string[];
}

export interface NodeCompileWithPackagesResult extends CompileResult {
  packageOutput?: string[];
}

export type NodeSpreadsheetFixtureWithPackagesResult = SpreadsheetFixtureRunResult & {
  packageOutput?: string[];
  packageDiagnostics?: Diagnostic[];
};

export interface PackageArtifactCacheRequest {
  action: "path" | "list" | "clear";
  packageCacheParent?: string;
  artifactRoot?: string;
  yes?: boolean;
  confirmClear?: boolean;
}

export interface PackageArtifactCacheEntry {
  path: string;
  importPath: string;
  size: number;
  packageName?: string;
  cacheKey?: string;
  sourceHash?: string;
  layoutVersion?: string;
  compilerVersion?: string;
  backend?: string;
  hostSpecVersion?: string;
  capabilityPolicy?: string;
  dependencies?: string[];
  dependencyCacheKeys?: string[];
  exports?: BuildExport[];
}

export interface PackageArtifactCacheResult {
  ok: boolean;
  diagnostics: string[];
  action: PackageArtifactCacheRequest["action"];
  root: string;
  entries?: PackageArtifactCacheEntry[];
  cleared?: string[];
}

export function defaultPackageCacheParent(): string {
  return join(homedir(), "go", "pkg");
}

export function createNodeArtifactStore(): BuildArtifactStore {
  return {
    read(artifactPath: string): string | undefined {
      try {
        return readFileSync(artifactPath, "utf8");
      } catch (error) {
        if (isNodeErrorCode(error, "ENOENT")) return undefined;
        throw error;
      }
    },
    writeAtomic(artifactPath: string, source: string): void {
      const dir = dirname(artifactPath);
      mkdirSync(dir, { recursive: true });
      const tmp = join(dir, `.${basename(artifactPath)}.${process.pid}.tmp`);
      writeFileSync(tmp, source, "utf8");
      renameSync(tmp, artifactPath);
    }
  };
}

export function createNodeSourcePackageProvider(sourceRoots: string[] = []): BuildSourcePackageProvider | undefined {
  const roots = nodeSourceRoots(sourceRoots);
  if (roots.length === 0) return undefined;
  return {
    load(importPath: string): SourceFile[] | undefined {
      const parts = String(importPath || "").split("/").filter(Boolean);
      if (parts.length === 0 || parts.some((part) => part === "." || part === ".." || part.includes(sep))) {
        throw new Error(`invalid import path: ${importPath}`);
      }
      for (const root of roots) {
        const dir = nodeSourcePackageDir(root, importPath, parts);
        if (!dir) continue;
        let stat;
        try {
          stat = statSync(dir);
        } catch (error) {
          if (isNodeErrorCode(error, "ENOENT")) continue;
          throw error;
        }
        if (!stat.isDirectory()) throw new Error(`${dir} is not a directory`);
        const names = readdirSync(dir)
          .filter((name) => !name.startsWith(".") && !name.startsWith("_") && name.endsWith(".go") && !name.endsWith("_test.go"))
          .sort();
        const loaded = names.flatMap((name) => {
          const filename = join(dir, name);
          const source = readFileSync(filename, "utf8");
          if (!goSourceFileMatchesBuildContext(name, source, root.goos, root.goarch, root.tags)) return [];
          return {
            filename,
            source
          };
        });
        if (loaded.length === 0) throw new Error(`${dir} contains no Go source files matching GOOS=${root.goos} GOARCH=${root.goarch}`);
        return loaded;
      }
      return undefined;
    },
    isStandardLibraryPackage(importPath: string): boolean {
      const parts = String(importPath || "").split("/").filter(Boolean);
      if (parts.length === 0 || parts.some((part) => part === "." || part === ".." || part.includes(sep))) return false;
      return roots.some((root) => {
        if (!root.standardLibrary) return false;
        const dir = nodeSourcePackageDir(root, importPath, parts);
        if (!dir) return false;
        try {
          return statSync(dir).isDirectory();
        } catch {
          return false;
        }
      });
    }
  };
}

interface NodeSourceRoot {
  path: string;
  standardLibrary: boolean;
  goos: string;
  goarch: string;
  tags: Set<string>;
  modulePath?: string;
}

function nodeSourceRoots(sourceRoots: string[]): NodeSourceRoot[] {
  const roots: NodeSourceRoot[] = [];
  const seen = new Map<string, NodeSourceRoot>();
  const add = (root: string, standardLibrary: boolean): void => {
    const text = String(root || "").trim();
    if (text === "") return;
    const path = resolve(expandHome(text));
    if (!existsSync(path)) return;
    const existing = seen.get(path);
    if (existing) {
      if (standardLibrary && !existing.standardLibrary) {
        const native = nativeStandardLibraryBuildContext();
        existing.standardLibrary = true;
        existing.goos = native.goos;
        existing.goarch = native.goarch;
        existing.tags = buildTagSetForContext(existing.goos, existing.goarch, []);
      }
      return;
    }
    const native = standardLibrary ? nativeStandardLibraryBuildContext() : undefined;
    const goos = native?.goos ?? GOJR_GOOS;
    const goarch = native?.goarch ?? GOJR_GOARCH;
    const entry = {
      path,
      standardLibrary,
      goos,
      goarch,
      tags: buildTagSetForContext(goos, goarch, []),
      ...(standardLibrary ? {} : moduleSourceRoot(path))
    };
    seen.set(path, entry);
    roots.push(entry);
  };

  for (const root of sourceRoots) add(root, false);
  for (const root of candidateGOROOTSourceRoots()) add(root, true);
  for (const root of candidateGOPATHSourceRoots()) add(root, false);
  return roots;
}

function nativeStandardLibraryBuildContext(): { goos: string; goarch: string } {
  return {
    goos: nativeGOOS(),
    goarch: nativeGOARCH()
  };
}

function nativeGOOS(): string {
  switch (nodePlatform()) {
    case "win32": return "windows";
    case "sunos": return "solaris";
    default: return nodePlatform();
  }
}

function nativeGOARCH(): string {
  switch (nodeArch()) {
    case "x64": return "amd64";
    case "ia32": return "386";
    case "mipsel": return "mipsle";
    default: return nodeArch();
  }
}

function nodeSourcePackageDir(root: NodeSourceRoot, importPath: string, parts: string[]): string | undefined {
  if (root.modulePath) {
    if (importPath !== root.modulePath && !importPath.startsWith(`${root.modulePath}/`)) return undefined;
    const suffix = importPath === root.modulePath ? "" : importPath.slice(root.modulePath.length + 1);
    const suffixParts = suffix === "" ? [] : suffix.split("/").filter(Boolean);
    const dir = resolve(root.path, ...suffixParts);
    return isWithinRoot(root.path, dir) ? dir : undefined;
  }
  const dir = resolve(root.path, ...parts);
  return isWithinRoot(root.path, dir) ? dir : undefined;
}

function isWithinRoot(root: string, dir: string): boolean {
  const rel = relative(root, dir);
  return rel === "" || (!rel.startsWith("..") && !isAbsolute(rel));
}

function moduleSourceRoot(root: string): { modulePath: string } | {} {
  try {
    const modulePath = parseGoModModule(readFileSync(join(root, "go.mod"), "utf8"));
    return modulePath ? { modulePath } : {};
  } catch {
    return {};
  }
}

function parseGoModModule(source: string): string | undefined {
  for (const line of source.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (trimmed === "" || trimmed.startsWith("//")) continue;
    const match = /^module\s+(\S+)\s*$/.exec(trimmed);
    if (match?.[1]) return match[1];
    return undefined;
  }
  return undefined;
}

function candidateGOROOTSourceRoots(): string[] {
  const roots: string[] = [];
  const addGOROOT = (root: string | undefined): void => {
    const text = String(root || "").trim();
    if (text === "") return;
    roots.push(join(expandHome(text), "src"));
  };

  addGOROOT(process.env.GOROOT);
  roots.push(
    "/usr/local/go1.27rc1/src",
    "/usr/local/go1.26.4/src",
    "/usr/local/go/src",
    "/opt/homebrew/opt/go/libexec/src"
  );
  return roots;
}

function candidateGOPATHSourceRoots(): string[] {
  const gopath = String(process.env.GOPATH || "").trim();
  const roots = gopath === "" ? [join(homedir(), "go")] : gopath.split(delimiter);
  return roots
    .map((root) => String(root || "").trim())
    .filter((root) => root !== "")
    .map((root) => join(expandHome(root), "src"));
}

export function buildPackagesOnNode(request: NodeBuildPackageRequest): BuildPackageReport {
  return buildPackages(normalizeNodeBuildRequest(request), createNodeArtifactStore());
}

export function inspectPackageJavaScriptOnNode(request: NodeBuildPackageRequest): InspectPackageJavaScriptReport {
  return inspectPackageJavaScript(normalizeNodeBuildRequest(request));
}

export async function evaluateSourceWithPackagesOnNode(request: NodeSourcePackageRequest): Promise<NodeEvaluationWithPackagesResult> {
  const options = evaluationOptionsFromNodeRequest(request);
  const rootFiles = rootSourceFilesFromRequest(request);
  const loaded = await loadSourcePackagesForRootFilesOnNode(rootFiles, request.packages ?? [], options, request.sourceRoots ?? []);
  if (hasErrorDiagnostics(loaded.diagnostics)) {
    return {
      diagnostics: loaded.diagnostics,
      output: [],
      packageOutput: loaded.output
    };
  }
  const result = await evaluateSource(request.source ?? "", {
    ...options,
    packages: loaded.packages,
    packageInfos: loaded.packageInfos,
    packageContexts: loaded.packageContexts
  });
  return {
    ...result,
    packageOutput: loaded.output
  };
}

export async function evaluateSourceFilesWithPackagesOnNode(request: NodeSourcePackageRequest): Promise<NodeEvaluationWithPackagesResult> {
  const options = evaluationOptionsFromNodeRequest(request);
  const rootFiles = rootSourceFilesFromRequest(request);
  const loaded = await loadSourcePackagesForRootFilesOnNode(rootFiles, request.packages ?? [], options, request.sourceRoots ?? []);
  if (hasErrorDiagnostics(loaded.diagnostics)) {
    return {
      diagnostics: loaded.diagnostics,
      output: [],
      packageOutput: loaded.output
    };
  }
  const result = await evaluateSourceFiles(request.files ?? [], {
    ...options,
    packages: loaded.packages,
    packageInfos: loaded.packageInfos,
    packageContexts: loaded.packageContexts
  });
  return {
    ...result,
    packageOutput: loaded.output
  };
}

export async function testSourceFilesWithPackagesOnNode(request: NodeSourcePackageRequest): Promise<NodeEvaluationWithPackagesResult> {
  const options = evaluationOptionsFromNodeRequest(request);
  const rootFiles = rootSourceFilesFromRequest(request);
  const packageSources = packageSourceMapFromSpecs(request.packages ?? []);
  if (request.artifactRoot || request.packageCacheParent) {
    const buildRequest: NodeBuildPackageRequest = {
      files: rootFiles,
      sourceRoots: request.sourceRoots ?? []
    };
    if (request.importPath) buildRequest.importPath = `${request.importPath}.test`;
    if (packageSources) buildRequest.packageSources = packageSources;
    if (request.artifactRoot) buildRequest.artifactRoot = request.artifactRoot;
    if (request.packageCacheParent) buildRequest.packageCacheParent = request.packageCacheParent;
    if (request.compilerVersion) buildRequest.compilerVersion = request.compilerVersion;
    if (request.progress) buildRequest.progress = request.progress;
    const build = buildPackagesOnNode(buildRequest);
    if (!build.ok || hasErrorDiagnostics(build.diagnostics)) {
      return {
        diagnostics: build.diagnostics,
        output: []
      };
    }
  }
  const loaded = await loadSourcePackagesForRootFilesOnNode(rootFiles, request.packages ?? [], options, request.sourceRoots ?? []);
  if (hasErrorDiagnostics(loaded.diagnostics)) {
    return {
      diagnostics: loaded.diagnostics,
      output: [],
      packageOutput: loaded.output
    };
  }
  const result = await testSourceFiles(request.files ?? [], {
    ...options,
    ...(request.importPath ? { importPath: request.importPath } : {}),
    ...(request.packageName ? { packageName: request.packageName } : {}),
    packages: loaded.packages,
    packageInfos: loaded.packageInfos,
    packageContexts: loaded.packageContexts,
    ...(request.progress ? {
      onProgress(event) {
        const line = formatEvaluationProgressEvent(event);
        if (line) process.stderr.write(`${line}\n`);
      }
    } : {})
  });
  return {
    ...result,
    packageOutput: loaded.output
  };
}

export async function runMainSourceFilesWithPackagesOnNode(request: NodeSourcePackageRequest): Promise<NodeEvaluationWithPackagesResult> {
  const options = evaluationOptionsFromNodeRequest(request);
  const rootFiles = rootSourceFilesFromRequest(request);
  const packageSources = packageSourceMapFromSpecs(request.packages ?? []);
  const buildRequest: NodeBuildPackageRequest = {
    files: rootFiles,
    sourceRoots: request.sourceRoots ?? []
  };
  if (request.importPath) buildRequest.importPath = request.importPath;
  if (packageSources) buildRequest.packageSources = packageSources;
  if (request.artifactRoot) buildRequest.artifactRoot = request.artifactRoot;
  if (request.packageCacheParent) buildRequest.packageCacheParent = request.packageCacheParent;
  if (request.compilerVersion) buildRequest.compilerVersion = request.compilerVersion;
  if (request.progress) buildRequest.progress = request.progress;
  const build = buildPackagesOnNode(buildRequest);
  if (!build.ok || hasErrorDiagnostics(build.diagnostics)) {
    return {
      diagnostics: build.diagnostics,
      output: []
    };
  }
  const provider = createNodeSourcePackageProvider(request.sourceRoots ?? []);
  const result = await runMainSourcePackageFiles(rootFiles, {
    ...options,
    ...(request.importPath ? { importPath: request.importPath } : {}),
    ...(request.packageName ? { packageName: request.packageName } : {}),
    sourcePackages: request.packages ?? [],
    ...(provider ? { sourcePackageProvider: provider } : {}),
    ...(request.progress ? {
      onProgress(event) {
        const line = formatEvaluationProgressEvent(event);
        if (line) process.stderr.write(`${line}\n`);
      }
    } : {})
  });
  return result;
}

function packageSourceMapFromSpecs(specs: SourcePackageSpec[]): Record<string, SourceFile[]> | undefined {
  if (specs.length === 0) return undefined;
  const sources: Record<string, SourceFile[]> = {};
  for (const spec of specs) {
    if (!spec?.importPath) continue;
    sources[spec.importPath] = spec.files ?? [];
  }
  return Object.keys(sources).length === 0 ? undefined : sources;
}

export function compileSourceFilesWithPackagesOnNode(request: NodeSourcePackageRequest): NodeCompileWithPackagesResult {
  const options = evaluationOptionsFromNodeRequest(request);
  const rootFiles = rootSourceFilesFromRequest(request);
  const loaded = compileSourcePackagesForRootFilesOnNode(rootFiles, request.packages ?? [], options, request.sourceRoots ?? []);
  if (hasErrorDiagnostics(loaded.diagnostics)) {
    const checked = compileSourceFiles([], options).checked;
    return {
      diagnostics: loaded.diagnostics,
      checked,
      packageOutput: []
    };
  }
  const result = compileSourceFiles(request.files ?? [], {
    ...options,
    packageInfos: loaded.packageInfos
  });
  return {
    ...result,
    packageOutput: []
  };
}

export async function runSpreadsheetFixtureWithPackagesOnNode(
  request: NodeSourcePackageRequest & { fixtureJSON?: string }
): Promise<NodeSpreadsheetFixtureWithPackagesResult> {
  const fixture = parseSpreadsheetFixtureJson(request.fixtureJSON ?? "{}");
  const rootFiles = collectSpreadsheetFixtureFormulaSourceFiles(fixture);
  const loaded = await loadSourcePackagesForRootFilesOnNode(rootFiles, request.packages ?? [], {}, request.sourceRoots ?? []);
  if (hasErrorDiagnostics(loaded.diagnostics)) {
    return {
      ok: false,
      unstable: false,
      diagnostics: [],
      evaluated: [],
      sheets: {},
      observedDeps: {},
      packageOutput: loaded.output,
      packageDiagnostics: loaded.diagnostics
    };
  }
  const result = await runSpreadsheetFixture(fixture, {
    packages: loaded.packages,
    packageInfos: loaded.packageInfos,
    packageContexts: loaded.packageContexts
  });
  return {
    ...result,
    packageOutput: loaded.output
  };
}

export async function loadSourcePackagesForRootFilesOnNode(
  rootFiles: SourceFile[],
  specs: SourcePackageSpec[] = [],
  baseOptions: EvaluationOptions = {},
  sourceRoots: string[] = []
): Promise<NodeLoadedSourcePackages> {
  const diagnostics: Diagnostic[] = [];
  const explicit = new Map<string, SourcePackageSpec>();
  const initialSpecs: SourcePackageSpec[] = [];
  for (const spec of specs) {
    if (!spec?.importPath) continue;
    explicit.set(spec.importPath, spec);
    initialSpecs.push(spec);
  }

  const provider = createNodeSourcePackageProvider(sourceRoots);
  const rootImports = collectSourceImportPaths(rootFiles);
  diagnostics.push(...rootImports.diagnostics);
  if (!hasErrorDiagnostics(diagnostics)) {
    for (const importPath of rootImports.imports) {
      if (explicit.has(importPath)) continue;
      if (isAmbientSourceImport(importPath)) continue;
      const files = loadStubOrSourcePackageFromProvider(provider, importPath, rootFiles[0]?.filename ?? REPL_FILENAME, diagnostics);
      if (!files) continue;
      const spec = { importPath, files };
      explicit.set(importPath, spec);
      initialSpecs.push(spec);
    }
  }

  if (hasErrorDiagnostics(diagnostics)) {
    return { packages: {}, packageInfos: {}, packageContexts: {}, diagnostics, output: [] };
  }

  const graphOptions: EvaluationOptions & { sourcePackageProvider?: BuildSourcePackageProvider } = { ...baseOptions };
  if (provider) graphOptions.sourcePackageProvider = provider;
  const result = await evaluateSourcePackageGraph(initialSpecs, graphOptions);
  return {
    packages: result.packages,
    packageInfos: result.packageInfos,
    packageContexts: result.packageContexts,
    diagnostics: [...diagnostics, ...result.diagnostics],
    output: result.output
  };
}

export function packageCacheOnNode(request: PackageArtifactCacheRequest): PackageArtifactCacheResult {
  try {
    const root = resolveNodeArtifactRoot(request);
    switch (request.action) {
      case "path":
        return { ok: true, diagnostics: [], action: request.action, root };
      case "list":
        return { ok: true, diagnostics: [], action: request.action, root, entries: listPackageArtifactCache(root) };
      case "clear": {
        if (!(request.yes ?? request.confirmClear ?? false)) {
          return { ok: false, diagnostics: ["gojr cache clear requires --yes"], action: request.action, root };
        }
        const cleared = clearPackageArtifactCache(root);
        return { ok: true, diagnostics: [], action: request.action, root, cleared };
      }
      default:
        return { ok: false, diagnostics: [`unknown cache command ${JSON.stringify((request as { action?: unknown }).action)}`], action: "path", root };
    }
  } catch (error) {
    return {
      ok: false,
      diagnostics: [error instanceof Error ? error.message : String(error)],
      action: request.action,
      root: ""
    };
  }
}

export function listPackageArtifactCache(root: string): PackageArtifactCacheEntry[] {
  if (!existsSync(root)) return [];
  const rootStat = statSync(root);
  if (!rootStat.isDirectory()) throw new Error(`${root} is not a directory`);
  const entries: PackageArtifactCacheEntry[] = [];
  walkFiles(root, (artifactPath) => {
    if (!artifactPath.endsWith(".a")) return;
    const stat = statSync(artifactPath);
    const entry: PackageArtifactCacheEntry = {
      path: artifactPath,
      importPath: cacheImportPath(root, artifactPath),
      size: stat.size
    };
    const archive = parseGoJuniorPackageArchive(readFileSync(artifactPath, "utf8"));
    if (archive) {
      entry.importPath = archive.pkgdef.importPath || entry.importPath;
      entry.packageName = archive.pkgdef.packageName;
      entry.cacheKey = archive.pkgdef.cacheKey;
      entry.sourceHash = archive.pkgdef.sourceHash;
      entry.layoutVersion = archive.pkgdef.layoutVersion;
      entry.compilerVersion = archive.pkgdef.compilerVersion;
      entry.backend = archive.pkgdef.backend;
      entry.hostSpecVersion = archive.pkgdef.hostSpecVersion;
      entry.capabilityPolicy = archive.pkgdef.capabilityPolicy;
      entry.dependencies = archive.pkgdef.dependencies;
      entry.dependencyCacheKeys = archive.pkgdef.dependencyCacheKeys;
      entry.exports = archive.pkgdef.exports;
    }
    entries.push(entry);
  });
  return entries.sort((left, right) => left.path.localeCompare(right.path));
}

export function clearPackageArtifactCache(root: string): string[] {
  const entries = listPackageArtifactCache(root);
  if (entries.length === 0) return [];
  rmSync(root, { recursive: true, force: true });
  return entries.map((entry) => entry.path);
}

function compileSourcePackagesForRootFilesOnNode(
  rootFiles: SourceFile[],
  specs: SourcePackageSpec[] = [],
  baseOptions: EvaluationOptions = {},
  sourceRoots: string[] = []
): { packageInfos: Record<string, GoTypesPackage>; diagnostics: Diagnostic[] } {
  const diagnostics: Diagnostic[] = [];
  const packageInfos: Record<string, GoTypesPackage> = {};
  const explicit = new Map<string, SourcePackageSpec>();
  for (const spec of specs) {
    if (spec?.importPath) explicit.set(spec.importPath, spec);
  }

  const provider = createNodeSourcePackageProvider(sourceRoots);
  const loading: string[] = [];
  const loaded = new Set<string>();

  const loadImport = (importPath: string, requestedFrom: string): void => {
    if (loaded.has(importPath) || hasErrorDiagnostics(diagnostics)) return;
    if (loading.includes(importPath)) {
      diagnostics.push(nodeDiagnostic(requestedFrom, "GOJR_BUILD001", `package import cycle detected: ${[...loading, importPath].join(" -> ")}`));
      return;
    }

    let spec = explicit.get(importPath);
    if (!spec) {
      if (isAmbientSourceImport(importPath)) {
        loaded.add(importPath);
        return;
      }
      const files = loadStubOrSourcePackageFromProvider(provider, importPath, requestedFrom, diagnostics);
      if (files) {
        spec = { importPath, files };
        explicit.set(importPath, spec);
      }
    }
    if (!spec) return;

    loading.push(importPath);
    const imports = collectSourceImportPaths(spec.files);
    diagnostics.push(...imports.diagnostics);
    if (!hasErrorDiagnostics(diagnostics)) {
      for (const dependencyPath of imports.imports) {
        loadImport(dependencyPath, sourcePackageFilename(spec));
      }
    }
    if (!hasErrorDiagnostics(diagnostics)) {
      const result = compilePackageSourceFiles(spec.files, {
        ...baseOptions,
        packageInfos,
        importPath: spec.importPath,
        ...(spec.packageName ? { packageName: spec.packageName } : {})
      });
      diagnostics.push(...result.diagnostics);
      if (!hasErrorDiagnostics(diagnostics)) {
        packageInfos[spec.importPath] = result.packageInfo;
        loaded.add(importPath);
      }
    }
    loading.pop();
  };

  for (const spec of specs) {
    if (spec?.importPath) loadImport(spec.importPath, sourcePackageFilename(spec));
  }
  const rootImports = collectSourceImportPaths(rootFiles);
  diagnostics.push(...rootImports.diagnostics);
  if (!hasErrorDiagnostics(diagnostics)) {
    for (const importPath of rootImports.imports) {
      loadImport(importPath, rootFiles[0]?.filename ?? REPL_FILENAME);
    }
  }
  return { packageInfos, diagnostics };
}

function isAmbientSourceImport(importPath: string): boolean {
  return isHostResolvedSourceImport(importPath);
}

function rootSourceFilesFromRequest(request: NodeSourcePackageRequest): SourceFile[] {
  if (request.files && request.files.length > 0) return request.files;
  return [{
    filename: REPL_FILENAME,
    source: request.source ?? ""
  }];
}

function evaluationOptionsFromNodeRequest(request: NodeSourcePackageRequest): EvaluationOptions {
  const options: EvaluationOptions = {};
  if (request.sheetJSON) options.sheet = parseSheetJson(request.sheetJSON);
  if (request.sheetsJSON) options.sheets = parseSheetsJson(request.sheetsJSON);
  if (request.argv) options.argv = request.argv;
  if (typeof request.testVerbose === "boolean") options.testVerbose = request.testVerbose;
  return options;
}

function loadSourcePackageFromProvider(
  provider: BuildSourcePackageProvider | undefined,
  importPath: string,
  requestedFrom: string,
  diagnostics: Diagnostic[]
): SourceFile[] | undefined {
  if (!provider) return undefined;
  try {
    return provider.load(importPath);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    diagnostics.push(nodeDiagnostic(requestedFrom, "GOJR_BUILD001", `could not load package ${importPath}: ${message}`));
    return undefined;
  }
}

function loadStubOrSourcePackageFromProvider(
  provider: BuildSourcePackageProvider | undefined,
  importPath: string,
  requestedFrom: string,
  diagnostics: Diagnostic[]
): SourceFile[] | undefined {
  return stubSourcePackageFiles(importPath) ?? loadSourcePackageFromProvider(provider, importPath, requestedFrom, diagnostics);
}

function sourcePackageFilename(spec: SourcePackageSpec): string {
  return spec.files[0]?.filename ?? REPL_FILENAME;
}

function nodeDiagnostic(filename: string, code: string, message: string): Diagnostic {
  return {
    filename,
    code,
    severity: "error",
    message
  };
}

function normalizeNodeBuildRequest(request: NodeBuildPackageRequest): BuildPackageRequest {
  const normalized: NodeBuildPackageRequest = { ...request };
  if (!normalized.artifactRoot && !normalized.packageCacheParent) {
    normalized.packageCacheParent = defaultPackageCacheParent();
  }
  if (normalized.artifactRoot) normalized.artifactRoot = expandHome(normalized.artifactRoot);
  if (normalized.packageCacheParent) normalized.packageCacheParent = expandHome(normalized.packageCacheParent);
  const provider = createNodeSourcePackageProvider(normalized.sourceRoots ?? []);
  if (provider) normalized.sourcePackageProvider = provider;
  if (normalized.progress) {
    normalized.onProgress = (event) => {
      const line = formatBuildProgressEvent(event);
      if (line) process.stderr.write(`${line}\n`);
    };
  }
  return normalized;
}

function resolveNodeArtifactRoot(request: Pick<PackageArtifactCacheRequest, "artifactRoot" | "packageCacheParent">): string {
  if (request.artifactRoot && request.packageCacheParent) {
    throw new Error("-pkgdir and -artifact-root are mutually exclusive");
  }
  if (request.artifactRoot) {
    return resolveArtifactRoot({ artifactRoot: expandHome(request.artifactRoot) });
  }
  return resolveArtifactRoot({
    packageCacheParent: expandHome(request.packageCacheParent || defaultPackageCacheParent())
  });
}

function walkFiles(root: string, visit: (path: string) => void): void {
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const path = join(root, entry.name);
    if (entry.isDirectory()) {
      walkFiles(path, visit);
    } else if (entry.isFile()) {
      visit(path);
    }
  }
}

function cacheImportPath(root: string, artifactPath: string): string {
  const rel = relative(root, artifactPath);
  const slash = rel.split(sep).join("/");
  return slash.endsWith(".a") ? slash.slice(0, -2) : slash;
}

function expandHome(path: string): string {
  if (path === "~") return homedir();
  if (path.startsWith("~/")) return join(homedir(), path.slice(2));
  return path;
}

function isNodeErrorCode(error: unknown, code: string): boolean {
  return typeof error === "object" && error !== null && (error as { code?: unknown }).code === code;
}
