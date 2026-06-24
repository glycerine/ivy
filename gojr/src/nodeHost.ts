import { existsSync, mkdirSync, readFileSync, readdirSync, renameSync, rmSync, statSync, writeFileSync } from "node:fs";
import { basename, dirname, isAbsolute, join, relative, resolve, sep } from "node:path";
import { homedir } from "node:os";
import {
  buildPackages,
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
import type { SourceFile } from "./diagnostics.js";

export interface NodeBuildPackageRequest extends BuildPackageRequest {
  sourceRoots?: string[];
}

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
