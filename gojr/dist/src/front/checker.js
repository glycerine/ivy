import { REPL_FILENAME, diagnosticFilename } from "../diagnostics.js";
import { parseFrontSource, parseFrontSourceFiles } from "./parser.js";
import { TokenKind } from "./token.js";
import { ArrayType, BasicKind, BasicType, ConstObject, FuncObject, InterfaceType, MapType, NamedType, newUniverse, ObjectKind, PackageInfo, PackageNameObject, PointerType, Scope, SignatureType, SliceType, StructType, TypeNameObject, VarObject, assignableTo, methodSet, tuple, varOf } from "./types.js";
export function checkFrontSource(source, filename, config = {}) {
    const parsed = parseFrontSource(source, filename);
    const result = checkFrontFiles(parsed.file ? [parsed.file] : [], config, parsed.diagnostics, parsed.statements);
    return {
        ...result,
        ...(parsed.file ? { file: parsed.file } : {})
    };
}
export function checkFrontSourceFiles(sourceFiles, config = {}) {
    const parsed = parseFrontSourceFiles(sourceFiles);
    const result = checkFrontFiles(parsed.files, config, parsed.diagnostics, parsed.statements);
    return {
        ...result,
        files: parsed.files
    };
}
export function checkFrontFiles(files, config = {}, parserDiagnostics = [], statements = []) {
    const checker = new FrontChecker(config, parserDiagnostics);
    return checker.check(files, statements);
}
class FrontChecker {
    config;
    universe;
    diagnostics;
    info = {
        types: new Map(),
        defs: new Map(),
        uses: new Map(),
        scopes: new Map(),
        selections: new Map()
    };
    pkg;
    constructor(config, parserDiagnostics) {
        this.config = config;
        this.universe = config.universe ?? newUniverse();
        this.diagnostics = [...parserDiagnostics];
        this.pkg = new PackageInfo(config.packagePath ?? "", config.packageName ?? "main", this.universe.scope);
    }
    check(files, statements = []) {
        const firstName = files.find((file) => file.name)?.name?.name;
        const packageName = this.config.packageName ?? firstName ?? "main";
        const packagePath = this.config.packagePath ?? packageName;
        this.pkg = new PackageInfo(packagePath, packageName, this.universe.scope);
        for (const file of files) {
            this.info.scopes.set(file, this.pkg.scope);
            this.declareImports(file);
        }
        this.declareAutomaticImports();
        for (const file of files)
            this.declareTopLevelTypeNames(file);
        for (const file of files)
            this.declareTopLevelObjects(file);
        for (const file of files)
            this.checkTopLevelBodies(file);
        for (const statement of statements)
            this.checkStmt(statement, this.pkg.scope);
        return {
            pkg: this.pkg,
            info: this.info,
            diagnostics: this.diagnostics,
            universe: this.universe
        };
    }
    declareImports(file) {
        for (const spec of file.imports) {
            const path = unquote(spec.path.value);
            const imported = this.config.importer?.import(path) ?? standardPackageInfo(path, this.universe) ?? new PackageInfo(path, importName(path), this.universe.scope);
            const localName = spec.name?.name === "." ? imported.name : spec.name?.name ?? imported.name;
            const existing = this.pkg.scope.lookup(localName);
            if (existing instanceof PackageNameObject && existing.imported.path === imported.path) {
                if (spec.name)
                    this.info.defs.set(spec.name, existing);
                continue;
            }
            const object = new PackageNameObject(localName, this.universe.basic.any, imported, this.pkg.scope, this.pkg);
            this.insert(this.pkg.scope, object, spec.name ?? spec.path);
            if (spec.name)
                this.info.defs.set(spec.name, object);
        }
    }
    declareAutomaticImports() {
        if (this.pkg.scope.lookup("fmt"))
            return;
        const fmt = fmtPackageInfo(this.universe);
        this.insert(this.pkg.scope, new PackageNameObject("fmt", this.universe.basic.any, fmt, this.pkg.scope, this.pkg));
    }
    declareTopLevelTypeNames(file) {
        for (const declaration of file.declarations) {
            if (declaration.kind !== "GenDecl" || declaration.token !== TokenKind.Type)
                continue;
            for (const spec of declaration.specs) {
                if (spec.kind !== "TypeSpec")
                    continue;
                const placeholder = new TypeNameObject(spec.name.name, this.universe.basic.invalid, this.pkg.scope, this.pkg);
                const named = new NamedType(placeholder, this.universe.basic.invalid);
                placeholder.setType(named);
                this.insert(this.pkg.scope, placeholder, spec.name);
            }
        }
    }
    declareTopLevelObjects(file) {
        for (const declaration of file.declarations) {
            if (declaration.kind === "GenDecl") {
                this.declareGenDecl(declaration, this.pkg.scope, true);
            }
            else if (declaration.kind === "FuncDecl") {
                this.declareFuncDecl(declaration);
            }
        }
    }
    declareGenDecl(declaration, scope, topLevel) {
        for (const spec of declaration.specs) {
            if (spec.kind === "TypeSpec") {
                this.bindTypeSpec(spec, scope);
            }
            else if (spec.kind === "ValueSpec") {
                this.bindValueSpec(spec, scope, declaration.token, topLevel);
            }
        }
    }
    bindTypeSpec(spec, scope) {
        const object = scope.lookup(spec.name.name);
        const typeName = object instanceof TypeNameObject
            ? object
            : new TypeNameObject(spec.name.name, this.universe.basic.invalid, scope, this.pkg);
        if (!(object instanceof TypeNameObject))
            this.insert(scope, typeName, spec.name);
        const underlying = this.resolveType(spec.type, scope);
        if (typeName.type instanceof NamedType && !spec.alias) {
            typeName.type.setUnderlying(underlying);
        }
        else {
            typeName.setType(spec.alias ? underlying : new NamedType(typeName, underlying));
        }
        this.info.defs.set(spec.name, typeName);
    }
    bindValueSpec(spec, scope, token, topLevel) {
        const explicitType = spec.type ? this.resolveType(spec.type, scope) : undefined;
        for (const [index, name] of spec.names.entries()) {
            const valueType = spec.values[index] ? this.checkExpr(spec.values[index], scope).type : undefined;
            const type = explicitType ?? (token === TokenKind.Const ? valueType : defaultType(valueType, this.universe)) ?? this.universe.basic.invalid;
            const object = token === TokenKind.Const
                ? new ConstObject(name.name, type, undefined, scope, this.pkg)
                : new VarObject(name.name, type, false, scope, this.pkg);
            this.insert(scope, object, name);
            if (!topLevel && spec.values[index]) {
                this.checkAssignable(spec.values[index], type, scope);
            }
        }
    }
    declareFuncDecl(declaration) {
        const receiver = declaration.receiver ? this.receiverVar(declaration.receiver) : undefined;
        const signature = this.signatureFromFuncType(declaration.type, this.pkg.scope, receiver);
        const object = new FuncObject(declaration.name.name, signature, this.pkg.scope, this.pkg);
        if (receiver) {
            const named = this.receiverNamedType(receiver.type);
            if (named)
                named.addMethod(object);
            this.info.defs.set(declaration.name, object);
            return;
        }
        this.insert(this.pkg.scope, object, declaration.name);
    }
    checkTopLevelBodies(file) {
        for (const declaration of file.declarations) {
            if (declaration.kind === "FuncDecl" && declaration.body) {
                const receiver = declaration.receiver ? this.receiverVar(declaration.receiver) : undefined;
                const signature = this.signatureFromFuncType(declaration.type, this.pkg.scope, receiver);
                const scope = new Scope(this.pkg.scope, `func ${declaration.name.name}`);
                this.info.scopes.set(declaration.body, scope);
                if (receiver && declaration.receiver)
                    this.insert(scope, receiver, firstName(declaration.receiver));
                for (const param of signature.params.variables)
                    this.insert(scope, param);
                for (const result of signature.results.variables) {
                    if (result.name)
                        this.insert(scope, result);
                }
                this.checkBlock(declaration.body, scope, signature);
            }
        }
    }
    checkBlock(block, parent, signature) {
        const scope = this.info.scopes.get(block) ?? new Scope(parent, "block");
        this.info.scopes.set(block, scope);
        for (const statement of block.statements)
            this.checkStmt(statement, scope, signature);
    }
    checkStmt(statement, scope, signature) {
        switch (statement.kind) {
            case "DeclStmt":
                if (statement.decl.kind === "GenDecl")
                    this.declareGenDecl(statement.decl, scope, false);
                break;
            case "BlockStmt":
                this.checkBlock(statement, new Scope(scope, "block"), signature);
                break;
            case "LabeledStmt":
                this.checkStmt(statement.stmt, scope, signature);
                break;
            case "ExprStmt":
                this.checkExpr(statement.expr, scope);
                break;
            case "AssignStmt":
                this.checkAssign(statement, scope);
                break;
            case "IncDecStmt":
                this.checkExpr(statement.expr, scope);
                break;
            case "ReturnStmt":
                this.checkReturn(statement.results, scope, signature);
                break;
            case "IfStmt":
                if (statement.init)
                    this.checkStmt(statement.init, scope, signature);
                this.checkExpr(statement.condition, scope);
                this.checkBlock(statement.body, new Scope(scope, "if"), signature);
                if (statement.else)
                    this.checkStmt(statement.else, scope, signature);
                break;
            case "ForStmt":
                if (statement.init)
                    this.checkStmt(statement.init, scope, signature);
                if (statement.condition)
                    this.checkExpr(statement.condition, scope);
                if (statement.post)
                    this.checkStmt(statement.post, scope, signature);
                this.checkBlock(statement.body, new Scope(scope, "for"), signature);
                break;
            case "RangeStmt":
                this.checkRange(statement, scope, signature);
                break;
            case "SwitchStmt":
                if (statement.init)
                    this.checkStmt(statement.init, scope, signature);
                if (statement.tag)
                    this.checkExpr(statement.tag, scope);
                for (const clause of statement.body) {
                    for (const item of clause.list)
                        this.checkExpr(item, scope);
                    for (const item of clause.body)
                        this.checkStmt(item, scope, signature);
                }
                break;
            case "TypeSwitchStmt":
                if (statement.init)
                    this.checkStmt(statement.init, scope, signature);
                this.checkStmt(statement.assign, scope, signature);
                for (const clause of statement.body) {
                    for (const item of clause.list)
                        this.resolveType(item, scope);
                    for (const item of clause.body)
                        this.checkStmt(item, scope, signature);
                }
                break;
            case "DeferStmt":
                this.checkExpr(statement.call, scope);
                break;
            default:
                break;
        }
    }
    checkAssign(statement, scope) {
        const rhs = statement.rhs.map((expr) => this.checkExpr(expr, scope).type);
        for (const [index, lhs] of statement.lhs.entries()) {
            if (statement.token === TokenKind.Define && lhs.kind === "Ident") {
                const type = defaultType(rhs[index] ?? rhs[0], this.universe) ?? this.universe.basic.invalid;
                this.insert(scope, new VarObject(lhs.name, type, false, scope, this.pkg), lhs);
                continue;
            }
            const lhsType = this.checkExpr(lhs, scope).type;
            if (statement.rhs[index])
                this.checkAssignable(statement.rhs[index], lhsType, scope);
        }
    }
    checkRange(statement, scope, signature) {
        const source = this.checkExpr(statement.source, scope).type.underlying();
        const keyType = this.universe.basic.int64;
        let valueType = this.universe.basic.any;
        if (source instanceof SliceType || source instanceof ArrayType)
            valueType = source.element;
        if (source instanceof MapType)
            valueType = source.value;
        if (source instanceof BasicType && source.basicKind === BasicKind.String)
            valueType = this.universe.basic.string;
        if (statement.token === TokenKind.Define) {
            if (statement.key?.kind === "Ident")
                this.insert(scope, new VarObject(statement.key.name, keyType, false, scope, this.pkg), statement.key);
            if (statement.value?.kind === "Ident")
                this.insert(scope, new VarObject(statement.value.name, valueType, false, scope, this.pkg), statement.value);
        }
        else {
            if (statement.key)
                this.checkExpr(statement.key, scope);
            if (statement.value)
                this.checkExpr(statement.value, scope);
        }
        this.checkBlock(statement.body, new Scope(scope, "range"), signature);
    }
    checkReturn(results, scope, signature) {
        const expected = signature?.results.variables ?? [];
        for (const [index, expr] of results.entries()) {
            const actual = this.checkExpr(expr, scope).type;
            const target = expected[index]?.type;
            if (target && !assignableTo(actual, target)) {
                this.error(`cannot use ${actual.typeString()} as ${target.typeString()} in return`, expr.span);
            }
        }
    }
    checkExpr(expr, scope) {
        const cached = this.info.types.get(expr);
        if (cached)
            return cached;
        let result;
        switch (expr.kind) {
            case "Ident":
                result = this.checkIdent(expr, scope);
                break;
            case "BasicLit":
                result = this.checkBasicLit(expr);
                break;
            case "CellRefExpr":
                result = { mode: "value", type: this.cellType(expr) };
                break;
            case "RangeRefExpr":
                result = { mode: "value", type: new SliceType(this.rangeElementType(expr)) };
                break;
            case "ParenExpr":
                result = this.checkExpr(expr.expr, scope);
                break;
            case "UnaryExpr":
                result = this.checkUnary(expr, scope);
                break;
            case "BinaryExpr":
                result = this.checkBinary(expr, scope);
                break;
            case "SelectorExpr":
                result = this.checkSelector(expr, scope);
                break;
            case "CallExpr":
                result = this.checkCall(expr, scope);
                break;
            case "IndexExpr":
                result = this.checkIndex(expr, scope);
                break;
            case "SliceExpr":
                result = { mode: "value", type: this.checkExpr(expr.object, scope).type };
                if (expr.low)
                    this.checkExpr(expr.low, scope);
                if (expr.high)
                    this.checkExpr(expr.high, scope);
                if (expr.max)
                    this.checkExpr(expr.max, scope);
                break;
            case "TypeAssertExpr":
                result = this.checkTypeAssert(expr, scope);
                break;
            case "CompositeLit":
                result = this.checkCompositeLit(expr, scope);
                break;
            case "FuncLit":
                result = this.checkFuncLit(expr, scope);
                break;
            case "ArrayType":
            case "StructType":
            case "FuncType":
            case "InterfaceType":
            case "MapType":
            case "StarExpr":
            case "Ellipsis":
                result = { mode: "type", type: this.resolveType(expr, scope) };
                break;
            default:
                result = { mode: "invalid", type: this.universe.basic.invalid };
                break;
        }
        this.info.types.set(expr, result);
        return result;
    }
    checkIdent(expr, scope) {
        const found = scope.lookupParent(expr.name);
        if (!found) {
            this.error(`${expr.name} is not declared`, expr.span);
            return { mode: "invalid", type: this.universe.basic.invalid };
        }
        this.info.uses.set(expr, found.object);
        switch (found.object.kind) {
            case ObjectKind.TypeName:
                return { mode: "type", type: found.object.type };
            case ObjectKind.Const:
                return { mode: "constant", type: found.object.type, value: found.object instanceof ConstObject ? found.object.value : undefined };
            case ObjectKind.Builtin:
                return { mode: "builtin", type: found.object.type };
            case ObjectKind.Nil:
                return { mode: "nil", type: found.object.type };
            case ObjectKind.PackageName:
                return { mode: "package", type: found.object.type };
            default:
                return { mode: "value", type: found.object.type };
        }
    }
    checkBasicLit(expr) {
        if (expr.token === TokenKind.IntLiteral)
            return { mode: "constant", type: this.universe.basic.untypedInt, value: parseGoIntLiteral(expr.value) };
        if (expr.token === TokenKind.FloatLiteral)
            return { mode: "constant", type: this.universe.basic.untypedFloat, value: parseGoFloatLiteral(expr.value) };
        if (expr.token === TokenKind.ImagLiteral)
            return { mode: "constant", type: this.universe.basic.untypedComplex, value: expr.value };
        if (expr.token === TokenKind.RuneLiteral)
            return { mode: "constant", type: this.universe.basic.untypedInt, value: parseGoRuneLiteral(expr.value) };
        return { mode: "constant", type: this.universe.basic.untypedString, value: unquote(expr.value) };
    }
    checkUnary(expr, scope) {
        const operand = this.checkExpr(expr.expr, scope);
        if (expr.op === TokenKind.Amp)
            return { mode: "value", type: new PointerType(operand.type) };
        if (expr.op === TokenKind.Arrow) {
            this.error("channels are not supported in Go-junior", expr.span, "GOJR_TYPE_UNSUPPORTED");
            return { mode: "invalid", type: this.universe.basic.invalid };
        }
        return { mode: operand.mode === "constant" ? "constant" : "value", type: operand.type };
    }
    checkBinary(expr, scope) {
        const left = this.checkExpr(expr.left, scope).type;
        const right = this.checkExpr(expr.right, scope).type;
        if ([TokenKind.Equal, TokenKind.NotEqual, TokenKind.Less, TokenKind.LessEqual, TokenKind.Greater, TokenKind.GreaterEqual].includes(expr.op)) {
            return { mode: "value", type: this.universe.basic.bool };
        }
        if (expr.op === TokenKind.AndAnd || expr.op === TokenKind.OrOr)
            return { mode: "value", type: this.universe.basic.bool };
        if (expr.op === TokenKind.Plus) {
            if (isStringLike(left) || isStringLike(right)) {
                if (isStringLike(left) && isStringLike(right)) {
                    return { mode: "value", type: isUntyped(left) && isUntyped(right) ? this.universe.basic.untypedString : this.universe.basic.string };
                }
                this.error(`invalid operation: ${left.typeString()} + ${right.typeString()} (mismatched types ${left.typeString()} and ${right.typeString()})`, expr.span);
                return { mode: "invalid", type: this.universe.basic.invalid };
            }
            if (!isNumericLike(left) || !isNumericLike(right)) {
                this.error(`invalid operation: ${left.typeString()} + ${right.typeString()} (operator + not defined for those types)`, expr.span);
                return { mode: "invalid", type: this.universe.basic.invalid };
            }
        }
        return { mode: isUntyped(left) && isUntyped(right) ? "constant" : "value", type: promoteNumeric(left, right, this.universe) };
    }
    checkSelector(expr, scope) {
        if (expr.object.kind === "Ident") {
            const object = scope.lookupParent(expr.object.name)?.object;
            if (object instanceof PackageNameObject) {
                this.info.uses.set(expr.object, object);
                const selected = object.imported.scope.lookup(expr.selector.name);
                if (selected) {
                    this.info.uses.set(expr.selector, selected);
                    this.info.selections.set(expr, selected);
                    return { mode: selected.kind === ObjectKind.TypeName ? "type" : "value", type: selected.type };
                }
                if (object.imported.scope.names().length === 0)
                    return { mode: "value", type: this.universe.basic.any };
            }
        }
        const base = this.checkExpr(expr.object, scope).type;
        const field = lookupFieldOrMethod(base, expr.selector.name);
        if (field) {
            this.info.uses.set(expr.selector, field);
            this.info.selections.set(expr, field);
            return { mode: field.kind === ObjectKind.TypeName ? "type" : "value", type: field.type };
        }
        this.error(`${expr.selector.name} is not a field or method`, expr.selector.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
    }
    checkCall(expr, scope) {
        const callee = this.checkExpr(expr.fun, scope);
        for (const arg of expr.args)
            this.checkExpr(arg, scope);
        if (callee.mode === "type")
            return { mode: "value", type: callee.type };
        const builtin = expr.fun.kind === "Ident" ? expr.fun.name : undefined;
        const special = builtin ? this.checkBuiltinCall(builtin, expr, scope) : undefined;
        if (special)
            return special;
        const fun = callee.type;
        if (fun instanceof SignatureType) {
            const results = fun.results.variables;
            if (results.length === 0)
                return { mode: "value", type: this.universe.basic.untypedNil };
            if (results.length === 1)
                return { mode: "value", type: results[0]?.type ?? this.universe.basic.invalid };
            return { mode: "value", type: fun.results };
        }
        if (fun instanceof BasicType && fun.basicKind === BasicKind.Any)
            return { mode: "value", type: this.universe.basic.any };
        this.error("cannot call non-function value", expr.fun.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
    }
    checkBuiltinCall(name, expr, scope) {
        switch (name) {
            case "make":
                return { mode: "value", type: expr.args[0] ? this.resolveType(expr.args[0], scope) : this.universe.basic.invalid };
            case "new":
                return { mode: "value", type: new PointerType(expr.args[0] ? this.resolveType(expr.args[0], scope) : this.universe.basic.invalid) };
            case "append":
                return { mode: "value", type: expr.args[0] ? this.checkExpr(expr.args[0], scope).type : this.universe.basic.invalid };
            case "len":
            case "cap":
            case "copy":
                return { mode: "value", type: this.universe.basic.int64 };
            case "complex":
                return { mode: "value", type: this.universe.basic.complex128 };
            case "real":
            case "imag":
                return { mode: "value", type: this.universe.basic.float64 };
            case "min":
            case "max": {
                const first = expr.args[0] ? this.checkExpr(expr.args[0], scope).type : this.universe.basic.invalid;
                const second = expr.args[1] ? this.checkExpr(expr.args[1], scope).type : first;
                return { mode: "value", type: defaultType(promoteNumeric(first, second, this.universe), this.universe) ?? this.universe.basic.invalid };
            }
            case "panic":
            case "panicOn":
            case "print":
            case "println":
            case "delete":
            case "clear":
                return { mode: "value", type: this.universe.basic.untypedNil };
            default:
                return undefined;
        }
    }
    checkIndex(expr, scope) {
        const object = this.checkExpr(expr.object, scope).type.underlying();
        this.checkExpr(expr.index, scope);
        if (object instanceof ArrayType || object instanceof SliceType)
            return { mode: "value", type: object.element };
        if (object instanceof MapType)
            return { mode: "value", type: object.value };
        if (object instanceof BasicType && object.basicKind === BasicKind.String)
            return { mode: "value", type: this.universe.basic.string };
        this.error("cannot index value", expr.span);
        return { mode: "invalid", type: this.universe.basic.invalid };
    }
    checkTypeAssert(expr, scope) {
        this.checkExpr(expr.object, scope);
        if (expr.typeSwitch)
            return { mode: "value", type: this.universe.basic.any };
        const type = expr.type ? this.resolveType(expr.type, scope) : this.universe.basic.any;
        return { mode: "value", type };
    }
    checkCompositeLit(expr, scope) {
        const type = expr.type ? this.resolveType(expr.type, scope) : this.universe.basic.invalid;
        for (const element of expr.elements) {
            if (element.kind === "KeyValueExpr") {
                this.checkExpr(element.key, scope);
                this.checkExpr(element.value, scope);
            }
            else {
                this.checkExpr(element, scope);
            }
        }
        return { mode: "value", type };
    }
    checkFuncLit(expr, scope) {
        const signature = this.signatureFromFuncType(expr.type, scope);
        const child = new Scope(scope, "func literal");
        for (const param of signature.params.variables)
            this.insert(child, param);
        for (const result of signature.results.variables) {
            if (result.name)
                this.insert(child, result);
        }
        this.checkBlock(expr.body, child, signature);
        return { mode: "value", type: signature };
    }
    resolveType(expr, scope) {
        switch (expr.kind) {
            case "Ident": {
                const object = scope.lookupParent(expr.name)?.object;
                if (object?.kind === ObjectKind.TypeName) {
                    this.info.uses.set(expr, object);
                    return object.type;
                }
                this.error(`${expr.name} is not a type`, expr.span);
                return this.universe.basic.invalid;
            }
            case "SelectorExpr": {
                const selected = this.checkSelector(expr, scope);
                return selected.type;
            }
            case "StarExpr":
                return new PointerType(this.resolveType(expr.expr, scope));
            case "ArrayType":
                return this.resolveArrayType(expr, scope);
            case "MapType":
                return new MapType(this.resolveType(expr.key, scope), this.resolveType(expr.value, scope));
            case "StructType":
                return this.resolveStructType(expr, scope);
            case "InterfaceType":
                return this.resolveInterfaceType(expr, scope);
            case "FuncType":
                return this.signatureFromFuncType(expr, scope);
            case "Ellipsis":
                return new SliceType(expr.element ? this.resolveType(expr.element, scope) : this.universe.basic.invalid);
            default:
                this.error("expected type", expr.span);
                return this.universe.basic.invalid;
        }
    }
    resolveArrayType(expr, scope) {
        const element = this.resolveType(expr.element, scope);
        if (!expr.length)
            return new SliceType(element);
        if (expr.length.kind === "BasicLit" && expr.length.token === TokenKind.IntLiteral) {
            return new ArrayType(Number(expr.length.value), element);
        }
        this.checkExpr(expr.length, scope);
        return new ArrayType(expr.inferredLength ? -1 : 0, element);
    }
    resolveStructType(expr, scope) {
        const fields = expr.fields.fields.flatMap((field) => {
            const type = this.resolveType(field.type, scope);
            const tag = field.tag ? unquote(field.tag.value) : undefined;
            if (field.names.length === 0)
                return [{ name: embeddedFieldName(field.type), type, embedded: true, ...(tag ? { tag } : {}) }];
            return field.names.map((name) => ({ name: name.name, type, embedded: false, ...(tag ? { tag } : {}) }));
        });
        return new StructType(fields);
    }
    resolveInterfaceType(expr, scope) {
        const methods = expr.methods.fields.flatMap((field) => {
            if (field.type.kind !== "FuncType") {
                const embedded = this.resolveType(field.type, scope).underlying();
                return embedded instanceof InterfaceType ? embedded.methods : [];
            }
            return field.names.map((name) => new FuncObject(name.name, this.signatureFromFuncType(field.type, scope), scope, this.pkg));
        });
        return new InterfaceType(methods).complete();
    }
    signatureFromFuncType(expr, scope, receiver) {
        const params = this.fieldListTuple(expr.params, scope, true);
        const results = expr.results ? this.fieldListTuple(expr.results, scope, false) : tuple();
        const variadic = params.variables[params.variables.length - 1]?.type instanceof SliceType &&
            expr.params.fields[expr.params.fields.length - 1]?.type.kind === "Ellipsis";
        return new SignatureType(receiver, params, results, variadic);
    }
    fieldListTuple(list, scope, params) {
        const variables = [];
        for (const field of list.fields) {
            const type = this.fieldType(field, scope);
            if (field.names.length === 0) {
                variables.push(varOf("", type));
            }
            else {
                for (const name of field.names) {
                    variables.push(varOf(name.name, type));
                }
            }
        }
        if (params)
            return tuple(...variables);
        return tuple(...variables);
    }
    fieldType(field, scope) {
        if (field.type.kind === "Ellipsis") {
            return new SliceType(field.type.element ? this.resolveType(field.type.element, scope) : this.universe.basic.invalid);
        }
        return this.resolveType(field.type, scope);
    }
    receiverVar(list) {
        const field = list.fields[0];
        if (!field)
            return undefined;
        const name = field.names[0]?.name ?? "";
        return new VarObject(name, this.resolveType(field.type, this.pkg.scope), false, this.pkg.scope, this.pkg);
    }
    receiverNamedType(type) {
        if (type instanceof NamedType)
            return type;
        if (type instanceof PointerType && type.base instanceof NamedType)
            return type.base;
        return undefined;
    }
    cellType(expr) {
        const namespace = this.config.sheetNamespaces?.[expr.namespace.name];
        return namespace?.cells?.[expr.address.raw] ?? namespace?.defaultType ?? this.universe.basic.any;
    }
    rangeElementType(expr) {
        const namespace = this.config.sheetNamespaces?.[expr.namespace.name];
        return namespace?.cells?.[expr.start.raw] ?? namespace?.defaultType ?? this.universe.basic.any;
    }
    checkAssignable(expr, target, scope) {
        const actual = this.checkExpr(expr, scope).type;
        if (!assignableTo(actual, target)) {
            this.error(`cannot use ${actual.typeString()} as ${target.typeString()}`, expr.span);
        }
    }
    insert(scope, object, identNode) {
        const existing = scope.insert(object);
        if (identNode && identNode.kind === "Ident")
            this.info.defs.set(identNode, existing ? undefined : object);
        if (existing && identNode)
            this.error(`${object.name} already declared`, identNode.span);
    }
    error(message, span, code = "GOJR_TYPE001") {
        this.diagnostics.push({
            filename: diagnosticFilename(span, REPL_FILENAME),
            code,
            severity: "error",
            message,
            ...(span ? { span } : {})
        });
    }
}
function unquote(value) {
    if (value.length >= 2 && value.startsWith("`") && value.endsWith("`")) {
        return value.slice(1, -1).replace(/\r/g, "");
    }
    if (value.length >= 2 && value.startsWith("\"") && value.endsWith("\"")) {
        return decodeGoEscaped(value.slice(1, -1));
    }
    return value;
}
function parseGoIntLiteral(value) {
    const text = value.replace(/_/g, "");
    if (/^0[0-7]+$/.test(text))
        return BigInt(`0o${text.slice(1)}`);
    return BigInt(text);
}
function parseGoFloatLiteral(value) {
    const text = value.replace(/_/g, "");
    const hex = /^0[xX]([0-9a-fA-F]*)(?:\.([0-9a-fA-F]*))?[pP]([+-]?[0-9]+)$/.exec(text);
    if (!hex)
        return Number(text);
    const whole = hex[1] || "0";
    const frac = hex[2] || "";
    const exponent = Number(hex[3]);
    const wholeValue = Number.parseInt(whole, 16);
    let fracValue = 0;
    for (let index = 0; index < frac.length; index += 1) {
        fracValue += Number.parseInt(frac[index] ?? "0", 16) / 16 ** (index + 1);
    }
    return (wholeValue + fracValue) * 2 ** exponent;
}
function parseGoRuneLiteral(value) {
    const decoded = decodeGoEscaped(value.slice(1, -1));
    return BigInt([...decoded][0]?.codePointAt(0) ?? 0);
}
function decodeGoEscaped(value) {
    let decoded = "";
    for (let index = 0; index < value.length; index += 1) {
        const char = value[index] ?? "";
        if (char !== "\\") {
            decoded += char;
            continue;
        }
        const next = value[index + 1] ?? "";
        index += 1;
        switch (next) {
            case "a":
                decoded += "\x07";
                break;
            case "b":
                decoded += "\b";
                break;
            case "f":
                decoded += "\f";
                break;
            case "n":
                decoded += "\n";
                break;
            case "r":
                decoded += "\r";
                break;
            case "t":
                decoded += "\t";
                break;
            case "v":
                decoded += "\x0b";
                break;
            case "\\":
            case "\"":
            case "'":
                decoded += next;
                break;
            case "x":
                decoded += codePointFromEscape(value.slice(index + 1, index + 3), 16);
                index += 2;
                break;
            case "u":
                decoded += codePointFromEscape(value.slice(index + 1, index + 5), 16);
                index += 4;
                break;
            case "U":
                decoded += codePointFromEscape(value.slice(index + 1, index + 9), 16);
                index += 8;
                break;
            default:
                if (/^[0-7]$/.test(next)) {
                    const digits = next + value.slice(index + 1, index + 3);
                    decoded += codePointFromEscape(digits, 8);
                    index += 2;
                }
                else {
                    decoded += next;
                }
                break;
        }
    }
    return decoded;
}
function codePointFromEscape(digits, radix) {
    const value = Number.parseInt(digits, radix);
    if (!Number.isFinite(value))
        return "";
    try {
        return String.fromCodePoint(value);
    }
    catch {
        return "";
    }
}
function importName(path) {
    const parts = path.split("/").filter(Boolean);
    return parts[parts.length - 1] ?? path;
}
function firstName(list) {
    return list.fields[0]?.names[0];
}
function embeddedFieldName(expr) {
    if (expr.kind === "Ident")
        return expr.name;
    if (expr.kind === "SelectorExpr")
        return expr.selector.name;
    if (expr.kind === "StarExpr")
        return embeddedFieldName(expr.expr);
    return "";
}
function lookupFieldOrMethod(type, name) {
    const actual = type instanceof PointerType ? type.base.underlying() : type.underlying();
    if (actual instanceof StructType) {
        const field = actual.fields.find((item) => item.name === name);
        if (field)
            return new VarObject(field.name, field.type, field.embedded);
        const promoted = promotedFieldOrMethod(actual, name);
        if (promoted)
            return promoted;
    }
    const directMethod = methodSet(type).find((method) => method.name === name);
    if (directMethod)
        return directMethod;
    if (type instanceof NamedType)
        return type.methodSet().find((method) => method.name === name);
    return undefined;
}
function promotedFieldOrMethod(type, name, seen = new Set()) {
    if (seen.has(type))
        return undefined;
    seen.add(type);
    const matches = [];
    for (const field of type.fields.filter((item) => item.embedded)) {
        const fieldType = field.type instanceof PointerType ? field.type.base : field.type;
        const directMethod = methodSet(field.type).find((method) => method.name === name);
        if (directMethod) {
            matches.push(directMethod);
            continue;
        }
        const underlying = fieldType.underlying();
        if (!(underlying instanceof StructType))
            continue;
        const directField = underlying.fields.find((item) => item.name === name);
        if (directField) {
            matches.push(new VarObject(directField.name, directField.type, directField.embedded));
            continue;
        }
        const promoted = promotedFieldOrMethod(underlying, name, seen);
        if (promoted)
            matches.push(promoted);
    }
    return matches.length === 1 ? matches[0] : undefined;
}
function isStringLike(type) {
    return type instanceof BasicType && (type.basicKind === BasicKind.String || type.basicKind === BasicKind.UntypedString);
}
function isNumericLike(type) {
    return type instanceof BasicType && type.info.has("numeric");
}
function isUntyped(type) {
    return type instanceof BasicType && type.info.has("untyped");
}
function defaultType(type, universe) {
    if (!(type instanceof BasicType))
        return type;
    switch (type.basicKind) {
        case BasicKind.UntypedBool:
            return universe.basic.bool;
        case BasicKind.UntypedInt:
            return universe.basic.int64;
        case BasicKind.UntypedFloat:
            return universe.basic.float64;
        case BasicKind.UntypedComplex:
            return universe.basic.complex128;
        case BasicKind.UntypedString:
            return universe.basic.string;
        default:
            return type;
    }
}
function fmtPackageInfo(universe) {
    const pkg = new PackageInfo("fmt", "fmt", universe.scope);
    const scope = pkg.scope;
    scope.insert(new FuncObject("Printf", new SignatureType(undefined, tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any))), tuple(varOf("", universe.basic.int64)), true), scope, pkg));
    scope.insert(new FuncObject("Sprintf", new SignatureType(undefined, tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any))), tuple(varOf("", universe.basic.string)), true), scope, pkg));
    scope.insert(new FuncObject("Println", new SignatureType(undefined, tuple(varOf("args", new SliceType(universe.basic.any))), tuple(varOf("", universe.basic.int64)), true), scope, pkg));
    return pkg;
}
function testingPackageInfo(universe) {
    const pkg = new PackageInfo("testing", "testing", universe.scope);
    const scope = pkg.scope;
    const typeName = new TypeNameObject("T", universe.basic.invalid, scope, pkg);
    const tType = new NamedType(typeName, new StructType([]));
    typeName.setType(tType);
    scope.insert(typeName);
    const receiver = varOf("", new PointerType(tType));
    const noResults = tuple();
    const addMethod = (name, params = tuple(), results = noResults, variadic = false) => {
        tType.addMethod(new FuncObject(name, new SignatureType(receiver, params, results, variadic), scope, pkg));
    };
    const anyArgs = tuple(varOf("args", new SliceType(universe.basic.any)));
    const formattedArgs = tuple(varOf("format", universe.basic.string), varOf("args", new SliceType(universe.basic.any)));
    addMethod("Fail");
    addMethod("FailNow");
    addMethod("Failed", tuple(), tuple(varOf("", universe.basic.bool)));
    addMethod("Fatal", anyArgs, noResults, true);
    addMethod("Fatalf", formattedArgs, noResults, true);
    addMethod("Error", anyArgs, noResults, true);
    addMethod("Errorf", formattedArgs, noResults, true);
    addMethod("Log", anyArgs, noResults, true);
    addMethod("Logf", formattedArgs, noResults, true);
    addMethod("Name", tuple(), tuple(varOf("", universe.basic.string)));
    addMethod("Helper");
    addMethod("Skip", anyArgs, noResults, true);
    addMethod("Skipf", formattedArgs, noResults, true);
    addMethod("SkipNow");
    addMethod("Skipped", tuple(), tuple(varOf("", universe.basic.bool)));
    scope.insert(new FuncObject("Short", new SignatureType(undefined, tuple(), tuple(varOf("", universe.basic.bool))), scope, pkg));
    scope.insert(new FuncObject("Verbose", new SignatureType(undefined, tuple(), tuple(varOf("", universe.basic.bool))), scope, pkg));
    return pkg;
}
function standardPackageInfo(path, universe) {
    if (path === "fmt")
        return fmtPackageInfo(universe);
    if (path === "testing")
        return testingPackageInfo(universe);
    return undefined;
}
function promoteNumeric(left, right, universe) {
    if (left instanceof BasicType && (left.basicKind === BasicKind.Complex64 || left.basicKind === BasicKind.Complex128))
        return left;
    if (right instanceof BasicType && (right.basicKind === BasicKind.Complex64 || right.basicKind === BasicKind.Complex128))
        return right;
    if (left instanceof BasicType && left.basicKind === BasicKind.UntypedComplex)
        return left;
    if (right instanceof BasicType && right.basicKind === BasicKind.UntypedComplex)
        return right;
    if (left instanceof BasicType && left.basicKind === BasicKind.Float64)
        return left;
    if (right instanceof BasicType && right.basicKind === BasicKind.Float64)
        return right;
    if (left instanceof BasicType && left.basicKind === BasicKind.UntypedFloat)
        return left;
    if (right instanceof BasicType && right.basicKind === BasicKind.UntypedFloat)
        return right;
    if (left instanceof BasicType && left.basicKind === BasicKind.Int64)
        return left;
    if (right instanceof BasicType && right.basicKind === BasicKind.Int64)
        return right;
    if (left instanceof BasicType && left.basicKind === BasicKind.UntypedInt && right instanceof BasicType && right.basicKind === BasicKind.UntypedInt) {
        return universe.basic.untypedInt;
    }
    return left instanceof BasicType ? left : universe.basic.invalid;
}
