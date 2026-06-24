import { existsSync, mkdirSync, readFileSync, readdirSync, renameSync, rmSync, statSync, writeFileSync } from "node:fs";
import { basename, dirname, isAbsolute, join, relative, resolve, sep } from "node:path";
import { homedir } from "node:os";
import {
  buildPackages,
  collectSourceImportPaths,
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
import { parseSheetJson, parseSheetsJson } from "./jsonInput.js";
import {
  evaluateSource,
  evaluateSourceFiles,
  evaluateSourcePackageGraph,
  testSourceFiles,
  type EvaluationOptions,
  type EvaluationResult,
  type RuntimeObject,
  type SourcePackageSpec
} from "./runtime.js";
import type { Package as GoTypesPackage } from "./go/types/index.js";

export interface NodeBuildPackageRequest extends BuildPackageRequest {
  sourceRoots?: string[];
}

export interface NodeSourcePackageRequest {
  source?: string;
  files?: SourceFile[];
  packages?: SourcePackageSpec[];
  sourceRoots?: string[];
  sheetJSON?: string;
  sheetsJSON?: string;
}

export type NodeEvaluationWithPackagesResult = EvaluationResult & {
  packageOutput?: string[];
};

export interface NodeLoadedSourcePackages {
  packages: Record<string, RuntimeObject>;
  packageInfos: Record<string, GoTypesPackage>;
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
  const roots = [...new Set(sourceRoots
    .map((root) => String(root || "").trim())
    .filter((root) => root !== "")
    .map((root) => resolve(expandHome(root))))];
  if (roots.length === 0) return undefined;
  return {
    load(importPath: string): SourceFile[] | undefined {
      const parts = String(importPath || "").split("/").filter(Boolean);
      if (parts.length === 0 || parts.some((part) => part === "." || part === ".." || part.includes(sep))) {
        throw new Error(`invalid import path: ${importPath}`);
      }
      for (const root of roots) {
        const dir = resolve(root, ...parts);
        const rel = relative(root, dir);
        if (rel === "" || rel.startsWith("..") || isAbsolute(rel)) continue;
        let stat;
        try {
          stat = statSync(dir);
        } catch (error) {
          if (isNodeErrorCode(error, "ENOENT")) continue;
          throw error;
        }
        if (!stat.isDirectory()) throw new Error(`${dir} is not a directory`);
        const names = readdirSync(dir)
          .filter((name) => !name.startsWith(".") && name.endsWith(".go") && !name.endsWith("_test.go"))
          .sort();
        if (names.length === 0) throw new Error(`${dir} contains no non-test .go files`);
        return names.map((name) => {
          const filename = join(dir, name);
          return {
            filename,
            source: readFileSync(filename, "utf8")
          };
        });
      }
      return undefined;
    }
  };
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
    packageInfos: loaded.packageInfos
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
    packageInfos: loaded.packageInfos
  });
  return {
    ...result,
    packageOutput: loaded.output
  };
}

export async function testSourceFilesWithPackagesOnNode(request: NodeSourcePackageRequest): Promise<NodeEvaluationWithPackagesResult> {
  const rootFiles = rootSourceFilesFromRequest(request);
  const loaded = await loadSourcePackagesForRootFilesOnNode(rootFiles, request.packages ?? [], {}, request.sourceRoots ?? []);
  if (hasErrorDiagnostics(loaded.diagnostics)) {
    return {
      diagnostics: loaded.diagnostics,
      output: [],
      packageOutput: loaded.output
    };
  }
  const result = await testSourceFiles(request.files ?? [], {
    packages: loaded.packages,
    packageInfos: loaded.packageInfos
  });
  return {
    ...result,
    packageOutput: loaded.output
  };
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
    packageInfos: loaded.packageInfos
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
      const files = loadSourcePackageFromProvider(provider, importPath, rootFiles[0]?.filename ?? REPL_FILENAME, diagnostics);
      if (!files) continue;
      const spec = { importPath, files };
      explicit.set(importPath, spec);
      initialSpecs.push(spec);
    }
  }

  if (hasErrorDiagnostics(diagnostics)) {
    return { packages: {}, packageInfos: {}, diagnostics, output: [] };
  }

  const graphOptions: EvaluationOptions & { sourcePackageProvider?: BuildSourcePackageProvider } = { ...baseOptions };
  if (provider) graphOptions.sourcePackageProvider = provider;
  const result = await evaluateSourcePackageGraph(initialSpecs, graphOptions);
  return {
    packages: result.packages,
    packageInfos: result.packageInfos,
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
      const files = loadSourcePackageFromProvider(provider, importPath, requestedFrom, diagnostics);
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
