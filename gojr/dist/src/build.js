import { REPL_FILENAME } from "./diagnostics.js";
import { parseFrontSourceFiles } from "./front/parser.js";
import { frontFilesToProgramAst } from "./frontToAst.js";
import { isIntrinsicPackageImport } from "./intrinsicPackages.js";
import { stubSourcePackageFiles } from "./stubPackages.js";
import { checkGoJuniorFiles, isGoJuniorSyntheticCheckName, standardTypePackage } from "./typecheck.js";
import { Builtin as GoTypesBuiltin, Const as GoTypesConst, Func as GoTypesFunc, TypeName as GoTypesTypeName, Unsafe as GoTypesUnsafe, Var as GoTypesVar } from "./go/types/index.js";
const ARTIFACT_LAYOUT_VERSION = "gojr-js-v4";
export const GOJR_GOOS = "gojr";
export const GOJR_GOARCH = "js";
const DEFAULT_COMPILER_VERSION = "gojr-dev";
const DEFAULT_BACKEND = "js-source-envelope";
const DEFAULT_HOST_SPEC_VERSION = "host-v0";
const DEFAULT_CAPABILITY_POLICY = "default";
const AR_MAGIC = "!<arch>\n";
const PKGDEF_MEMBER = "__.PKGDEF";
const JAVASCRIPT_MEMBER = "_gojr.js";
const GOJR_EXPORT_MAGIC = "$$gojr iexport v1\n";
export function buildPackage(request, store) {
    return buildPackages(request, store);
}
export function buildPackages(request, store) {
    return new PackageGraphBuilder(request, store).build();
}
export function buildStandardLibraryPackage(request, store) {
    const provider = request.sourcePackageProvider ?? createStandardLibrarySourcePackageProvider(request.standardLibrary);
    const files = provider.load(request.importPath);
    if (!files) {
        return emptyBuildReport([buildDiagnostic(request.importPath, `standard library package ${request.importPath} is not available under ${request.standardLibrary.sourceRoot}`)]);
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
export function createStandardLibrarySourcePackageProvider(options) {
    const sourceRoot = trimTrailingSlash(options.sourceRoot);
    const goos = options.goos ?? GOJR_GOOS;
    const goarch = options.goarch ?? GOJR_GOARCH;
    const tags = buildTagSet(goos, goarch, options.buildTags);
    return {
        load(importPath) {
            const parts = importPathParts(importPath);
            if (!parts)
                throw new Error(`invalid import path: ${importPath}`);
            const dir = joinSlash(sourceRoot, ...parts);
            let entries;
            try {
                entries = options.host.readDir(dir);
            }
            catch {
                return undefined;
            }
            const names = entries
                .map((entry) => typeof entry === "string" ? { name: entry, isFile: true } : entry)
                .filter((entry) => entry.isFile !== false)
                .map((entry) => entry.name)
                .filter((name) => !name.startsWith(".") && !name.startsWith("_") && name.endsWith(".go") && !name.endsWith("_test.go"))
                .sort();
            const files = [];
            for (const name of names) {
                const filename = joinSlash(dir, name);
                const source = options.host.readFile(filename);
                if (goSourceFileMatchesBuildContext(name, source, goos, goarch, tags)) {
                    files.push({ filename, source });
                }
            }
            if (files.length === 0)
                throw new Error(`${dir} contains no Go source files matching GOOS=${goos} GOARCH=${goarch}`);
            return files;
        },
        isStandardLibraryPackage(importPath) {
            return importPathParts(importPath) !== undefined;
        }
    };
}
export function inspectPackageJavaScript(request) {
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
export function collectSourceImportPaths(files) {
    const parsed = parseFrontSourceFiles(files.map(ensureSourceFile));
    return {
        diagnostics: parsed.diagnostics,
        imports: uniqueSorted(parsed.files.flatMap((file) => file.imports.map(importPathFromSpec)))
    };
}
class PackageGraphBuilder {
    request;
    store;
    packageSources = new Map();
    packageInfos = new Map();
    nodes = new Map();
    visiting = new Set();
    failedPackageLoads = new Set();
    standardLibraryPackages = new Set();
    diagnostics = [];
    constructor(request, store) {
        this.request = request;
        this.store = store;
        for (const [importPath, files] of Object.entries(request.packageSources ?? {})) {
            this.packageSources.set(importPath, files.map(ensureSourceFile));
        }
        for (const importPath of request.standardLibraryPackages ?? []) {
            this.standardLibraryPackages.add(importPath);
        }
    }
    build() {
        const root = this.buildOne(this.request.importPath, this.request.files, []);
        if (this.hasErrors() || !root)
            return emptyBuildReport(this.diagnostics);
        const artifacts = [];
        const built = [];
        const skipped = [];
        for (const node of this.orderedNodes(root)) {
            const reportArtifact = this.reportArtifact(node, "built");
            const existing = this.store?.read?.(node.artifactPath);
            const existingArtifact = existing === undefined ? undefined : parseGeneratedArtifactSource(existing);
            if (existingArtifact?.cacheKey === node.cacheKey && existingArtifact.layoutVersion === ARTIFACT_LAYOUT_VERSION) {
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
    buildOne(importPathHint, rawFiles, stack) {
        const files = rawFiles.map(ensureSourceFile);
        const parsed = parseFrontSourceFiles(files);
        this.diagnostics.push(...parsed.diagnostics);
        this.diagnostics.push(...packageDiagnostics(parsed.files, files));
        if (this.hasErrors())
            return undefined;
        const packageName = parsed.files.find((file) => file.name)?.name?.name ?? "main";
        const importPath = importPathHint ?? packageName;
        const existing = this.nodes.get(importPath);
        if (existing)
            return existing;
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
        this.progress({
            action: "checking",
            importPath,
            packageName,
            dependencyCount: dependencies.length,
            fileCount: files.length,
            standardLibrary: this.standardLibraryPackages.has(importPath)
        });
        const sourceDependencies = [];
        for (const dependencyPath of dependencies) {
            if (preferAmbientBuildImport(dependencyPath)) {
                const dependency = this.buildAmbientStandardPackage(dependencyPath);
                if (dependency)
                    sourceDependencies.push(dependency);
                continue;
            }
            const dependencyFiles = this.sourceFilesForImport(dependencyPath, files[0]?.filename ?? REPL_FILENAME);
            if (!dependencyFiles) {
                if (isAmbientBuildImport(dependencyPath)) {
                    const dependency = this.buildAmbientStandardPackage(dependencyPath);
                    if (dependency)
                        sourceDependencies.push(dependency);
                    continue;
                }
                if (!this.failedPackageLoads.has(dependencyPath)) {
                    this.diagnostics.push(buildDiagnostic(files[0]?.filename ?? REPL_FILENAME, `package ${dependencyPath} is not available to gojr build; provide it with packageSources or --pkg`));
                }
                continue;
            }
            const dependency = this.buildOne(dependencyPath, dependencyFiles, [...stack, importPath]);
            if (dependency)
                sourceDependencies.push(dependency);
        }
        if (!this.hasErrors()) {
            const checked = importPath === "unsafe"
                ? { diagnostics: [], pkg: GoTypesUnsafe }
                : checkGoJuniorFiles(parsed.files, parsed.statements, [], {
                    packageName,
                    packagePath: importPath,
                    autoImportFmt: false,
                    importer: {
                        import: (path) => this.packageInfos.get(path)
                    }
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
                const functions = frontFilesToProgramAst(parsed.files, [], []).functions;
                const node = {
                    importPath,
                    packageName,
                    files,
                    sourceHash,
                    cacheKey,
                    dependencies,
                    dependencyCacheKeys,
                    exports,
                    artifactPath,
                    artifactSource: generatedArtifactSource(pkgdef, files, functions)
                };
                this.nodes.set(importPath, node);
                this.packageInfos.set(importPath, checked.pkg);
            }
        }
        this.visiting.delete(importPath);
        return this.nodes.get(importPath);
    }
    buildAmbientStandardPackage(importPath) {
        const existing = this.nodes.get(importPath);
        if (existing)
            return existing;
        const pkg = standardTypePackage(importPath);
        if (!pkg)
            return undefined;
        this.standardLibraryPackages.add(importPath);
        this.packageInfos.set(importPath, pkg);
        const packageName = pkg.Name();
        const goos = this.request.goos ?? GOJR_GOOS;
        const goarch = this.request.goarch ?? GOJR_GOARCH;
        const buildTags = resolvedBuildTags(this.request);
        const dependencies = [];
        this.progress({
            action: "checking",
            importPath,
            packageName,
            dependencyCount: dependencies.length,
            fileCount: 1,
            standardLibrary: true
        });
        const dependencyCacheKeys = [];
        const exports = uniqueExports(packageScopeObjects(pkg));
        const sourceHash = stableHash([
            "ambient-stdlib",
            importPath,
            packageName,
            JSON.stringify(exports)
        ].join("\0"));
        const cacheKey = stableHash([
            ARTIFACT_LAYOUT_VERSION,
            this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
            this.request.backend ?? DEFAULT_BACKEND,
            this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
            this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
            goos,
            goarch,
            buildTags.join("\n"),
            "stdlib",
            importPath,
            sourceHash,
            "",
            ""
        ].join("\0"));
        const artifactPath = artifactPathForImportPath(resolveArtifactRoot(this.request), importPath);
        const syntheticFile = {
            filename: `gojr:stdlib/${importPath}`,
            source: `package ${packageName}\n`
        };
        const pkgdef = packageExportData({
            layoutVersion: ARTIFACT_LAYOUT_VERSION,
            compilerVersion: this.request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
            backend: this.request.backend ?? DEFAULT_BACKEND,
            hostSpecVersion: this.request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
            capabilityPolicy: this.request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
            goos,
            goarch,
            buildTags,
            standardLibrary: true,
            importPath,
            packageName,
            sourceHash,
            cacheKey,
            dependencies,
            dependencyCacheKeys,
            exports,
            sources: [{
                    filename: syntheticFile.filename,
                    hash: stableHash(syntheticFile.source)
                }]
        });
        const node = {
            importPath,
            packageName,
            files: [syntheticFile],
            sourceHash,
            cacheKey,
            dependencies,
            dependencyCacheKeys,
            exports,
            artifactPath,
            artifactSource: generatedArtifactSource(pkgdef, [syntheticFile], [])
        };
        this.nodes.set(importPath, node);
        return node;
    }
    progress(event) {
        try {
            this.request.onProgress?.(event);
        }
        catch {
            // Progress sinks are observability hooks and must not affect compilation.
        }
    }
    sourceFilesForImport(importPath, filename) {
        const explicit = this.packageSources.get(importPath);
        if (explicit)
            return explicit;
        const stub = stubSourcePackageFiles(importPath);
        if (stub) {
            const files = stub.map(ensureSourceFile);
            this.packageSources.set(importPath, files);
            this.standardLibraryPackages.add(importPath);
            return files;
        }
        try {
            const loaded = this.request.sourcePackageProvider?.load(importPath);
            if (!loaded)
                return undefined;
            const files = loaded.map(ensureSourceFile);
            this.packageSources.set(importPath, files);
            if (this.request.sourcePackageProvider?.isStandardLibraryPackage?.(importPath)) {
                this.standardLibraryPackages.add(importPath);
            }
            return files;
        }
        catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.failedPackageLoads.add(importPath);
            this.diagnostics.push(buildDiagnostic(filename, `could not load package ${importPath}: ${message}`));
            return undefined;
        }
    }
    orderedNodes(root) {
        const out = [];
        const seen = new Set();
        const visit = (node) => {
            if (seen.has(node.importPath))
                return;
            seen.add(node.importPath);
            for (const dependencyPath of node.dependencies) {
                const dependency = this.nodes.get(dependencyPath);
                if (dependency)
                    visit(dependency);
            }
            out.push(node);
        };
        visit(root);
        return out;
    }
    reportArtifact(node, action) {
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
    hasErrors() {
        return this.diagnostics.some((diagnostic) => diagnostic.severity === "error");
    }
}
function parseGeneratedArtifactSource(source) {
    const archive = parseGoJuniorPackageArchive(source);
    if (archive) {
        return {
            layoutVersion: archive.pkgdef.layoutVersion,
            cacheKey: archive.pkgdef.cacheKey
        };
    }
    const artifactJSON = exportedConstJSON(source, "gojrPackageArtifact");
    if (!artifactJSON)
        return undefined;
    try {
        const value = JSON.parse(artifactJSON);
        return {
            ...(typeof value.layoutVersion === "string" ? { layoutVersion: value.layoutVersion } : {}),
            ...(typeof value.cacheKey === "string" ? { cacheKey: value.cacheKey } : {})
        };
    }
    catch {
        return undefined;
    }
}
function exportedConstJSON(source, name) {
    const startText = `export const ${name} = `;
    const start = source.indexOf(startText);
    if (start < 0)
        return undefined;
    const jsonStart = start + startText.length;
    const end = source.indexOf(";\n", jsonStart);
    if (end < 0)
        return undefined;
    return source.slice(jsonStart, end);
}
export function artifactPathForImportPath(artifactRoot, importPath) {
    const parts = importPathParts(importPath);
    if (!parts) {
        throw new Error(`invalid import path: ${importPath}`);
    }
    return joinSlash(artifactRoot, ...parts) + ".a";
}
export function resolveArtifactRoot(request) {
    if (request.artifactRoot && request.artifactRoot.trim() !== "")
        return trimTrailingSlash(request.artifactRoot);
    if (request.packageCacheParent && request.packageCacheParent.trim() !== "")
        return joinSlash(request.packageCacheParent, "gojr_js");
    return "~/go/pkg/gojr_js";
}
function resolvedBuildTags(request) {
    return [...buildTagSet(request.goos ?? GOJR_GOOS, request.goarch ?? GOJR_GOARCH, request.buildTags)].sort();
}
function buildTagSet(goos, goarch, extra) {
    const tags = new Set();
    tags.add(goos);
    tags.add(goarch);
    if (isUnixGOOS(goos))
        tags.add("unix");
    for (let minor = 1; minor <= 27; minor += 1) {
        tags.add(`go1.${minor}`);
    }
    for (const tag of extra ?? []) {
        const text = tag.trim();
        if (text !== "")
            tags.add(text);
    }
    return tags;
}
function isUnixGOOS(goos) {
    return new Set([
        "aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos",
        "ios", "linux", "netbsd", "openbsd", "solaris"
    ]).has(goos);
}
export function goSourceFileMatchesBuildContext(filename, source, goos, goarch, tags) {
    return goSourceFileNameMatchesBuildContext(filename, goos, goarch) && goSourceMatchesBuildConstraints(source, tags);
}
export function buildTagSetForContext(goos, goarch, extra) {
    return buildTagSet(goos, goarch, extra);
}
function goSourceMatchesBuildConstraints(source, tags) {
    const lines = leadingCommentAndBlankLines(source);
    const goBuild = lines
        .map((line) => line.match(/^\/\/go:build\s+(.+)$/)?.[1]?.trim())
        .filter((line) => !!line);
    if (goBuild.length > 0) {
        return goBuild.every((expr) => evaluateGoBuildExpression(expr, tags));
    }
    const plusBuild = lines
        .map((line) => line.match(/^\/\/\s*\+build\s+(.+)$/)?.[1]?.trim())
        .filter((line) => !!line);
    return plusBuild.length === 0 || plusBuild.every((line) => evaluatePlusBuildLine(line, tags));
}
function leadingCommentAndBlankLines(source) {
    const out = [];
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
function evaluatePlusBuildLine(line, tags) {
    return line.split(/\s+/)
        .filter((option) => option !== "")
        .some((option) => option.split(",").every((term) => {
        if (term.startsWith("!"))
            return !tags.has(term.slice(1));
        return tags.has(term);
    }));
}
function evaluateGoBuildExpression(expr, tags) {
    const tokens = expr.match(/[A-Za-z0-9_.]+|&&|\|\||!|\(|\)/g) ?? [];
    let index = 0;
    const parseOr = () => {
        let value = parseAnd();
        while (tokens[index] === "||") {
            index += 1;
            value = parseAnd() || value;
        }
        return value;
    };
    const parseAnd = () => {
        let value = parseUnary();
        while (tokens[index] === "&&") {
            index += 1;
            value = parseUnary() && value;
        }
        return value;
    };
    const parseUnary = () => {
        const token = tokens[index];
        if (token === "!") {
            index += 1;
            return !parseUnary();
        }
        if (token === "(") {
            index += 1;
            const value = parseOr();
            if (tokens[index] === ")")
                index += 1;
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
function packageDiagnostics(files, sourceFiles) {
    if (sourceFiles.length === 0) {
        return [buildDiagnostic(REPL_FILENAME, "no source files supplied")];
    }
    if (files.length === 0)
        return [];
    const firstName = files.find((file) => file.name)?.name;
    if (!firstName) {
        return [buildDiagnostic(files[0]?.span?.filename ?? sourceFiles[0]?.filename ?? REPL_FILENAME, "package source files must declare a package")];
    }
    const diagnostics = [];
    for (const file of files) {
        if (!file.name) {
            diagnostics.push(buildDiagnostic(file.span?.filename ?? REPL_FILENAME, "package source files must declare a package", file.span));
            continue;
        }
        if (file.name.name !== firstName.name) {
            diagnostics.push(buildDiagnostic(file.name.span?.filename ?? file.span?.filename ?? REPL_FILENAME, `package ${file.name.name} does not match package ${firstName.name}`, file.name.span));
        }
    }
    return diagnostics;
}
function goSourceFileNameMatchesBuildContext(filename, goos, goarch) {
    const name = filename.split("/").pop()?.replace(/\.go$/, "") ?? filename.replace(/\.go$/, "");
    const parts = name.split("_");
    if (parts.length < 2)
        return true;
    const last = parts[parts.length - 1] ?? "";
    const prev = parts[parts.length - 2] ?? "";
    if (knownGOOS.has(prev) && knownGOARCH.has(last)) {
        return prev === goos && last === goarch;
    }
    if (knownGOOS.has(last))
        return last === goos;
    if (knownGOARCH.has(last))
        return last === goarch;
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
function isAmbientBuildImport(path) {
    return isIntrinsicPackageImport(path);
}
function preferAmbientBuildImport(path) {
    return isIntrinsicPackageImport(path);
}
function buildDiagnostic(filename, message, span) {
    return {
        filename,
        code: "GOJR_BUILD001",
        severity: "error",
        message,
        ...(span ? { span } : {})
    };
}
function ensureSourceFile(file) {
    return {
        filename: file.filename || REPL_FILENAME,
        source: file.source ?? ""
    };
}
function emptyBuildReport(diagnostics) {
    return {
        ok: false,
        diagnostics,
        artifacts: [],
        built: [],
        skipped: []
    };
}
function importPathFromSpec(spec) {
    return unquoteStringLiteral(spec.path.value);
}
function packageScopeObjects(pkg) {
    return pkg.Scope().Names().flatMap((name) => {
        if (isGoJuniorSyntheticCheckName(name) || name === "fmt")
            return [];
        const object = pkg.Scope().Lookup(name);
        if (object === null || object.constructor.name === "PkgName")
            return [];
        if (object.Pkg() !== pkg && !(pkg.Path() === "unsafe" && object instanceof GoTypesBuiltin))
            return [];
        return [object];
    });
}
function uniqueExports(objects) {
    const exports = new Map();
    for (const object of objects) {
        const kind = exportKindForObject(object);
        if (!kind || !object.Exported())
            continue;
        const typ = object.Type();
        const item = {
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
function exportKindForObject(object) {
    if (object instanceof GoTypesConst)
        return "const";
    if (object instanceof GoTypesFunc)
        return "func";
    if (object instanceof GoTypesTypeName)
        return "type";
    if (object instanceof GoTypesVar)
        return "var";
    if (object instanceof GoTypesBuiltin)
        return "func";
    return undefined;
}
function packageExportData(data) {
    const exportIndex = {};
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
function generatedArtifactSource(pkgdef, files, functions) {
    const javascript = generatedArtifactJavaScript(pkgdef, files, functions);
    return writeArArchive([
        { name: PKGDEF_MEMBER, data: GOJR_EXPORT_MAGIC + JSON.stringify(pkgdef, null, 2) + "\n" },
        { name: JAVASCRIPT_MEMBER, data: javascript }
    ]);
}
function generatedArtifactJavaScript(artifact, files, functions) {
    const sources = files.map((file) => ({
        filename: file.filename,
        hash: stableHash(file.source),
        source: file.source
    }));
    const compiledFunctions = generatedCompiledFunctionBodyLines(functions);
    return [
        "// Code generated by gojr build; DO NOT EDIT.",
        "// This member is the generated JavaScript payload inside the Go-junior package archive.",
        "// __.PKGDEF carries export metadata; this module carries executable package input.",
        `export const gojrPackageArtifact = ${JSON.stringify(artifact, null, 2)};`,
        `export const gojrPackageSources = ${JSON.stringify(sources, null, 2)};`,
        "export async function instantiateGoJrPackage(runtime, options = {}) {",
        "  if (!runtime || typeof runtime.evaluatePackageSourceFiles !== \"function\") {",
        "    throw new Error(\"gojr package artifact requires runtime.evaluatePackageSourceFiles\");",
        "  }",
        "  const __gojrPackageResult = await runtime.evaluatePackageSourceFiles(gojrPackageSources, {",
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
        "export default {",
        "  artifact: gojrPackageArtifact,",
        "  sources: gojrPackageSources,",
        "  compiledFunctions: gojrCompiledFunctionBodies,",
        "  instantiateGoJrPackage",
        "};",
        ""
    ].join("\n");
}
function generatedCompiledFunctionBodyLines(functions) {
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
function compiledFunctionDescriptors(functions) {
    return functions.map((declaration, index) => {
        const receiver = declaration.receiver?.type.text;
        const name = declaration.name;
        const baseKey = receiver ? `${receiver}.${name}` : name;
        const key = name === "init" || functions.some((other, otherIndex) => otherIndex !== index &&
            other.name === name &&
            (other.receiver?.type.text ?? "") === (receiver ?? ""))
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
function compiledFunctionEntryLines(descriptor, last) {
    const jsName = compiledFunctionJavaScriptName(descriptor);
    return [
        `  ${JSON.stringify(descriptor.key)}: async function ${jsName}(__gojrPackageResult, ...__gojrArgs) {`,
        `    return await __gojrCallCompiledFunction(__gojrPackageResult, ${JSON.stringify(descriptor.name)}, __gojrArgs);`,
        `  }${last ? "" : ","}`
    ];
}
function compiledFunctionJavaScriptName(descriptor) {
    const text = `gojr$${descriptor.key}`;
    return text.replace(/[^A-Za-z0-9_$]/g, "_").replace(/^[^A-Za-z_$]/, "_");
}
export function parseGoJuniorPackageArchive(source) {
    const members = readArArchive(source);
    if (!members)
        return undefined;
    const pkgdefMember = members[0];
    if (pkgdefMember?.name !== PKGDEF_MEMBER || !pkgdefMember.data.startsWith(GOJR_EXPORT_MAGIC))
        return undefined;
    const javascript = members.find((member) => member.name === JAVASCRIPT_MEMBER)?.data ?? "";
    try {
        const pkgdef = JSON.parse(pkgdefMember.data.slice(GOJR_EXPORT_MAGIC.length));
        if (pkgdef.exportFormat !== "gojr-iexport" || pkgdef.exportVersion !== 1)
            return undefined;
        return {
            pkgdef,
            javascript,
            members
        };
    }
    catch {
        return undefined;
    }
}
function writeArArchive(members) {
    return AR_MAGIC + members.map((member) => writeArMember(member.name, member.data)).join("");
}
function writeArMember(name, data) {
    const arName = name;
    if (arName.length > 16)
        throw new Error(`ar member name is too long: ${name}`);
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
function readArArchive(source) {
    if (!source.startsWith(AR_MAGIC))
        return undefined;
    const bytes = utf8Encode(source);
    const members = [];
    let offset = utf8ByteLength(AR_MAGIC);
    while (offset < bytes.length) {
        if (offset + 60 > bytes.length)
            return undefined;
        const header = asciiDecode(bytes.subarray(offset, offset + 60));
        if (header.slice(58, 60) !== "`\n")
            return undefined;
        const rawName = header.slice(0, 16).trim();
        const sizeText = header.slice(48, 58).trim();
        const size = Number.parseInt(sizeText, 10);
        if (!Number.isFinite(size) || size < 0)
            return undefined;
        const dataStart = offset + 60;
        const dataEnd = dataStart + size;
        if (dataEnd > bytes.length)
            return undefined;
        members.push({
            name: rawName.endsWith("/") ? rawName.slice(0, -1) : rawName,
            data: utf8Decode(bytes.subarray(dataStart, dataEnd))
        });
        offset = dataEnd + (size % 2);
    }
    return members;
}
function utf8ByteLength(text) {
    return utf8Encode(text).length;
}
function utf8Encode(text) {
    return new TextEncoder().encode(text);
}
function utf8Decode(bytes) {
    return new TextDecoder("utf-8").decode(bytes);
}
function asciiDecode(bytes) {
    let out = "";
    for (const byte of bytes)
        out += String.fromCharCode(byte);
    return out;
}
function hashSourceFiles(files) {
    return stableHash(files
        .map((file) => `${file.filename}\0${file.source.length}\0${file.source}`)
        .join("\0"));
}
function stableHash(text) {
    let hash = 0xcbf29ce484222325n;
    const prime = 0x100000001b3n;
    const mask = 0xffffffffffffffffn;
    for (let index = 0; index < text.length; index++) {
        hash ^= BigInt(text.charCodeAt(index));
        hash = (hash * prime) & mask;
    }
    return hash.toString(16).padStart(16, "0");
}
function unquoteStringLiteral(value) {
    if (value.length < 2)
        return value;
    const quote = value[0];
    if ((quote !== `"` && quote !== "`") || value[value.length - 1] !== quote)
        return value;
    if (quote === "`")
        return value.slice(1, -1);
    try {
        return JSON.parse(value);
    }
    catch {
        return value.slice(1, -1);
    }
}
function uniqueSorted(values) {
    return [...new Set(values)].sort();
}
function importPathParts(importPath) {
    const text = importPath.trim();
    if (text === "" || text.startsWith("/") || text.includes("\\") || text.includes("//"))
        return undefined;
    const parts = text.split("/");
    if (parts.length === 0 || parts.some((part) => part === "" || part === "." || part === ".."))
        return undefined;
    return parts;
}
function joinSlash(first, ...rest) {
    return [first, ...rest]
        .filter((part) => part !== "")
        .map((part, index) => index === 0 ? trimTrailingSlash(part) : trimSlashes(part))
        .join("/");
}
function trimSlashes(text) {
    return text.replace(/^\/+|\/+$/g, "");
}
function trimTrailingSlash(text) {
    return text.replace(/\/+$/g, "") || "/";
}
