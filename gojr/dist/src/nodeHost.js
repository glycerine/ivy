import { existsSync, mkdirSync, readFileSync, readdirSync, renameSync, rmSync, statSync, writeFileSync } from "node:fs";
import { basename, delimiter, dirname, isAbsolute, join, relative, resolve, sep } from "node:path";
import { arch as nodeArch, homedir, platform as nodePlatform } from "node:os";
import { buildTagSetForContext, buildPackages, collectSourceImportPaths, GOJR_GOARCH, GOJR_GOOS, goSourceFileMatchesBuildContext, inspectPackageJavaScript, parseGoJuniorPackageArchive, resolveArtifactRoot } from "./build.js";
import { hasErrorDiagnostics, REPL_FILENAME } from "./diagnostics.js";
import { formatBuildProgressEvent } from "./hostProtocol.js";
import { compilePackageSourceFiles, compileSourceFiles } from "./compile.js";
import { collectSpreadsheetFixtureFormulaSourceFiles, parseSpreadsheetFixtureJson, runSpreadsheetFixture } from "./fixture.js";
import { isHostResolvedSourceImport } from "./intrinsicPackages.js";
import { stubSourcePackageFiles } from "./stubPackages.js";
import { parseSheetJson, parseSheetsJson } from "./jsonInput.js";
import { evaluateSource, evaluateSourceFiles, evaluateSourcePackageGraph, runMainSourcePackageFiles, testSourceFiles } from "./runtime.js";
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
    const roots = nodeSourceRoots(sourceRoots);
    if (roots.length === 0)
        return undefined;
    return {
        load(importPath) {
            const parts = String(importPath || "").split("/").filter(Boolean);
            if (parts.length === 0 || parts.some((part) => part === "." || part === ".." || part.includes(sep))) {
                throw new Error(`invalid import path: ${importPath}`);
            }
            for (const root of roots) {
                const dir = nodeSourcePackageDir(root, importPath, parts);
                if (!dir)
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
                    .filter((name) => !name.startsWith(".") && !name.startsWith("_") && name.endsWith(".go") && !name.endsWith("_test.go"))
                    .sort();
                const loaded = names.flatMap((name) => {
                    const filename = join(dir, name);
                    const source = readFileSync(filename, "utf8");
                    if (!goSourceFileMatchesBuildContext(name, source, root.goos, root.goarch, root.tags))
                        return [];
                    return {
                        filename,
                        source
                    };
                });
                if (loaded.length === 0)
                    throw new Error(`${dir} contains no Go source files matching GOOS=${root.goos} GOARCH=${root.goarch}`);
                return loaded;
            }
            return undefined;
        },
        isStandardLibraryPackage(importPath) {
            const parts = String(importPath || "").split("/").filter(Boolean);
            if (parts.length === 0 || parts.some((part) => part === "." || part === ".." || part.includes(sep)))
                return false;
            return roots.some((root) => {
                if (!root.standardLibrary)
                    return false;
                const dir = nodeSourcePackageDir(root, importPath, parts);
                if (!dir)
                    return false;
                try {
                    return statSync(dir).isDirectory();
                }
                catch {
                    return false;
                }
            });
        }
    };
}
function nodeSourceRoots(sourceRoots) {
    const roots = [];
    const seen = new Map();
    const add = (root, standardLibrary) => {
        const text = String(root || "").trim();
        if (text === "")
            return;
        const path = resolve(expandHome(text));
        if (!existsSync(path))
            return;
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
    for (const root of sourceRoots)
        add(root, false);
    for (const root of candidateGOROOTSourceRoots())
        add(root, true);
    for (const root of candidateGOPATHSourceRoots())
        add(root, false);
    return roots;
}
function nativeStandardLibraryBuildContext() {
    return {
        goos: nativeGOOS(),
        goarch: nativeGOARCH()
    };
}
function nativeGOOS() {
    switch (nodePlatform()) {
        case "win32": return "windows";
        case "sunos": return "solaris";
        default: return nodePlatform();
    }
}
function nativeGOARCH() {
    switch (nodeArch()) {
        case "x64": return "amd64";
        case "ia32": return "386";
        case "mipsel": return "mipsle";
        default: return nodeArch();
    }
}
function nodeSourcePackageDir(root, importPath, parts) {
    if (root.modulePath) {
        if (importPath !== root.modulePath && !importPath.startsWith(`${root.modulePath}/`))
            return undefined;
        const suffix = importPath === root.modulePath ? "" : importPath.slice(root.modulePath.length + 1);
        const suffixParts = suffix === "" ? [] : suffix.split("/").filter(Boolean);
        const dir = resolve(root.path, ...suffixParts);
        return isWithinRoot(root.path, dir) ? dir : undefined;
    }
    const dir = resolve(root.path, ...parts);
    return isWithinRoot(root.path, dir) ? dir : undefined;
}
function isWithinRoot(root, dir) {
    const rel = relative(root, dir);
    return rel === "" || (!rel.startsWith("..") && !isAbsolute(rel));
}
function moduleSourceRoot(root) {
    try {
        const modulePath = parseGoModModule(readFileSync(join(root, "go.mod"), "utf8"));
        return modulePath ? { modulePath } : {};
    }
    catch {
        return {};
    }
}
function parseGoModModule(source) {
    for (const line of source.split(/\r?\n/)) {
        const trimmed = line.trim();
        if (trimmed === "" || trimmed.startsWith("//"))
            continue;
        const match = /^module\s+(\S+)\s*$/.exec(trimmed);
        if (match?.[1])
            return match[1];
        return undefined;
    }
    return undefined;
}
function candidateGOROOTSourceRoots() {
    const roots = [];
    const addGOROOT = (root) => {
        const text = String(root || "").trim();
        if (text === "")
            return;
        roots.push(join(expandHome(text), "src"));
    };
    addGOROOT(process.env.GOROOT);
    roots.push("/usr/local/go1.27rc1/src", "/usr/local/go1.26.4/src", "/usr/local/go/src", "/opt/homebrew/opt/go/libexec/src");
    return roots;
}
function candidateGOPATHSourceRoots() {
    const gopath = String(process.env.GOPATH || "").trim();
    const roots = gopath === "" ? [join(homedir(), "go")] : gopath.split(delimiter);
    return roots
        .map((root) => String(root || "").trim())
        .filter((root) => root !== "")
        .map((root) => join(expandHome(root), "src"));
}
export function buildPackagesOnNode(request) {
    return buildPackages(normalizeNodeBuildRequest(request), createNodeArtifactStore());
}
export function inspectPackageJavaScriptOnNode(request) {
    return inspectPackageJavaScript(normalizeNodeBuildRequest(request));
}
export async function evaluateSourceWithPackagesOnNode(request) {
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
export async function evaluateSourceFilesWithPackagesOnNode(request) {
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
export async function testSourceFilesWithPackagesOnNode(request) {
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
export async function runMainSourceFilesWithPackagesOnNode(request) {
    const options = evaluationOptionsFromNodeRequest(request);
    const rootFiles = rootSourceFilesFromRequest(request);
    const packageSources = packageSourceMapFromSpecs(request.packages ?? []);
    const buildRequest = {
        files: rootFiles,
        sourceRoots: request.sourceRoots ?? []
    };
    if (request.importPath)
        buildRequest.importPath = request.importPath;
    if (packageSources)
        buildRequest.packageSources = packageSources;
    if (request.artifactRoot)
        buildRequest.artifactRoot = request.artifactRoot;
    if (request.packageCacheParent)
        buildRequest.packageCacheParent = request.packageCacheParent;
    if (request.progress)
        buildRequest.progress = request.progress;
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
        ...(provider ? { sourcePackageProvider: provider } : {})
    });
    return result;
}
function packageSourceMapFromSpecs(specs) {
    if (specs.length === 0)
        return undefined;
    const sources = {};
    for (const spec of specs) {
        if (!spec?.importPath)
            continue;
        sources[spec.importPath] = spec.files ?? [];
    }
    return Object.keys(sources).length === 0 ? undefined : sources;
}
export function compileSourceFilesWithPackagesOnNode(request) {
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
export async function runSpreadsheetFixtureWithPackagesOnNode(request) {
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
export async function loadSourcePackagesForRootFilesOnNode(rootFiles, specs = [], baseOptions = {}, sourceRoots = []) {
    const diagnostics = [];
    const explicit = new Map();
    const initialSpecs = [];
    for (const spec of specs) {
        if (!spec?.importPath)
            continue;
        explicit.set(spec.importPath, spec);
        initialSpecs.push(spec);
    }
    const provider = createNodeSourcePackageProvider(sourceRoots);
    const rootImports = collectSourceImportPaths(rootFiles);
    diagnostics.push(...rootImports.diagnostics);
    if (!hasErrorDiagnostics(diagnostics)) {
        for (const importPath of rootImports.imports) {
            if (explicit.has(importPath))
                continue;
            if (isAmbientSourceImport(importPath))
                continue;
            const files = loadStubOrSourcePackageFromProvider(provider, importPath, rootFiles[0]?.filename ?? REPL_FILENAME, diagnostics);
            if (!files)
                continue;
            const spec = { importPath, files };
            explicit.set(importPath, spec);
            initialSpecs.push(spec);
        }
    }
    if (hasErrorDiagnostics(diagnostics)) {
        return { packages: {}, packageInfos: {}, diagnostics, output: [] };
    }
    const graphOptions = { ...baseOptions };
    if (provider)
        graphOptions.sourcePackageProvider = provider;
    const result = await evaluateSourcePackageGraph(initialSpecs, graphOptions);
    return {
        packages: result.packages,
        packageInfos: result.packageInfos,
        diagnostics: [...diagnostics, ...result.diagnostics],
        output: result.output
    };
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
function compileSourcePackagesForRootFilesOnNode(rootFiles, specs = [], baseOptions = {}, sourceRoots = []) {
    const diagnostics = [];
    const packageInfos = {};
    const explicit = new Map();
    for (const spec of specs) {
        if (spec?.importPath)
            explicit.set(spec.importPath, spec);
    }
    const provider = createNodeSourcePackageProvider(sourceRoots);
    const loading = [];
    const loaded = new Set();
    const loadImport = (importPath, requestedFrom) => {
        if (loaded.has(importPath) || hasErrorDiagnostics(diagnostics))
            return;
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
        if (!spec)
            return;
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
        if (spec?.importPath)
            loadImport(spec.importPath, sourcePackageFilename(spec));
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
function isAmbientSourceImport(importPath) {
    return isHostResolvedSourceImport(importPath);
}
function rootSourceFilesFromRequest(request) {
    if (request.files && request.files.length > 0)
        return request.files;
    return [{
            filename: REPL_FILENAME,
            source: request.source ?? ""
        }];
}
function evaluationOptionsFromNodeRequest(request) {
    const options = {};
    if (request.sheetJSON)
        options.sheet = parseSheetJson(request.sheetJSON);
    if (request.sheetsJSON)
        options.sheets = parseSheetsJson(request.sheetsJSON);
    return options;
}
function loadSourcePackageFromProvider(provider, importPath, requestedFrom, diagnostics) {
    if (!provider)
        return undefined;
    try {
        return provider.load(importPath);
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        diagnostics.push(nodeDiagnostic(requestedFrom, "GOJR_BUILD001", `could not load package ${importPath}: ${message}`));
        return undefined;
    }
}
function loadStubOrSourcePackageFromProvider(provider, importPath, requestedFrom, diagnostics) {
    return stubSourcePackageFiles(importPath) ?? loadSourcePackageFromProvider(provider, importPath, requestedFrom, diagnostics);
}
function sourcePackageFilename(spec) {
    return spec.files[0]?.filename ?? REPL_FILENAME;
}
function nodeDiagnostic(filename, code, message) {
    return {
        filename,
        code,
        severity: "error",
        message
    };
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
    if (normalized.progress) {
        normalized.onProgress = (event) => {
            const line = formatBuildProgressEvent(event);
            if (line)
                process.stderr.write(`${line}\n`);
        };
    }
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
