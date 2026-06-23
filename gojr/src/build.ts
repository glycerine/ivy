import { Diagnostic, REPL_FILENAME, SourceFile, SourceSpan } from "./diagnostics.js";
import { File, ImportSpec } from "./front/ast.js";
import { checkFrontFiles } from "./front/checker.js";
import { parseFrontSourceFiles } from "./front/parser.js";
import { newUniverse, ObjectKind, PackageInfo, TypeObject, Universe } from "./front/types.js";

export interface BuildPackageRequest {
  importPath?: string;
  files: SourceFile[];
  packageSources?: Record<string, SourceFile[]>;
  sourcePackageProvider?: BuildSourcePackageProvider;
  artifactRoot?: string;
  packageCacheParent?: string;
  compilerVersion?: string;
  backend?: string;
  hostSpecVersion?: string;
  capabilityPolicy?: string;
}

export interface BuildSourcePackageProvider {
  load(importPath: string): SourceFile[] | undefined;
}

export interface BuildArtifactStore {
  read?(path: string): string | undefined;
  writeAtomic(path: string, source: string): void;
}

export interface BuildExport {
  name: string;
  kind: "const" | "func" | "type" | "var";
  typeText: string;
  underlyingTypeText?: string;
}

export interface BuildArtifactReport {
  importPath: string;
  packageName: string;
  artifactPath: string;
  action: "built" | "skipped";
  sourceHash: string;
  cacheKey: string;
  dependencies: string[];
  dependencyCacheKeys: string[];
  exports: BuildExport[];
}

export interface BuildPackageReport {
  ok: boolean;
  diagnostics: Diagnostic[];
  artifacts: BuildArtifactReport[];
  built: string[];
  skipped: string[];
}

export interface InspectPackageJavaScriptReport extends BuildPackageReport {
  source: string;
}

export interface SourceImportPathsResult {
  diagnostics: Diagnostic[];
  imports: string[];
}

const ARTIFACT_LAYOUT_VERSION = "gojr-js-v2";
const DEFAULT_COMPILER_VERSION = "gojr-dev";
const DEFAULT_BACKEND = "js-source-envelope";
const DEFAULT_HOST_SPEC_VERSION = "host-v0";
const DEFAULT_CAPABILITY_POLICY = "default";

export function buildPackage(request: BuildPackageRequest, store?: BuildArtifactStore): BuildPackageReport {
  return buildPackages(request, store);
}

export function buildPackages(request: BuildPackageRequest, store?: BuildArtifactStore): BuildPackageReport {
  return new PackageGraphBuilder(request, store).build();
}

export function inspectPackageJavaScript(request: BuildPackageRequest): InspectPackageJavaScriptReport {
  let source = "";
  const report = buildPackages(request, {
    read() {
      return undefined;
    },
    writeAtomic(_path, nextSource) {
      source = nextSource;
    }
  });
  return {
    ...report,
    source
  };
}

export function collectSourceImportPaths(files: SourceFile[]): SourceImportPathsResult {
  const parsed = parseFrontSourceFiles(files.map(ensureSourceFile));
  return {
    diagnostics: parsed.diagnostics,
    imports: uniqueSorted(parsed.files.flatMap((file) => file.imports.map(importPathFromSpec)))
  };
}

interface PackageBuildNode {
  importPath: string;
  packageName: string;
  files: SourceFile[];
  sourceHash: string;
  cacheKey: string;
  dependencies: string[];
  dependencyCacheKeys: string[];
  exports: BuildExport[];
  artifactPath: string;
  artifactSource: string;
}

class PackageGraphBuilder {
  private readonly universe: Universe = newUniverse();
  private readonly packageSources = new Map<string, SourceFile[]>();
  private readonly packageInfos = new Map<string, PackageInfo>();
  private readonly nodes = new Map<string, PackageBuildNode>();
  private readonly visiting = new Set<string>();
  private readonly failedPackageLoads = new Set<string>();
  private readonly diagnostics: Diagnostic[] = [];

  public constructor(
    private readonly request: BuildPackageRequest,
    private readonly store?: BuildArtifactStore
  ) {
    for (const [importPath, files] of Object.entries(request.packageSources ?? {})) {
      this.packageSources.set(importPath, files.map(ensureSourceFile));
    }
  }

