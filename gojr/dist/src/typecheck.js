import { REPL_FILENAME, diagnosticFilename } from "./diagnostics.js";
import { parseFrontSource, parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
import { Config, Info, Int, Bool, NewChecker, NewFunc, NewNamed, NewPackage, NewPointer, NewPkgName, NewSignatureType, NewSlice, NewStruct, NewTypeName, NewTuple, NewVar, NoPos, String as GoString, Typ, emptyInterface, init as initGoTypesUniverse, Unsafe } from "./go/types/index.js";
export function checkGoJuniorSource(source, filename, config = {}) {
    const parsed = parseFrontSource(source, filename);
    const result = checkGoJuniorFiles(parsed.file ? [parsed.file] : [], parsed.statements, parsed.diagnostics, config);
    return {
        ...result,
        ...(parsed.file ? { file: parsed.file } : {})
    };
}
export function checkGoJuniorSourceFiles(sourceFiles, config = {}) {
    const parsed = parseFrontSourceFiles(sourceFiles);
    return checkGoJuniorFiles(parsed.files, parsed.statements, parsed.diagnostics, config);
}
export function checkGoJuniorFiles(files, statements = [], parserDiagnostics = [], config = {}) {
    initGoTypesUniverse();
    const packageName = config.packageName ?? files.find((file) => file.name)?.name?.name ?? "main";
    const packagePath = config.packagePath ?? packageName;
    const diagnostics = [...parserDiagnostics];
    const fset = new FrontFileSet(files);
    const info = new Info();
    info.Types = new Map();
    info.Defs = new Map();
    info.Uses = new Map();
    info.Scopes = new Map();
    info.Selections = new Map();
    const checkFiles = filesForChecking(files, statements, packageName);
    const pkg = NewPackage(packagePath, packageName);
    seedPackageScope(pkg, config);
    const conf = new Config();
    conf.Importer = {
        Import(path) {
            const imported = config.importer?.import(path) ?? standardPackage(path);
            if (imported === undefined)
                return [null, new Error(`package ${path} is not available`)];
            return [imported, null];
        }
    };
    conf.DisableUnusedImportCheck = config.disableUnusedImportCheck ?? true;
    conf.GoJuniorSheetNamespaces = config.sheetNamespaces ?? null;
    conf.Error = (err) => diagnostics.push(diagnosticFromGoTypesError(err, fset));
    try {
        const err = NewChecker(conf, fset, pkg, info).Files(checkFiles);
        if (err !== null && err !== undefined && diagnostics.length === parserDiagnostics.length) {
            diagnostics.push(diagnosticFromGoTypesError(err, fset));
        }
    }
    catch (error) {
        diagnostics.push({
            filename: diagnosticFilename(undefined, firstFilename(files)),
            code: "GOJR_TYPE001",
            severity: "error",
            message: error instanceof Error ? error.message : String(error),
            ...(error instanceof Error && error.stack ? { stack: error.stack } : {})
        });
    }
    filterTopLevelExpressionUnusedDiagnostics(diagnostics, statements);
    return {
        pkg,
        info,
        diagnostics,
        files,
        statements
    };
}
function filterTopLevelExpressionUnusedDiagnostics(diagnostics, statements) {
    const expressionOffsets = new Set(statements
        .filter((statement) => statement.kind === "ExprStmt" && statement.span !== undefined)
        .map((statement) => statement.span.offset));
    if (expressionOffsets.size === 0)
        return;
    const kept = diagnostics.filter((diagnostic) => !(diagnostic.code === "GOJR_TYPE001" &&
        / is not used$/.test(diagnostic.message) &&
        diagnostic.span !== undefined &&
        expressionOffsets.has(diagnostic.span.offset)));
    diagnostics.splice(0, diagnostics.length, ...kept);
}
class FrontFileSet {
    files;
    constructor(files) {
        this.files = files;
    }
    Position(pos) {
        const span = this.span(pos);
        return `${span.filename}:${span.line}:${span.column}`;
    }
    span(pos) {
        const offset = positionOffset(pos);
        const file = fileForOffset(this.files, offset);
        return spanFromOffset(file?.span?.filename ?? firstFilename(this.files), offset);
    }
}
function filesForChecking(files, statements, packageName) {
    const hasPackageClause = files.some((file) => file.name);
    if (statements.length === 0) {
        if (hasPackageClause)
            return files;
        return [syntheticPackageFile(files, statements, packageName, true)];
    }
    const promoted = promoteTopLevelShortDeclarations(statements);
    const synthetic = syntheticPackageFile(files, promoted.statements, packageName, !hasPackageClause, promoted.declarations);
    return hasPackageClause ? [...files, synthetic] : [synthetic];
}
function syntheticPackageFile(files, statements, packageName, includeDeclarations, extraDeclarations = []) {
    const first = files[0];
    const name = first?.name ?? ident(packageName, first?.span);
    const start = statements[0]?.span ?? first?.span;
    const declarations = includeDeclarations ? files.flatMap((file) => file.declarations) : importDeclarations(files);
    declarations.push(...extraDeclarations);
    if (statements.length > 0) {
        declarations.push(syntheticFunctionDecl(statements, start));
    }
    return {
        kind: "File",
        name,
        declarations,
        imports: declarations.flatMap((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Import
            ? decl.specs.filter((spec) => spec.kind === "ImportSpec")
            : []),
        unresolved: [],
        comments: [],
        ...(start ? { span: start } : {})
    };
}
function promoteTopLevelShortDeclarations(statements) {
    const declarations = [];
    const remaining = [];
    for (const statement of statements) {
        if (statement.kind === "AssignStmt" && statement.token === TokenKind.Define && allIdentifiers(statement.lhs)) {
            declarations.push(shortDeclarationAsVar(statement, statement.lhs));
            continue;
        }
        remaining.push(statement);
    }
    return { declarations, statements: remaining };
}
function allIdentifiers(exprs) {
    return exprs.every((expr) => expr.kind === "Ident");
}
function shortDeclarationAsVar(statement, names) {
    return {
        kind: "GenDecl",
        token: TokenKind.Var,
        grouped: false,
        specs: [{
                kind: "ValueSpec",
                names,
                values: statement.rhs,
                ...(statement.span ? { span: statement.span } : {})
            }],
        ...(statement.span ? { span: statement.span } : {})
    };
}
function syntheticFunctionDecl(statements, span) {
    const resultCount = maxReturnValueCount(statements);
    const type = {
        kind: "FuncType",
        params: { kind: "FieldList", fields: [], ...(span ? { span } : {}) },
        ...(resultCount > 0 ? { results: syntheticAnyResults(resultCount, span) } : {}),
        ...(span ? { span } : {})
    };
    return {
        kind: "FuncDecl",
        name: ident("__gojr_check_statements", span),
        type,
        body: {
            kind: "BlockStmt",
            statements,
            ...(span ? { span } : {})
        },
        ...(span ? { span } : {})
    };
}
function syntheticAnyResults(count, span) {
    return {
        kind: "FieldList",
        fields: globalThis.Array.from({ length: count }, () => ({
            kind: "Field",
            names: [],
            type: ident("any", span),
            ...(span ? { span } : {})
        })),
        ...(span ? { span } : {})
    };
}
function maxReturnValueCount(statements) {
    let max = 0;
    for (const statement of statements) {
        max = Math.max(max, returnValueCount(statement));
    }
    return max;
}
function returnValueCount(statement) {
    switch (statement.kind) {
        case "ReturnStmt":
            return statement.results.length;
        case "BlockStmt":
            return maxReturnValueCount(statement.statements);
        case "LabeledStmt":
            return returnValueCount(statement.stmt);
        case "IfStmt":
            return Math.max(statement.body ? returnValueCount(statement.body) : 0, statement.else ? returnValueCount(statement.else) : 0);
        case "SwitchStmt":
        case "TypeSwitchStmt":
            return Math.max(0, ...statement.body.map((clause) => maxReturnValueCount(clause.body)));
        case "SelectStmt":
            return Math.max(0, ...statement.body.map((clause) => maxReturnValueCount(clause.body)));
        case "ForStmt":
        case "RangeStmt":
            return returnValueCount(statement.body);
        default:
            return 0;
    }
}
function importDeclarations(files) {
    return files.flatMap((file) => file.declarations.filter((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Import));
}
function seedPackageScope(pkg, config) {
    if (config.autoImportFmt ?? true) {
        pkg.Scope().Insert(NewPkgName(NoPos, pkg, "fmt", fmtPackage()));
    }
    for (const object of config.predeclaredPackageObjects ?? []) {
        if (object.Pkg() !== pkg)
            continue;
        if (pkg.Scope().Lookup(object.Name()) === null)
            pkg.Scope().Insert(object);
    }
}
function standardPackage(path) {
    if (path === "fmt")
        return fmtPackage();
    if (path === "testing")
        return testingPackage();
    if (path === "unsafe")
        return Unsafe;
    return undefined;
}
function fmtPackage() {
    const pkg = NewPackage("fmt", "fmt");
    if (pkg.Scope().Lookup("Printf") !== null)
        return pkg;
    const stringType = Typ[GoString];
    const intType = Typ[Int];
    const argsType = NewSlice(emptyInterface);
    const printfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType)), true);
    const sprintfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Printf", printfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprintf", sprintfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Println", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), null, true)));
    pkg.MarkComplete();
    return pkg;
}
function testingPackage() {
    const pkg = NewPackage("testing", "testing");
    if (pkg.Scope().Lookup("T") !== null)
        return pkg;
    const boolType = Typ[Bool];
    const stringType = Typ[GoString];
    const argsType = NewSlice(emptyInterface);
    const tName = NewTypeName(NoPos, pkg, "T", null);
    const tType = NewNamed(tName, NewStruct([], null), null);
    const tPtr = NewPointer(tType);
    const recv = NewVar(NoPos, pkg, "t", tPtr);
    pkg.Scope().Insert(tName);
    const addMethod = (name, result, parameters = [], variadic = false) => {
        const params = NewTuple(...parameters);
        const results = result === null ? null : NewTuple(NewVar(NoPos, pkg, "", result));
        tType.AddMethod(NewFunc(NoPos, pkg, name, NewSignatureType(recv, null, null, params, results, variadic)));
    };
    addMethod("Fail", null);
    addMethod("FailNow", null);
    addMethod("Failed", boolType);
    addMethod("Fatal", null, [NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Fatalf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Error", null, [NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Errorf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Log", null, [NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Logf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Name", stringType);
    addMethod("Helper", null);
    addMethod("Skip", null, [NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("Skipf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
    addMethod("SkipNow", null);
    addMethod("Skipped", boolType);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Short", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Verbose", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function diagnosticFromGoTypesError(error, fset) {
    const err = error;
    const span = fset.span(err.go116start ?? err.Pos);
    return {
        filename: diagnosticFilename(span),
        code: "GOJR_TYPE001",
        severity: "error",
        message: err.Msg ?? err.message ?? err.Error?.() ?? String(error),
        span
    };
}
function ident(name, span) {
    return {
        kind: "Ident",
        name,
        ...(span ? { span } : {})
    };
}
function positionOffset(pos) {
    const n = typeof pos === "number" ? pos : Number(pos);
    return Number.isFinite(n) && n > 0 ? n - 1 : 0;
}
function fileForOffset(files, offset) {
    for (const file of files) {
        const start = file.span?.offset ?? 0;
        const end = start + (file.span?.length ?? 0);
        if (offset >= start && (file.span === undefined || offset <= end))
            return file;
    }
    return files[0];
}
function firstFilename(files) {
    return files[0]?.span?.filename ?? REPL_FILENAME;
}
function spanFromOffset(filename, offset) {
    return {
        filename,
        offset,
        length: 1,
        line: 1,
        column: offset + 1
    };
}
