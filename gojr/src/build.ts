import { Diagnostic, REPL_FILENAME, SourceFile, SourceSpan } from "./diagnostics.js";
import type { FunctionDecl, ProgramAst } from "./ast.js";
import { blake3HashString } from "./blake3.js";
import { File, ImportSpec } from "./front/ast.js";
import { parseFrontSourceFiles } from "./front/parser.js";
import { frontFilesToProgramAst } from "./frontToAst.js";
import { intrinsicPackageName, isIntrinsicPackageImport } from "./intrinsicPackages.js";
import { isStubSourcePackageStandardLibrary, stubSourcePackageFiles } from "./stubPackages.js";
import { checkGoJuniorFiles, isGoJuniorSyntheticCheckName, standardTypePackage, type GoJuniorCheckResult } from "./typecheck.js";
import { emitStage1Package, GOJR_STAGE1_BACKEND } from "./emitter/package.js";
import { Codebase, type CodebaseUpdateTxn } from "./codebase.js";
import {
  Builtin as GoTypesBuiltin,
  Const as GoTypesConst,
  Func as GoTypesFunc,
  RelativeTo as GoTypesRelativeTo,
  TypeName as GoTypesTypeName,
  TypeString as GoTypesTypeString,
  Var as GoTypesVar,
  type Object as GoTypesObject,
  type Package as GoTypesPackage,
  type Type as GoTypesType
} from "./go/types/index.js";

export interface BuildPackageRequest {
  importPath?: string;
  files: SourceFile[];
  packageSources?: Record<string, SourceFile[]>;
  sourcePackageProvider?: BuildSourcePackageProvider;
  artifactRoot?: string;
  packageCacheParent?: string;
  goos?: string;
  goarch?: string;
  buildTags?: string[];
  compilerVersion?: string;
  backend?: string;
  hostSpecVersion?: string;
  capabilityPolicy?: string;
  standardLibraryPackages?: string[];
  onProgress?: BuildProgressSink;
}

export interface BuildSourcePackageProvider {
  load(importPath: string): SourceFile[] | undefined;
  isStandardLibraryPackage?(importPath: string): boolean;
}

export type BuildProgressAction = "checking" | "built" | "cached";

export interface BuildProgressEvent {
  action: BuildProgressAction;
  importPath: string;
  packageName?: string;
  artifactPath?: string;
  dependencyCount?: number;
  fileCount?: number;
  standardLibrary?: boolean;
}

export type BuildProgressSink = (event: BuildProgressEvent) => void;

export interface BuildArtifactStore {
  read?(path: string): string | undefined;
  mtimeMs?(path: string): number | undefined;
  writeAtomic(path: string, source: string): void;
}

export interface BuildExport {
  name: string;
  kind: "const" | "func" | "type" | "var";
  typeText: string;
  underlyingTypeText?: string;
}

export interface GoJuniorPackageExportIndexEntry {
  exportIndex: number;
  kind: BuildExport["kind"];
  typeText: string;
  underlyingTypeText?: string;
}

export interface GoJuniorPackageExportData {
  exportFormat: "gojr-iexport";
  exportVersion: 1;
  exportIndex: Record<string, GoJuniorPackageExportIndexEntry>;
  layoutVersion: string;
  compilerVersion: string;
  backend: string;
  hostSpecVersion: string;
  capabilityPolicy: string;
  goos: string;
  goarch: string;
  buildTags: string[];
  standardLibrary: boolean;
  importPath: string;
  packageName: string;
  sourceHash: string;
  cacheKey: string;
  dependencies: string[];
  dependencyCacheKeys: string[];
  exports: BuildExport[];
  sources: Array<{ filename: string; hash: string }>;
  runtime?: GoJuniorPackageRuntimePayload;
}

export interface GoJuniorPackageArchive {
  pkgdef: GoJuniorPackageExportData;
  javascript: string;
  wasmBase64?: string;
  runtime?: GoJuniorPackageRuntimePayload;
  members: Array<{ name: string; data: string }>;
}

export interface GoJuniorPackageRuntimePayload {
  ast: ProgramAst;
  runtimePlan: GoJuniorPackageRuntimePlan;
}

export interface GoJuniorPackageSourcePayload {
  filename: string;
  hash: string;
  source: string;
}

export interface GoJuniorPackageRuntimeVariable {
  name: string;
  typeText?: string;
}

export interface GoJuniorPackageRuntimeConstant {
  name: string;
  typeText?: string;
  value: unknown;
}