  public build(): BuildPackageReport {
    const root = this.buildOne(this.request.importPath, this.request.files, []);
    if (this.hasErrors() || !root) return emptyBuildReport(this.diagnostics);

    const artifacts: BuildArtifactReport[] = [];
    const built: string[] = [];
    const skipped: string[] = [];
    for (const node of this.orderedNodes(root)) {
      const reportArtifact = this.reportArtifact(node, "built");
      const existing = this.store?.read?.(node.artifactPath);
      const existingArtifact = existing === undefined ? undefined : parseGeneratedArtifactSource(existing);
      if (existingArtifact?.cacheKey === node.cacheKey && existingArtifact.layoutVersion === ARTIFACT_LAYOUT_VERSION) {
        artifacts.push({ ...reportArtifact, action: "skipped" });
        skipped.push(node.artifactPath);
        continue;
      }
      this.store?.writeAtomic(node.artifactPath, node.artifactSource);
      artifacts.push(reportArtifact);
      built.push(node.artifactPath);
    }

    return {
      ok: true,
      diagnostics: this.diagnostics,
      artifacts,
      built,
      skipped
    };
  }

  private buildOne(importPathHint: string | undefined, rawFiles: SourceFile[], stack: string[]): PackageBuildNode | undefined {
    const files = rawFiles.map(ensureSourceFile);
    const parsed = parseFrontSourceFiles(files);
    this.diagnostics.push(...parsed.diagnostics);
    this.diagnostics.push(...packageDiagnostics(parsed.files, files));
    if (this.hasErrors()) return undefined;

    const packageName = parsed.files.find((file) => file.name)?.name?.name ?? "main";
    const importPath = importPathHint ?? packageName;
    const existing = this.nodes.get(importPath);
    if (existing) return existing;
    if (this.visiting.has(importPath)) {
      const cycle = [...stack, importPath].join(" -> ");
      this.diagnostics.push(buildDiagnostic(files[0]?.filename ?? REPL_FILENAME, `package import cycle detected: ${cycle}`));
      return undefined;
    }

    if (!this.packageSources.has(importPath)) {
      this.packageSources.set(importPath, files);
    }
    this.visiting.add(importPath);
    const dependencies = uniqueSorted(parsed.files.flatMap((file) => file.imports.map(importPathFromSpec)));
    const sourceDependencies: PackageBuildNode[] = [];
    for (const dependencyPath of dependencies) {
      if (isStandardBuildImport(dependencyPath) && !this.packageSources.has(dependencyPath)) continue;
      const dependencyFiles = this.sourceFilesForImport(dependencyPath, files[0]?.filename ?? REPL_FILENAME);
      if (!dependencyFiles) {
        if (!isStandardBuildImport(dependencyPath) && !this.failedPackageLoads.has(dependencyPath)) {
          this.diagnostics.push(buildDiagnostic(
            files[0]?.filename ?? REPL_FILENAME,
            `package ${dependencyPath} is not available to gojr build; provide it with packageSources or --pkg`
          ));
        }
        continue;
      }
      const dependency = this.buildOne(dependencyPath, dependencyFiles, [...stack, importPath]);
      if (dependency) sourceDependencies.push(dependency);
    }

    if (!this.hasErrors()) {
      const checked = checkFrontFiles(parsed.files, {
        universe: this.universe,
        packageName,
        packagePath: importPath,
        importer: {
          import: (path) => this.packageInfos.get(path)
        }
      });
      this.diagnostics.push(...checked.diagnostics);
      if (!this.hasErrors()) {
        const sourceHash = hashSourceFiles(files);
        const dependencyCacheKeys = sourceDependencies.map((dependency) => `${dependency.importPath}:${dependency.cacheKey}`).sort();
        const exports = uniqueExports(checked.pkg.scope.children());
        const cacheKey = stableHash([
          ARTIFACT_LAYOUT_VERSION,
          this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
          this.request.backend ?? DEFAULT_BACKEND,
          this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
          this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
          importPath,
          sourceHash,
          dependencies.join("\n"),
          dependencyCacheKeys.join("\n")
        ].join("\0"));
        const artifactPath = artifactPathForImportPath(resolveArtifactRoot(this.request), importPath);
        const artifact = {
          layoutVersion: ARTIFACT_LAYOUT_VERSION,
          compilerVersion: this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
          backend: this.request.backend ?? DEFAULT_BACKEND,
          hostSpecVersion: this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
          capabilityPolicy: this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
          importPath,
          packageName,
          sourceHash,
          cacheKey,
          dependencies,
          dependencyCacheKeys,
          exports,
          sources: files.map((file) => ({
            filename: file.filename,
            hash: stableHash(file.source)
          }))
        };
        const node: PackageBuildNode = {
          importPath,
          packageName,
          files,
          sourceHash,
          cacheKey,
          dependencies,
          dependencyCacheKeys,
          exports,
          artifactPath,
          artifactSource: generatedArtifactSource(artifact)
        };
        this.nodes.set(importPath, node);
        this.packageInfos.set(importPath, checked.pkg);
      }
    }

    this.visiting.delete(importPath);
    return this.nodes.get(importPath);
  }

