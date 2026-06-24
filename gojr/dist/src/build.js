import { REPL_FILENAME } from "./diagnostics.js";
import { parseFrontSourceFiles } from "./front/parser.js";
import { checkGoJuniorFiles } from "./typecheck.js";
import { Const as GoTypesConst, Func as GoTypesFunc, TypeName as GoTypesTypeName, Var as GoTypesVar } from "./go/types/index.js";
const ARTIFACT_LAYOUT_VERSION = "gojr-js-v2";
export const GOJR_GOOS = "gojr";
export const GOJR_GOARCH = "js";
const DEFAULT_COMPILER_VERSION = "gojr-dev";
const DEFAULT_BACKEND = "js-source-envelope";
const DEFAULT_HOST_SPEC_VERSION = "host-v0";
const DEFAULT_CAPABILITY_POLICY = "default";
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
                if (goSourceMatchesBuildConstraints(source, tags)) {
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
            source = nextSource;
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
        const sourceDependencies = [];
        for (const dependencyPath of dependencies) {
            if (isStandardBuildImport(dependencyPath) && !this.packageSources.has(dependencyPath))
                continue;
            const dependencyFiles = this.sourceFilesForImport(dependencyPath, files[0]?.filename ?? REPL_FILENAME);
            if (!dependencyFiles) {
                if (!isStandardBuildImport(dependencyPath) && !this.failedPackageLoads.has(dependencyPath)) {
                    this.diagnostics.push(buildDiagnostic(files[0]?.filename ?? REPL_FILENAME, `package ${dependencyPath} is not available to gojr build; provide it with packageSources or --pkg`));
                }
                continue;
            }
            const dependency = this.buildOne(dependencyPath, dependencyFiles, [...stack, importPath]);
            if (dependency)
                sourceDependencies.push(dependency);
        }
        if (!this.hasErrors()) {
            const checked = checkGoJuniorFiles(parsed.files, parsed.statements, [], {
                packageName,
                packagePath: importPath,
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
                const artifact = {
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
                };
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
                    artifactSource: generatedArtifactSource(artifact)
                };
                this.nodes.set(importPath, node);
                this.packageInfos.set(importPath, checked.pkg);
            }
        }
        this.visiting.delete(importPath);
        return this.nodes.get(importPath);
    }
    sourceFilesForImport(importPath, filename) {
        const explicit = this.packageSources.get(importPath);
        if (explicit)
            return explicit;
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
    const match = /export const gojrPackageArtifact = ([\s\S]*);\s*$/.exec(source);
    if (!match?.[1])
        return undefined;
    try {
        const value = JSON.parse(match[1]);
        return {
            ...(typeof value.layoutVersion === "string" ? { layoutVersion: value.layoutVersion } : {}),
            ...(typeof value.cacheKey === "string" ? { cacheKey: value.cacheKey } : {})
        };
    }
    catch {
        return undefined;
    }
}
export function artifactPathForImportPath(artifactRoot, importPath) {
    const parts = importPathParts(importPath);
    if (!parts) {
        throw new Error(`invalid import path: ${importPath}`);
    }
    return joinSlash(artifactRoot, ...parts) + ".js";
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
function isStandardBuildImport(path) {
    return path === "fmt" || path === "testing";
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
        const object = pkg.Scope().Lookup(name);
        return object === null ? [] : [object];
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
    return undefined;
}
function generatedArtifactSource(artifact) {
    return [
        "// Code generated by gojr build; DO NOT EDIT.",
        "// This envelope is the package artifact cache format used before full JS lowering.",
        `export const gojrPackageArtifact = ${JSON.stringify(artifact, null, 2)};`,
        ""
    ].join("\n");
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
