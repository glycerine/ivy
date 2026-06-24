// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { atPos, debug, nopos, tracePos, environment, registerCheckerMethod } from "./check.js";
import { Const, Func, LocalVar, NewConst, NewTypeName, TypeName, Var, newVar, packagePrefix } from "./object.js";
import { Typ } from "./universe.js";
import { Invalid } from "./basic.js";
import { Alias, asNamed, unalias } from "./alias.js";
import { Named } from "./named.js";
import { Interface } from "./interface.js";
import { TypeParam } from "./typeparam.js";
import { Signature } from "./signature.js";
import { bindTParams } from "./typelists.js";
import { objset } from "./objset.js";
import { Struct } from "./struct.js";
import { assert } from "./util.js";
import { cmpPos } from "./check.js";
import { go1_18, go1_23, go1_9 } from "./version.js";
import { operand } from "./operand.js";
import { fieldListNumFields, funcDeclBody, funcDeclName, funcDeclRecv, funcDeclType, genDeclSpecs, genDeclTok, identName, nodeEnd, nodeKind, nodePos, specName, specNames, specType, specValues } from "./astcompat.js";
import { TypeString } from "./typestring.js";
import { TokenKind } from "../../front/token.js";
registerCheckerMethod("declare", function declare(scope, id, obj, pos) {
    // spec: "The blank identifier, represented by the underscore
    // character _, may be used in a declaration like any other
    // identifier but the declaration does not introduce a new
    // binding."
    if (obj.Name() !== "_") {
        const alt = scope.Insert(obj);
        if (alt !== null) {
            const err = this.newError("DuplicateDecl");
            err.addf(obj, "%s redeclared in this block", obj.Name());
            err.addAltDecl(alt);
            err.report();
            return;
        }
        obj.setScopePos(pos);
    }
    if (id !== null && id !== undefined) {
        this.recordDef(id, obj);
    }
});
// pathString returns a string of the form a->b-> ... ->g for a path [a, b, ... g].
export function pathString(path) {
    let s = "";
    for (let i = 0; i < path.length; i++) {
        const p = path[i];
        if (i > 0) {
            s += "->";
        }
        s += p.Name();
    }
    return s;
}
// objDecl type-checks the declaration of obj in its respective (file) environment.
registerCheckerMethod("objDecl", function objDecl(obj) {
    if (tracePos) {
        this.pushPos(new atPos(obj.Pos()));
    }
    try {
        if (this.conf._Trace && obj.Type() === null) {
            if (this.indent === 0) {
                console.log(); // empty line between top-level objects for readability
            }
            this.trace(obj.Pos(), "-- checking %s (objPath = %s)", obj, pathString(this.objPath.filter((x) => x !== null)));
            this.indent++;
        }
        // Checking the declaration of an object means determining its type
        // (and also its value for constants). An object (and thus its type)
        // may be in 1 of 3 states:
        //
        // - not in Checker.objPathIdx and type == nil : type is not yet known (white)
        // -     in Checker.objPathIdx                 : type is pending       (grey)
        // - not in Checker.objPathIdx and type != nil : type is known         (black)
        if (this.objPathIdx?.has(obj)) {
            if (obj instanceof Const || obj instanceof Var) {
                if (!this.validCycle(obj) || obj.Type() === null) {
                    obj.setType(Typ[Invalid]);
                }
            }
            else if (obj instanceof TypeName) {
                if (!this.validCycle(obj)) {
                    obj.setType(Typ[Invalid]);
                }
            }
            else if (obj instanceof Func) {
                if (!this.validCycle(obj)) {
                    // Don't set type to Typ[Invalid]; plenty of code asserts that
                    // functions have a *Signature type. Instead, leave the type
                    // as an empty signature, which makes it impossible to
                    // initialize a variable with the function.
                }
            }
            else {
                throw new Error("unreachable");
            }
            assert(obj.Type() !== null);
            return;
        }
        if (obj.Type() !== null) { // black, meaning it's already type-checked
            return;
        }
        // white, meaning it must be type-checked
        this.push(obj); // mark as grey
        try {
            const d = this.objMap.get(obj);
            assert(d !== undefined);
            // save/restore current environment and set up object environment
            const saved = Object.assign(new environment(), this);
            this.scope = d.file;
            this.version = d.version;
            try {
                if (obj instanceof Const) {
                    this.decl = d; // new package-level const decl
                    this.constDecl(obj, d.vtyp, d.init, d.inherited);
                }
                else if (obj instanceof Var) {
                    this.decl = d; // new package-level var decl
                    this.varDecl(obj, d.lhs, d.vtyp, d.init);
                }
                else if (obj instanceof TypeName) {
                    // invalid recursive types are detected via path
                    this.typeDecl(obj, d.tdecl);
                    this.collectMethods(obj); // methods can only be added to top-level types
                }
                else if (obj instanceof Func) {
                    // functions may be recursive - no need to track dependencies
                    this.funcDecl(obj, d);
                }
                else {
                    throw new Error("unreachable");
                }
            }
            finally {
                this.decl = saved.decl;
                this.scope = saved.scope;
                this.version = saved.version;
                this.iota = saved.iota;
                this.errpos = saved.errpos;
                this.inTParamList = saved.inTParamList;
                this.sig = saved.sig;
                this.isPanic = saved.isPanic;
                this.hasLabel = saved.hasLabel;
                this.hasCallOrRecv = saved.hasCallOrRecv;
                this.exprPos = saved.exprPos;
            }
        }
        finally {
            this.pop();
        }
    }
    finally {
        if (tracePos) {
            this.popPos();
        }
    }
});
// validCycle checks if the cycle starting with obj is valid and
// reports an error if it is not.
registerCheckerMethod("validCycle", function validCycle(obj) {
    const start = this.objPathIdx?.get(obj);
    assert(start !== undefined);
    const cycle = this.objPath.slice(start).filter((x) => x !== null);
    let tparCycle = false; // if set, the cycle is through a type parameter list
    let nval = 0; // number of (constant or variable) values in the cycle
    let ndef = 0; // number of type definitions in the cycle
    for (const obj of cycle) {
        if (obj instanceof Const || obj instanceof Var) {
            nval++;
        }
        else if (obj instanceof TypeName) {
            // If we reach a generic type that is part of a cycle
            // and we are in a type parameter list, we have a cycle
            // through a type parameter list.
            if (this.inTParamList && isGeneric(obj.typ)) {
                tparCycle = true;
                break;
            }
            if (!obj.IsAlias()) {
                ndef++;
            }
        }
        else if (obj instanceof Func) {
            // ignored for now
        }
        else {
            throw new Error("unreachable");
        }
    }
    if (this.conf._Trace) {
        this.trace(obj.Pos(), "## cycle detected: objPath = %s->%s (len = %d)", pathString(cycle), obj.Name(), cycle.length);
    }
    // Cycles through type parameter lists are ok (go.dev/issue/68162).
    if (tparCycle) {
        return true;
    }
    // A cycle involving only constants and variables is invalid but we
    // ignore them here because they are reported via the initialization
    // cycle check.
    if (nval === cycle.length) {
        return true;
    }
    // A cycle involving only types (and possibly functions) must have at least
    // one type definition to be permitted: If there is no type definition, we
    // have a sequence of alias type names which will expand ad infinitum.
    if (nval === 0 && ndef > 0) {
        return true;
    }
    this.cycleError(cycle, firstInSrc(cycle));
    return false;
});
// cycleError reports a declaration cycle starting with the object at cycle[start].
registerCheckerMethod("cycleError", function cycleError(cycle, start) {
    // name returns the (possibly qualified) object name.
    // This is needed because with generic types, cycles
    // may refer to imported types. See go.dev/issue/50788.
    const name = (obj) => {
        // include any type arguments in the reported error message
        const n = asNamed(obj.Type());
        if (n !== null && n.inst !== null) {
            return TypeString(n, null);
        }
        return packagePrefix(obj.Pkg(), null) + obj.Name();
    };
    // If obj is a type alias, mark it as valid (not broken) in order to avoid follow-on errors.
    let obj = cycle[start];
    const tname = obj instanceof TypeName ? obj : null;
    if (tname !== null) {
        const a = tname.Type();
        if (a instanceof Alias) {
            a.fromRHS = Typ[Invalid];
        }
    }
    // report a more concise error for self references
    if (cycle.length === 1) {
        if (tname !== null) {
            this.errorf(obj, "InvalidDeclCycle", "invalid recursive type: %s refers to itself", name(obj));
        }
        else {
            this.errorf(obj, "InvalidDeclCycle", "invalid cycle in declaration: %s refers to itself", name(obj));
        }
        return;
    }
    const err = this.newError("InvalidDeclCycle");
    if (tname !== null) {
        err.addf(obj, "invalid recursive type %s", name(obj));
    }
    else {
        err.addf(obj, "invalid cycle in declaration of %s", name(obj));
    }
    // "cycle[i] refers to cycle[j]" for (i,j) = (s,s+1), (s+1,s+2), ..., (n-1,0), (0,1), ..., (s-1,s) for len(cycle) = n, s = start.
    for (let i = 0; i < cycle.length; i++) {
        const next = cycle[(start + i + 1) % cycle.length];
        err.addf(obj, "%s refers to %s", name(obj), name(next));
        obj = next;
    }
    err.report();
});
// firstInSrc reports the index of the object with the "smallest"
// source position in path. path must not be empty.
export function firstInSrc(path) {
    let fst = 0;
    let pos = path[0].Pos();
    for (let i = 0; i < path.slice(1).length; i++) {
        const t = path.slice(1)[i];
        if (cmpPos(t.Pos(), pos) < 0) {
            fst = i + 1;
            pos = t.Pos();
        }
    }
    return fst;
}
export class importDecl {
    spec;
    kind = "importDecl";
    constructor(spec) {
        this.spec = spec;
    }
    node() { return this.spec; }
}
export class constDecl {
    spec;
    iota;
    typ;
    init;
    inherited;
    kind = "constDecl";
    constructor(spec, iota, typ, init, inherited) {
        this.spec = spec;
        this.iota = iota;
        this.typ = typ;
        this.init = init;
        this.inherited = inherited;
    }
    node() { return this.spec; }
}
export class varDecl {
    spec;
    kind = "varDecl";
    constructor(spec) {
        this.spec = spec;
    }
    node() { return this.spec; }
}
export class typeDecl {
    spec;
    kind = "typeDecl";
    constructor(spec) {
        this.spec = spec;
    }
    node() { return this.spec; }
}
export class funcDecl {
    decl;
    kind = "funcDecl";
    constructor(decl) {
        this.decl = decl;
    }
    node() { return this.decl; }
}
registerCheckerMethod("walkDecls", function walkDecls(decls, f) {
    for (const d of decls) {
        this.walkDecl(d, f);
    }
});
registerCheckerMethod("walkDecl", function walkDecl(d, f) {
    switch (nodeKind(d)) {
        case "BadDecl":
            // ignore
            break;
        case "GenDecl": {
            let last = null; // last ValueSpec with type or init exprs seen
            const specs = genDeclSpecs(d);
            for (let iota = 0; iota < specs.length; iota++) {
                const s = specs[iota];
                switch (nodeKind(s)) {
                    case "ImportSpec":
                        f(new importDecl(s));
                        break;
                    case "ValueSpec":
                        switch (genDeclTok(d)) {
                            case "CONST": {
                                // determine which initialization expressions to use
                                let inherited = true;
                                switch (true) {
                                    case specType(s) !== null && specType(s) !== undefined || specValues(s).length > 0:
                                        last = s;
                                        inherited = false;
                                        break;
                                    case last === null:
                                        last = {}; // make sure last exists
                                        inherited = false;
                                        break;
                                }
                                this.arityMatch(s, last);
                                f(new constDecl(s, iota, specType(last), specValues(last), inherited));
                                break;
                            }
                            case "VAR":
                                this.arityMatch(s, null);
                                f(new varDecl(s));
                                break;
                            default:
                                this.errorf(s, "InvalidSyntaxTree", "invalid token %s", genDeclTok(d));
                        }
                        break;
                    case "TypeSpec":
                        f(new typeDecl(s));
                        break;
                    default:
                        this.errorf(s, "InvalidSyntaxTree", "unknown ast.Spec node %T", s);
                }
            }
            break;
        }
        case "FuncDecl":
            f(new funcDecl(d));
            break;
        default:
            this.errorf(d, "InvalidSyntaxTree", "unknown ast.Decl node %T", d);
    }
});
registerCheckerMethod("constDecl", function constDeclMethod(obj, typ, init, inherited) {
    assert(obj.typ === null);
    // use the correct value of iota
    const savedIota = this.iota;
    const savedErrpos = this.errpos;
    this.iota = obj.val;
    this.errpos = null;
    try {
        // provide valid constant value under all circumstances
        obj.val = { unknown: true };
        // determine type, if any
        if (typ !== null && typ !== undefined) {
            const t = this.typ(typ);
            if (!isConstType(t)) {
                if (isValid(t.Underlying())) {
                    this.errorf(typ, "InvalidConstType", "invalid constant type %s", t);
                }
                obj.typ = Typ[Invalid];
                return;
            }
            obj.typ = t;
        }
        // check initialization
        const x = new operand();
        if (init !== null && init !== undefined) {
            if (inherited) {
                this.errpos = new atPos(obj.pos);
            }
            this.expr(null, x, init);
        }
        this.initConst(obj, x);
    }
    finally {
        this.iota = savedIota;
        this.errpos = savedErrpos;
    }
});
registerCheckerMethod("varDecl", function varDeclMethod(obj, lhs, typ, init) {
    assert(obj.typ === null);
    // determine type, if any
    if (typ !== null && typ !== undefined) {
        obj.typ = this.varType(typ);
    }
    // check initialization
    if (init === null || init === undefined) {
        if (typ === null || typ === undefined) {
            obj.typ = Typ[Invalid];
        }
        return;
    }
    if (lhs === null || lhs.length === 1) {
        assert(lhs === null || lhs[0] === obj);
        const x = new operand();
        this.expr(newTarget(obj.typ, obj.name), x, init);
        this.initVar(obj, x, "variable declaration");
        return;
    }
    if (debug) {
        if (!lhs.includes(obj)) {
            throw new Error("inconsistent lhs");
        }
    }
    if (typ !== null && typ !== undefined) {
        for (const lhsVar of lhs) {
            lhsVar.typ = obj.typ;
        }
    }
    this.initVars(lhs, [init], null);
});
// isImportedConstraint reports whether typ is an imported type constraint.
registerCheckerMethod("isImportedConstraint", function isImportedConstraint(typ) {
    const named = asNamed(typ);
    if (named === null || named.obj.pkg === this.pkg || named.obj.pkg === null) {
        return false;
    }
    const u = named.Underlying();
    return u instanceof Interface && !u.IsMethodSet();
});
function typeSpecTypeParams(spec) {
    const s = spec;
    return s?.TypeParams ?? s?.typeParams ?? null;
}
function typeSpecType(spec) {
    const s = spec;
    return s?.Type ?? s?.type ?? null;
}
function typeSpecIsAlias(spec) {
    const s = spec;
    return s?.Assign?.IsValid?.() ?? s?.alias === true;
}
function fieldListFields(list) {
    const l = list;
    return l?.List ?? l?.fields ?? [];
}
function fieldNames(field) {
    const f = field;
    return f?.Names ?? f?.names ?? [];
}
function fieldType(field) {
    const f = field;
    return f?.Type ?? f?.type ?? null;
}
registerCheckerMethod("typeDecl", function typeDeclMethod(obj, tdecl) {
    assert(obj.typ === null);
    let versionErr = false;
    let rhs = Typ[Invalid];
    this.later(() => {
        const t = asNamed(obj.typ);
        if (t !== null) {
            this.validType(t);
        }
        const tdType = typeSpecType(tdecl);
        void (!versionErr && this.isImportedConstraint(rhs) && this.verifyVersionf(tdType, go1_18, "using type constraint %s", rhs));
    }).describef(obj, "validType(%s)", obj.Name());
    const typeParams = typeSpecTypeParams(tdecl);
    const tdType = typeSpecType(tdecl);
    const tparam0 = fieldListNumFields(typeParams) > 0 ? fieldListFields(typeParams)[0] : null;
    // alias declaration
    if (typeSpecIsAlias(tdecl)) {
        if (!versionErr && tparam0 !== null && tparam0 !== undefined && !this.verifyVersionf(tparam0, go1_23, "generic type alias")) {
            versionErr = true;
        }
        if (!versionErr && !this.verifyVersionf(new atPos(nopos), go1_9, "type alias")) {
            versionErr = true;
        }
        const alias = this.newAlias(obj, null);
        try {
            let closeTypeParamScope = false;
            if (tparam0 !== null && tparam0 !== undefined) {
                this.openScope(tdecl, "type parameters");
                closeTypeParamScope = true;
            }
            try {
                if (tparam0 !== null && tparam0 !== undefined) {
                    const dst = { value: alias.tparams };
                    this.collectTypeParams(dst, typeParams);
                    alias.tparams = dst.value;
                }
                rhs = this.declaredType(tdType, obj);
            }
            finally {
                if (closeTypeParamScope) {
                    this.closeScope();
                }
            }
            assert(rhs !== null);
            alias.fromRHS = rhs;
            if (rhs instanceof TypeParam && alias.tparams !== null && alias.tparams.list().includes(rhs)) {
                this.error(tdType, "MisplacedTypeParam", "cannot use type parameter declared in alias declaration as RHS");
                alias.fromRHS = Typ[Invalid];
            }
        }
        finally {
            if (alias.fromRHS === null) {
                alias.fromRHS = Typ[Invalid];
                unalias(alias);
            }
        }
        return;
    }
    // type definition or generic type declaration
    if (!versionErr && tparam0 !== null && tparam0 !== undefined && !this.verifyVersionf(tparam0, go1_18, "type parameter")) {
        versionErr = true;
    }
    const named = this.newNamed(obj, null, null);
    let closeTypeParamScope = false;
    if (typeParams !== null && typeParams !== undefined) {
        this.openScope(tdecl, "type parameters");
        closeTypeParamScope = true;
    }
    try {
        if (typeParams !== null && typeParams !== undefined) {
            const dst = { value: named.tparams };
            this.collectTypeParams(dst, typeParams);
            named.tparams = dst.value;
        }
        rhs = this.declaredType(tdType, obj);
    }
    finally {
        if (closeTypeParamScope) {
            this.closeScope();
        }
    }
    assert(rhs !== null);
    named.fromRHS = rhs;
    if (isTypeParam(rhs)) {
        this.error(tdType, "MisplacedTypeParam", "cannot use a type parameter as RHS in type declaration");
        named.fromRHS = Typ[Invalid];
    }
});
registerCheckerMethod("collectTypeParams", function collectTypeParams(dst, list) {
    const tparams = [];
    const fields = fieldListFields(list);
    const scopePos = nodePos(list) || nopos;
    for (const f of fields) {
        for (const name of fieldNames(f)) {
            tparams.push(this.declareTypeParam(name, scopePos));
        }
    }
    dst.value = bindTParams(tparams);
    assert(!this.inTParamList);
    this.inTParamList = true;
    try {
        let index = 0;
        for (const f of fields) {
            let bound;
            const ftype = fieldType(f);
            if (ftype !== null && ftype !== undefined) {
                bound = this.bound(ftype);
                if (isTypeParam(bound)) {
                    this.error(ftype, "MisplacedTypeParam", "cannot use a type parameter as constraint");
                    bound = Typ[Invalid];
                }
            }
            else {
                bound = Typ[Invalid];
            }
            const names = fieldNames(f);
            for (let i = 0; i < names.length; i++) {
                tparams[index + i].bound = bound;
            }
            index += names.length;
        }
    }
    finally {
        this.inTParamList = false;
    }
});
registerCheckerMethod("bound", function bound(x) {
    let wrap = false;
    const expr = x;
    switch (expr.kind) {
        case "UnaryExpr":
            wrap = expr.Op === "TILDE" || expr.op === TokenKind.Tilde;
            break;
        case "BinaryExpr":
            wrap = expr.Op === "OR" || expr.op === TokenKind.Or;
            break;
    }
    if (wrap) {
        const t = this.typ({ kind: "InterfaceType", methods: { kind: "FieldList", fields: [{ kind: "Field", names: [], type: x }] } });
        if (t instanceof Interface) {
            t.implicit = true;
        }
        return t;
    }
    return this.typ(x);
});
registerCheckerMethod("declareTypeParam", function declareTypeParam(name, scopePos) {
    const tname = NewTypeName(nodePos(name) || nopos, this.pkg, identName(name), null);
    const tpar = this.newTypeParam(tname, Typ[Invalid]); // assigns type to tname as a side-effect
    this.declare(this.scope, name, tname, scopePos);
    return tpar;
});
registerCheckerMethod("collectMethods", function collectMethods(obj) {
    const methods = this.methods?.get(obj);
    if (methods === undefined) {
        return;
    }
    this.methods?.delete(obj);
    const mset = new objset();
    const base = asNamed(obj.typ);
    if (base !== null) {
        this.later(() => {
            this.checkFieldUniqueness(base);
        }).describef(obj, "verifying field uniqueness for %v", base);
        for (let i = 0; i < base.NumMethods(); i++) {
            const m = base.Method(i);
            assert(m.name !== "_");
            assert(mset.insert(m) === null);
        }
    }
    for (const m of methods) {
        assert(m.name !== "_");
        const alt = mset.insert(m);
        if (alt !== null) {
            this.errorf(m, "DuplicateMethod", "method %s.%s already declared", obj.Name(), m.name);
            continue;
        }
        if (base !== null) {
            base.AddMethod(m);
        }
    }
});
registerCheckerMethod("checkFieldUniqueness", function checkFieldUniqueness(base) {
    const t = base.Underlying();
    if (t instanceof Struct) {
        const mset = new objset();
        for (let i = 0; i < base.NumMethods(); i++) {
            const m = base.Method(i);
            assert(m.name !== "_");
            assert(mset.insert(m) === null);
        }
        for (const fld of t.fields ?? []) {
            if (fld.name !== "_") {
                const alt = mset.insert(fld);
                if (alt !== null) {
                    const err = this.newError("DuplicateFieldAndMethod");
                    err.addf(alt, "field and method with the same name %s", fld.name);
                    err.addAltDecl(fld);
                    err.report();
                }
            }
        }
    }
});
registerCheckerMethod("funcDecl", function funcDeclMethod(obj, decl) {
    assert(obj.typ === null);
    assert(this.iota === null);
    const sig = new Signature();
    obj.typ = sig; // guard against cycles
    const fdecl = decl.fdecl;
    const recv = funcDeclRecv(fdecl);
    const ftyp = funcDeclType(fdecl);
    const body = funcDeclBody(fdecl);
    this.funcType(sig, recv, ftyp);
    if (sig.scope !== null) {
        sig.scope.pos = nodePos(fdecl) || nopos;
        sig.scope.end = nodeEnd(fdecl) || nopos;
    }
    const typeParams = ftyp?.TypeParams ?? ftyp?.typeParams;
    if (fieldListNumFields(typeParams) > 0 && (body === null || body === undefined)) {
        this.softErrorf(funcDeclName(fdecl), "BadDecl", "generic function is missing function body");
    }
    if (!this.conf.IgnoreFuncBodies && body !== null && body !== undefined) {
        this.later(() => {
            this.funcBody(decl, obj.name, sig, body, null);
        }).describef(obj, "func %s", obj.name);
    }
});
registerCheckerMethod("declStmt", function declStmt(d) {
    const pkg = this.pkg;
    this.walkDecl(d, (d) => {
        if (d instanceof constDecl) {
            const top = this.delayed.length;
            const names = specNames(d.spec);
            const lhs = new Array(names.length);
            for (let i = 0; i < names.length; i++) {
                const name = names[i];
                const obj = NewConst(nodePos(name) || nopos, pkg, identName(name), null, d.iota);
                lhs[i] = obj;
                let init = null;
                if (i < d.init.length) {
                    init = d.init[i];
                }
                this.constDecl(obj, d.typ, init, d.inherited);
            }
            this.processDelayed(top);
            const scopePos = nodeEnd(d.spec) || nopos;
            for (let i = 0; i < names.length; i++) {
                this.declare(this.scope, names[i], lhs[i], scopePos);
            }
        }
        else if (d instanceof varDecl) {
            const top = this.delayed.length;
            const names = specNames(d.spec);
            const values = specValues(d.spec);
            const lhs0 = new Array(names.length);
            for (let i = 0; i < names.length; i++) {
                const name = names[i];
                lhs0[i] = newVar(LocalVar, nodePos(name) || nopos, pkg, identName(name), null);
            }
            for (let i = 0; i < lhs0.length; i++) {
                const obj = lhs0[i];
                let lhs = null;
                let init = null;
                switch (values.length) {
                    case names.length:
                        init = values[i];
                        break;
                    case 1:
                        lhs = lhs0;
                        init = values[0];
                        break;
                    default:
                        if (i < values.length) {
                            init = values[i];
                        }
                }
                this.varDecl(obj, lhs, specType(d.spec), init);
                if (values.length === 1) {
                    break;
                }
            }
            this.processDelayed(top);
            const scopePos = nodeEnd(d.spec) || nopos;
            for (let i = 0; i < names.length; i++) {
                this.declare(this.scope, names[i], lhs0[i], scopePos);
            }
        }
        else if (d instanceof typeDecl) {
            const name = specName(d.spec);
            const obj = NewTypeName(nodePos(name) || nopos, pkg, identName(name), null);
            const scopePos = nodePos(name) || nopos;
            this.declare(this.scope, name, obj, scopePos);
            this.push(obj); // mark as grey
            this.typeDecl(obj, d.spec);
            this.pop();
        }
        else {
            this.errorf(d.node(), "InvalidSyntaxTree", "unknown ast.Decl node %T", d.node());
        }
    });
});
function isGeneric(t) {
    return t instanceof Named && t.TypeParams() !== null && t.TypeParams().Len() > 0;
}
function isConstType(t) {
    return t !== null;
}
function isValid(t) {
    return t !== Typ[Invalid];
}
function isTypeParam(t) {
    return t instanceof TypeParam;
}
function newTarget(typ, name) {
    return { typ, name };
}
