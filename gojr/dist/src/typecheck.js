import { REPL_FILENAME, diagnosticFilename } from "./diagnostics.js";
import { parseFrontSource, parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
import { Config, Info, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr, Float32, Float64, Bool, NewChecker, NewFunc, NewInterfaceType, NewNamed, NewPackage, NewPointer, NewPkgName, NewSignatureType, NewSlice, NewStruct, NewTerm, NewTypeName, NewTypeParam, NewTuple, NewUnion, NewVar, NoPos, String as GoString, Typ, emptyInterface, ensureUniverseInitialized, UniverseLookup, Unsafe } from "./go/types/index.js";
export const GOJR_SYNTHETIC_CHECK_PREFIX = "__gojr_check_statements";
export function isGoJuniorSyntheticCheckName(name) {
    return name === GOJR_SYNTHETIC_CHECK_PREFIX || name.startsWith(`${GOJR_SYNTHETIC_CHECK_PREFIX}_`);
}
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
    ensureUniverseInitialized();
    const packageName = config.packageInstance?.Name() ?? config.packageName ?? files.find((file) => file.name)?.name?.name ?? "main";
    const packagePath = config.packageInstance?.Path() ?? config.packagePath ?? packageName;
    const diagnostics = [...parserDiagnostics];
    const fset = new FrontFileSet(files);
    const info = new Info();
    info.Types = new Map();
    info.Defs = new Map();
    info.Uses = new Map();
    info.Scopes = new Map();
    info.Selections = new Map();
    const checkFiles = filesForChecking(files, statements, packageName, config.syntheticFunctionName);
    const pkg = config.packageInstance ?? NewPackage(packagePath, packageName);
    seedPackageScope(pkg, config);
    const conf = new Config();
    conf.Importer = {
        Import(path) {
            const imported = config.importer?.import(path) ?? standardTypePackage(path);
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
function filesForChecking(files, statements, packageName, syntheticFunctionName = GOJR_SYNTHETIC_CHECK_PREFIX) {
    const hasPackageClause = files.some((file) => file.name);
    if (statements.length === 0) {
        if (hasPackageClause)
            return files;
        return [syntheticPackageFile(files, statements, packageName, true)];
    }
    const promoted = promoteTopLevelShortDeclarations(statements);
    const synthetic = syntheticPackageFile(files, promoted.statements, packageName, !hasPackageClause, promoted.declarations, syntheticFunctionName);
    return hasPackageClause ? [...files, synthetic] : [synthetic];
}
function syntheticPackageFile(files, statements, packageName, includeDeclarations, extraDeclarations = [], syntheticFunctionName = GOJR_SYNTHETIC_CHECK_PREFIX) {
    const first = files[0];
    const name = first?.name ?? ident(packageName, first?.span);
    const start = statements[0]?.span ?? first?.span;
    const declarations = includeDeclarations ? files.flatMap((file) => file.declarations) : importDeclarations(files);
    declarations.push(...extraDeclarations);
    if (statements.length > 0) {
        declarations.push(syntheticFunctionDecl(statements, start, syntheticFunctionName));
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
function syntheticFunctionDecl(statements, span, name = GOJR_SYNTHETIC_CHECK_PREFIX) {
    const resultCount = maxReturnValueCount(statements);
    const type = {
        kind: "FuncType",
        params: { kind: "FieldList", fields: [], ...(span ? { span } : {}) },
        ...(resultCount > 0 ? { results: syntheticAnyResults(resultCount, span) } : {}),
        ...(span ? { span } : {})
    };
    return {
        kind: "FuncDecl",
        name: ident(name, span),
        type,
        body: {
            kind: "BlockStmt",
            statements: resultCount > 0 ? [...statements, syntheticNilReturn(resultCount, span)] : statements,
            ...(span ? { span } : {})
        },
        ...(span ? { span } : {})
    };
}
function syntheticNilReturn(count, span) {
    return {
        kind: "ReturnStmt",
        results: globalThis.Array.from({ length: count }, () => ident("nil", span)),
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
export function standardTypePackage(path) {
    if (path === "cmp")
        return cmpPackage();
    if (path === "fmt")
        return fmtPackage();
    if (path === "math")
        return mathPackage();
    if (path === "testing")
        return testingPackage();
    if (path === "unsafe")
        return Unsafe;
    return undefined;
}
function cmpPackage() {
    const pkg = NewPackage("cmp", "cmp");
    if (pkg.Scope().Lookup("Compare") !== null)
        return pkg;
    const intType = Typ[Int];
    const boolType = Typ[Bool];
    const orderedConstraint = NewInterfaceType(null, [NewUnion([
            NewTerm(true, Typ[Int]),
            NewTerm(true, Typ[Int8]),
            NewTerm(true, Typ[Int16]),
            NewTerm(true, Typ[Int32]),
            NewTerm(true, Typ[Int64]),
            NewTerm(true, Typ[Uint]),
            NewTerm(true, Typ[Uint8]),
            NewTerm(true, Typ[Uint16]),
            NewTerm(true, Typ[Uint32]),
            NewTerm(true, Typ[Uint64]),
            NewTerm(true, Typ[Uintptr]),
            NewTerm(true, Typ[Float32]),
            NewTerm(true, Typ[Float64]),
            NewTerm(true, Typ[GoString])
        ])]).Complete();
    const orderedName = NewTypeName(NoPos, pkg, "Ordered", null);
    const orderedType = NewNamed(orderedName, orderedConstraint, null);
    orderedName.setType(orderedType);
    pkg.Scope().Insert(orderedName);
    const orderedTypeParam = () => NewTypeParam(NewTypeName(NoPos, pkg, "T", null), orderedType);
    const comparableTypeParam = () => {
        const comparable = UniverseLookup("comparable");
        const comparableType = comparable?.Type?.() ?? null;
        if (comparableType === null) {
            throw new Error("go/types: predeclared comparable is not initialized");
        }
        return NewTypeParam(NewTypeName(NoPos, pkg, "T", null), comparableType);
    };
    const compareT = orderedTypeParam();
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Compare", NewSignatureType(null, null, [compareT], NewTuple(NewVar(NoPos, pkg, "x", compareT), NewVar(NoPos, pkg, "y", compareT)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    const lessT = orderedTypeParam();
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Less", NewSignatureType(null, null, [lessT], NewTuple(NewVar(NoPos, pkg, "x", lessT), NewVar(NoPos, pkg, "y", lessT)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    const orT = comparableTypeParam();
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Or", NewSignatureType(null, null, [orT], NewTuple(NewVar(NoPos, pkg, "vals", NewSlice(orT))), NewTuple(NewVar(NoPos, pkg, "", orT)), true)));
    pkg.MarkComplete();
    return pkg;
}
function mathPackage() {
    const pkg = NewPackage("math", "math");
    if (pkg.Scope().Lookup("NaN") !== null)
        return pkg;
    const float64Type = Typ[Float64];
    const uint32Type = Typ[Uint32];
    const uint64Type = Typ[Uint64];
    const intType = Typ[Int];
    const boolType = Typ[Bool];
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NaN", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Inf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "sign", intType)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "IsNaN", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", float64Type)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float32bits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", Typ[Float32])), NewTuple(NewVar(NoPos, pkg, "", uint32Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float32frombits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", uint32Type)), NewTuple(NewVar(NoPos, pkg, "", Typ[Float32])), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float64bits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", float64Type)), NewTuple(NewVar(NoPos, pkg, "", uint64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float64frombits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", uint64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.MarkComplete();
    return pkg;
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
