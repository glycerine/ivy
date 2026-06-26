import { REPL_FILENAME, diagnosticFilename } from "../diagnostics.js";
import { Bad, Fun, Lbl, NewObj, NewScope, NormalizeAst, parseCellAddress, Typ, Unparen, Var, Walk, Con, ident } from "./ast.js";
import { scanSource } from "./scanner.js";
import { isAssignmentToken, isIdentifierLike, TokenKind } from "./token.js";
export const basic = "basic";
export const labelOk = "labelOk";
export const rangeOk = "rangeOk";
export const maxNestLev = 1e5;
export const declStart = new Set([TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var, TokenKind.Func]);
export const stmtStart = new Set([
    TokenKind.Break,
    TokenKind.Const,
    TokenKind.Continue,
    TokenKind.Defer,
    TokenKind.Fallthrough,
    TokenKind.For,
    TokenKind.Go,
    TokenKind.Goto,
    TokenKind.If,
    TokenKind.Return,
    TokenKind.Select,
    TokenKind.Switch,
    TokenKind.Type,
    TokenKind.Var
]);
export const exprEnd = new Set([TokenKind.Comma, TokenKind.Semicolon, TokenKind.Colon, TokenKind.RParen, TokenKind.RBracket, TokenKind.RBrace]);
export class bailout {
    pos;
    msg;
    constructor(pos, msg = "") {
        this.pos = pos;
        this.msg = msg;
    }
}
export function assert(condition, message = "assertion failed") {
    if (!condition) {
        throw new Error(message);
    }
}
export function trace(p, msg) {
    p.printTrace(msg, "(");
    p.indent += 1;
    return p;
}
export function un(p) {
    p.indent -= 1;
    p.printTrace(")");
}
export function incNestLev(p) {
    p.nestLev += 1;
    if (p.nestLev > maxNestLev) {
        p.error("exceeded max nesting depth", p.currentSpan());
        throw new bailout();
    }
    return p;
}
export function decNestLev(p) {
    p.nestLev -= 1;
}
export function isTypeSwitchAssert(x) {
    return x.kind === "TypeAssertExpr" && x.typeSwitch;
}
export function packIndexExpr(x, _lbrack, exprs, rbrack) {
    switch (exprs.length) {
        case 0:
            throw new Error("internal error: packIndexExpr with empty expr slice");
        case 1:
            return {
                kind: "IndexExpr",
                object: x,
                index: exprs[0],
                span: mergeSpans(x.span, rbrack ?? exprs[0]?.span)
            };
        default:
            return {
                kind: "IndexListExpr",
                object: x,
                indices: exprs,
                span: mergeSpans(x.span, rbrack ?? exprs[exprs.length - 1]?.span)
            };
    }
}
export const PackageClauseOnly = 1 << 0;
export const ImportsOnly = 1 << 1;
export const ParseComments = 1 << 2;
export const Trace = 1 << 3;
export const DeclarationErrors = 1 << 4;
export const SpuriousErrors = 1 << 5;
export const SkipObjectResolution = 1 << 6;
export const AllErrors = SpuriousErrors;
export const debugResolve = false;
export const maxScopeDepth = 1e3;
export const unresolved = NewObj(Bad, "");
export function resolveFile(file, handle, declErr) {
    const pkgScope = NewScope();
    const r = new resolver(handle, declErr, pkgScope, pkgScope, 1);
    for (const decl of file.declarations) {
        Walk(r, decl);
    }
    r.closeScope();
    assert(r.topScope === undefined, "unbalanced scopes");
    assert(r.labelScope === undefined, "unbalanced label scopes");
    let i = 0;
    for (const identNode of r.unresolved) {
        assert(identNode.Obj === unresolved, "object already resolved");
        const obj = r.pkgScope.Lookup(identNode.name);
        if (obj === undefined) {
            delete identNode.Obj;
            r.unresolved[i] = identNode;
            i += 1;
        }
        else if (debugResolve) {
            identNode.Obj = obj;
            r.trace("resolved %s@%v to package object %v", identNode.name, identNode.span?.offset ?? 0, identNode.Obj.Pos());
        }
        else {
            identNode.Obj = obj;
        }
    }
    file.scope = r.pkgScope;
    file.unresolved = r.unresolved.slice(0, i);
}
export class resolver {
    handle;
    declErr;
    topScope;
    pkgScope;
    depth;
    unresolved = [];
    labelScope;
    targetStack = [];
    constructor(handle, declErr, topScope, pkgScope, depth) {
        this.handle = handle;
        this.declErr = declErr;
        this.topScope = topScope;
        this.pkgScope = pkgScope;
        this.depth = depth;
    }
    trace(format, ...args) {
        globalThis.console?.log(`${". ".repeat(this.depth)}${this.sprintf(format, ...args)}`);
    }
    sprintf(format, ...args) {
        return resolverSprintf(format, args);
    }
    openScope(pos) {
        this.depth += 1;
        if (this.depth > maxScopeDepth) {
            throw new bailout(posSpanForResolver(this.handle, pos), "exceeded max scope depth during object resolution");
        }
        if (debugResolve) {
            this.trace("opening scope @%v", pos);
        }
        this.topScope = NewScope(this.topScope);
    }
    closeScope() {
        this.depth -= 1;
        if (debugResolve) {
            this.trace("closing scope");
        }
        this.topScope = this.topScope?.Outer;
    }
    openLabelScope() {
        this.labelScope = NewScope(this.labelScope);
        this.targetStack.push([]);
    }
    closeLabelScope() {
        const n = this.targetStack.length - 1;
        const scope = this.labelScope;
        for (const identNode of this.targetStack[n] ?? []) {
            const obj = scope?.Lookup(identNode.name);
            if (obj === undefined && this.declErr) {
                delete identNode.Obj;
                this.declErr(identPos(identNode), `label ${identNode.name} undefined`);
            }
            else if (obj !== undefined) {
                identNode.Obj = obj;
            }
        }
        this.targetStack.length = Math.max(0, n);
        this.labelScope = this.labelScope?.Outer;
    }
    declare(decl, data, scope, kind, ...idents) {
        if (scope === undefined) {
            return;
        }
        for (const identNode of idents) {
            if (identNode.Obj !== undefined) {
                throw new globalThis.Error(`${identPos(identNode)}: identifier ${identNode.name} already declared or resolved`);
            }
            const obj = NewObj(kind, identNode.name);
            obj.Decl = decl;
            obj.Data = data;
            if (!isIdentNode(decl)) {
                identNode.Obj = obj;
            }
            if (identNode.name !== "_") {
                if (debugResolve) {
                    this.trace("declaring %s@%v", identNode.name, identPos(identNode));
                }
                const alt = scope.Insert(obj);
                if (alt !== undefined && this.declErr) {
                    let prevDecl = "";
                    const pos = alt.Pos();
                    if (pos !== 0) {
                        prevDecl = this.sprintf("\n\tprevious declaration at %v", pos);
                    }
                    this.declErr(identPos(identNode), `${identNode.name} redeclared in this block${prevDecl}`);
                }
            }
        }
    }
    shortVarDecl(decl) {
        let n = 0;
        for (const x of decl.lhs) {
            if (x.kind === "Ident") {
                assert(x.Obj === undefined, "identifier already declared or resolved");
                const obj = NewObj(Var, x.name);
                obj.Decl = decl;
                x.Obj = obj;
                if (x.name !== "_") {
                    if (debugResolve) {
                        this.trace("declaring %s@%v", x.name, identPos(x));
                    }
                    const alt = this.topScope?.Insert(obj);
                    if (alt !== undefined) {
                        x.Obj = alt;
                    }
                    else {
                        n += 1;
                    }
                }
            }
        }
        if (n === 0 && this.declErr && decl.lhs[0]) {
            this.declErr(nodePosForResolver(decl.lhs[0]), "no new variables on left side of :=");
        }
    }
    resolve(identNode, collectUnresolved) {
        if (identNode.Obj !== undefined) {
            throw new globalThis.Error(this.sprintf("%v: identifier %s already declared or resolved", identPos(identNode), identNode.name));
        }
        if (identNode.name === "_") {
            return;
        }
        for (let s = this.topScope; s !== undefined; s = s.Outer) {
            const obj = s.Lookup(identNode.name);
            if (obj !== undefined) {
                if (debugResolve) {
                    this.trace("resolved %v:%s to %v", identPos(identNode), identNode.name, obj.Name);
                }
                assert(obj.Name !== "", "obj with no name");
                if (!isIdentNode(obj.Decl)) {
                    identNode.Obj = obj;
                }
                return;
            }
        }
        if (collectUnresolved) {
            identNode.Obj = unresolved;
            this.unresolved.push(identNode);
        }
    }
    walkExprs(list) {
        for (const node of list) {
            Walk(this, node);
        }
    }
    walkLHS(list) {
        for (const expr of list) {
            const node = Unparen(expr);
            if (node.kind !== "Ident") {
                Walk(this, node);
            }
        }
    }
    walkStmts(list) {
        for (const stmt of list) {
            Walk(this, stmt);
        }
    }
    Visit(node) {
        if (debugResolve && node !== undefined) {
            this.trace("node %s@%v", node.kind, nodePosForResolver(node));
        }
        if (node === undefined) {
            return undefined;
        }
        switch (node.kind) {
            case "Ident":
                this.resolve(node, true);
                break;
            case "FuncLit":
                this.openScope(nodePosForResolver(node));
                this.walkFuncType(node.type);
                this.walkBody(node.body);
                this.closeScope();
                break;
            case "SelectorExpr":
                Walk(this, node.object);
                break;
            case "StructType":
                this.openScope(nodePosForResolver(node));
                this.walkFieldList(node.fields, Var);
                this.closeScope();
                break;
            case "FuncType":
                this.openScope(nodePosForResolver(node));
                this.walkFuncType(node);
                this.closeScope();
                break;
            case "CompositeLit":
                if (node.type !== undefined) {
                    Walk(this, node.type);
                }
                for (const element of node.elements) {
                    if (element.kind === "KeyValueExpr") {
                        if (element.key.kind === "Ident") {
                            this.resolve(element.key, false);
                        }
                        else {
                            Walk(this, element.key);
                        }
                        Walk(this, element.value);
                    }
                    else {
                        Walk(this, element);
                    }
                }
                break;
            case "InterfaceType":
                this.openScope(nodePosForResolver(node));
                this.walkFieldList(node.methods, Fun);
                this.closeScope();
                break;
            case "LabeledStmt":
                this.declare(node, undefined, this.labelScope, Lbl, node.label);
                Walk(this, node.stmt);
                break;
            case "AssignStmt":
                this.walkExprs(node.rhs);
                if (node.token === TokenKind.Define) {
                    this.shortVarDecl(node);
                }
                else {
                    this.walkExprs(node.lhs);
                }
                break;
            case "BranchStmt":
                if (node.token !== TokenKind.Fallthrough && node.label !== undefined && this.targetStack.length > 0) {
                    this.targetStack[this.targetStack.length - 1]?.push(node.label);
                }
                break;
            case "BlockStmt":
                this.openScope(nodePosForResolver(node));
                this.walkStmts(node.statements);
                this.closeScope();
                break;
            case "IfStmt":
                this.openScope(nodePosForResolver(node));
                if (node.init !== undefined) {
                    Walk(this, node.init);
                }
                Walk(this, node.condition);
                Walk(this, node.body);
                if (node.else !== undefined) {
                    Walk(this, node.else);
                }
                this.closeScope();
                break;
            case "CaseClause":
                this.walkExprs(node.list);
                this.openScope(nodePosForResolver(node));
                this.walkStmts(node.body);
                this.closeScope();
                break;
            case "SwitchStmt":
                this.openScope(nodePosForResolver(node));
                if (node.init !== undefined) {
                    Walk(this, node.init);
                }
                if (node.tag !== undefined) {
                    if (node.init !== undefined) {
                        this.openScope(nodePosForResolver(node.tag));
                        Walk(this, node.tag);
                        this.closeScope();
                    }
                    else {
                        Walk(this, node.tag);
                    }
                }
                this.walkStmts(node.body);
                this.closeScope();
                break;
            case "TypeSwitchStmt":
                if (node.init !== undefined) {
                    this.openScope(nodePosForResolver(node));
                    Walk(this, node.init);
                    this.closeScope();
                }
                this.openScope(nodePosForResolver(node.assign));
                Walk(this, node.assign);
                this.walkStmts(node.body);
                this.closeScope();
                break;
            case "CommClause":
                this.openScope(nodePosForResolver(node));
                if (node.comm !== undefined) {
                    Walk(this, node.comm);
                }
                this.walkStmts(node.body);
                this.closeScope();
                break;
            case "SelectStmt":
                this.walkStmts(node.body);
                break;
            case "ForStmt":
                this.openScope(nodePosForResolver(node));
                if (node.init !== undefined) {
                    Walk(this, node.init);
                }
                if (node.condition !== undefined) {
                    Walk(this, node.condition);
                }
                if (node.post !== undefined) {
                    Walk(this, node.post);
                }
                Walk(this, node.body);
                this.closeScope();
                break;
            case "RangeStmt": {
                this.openScope(nodePosForResolver(node));
                Walk(this, node.source);
                const lhs = [node.key, node.value].filter((expr) => expr !== undefined);
                if (lhs.length > 0) {
                    if (node.token === TokenKind.Define) {
                        const as = {
                            kind: "AssignStmt",
                            lhs,
                            token: TokenKind.Define,
                            rhs: [node.source],
                            ...(node.span ? { span: node.span } : {})
                        };
                        this.walkLHS(lhs);
                        this.shortVarDecl(as);
                    }
                    else {
                        this.walkExprs(lhs);
                    }
                }
                Walk(this, node.body);
                this.closeScope();
                break;
            }
            case "GenDecl":
                switch (node.token) {
                    case TokenKind.Const:
                    case TokenKind.Var:
                        for (let i = 0; i < node.specs.length; i += 1) {
                            const spec = node.specs[i];
                            if (spec?.kind !== "ValueSpec") {
                                continue;
                            }
                            const kind = node.token === TokenKind.Var ? Var : Con;
                            this.walkExprs(spec.values);
                            if (spec.type !== undefined) {
                                Walk(this, spec.type);
                            }
                            this.declare(spec, i, this.topScope, kind, ...spec.names);
                        }
                        break;
                    case TokenKind.Type:
                        for (const spec of node.specs) {
                            if (spec.kind !== "TypeSpec") {
                                continue;
                            }
                            this.declare(spec, undefined, this.topScope, Typ, spec.name);
                            if (spec.typeParams !== undefined) {
                                this.openScope(nodePosForResolver(spec));
                                this.walkTParams(spec.typeParams);
                                this.closeScope();
                            }
                            Walk(this, spec.type);
                        }
                        break;
                }
                break;
            case "FuncDecl":
                this.openScope(nodePosForResolver(node));
                this.walkRecv(node.receiver);
                if (node.type.typeParams !== undefined) {
                    this.walkTParams(node.type.typeParams);
                }
                this.resolveList(node.type.params);
                this.resolveList(node.type.results);
                this.declareList(node.receiver, Var);
                this.declareList(node.type.params, Var);
                this.declareList(node.type.results, Var);
                this.walkBody(node.body);
                if (node.receiver === undefined && node.name.name !== "init") {
                    this.declare(node, undefined, this.pkgScope, Fun, node.name);
                }
                this.closeScope();
                break;
            default:
                return this;
        }
        return undefined;
    }
    walkFuncType(typ) {
        this.resolveList(typ.params);
        this.resolveList(typ.results);
        this.declareList(typ.params, Var);
        this.declareList(typ.results, Var);
    }
    resolveList(list) {
        if (list === undefined) {
            return;
        }
        for (const field of list.fields) {
            Walk(this, field.type);
        }
    }
    declareList(list, kind) {
        if (list === undefined) {
            return;
        }
        for (const field of list.fields) {
            this.declare(field, undefined, this.topScope, kind, ...field.names);
        }
    }
    walkRecv(recv) {
        if (recv === undefined || recv.fields.length === 0) {
            return;
        }
        let typ = recv.fields[0]?.type;
        if (typ?.kind === "StarExpr") {
            typ = typ.expr;
        }
        let declareExprs = [];
        let resolveExprs = [];
        switch (typ?.kind) {
            case "IndexExpr":
                declareExprs = [typ.index];
                resolveExprs.push(typ.object);
                break;
            case "IndexListExpr":
                declareExprs = typ.indices;
                resolveExprs.push(typ.object);
                break;
            default:
                if (typ !== undefined) {
                    resolveExprs.push(typ);
                }
                break;
        }
        for (const expr of declareExprs) {
            if (expr.kind === "Ident") {
                this.declare(expr, undefined, this.topScope, Typ, expr);
            }
            else {
                resolveExprs.push(expr);
            }
        }
        for (const expr of resolveExprs) {
            Walk(this, expr);
        }
        for (const field of recv.fields.slice(1)) {
            Walk(this, field.type);
        }
    }
    walkFieldList(list, kind) {
        if (list === undefined) {
            return;
        }
        this.resolveList(list);
        this.declareList(list, kind);
    }
    walkTParams(list) {
        this.declareList(list, Typ);
        this.resolveList(list);
    }
    walkBody(body) {
        if (body === undefined) {
            return;
        }
        this.openLabelScope();
        this.walkStmts(body.statements);
        this.closeLabelScope();
    }
}
function resolverSprintf(format, args) {
    let i = 0;
    return format.replace(/%[sv]/g, () => String(args[i++]));
}
function posSpanForResolver(handle, pos) {
    const position = handle?.Position?.(pos);
    return {
        filename: position?.filename ?? position?.Filename ?? REPL_FILENAME,
        offset: position?.offset ?? position?.Offset ?? pos,
        length: 0,
        line: position?.line ?? position?.Line ?? 1,
        column: position?.column ?? position?.Column ?? 1
    };
}
function nodePosForResolver(node) {
    return node.span?.offset ?? 0;
}
function identPos(node) {
    return node.span?.offset ?? 0;
}
function isIdentNode(value) {
    return typeof value === "object" && value !== null && "kind" in value && value.kind === "Ident";
}
export function parseFrontSource(source, filename) {
    const scanned = scanSource(source, filename);
    const p = new parser(scanned.tokens, scanned.diagnostics, filename);
    const result = p.parseFile();
    if (result.file)
        NormalizeAst(result.file);
    for (const statement of result.statements)
        NormalizeAst(statement);
    return result;
}
export function parseFrontSourceFiles(files) {
    const results = files.map((file) => parseFrontSource(file.source, file.filename));
    return {
        files: results.flatMap((result) => result.file ? [result.file] : []),
        statements: results.flatMap((result) => result.statements),
        diagnostics: results.flatMap((result) => result.diagnostics),
        results
    };
}
export function readSource(filename, src) {
    if (src !== null && src !== undefined) {
        if (typeof src === "string") {
            return [src, undefined];
        }
        if (src instanceof Uint8Array) {
            return [new TextDecoder().decode(src), undefined];
        }
        if (src instanceof ArrayBuffer) {
            return [new TextDecoder().decode(src), undefined];
        }
        if (hasBytes(src)) {
            return [new TextDecoder().decode(src.Bytes()), undefined];
        }
        if (hasRead(src)) {
            const text = src.Read();
            if (typeof text === "string") {
                return [text, undefined];
            }
            if (text instanceof Uint8Array) {
                return [new TextDecoder().decode(text), undefined];
            }
        }
        return [undefined, new Error("invalid source")];
    }
    const reader = hostReadFile();
    if (!reader) {
        return [undefined, new Error(`could not read ${filename}: no host file reader available`)];
    }
    try {
        return [reader(filename), undefined];
    }
    catch (err) {
        return [undefined, err instanceof Error ? err : new Error(String(err))];
    }
}
export function ParseFile(_fset, filename, src, mode = 0) {
    if (_fset === null || _fset === undefined) {
        throw new Error("parser.ParseFile: no token.FileSet provided (fset == nil)");
    }
    const [text, readErr] = readSource(filename, src);
    if (readErr || text === undefined) {
        return [undefined, readErr];
    }
    const result = parseFrontSource(text, filename);
    let file = result.file;
    if (!file) {
        file = {
            kind: "File",
            name: ident(""),
            declarations: [],
            imports: [],
            unresolved: [],
            comments: []
        };
    }
    if ((mode & PackageClauseOnly) !== 0) {
        file = { ...file, declarations: [], imports: [], unresolved: [] };
    }
    else if ((mode & ImportsOnly) !== 0) {
        const declarations = file.declarations.filter((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Import);
        file = { ...file, declarations, imports: declarations.flatMap((decl) => decl.kind === "GenDecl" ? decl.specs.filter((spec) => spec.kind === "ImportSpec") : []) };
    }
    if ((mode & SkipObjectResolution) === 0 && (mode & PackageClauseOnly) === 0) {
        const declErr = (mode & DeclarationErrors) !== 0
            ? (pos, message) => {
                const span = posSpanForResolver(_fset, pos);
                result.diagnostics.push({
                    filename: diagnosticFilename(span, filename),
                    code: "GOJR_PARSE_FRONT001",
                    severity: "error",
                    message,
                    span
                });
            }
            : undefined;
        resolveFile(file, _fset, declErr);
    }
    return [file, diagnosticAsError(result.diagnostics[0])];
}
export function ParseDir(fset, path, filter, mode = 0) {
    const reader = hostReadDir();
    if (!reader) {
        return [undefined, new Error(`could not read directory ${path}: no host directory reader available`)];
    }
    const packages = new Map();
    let first;
    for (const entry of reader(path)) {
        if (!entry.isFile || !entry.name.endsWith(".go")) {
            continue;
        }
        if (filter && !filter(entry)) {
            continue;
        }
        const filename = `${path.replace(/\/$/u, "")}/${entry.name}`;
        const [src, err] = ParseFile(fset, filename, undefined, mode);
        if (src && !err) {
            const name = src.name?.name ?? "";
            let pkg = packages.get(name);
            if (!pkg) {
                pkg = { kind: "Package", name, files: [] };
                packages.set(name, pkg);
            }
            if (Array.isArray(pkg.files)) {
                pkg.files.push(src);
            }
        }
        else if (!first) {
            first = err;
        }
    }
    return [packages, first];
}
export function ParseExprFrom(fset, filename, src, mode = 0) {
    if (fset === null || fset === undefined) {
        throw new Error("parser.ParseExprFrom: no token.FileSet provided (fset == nil)");
    }
    const [text, readErr] = readSource(filename, src);
    if (readErr || text === undefined) {
        return [undefined, readErr];
    }
    const result = parseFrontSource(`package p\nvar _ = ${text}`, filename || REPL_FILENAME);
    const declaration = result.file?.declarations.find((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Var);
    const spec = declaration?.kind === "GenDecl" ? declaration.specs[0] : undefined;
    const expr = spec?.kind === "ValueSpec" ? spec.values[0] : undefined;
    return [expr, diagnosticAsError(result.diagnostics[0])];
}
export function ParseExpr(x) {
    return ParseExprFrom({}, "", x, 0);
}
class parser {
    tokens;
    filename;
    index = 0;
    indent = 0;
    nestLev = 0;
    // Faithful port of go/parser.parser.exprLev:
    // exprLev < 0 means we are parsing an if/for/switch control clause, where
    // a following "{" may be the statement body rather than a composite literal.
    // Parenthesized/call/index subexpressions increment it back into expression
    // context, exactly like the standard parser.
    exprLev = 0;
    // Faithful port of go/parser.parser.inRhs: while parsing right-hand-side
    // expression lists, "=" is treated as equality for tolerant parsing.
    inRhs = false;
    allowBareIdentifierComposite = true;
    allowSpreadsheetRanges = true;
    diagnostics;
    constructor(tokens, diagnostics, filename) {
        this.tokens = tokens;
        this.filename = filename;
        this.diagnostics = [...diagnostics];
    }
    parseFile() {
        let name;
        if (this.match(TokenKind.Package)) {
            name = this.parseIdent("expected package name");
            this.consumeSemi();
        }
        const declarations = [];
        const imports = [];
        const statements = [];
        while (!this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.EOF))
                break;
            if (this.startsDecl()) {
                const declaration = this.parseDecl();
                declarations.push(declaration);
                if (declaration.kind === "GenDecl" && declaration.token === TokenKind.Import) {
                    imports.push(...declaration.specs.filter((spec) => spec.kind === "ImportSpec"));
                }
            }
            else {
                statements.push(this.parseStatement());
            }
            this.consumeSemi();
        }
        return {
            tokens: this.tokens,
            diagnostics: this.diagnostics,
            statements,
            file: {
                kind: "File",
                ...(name ? { name } : {}),
                declarations,
                imports,
                unresolved: [],
                comments: [],
                span: mergeSpans(name?.span, declarations[declarations.length - 1]?.span)
            }
        };
    }
    parseDecl() {
        if (this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var))
            return this.parseGenDecl();
        if (this.at(TokenKind.Func))
            return this.parseFuncDecl();
        const token = this.peek();
        this.error(`expected declaration, found ${token.lexeme || token.kind}`, token.span);
        this.advance();
        return { kind: "BadDecl", span: token.span };
    }
    startsDecl() {
        return this.atAny(TokenKind.Import, TokenKind.Const, TokenKind.Type, TokenKind.Var, TokenKind.Func);
    }
    parseGenDecl() {
        const start = this.advance();
        const token = start.kind;
        const specs = [];
        let grouped = false;
        if (this.match(TokenKind.LParen)) {
            grouped = true;
            while (!this.at(TokenKind.RParen) && !this.at(TokenKind.EOF)) {
                this.skipSemis();
                if (this.at(TokenKind.RParen))
                    break;
                specs.push(this.parseSpec(token));
                this.consumeSemi();
            }
            this.expect(TokenKind.RParen, "expected ')' after declaration group");
        }
        else {
            specs.push(this.parseSpec(token));
        }
        return {
            kind: "GenDecl",
            token,
            specs,
            grouped,
            span: mergeSpans(start.span, specs[specs.length - 1]?.span)
        };
    }
    parseSpec(token) {
        if (token === TokenKind.Import)
            return this.parseImportSpec();
        if (token === TokenKind.Type)
            return this.parseTypeSpec();
        return this.parseValueSpec();
    }
    parseImportSpec() {
        const start = this.peek();
        let name;
        if (this.at(TokenKind.Identifier) && this.peek(1).kind === TokenKind.StringLiteral) {
            name = this.parseIdent("expected import alias");
        }
        else if (this.at(TokenKind.Dot) && this.peek(1).kind === TokenKind.StringLiteral) {
            const dot = this.advance();
            name = ident(".", dot.span);
        }
        const path = this.expectBasicLit(TokenKind.StringLiteral, "expected import path string");
        return {
            kind: "ImportSpec",
            ...(name ? { name } : {}),
            path,
            span: mergeSpans(start.span, path.span)
        };
    }
    parseTypeSpec() {
        const name = this.parseIdent("expected type name");
        let typeParams;
        let alias = false;
        let type;
        if (this.match(TokenKind.LBracket)) {
            const open = this.previous();
            if (isIdentifierLike(this.peek().kind)) {
                const firstName = this.parseIdent("expected type parameter name or array length");
                let expression = firstName;
                if (!this.at(TokenKind.LBracket)) {
                    this.withExpressionLevel(() => {
                        expression = this.parseBinaryExpression(this.parsePrimaryFrom(expression), 1);
                    });
                }
                const { name: paramName, type: paramType } = extractName(expression, this.at(TokenKind.Comma));
                if (paramName && (paramType || !this.at(TokenKind.RBracket))) {
                    const spec = {
                        kind: "TypeSpec",
                        name,
                        type: badExpr(open.span),
                        alias: false
                    };
                    this.parseGenericType(spec, open, paramName, paramType);
                    return {
                        ...spec,
                        span: mergeSpans(name.span, spec.type.span)
                    };
                }
                else {
                    type = this.parseArrayTypeAfterOpen(open, expression);
                }
            }
            else {
                type = this.parseArrayTypeAfterOpen(open);
            }
        }
        else {
            alias = this.match(TokenKind.Assign);
            type = this.parseType();
        }
        return {
            kind: "TypeSpec",
            name,
            ...(typeParams ? { typeParams } : {}),
            type,
            alias,
            span: mergeSpans(name.span, type.span)
        };
    }
    parseGenericType(spec, open, name0, typ0) {
        spec.typeParams = this.parseTypeParameterListAfterOpen(open, name0, typ0);
        spec.alias = this.match(TokenKind.Assign);
        spec.type = this.parseType();
    }
    parseValueSpec() {
        const names = this.parseIdentList();
        let type;
        let values = [];
        if (!this.atAny(TokenKind.Assign, TokenKind.Semicolon, TokenKind.RParen, TokenKind.EOF)) {
            type = this.parseType();
        }
        if (this.match(TokenKind.Assign)) {
            values = this.parseExpressionList(true);
        }
        return {
            kind: "ValueSpec",
            names,
            ...(type ? { type } : {}),
            values,
            span: mergeSpans(names[0]?.span, values[values.length - 1]?.span ?? type?.span ?? names[names.length - 1]?.span)
        };
    }
    parseFuncDecl() {
        const start = this.expect(TokenKind.Func, "expected func");
        let receiver;
        if (this.at(TokenKind.LParen) && this.looksLikeReceiver()) {
            receiver = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        }
        const name = this.parseIdent("expected function name");
        const typeParams = this.at(TokenKind.LBracket) ? this.parseTypeParamList() : undefined;
        const type = this.parseSignature(start.span, typeParams);
        const body = this.at(TokenKind.LBrace) ? this.parseBlock() : undefined;
        return {
            kind: "FuncDecl",
            ...(receiver ? { receiver } : {}),
            name,
            type,
            ...(body ? { body } : {}),
            span: mergeSpans(start.span, body?.span ?? type.span)
        };
    }
    parseStatement() {
        this.skipSemis();
        if (this.at(TokenKind.RBrace)) {
            return { kind: "EmptyStmt", implicit: true, span: this.peek().span };
        }
        if (this.at(TokenKind.LBrace))
            return this.parseBlock();
        if (this.atAny(TokenKind.Const, TokenKind.Type, TokenKind.Var)) {
            return { kind: "DeclStmt", decl: this.parseGenDecl() };
        }
        if (this.match(TokenKind.Return)) {
            const start = this.previous();
            const results = this.atAny(TokenKind.Semicolon, TokenKind.RBrace, TokenKind.EOF) ? [] : this.parseExpressionList(true);
            return { kind: "ReturnStmt", results, span: mergeSpans(start.span, results[results.length - 1]?.span ?? start.span) };
        }
        if (this.atAny(TokenKind.Break, TokenKind.Continue, TokenKind.Goto, TokenKind.Fallthrough)) {
            const token = this.advance();
            const acceptsLabel = token.kind === TokenKind.Goto ||
                ((token.kind === TokenKind.Break || token.kind === TokenKind.Continue) && this.at(TokenKind.Identifier));
            const label = acceptsLabel ? this.parseIdent("expected label") : undefined;
            return {
                kind: "BranchStmt",
                token: token.kind,
                ...(label ? { label } : {}),
                span: mergeSpans(token.span, label?.span)
            };
        }
        if (this.match(TokenKind.Defer)) {
            const start = this.previous();
            const expression = this.parseRhsExpression();
            const call = expression.kind === "CallExpr"
                ? expression
                : { kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) };
            return { kind: "DeferStmt", call, span: mergeSpans(start.span, call.span) };
        }
        if (this.at(TokenKind.Go))
            return this.parseGoStmt();
        if (this.at(TokenKind.Select))
            return this.parseSelectStmt();
        if (this.at(TokenKind.For))
            return this.parseForStmt();
        if (this.at(TokenKind.If))
            return this.parseIfStmt();
        if (this.at(TokenKind.Switch))
            return this.parseSwitchStmt();
        return this.parseSimpleStmt("labelOk");
    }
    parseSimpleStmt(mode = "basic") {
        return this.parseSimpleStmtWithRange(mode).statement;
    }
    parseSimpleStmtWithRange(mode) {
        const lhs = this.parseExpressionList(false);
        if (this.match(TokenKind.Arrow)) {
            if (lhs.length !== 1) {
                this.error("send statement expects one channel expression", lhs[1]?.span ?? lhs[0]?.span);
            }
            const value = this.parseRhsExpression();
            return {
                statement: {
                    kind: "SendStmt",
                    channel: lhs[0] ?? badExpr(value.span),
                    value,
                    span: mergeSpans(lhs[0]?.span, value.span)
                },
                isRange: false
            };
        }
        if (isAssignmentToken(this.peek().kind)) {
            const token = this.advance();
            let rhs;
            let isRange = false;
            if ((token.kind === TokenKind.Assign || token.kind === TokenKind.Define) && mode === "rangeOk" && this.match(TokenKind.Range)) {
                rhs = [this.parseRhsExpression()];
                isRange = true;
            }
            else {
                rhs = this.parseExpressionList(true);
            }
            return {
                statement: {
                    kind: "AssignStmt",
                    lhs,
                    token: token.kind,
                    rhs,
                    span: mergeSpans(lhs[0]?.span, rhs[rhs.length - 1]?.span)
                },
                isRange
            };
        }
        if (this.match(TokenKind.Colon)) {
            const colon = this.previous();
            if (mode === "labelOk" && lhs[0]?.kind === "Ident" && lhs.length === 1) {
                const stmt = this.parseStatement();
                return {
                    statement: {
                        kind: "LabeledStmt",
                        label: lhs[0],
                        stmt,
                        span: mergeSpans(lhs[0].span, stmt.span)
                    },
                    isRange: false
                };
            }
            this.error("illegal label declaration", colon.span);
            return { statement: { kind: "BadStmt", span: mergeSpans(lhs[0]?.span, colon.span) }, isRange: false };
        }
        if (this.at(TokenKind.PlusPlus) || this.at(TokenKind.MinusMinus)) {
            const token = this.advance();
            return {
                statement: {
                    kind: "IncDecStmt",
                    expr: lhs[0] ?? badExpr(token.span),
                    token: token.kind,
                    span: mergeSpans(lhs[0]?.span, token.span)
                },
                isRange: false
            };
        }
        if (lhs.length > 1)
            this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
        return {
            statement: { kind: "ExprStmt", expr: lhs[0] ?? badExpr(this.peek().span), span: mergeSpans(lhs[0]?.span, lhs[0]?.span) },
            isRange: false
        };
    }
    parseGoStmt() {
        const start = this.expect(TokenKind.Go, "expected go");
        const expression = this.parseRhsExpression();
        if (expression.kind !== "CallExpr") {
            this.error("go statement requires function call", expression.span);
            return {
                kind: "GoStmt",
                call: { kind: "CallExpr", fun: expression, args: [], ellipsis: false, span: mergeSpans(expression.span, expression.span) },
                span: mergeSpans(start.span, expression.span)
            };
        }
        return { kind: "GoStmt", call: expression, span: mergeSpans(start.span, expression.span) };
    }
    parseIfHeader() {
        let init;
        let conditionStatement;
        if (this.at(TokenKind.LBrace)) {
            this.error("missing condition in if statement", this.peek().span);
            return { condition: badExpr(this.peek().span) };
        }
        if (!this.at(TokenKind.Semicolon)) {
            if (this.match(TokenKind.Var))
                this.error("var declaration not allowed in if initializer", this.previous().span);
            init = this.parseSimpleStmt("basic");
        }
        if (!this.at(TokenKind.LBrace)) {
            this.expect(TokenKind.Semicolon, "expected ';' after if init statement");
            if (!this.at(TokenKind.LBrace))
                conditionStatement = this.parseSimpleStmt("basic");
        }
        else {
            conditionStatement = init;
            init = undefined;
        }
        return {
            ...(init ? { init } : {}),
            condition: this.statementExpression(conditionStatement, "boolean expression") ?? badExpr(this.peek().span)
        };
    }
    parseIfStmt() {
        const start = this.expect(TokenKind.If, "expected if");
        const { init, condition } = this.withControlClause(() => this.parseIfHeader());
        const body = this.parseBlock();
        let elseStmt;
        if (this.match(TokenKind.Else)) {
            elseStmt = this.at(TokenKind.If) ? this.parseIfStmt() : this.parseBlock();
        }
        return {
            kind: "IfStmt",
            ...(init ? { init } : {}),
            condition,
            body,
            ...(elseStmt ? { else: elseStmt } : {}),
            span: mergeSpans(start.span, elseStmt?.span ?? body.span)
        };
    }
    parseForStmt() {
        const start = this.expect(TokenKind.For, "expected for");
        const header = this.withControlClause(() => {
            let first;
            let second;
            let third;
            let isRange = false;
            if (!this.at(TokenKind.LBrace)) {
                if (!this.at(TokenKind.Semicolon)) {
                    if (this.match(TokenKind.Range)) {
                        const source = this.parseRhsExpression();
                        second = { kind: "AssignStmt", lhs: [], token: TokenKind.Assign, rhs: [source], span: source.span ?? start.span };
                        isRange = true;
                    }
                    else {
                        const parsed = this.parseSimpleStmtWithRange("rangeOk");
                        second = parsed.statement;
                        isRange = parsed.isRange;
                    }
                }
                if (!isRange && this.match(TokenKind.Semicolon)) {
                    first = second;
                    second = undefined;
                    if (!this.at(TokenKind.Semicolon))
                        second = this.parseSimpleStmt("basic");
                    this.expect(TokenKind.Semicolon, "expected ';' in for clause");
                    if (!this.at(TokenKind.LBrace))
                        third = this.parseSimpleStmt("basic");
                }
            }
            return { first, second, third, isRange };
        });
        const body = this.parseBlock();
        if (header.isRange) {
            const assign = header.second?.kind === "AssignStmt" ? header.second : undefined;
            const lhs = assign?.lhs ?? [];
            if (lhs.length > 2)
                this.error("expected at most 2 expressions", lhs[2]?.span ?? assign?.span);
            const source = assign?.rhs[0] ?? badExpr(header.second?.span ?? start.span);
            return {
                kind: "RangeStmt",
                ...(lhs[0] ? { key: lhs[0] } : {}),
                ...(lhs[1] ? { value: lhs[1] } : {}),
                token: assign?.token === TokenKind.Define ? TokenKind.Define : TokenKind.Assign,
                source,
                body,
                span: mergeSpans(start.span, body.span)
            };
        }
        const condition = this.statementExpression(header.second, "boolean or range expression");
        return {
            kind: "ForStmt",
            ...(header.first ? { init: header.first } : {}),
            ...(condition ? { condition } : {}),
            ...(header.third ? { post: header.third } : {}),
            body,
            span: mergeSpans(start.span, body.span)
        };
    }
    parseSwitchStmt() {
        const start = this.expect(TokenKind.Switch, "expected switch");
        const { init, tag, assign, typeSwitch } = this.withControlClause(() => {
            let init;
            let tag;
            let assign;
            let typeSwitch = false;
            if (!this.at(TokenKind.LBrace)) {
                const first = this.at(TokenKind.Semicolon) ? undefined : this.parseSimpleStmt("basic");
                if (this.match(TokenKind.Semicolon)) {
                    init = first;
                    if (!this.at(TokenKind.LBrace)) {
                        const second = this.parseSimpleStmt("basic");
                        if (this.isTypeSwitchGuard(second)) {
                            assign = second;
                            typeSwitch = true;
                        }
                        else if (second.kind === "ExprStmt") {
                            tag = second.expr;
                        }
                        else {
                            this.error("expected switch expression or type switch guard after ';'", second.span);
                        }
                    }
                }
                else if (first && this.isTypeSwitchGuard(first)) {
                    assign = first;
                    typeSwitch = true;
                }
                else if (first?.kind === "ExprStmt") {
                    tag = first.expr;
                }
                else if (first) {
                    this.error("expected ';' after switch init statement", first.span);
                    init = first;
                }
            }
            return { init, tag, assign, typeSwitch };
        });
        const { clauses, span } = this.parseSwitchBody();
        if (typeSwitch) {
            return {
                kind: "TypeSwitchStmt",
                ...(init ? { init } : {}),
                assign: assign ?? { kind: "BadStmt", span: start.span },
                body: clauses,
                span: mergeSpans(start.span, span)
            };
        }
        return {
            kind: "SwitchStmt",
            ...(init ? { init } : {}),
            ...(tag ? { tag } : {}),
            body: clauses,
            span: mergeSpans(start.span, span)
        };
    }
    parseSwitchBody() {
        const start = this.expect(TokenKind.LBrace, "expected '{' after switch");
        const clauses = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
                clauses.push(this.parseCaseClause());
            }
            else {
                const token = this.peek();
                this.error("expected case or default in switch body", token.span);
                this.advance();
            }
        }
        const end = this.expect(TokenKind.RBrace, "expected '}' after switch body");
        return { clauses, span: mergeSpans(start.span, end.span) };
    }
    parseSelectStmt() {
        const start = this.expect(TokenKind.Select, "expected select");
        const open = this.expect(TokenKind.LBrace, "expected '{' after select");
        const clauses = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            if (this.at(TokenKind.Case) || this.at(TokenKind.Default)) {
                clauses.push(this.parseCommClause());
            }
            else {
                const token = this.peek();
                this.error("expected case or default in select body", token.span);
                this.advance();
            }
        }
        const close = this.expect(TokenKind.RBrace, "expected '}' after select body");
        return {
            kind: "SelectStmt",
            body: clauses,
            span: mergeSpans(start.span, close.span ?? open.span)
        };
    }
    parseCommClause() {
        const start = this.advance();
        const isDefault = start.kind === TokenKind.Default;
        let comm;
        if (!isDefault && !this.at(TokenKind.Colon)) {
            const lhs = this.parseExpressionList(false);
            if (this.match(TokenKind.Arrow)) {
                if (lhs.length > 1)
                    this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
                const value = this.parseRhsExpression();
                comm = {
                    kind: "SendStmt",
                    channel: lhs[0] ?? badExpr(value.span),
                    value,
                    span: mergeSpans(lhs[0]?.span, value.span)
                };
            }
            else if (this.at(TokenKind.Assign) || this.at(TokenKind.Define)) {
                if (lhs.length > 2)
                    this.error("expected 1 or 2 expressions", lhs[2]?.span ?? lhs[0]?.span);
                const token = this.advance();
                const rhs = this.parseRhsExpression();
                comm = {
                    kind: "AssignStmt",
                    lhs,
                    token: token.kind,
                    rhs: [rhs],
                    span: mergeSpans(lhs[0]?.span, rhs.span)
                };
            }
            else {
                if (lhs.length > 1)
                    this.error("expected 1 expression", lhs[1]?.span ?? lhs[0]?.span);
                comm = {
                    kind: "ExprStmt",
                    expr: lhs[0] ?? badExpr(start.span),
                    span: mergeSpans(lhs[0]?.span, lhs[0]?.span)
                };
            }
        }
        this.expect(TokenKind.Colon, "expected ':' after select case");
        const body = [];
        while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
            this.skipSemis();
            if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF))
                break;
            body.push(this.parseStatement());
            this.consumeSemi();
        }
        return {
            kind: "CommClause",
            ...(comm ? { comm } : {}),
            body,
            default: isDefault,
            span: mergeSpans(start.span, body[body.length - 1]?.span ?? comm?.span ?? start.span)
        };
    }
    parseCaseClause() {
        const start = this.advance();
        const isDefault = start.kind === TokenKind.Default;
        const list = isDefault ? [] : this.withSpreadsheetRanges(false, () => this.parseExpressionList(true));
        this.expect(TokenKind.Colon, "expected ':' after switch case");
        const body = [];
        while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
            this.skipSemis();
            if (this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF))
                break;
            body.push(this.parseStatement());
            this.consumeSemi();
        }
        return {
            kind: "CaseClause",
            list,
            body,
            default: isDefault,
            span: mergeSpans(start.span, body[body.length - 1]?.span ?? list[list.length - 1]?.span ?? start.span)
        };
    }
    isTypeSwitchGuard(statement) {
        if (statement.kind === "ExprStmt") {
            return statement.expr.kind === "TypeAssertExpr" && statement.expr.typeSwitch;
        }
        if (statement.kind !== "AssignStmt" || statement.rhs.length !== 1)
            return false;
        const rhs = statement.rhs[0];
        return rhs?.kind === "TypeAssertExpr" && rhs.typeSwitch;
    }
    parseBlock() {
        const start = this.expect(TokenKind.LBrace, "expected '{'");
        const statements = [];
        while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(TokenKind.RBrace))
                break;
            statements.push(this.parseStatement());
            this.consumeSemi();
        }
        const end = this.expect(TokenKind.RBrace, "expected '}'");
        return {
            kind: "BlockStmt",
            statements,
            span: mergeSpans(start.span, end.span)
        };
    }
    parseExpressionList(inRhs = false) {
        return this.withRhs(inRhs, () => {
            const expressions = [this.parseExpression()];
            while (this.match(TokenKind.Comma)) {
                expressions.push(this.parseExpression());
            }
            return expressions;
        });
    }
    parseRhsExpression() {
        return this.withRhs(true, () => this.parseExpression());
    }
    parseExpression(minPrecedence = 1) {
        const left = this.parseUnary();
        return this.parseBinaryExpression(left, minPrecedence);
    }
    parseBinaryExpression(leftOperand, minPrecedence = 1) {
        let left = leftOperand;
        while (true) {
            const operatorKind = this.inRhs && this.peek().kind === TokenKind.Assign
                ? TokenKind.Equal
                : this.peek().kind;
            const precedence = binaryPrecedence(operatorKind);
            if (precedence < minPrecedence)
                break;
            const operator = this.advance();
            const right = this.parseExpression(precedence + 1);
            left = {
                kind: "BinaryExpr",
                left,
                op: operatorKind,
                right,
                span: mergeSpans(left.span, right.span)
            };
        }
        return left;
    }
    statementExpression(statement, want) {
        if (!statement)
            return undefined;
        if (statement.kind === "ExprStmt")
            return statement.expr;
        const found = statement.kind === "AssignStmt" ? "assignment" : "simple statement";
        this.error(`expected ${want}, found ${found} (missing parentheses around composite literal?)`, statement.span);
        return badExpr(statement.span);
    }
    parseUnary() {
        if (this.atAny(TokenKind.Plus, TokenKind.Minus, TokenKind.Bang, TokenKind.Caret, TokenKind.Amp, TokenKind.Arrow)) {
            const operator = this.advance();
            const expr = this.parseUnary();
            return {
                kind: "UnaryExpr",
                op: operator.kind,
                expr,
                span: mergeSpans(operator.span, expr.span)
            };
        }
        if (this.match(TokenKind.Star)) {
            const start = this.previous();
            const expr = this.parseUnary();
            return { kind: "StarExpr", expr, span: mergeSpans(start.span, expr.span) };
        }
        return this.parsePrimary();
    }
    parsePrimary() {
        return this.parsePrimaryFrom(this.parseOperand());
    }
    parsePrimaryFrom(start) {
        let expression = start;
        while (true) {
            if (this.match(TokenKind.Dot)) {
                const dot = this.previous();
                if (this.match(TokenKind.LParen)) {
                    if (this.match(TokenKind.Type)) {
                        const close = this.expect(TokenKind.RParen, "expected ')' after type switch guard");
                        expression = {
                            kind: "TypeAssertExpr",
                            object: expression,
                            typeSwitch: true,
                            span: mergeSpans(expression.span, close.span)
                        };
                    }
                    else {
                        const type = this.parseType();
                        const close = this.expect(TokenKind.RParen, "expected ')' after type assertion");
                        expression = {
                            kind: "TypeAssertExpr",
                            object: expression,
                            type,
                            typeSwitch: false,
                            span: mergeSpans(expression.span, close.span)
                        };
                    }
                    continue;
                }
                const cellCandidate = this.peek();
                const cellStart = (cellCandidate.kind === TokenKind.CellAddress || cellCandidate.kind === TokenKind.Identifier)
                    ? parseCellAddress(cellCandidate.lexeme)
                    : undefined;
                const rangeEndCandidate = this.peek(2);
                const hasSpreadsheetRangeEnd = this.allowSpreadsheetRanges &&
                    this.peek(1).kind === TokenKind.Colon &&
                    (rangeEndCandidate.kind === TokenKind.CellAddress || rangeEndCandidate.kind === TokenKind.Identifier) &&
                    !!parseCellAddress(rangeEndCandidate.lexeme);
                if (cellStart && expression.kind === "Ident" && (cellCandidate.kind === TokenKind.CellAddress || hasSpreadsheetRangeEnd)) {
                    const cellToken = this.advance();
                    if (this.match(TokenKind.Colon)) {
                        const { token: endToken, address: end } = this.expectCellAddress("expected cell address after ':'");
                        if (!end) {
                            this.error("malformed spreadsheet range end", endToken.span);
                            expression = badExpr(endToken.span);
                        }
                        else {
                            expression = {
                                kind: "RangeRefExpr",
                                namespace: expression,
                                start: cellStart,
                                end,
                                span: mergeSpans(expression.span, endToken.span)
                            };
                        }
                    }
                    else {
                        expression = {
                            kind: "CellRefExpr",
                            namespace: expression,
                            address: cellStart,
                            span: mergeSpans(expression.span, cellToken.span)
                        };
                    }
                    continue;
                }
                const selector = this.parseIdent("expected selector after '.'");
                expression = {
                    kind: "SelectorExpr",
                    object: expression,
                    selector,
                    span: mergeSpans(expression.span, selector.span ?? dot.span)
                };
                continue;
            }
            if (this.match(TokenKind.LParen)) {
                const args = [];
                let ellipsis = false;
                if (!this.at(TokenKind.RParen)) {
                    args.push(this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression())));
                    if (this.match(TokenKind.Ellipsis))
                        ellipsis = true;
                    while (this.match(TokenKind.Comma) && !this.at(TokenKind.RParen)) {
                        args.push(this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression())));
                        if (this.match(TokenKind.Ellipsis))
                            ellipsis = true;
                    }
                }
                const close = this.expect(TokenKind.RParen, "expected ')' after arguments");
                expression = {
                    kind: "CallExpr",
                    fun: expression,
                    args,
                    ellipsis,
                    span: mergeSpans(expression.span, close.span)
                };
                continue;
            }
            if (this.match(TokenKind.LBracket)) {
                const open = this.previous();
                const low = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket)
                    ? undefined
                    : this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
                if (this.match(TokenKind.Colon)) {
                    const high = this.at(TokenKind.Colon) || this.at(TokenKind.RBracket)
                        ? undefined
                        : this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
                    const max = this.match(TokenKind.Colon)
                        ? this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()))
                        : undefined;
                    const close = this.expect(TokenKind.RBracket, "expected ']' after slice");
                    expression = {
                        kind: "SliceExpr",
                        object: expression,
                        ...(low ? { low } : {}),
                        ...(high ? { high } : {}),
                        ...(max ? { max } : {}),
                        span: mergeSpans(expression.span, close.span)
                    };
                }
                else if (this.match(TokenKind.Comma)) {
                    const indices = [low ?? badExpr(open.span)];
                    while (!this.at(TokenKind.RBracket) && !this.at(TokenKind.EOF)) {
                        indices.push(this.withExpressionLevel(() => this.parseType()));
                        if (!this.match(TokenKind.Comma))
                            break;
                    }
                    const close = this.expect(TokenKind.RBracket, "expected ']' after type arguments");
                    expression = indices.length === 1
                        ? {
                            kind: "IndexExpr",
                            object: expression,
                            index: indices[0],
                            span: mergeSpans(expression.span, close.span)
                        }
                        : {
                            kind: "IndexListExpr",
                            object: expression,
                            indices,
                            span: mergeSpans(expression.span, close.span)
                        };
                }
                else {
                    const close = this.expect(TokenKind.RBracket, "expected ']' after index");
                    expression = {
                        kind: "IndexExpr",
                        object: expression,
                        index: low ?? badExpr(close.span),
                        span: mergeSpans(expression.span, close.span)
                    };
                }
                continue;
            }
            if (this.at(TokenKind.LBrace) && this.canUseCompositeLiteralType(expression)) {
                this.advance();
                expression = this.finishCompositeLiteral(expression);
                continue;
            }
            break;
        }
        return expression;
    }
    parseOperand() {
        const token = this.peek();
        if (this.at(TokenKind.Identifier) || this.at(TokenKind.CellAddress))
            return this.parseIdent("expected identifier");
        if (this.at(TokenKind.True) || this.at(TokenKind.False) || this.at(TokenKind.Nil)) {
            const keyword = this.advance();
            return ident(keyword.lexeme, keyword.span);
        }
        if (this.at(TokenKind.IntLiteral) || this.at(TokenKind.FloatLiteral) || this.at(TokenKind.ImagLiteral) || this.at(TokenKind.RuneLiteral) || this.at(TokenKind.StringLiteral)) {
            return this.expectBasicLit(this.peek().kind, "expected literal");
        }
        if (this.match(TokenKind.LParen)) {
            const start = this.previous();
            const expr = this.withExpressionLevel(() => this.withBareIdentifierComposites(true, () => this.parseRhsExpression()));
            const close = this.expect(TokenKind.RParen, "expected ')'");
            return { kind: "ParenExpr", expr, span: mergeSpans(start.span, close.span) };
        }
        if (this.atAny(TokenKind.LBracket, TokenKind.Map, TokenKind.Struct, TokenKind.Interface, TokenKind.Chan))
            return this.parseType();
        if (this.at(TokenKind.Func))
            return this.parseFuncTypeOrLit();
        this.error(`expected expression, found ${token.lexeme || token.kind}`, token.span);
        this.advance();
        return badExpr(token.span);
    }
    canUseCompositeLiteralType(expression) {
        if (expression.kind === "ArrayType" || expression.kind === "MapType" || expression.kind === "StructType") {
            return true;
        }
        if ((expression.kind === "Ident" && !["true", "false", "nil"].includes(expression.name)) ||
            expression.kind === "SelectorExpr" ||
            expression.kind === "IndexExpr" ||
            expression.kind === "IndexListExpr") {
            return this.exprLev >= 0 && this.allowBareIdentifierComposite;
        }
        return false;
    }
    finishCompositeLiteral(type, startSpan = type?.span) {
        const elements = [];
        this.withExpressionLevel(() => {
            while (!this.at(TokenKind.RBrace) && !this.at(TokenKind.EOF)) {
                this.skipSemis();
                if (this.at(TokenKind.RBrace))
                    break;
                const first = this.parseCompositeLiteralElement();
                if (this.match(TokenKind.Colon)) {
                    const value = this.parseCompositeLiteralElement();
                    elements.push({
                        kind: "KeyValueExpr",
                        key: first,
                        value,
                        span: mergeSpans(first.span, value.span)
                    });
                }
                else {
                    elements.push(first);
                }
                this.match(TokenKind.Comma);
                this.consumeSemi();
            }
        });
        const close = this.expect(TokenKind.RBrace, "expected '}' after composite literal");
        return {
            kind: "CompositeLit",
            ...(type ? { type } : {}),
            elements,
            span: mergeSpans(startSpan, close.span)
        };
    }
    parseCompositeLiteralElement() {
        if (this.match(TokenKind.LBrace)) {
            const start = this.previous();
            return this.finishCompositeLiteral(undefined, start.span);
        }
        return this.parseRhsExpression();
    }
    parseType() {
        let left = this.parseTypeTerm();
        while (this.match(TokenKind.Or)) {
            const operator = this.previous();
            const right = this.parseTypeTerm();
            left = {
                kind: "BinaryExpr",
                left,
                op: operator.kind,
                right,
                span: mergeSpans(left.span, right.span)
            };
        }
        return left;
    }
    parseTypeTerm() {
        const start = this.peek();
        if (this.match(TokenKind.Tilde)) {
            const expr = this.parseTypeTerm();
            return {
                kind: "UnaryExpr",
                op: TokenKind.Tilde,
                expr,
                span: mergeSpans(start.span, expr.span)
            };
        }
        if (this.match(TokenKind.Arrow)) {
            const chan = this.expect(TokenKind.Chan, "expected chan after '<-' in channel type");
            const value = this.parseType();
            return { kind: "ChanType", direction: "receive", value, span: mergeSpans(start.span, value.span ?? chan.span) };
        }
        if (this.match(TokenKind.Star)) {
            const expr = this.parseTypeTerm();
            return { kind: "StarExpr", expr, span: mergeSpans(start.span, expr.span) };
        }
        if (this.match(TokenKind.LBracket)) {
            return this.parseArrayTypeAfterOpen(this.previous());
        }
        if (this.match(TokenKind.Map)) {
            this.expect(TokenKind.LBracket, "expected '[' after map");
            const key = this.parseType();
            this.expect(TokenKind.RBracket, "expected ']' after map key type");
            const value = this.parseType();
            return { kind: "MapType", key, value, span: mergeSpans(start.span, value.span) };
        }
        if (this.match(TokenKind.Chan)) {
            const sendOnly = this.match(TokenKind.Arrow);
            const value = this.parseType();
            return {
                kind: "ChanType",
                direction: sendOnly ? "send" : "both",
                value,
                span: mergeSpans(start.span, value.span)
            };
        }
        if (this.match(TokenKind.Struct))
            return this.parseStructType(start.span);
        if (this.match(TokenKind.Interface))
            return this.parseInterfaceType(start.span);
        if (this.at(TokenKind.Func))
            return this.parseFuncType();
        if (this.match(TokenKind.LParen)) {
            const type = this.parseType();
            const close = this.expect(TokenKind.RParen, "expected ')' after type");
            return { kind: "ParenExpr", expr: type, span: mergeSpans(start.span, close.span) };
        }
        return this.parseTypeName();
    }
    parseTypeName() {
        let expression = this.parseIdent("expected type name");
        while (this.match(TokenKind.Dot)) {
            const selector = this.parseIdent("expected selector in qualified type");
            expression = {
                kind: "SelectorExpr",
                object: expression,
                selector,
                span: mergeSpans(expression.span, selector.span)
            };
        }
        if (this.match(TokenKind.LBracket)) {
            const { indices, close } = this.parseTypeArgumentList();
            expression = indices.length === 1 ? {
                kind: "IndexExpr",
                object: expression,
                index: indices[0],
                span: mergeSpans(expression.span, close.span)
            } : {
                kind: "IndexListExpr",
                object: expression,
                indices,
                span: mergeSpans(expression.span, close.span)
            };
        }
        return expression;
    }
    parseArrayTypeAfterOpen(open, parsedLength) {
        let length = parsedLength;
        let inferredLength = false;
        if (!length) {
            this.withExpressionLevel(() => {
                if (this.match(TokenKind.Ellipsis)) {
                    inferredLength = true;
                }
                else if (!this.at(TokenKind.RBracket)) {
                    length = this.parseRhsExpression();
                }
            });
        }
        if (this.match(TokenKind.Comma)) {
            this.error("unexpected comma; expecting ]", this.previous().span);
        }
        this.expect(TokenKind.RBracket, "expected ']' in array or slice type");
        const element = this.parseTypeTerm();
        return {
            kind: "ArrayType",
            ...(length ? { length } : {}),
            element,
            inferredLength,
            span: mergeSpans(open.span, element.span)
        };
    }
    parseTypeArgumentList() {
        const indices = [];
        while (!this.at(TokenKind.RBracket) && !this.at(TokenKind.EOF)) {
            indices.push(this.withExpressionLevel(() => this.parseType()));
            if (!this.match(TokenKind.Comma))
                break;
        }
        return {
            indices: indices.length > 0 ? indices : [badExpr(this.peek().span)],
            close: this.expect(TokenKind.RBracket, "expected ']' after type arguments")
        };
    }
    parseStructType(start) {
        const fields = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
        return { kind: "StructType", fields, span: mergeSpans(start, fields.span) };
    }
    parseInterfaceType(start) {
        const methods = this.parseFieldList(TokenKind.LBrace, TokenKind.RBrace);
        return { kind: "InterfaceType", methods, span: mergeSpans(start, methods.span) };
    }
    parseSignature(start, typeParams) {
        const params = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        let results;
        if (this.at(TokenKind.LParen)) {
            results = this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        }
        else if (this.startsType()) {
            const type = this.parseType();
            results = {
                kind: "FieldList",
                fields: [{ kind: "Field", names: [], type, span: mergeSpans(type.span, type.span) }],
                span: mergeSpans(type.span, type.span)
            };
        }
        return {
            kind: "FuncType",
            ...(typeParams ? { typeParams } : {}),
            params,
            ...(results ? { results } : {}),
            span: mergeSpans(start, results?.span ?? params.span)
        };
    }
    parseFuncType() {
        const start = this.expect(TokenKind.Func, "expected func");
        if (this.at(TokenKind.LBracket)) {
            const typeParams = this.parseTypeParamList();
            this.error("function type must have no type parameters", typeParams.span);
        }
        return this.parseSignature(start.span);
    }
    parseFuncTypeOrLit() {
        const type = this.parseFuncType();
        if (!this.at(TokenKind.LBrace)) {
            return type;
        }
        const body = this.withExpressionLevel(() => this.parseBlock());
        return { kind: "FuncLit", type, body, span: mergeSpans(type.span, body.span) };
    }
    parseTypeParamList() {
        return this.parseFieldList(TokenKind.LBracket, TokenKind.RBracket);
    }
    parseTypeParameterListAfterOpen(open, firstName, firstType) {
        const fields = this.parseParameterList(TokenKind.RBracket, firstName, firstType, false);
        const close = this.expect(TokenKind.RBracket, "expected ']'");
        if (fields.length === 0)
            this.error("empty type parameter list", close.span);
        return { kind: "FieldList", fields, span: mergeSpans(open.span, close.span) };
    }
    parseParameterList(closing, firstName, firstType, allowEllipsis = false) {
        const typeParams = closing === TokenKind.RBracket;
        const params = [];
        let name0 = firstName;
        let type0 = firstType;
        let named = 0;
        let typed = 0;
        while (name0 || (!this.at(closing) && !this.at(TokenKind.EOF))) {
            let param;
            if (type0) {
                param = { ...(name0 ? { name: name0 } : {}), type: typeParams ? this.embeddedElem(type0) : type0 };
            }
            else {
                param = this.parseParamDecl(name0, typeParams);
            }
            name0 = undefined;
            type0 = undefined;
            if (param.name || param.type) {
                params.push(param);
                if (param.name && param.type)
                    named += 1;
                if (param.type)
                    typed += 1;
            }
            if (this.at(closing) || this.at(TokenKind.EOF))
                break;
            if (!this.match(TokenKind.Comma))
                break;
        }
        if (params.length === 0)
            return [];
        if (named === 0) {
            for (const param of params) {
                if (param.name && !param.type) {
                    param.type = param.name;
                    delete param.name;
                }
            }
            if (typeParams) {
                const message = named === typed ? "missing type constraint" : `missing type parameter name${params.length === 1 ? " or invalid array length" : ""}`;
                this.error(message, this.peek().span);
            }
        }
        else if (named !== params.length) {
            let type;
            let errorSpan;
            for (let index = params.length - 1; index >= 0; index -= 1) {
                const param = params[index];
                if (param.type) {
                    type = param.type;
                    if (!param.name) {
                        errorSpan = param.type.span;
                        param.name = ident("_", errorSpan);
                    }
                }
                else if (type) {
                    param.type = type;
                }
                else {
                    errorSpan = param.name?.span;
                    param.type = badExpr(errorSpan);
                }
            }
            if (errorSpan) {
                const message = named === typed
                    ? (typeParams ? "missing type constraint" : "missing parameter type")
                    : (typeParams ? `missing type parameter name${params.length === 1 ? " or invalid array length" : ""}` : "missing parameter name");
                this.error(message, errorSpan);
            }
        }
        let reportedEllipsis = false;
        for (let index = 0; index < params.length; index += 1) {
            const param = params[index];
            if (param.type?.kind === "Ellipsis" && (!allowEllipsis || index + 1 < params.length)) {
                if (!reportedEllipsis) {
                    this.error(allowEllipsis ? "can only use ... with final parameter" : "invalid use of ...", param.type.span);
                    reportedEllipsis = true;
                }
                param.type = badExpr(param.type.span);
            }
        }
        if (named === 0) {
            return params.map((param) => ({
                kind: "Field",
                names: [],
                type: param.type ?? badExpr(param.name?.span),
                span: mergeSpans(param.type?.span ?? param.name?.span, param.type?.span ?? param.name?.span)
            }));
        }
        const fields = [];
        let names = [];
        let currentType;
        const flush = () => {
            if (!currentType || names.length === 0)
                return;
            fields.push({
                kind: "Field",
                names,
                type: currentType,
                span: mergeSpans(names[0]?.span, currentType.span)
            });
            names = [];
        };
        for (const param of params) {
            if (param.type !== currentType) {
                flush();
                currentType = param.type;
            }
            names.push(param.name ?? ident("_", param.type?.span));
        }
        flush();
        return fields;
    }
    parseParamDecl(name0, typeSetsOK) {
        let name = name0;
        let type;
        if (name || isIdentifierLike(this.peek().kind)) {
            if (!name)
                name = this.parseIdent("expected parameter name or type");
            if (this.startsType() || this.at(TokenKind.LParen)) {
                type = this.parseType();
            }
            else if (this.match(TokenKind.Ellipsis)) {
                const dots = this.previous();
                type = { kind: "Ellipsis", element: this.parseType(), span: dots.span };
            }
            else if (this.match(TokenKind.Dot)) {
                const selector = this.parseIdent("expected selector in qualified type");
                type = {
                    kind: "SelectorExpr",
                    object: name,
                    selector,
                    span: mergeSpans(name.span, selector.span)
                };
                name = undefined;
            }
            else if (typeSetsOK && this.at(TokenKind.Or)) {
                type = this.embeddedElem(name);
                name = undefined;
            }
        }
        else if (this.startsType() || this.at(TokenKind.LParen)) {
            type = this.parseType();
        }
        else if (this.match(TokenKind.Ellipsis)) {
            const dots = this.previous();
            type = { kind: "Ellipsis", element: this.parseType(), span: dots.span };
        }
        else {
            this.error(`expected parameter name or type, found ${this.peek().lexeme || this.peek().kind}`, this.peek().span);
            this.advance();
            return { type: badExpr(this.previous().span) };
        }
        if (typeSetsOK && type && this.at(TokenKind.Or)) {
            type = this.embeddedElem(type);
        }
        return { ...(name ? { name } : {}), ...(type ? { type } : {}) };
    }
    embeddedElem(initial) {
        let expr = initial;
        while (this.match(TokenKind.Or)) {
            const operator = this.previous();
            const right = this.parseTypeTerm();
            expr = {
                kind: "BinaryExpr",
                left: expr,
                op: operator.kind,
                right,
                span: mergeSpans(expr.span, right.span)
            };
        }
        return expr;
    }
    parseFieldList(open, close) {
        const start = this.expect(open, `expected '${tokenDisplay(open)}'`);
        const fields = [];
        while (!this.at(close) && !this.at(TokenKind.EOF)) {
            this.skipSemis();
            if (this.at(close))
                break;
            fields.push(this.parseField(close));
            if (!this.match(TokenKind.Comma))
                this.consumeSemi();
        }
        const end = this.expect(close, `expected '${tokenDisplay(close)}'`);
        return { kind: "FieldList", fields, span: mergeSpans(start.span, end.span) };
    }
    parseField(close) {
        const start = this.peek();
        if (isIdentifierLike(this.peek().kind) && this.peek(1).kind === TokenKind.LParen) {
            const name = this.parseIdent("expected method name");
            const type = this.parseSignature(name.span ?? start.span);
            return {
                kind: "Field",
                names: [name],
                type,
                span: mergeSpans(name.span, type.span)
            };
        }
        if (isIdentifierLike(this.peek().kind) && this.fieldHasExplicitNames(close)) {
            const names = this.parseIdentList();
            const type = this.match(TokenKind.Ellipsis)
                ? { kind: "Ellipsis", element: this.parseType(), span: start.span }
                : this.parseType();
            const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
            return {
                kind: "Field",
                names,
                type,
                ...(tag ? { tag } : {}),
                span: mergeSpans(names[0]?.span, tag?.span ?? type.span)
            };
        }
        const type = this.match(TokenKind.Ellipsis)
            ? { kind: "Ellipsis", element: this.parseType(), span: start.span }
            : this.parseType();
        const tag = this.at(TokenKind.StringLiteral) ? this.expectBasicLit(TokenKind.StringLiteral, "expected struct tag") : undefined;
        return { kind: "Field", names: [], type, ...(tag ? { tag } : {}), span: mergeSpans(type.span, tag?.span ?? type.span) };
    }
    fieldHasExplicitNames(close) {
        let offset = 0;
        if (!isIdentifierLike(this.peek(offset).kind))
            return false;
        offset += 1;
        while (this.peek(offset).kind === TokenKind.Comma && isIdentifierLike(this.peek(offset + 1).kind)) {
            offset += 2;
        }
        const afterNames = this.peek(offset).kind;
        if (offset === 1 && afterNames === TokenKind.LBracket) {
            const afterBracket = this.kindAfterBalancedBrackets(offset);
            if (afterBracket === close ||
                afterBracket === TokenKind.Semicolon ||
                afterBracket === TokenKind.RBrace ||
                afterBracket === TokenKind.RParen ||
                afterBracket === TokenKind.RBracket ||
                afterBracket === TokenKind.StringLiteral ||
                afterBracket === TokenKind.Comma ||
                afterBracket === TokenKind.EOF) {
                return false;
            }
        }
        if (afterNames === close || afterNames === TokenKind.Semicolon || afterNames === TokenKind.RBrace || afterNames === TokenKind.RParen || afterNames === TokenKind.RBracket) {
            return false;
        }
        return this.startsType(afterNames) || afterNames === TokenKind.Ellipsis;
    }
    kindAfterBalancedBrackets(openOffset) {
        let depth = 0;
        for (let offset = openOffset;; offset += 1) {
            const kind = this.peek(offset).kind;
            if (kind === TokenKind.EOF)
                return TokenKind.EOF;
            if (kind === TokenKind.LBracket)
                depth += 1;
            if (kind === TokenKind.RBracket) {
                depth -= 1;
                if (depth === 0)
                    return this.peek(offset + 1).kind;
            }
        }
    }
    parseIdentList() {
        const names = [this.parseIdent("expected identifier")];
        while (this.match(TokenKind.Comma)) {
            names.push(this.parseIdent("expected identifier after ','"));
        }
        return names;
    }
    parseIdent(message) {
        const token = this.peek();
        if (isIdentifierLike(token.kind)) {
            this.advance();
            return ident(token.lexeme, token.span);
        }
        this.error(message, token.span);
        this.advance();
        return ident("<missing>", token.span);
    }
    expectCellAddress(message) {
        const token = this.peek();
        if (token.kind === TokenKind.CellAddress || token.kind === TokenKind.Identifier) {
            this.advance();
            const address = parseCellAddress(token.lexeme);
            if (address)
                return { token, address };
            this.error(message, token.span);
            return { token };
        }
        this.error(message, token.span);
        this.advance();
        return { token };
    }
    expectBasicLit(kind, message) {
        const token = this.peek();
        const end = this.end(token);
        if (this.match(kind)) {
            return { kind: "BasicLit", token: kind, value: token.lexeme, span: mergeSpans(token.span, end) };
        }
        this.error(message, token.span);
        return { kind: "BasicLit", token: kind, value: "", span: token.span };
    }
    looksLikeReceiver() {
        let depth = 0;
        for (let offset = 0; offset < 32; offset += 1) {
            const kind = this.peek(offset).kind;
            if (kind === TokenKind.LParen)
                depth += 1;
            if (kind === TokenKind.RParen) {
                depth -= 1;
                if (depth === 0)
                    return this.peek(offset + 1).kind === TokenKind.Identifier;
            }
            if (kind === TokenKind.EOF || kind === TokenKind.LBrace)
                return false;
        }
        return false;
    }
    looksLikeRangeClause() {
        let parens = 0;
        let brackets = 0;
        let braces = 0;
        for (let offset = 0; offset < 64; offset += 1) {
            const kind = this.peek(offset).kind;
            if (kind === TokenKind.LParen)
                parens += 1;
            else if (kind === TokenKind.RParen && parens > 0)
                parens -= 1;
            else if (kind === TokenKind.LBracket)
                brackets += 1;
            else if (kind === TokenKind.RBracket && brackets > 0)
                brackets -= 1;
            else if (kind === TokenKind.LBrace) {
                if (parens === 0 && brackets === 0 && braces === 0)
                    return false;
                braces += 1;
            }
            else if (kind === TokenKind.RBrace && braces > 0)
                braces -= 1;
            if (kind === TokenKind.Range)
                return true;
            if ((kind === TokenKind.Semicolon || kind === TokenKind.EOF) && parens === 0 && brackets === 0 && braces === 0)
                return false;
        }
        return false;
    }
    startsType(kind = this.peek().kind) {
        return isIdentifierLike(kind) ||
            kind === TokenKind.Star ||
            kind === TokenKind.Tilde ||
            kind === TokenKind.LBracket ||
            kind === TokenKind.Map ||
            kind === TokenKind.Chan ||
            kind === TokenKind.Arrow ||
            kind === TokenKind.Struct ||
            kind === TokenKind.Interface ||
            kind === TokenKind.Func ||
            kind === TokenKind.LParen;
    }
    withBareIdentifierComposites(enabled, fn) {
        const previous = this.allowBareIdentifierComposite;
        this.allowBareIdentifierComposite = enabled;
        try {
            return fn();
        }
        finally {
            this.allowBareIdentifierComposite = previous;
        }
    }
    withControlClause(fn) {
        const previous = this.exprLev;
        this.exprLev = -1;
        try {
            return fn();
        }
        finally {
            this.exprLev = previous;
        }
    }
    withExpressionLevel(fn) {
        const previous = this.exprLev;
        this.exprLev += 1;
        try {
            return fn();
        }
        finally {
            this.exprLev = previous;
        }
    }
    withRhs(enabled, fn) {
        const previous = this.inRhs;
        this.inRhs = enabled;
        try {
            return fn();
        }
        finally {
            this.inRhs = previous;
        }
    }
    withSpreadsheetRanges(enabled, fn) {
        const previous = this.allowSpreadsheetRanges;
        this.allowSpreadsheetRanges = enabled;
        try {
            return fn();
        }
        finally {
            this.allowSpreadsheetRanges = previous;
        }
    }
    printTrace(...args) {
        const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . ";
        const prefix = dots.slice(0, Math.min(dots.length, 2 * this.indent));
        globalThis.console?.log(`${this.peek().span?.line ?? 0}:${this.peek().span?.column ?? 0}: ${prefix}${args.join(" ")}`);
    }
    init(_file, _text, _mode) {
    }
    next0() {
        this.advance();
    }
    next() {
        this.advance();
    }
    lineFor(pos) {
        return pos?.line ?? 0;
    }
    atComma(_context, _follow) {
        if (this.at(TokenKind.Comma)) {
            return true;
        }
        if (this.at(TokenKind.Semicolon)) {
            this.error("missing ',' before newline in composite literal", this.peek().span);
            return true;
        }
        return false;
    }
    errorExpected(pos, msg) {
        this.error(`expected ${msg}`, pos);
    }
    expect2(kind) {
        const token = this.expect(kind, `expected ${tokenDisplay(kind)}`);
        return [token.span, token.lexeme];
    }
    expectClosing(kind, context) {
        return this.expect(kind, `expected ${tokenDisplay(kind)} closing ${context}`).span;
    }
    expectSemi() {
        switch (this.peek().kind) {
            case TokenKind.Semicolon:
            case TokenKind.RParen:
            case TokenKind.RBrace:
                this.consumeSemi();
                return;
            default:
                this.error("expected ';'", this.peek().span);
        }
    }
    consumeComment() {
    }
    consumeCommentGroup() {
    }
    tokPrec() {
        let kind = this.peek().kind;
        if (this.inRhs && kind === TokenKind.Assign) {
            kind = TokenKind.Equal;
        }
        return [kind, binaryPrecedence(kind)];
    }
    makeExpr(statement, want) {
        return this.statementExpression(statement, want);
    }
    parseStmt() {
        return this.parseStatement();
    }
    parseStmtList() {
        const list = [];
        while (!this.atAny(TokenKind.Case, TokenKind.Default, TokenKind.RBrace, TokenKind.EOF)) {
            list.push(this.parseStatement());
            this.consumeSemi();
        }
        return list;
    }
    parseBlockStmt() {
        return this.parseBlock();
    }
    parseBody() {
        return this.parseBlock();
    }
    parseReturnStmt() {
        if (!this.at(TokenKind.Return)) {
            this.error("expected return", this.peek().span);
        }
        return this.parseStatement();
    }
    parseBranchStmt() {
        return this.parseStatement();
    }
    parseDeferStmt() {
        return this.parseStatement();
    }
    parseGoStmtAlias() {
        return this.parseGoStmt();
    }
    parseExpr() {
        return this.parseExpression();
    }
    parseRhs() {
        return this.parseRhsExpression();
    }
    parseBinaryExpr(x, prec1 = 1) {
        return this.parseBinaryExpression(x ?? this.parseUnary(), prec1);
    }
    parseUnaryExpr() {
        return this.parseUnary();
    }
    parsePrimaryExpr(x) {
        return x ? this.parsePrimaryFrom(x) : this.parsePrimary();
    }
    parseCallExpr(fun) {
        return this.parsePrimaryFrom(fun);
    }
    parseCallOrConversion(fun) {
        return this.parsePrimaryFrom(fun);
    }
    parseSelector(x) {
        if (!this.at(TokenKind.Dot)) {
            return x;
        }
        const dot = this.advance();
        const selector = this.parseIdent("expected selector");
        return { kind: "SelectorExpr", object: x, selector, span: mergeSpans(x.span, selector.span ?? dot.span) };
    }
    parseTypeAssertion(x) {
        return this.parsePrimaryFrom(x);
    }
    parseIndexOrSliceOrInstance(x) {
        return this.parsePrimaryFrom(x);
    }
    parseTypeInstance(x) {
        return this.parsePrimaryFrom(x);
    }
    parseElement() {
        return this.parseCompositeLiteralElement();
    }
    parseElementList() {
        const list = [];
        while (!this.atAny(TokenKind.RBrace, TokenKind.EOF)) {
            list.push(this.parseElement());
            if (!this.match(TokenKind.Comma)) {
                break;
            }
        }
        return list;
    }
    parseLiteralValue(type) {
        const open = this.expect(TokenKind.LBrace, "expected literal value");
        const elements = this.parseElementList();
        const close = this.expect(TokenKind.RBrace, "expected '}' after literal value");
        return { kind: "CompositeLit", ...(type ? { type } : {}), elements, span: mergeSpans(open.span, close.span) };
    }
    parseValue() {
        return this.parseRhsExpression();
    }
    parseExprList() {
        return this.parseExpressionList(false);
    }
    parseList(inRhs) {
        return this.parseExpressionList(inRhs);
    }
    parsePointerType() {
        const star = this.expect(TokenKind.Star, "expected '*'");
        const expr = this.parseType();
        return { kind: "StarExpr", expr, span: mergeSpans(star.span, expr.span) };
    }
    parseDotsType() {
        const dots = this.expect(TokenKind.Ellipsis, "expected '...'");
        const element = this.parseType();
        return { kind: "Ellipsis", element, span: mergeSpans(dots.span, element.span) };
    }
    parseArrayType() {
        const open = this.expect(TokenKind.LBracket, "expected '['");
        return this.parseArrayTypeAfterOpen(open);
    }
    parseArrayFieldOrTypeInstance(name) {
        const open = this.expect(TokenKind.LBracket, "expected '['");
        const type = this.parseArrayTypeAfterOpen(open, name);
        return [name, type];
    }
    parseMapType() {
        const start = this.expect(TokenKind.Map, "expected map");
        this.expect(TokenKind.LBracket, "expected '[' after map");
        const key = this.parseType();
        this.expect(TokenKind.RBracket, "expected ']' after map key");
        const value = this.parseType();
        return { kind: "MapType", key, value, span: mergeSpans(start.span, value.span) };
    }
    parseChanType() {
        return this.parseTypeTerm();
    }
    parseQualifiedIdent(name) {
        const object = name ?? this.parseIdent("expected identifier");
        if (this.match(TokenKind.Dot)) {
            const selector = this.parseIdent("expected selector");
            return { kind: "SelectorExpr", object, selector, span: mergeSpans(object.span, selector.span) };
        }
        return object;
    }
    tryIdentOrType() {
        if (isIdentifierLike(this.peek().kind)) {
            return this.parseTypeName();
        }
        if (this.startsType()) {
            return this.parseType();
        }
        return undefined;
    }
    embeddedTerm() {
        return this.parseTypeTerm();
    }
    parseFieldDecl() {
        return this.parseField(TokenKind.RBrace);
    }
    parseMethodSpec() {
        return this.parseField(TokenKind.RBrace);
    }
    parseParameters(acceptTParams = false) {
        if (this.at(TokenKind.LParen)) {
            return this.parseFieldList(TokenKind.LParen, TokenKind.RParen);
        }
        if (acceptTParams && this.at(TokenKind.LBracket)) {
            return this.parseFieldList(TokenKind.LBracket, TokenKind.RBracket);
        }
        return undefined;
    }
    parseTypeParameters() {
        return this.at(TokenKind.LBracket) ? this.parseTypeParamList() : undefined;
    }
    consumeSemi() {
        this.match(TokenKind.Semicolon);
    }
    skipSemis() {
        while (this.match(TokenKind.Semicolon)) {
            // consume all empty statements between declarations/statements
        }
    }
    match(kind) {
        if (!this.at(kind))
            return false;
        this.advance();
        return true;
    }
    expect(kind, message) {
        if (this.at(kind))
            return this.advance();
        const token = this.peek();
        this.error(message, token.span);
        return token;
    }
    expectAny(kinds, message) {
        if (kinds.includes(this.peek().kind))
            return this.advance();
        const token = this.peek();
        this.error(message, token.span);
        return token;
    }
    at(kind) {
        return this.peek().kind === kind;
    }
    atAny(...kinds) {
        const current = this.peek().kind;
        return kinds.includes(current);
    }
    advance() {
        const token = this.peek();
        if (!this.at(TokenKind.EOF))
            this.index += 1;
        return token;
    }
    previous() {
        return this.tokens[Math.max(0, this.index - 1)] ?? this.peek();
    }
    peek(ahead = 0) {
        return this.tokens[this.index + ahead] ?? this.tokens[this.tokens.length - 1] ?? eofToken();
    }
    currentSpan() {
        return this.peek().span;
    }
    end(token = this.peek()) {
        if (!token.span)
            return undefined;
        return {
            ...token.span,
            offset: token.span.offset + token.span.length,
            length: 0,
            column: token.span.column + token.span.length
        };
    }
    error(message, span, code = "GOJR_PARSE_FRONT001") {
        this.diagnostics.push({
            filename: diagnosticFilename(span, this.filename),
            code,
            severity: "error",
            message,
            ...(span ? { span } : {})
        });
    }
}
function binaryPrecedence(kind) {
    switch (kind) {
        case TokenKind.OrOr: return 1;
        case TokenKind.AndAnd: return 2;
        case TokenKind.Equal:
        case TokenKind.NotEqual:
        case TokenKind.Less:
        case TokenKind.LessEqual:
        case TokenKind.Greater:
        case TokenKind.GreaterEqual:
            return 3;
        case TokenKind.Plus:
        case TokenKind.Minus:
        case TokenKind.Or:
        case TokenKind.Caret:
            return 4;
        case TokenKind.Star:
        case TokenKind.Slash:
        case TokenKind.Percent:
        case TokenKind.Shl:
        case TokenKind.Shr:
        case TokenKind.Amp:
        case TokenKind.BitClear:
            return 5;
        default:
            return 0;
    }
}
function extractName(expr, force) {
    if (expr.kind === "Ident") {
        return { name: expr };
    }
    if (expr.kind === "BinaryExpr") {
        if (expr.op === TokenKind.Star && expr.left.kind === "Ident" && (force || isTypeElem(expr.right))) {
            return {
                name: expr.left,
                type: {
                    kind: "StarExpr",
                    expr: expr.right,
                    span: mergeSpans(expr.left.span, expr.right.span)
                }
            };
        }
        if (expr.op === TokenKind.Or) {
            const split = extractName(expr.left, force || isTypeElem(expr.right));
            if (split.name && split.type) {
                return {
                    name: split.name,
                    type: {
                        kind: "BinaryExpr",
                        left: split.type,
                        op: TokenKind.Or,
                        right: expr.right,
                        span: mergeSpans(split.type.span, expr.right.span)
                    }
                };
            }
        }
    }
    if (expr.kind === "CallExpr" && expr.fun.kind === "Ident" && expr.args.length === 1 && !expr.ellipsis && (force || isTypeElem(expr.args[0]))) {
        return {
            name: expr.fun,
            type: {
                kind: "ParenExpr",
                expr: expr.args[0],
                span: mergeSpans(expr.fun.span, expr.span)
            }
        };
    }
    return { type: expr };
}
function isTypeElem(expr) {
    switch (expr.kind) {
        case "ArrayType":
        case "StructType":
        case "FuncType":
        case "InterfaceType":
        case "MapType":
        case "ChanType":
            return true;
        case "BinaryExpr":
            return isTypeElem(expr.left) || isTypeElem(expr.right);
        case "UnaryExpr":
            return expr.op === TokenKind.Tilde;
        case "ParenExpr":
            return isTypeElem(expr.expr);
        default:
            return false;
    }
}
function badExpr(span) {
    return { kind: "BadExpr", ...(span ? { span } : {}) };
}
function mergeSpans(start, end) {
    if (!start && !end)
        return eofToken().span;
    if (!start)
        return end ?? eofToken().span;
    if (!end)
        return start;
    const endOffset = end.offset + end.length;
    return {
        filename: start.filename,
        offset: start.offset,
        length: Math.max(0, endOffset - start.offset),
        line: start.line,
        column: start.column
    };
}
function tokenDisplay(kind) {
    switch (kind) {
        case TokenKind.LParen:
            return "(";
        case TokenKind.RParen:
            return ")";
        case TokenKind.LBrace:
            return "{";
        case TokenKind.RBrace:
            return "}";
        case TokenKind.LBracket:
            return "[";
        case TokenKind.RBracket:
            return "]";
        default:
            return kind;
    }
}
function eofToken() {
    return {
        kind: TokenKind.EOF,
        lexeme: "",
        span: { filename: REPL_FILENAME, offset: 0, length: 0, line: 1, column: 1 }
    };
}
function hasBytes(value) {
    return typeof value === "object" && value !== null && "Bytes" in value && typeof value.Bytes === "function";
}
function hasRead(value) {
    return typeof value === "object" && value !== null && "Read" in value && typeof value.Read === "function";
}
function diagnosticAsError(diagnostic) {
    if (!diagnostic) {
        return undefined;
    }
    return new Error(`${diagnostic.filename}:${diagnostic.span?.line ?? 0}:${diagnostic.span?.column ?? 0}: ${diagnostic.message}`);
}
function hostReadFile() {
    const host = globalThis;
    return host.__gojrReadFile;
}
function hostReadDir() {
    const host = globalThis;
    return host.__gojrReadDir;
}