export interface GoJuniorPackageRuntimePlan {
  importPath: string;
  packageName: string;
  exportedNames: string[];
  constants: GoJuniorPackageRuntimeConstant[];
  variables: GoJuniorPackageRuntimeVariable[];
  varInitOrder: string[][];
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

export interface StandardLibrarySourceHost {
  readDir(path: string): Array<string | { name: string; isFile?: boolean }>;
  readFile(path: string): string;
}

export interface StandardLibrarySourcePackageProviderOptions {
  sourceRoot: string;
  host: StandardLibrarySourceHost;
  goos?: string;
  goarch?: string;
  buildTags?: string[];
}

export interface BuildStandardLibraryPackageRequest extends Omit<BuildPackageRequest, "files"> {
  importPath: string;
  standardLibrary: StandardLibrarySourcePackageProviderOptions;
}

const ARTIFACT_LAYOUT_VERSION = "gojr-js-v4";
export const GOJR_GOOS = "js";
export const GOJR_GOARCH = "gojr";
const GOJR_DEFAULT_BUILD_TAGS = ["codec.safe"];
const DEFAULT_COMPILER_VERSION = "gojr-dev";
const DEFAULT_BACKEND = GOJR_STAGE1_BACKEND;
const DEFAULT_HOST_SPEC_VERSION = "host-v0";
const DEFAULT_CAPABILITY_POLICY = "default";
const AR_MAGIC = "!<arch>\n";
const PKGDEF_MEMBER = "__.PKGDEF";
const JAVASCRIPT_MEMBER = "_gojr.js";
const WASM_MEMBER = "_gojr.wasm";
const GOJR_EXPORT_MAGIC = "$$gojr iexport v1\n";

export function buildPackage(request: BuildPackageRequest, store?: BuildArtifactStore): BuildPackageReport {
  return buildPackages(request, store);
}

export function buildPackages(request: BuildPackageRequest, store?: BuildArtifactStore): BuildPackageReport {
  return new PackageGraphBuilder(request, store).build();
}

export function buildStandardLibraryPackage(request: BuildStandardLibraryPackageRequest, store?: BuildArtifactStore): BuildPackageReport {
  const provider = request.sourcePackageProvider ?? createStandardLibrarySourcePackageProvider(request.standardLibrary);
  const files = isIntrinsicPackageImport(request.importPath)
    ? [intrinsicSyntheticSourceFile(request.importPath)]
    : provider.load(request.importPath);
  if (!files) {
    return emptyBuildReport([buildDiagnostic(
      request.importPath,
      `standard library package ${request.importPath} is not available under ${request.standardLibrary.sourceRoot}`
    )]);
  }
  const { standardLibrary, ...rest } = request;
  const buildTags = request.buildTags ?? standardLibrary.buildTags;
  return buildPackages({
    ...rest,
    files,
    sourcePackageProvider: provider,
    goos: request.goos ?? standardLibrary.goos ?? GOJR_GOOS,
    goarch: request.goarch ?? standardLibrary.goarch ?? GOJR_GOARCH,
    ...(buildTags ? { buildTags } : {}),
    standardLibraryPackages: uniqueSorted([...(request.standardLibraryPackages ?? []), request.importPath])
  }, store);
}

export function createStandardLibrarySourcePackageProvider(options: StandardLibrarySourcePackageProviderOptions): BuildSourcePackageProvider {
  const sourceRoot = trimTrailingSlash(options.sourceRoot);
  const goos = options.goos ?? GOJR_GOOS;
  const goarch = options.goarch ?? GOJR_GOARCH;
  const tags = buildTagSet(goos, goarch, options.buildTags);
  return {
    load(importPath: string): SourceFile[] | undefined {
      const parts = importPathParts(importPath);
      if (!parts) throw new Error(`invalid import path: ${importPath}`);
      const dir = joinSlash(sourceRoot, ...parts);
      let entries: Array<string | { name: string; isFile?: boolean }>;
      try {
        entries = options.host.readDir(dir);
      } catch {
        return undefined;
      }
      const names = entries
        .map((entry) => typeof entry === "string" ? { name: entry, isFile: true } : entry)
        .filter((entry) => entry.isFile !== false)
        .map((entry) => entry.name)
      .filter((name) => !name.startsWith(".") && !name.startsWith("_") && name.endsWith(".go") && !name.endsWith("_test.go"))
      .sort();
      const files: SourceFile[] = [];
      for (const name of names) {
        const filename = joinSlash(dir, name);
        const source = options.host.readFile(filename);
        if (goSourceFileMatchesBuildContext(name, source, goos, goarch, tags)) {
          files.push({ filename, source });
        }
      }
      if (files.length === 0) throw new Error(`${dir} contains no Go source files matching GOOS=${goos} GOARCH=${goarch}`);
      return files;
    },
    isStandardLibraryPackage(importPath: string): boolean {
      return importPathParts(importPath) !== undefined;
    }
  };
}

export function inspectPackageJavaScript(request: BuildPackageRequest): InspectPackageJavaScriptReport {
  let source = "";
  const report = buildPackages(request, {
    read() {
      return undefined;
    },
    writeAtomic(_path, nextSource) {
      source = parseGoJuniorPackageArchive(nextSource)?.javascript ?? nextSource;
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
  private readonly packageSources = new Map<string, SourceFile[]>();
  private readonly codebase = new Codebase();
  private readonly nodes = new Map<string, PackageBuildNode>();
  private readonly visiting = new Set<string>();
  private readonly failedPackageLoads = new Set<string>();
  private readonly standardLibraryPackages = new Set<string>();
  private readonly diagnostics: Diagnostic[] = [];
  private activeTxn: CodebaseUpdateTxn | undefined;

  public constructor(
    private readonly request: BuildPackageRequest,
    private readonly store?: BuildArtifactStore
  ) {
    for (const [importPath, files] of Object.entries(request.packageSources ?? {})) {
      this.packageSources.set(importPath, files.map(ensureSourceFile));
    }
    for (const importPath of request.standardLibraryPackages ?? []) {
      this.standardLibraryPackages.add(importPath);
    }
  }

  public build(): BuildPackageReport {
    const cached = this.buildFromFreshArtifacts();
    if (cached) return cached;

    const txn = this.codebase.NewUpdateTxn();
    this.activeTxn = txn;
    const root = this.buildOne(this.request.importPath, this.request.files, []);
    if (this.hasErrors() || !root) {
      if (!txn.IsClosed()) txn.Rollback();
      this.activeTxn = undefined;
      return emptyBuildReport(this.diagnostics);
    }
    txn.Commit();
    this.activeTxn = undefined;

    const artifacts: BuildArtifactReport[] = [];
    const built: string[] = [];
    const skipped: string[] = [];
    for (const node of this.orderedNodes(root)) {
      const reportArtifact = this.reportArtifact(node, "built");
      const existing = this.store?.read?.(node.artifactPath);
      const existingArtifact = existing === undefined ? undefined : parseGeneratedArtifactSource(existing);
      if (
        existingArtifact?.cacheKey === node.cacheKey &&
        existingArtifact.layoutVersion === ARTIFACT_LAYOUT_VERSION &&
        existingArtifact.compilerVersion === (this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION)
      ) {
        this.progress({
          action: "cached",
          importPath: node.importPath,
          packageName: node.packageName,
          artifactPath: node.artifactPath,
          dependencyCount: node.dependencies.length,
          fileCount: node.files.length,
          standardLibrary: this.standardLibraryPackages.has(node.importPath)
        });
        artifacts.push({ ...reportArtifact, action: "skipped" });
        skipped.push(node.artifactPath);
        continue;
      }
      this.store?.writeAtomic(node.artifactPath, node.artifactSource);
      this.progress({
        action: "built",
        importPath: node.importPath,
        packageName: node.packageName,
        artifactPath: node.artifactPath,
        dependencyCount: node.dependencies.length,
        fileCount: node.files.length,
        standardLibrary: this.standardLibraryPackages.has(node.importPath)
      });
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

  private buildFromFreshArtifacts(): BuildPackageReport | undefined {
    if (!this.store?.read || !this.request.importPath) return undefined;

    const nodes = new Map<string, PackageBuildNode>();
    const visiting = new Set<string>();
    const rootFiles = this.request.files.map(ensureSourceFile);
    const root = this.freshArtifactNode(this.request.importPath, rootFiles, nodes, visiting);
    if (!root) return undefined;

    const ordered = orderedNodesFrom(root, nodes);
    const artifacts = ordered.map((node) => this.reportArtifact(node, "skipped"));
    const skipped = ordered.map((node) => node.artifactPath);
    for (const node of ordered) {
      this.progress({
        action: "cached",
        importPath: node.importPath,
        packageName: node.packageName,
        artifactPath: node.artifactPath,
        dependencyCount: node.dependencies.length,
        fileCount: node.files.length,
        standardLibrary: this.standardLibraryPackages.has(node.importPath)
      });
    }
    return {
      ok: true,
      diagnostics: [],
      artifacts,
      built: [],
      skipped
    };
  }

  private freshArtifactNode(
    importPath: string,
    files: SourceFile[] | undefined,
    nodes: Map<string, PackageBuildNode>,
    visiting: Set<string>
  ): PackageBuildNode | undefined {
    const existing = nodes.get(importPath);
    if (existing) return existing;
    if (visiting.has(importPath)) return undefined;

    const artifactPath = artifactPathForImportPath(resolveArtifactRoot(this.request), importPath);
    const source = this.store?.read?.(artifactPath);
    if (source === undefined) return undefined;
    const archive = parseGoJuniorPackageArchive(source);
    if (!archive || !this.artifactMetadataMatchesRequest(archive.pkgdef, importPath)) return undefined;

    let artifactFiles: SourceFile[] | undefined;
    if (files) {
      artifactFiles = files;
      if (hashSourceFiles(artifactFiles) !== archive.pkgdef.sourceHash) return undefined;
    } else if (this.artifactSourcesOlderThanArtifact(artifactPath, archive.pkgdef)) {
      artifactFiles = sourceFilesFromPkgdef(archive.pkgdef);
    } else {
      artifactFiles = this.freshArtifactSourceFiles(importPath, archive.pkgdef);
      if (!artifactFiles) return undefined;
      const sourceHash = archive.pkgdef.standardLibrary && isAmbientBuildImport(importPath)
        ? archive.pkgdef.sourceHash
        : hashSourceFiles(artifactFiles);
      if (sourceHash !== archive.pkgdef.sourceHash) return undefined;
    }

    visiting.add(importPath);
    const dependencies = [...archive.pkgdef.dependencies].sort();
    const dependencyCacheKeys: string[] = [];
    for (const dependencyPath of dependencies) {
      const dependency = this.freshArtifactNode(dependencyPath, undefined, nodes, visiting);
      if (!dependency) {
        visiting.delete(importPath);
        return undefined;
      }
      dependencyCacheKeys.push(`${dependency.importPath}:${dependency.cacheKey}`);
    }
    visiting.delete(importPath);

    const sortedDependencyCacheKeys = dependencyCacheKeys.sort();
    if (!stringArraysEqual(sortedDependencyCacheKeys, [...archive.pkgdef.dependencyCacheKeys].sort())) return undefined;

    const node: PackageBuildNode = {
      importPath,
      packageName: archive.pkgdef.packageName,
      files: artifactFiles,
      sourceHash: archive.pkgdef.sourceHash,
      cacheKey: archive.pkgdef.cacheKey,
      dependencies,
      dependencyCacheKeys: sortedDependencyCacheKeys,
      exports: archive.pkgdef.exports,
      artifactPath,
      artifactSource: source
    };
    nodes.set(importPath, node);
    if (archive.pkgdef.standardLibrary) this.standardLibraryPackages.add(importPath);
    return node;
  }

  private artifactMetadataMatchesRequest(pkgdef: GoJuniorPackageExportData, importPath: string): boolean {
    return pkgdef.importPath === importPath &&
      pkgdef.layoutVersion === ARTIFACT_LAYOUT_VERSION &&
      pkgdef.compilerVersion === (this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION) &&
      pkgdef.backend === (this.request.backend ?? DEFAULT_BACKEND) &&
      pkgdef.hostSpecVersion === (this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION) &&
      pkgdef.capabilityPolicy === (this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY) &&
      pkgdef.goos === (this.request.goos ?? GOJR_GOOS) &&
      pkgdef.goarch === (this.request.goarch ?? GOJR_GOARCH) &&
      stringArraysEqual(pkgdef.buildTags, resolvedBuildTags(this.request));
  }

  private freshArtifactSourceFiles(importPath: string, pkgdef: GoJuniorPackageExportData): SourceFile[] | undefined {
    if (pkgdef.standardLibrary && isAmbientBuildImport(importPath)) {
      return sourceFilesFromPkgdef(pkgdef);
    }

    const files = this.sourceFilesForImportFast(importPath);
    if (!files) return undefined;
    return files.length === pkgdef.sources.length ? files : undefined;
  }

  private artifactSourcesOlderThanArtifact(artifactPath: string, pkgdef: GoJuniorPackageExportData): boolean {
    if (!this.store?.mtimeMs) return false;
    const artifactMtime = this.store.mtimeMs(artifactPath);
    if (!isFiniteMtime(artifactMtime)) return false;
    const directories = new Set<string>();
    for (const source of pkgdef.sources) {
      if (isSyntheticSourceFilename(source.filename)) continue;
      const sourceMtime = this.store.mtimeMs(source.filename);
      if (!isFiniteMtime(sourceMtime) || sourceMtime > artifactMtime) return false;
      const dir = sourceFilenameDirectory(source.filename);
      if (dir) directories.add(dir);
    }
    for (const dir of directories) {
      const dirMtime = this.store.mtimeMs(dir);
      if (!isFiniteMtime(dirMtime) || dirMtime > artifactMtime) return false;
    }
    return true;
  }

  private sourceFilesForImportFast(importPath: string): SourceFile[] | undefined {
    const explicit = this.packageSources.get(importPath);
    if (explicit) return explicit;
    const stub = stubSourcePackageFiles(importPath);
    if (stub) {
      const files = stub.map(ensureSourceFile);
      this.packageSources.set(importPath, files);
      if (isStubSourcePackageStandardLibrary(importPath)) this.standardLibraryPackages.add(importPath);
      return files;
    }
    try {
      const loaded = this.request.sourcePackageProvider?.load(importPath);
      if (!loaded) return undefined;
      const files = loaded.map(ensureSourceFile);
      this.packageSources.set(importPath, files);
      if (this.request.sourcePackageProvider?.isStandardLibraryPackage?.(importPath)) {
        this.standardLibraryPackages.add(importPath);
      }
      return files;
    } catch {
      return undefined;
    }
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
    const allDependencies = uniqueSorted(parsed.files.flatMap((file) => file.imports.map(importPathFromSpec)));
    const dependencies = allDependencies.filter((dependencyPath) => !isIntrinsicPackageImport(dependencyPath));
    this.progress({
      action: "checking",
      importPath,
      packageName,
      dependencyCount: allDependencies.length,
      fileCount: files.length,
      standardLibrary: this.standardLibraryPackages.has(importPath)
    });
    const sourceDependencies: PackageBuildNode[] = [];
    for (const dependencyPath of allDependencies) {
      if (isIntrinsicPackageImport(dependencyPath)) {
        this.registerIntrinsicPackage(dependencyPath);
        continue;
      }
      const dependencyFiles = this.sourceFilesForImport(dependencyPath, files[0]?.filename ?? REPL_FILENAME);
      if (!dependencyFiles) {
        if (!this.failedPackageLoads.has(dependencyPath)) {
          this.diagnostics.push(buildDiagnostic(
            files[0]?.filename ?? REPL_FILENAME,
            `package ${dependencyPath} is not available to gojr build; provide it with packageSources or --pkg`
          ));
        }
        continue;
      }
      const dependency = this.buildOne(dependencyPath, dependencyFiles, [...stack, importPath]);
      if (dependency) sourceDependencies.push(dependency);
      if (this.hasErrors()) break;
    }

    if (!this.hasErrors()) {
      const ambientTypePackage = isIntrinsicPackageImport(importPath)
        ? standardTypePackage(importPath)
        : undefined;
      const checked: Pick<GoJuniorCheckResult, "diagnostics" | "pkg"> & Partial<Pick<GoJuniorCheckResult, "info">> = ambientTypePackage
        ? this.ambientCheckedPackage(importPath, ambientTypePackage)
        : checkGoJuniorFiles(parsed.files, parsed.statements, [], {
          packageName,
          packagePath: importPath,
          autoImportFmt: false,
          codebaseTxn: this.requireActiveTxn(),
          rollbackOnErrors: false
        });
      this.diagnostics.push(...checked.diagnostics);
      if (!this.hasErrors()) {
        const goos = this.request.goos ?? GOJR_GOOS;
        const goarch = this.request.goarch ?? GOJR_GOARCH;
        const buildTags = resolvedBuildTags(this.request);
        const standardLibrary = this.standardLibraryPackages.has(importPath);
        const sourceHash = hashSourceFiles(files);
        const dependencyCacheKeys = sourceDependencies.map((dependency) => `${dependency.importPath}:${dependency.cacheKey}`).sort();
        const exports = uniqueExports(packageScopeObjects(checked.pkg));
        const cacheKey = stableHash([
          ARTIFACT_LAYOUT_VERSION,
          this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
          this.request.backend ?? DEFAULT_BACKEND,
          this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
          this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
          goos,
          goarch,
          buildTags.join("\n"),
          standardLibrary ? "stdlib" : "workspace",
          importPath,
          sourceHash,
          dependencies.join("\n"),
          dependencyCacheKeys.join("\n")
        ].join("\0"));
        const artifactPath = artifactPathForImportPath(resolveArtifactRoot(this.request), importPath);
        const pkgdef = packageExportData({
          layoutVersion: ARTIFACT_LAYOUT_VERSION,
          compilerVersion: this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
          backend: this.request.backend ?? DEFAULT_BACKEND,
          hostSpecVersion: this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
          capabilityPolicy: this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
          goos,
          goarch,
          buildTags,
          standardLibrary,
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
        });
        const ast = frontFilesToProgramAst(parsed.files, [], [], "info" in checked ? checked.info : undefined, checked.pkg);
        const runtimePlan = packageRuntimePlan(
          importPath,
          packageName,
          checked.pkg,
          "info" in checked ? checked.info.InitOrder ?? [] : []
        );
        const backend = this.request.backend ?? DEFAULT_BACKEND;
        const artifactSource = backend === GOJR_STAGE1_BACKEND
          ? this.stage1ArtifactSource(pkgdef, ast)
          : generatedArtifactSource(pkgdef, files, ast, runtimePlan);
        if (this.hasErrors()) {
          this.visiting.delete(importPath);
          return undefined;
        }
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
          artifactSource
        };
        this.nodes.set(importPath, node);
      }
    }

    this.visiting.delete(importPath);
    return this.nodes.get(importPath);
  }

  private registerIntrinsicPackage(importPath: string): boolean {
    const pkg = standardTypePackage(importPath);
    if (!pkg) return false;
    this.standardLibraryPackages.add(importPath);
    this.codebase.SetPackageInfo(this.requireActiveTxn(), importPath, pkg);
    return true;
  }

  private ambientCheckedPackage(importPath: string, pkg: GoTypesPackage): { diagnostics: Diagnostic[]; pkg: GoTypesPackage } {
    this.codebase.SetPackageInfo(this.requireActiveTxn(), importPath, pkg);
    return { diagnostics: [], pkg };
  }

  private requireActiveTxn(): CodebaseUpdateTxn {
    if (!this.activeTxn || this.activeTxn.IsClosed()) {
      throw new Error("internal error: build package graph has no active Codebase transaction");
    }
    return this.activeTxn;
  }

  private progress(event: BuildProgressEvent): void {
    try {
      this.request.onProgress?.(event);
    } catch {
      // Progress sinks are observability hooks and must not affect compilation.
    }
  }

  private sourceFilesForImport(importPath: string, filename: string): SourceFile[] | undefined {
    if (isIntrinsicPackageImport(importPath)) {
      this.registerIntrinsicPackage(importPath);
      return undefined;
    }
    const explicit = this.packageSources.get(importPath);
    if (explicit) return explicit;
    const stub = stubSourcePackageFiles(importPath);
    if (stub) {
      const files = stub.map(ensureSourceFile);
      this.packageSources.set(importPath, files);
      if (isStubSourcePackageStandardLibrary(importPath)) {
        this.standardLibraryPackages.add(importPath);
      }
      return files;
    }
    try {
      const loaded = this.request.sourcePackageProvider?.load(importPath);
      if (!loaded) return undefined;
      const files = loaded.map(ensureSourceFile);
      this.packageSources.set(importPath, files);
      if (this.request.sourcePackageProvider?.isStandardLibraryPackage?.(importPath)) {
        this.standardLibraryPackages.add(importPath);
      }
      return files;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      this.failedPackageLoads.add(importPath);
      this.diagnostics.push(buildDiagnostic(filename, `could not load package ${importPath}: ${message}`));
      return undefined;
    }
  }

  private orderedNodes(root: PackageBuildNode): PackageBuildNode[] {
    return orderedNodesFrom(root, this.nodes);
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

  private stage1ArtifactSource(pkgdef: GoJuniorPackageExportData, ast: ProgramAst): string {
    const emitted = emitStage1Package(pkgdef, ast);
    this.diagnostics.push(...emitted.diagnostics);
    if (this.hasErrors()) return "";
    return generatedCompiledPackageArtifactSource(pkgdef, emitted.javascript, emitted.wasmBase64);
  }
}

function orderedNodesFrom(root: PackageBuildNode, nodes: Map<string, PackageBuildNode>): PackageBuildNode[] {
  const out: PackageBuildNode[] = [];
  const seen = new Set<string>();
  const visit = (node: PackageBuildNode): void => {
    if (seen.has(node.importPath)) return;
    seen.add(node.importPath);
    for (const dependencyPath of node.dependencies) {
      const dependency = nodes.get(dependencyPath);
      if (dependency) visit(dependency);
    }
    out.push(node);
  };
  visit(root);
  return out;
}

function stringArraysEqual(left: string[], right: string[]): boolean {
  if (left.length !== right.length) return false;
  return left.every((value, index) => value === right[index]);
}

function parseGeneratedArtifactSource(source: string): { layoutVersion?: string; compilerVersion?: string; cacheKey?: string } | undefined {
  const archive = parseGoJuniorPackageArchive(source);
  if (archive) {
    return {
      layoutVersion: archive.pkgdef.layoutVersion,
      compilerVersion: archive.pkgdef.compilerVersion,
      cacheKey: archive.pkgdef.cacheKey
    };
  }
  const artifactJSON = exportedConstJSON(source, "gojrPackageArtifact");
  if (!artifactJSON) return undefined;
  try {
    const value = JSON.parse(artifactJSON) as { layoutVersion?: unknown; compilerVersion?: unknown; cacheKey?: unknown };
    return {
      ...(typeof value.layoutVersion === "string" ? { layoutVersion: value.layoutVersion } : {}),
      ...(typeof value.compilerVersion === "string" ? { compilerVersion: value.compilerVersion } : {}),
      ...(typeof value.cacheKey === "string" ? { cacheKey: value.cacheKey } : {})
    };
  } catch {
    return undefined;
  }
}

function exportedConstJSON(source: string, name: string): string | undefined {
  const startText = `export const ${name} = `;
  const start = source.indexOf(startText);
  if (start < 0) return undefined;
  const jsonStart = start + startText.length;
  const end = source.indexOf(";\n", jsonStart);
  if (end < 0) return undefined;
  return source.slice(jsonStart, end);
}

export function artifactPathForImportPath(artifactRoot: string, importPath: string): string {
  const parts = importPathParts(importPath);
  if (!parts) {
    throw new Error(`invalid import path: ${importPath}`);
  }
  return joinSlash(artifactRoot, ...parts) + ".a";
}

export function resolveArtifactRoot(request: Pick<BuildPackageRequest, "artifactRoot" | "packageCacheParent">): string {
  if (request.artifactRoot && request.artifactRoot.trim() !== "") return trimTrailingSlash(request.artifactRoot);
  if (request.packageCacheParent && request.packageCacheParent.trim() !== "") return joinSlash(request.packageCacheParent, "js_gojr");
  return "~/go/pkg/js_gojr";
}

function resolvedBuildTags(request: Pick<BuildPackageRequest, "goos" | "goarch" | "buildTags">): string[] {
  return [...buildTagSet(request.goos ?? GOJR_GOOS, request.goarch ?? GOJR_GOARCH, request.buildTags)].sort();
}

function buildTagSet(goos: string, goarch: string, extra: string[] | undefined): Set<string> {
  const tags = new Set<string>();
  tags.add(goos);
  tags.add(goarch);
  for (const tag of GOJR_DEFAULT_BUILD_TAGS) tags.add(tag);
  if (isUnixGOOS(goos)) tags.add("unix");
  for (let minor = 1; minor <= 27; minor += 1) {
    tags.add(`go1.${minor}`);
  }
  for (const tag of extra ?? []) {
    const text = tag.trim();
    if (text !== "") tags.add(text);
  }
  return tags;
}

function isUnixGOOS(goos: string): boolean {
  return new Set([
    "aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos",
    "ios", "linux", "netbsd", "openbsd", "solaris"
  ]).has(goos);
}

export function goSourceFileMatchesBuildContext(filename: string, source: string, goos: string, goarch: string, tags: Set<string>): boolean {
  return goSourceFileNameMatchesBuildContext(filename, goos, goarch) && goSourceMatchesBuildConstraints(source, tags);
}

export function buildTagSetForContext(goos: string, goarch: string, extra: string[] | undefined): Set<string> {
  return buildTagSet(goos, goarch, extra);
}

function goSourceMatchesBuildConstraints(source: string, tags: Set<string>): boolean {
  const lines = leadingCommentAndBlankLines(source);
  const goBuild = lines
    .map((line) => line.match(/^\/\/go:build\s+(.+)$/)?.[1]?.trim())
    .filter((line): line is string => !!line);
  if (goBuild.length > 0) {
    return goBuild.every((expr) => evaluateGoBuildExpression(expr, tags));
  }
  const plusBuild = lines
    .map((line) => line.match(/^\/\/\s*\+build\s+(.+)$/)?.[1]?.trim())
    .filter((line): line is string => !!line);
  return plusBuild.length === 0 || plusBuild.every((line) => evaluatePlusBuildLine(line, tags));
}

function leadingCommentAndBlankLines(source: string): string[] {
  const out: string[] = [];
  for (const line of source.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (trimmed === "" || trimmed.startsWith("//")) {
      out.push(trimmed);
      continue;
    }
    break;
  }
  return out;
}

function evaluatePlusBuildLine(line: string, tags: Set<string>): boolean {
  return line.split(/\s+/)
    .filter((option) => option !== "")
    .some((option) => option.split(",").every((term) => {
      if (term.startsWith("!")) return !tags.has(term.slice(1));
      return tags.has(term);
    }));
}

function evaluateGoBuildExpression(expr: string, tags: Set<string>): boolean {
  const tokens = expr.match(/[A-Za-z0-9_.]+|&&|\|\||!|\(|\)/g) ?? [];
  let index = 0;

  const parseOr = (): boolean => {
    let value = parseAnd();
    while (tokens[index] === "||") {
      index += 1;
      value = parseAnd() || value;
    }
    return value;
  };

  const parseAnd = (): boolean => {
    let value = parseUnary();
    while (tokens[index] === "&&") {
      index += 1;
      value = parseUnary() && value;
    }
    return value;
  };

  const parseUnary = (): boolean => {
    const token = tokens[index];
    if (token === "!") {
      index += 1;
      return !parseUnary();
    }
    if (token === "(") {
      index += 1;
      const value = parseOr();
      if (tokens[index] === ")") index += 1;
      return value;
    }
    if (token && /^[A-Za-z0-9_.]+$/.test(token)) {
      index += 1;
      return tags.has(token);
    }
    index += 1;
    return false;
  };

  const value = parseOr();
  return index >= tokens.length && value;
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

function goSourceFileNameMatchesBuildContext(filename: string, goos: string, goarch: string): boolean {
  const name = filename.split("/").pop()?.replace(/\.go$/, "") ?? filename.replace(/\.go$/, "");
  const parts = name.split("_");
  if (parts.length < 2) return true;
  const last = parts[parts.length - 1] ?? "";
  const prev = parts[parts.length - 2] ?? "";
  if (knownGOOS.has(prev) && knownGOARCH.has(last)) {
    return prev === goos && last === goarch;
  }
  if (knownGOOS.has(last)) return last === goos;
  if (knownGOARCH.has(last)) return last === goarch;
  return true;
}

const knownGOOS = new Set([
  "aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios",
  "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows",
  GOJR_GOOS
]);

const knownGOARCH = new Set([
  "386", "amd64", "amd64p32", "arm", "arm64", "arm64be", "loong64", "mips",
  "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "ppc64",
  "ppc64le", "riscv", "riscv64", "s390", "s390x", "sparc", "sparc64",
  ["w", "a", "s", "m"].join(""),
  GOJR_GOARCH
]);

function isAmbientBuildImport(path: string): boolean {
  return isIntrinsicPackageImport(path);
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

function intrinsicSyntheticSourceFile(importPath: string): SourceFile {
  return {
    filename: `gojr:intrinsic/${importPath}`,
    source: `package ${intrinsicPackageName(importPath) ?? importPath.split("/").filter(Boolean).at(-1) ?? importPath}\n`
  };
}

function sourceFilesFromPkgdef(pkgdef: GoJuniorPackageExportData): SourceFile[] {
  return pkgdef.sources.map((source) => ({
    filename: source.filename,
    source: ""
  }));
}

function isFiniteMtime(value: number | undefined): value is number {
  return typeof value === "number" && Number.isFinite(value) && value >= 0;
}

function isSyntheticSourceFilename(filename: string): boolean {
  return filename.startsWith("gojr:");
}

function sourceFilenameDirectory(filename: string): string | undefined {
  const slash = Math.max(filename.lastIndexOf("/"), filename.lastIndexOf("\\"));
  if (slash <= 0) return undefined;
  return filename.slice(0, slash);
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

function packageScopeObjects(pkg: GoTypesPackage): GoTypesObject[] {
  return pkg.Scope().Names().flatMap((name) => {
    if (isGoJuniorSyntheticCheckName(name) || name === "fmt") return [];
    const object = pkg.Scope().Lookup(name);
    if (object === null || object.constructor.name === "PkgName") return [];
    if (object.Pkg() !== pkg && !(pkg.Path() === "unsafe" && object instanceof GoTypesBuiltin)) return [];
    return [object];
  });
}

function uniqueExports(objects: GoTypesObject[]): BuildExport[] {
  const exports = new Map<string, BuildExport>();
  for (const object of objects) {
    const kind = exportKindForObject(object);
    if (!kind || !object.Exported()) continue;
    const typ = object.Type();
    const item: BuildExport = {
      name: object.Name(),
      kind,
      typeText: typ?.String() ?? "<nil>"
    };
    if (kind === "type") {
      item.underlyingTypeText = typ?.Underlying().String() ?? "<nil>";
    }
    exports.set(`${kind}:${object.Name()}`, item);
  }
  return [...exports.values()].sort((left, right) => left.name.localeCompare(right.name) || left.kind.localeCompare(right.kind));
}

function exportKindForObject(object: GoTypesObject): BuildExport["kind"] | undefined {
  if (object instanceof GoTypesConst) return "const";
  if (object instanceof GoTypesFunc) return "func";
  if (object instanceof GoTypesTypeName) return "type";
  if (object instanceof GoTypesVar) return "var";
  if (object instanceof GoTypesBuiltin) return "func";
  return undefined;
}

function packageExportData(data: Omit<GoJuniorPackageExportData, "exportFormat" | "exportVersion" | "exportIndex">): GoJuniorPackageExportData {
  const exportIndex: Record<string, GoJuniorPackageExportIndexEntry> = {};
  data.exports.forEach((item, exportIndexNumber) => {
    exportIndex[item.name] = {
      exportIndex: exportIndexNumber,
      kind: item.kind,
      typeText: item.typeText,
      ...(item.underlyingTypeText ? { underlyingTypeText: item.underlyingTypeText } : {})
    };
  });
  return {
    exportFormat: "gojr-iexport",
    exportVersion: 1,
    exportIndex,
    ...data
  };
}

function packageRuntimePlan(
  importPath: string,
  packageName: string,
  pkg: GoTypesPackage,
  initOrder: Array<{ Lhs: Array<{ Name(): string }> }>
): GoJuniorPackageRuntimePlan {
  const objects = packageScopeObjects(pkg);
  return {
    importPath,
    packageName,
    exportedNames: objects.filter((object) => object.Exported()).map((object) => object.Name()).sort(),
    constants: objects.flatMap((object): GoJuniorPackageRuntimeConstant[] => {
      if (!(object instanceof GoTypesConst)) return [];
      const type = object.Type();
      const typeText = type ? checkedConstantBuildTypeText(type, pkg) : undefined;
      return [{
        name: object.Name(),
        value: object.Val(),
        ...(typeText ? { typeText } : {})
      }];
    }).sort((left, right) => left.name.localeCompare(right.name)),
    variables: objects.flatMap((object): GoJuniorPackageRuntimeVariable[] => {
      if (!(object instanceof GoTypesVar)) return [];
      const type = object.Type();
      return [{
        name: object.Name(),
        ...(type ? { typeText: GoTypesTypeString(type, GoTypesRelativeTo(pkg)) } : {})
      }];
    }).sort((left, right) => left.name.localeCompare(right.name)),
    varInitOrder: initOrder.map((initializer) =>
      initializer.Lhs.map((object) => object.Name()).filter((name) => name !== "_")
    )
  };
}

function checkedConstantBuildTypeText(type: GoTypesType, pkg: GoTypesPackage): string | undefined {
  const text = GoTypesTypeString(type, GoTypesRelativeTo(pkg));
  return text.startsWith("untyped ") ? undefined : text;
}

function generatedArtifactSource(pkgdef: GoJuniorPackageExportData, files: SourceFile[], ast: ProgramAst, runtimePlan: GoJuniorPackageRuntimePlan): string {
  const javascript = generatedArtifactJavaScript(pkgdef, files, ast, runtimePlan);
  const pkgdefWithRuntime: GoJuniorPackageExportData = {
    ...pkgdef,
    runtime: {
      ast,
      runtimePlan
    }
  };
  return writeArArchive([
    { name: PKGDEF_MEMBER, data: GOJR_EXPORT_MAGIC + artifactJSONString(pkgdefWithRuntime) + "\n" },
    { name: JAVASCRIPT_MEMBER, data: javascript }
  ]);
}

export function generatedMixedWasmArtifactSource(pkgdef: GoJuniorPackageExportData, javascript: string, wasmBase64: string): string {
  return generatedCompiledPackageArtifactSource(pkgdef, javascript, wasmBase64);
}

export function generatedCompiledPackageArtifactSource(pkgdef: GoJuniorPackageExportData, javascript: string, wasmBase64?: string): string {
  const members = [
    { name: PKGDEF_MEMBER, data: GOJR_EXPORT_MAGIC + artifactJSONString(pkgdef) + "\n" },
    { name: JAVASCRIPT_MEMBER, data: javascript }
  ];
  if (wasmBase64 !== undefined) members.push({ name: WASM_MEMBER, data: wasmBase64 });
  return writeArArchive([
    ...members
  ]);
}

function generatedArtifactJavaScript(
  artifact: GoJuniorPackageExportData,
  files: SourceFile[],
  ast: ProgramAst,
  _runtimePlan: GoJuniorPackageRuntimePlan
): string {
  const sources: GoJuniorPackageSourcePayload[] = files.map((file) => ({
    filename: file.filename,
    hash: stableHash(file.source),
    source: file.source
  }));
  const compiledFunctions = generatedCompiledFunctionBodyLines(ast.functions);
  const artifactHeader: GoJuniorPackageExportData = {
    ...artifact
  };
  delete artifactHeader.runtime;
  return [
    "// Code generated by gojr build; DO NOT EDIT.",
    "// This member is the generated JavaScript payload inside the Go-junior package archive.",
    "// __.PKGDEF carries export metadata; this module carries runtime-ready package input.",
    `export const gojrPackageArtifact = ${JSON.stringify(artifactHeader, null, 2)};`,
    `export const gojrPackageSources = ${JSON.stringify(sources, null, 2)};`,
    "export async function instantiateGoJrPackage(runtime, options = {}) {",
    "  if (!runtime || typeof runtime.evaluatePackageArtifact !== \"function\") {",
    "    throw new Error(\"gojr package artifact requires runtime.evaluatePackageArtifact\");",
    "  }",
    "  const __gojrPayload = options.runtimePayload || (options.packageArtifact && options.packageArtifact.runtime);",
    "  if (!__gojrPayload || !__gojrPayload.ast || !__gojrPayload.runtimePlan) {",
    "    throw new Error(\"gojr package artifact runtime payload must be supplied from __.PKGDEF\");",
    "  }",
    "  const __gojrPackageResult = await runtime.evaluatePackageArtifact(__gojrPayload.ast, __gojrPayload.runtimePlan, {",
    "    ...options,",
    "    importPath: gojrPackageArtifact.importPath,",
    "    packageName: gojrPackageArtifact.packageName",
    "  });",
    "  return {",
    "    ...__gojrPackageResult,",
    "    compiledFunctions: __gojrBindCompiledFunctions(__gojrPackageResult)",
    "  };",
    "}",
    ...compiledFunctions,
    "const gojrPackageDefault = {",
    "  artifact: gojrPackageArtifact,",
    "  sources: gojrPackageSources,",
    "  compiledFunctions: gojrCompiledFunctionBodies,",
    "  instantiateGoJrPackage",
    "};",
    "export { gojrPackageDefault as default };",
    ""
  ].join("\n");
}

function artifactJSONString(value: unknown): string {
  return JSON.stringify(value, (_key, item) => {
    if (typeof item === "bigint") return { __gojrBigInt: item.toString() };
    return item;
  }, 2);
}

interface CompiledFunctionDescriptor {
  key: string;
  name: string;
  receiver?: string;
  source?: string;
}

function generatedCompiledFunctionBodyLines(functions: FunctionDecl[]): string[] {
  const descriptors = compiledFunctionDescriptors(functions);
  return [
    `export const gojrCompiledFunctionDescriptors = ${JSON.stringify(descriptors, null, 2)};`,
    "async function __gojrCallCompiledFunction(__gojrPackageResult, __gojrName, __gojrArgs) {",
    "  const __gojrPackage = __gojrPackageResult && __gojrPackageResult.package ? __gojrPackageResult.package : __gojrPackageResult;",
    "  const __gojrContext = __gojrPackageResult && (__gojrPackageResult.context || __gojrPackageResult.__gojrContext);",
    "  const __gojrFn = __gojrPackage && __gojrPackage[__gojrName];",
    "  if (!__gojrFn) throw new Error(`gojr compiled package function ${__gojrName} is not available`);",
    "  if (__gojrFn.kind === \"GoJuniorFunction\" && typeof __gojrFn.call === \"function\") {",
    "    if (!__gojrContext) throw new Error(`gojr compiled package function ${__gojrName} requires an EvaluationContext`);",
    "    return await __gojrFn.call(__gojrArgs, __gojrContext);",
    "  }",
    "  if (typeof __gojrFn === \"function\") return await __gojrFn(...__gojrArgs);",
    "  if (__gojrFn && typeof __gojrFn.call === \"function\") return await __gojrFn.call(__gojrPackage, ...__gojrArgs);",
    "  throw new Error(`gojr package member ${__gojrName} is not callable`);",
    "}",
    "function __gojrBindCompiledFunctions(__gojrPackageResult) {",
    "  const __gojrBound = {};",
    "  for (const [__gojrKey, __gojrFn] of Object.entries(gojrCompiledFunctionBodies)) {",
    "    __gojrBound[__gojrKey] = async (...__gojrArgs) => await __gojrFn(__gojrPackageResult, ...__gojrArgs);",
    "  }",
    "  return __gojrBound;",
    "}",
    "export const gojrCompiledFunctionBodies = {",
    ...descriptors.flatMap((descriptor, index) => compiledFunctionEntryLines(descriptor, index === descriptors.length - 1)),
    "};",
    "export async function instantiateGoJrCompiledPackage(runtime, options = {}) {",
    "  const __gojrPackageResult = await instantiateGoJrPackage(runtime, options);",
    "  return {",
    "    ...__gojrPackageResult,",
    "    compiledFunctions: __gojrBindCompiledFunctions(__gojrPackageResult)",
    "  };",
    "}"
  ];
}

function compiledFunctionDescriptors(functions: FunctionDecl[]): CompiledFunctionDescriptor[] {
  return functions.map((declaration, index) => {
    const receiver = declaration.receiver?.type.text;
    const name = declaration.name;
    const baseKey = receiver ? `${receiver}.${name}` : name;
    const key = name === "init" || functions.some((other, otherIndex) =>
      otherIndex !== index &&
      other.name === name &&
      (other.receiver?.type.text ?? "") === (receiver ?? "")
    )
      ? `${baseKey}#${index}`
      : baseKey;
    return {
      key,
      name,
      ...(receiver ? { receiver } : {}),
      ...(declaration.source ? { source: declaration.source } : {})
    };
  });
}

function compiledFunctionEntryLines(descriptor: CompiledFunctionDescriptor, last: boolean): string[] {
  const jsName = compiledFunctionJavaScriptName(descriptor);
  return [
    `  ${JSON.stringify(descriptor.key)}: async function ${jsName}(__gojrPackageResult, ...__gojrArgs) {`,
    `    return await __gojrCallCompiledFunction(__gojrPackageResult, ${JSON.stringify(descriptor.name)}, __gojrArgs);`,
    `  }${last ? "" : ","}`
  ];
}

function compiledFunctionJavaScriptName(descriptor: CompiledFunctionDescriptor): string {
  const text = `gojr$${descriptor.key}`;
  return text.replace(/[^A-Za-z0-9_$]/g, "_").replace(/^[^A-Za-z_$]/, "_");
}

export function parseGoJuniorPackageArchive(source: string): GoJuniorPackageArchive | undefined {
  const members = readArArchive(source);
  if (!members) return undefined;
  const pkgdefMember = members[0];
  if (pkgdefMember?.name !== PKGDEF_MEMBER || !pkgdefMember.data.startsWith(GOJR_EXPORT_MAGIC)) return undefined;
  const javascript = members.find((member) => member.name === JAVASCRIPT_MEMBER)?.data ?? "";
  const wasmBase64 = members.find((member) => member.name === WASM_MEMBER)?.data;
  try {
    const pkgdef = reviveArtifactValue(JSON.parse(pkgdefMember.data.slice(GOJR_EXPORT_MAGIC.length))) as GoJuniorPackageExportData;
    if (pkgdef.exportFormat !== "gojr-iexport" || pkgdef.exportVersion !== 1) return undefined;
    return {
      pkgdef,
      javascript,
      ...(wasmBase64 !== undefined ? { wasmBase64 } : {}),
      ...(pkgdef.runtime ? { runtime: pkgdef.runtime } : {}),
      members
    };
  } catch {
    return undefined;
  }
}

function reviveArtifactValue(value: unknown): unknown {
  if (Array.isArray(value)) return value.map((item) => reviveArtifactValue(item));
  if (value && typeof value === "object") {
    const record = value as Record<string, unknown>;
    if (typeof record.__gojrBigInt === "string") return BigInt(record.__gojrBigInt);
    const out: Record<string, unknown> = {};
    for (const [key, item] of Object.entries(record)) out[key] = reviveArtifactValue(item);
    return out;
  }
  return value;
}

function writeArArchive(members: Array<{ name: string; data: string }>): string {
  return AR_MAGIC + members.map((member) => writeArMember(member.name, member.data)).join("");
}

function writeArMember(name: string, data: string): string {
  const arName = name;
  if (arName.length > 16) throw new Error(`ar member name is too long: ${name}`);
  const size = utf8ByteLength(data);
  const header = [
    arName.padEnd(16, " "),
    "0".padEnd(12, " "),
    "0".padEnd(6, " "),
    "0".padEnd(6, " "),
    "100644".padEnd(8, " "),
    String(size).padEnd(10, " "),
    "`\n"
  ].join("");
  return header + data + (size % 2 === 1 ? "\n" : "");
}

function readArArchive(source: string): Array<{ name: string; data: string }> | undefined {
  if (!source.startsWith(AR_MAGIC)) return undefined;
  const bytes = utf8Encode(source);
  const members: Array<{ name: string; data: string }> = [];
  let offset = utf8ByteLength(AR_MAGIC);
  while (offset < bytes.length) {
    if (offset + 60 > bytes.length) return undefined;
    const header = asciiDecode(bytes.subarray(offset, offset + 60));
    if (header.slice(58, 60) !== "`\n") return undefined;
    const rawName = header.slice(0, 16).trim();
    const sizeText = header.slice(48, 58).trim();
    const size = Number.parseInt(sizeText, 10);
    if (!Number.isFinite(size) || size < 0) return undefined;
    const dataStart = offset + 60;
    const dataEnd = dataStart + size;
    if (dataEnd > bytes.length) return undefined;
    members.push({
      name: rawName.endsWith("/") ? rawName.slice(0, -1) : rawName,
      data: utf8Decode(bytes.subarray(dataStart, dataEnd))
    });
    offset = dataEnd + (size % 2);
  }
  return members;
}

function utf8ByteLength(text: string): number {
  return utf8Encode(text).length;
}

function utf8Encode(text: string): Uint8Array {
  return new TextEncoder().encode(text);
}

function utf8Decode(bytes: Uint8Array): string {
  return new TextDecoder("utf-8").decode(bytes);
}

function asciiDecode(bytes: Uint8Array): string {
  let out = "";
  for (const byte of bytes) out += String.fromCharCode(byte);
  return out;
}

function hashSourceFiles(files: SourceFile[]): string {
  return stableHash(files
    .map((file) => `${file.filename}\0${file.source.length}\0${file.source}`)
    .join("\0"));
}

function stableHash(text: string): string {
  return blake3HashString(text);
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

function importPathParts(importPath: string): string[] | undefined {
  const text = importPath.trim();
  if (text === "" || text.startsWith("/") || text.includes("\\") || text.includes("//")) return undefined;
  const parts = text.split("/");
  if (parts.length === 0 || parts.some((part) => part === "" || part === "." || part === "..")) return undefined;
  return parts;
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
