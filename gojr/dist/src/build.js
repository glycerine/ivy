import { REPL_FILENAME } from "./diagnostics.js";
import { checkFrontFiles } from "./front/checker.js";
import { parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
import { newUniverse } from "./front/types.js";
const ARTIFACT_LAYOUT_VERSION = "gojr-js-v1";
const DEFAULT_COMPILER_VERSION = "gojr-dev";
const DEFAULT_BACKEND = "js-source-envelope";
const DEFAULT_HOST_SPEC_VERSION = "host-v0";
const DEFAULT_CAPABILITY_POLICY = "default";
export function buildPackage(request, store) {
    return buildPackages(request, store);
}
export function buildPackages(request, store) {
    const files = request.files.map(ensureSourceFile);
    const parsed = parseFrontSourceFiles(files);
    const diagnostics = [
        ...parsed.diagnostics,
        ...packageDiagnostics(parsed.files, request)
    ];
    if (diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
        return emptyBuildReport(diagnostics);
    }
    const packageName = parsed.files.find((file) => file.name)?.name?.name ?? "main";
    const importPath = request.importPath ?? packageName;
    const checked = checkFrontFiles(parsed.files, {
        universe: newUniverse(),
        packageName,
        packagePath: importPath
    });
    diagnostics.push(...checked.diagnostics);
    if (diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
        return emptyBuildReport(diagnostics);
    }
    const sourceHash = hashSourceFiles(files);
    const dependencies = uniqueSorted(parsed.files.flatMap((file) => file.imports.map(importPathFromSpec)));
    const exports = uniqueExports(parsed.files);
    const cacheKey = stableHash([
        ARTIFACT_LAYOUT_VERSION,
        request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
        request.backend ?? DEFAULT_BACKEND,
        request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
        request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
        importPath,
        sourceHash,
        dependencies.join("\n")
    ].join("\0"));
    const artifactPath = artifactPathForImportPath(resolveArtifactRoot(request), importPath);
    const artifact = {
        layoutVersion: ARTIFACT_LAYOUT_VERSION,
        compilerVersion: request.compilerVersion ?? DEFAULT_COMPILER_VERSION,
        backend: request.backend ?? DEFAULT_BACKEND,
        hostSpecVersion: request.hostSpecVersion ?? DEFAULT_HOST_SPEC_VERSION,
        capabilityPolicy: request.capabilityPolicy ?? DEFAULT_CAPABILITY_POLICY,
        importPath,
        packageName,
        sourceHash,
        cacheKey,
        dependencies,
        exports,
        sources: files.map((file) => ({
            filename: file.filename,
            hash: stableHash(file.source)
        }))
    };
    const source = generatedArtifactSource(artifact);
    store?.writeAtomic(artifactPath, source);
    const reportArtifact = {
        importPath,
        packageName,
        artifactPath,
        action: "built",
        sourceHash,
        cacheKey,
        dependencies,
        exports
    };
    return {
        ok: true,
        diagnostics,
        artifacts: [reportArtifact],
        built: [artifactPath],
        skipped: []
    };
}
export function artifactPathForImportPath(artifactRoot, importPath) {
    const parts = importPath.split("/").filter(Boolean);
    if (parts.length === 0 || parts.some((part) => part === "." || part === "..")) {
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
function packageDiagnostics(files, request) {
    if (request.files.length === 0) {
        return [buildDiagnostic(REPL_FILENAME, "no source files supplied")];
    }
    if (files.length === 0)
        return [];
    const firstName = files.find((file) => file.name)?.name;
    if (!firstName) {
        return [buildDiagnostic(files[0]?.span?.filename ?? request.files[0]?.filename ?? REPL_FILENAME, "package source files must declare a package")];
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
function uniqueExports(files) {
    const exports = new Map();
    for (const file of files) {
        for (const declaration of file.declarations) {
            if (declaration.kind === "FuncDecl") {
                addExport(exports, declaration.name.name, "func", declaration);
            }
            else if (declaration.kind === "GenDecl") {
                collectGenDeclExports(exports, declaration);
            }
        }
    }
    return [...exports.values()].sort((left, right) => left.name.localeCompare(right.name) || left.kind.localeCompare(right.kind));
}
function collectGenDeclExports(exports, declaration) {
    const kind = exportKindForToken(declaration.token);
    if (!kind)
        return;
    for (const spec of declaration.specs) {
        for (const name of namesForSpec(spec))
            addExport(exports, name, kind);
    }
}
function exportKindForToken(token) {
    switch (token) {
        case TokenKind.Const: return "const";
        case TokenKind.Func: return "func";
        case TokenKind.Type: return "type";
        case TokenKind.Var: return "var";
        default: return undefined;
    }
}
function namesForSpec(spec) {
    if (spec.kind === "TypeSpec")
        return [spec.name.name];
    if (spec.kind === "ValueSpec")
        return spec.names.map((name) => name.name);
    return [];
}
function addExport(exports, name, kind, declaration) {
    void declaration;
    if (!isExportedName(name))
        return;
    exports.set(`${kind}:${name}`, { name, kind });
}
function isExportedName(name) {
    const first = name.codePointAt(0);
    return first !== undefined && first >= 0x41 && first <= 0x5a;
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