  private sourceFilesForImport(importPath: string, filename: string): SourceFile[] | undefined {
    const explicit = this.packageSources.get(importPath);
    if (explicit) return explicit;
    try {
      const loaded = this.request.sourcePackageProvider?.load(importPath);
      if (!loaded) return undefined;
      const files = loaded.map(ensureSourceFile);
      this.packageSources.set(importPath, files);
      return files;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      this.failedPackageLoads.add(importPath);
      this.diagnostics.push(buildDiagnostic(filename, `could not load package ${importPath}: ${message}`));
      return undefined;
    }
  }

  private orderedNodes(root: PackageBuildNode): PackageBuildNode[] {
    const out: PackageBuildNode[] = [];
    const seen = new Set<string>();
    const visit = (node: PackageBuildNode): void => {
      if (seen.has(node.importPath)) return;
      seen.add(node.importPath);
      for (const dependencyPath of node.dependencies) {
        const dependency = this.nodes.get(dependencyPath);
        if (dependency) visit(dependency);
      }
      out.push(node);
    };
    visit(root);
    return out;
  }

  private reportArtifact(node: PackageBuildNode, action: BuildArtifactReport["action"]): BuildArtifactReport {
    return {
      importPath: node.importPath,
      packageName: node.packageName,
      artifactPath: node.artifactPath,
      action,
      sourceHash: node.sourceHash,
      cacheKey: node.cacheKey,
      dependencies: node.dependencies,
      dependencyCacheKeys: node.dependencyCacheKeys,
      exports: node.exports
    };
  }

  private hasErrors(): boolean {
    return this.diagnostics.some((diagnostic) => diagnostic.severity === "error");
  }
}

function parseGeneratedArtifactSource(source: string): { layoutVersion?: string; cacheKey?: string } | undefined {
  const match = /export const gojrPackageArtifact = ([\s\S]*);\s*$/.exec(source);
  if (!match?.[1]) return undefined;
  try {
    const value = JSON.parse(match[1]) as { layoutVersion?: unknown; cacheKey?: unknown };
    return {
      ...(typeof value.layoutVersion === "string" ? { layoutVersion: value.layoutVersion } : {}),
      ...(typeof value.cacheKey === "string" ? { cacheKey: value.cacheKey } : {})
    };
  } catch {
    return undefined;
  }
}

export function artifactPathForImportPath(artifactRoot: string, importPath: string): string {
  const parts = importPath.split("/").filter(Boolean);
  if (parts.length === 0 || parts.some((part) => part === "." || part === "..")) {
    throw new Error(`invalid import path: ${importPath}`);
  }
  return joinSlash(artifactRoot, ...parts) + ".js";
}

export function resolveArtifactRoot(request: Pick<BuildPackageRequest, "artifactRoot" | "packageCacheParent">): string {
  if (request.artifactRoot && request.artifactRoot.trim() !== "") return trimTrailingSlash(request.artifactRoot);
  if (request.packageCacheParent && request.packageCacheParent.trim() !== "") return joinSlash(request.packageCacheParent, "gojr_js");
  return "~/go/pkg/gojr_js";
}

