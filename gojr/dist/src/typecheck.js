import { REPL_FILENAME, diagnosticFilename } from "./diagnostics.js";
import { parseFrontSource, parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
import { Config, Info, Int, Int8, Int16, Int32, Int64, UntypedInt, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr, Float32, Float64, Bool, NewArray, NewChecker, NewConst, NewField, NewFunc, NewInterfaceType, NewNamed, NewPackage, NewPointer, NewPkgName, NewSignatureType, NewSlice, NewStruct, NewTerm, NewTypeName, NewTypeParam, NewTuple, NewUnion, NewVar, NoPos, String as GoString, Typ, emptyInterface, ensureUniverseInitialized, UniverseLookup, Unsafe } from "./go/types/index.js";
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
    ensureUniverseInitialized();
    if (path === "cmp")
        return cmpPackage();
    if (path === "fmt")
        return fmtPackage();
    if (path === "internal/reflectlite")
        return reflectlitePackage();
    if (path === "iter")
        return iterPackage();
    if (path === "math")
        return mathPackage();
    if (path === "runtime")
        return runtimePackage();
    if (path === "syscall/js")
        return syscallJSPackage();
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
function iterPackage() {
    const pkg = NewPackage("iter", "iter");
    if (pkg.Scope().Lookup("Pull") !== null)
        return pkg;
    const boolType = Typ[Bool];
    const seqV = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const seqYield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "value", seqV)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const seqName = NewTypeName(NoPos, pkg, "Seq", null);
    const seqType = NewNamed(seqName, NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", seqYield)), null, false), null);
    seqType.SetTypeParams([seqV]);
    seqName.setType(seqType);
    pkg.Scope().Insert(seqName);
    const seq2K = NewTypeParam(NewTypeName(NoPos, pkg, "K", null), emptyInterface);
    const seq2V = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const seq2Yield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "key", seq2K), NewVar(NoPos, pkg, "value", seq2V)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const seq2Name = NewTypeName(NoPos, pkg, "Seq2", null);
    const seq2Type = NewNamed(seq2Name, NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", seq2Yield)), null, false), null);
    seq2Type.SetTypeParams([seq2K, seq2V]);
    seq2Name.setType(seq2Type);
    pkg.Scope().Insert(seq2Name);
    const pullV = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const pullYield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "value", pullV)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const pullSeq = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", pullYield)), null, false);
    const pullNext = NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "value", pullV), NewVar(NoPos, pkg, "ok", boolType)), false);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Pull", NewSignatureType(null, null, [pullV], NewTuple(NewVar(NoPos, pkg, "seq", pullSeq)), NewTuple(NewVar(NoPos, pkg, "next", pullNext), NewVar(NoPos, pkg, "stop", NewSignatureType(null, null, null, null, null, false))), false)));
    const pull2K = NewTypeParam(NewTypeName(NoPos, pkg, "K", null), emptyInterface);
    const pull2V = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const pull2Yield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "key", pull2K), NewVar(NoPos, pkg, "value", pull2V)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const pull2Seq = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", pull2Yield)), null, false);
    const pull2Next = NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "key", pull2K), NewVar(NoPos, pkg, "value", pull2V), NewVar(NoPos, pkg, "ok", boolType)), false);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Pull2", NewSignatureType(null, null, [pull2K, pull2V], NewTuple(NewVar(NoPos, pkg, "seq", pull2Seq)), NewTuple(NewVar(NoPos, pkg, "next", pull2Next), NewVar(NoPos, pkg, "stop", NewSignatureType(null, null, null, null, null, false))), false)));
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
    const untypedIntType = Typ[UntypedInt];
    const boolType = Typ[Bool];
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt", untypedIntType, 9223372036854775807n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt", untypedIntType, -9223372036854775808n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt8", untypedIntType, 127n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt8", untypedIntType, -128n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt16", untypedIntType, 32767n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt16", untypedIntType, -32768n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt32", untypedIntType, 2147483647n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt32", untypedIntType, -2147483648n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt64", untypedIntType, 9223372036854775807n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt64", untypedIntType, -9223372036854775808n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint", untypedIntType, 18446744073709551615n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint8", untypedIntType, 255n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint16", untypedIntType, 65535n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint32", untypedIntType, 4294967295n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint64", untypedIntType, 18446744073709551615n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxFloat32", float64Type, 3.4028234663852886e38));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxFloat64", float64Type, Number.MAX_VALUE));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NaN", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Inf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "sign", intType)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Exp", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", float64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Floor", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", float64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "IsNaN", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", float64Type)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Log", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", float64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
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
    const byteSliceType = NewSlice(Typ[Uint8]);
    const argsType = NewSlice(emptyInterface);
    const errorObject = UniverseLookup("error");
    const errorType = errorObject?.Type?.() ?? null;
    if (errorType === null) {
        throw new Error("go/types: predeclared error is not initialized");
    }
    const writerType = NewInterfaceType([
        NewFunc(NoPos, pkg, "Write", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", byteSliceType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), false))
    ], null).Complete();
    const printfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    const sprintfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true);
    const fprintfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "w", writerType), NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    const printSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    const fprintSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "w", writerType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Fprint", fprintSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Fprintf", fprintfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Fprintln", fprintSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Print", printSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Printf", printfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprintf", sprintfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Errorf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", errorType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprint", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprintln", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Append", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", byteSliceType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", byteSliceType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Appendf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", byteSliceType), NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", byteSliceType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Appendln", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", byteSliceType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", byteSliceType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Println", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true)));
    pkg.MarkComplete();
    return pkg;
}
function reflectlitePackage() {
    const pkg = NewPackage("internal/reflectlite", "reflectlite");
    if (pkg.Scope().Lookup("TypeOf") !== null)
        return pkg;
    const boolType = Typ[Bool];
    const intType = Typ[Int];
    const anyType = emptyInterface;
    const kindName = NewTypeName(NoPos, pkg, "Kind", null);
    const kindType = NewNamed(kindName, intType, null);
    kindName.setType(kindType);
    pkg.Scope().Insert(kindName);
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Invalid", kindType, 0n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Interface", kindType, 20n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Ptr", kindType, 22n));
    const typeName = NewTypeName(NoPos, pkg, "Type", null);
    const typeType = NewNamed(typeName, NewInterfaceType(null, null).Complete(), null);
    typeName.setType(typeType);
    pkg.Scope().Insert(typeName);
    const typeMethods = [
        NewFunc(NoPos, pkg, "AssignableTo", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "u", typeType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)),
        NewFunc(NoPos, pkg, "Comparable", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)),
        NewFunc(NoPos, pkg, "Elem", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", typeType)), false)),
        NewFunc(NoPos, pkg, "Implements", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "u", typeType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)),
        NewFunc(NoPos, pkg, "Kind", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", kindType)), false)),
        NewFunc(NoPos, pkg, "String", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", Typ[GoString])), false))
    ];
    typeType.SetUnderlying(NewInterfaceType(typeMethods, null).Complete());
    const valueName = NewTypeName(NoPos, pkg, "Value", null);
    const valueType = NewNamed(valueName, NewStruct([], null), null);
    valueName.setType(valueType);
    pkg.Scope().Insert(valueName);
    const valueRecv = NewVar(NoPos, pkg, "v", valueType);
    valueType.AddMethod(NewFunc(NoPos, pkg, "Elem", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "IsNil", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Kind", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", kindType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Len", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Set", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "x", valueType)), null, false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Type", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", typeType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "TypeOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "i", anyType)), NewTuple(NewVar(NoPos, pkg, "", typeType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Swapper", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "slice", anyType)), NewTuple(NewVar(NoPos, pkg, "", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "i", intType), NewVar(NoPos, pkg, "j", intType)), null, false))), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ValueOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "i", anyType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function syscallJSPackage() {
    const pkg = NewPackage("syscall/js", "js");
    if (pkg.Scope().Lookup("Value") !== null)
        return pkg;
    const anyType = emptyInterface;
    const boolType = Typ[Bool];
    const float64Type = Typ[Float64];
    const intType = Typ[Int];
    const stringType = Typ[GoString];
    const byteSliceType = NewSlice(Typ[Uint8]);
    const typeName = NewTypeName(NoPos, pkg, "Type", null);
    const typeType = NewNamed(typeName, intType, null);
    typeName.setType(typeType);
    pkg.Scope().Insert(typeName);
    typeType.AddMethod(NewFunc(NoPos, pkg, "String", NewSignatureType(NewVar(NoPos, pkg, "t", typeType), null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    [
        "TypeUndefined",
        "TypeNull",
        "TypeBoolean",
        "TypeNumber",
        "TypeString",
        "TypeSymbol",
        "TypeObject",
        "TypeFunction"
    ].forEach((name, index) => {
        pkg.Scope().Insert(NewConst(NoPos, pkg, name, typeType, BigInt(index)));
    });
    const valueName = NewTypeName(NoPos, pkg, "Value", null);
    const valueType = NewNamed(valueName, NewStruct([], null), null);
    valueName.setType(valueType);
    pkg.Scope().Insert(valueName);
    const valueRecv = NewVar(NoPos, pkg, "v", valueType);
    const valueMethods = [
        ["Bool", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Call", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "m", stringType), NewVar(NoPos, pkg, "args", NewSlice(anyType))), NewTuple(NewVar(NoPos, pkg, "", valueType)), true)],
        ["Delete", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "p", stringType)), null, false)],
        ["Equal", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "w", valueType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Float", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)],
        ["Get", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "p", stringType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)],
        ["Index", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "i", intType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)],
        ["InstanceOf", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "t", valueType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Int", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)],
        ["Invoke", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "args", NewSlice(anyType))), NewTuple(NewVar(NoPos, pkg, "", valueType)), true)],
        ["IsNaN", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["IsNull", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["IsUndefined", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Length", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)],
        ["New", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "args", NewSlice(anyType))), NewTuple(NewVar(NoPos, pkg, "", valueType)), true)],
        ["Set", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "p", stringType), NewVar(NoPos, pkg, "x", anyType)), null, false)],
        ["SetIndex", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "i", intType), NewVar(NoPos, pkg, "x", anyType)), null, false)],
        ["String", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)],
        ["Truthy", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Type", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", typeType)), false)]
    ];
    for (const [name, signature] of valueMethods) {
        valueType.AddMethod(NewFunc(NoPos, pkg, name, signature));
    }
    const funcName = NewTypeName(NoPos, pkg, "Func", null);
    const funcType = NewNamed(funcName, NewStruct([
        NewField(NoPos, pkg, "Value", valueType, true)
    ], null), null);
    funcName.setType(funcType);
    pkg.Scope().Insert(funcName);
    funcType.AddMethod(NewFunc(NoPos, pkg, "Release", NewSignatureType(NewVar(NoPos, pkg, "c", funcType), null, null, null, null, false)));
    const errorName = NewTypeName(NoPos, pkg, "Error", null);
    const errorType = NewNamed(errorName, NewStruct([
        NewField(NoPos, pkg, "Value", valueType, true)
    ], null), null);
    errorName.setType(errorType);
    pkg.Scope().Insert(errorName);
    errorType.AddMethod(NewFunc(NoPos, pkg, "Error", NewSignatureType(NewVar(NoPos, pkg, "e", errorType), null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    const valueErrorName = NewTypeName(NoPos, pkg, "ValueError", null);
    const valueErrorType = NewNamed(valueErrorName, NewStruct([
        NewField(NoPos, pkg, "Method", stringType, false),
        NewField(NoPos, pkg, "Type", typeType, false)
    ], null), null);
    valueErrorName.setType(valueErrorType);
    pkg.Scope().Insert(valueErrorName);
    valueErrorType.AddMethod(NewFunc(NoPos, pkg, "Error", NewSignatureType(NewVar(NoPos, pkg, "e", NewPointer(valueErrorType)), null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    const funcOfParam = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "this", valueType), NewVar(NoPos, pkg, "args", NewSlice(valueType))), NewTuple(NewVar(NoPos, pkg, "", anyType)), false);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "FuncOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "fn", funcOfParam)), NewTuple(NewVar(NoPos, pkg, "", funcType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Global", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Null", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Undefined", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ValueOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", anyType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "CopyBytesToGo", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "dst", byteSliceType), NewVar(NoPos, pkg, "src", valueType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "CopyBytesToJS", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "dst", valueType), NewVar(NoPos, pkg, "src", byteSliceType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function runtimePackage() {
    const pkg = NewPackage("runtime", "runtime");
    if (pkg.Scope().Lookup("GOOS") !== null)
        return pkg;
    const intType = Typ[Int];
    const boolType = Typ[Bool];
    const int64Type = Typ[Int64];
    const stringType = Typ[GoString];
    const uintptrType = Typ[Uintptr];
    const uint32Type = Typ[Uint32];
    const uint64Type = Typ[Uint64];
    const float64Type = Typ[Float64];
    const byteSliceType = NewSlice(Typ[Uint8]);
    const uintptrSliceType = NewSlice(uintptrType);
    pkg.Scope().Insert(NewConst(NoPos, pkg, "GOOS", stringType, "js"));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "GOARCH", stringType, "gojr"));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Compiler", stringType, "gojr"));
    pkg.Scope().Insert(NewVar(NoPos, pkg, "MemProfileRate", intType));
    const funcName = NewTypeName(NoPos, pkg, "Func", null);
    const funcType = NewNamed(funcName, NewStruct([], null), null);
    funcName.setType(funcType);
    pkg.Scope().Insert(funcName);
    const funcRecv = NewVar(NoPos, pkg, "f", NewPointer(funcType));
    funcType.AddMethod(NewFunc(NoPos, pkg, "Entry", NewSignatureType(funcRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", uintptrType)), false)));
    funcType.AddMethod(NewFunc(NoPos, pkg, "FileLine", NewSignatureType(funcRecv, null, null, NewTuple(NewVar(NoPos, pkg, "pc", uintptrType)), NewTuple(NewVar(NoPos, pkg, "file", stringType), NewVar(NoPos, pkg, "line", intType)), false)));
    funcType.AddMethod(NewFunc(NoPos, pkg, "Name", NewSignatureType(funcRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    const frameName = NewTypeName(NoPos, pkg, "Frame", null);
    const frameType = NewNamed(frameName, NewStruct([
        NewField(NoPos, pkg, "PC", uintptrType, false),
        NewField(NoPos, pkg, "Func", NewPointer(funcType), false),
        NewField(NoPos, pkg, "Function", stringType, false),
        NewField(NoPos, pkg, "File", stringType, false),
        NewField(NoPos, pkg, "Line", intType, false),
        NewField(NoPos, pkg, "Entry", uintptrType, false)
    ], null), null);
    frameName.setType(frameType);
    pkg.Scope().Insert(frameName);
    const framesName = NewTypeName(NoPos, pkg, "Frames", null);
    const framesType = NewNamed(framesName, NewStruct([], null), null);
    framesName.setType(framesType);
    pkg.Scope().Insert(framesName);
    const framesRecv = NewVar(NoPos, pkg, "ci", NewPointer(framesType));
    framesType.AddMethod(NewFunc(NoPos, pkg, "Next", NewSignatureType(framesRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "frame", frameType), NewVar(NoPos, pkg, "more", boolType)), false)));
    const stackRecordName = NewTypeName(NoPos, pkg, "StackRecord", null);
    const stackRecordType = NewNamed(stackRecordName, NewStruct([
        NewField(NoPos, pkg, "Stack0", NewArray(uintptrType, 32), false)
    ], null), null);
    stackRecordName.setType(stackRecordType);
    pkg.Scope().Insert(stackRecordName);
    stackRecordType.AddMethod(NewFunc(NoPos, pkg, "Stack", NewSignatureType(NewVar(NoPos, pkg, "r", NewPointer(stackRecordType)), null, null, null, NewTuple(NewVar(NoPos, pkg, "", uintptrSliceType)), false)));
    const memProfileRecordName = NewTypeName(NoPos, pkg, "MemProfileRecord", null);
    const memProfileRecordType = NewNamed(memProfileRecordName, NewStruct([
        NewField(NoPos, pkg, "AllocBytes", int64Type, false),
        NewField(NoPos, pkg, "FreeBytes", int64Type, false),
        NewField(NoPos, pkg, "AllocObjects", int64Type, false),
        NewField(NoPos, pkg, "FreeObjects", int64Type, false),
        NewField(NoPos, pkg, "Stack0", NewArray(uintptrType, 32), false)
    ], null), null);
    memProfileRecordName.setType(memProfileRecordType);
    pkg.Scope().Insert(memProfileRecordName);
    const memProfileRecordRecv = NewVar(NoPos, pkg, "r", NewPointer(memProfileRecordType));
    memProfileRecordType.AddMethod(NewFunc(NoPos, pkg, "InUseBytes", NewSignatureType(memProfileRecordRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", int64Type)), false)));
    memProfileRecordType.AddMethod(NewFunc(NoPos, pkg, "InUseObjects", NewSignatureType(memProfileRecordRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", int64Type)), false)));
    memProfileRecordType.AddMethod(NewFunc(NoPos, pkg, "Stack", NewSignatureType(memProfileRecordRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", uintptrSliceType)), false)));
    const blockProfileRecordName = NewTypeName(NoPos, pkg, "BlockProfileRecord", null);
    const blockProfileRecordType = NewNamed(blockProfileRecordName, NewStruct([
        NewField(NoPos, pkg, "Count", int64Type, false),
        NewField(NoPos, pkg, "Cycles", int64Type, false),
        NewField(NoPos, pkg, "StackRecord", stackRecordType, true)
    ], null), null);
    blockProfileRecordName.setType(blockProfileRecordType);
    pkg.Scope().Insert(blockProfileRecordName);
    const errorObject = UniverseLookup("error");
    const errorType = errorObject?.Type?.() ?? null;
    if (errorType === null) {
        throw new Error("go/types: predeclared error is not initialized");
    }
    const runtimeErrorInterface = NewInterfaceType([
        NewFunc(NoPos, pkg, "RuntimeError", NewSignatureType(null, null, null, null, null, false))
    ], [errorType]).Complete();
    const runtimeErrorName = NewTypeName(NoPos, pkg, "Error", null);
    const runtimeErrorType = NewNamed(runtimeErrorName, runtimeErrorInterface, null);
    runtimeErrorName.setType(runtimeErrorType);
    pkg.Scope().Insert(runtimeErrorName);
    const cleanupName = NewTypeName(NoPos, pkg, "Cleanup", null);
    const cleanupType = NewNamed(cleanupName, NewStruct([], null), null);
    cleanupName.setType(cleanupType);
    pkg.Scope().Insert(cleanupName);
    const cleanupRecv = NewVar(NoPos, pkg, "c", cleanupType);
    cleanupType.AddMethod(NewFunc(NoPos, pkg, "Stop", NewSignatureType(cleanupRecv, null, null, null, null, false)));
    const memStatsName = NewTypeName(NoPos, pkg, "MemStats", null);
    const bySizeStruct = NewStruct([
        NewField(NoPos, pkg, "Size", uint32Type, false),
        NewField(NoPos, pkg, "Mallocs", uint64Type, false),
        NewField(NoPos, pkg, "Frees", uint64Type, false)
    ], null);
    const memStatsType = NewNamed(memStatsName, NewStruct([
        NewField(NoPos, pkg, "Alloc", uint64Type, false),
        NewField(NoPos, pkg, "TotalAlloc", uint64Type, false),
        NewField(NoPos, pkg, "Sys", uint64Type, false),
        NewField(NoPos, pkg, "Lookups", uint64Type, false),
        NewField(NoPos, pkg, "Mallocs", uint64Type, false),
        NewField(NoPos, pkg, "Frees", uint64Type, false),
        NewField(NoPos, pkg, "HeapAlloc", uint64Type, false),
        NewField(NoPos, pkg, "HeapSys", uint64Type, false),
        NewField(NoPos, pkg, "HeapIdle", uint64Type, false),
        NewField(NoPos, pkg, "HeapInuse", uint64Type, false),
        NewField(NoPos, pkg, "HeapReleased", uint64Type, false),
        NewField(NoPos, pkg, "HeapObjects", uint64Type, false),
        NewField(NoPos, pkg, "StackInuse", uint64Type, false),
        NewField(NoPos, pkg, "StackSys", uint64Type, false),
        NewField(NoPos, pkg, "MSpanInuse", uint64Type, false),
        NewField(NoPos, pkg, "MSpanSys", uint64Type, false),
        NewField(NoPos, pkg, "MCacheInuse", uint64Type, false),
        NewField(NoPos, pkg, "MCacheSys", uint64Type, false),
        NewField(NoPos, pkg, "BuckHashSys", uint64Type, false),
        NewField(NoPos, pkg, "GCSys", uint64Type, false),
        NewField(NoPos, pkg, "OtherSys", uint64Type, false),
        NewField(NoPos, pkg, "NextGC", uint64Type, false),
        NewField(NoPos, pkg, "LastGC", uint64Type, false),
        NewField(NoPos, pkg, "PauseTotalNs", uint64Type, false),
        NewField(NoPos, pkg, "PauseNs", NewArray(uint64Type, 256), false),
        NewField(NoPos, pkg, "PauseEnd", NewArray(uint64Type, 256), false),
        NewField(NoPos, pkg, "NumGC", uint32Type, false),
        NewField(NoPos, pkg, "NumForcedGC", uint32Type, false),
        NewField(NoPos, pkg, "GCCPUFraction", float64Type, false),
        NewField(NoPos, pkg, "EnableGC", boolType, false),
        NewField(NoPos, pkg, "DebugGC", boolType, false),
        NewField(NoPos, pkg, "BySize", NewArray(bySizeStruct, 61), false)
    ], null), null);
    memStatsName.setType(memStatsType);
    pkg.Scope().Insert(memStatsName);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Caller", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "skip", intType)), NewTuple(NewVar(NoPos, pkg, "pc", uintptrType), NewVar(NoPos, pkg, "file", stringType), NewVar(NoPos, pkg, "line", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Callers", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "skip", intType), NewVar(NoPos, pkg, "pc", uintptrSliceType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "CallersFrames", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "callers", uintptrSliceType)), NewTuple(NewVar(NoPos, pkg, "", NewPointer(framesType))), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "FuncForPC", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "pc", uintptrType)), NewTuple(NewVar(NoPos, pkg, "", NewPointer(funcType))), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "BlockProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(blockProfileRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    const cleanupT = NewTypeParam(NewTypeName(NoPos, pkg, "T", null), emptyInterface);
    const cleanupS = NewTypeParam(NewTypeName(NoPos, pkg, "S", null), emptyInterface);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "AddCleanup", NewSignatureType(null, null, [cleanupT, cleanupS], NewTuple(NewVar(NoPos, pkg, "ptr", NewPointer(cleanupT)), NewVar(NoPos, pkg, "cleanup", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "", cleanupS)), null, false)), NewVar(NoPos, pkg, "arg", cleanupS)), NewTuple(NewVar(NoPos, pkg, "", cleanupType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GC", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GOMAXPROCS", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "n", intType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GOROOT", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Goexit", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GoroutineProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(stackRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Gosched", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "KeepAlive", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", emptyInterface)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "MemProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(memProfileRecordType)), NewVar(NoPos, pkg, "inuseZero", boolType)), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "MutexProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(blockProfileRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NumCPU", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NumGoroutine", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ReadMemStats", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "m", NewPointer(memStatsType))), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetBlockProfileRate", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "rate", intType)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetCPUProfileRate", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "hz", intType)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetFinalizer", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "obj", emptyInterface), NewVar(NoPos, pkg, "finalizer", emptyInterface)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetMutexProfileFraction", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "rate", intType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ThreadCreateProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(stackRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Stack", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "buf", byteSliceType), NewVar(NoPos, pkg, "all", boolType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Version", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
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
