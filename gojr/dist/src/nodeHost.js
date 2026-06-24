import { existsSync, mkdirSync, readFileSync, readdirSync, renameSync, rmSync, statSync, writeFileSync } from "node:fs";
import { basename, dirname, isAbsolute, join, relative, resolve, sep } from "node:path";
import { homedir } from "node:os";
import { buildPackages, inspectPackageJavaScript, parseGoJuniorPackageArchive, resolveArtifactRoot } from "./build.js";
export function defaultPackageCacheParent() {
    return join(homedir(), "go", "pkg");
}
export function createNodeArtifactStore() {
    return {
        read(artifactPath) {
            try {
                return readFileSync(artifactPath, "utf8");
            }
            catch (error) {
                if (isNodeErrorCode(error, "ENOENT"))
                    return undefined;
                throw error;
            }
        },
        writeAtomic(artifactPath, source) {
            const dir = dirname(artifactPath);
            mkdirSync(dir, { recursive: true });
            const tmp = join(dir, `.${basename(artifactPath)}.${process.pid}.tmp`);
            writeFileSync(tmp, source, "utf8");
            renameSync(tmp, artifactPath);
        }
    };
}
export function createNodeSourcePackageProvider(sourceRoots = []) {
    const roots = [...new Set(sourceRoots
            .map((root) => String(root || "").trim())
            .filter((root) => root !== "")
            .map((root) => resolve(expandHome(root))))];
    if (roots.length === 0)
        return undefined;
    return {
        load(importPath) {
            const parts = String(importPath || "").split("/").filter(Boolean);
            if (parts.length === 0 || parts.some((part) => part === "." || part === ".." || part.includes(sep))) {
                throw new Error(`invalid import path: ${importPath}`);
            }
            for (const root of roots) {
                const dir = resolve(root, ...parts);
                const rel = relative(root, dir);
                if (rel === "" || rel.startsWith("..") || isAbsolute(rel))
                    continue;
                let stat;
                try {
                    stat = statSync(dir);
                }
                catch (error) {
                    if (isNodeErrorCode(error, "ENOENT"))
                        continue;
                    throw error;
                }
                if (!stat.isDirectory())
                    throw new Error(`${dir} is not a directory`);
                const names = readdirSync(dir)
                    .filter((name) => !name.startsWith(".") && name.endsWith(".go") && !name.endsWith("_test.go"))
                    .sort();
                if (names.length === 0)
                    throw new Error(`${dir} contains no non-test .go files`);
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
export function buildPackagesOnNode(request) {
    return buildPackages(normalizeNodeBuildRequest(request), createNodeArtifactStore());
}
export function inspectPackageJavaScriptOnNode(request) {
    return inspectPackageJavaScript(normalizeNodeBuildRequest(request));
}
export function packageCacheOnNode(request) {
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
                return { ok: false, diagnostics: [`unknown cache command ${JSON.stringify(request.action)}`], action: "path", root };
        }
    }
    catch (error) {
        return {
            ok: false,
            diagnostics: [error instanceof Error ? error.message : String(error)],
            action: request.action,
            root: ""
        };
    }
}
export function listPackageArtifactCache(root) {
    if (!existsSync(root))
        return [];
    const rootStat = statSync(root);
    if (!rootStat.isDirectory())
        throw new Error(`${root} is not a directory`);
    const entries = [];
    walkFiles(root, (artifactPath) => {
        if (!artifactPath.endsWith(".a"))
            return;
        const stat = statSync(artifactPath);
        const entry = {
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
export function clearPackageArtifactCache(root) {
    const entries = listPackageArtifactCache(root);
    if (entries.length === 0)
        return [];
    rmSync(root, { recursive: true, force: true });
    return entries.map((entry) => entry.path);
}
function normalizeNodeBuildRequest(request) {
    const normalized = { ...request };
    if (!normalized.artifactRoot && !normalized.packageCacheParent) {
        normalized.packageCacheParent = defaultPackageCacheParent();
    }
    if (normalized.artifactRoot)
        normalized.artifactRoot = expandHome(normalized.artifactRoot);
    if (normalized.packageCacheParent)
        normalized.packageCacheParent = expandHome(normalized.packageCacheParent);
    const provider = createNodeSourcePackageProvider(normalized.sourceRoots ?? []);
    if (provider)
        normalized.sourcePackageProvider = provider;
    return normalized;
}
function resolveNodeArtifactRoot(request) {
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
function walkFiles(root, visit) {
    for (const entry of readdirSync(root, { withFileTypes: true })) {
        const path = join(root, entry.name);
        if (entry.isDirectory()) {
            walkFiles(path, visit);
        }
        else if (entry.isFile()) {
            visit(path);
        }
    }
}
function cacheImportPath(root, artifactPath) {
    const rel = relative(root, artifactPath);
    const slash = rel.split(sep).join("/");
    return slash.endsWith(".a") ? slash.slice(0, -2) : slash;
}
function expandHome(path) {
    if (path === "~")
        return homedir();
    if (path.startsWith("~/"))
        return join(homedir(), path.slice(2));
    return path;
}
function isNodeErrorCode(error, code) {
    return typeof error === "object" && error !== null && error.code === code;
}