function packageDiagnostics(files: File[], sourceFiles: SourceFile[]): Diagnostic[] {
  if (sourceFiles.length === 0) {
    return [buildDiagnostic(REPL_FILENAME, "no source files supplied")];
  }
  if (files.length === 0) return [];

  const firstName = files.find((file) => file.name)?.name;
  if (!firstName) {
    return [buildDiagnostic(files[0]?.span?.filename ?? sourceFiles[0]?.filename ?? REPL_FILENAME, "package source files must declare a package")];
  }

  const diagnostics: Diagnostic[] = [];
  for (const file of files) {
    if (!file.name) {
      diagnostics.push(buildDiagnostic(file.span?.filename ?? REPL_FILENAME, "package source files must declare a package", file.span));
      continue;
    }
    if (file.name.name !== firstName.name) {
      diagnostics.push(buildDiagnostic(
        file.name.span?.filename ?? file.span?.filename ?? REPL_FILENAME,
        `package ${file.name.name} does not match package ${firstName.name}`,
        file.name.span
      ));
    }
  }
  return diagnostics;
}

function isStandardBuildImport(path: string): boolean {
  return path === "fmt" || path === "testing";
}

function buildDiagnostic(filename: string, message: string, span?: SourceSpan): Diagnostic {
  return {
    filename,
    code: "GOJR_BUILD001",
    severity: "error",
    message,
    ...(span ? { span } : {})
  };
}

function ensureSourceFile(file: SourceFile): SourceFile {
  return {
    filename: file.filename || REPL_FILENAME,
    source: file.source ?? ""
  };
}

function emptyBuildReport(diagnostics: Diagnostic[]): BuildPackageReport {
  return {
    ok: false,
    diagnostics,
    artifacts: [],
    built: [],
    skipped: []
  };
}

function importPathFromSpec(spec: ImportSpec): string {
  return unquoteStringLiteral(spec.path.value);
}

function uniqueExports(objects: TypeObject[]): BuildExport[] {
  const exports = new Map<string, BuildExport>();
  for (const object of objects) {
    const kind = exportKindForObject(object.kind);
    if (!kind || !object.exported()) continue;
    const item: BuildExport = {
      name: object.name,
      kind,
      typeText: object.type.typeString()
    };
    if (kind === "type") {
      item.underlyingTypeText = object.type.underlying().typeString();
    }
    exports.set(`${kind}:${object.name}`, item);
  }
  return [...exports.values()].sort((left, right) => left.name.localeCompare(right.name) || left.kind.localeCompare(right.kind));
}

function exportKindForObject(kind: ObjectKind): BuildExport["kind"] | undefined {
  switch (kind) {
    case ObjectKind.Const: return "const";
    case ObjectKind.Func: return "func";
    case ObjectKind.TypeName: return "type";
    case ObjectKind.Var: return "var";
    default: return undefined;
  }
}

function generatedArtifactSource(artifact: unknown): string {
  return [
    "// Code generated by gojr build; DO NOT EDIT.",
    "// This envelope is the package artifact cache format used before full JS lowering.",
    `export const gojrPackageArtifact = ${JSON.stringify(artifact, null, 2)};`,
    ""
  ].join("\n");
}

function hashSourceFiles(files: SourceFile[]): string {
  return stableHash(files
    .map((file) => `${file.filename}\0${file.source.length}\0${file.source}`)
    .join("\0"));
}

function stableHash(text: string): string {
  let hash = 0xcbf29ce484222325n;
  const prime = 0x100000001b3n;
  const mask = 0xffffffffffffffffn;
  for (let index = 0; index < text.length; index++) {
    hash ^= BigInt(text.charCodeAt(index));
    hash = (hash * prime) & mask;
  }
  return hash.toString(16).padStart(16, "0");
}

function unquoteStringLiteral(value: string): string {
  if (value.length < 2) return value;
  const quote = value[0];
  if ((quote !== `"` && quote !== "`") || value[value.length - 1] !== quote) return value;
  if (quote === "`") return value.slice(1, -1);
  try {
    return JSON.parse(value) as string;
  } catch {
    return value.slice(1, -1);
  }
}

function uniqueSorted(values: string[]): string[] {
  return [...new Set(values)].sort();
}

function joinSlash(first: string, ...rest: string[]): string {
  return [first, ...rest]
    .filter((part) => part !== "")
    .map((part, index) => index === 0 ? trimTrailingSlash(part) : trimSlashes(part))
    .join("/");
}

function trimSlashes(text: string): string {
  return text.replace(/^\/+|\/+$/g, "");
}

function trimTrailingSlash(text: string): string {
  return text.replace(/\/+$/g, "") || "/";
}
