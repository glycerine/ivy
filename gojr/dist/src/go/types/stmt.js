// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// This file implements typechecking of statements.
import { EndOf, PosOf, Unparen } from "../../front/ast.js";
import { TokenKind } from "../../front/token.js";
import { Checker, atPos, cmpPos, debug, environment } from "./check.js";
import { NewScope, resolve } from "./scope.js";
import { LocalVar, Nil, ParamVar, RecvVar, ResultVar, Var, newVar } from "./object.js";
import { allBoolean, allNumeric, Comparable, hasNil, Identical, isTypeParam, IsInterface, isValid } from "./predicates.js";
import { exprKind, Typ } from "./universe.js";
import { BasicKind } from "./basic.js";
import { operand, builtin, constant_, typexpr } from "./operand.js";
import { constantKindOf, toInt } from "./const.js";
import { TypeString } from "./typestring.js";
import { assert } from "./util.js";
// decl may be nil
Checker.prototype.funcBody = function funcBody(decl, name, sig, body, iota) {
    if (this.conf.IgnoreFuncBodies) {
        throw new Error("function body not ignored");
    }
    if (this.conf._Trace) {
        this.trace(PosOf(body), "-- %s: %s", name, sig);
    }
    // save/restore current environment and set up function environment
    // (and use 0 indentation at function start)
    const env = new environment();
    env.decl = this.decl;
    env.scope = this.scope;
    env.version = this.version;
    env.iota = this.iota;
    env.errpos = this.errpos;
    env.inTParamList = this.inTParamList;
    env.sig = this.sig;
    env.isPanic = this.isPanic;
    env.hasLabel = this.hasLabel;
    env.hasCallOrRecv = this.hasCallOrRecv;
    env.exprPos = this.exprPos;
    const indent = this.indent;
    try {
        this.decl = decl;
        this.scope = sig.scope;
        this.version = this.version; // TODO(adonovan): would decl.version (if decl != nil) be better?
        this.iota = iota;
        this.errpos = null;
        this.inTParamList = false;
        this.sig = sig;
        this.isPanic = null;
        this.hasLabel = false;
        this.hasCallOrRecv = false;
        this.exprPos = 0;
        this.indent = 0;
        this.stmtList(0, body.statements);
        if (this.hasLabel) {
            this.labels(body);
        }
        if ((sig.results?.Len() ?? 0) > 0 && !this.isTerminating(body, "")) {
            this.error(new atPos(EndOf(body)), "MissingReturn", "missing return");
        }
        // spec: "Implementation restriction: A compiler may make it illegal to
        // declare a variable inside a function body if the variable is never used."
        this.usage(sig.scope);
    }
    finally {
        this.decl = env.decl;
        this.scope = env.scope;
        this.version = env.version;
        this.iota = env.iota;
        this.errpos = env.errpos;
        this.inTParamList = env.inTParamList;
        this.sig = env.sig;
        this.isPanic = env.isPanic;
        this.hasLabel = env.hasLabel;
        this.hasCallOrRecv = env.hasCallOrRecv;
        this.exprPos = env.exprPos;
        this.indent = indent;
    }
};
Checker.prototype.usage = function usage(scope) {
    if (scope === null) {
        return;
    }
    const needUse = (kind) => {
        return !(kind === RecvVar || kind === ParamVar || kind === ResultVar);
    };
    const unused = [];
    for (const [name, elem0] of scope.elems.entries()) {
        const elem = resolve(name, elem0);
        const v = elem instanceof Var ? elem : null;
        if (v !== null && needUse(v.kind) && !this.usedVars.get(v)) {
            unused.push(v);
        }
    }
    unused.sort((a, b) => cmpPos(a.pos, b.pos));
    for (const v of unused) {
        this.softErrorf(v, "UnusedVar", "declared and not used: %s", v.name);
    }
    for (const child of scope.children) {
        // Don't go inside function literal scopes a second time;
        // they are handled explicitly by funcBody.
        if (!child.isFunc) {
            this.usage(child);
        }
    }
};
// permissible control-flow statements
export const breakOk = 1 << 0;
export const continueOk = 1 << 1;
export const fallthroughOk = 1 << 2;
// additional context information
export const finalSwitchCase = 1 << 3;
export const inTypeSwitch = 1 << 4;
Checker.prototype.simpleStmt = function simpleStmt(s) {
    if (s !== null && s !== undefined) {
        this.stmt(0, s);
    }
};
export function trimTrailingEmptyStmts(list) {
    for (let i = list.length; i > 0; i--) {
        if (list[i - 1].kind !== "EmptyStmt") {
            return list.slice(0, i);
        }
    }
    return [];
}
Checker.prototype.stmtList = function stmtList(ctxt, list) {
    const ok = (ctxt & fallthroughOk) !== 0;
    let inner = ctxt & ~fallthroughOk;
    list = trimTrailingEmptyStmts(list); // trailing empty statements are "invisible" to fallthrough analysis
    for (let i = 0; i < list.length; i++) {
        let inner2 = inner;
        if (ok && i + 1 === list.length) {
            inner2 |= fallthroughOk;
        }
        this.stmt(inner2, list[i]);
    }
};
Checker.prototype.multipleDefaults = function multipleDefaults(list) {
    let first = null;
    for (const s of list) {
        let d = null;
        switch (s.kind) {
            case "CaseClause":
                if (s.default || s.list.length === 0) {
                    d = s;
                }
                break;
            case "CommClause":
                if (s.default || s.comm === undefined) {
                    d = s;
                }
                break;
            default:
                this.error(s, "InvalidSyntaxTree", "case/communication clause expected");
                break;
        }
        if (d !== null) {
            if (first !== null) {
                this.errorf(d, "DuplicateDefault", "multiple defaults (first at %s)", PosOf(first));
            }
            else {
                first = d;
            }
        }
    }
};
Checker.prototype.openScope = function openScope(node, comment) {
    const scope = NewScope(this.scope, PosOf(node), EndOf(node), comment);
    this.recordScope(node, scope);
    this.scope = scope;
};
Checker.prototype.closeScope = function closeScope() {
    this.scope = this.scope?.Parent() ?? null;
};
export function assignOp(op) {
    switch (op) {
        case TokenKind.PlusAssign:
            return TokenKind.Plus;
        case TokenKind.MinusAssign:
            return TokenKind.Minus;
        case TokenKind.StarAssign:
            return TokenKind.Star;
        case TokenKind.SlashAssign:
            return TokenKind.Slash;
        case TokenKind.PercentAssign:
            return TokenKind.Percent;
        case TokenKind.AmpAssign:
            return TokenKind.Amp;
        case TokenKind.OrAssign:
            return TokenKind.Or;
        case TokenKind.CaretAssign:
            return TokenKind.Caret;
        case TokenKind.ShlAssign:
            return TokenKind.Shl;
        case TokenKind.ShrAssign:
            return TokenKind.Shr;
        case TokenKind.BitClearAssign:
            return TokenKind.BitClear;
    }
    return TokenKind.Illegal;
}
Checker.prototype.suspendedCall = function suspendedCall(keyword, call) {
    const x = new operand();
    let msg = "";
    let code;
    switch (this.rawExpr(null, x, call, null, false)) {
        case exprKind.conversion:
            msg = "requires function call, not conversion";
            code = keyword === "go" ? "InvalidGo" : "InvalidDefer";
            break;
        case exprKind.expression:
            msg = "discards result of";
            code = "UnusedResults";
            break;
        case exprKind.statement:
            return;
        default:
            throw new Error("unreachable");
    }
    this.errorf(x, code, "%s %s %s", keyword, msg, x);
};
// goVal returns the Go value for val, or nil.
export function goVal(val) {
    // val should exist, but be conservative and check
    if (val === null || val === undefined) {
        return null;
    }
    // Match implementation restriction of other compilers.
    // gc only checks duplicates for integer, floating-point
    // and string values, so only create Go values for these
    // types.
    switch (constantKindOf(val)) {
        case "int":
            return toInt(val);
        case "float":
            return typeof val === "number" ? val : Number(val);
        case "string":
            return String(val);
    }
    return null;
}
export class valueType {
    pos;
    typ;
    constructor(pos, typ) {
        this.pos = pos;
        this.typ = typ;
    }
}
Checker.prototype.caseValues = function caseValues(x, values, seen) {
    L: for (const e of values) {
        const v = new operand();
        this.expr(null, v, e);
        if (!x.isValid() || !v.isValid()) {
            continue L;
        }
        this.convertUntyped(v, x.typ());
        if (!v.isValid()) {
            continue L;
        }
        // Order matters: By comparing v against x, error positions are at the case values.
        const res = v.clone(); // keep original v unchanged
        this.comparison(res, x, TokenKind.Equal, true);
        if (!res.isValid()) {
            continue L;
        }
        if (v.mode() !== constant_) {
            continue L; // we're done
        }
        // look for duplicate values
        const val = goVal(v.val);
        if (val !== null) {
            const key = caseValueKey(val);
            // look for duplicate types for a given value
            // (quadratic algorithm, but these lists tend to be very short)
            const bucket = seen.get(key) ?? [];
            for (const vt of bucket) {
                if (Identical(v.typ(), vt.typ)) {
                    const err = this.newError("DuplicateCase");
                    err.addf(v, "duplicate case %s in expression switch", v);
                    err.addf(new atPos(vt.pos), "previous case");
                    err.report();
                    continue L;
                }
            }
            bucket.push(new valueType(v.Pos(), v.typ()));
            seen.set(key, bucket);
        }
    }
};
Checker.prototype.isNil = function isNil(e) {
    // The only way to express the nil value is by literally writing nil (possibly in parentheses).
    const name = Unparen(e);
    if (name.kind === "Ident") {
        return this.lookup(name.name) instanceof Nil;
    }
    return false;
};
Checker.prototype.caseTypes = function caseTypes(x, types, seen) {
    let T = null;
    const dummy = new operand();
    L: for (const e of types) {
        // The spec allows the value nil instead of a type.
        if (this.isNil(e)) {
            T = null;
            this.expr(null, dummy, e); // run e through expr so we get the usual Info recordings
        }
        else {
            T = this.varType(e);
            if (!isValid(T)) {
                continue L;
            }
        }
        // look for duplicate types
        // (quadratic algorithm, but type switches tend to be reasonably small)
        for (const [t, other] of seen.entries()) {
            if (T === null && t === null || T !== null && t !== null && Identical(T, t)) {
                // talk about "case" rather than "type" because of nil case
                let Ts = "nil";
                if (T !== null) {
                    Ts = TypeString(T, (pkg) => this.qualifier(pkg));
                }
                const err = this.newError("DuplicateCase");
                err.addf(e, "duplicate case %s in type switch", Ts);
                err.addf(other, "previous case");
                err.report();
                continue L;
            }
        }
        seen.set(T, e);
        if (x !== null && T !== null) {
            this.typeAssertion(e, x, T, true);
        }
    }
    // spec: "In clauses with a case listing exactly one type, the variable has that type;
    // otherwise, the variable has the type of the expression in the TypeSwitchGuard.
    if (types.length !== 1 || T === null) {
        T = Typ[BasicKind.Invalid];
        if (x !== null) {
            T = x.typ();
        }
    }
    assert(T !== null);
    return T;
};
// TODO(gri) Once we are certain that typeHash is correct in all situations, use this version of caseTypes instead.
// (Currently it may be possible that different types have identical names and import paths due to ImporterFrom.)
Checker.prototype.caseTypes_currently_unused = function caseTypes_currently_unused(x, _xtyp, types, seen) {
    let T = null;
    const dummy = new operand();
    L: for (const e of types) {
        // The spec allows the value nil instead of a type.
        let hash;
        if (this.isNil(e)) {
            this.expr(null, dummy, e); // run e through expr so we get the usual Info recordings
            T = null;
            hash = "<nil>"; // avoid collision with a type named nil
        }
        else {
            T = this.varType(e);
            if (!isValid(T)) {
                continue L;
            }
            throw new Error("enable typeHash(T, nil)");
            // hash = typeHash(T, nil)
        }
        const other = seen.get(hash);
        if (other !== undefined) {
            // talk about "case" rather than "type" because of nil case
            let Ts = "nil";
            if (T !== null) {
                Ts = TypeString(T, (pkg) => this.qualifier(pkg));
            }
            const err = this.newError("DuplicateCase");
            err.addf(e, "duplicate case %s in type switch", Ts);
            err.addf(other, "previous case");
            err.report();
            continue L;
        }
        seen.set(hash, e);
        if (T !== null) {
            this.typeAssertion(e, x, T, true);
        }
    }
    // spec: "In clauses with a case listing exactly one type, the variable has that type;
    // otherwise, the variable has the type of the expression in the TypeSwitchGuard.
    if (types.length !== 1 || T === null) {
        T = Typ[BasicKind.Invalid];
        if (x !== null) {
            T = x.typ();
        }
    }
    assert(T !== null);
    return T;
};
// stmt typechecks statement s.
Checker.prototype.stmt = function stmt(ctxt, s) {
    // statements must end with the same top scope as they started with
    const scope = this.scope;
    // process collected function literals before scope changes
    const top = this.delayed.length;
    try {
        // reset context for statements of inner blocks
        let inner = ctxt & ~(fallthroughOk | finalSwitchCase | inTypeSwitch);
        switch (s.kind) {
            case "BadStmt":
            case "EmptyStmt":
                // ignore
                break;
            case "DeclStmt":
                this.declStmt(s.decl);
                break;
            case "LabeledStmt":
                this.hasLabel = true;
                this.stmt(ctxt, s.stmt);
                break;
            case "ExprStmt": {
                // spec: "With the exception of specific built-in functions,
                // function and method calls and receive operations can appear
                // in statement context. Such statements may be parenthesized."
                const x = new operand();
                const kind = this.rawExpr(null, x, s.expr, null, false);
                let msg;
                let code;
                switch (x.mode()) {
                    default:
                        if (kind === exprKind.statement) {
                            return;
                        }
                        msg = "is not used";
                        code = "UnusedExpr";
                        break;
                    case builtin:
                        msg = "must be called";
                        code = "UncalledBuiltin";
                        break;
                    case typexpr:
                        msg = "is not an expression";
                        code = "NotAnExpr";
                        break;
                }
                this.errorf(x, code, "%s %s", x, msg);
                break;
            }
            case "SendStmt": {
                const ch = new operand();
                const val = new operand();
                this.expr(null, ch, s.channel);
                this.genericExpr(val, s.value, null);
                if (!ch.isValid() || !val.isValid()) {
                    return;
                }
                const elem = this.chanElem(new atPos(PosOf(s)), ch, false);
                if (elem !== null) {
                    this.assignment(val, elem, "send");
                }
                break;
            }
            case "IncDecStmt": {
                let op;
                const tok = s.token;
                switch (tok) {
                    case TokenKind.PlusPlus:
                        op = TokenKind.Plus;
                        break;
                    case TokenKind.MinusMinus:
                        op = TokenKind.Minus;
                        break;
                    default:
                        this.errorf(new atPos(PosOf(s)), "InvalidSyntaxTree", "unknown inc/dec operation %s", tok);
                        return;
                }
                const x = new operand();
                this.expr(null, x, s.expr);
                if (!x.isValid()) {
                    return;
                }
                if (!allNumeric(x.typ())) {
                    this.errorf(s.expr, "NonNumericIncDec", "invalid operation: %s%s (non-numeric type %s)", s.expr, s.token, x.typ());
                    return;
                }
                const Y = { kind: "BasicLit", token: TokenKind.IntLiteral, value: "1", span: s.expr.span };
                this.binary(x, null, s.expr, Y, op, PosOf(s));
                if (!x.isValid()) {
                    return;
                }
                this.assignVar(s.expr, null, x, "assignment");
                break;
            }
            case "AssignStmt":
                switch (s.token) {
                    case TokenKind.Assign:
                    case TokenKind.Define:
                        if (s.lhs.length === 0) {
                            this.error(s, "InvalidSyntaxTree", "missing lhs in assignment");
                            return;
                        }
                        if (s.token === TokenKind.Define) {
                            this.shortVarDecl(new atPos(PosOf(s)), s.lhs, s.rhs);
                        }
                        else {
                            // regular assignment
                            this.assignVars(s.lhs, s.rhs);
                        }
                        break;
                    default: {
                        // assignment operations
                        if (s.lhs.length !== 1 || s.rhs.length !== 1) {
                            this.errorf(new atPos(PosOf(s)), "MultiValAssignOp", "assignment operation %s requires single-valued expressions", s.token);
                            return;
                        }
                        const op = assignOp(s.token);
                        if (op === TokenKind.Illegal) {
                            this.errorf(new atPos(PosOf(s)), "InvalidSyntaxTree", "unknown assignment operation %s", s.token);
                            return;
                        }
                        const x = new operand();
                        this.binary(x, null, s.lhs[0], s.rhs[0], op, PosOf(s));
                        if (!x.isValid()) {
                            return;
                        }
                        this.assignVar(s.lhs[0], null, x, "assignment");
                        break;
                    }
                }
                break;
            case "GoStmt":
                this.suspendedCall("go", s.call);
                break;
            case "DeferStmt":
                this.suspendedCall("defer", s.call);
                break;
            case "ReturnStmt": {
                const res = this.sig?.results ?? null;
                // Return with implicit results allowed for function with named results.
                // (If one is named, all are named.)
                if (s.results.length === 0 && (res?.Len() ?? 0) > 0 && res.vars[0].name !== "") {
                    // spec: "Implementation restriction: A compiler may disallow an empty expression
                    // list in a "return" statement if a different entity (constant, type, or variable)
                    // with the same name as a result parameter is in scope at the place of the return."
                    for (const obj of res.vars) {
                        const alt = this.lookup(obj.name);
                        if (alt !== null && alt !== obj) {
                            const err = this.newError("OutOfScopeResult");
                            err.addf(s, "result parameter %s not in scope at return", obj.name);
                            err.addf(alt, "inner declaration of %s", obj);
                            err.report();
                            // ok to continue
                        }
                    }
                }
                else {
                    let lhs = [];
                    if ((res?.Len() ?? 0) > 0) {
                        lhs = res.vars;
                    }
                    this.initVars(lhs, s.results, s);
                }
                break;
            }
            case "BranchStmt":
                if (s.label !== undefined) {
                    this.hasLabel = true;
                    return; // checked in 2nd pass (check.labels)
                }
                switch (s.token) {
                    case TokenKind.Break:
                        if ((ctxt & breakOk) === 0) {
                            this.error(s, "MisplacedBreak", "break not in for, switch, or select statement");
                        }
                        break;
                    case TokenKind.Continue:
                        if ((ctxt & continueOk) === 0) {
                            this.error(s, "MisplacedContinue", "continue not in for statement");
                        }
                        break;
                    case TokenKind.Fallthrough:
                        if ((ctxt & fallthroughOk) === 0) {
                            let msg;
                            switch (true) {
                                case (ctxt & finalSwitchCase) !== 0:
                                    msg = "cannot fallthrough final case in switch";
                                    break;
                                case (ctxt & inTypeSwitch) !== 0:
                                    msg = "cannot fallthrough in type switch";
                                    break;
                                default:
                                    msg = "fallthrough statement out of place";
                                    break;
                            }
                            this.error(s, "MisplacedFallthrough", msg);
                        }
                        break;
                    default:
                        this.errorf(s, "InvalidSyntaxTree", "branch statement: %s", s.token);
                        break;
                }
                break;
            case "BlockStmt":
                this.openScope(s, "block");
                try {
                    this.stmtList(inner, s.statements);
                }
                finally {
                    this.closeScope();
                }
                break;
            case "IfStmt": {
                this.openScope(s, "if");
                try {
                    this.simpleStmt(s.init);
                    const x = new operand();
                    this.expr(null, x, s.condition);
                    if (x.isValid() && !allBoolean(x.typ())) {
                        this.error(s.condition, "InvalidCond", "non-boolean condition in if statement");
                    }
                    this.stmt(inner, s.body);
                    // The parser produces a correct AST but if it was modified
                    // elsewhere the else branch may be invalid. Check again.
                    switch (s.else?.kind) {
                        case undefined:
                        case "BadStmt":
                            // valid or error already reported
                            break;
                        case "IfStmt":
                        case "BlockStmt":
                            this.stmt(inner, s.else);
                            break;
                        default:
                            this.error(s.else, "InvalidSyntaxTree", "invalid else branch in if statement");
                            break;
                    }
                }
                finally {
                    this.closeScope();
                }
                break;
            }
            case "SwitchStmt": {
                inner |= breakOk;
                this.openScope(s, "switch");
                try {
                    this.simpleStmt(s.init);
                    const x = new operand();
                    if (s.tag !== undefined) {
                        this.expr(null, x, s.tag);
                        // By checking assignment of x to an invisible temporary
                        // (as a compiler would), we get all the relevant checks.
                        this.assignment(x, null, "switch expression");
                        if (x.isValid() && !Comparable(x.typ()) && !hasNil(x.typ())) {
                            this.errorf(x, "InvalidExprSwitch", "cannot switch on %s (%s is not comparable)", x, x.typ());
                            x.invalidate();
                        }
                    }
                    else {
                        // spec: "A missing switch expression is
                        // equivalent to the boolean value true."
                        x.mode_ = constant_;
                        x.typ_ = Typ[BasicKind.Bool];
                        x.val = true;
                        x.expr = { kind: "Ident", name: "true", span: s.span };
                    }
                    this.multipleDefaults(s.body);
                    const seen = new Map(); // map of seen case values to positions and types
                    for (let i = 0; i < s.body.length; i++) {
                        const clause = s.body[i];
                        this.caseValues(x, clause.list, seen);
                        this.openScope(clause, "case");
                        try {
                            let innerCase = inner;
                            if (i + 1 < s.body.length) {
                                innerCase |= fallthroughOk;
                            }
                            else {
                                innerCase |= finalSwitchCase;
                            }
                            this.stmtList(innerCase, clause.body);
                        }
                        finally {
                            this.closeScope();
                        }
                    }
                }
                finally {
                    this.closeScope();
                }
                break;
            }
            case "TypeSwitchStmt": {
                inner |= breakOk | inTypeSwitch;
                this.openScope(s, "type switch");
                try {
                    this.simpleStmt(s.init);
                    let lhs = null; // lhs identifier or nil
                    let rhs;
                    switch (s.assign.kind) {
                        case "ExprStmt":
                            rhs = s.assign.expr;
                            break;
                        case "AssignStmt":
                            if (s.assign.lhs.length !== 1 || s.assign.token !== TokenKind.Define || s.assign.rhs.length !== 1) {
                                this.error(s, "InvalidSyntaxTree", "incorrect form of type switch guard");
                                return;
                            }
                            lhs = s.assign.lhs[0].kind === "Ident" ? s.assign.lhs[0] : null;
                            if (lhs === null) {
                                this.error(s, "InvalidSyntaxTree", "incorrect form of type switch guard");
                                return;
                            }
                            if (lhs.name === "_") {
                                // _ := x.(type) is an invalid short variable declaration
                                this.softErrorf(lhs, "NoNewVar", "no new variable on left side of :=");
                                lhs = null; // avoid declared and not used error below
                            }
                            else {
                                this.recordDef(lhs, null); // lhs variable is implicitly declared in each cause clause
                            }
                            rhs = s.assign.rhs[0];
                            break;
                        default:
                            this.error(s, "InvalidSyntaxTree", "incorrect form of type switch guard");
                            return;
                    }
                    // rhs must be of the form: expr.(type) and expr must be an ordinary interface
                    const expr = rhs.kind === "TypeAssertExpr" ? rhs : null;
                    if (expr === null || expr.type !== undefined) {
                        this.error(s, "InvalidSyntaxTree", "incorrect form of type switch guard");
                        return;
                    }
                    let sx = null; // switch expression against which cases are compared against; nil if invalid
                    {
                        const x = new operand();
                        this.expr(null, x, expr.object);
                        if (x.isValid()) {
                            if (isTypeParam(x.typ())) {
                                this.errorf(x, "InvalidTypeSwitch", "cannot use type switch on type parameter value %s", x);
                            }
                            else if (IsInterface(x.typ())) {
                                sx = x;
                            }
                            else {
                                this.errorf(x, "InvalidTypeSwitch", "%s is not an interface", x);
                            }
                        }
                    }
                    this.multipleDefaults(s.body);
                    const lhsVars = []; // list of implicitly declared lhs variables
                    const seen = new Map(); // map of seen types to positions
                    for (const clause of s.body) {
                        // Check each type in this type switch case.
                        const T = this.caseTypes(sx, clause.list, seen);
                        this.openScope(clause, "case");
                        try {
                            // If lhs exists, declare a corresponding variable in the case-local scope.
                            if (lhs !== null) {
                                const obj = newVar(LocalVar, PosOf(lhs), this.pkg, lhs.name, T);
                                this.declare(this.scope, null, obj, PosOf(clause));
                                this.recordImplicit(clause, obj);
                                // For the "declared and not used" error, all lhs variables act as
                                // one; i.e., if any one of them is 'used', all of them are 'used'.
                                // Collect them for later analysis.
                                lhsVars.push(obj);
                            }
                            this.stmtList(inner, clause.body);
                        }
                        finally {
                            this.closeScope();
                        }
                    }
                    // If lhs exists, we must have at least one lhs variable that was used.
                    // (We can't use check.usage because that only looks at one scope; and
                    // we don't want to use the same variable for all scopes and change the
                    // variable type underfoot.)
                    if (lhs !== null) {
                        let used = false;
                        for (const v of lhsVars) {
                            if (this.usedVars.get(v)) {
                                used = true;
                            }
                            this.usedVars.set(v, true); // avoid usage error when checking entire function
                        }
                        if (!used) {
                            this.softErrorf(lhs, "UnusedVar", "%s declared and not used", lhs.name);
                        }
                    }
                }
                finally {
                    this.closeScope();
                }
                break;
            }
            case "SelectStmt":
                inner |= breakOk;
                this.multipleDefaults(s.body);
                for (const clause of s.body) {
                    // clause.Comm must be a SendStmt, RecvStmt, or default case
                    let valid = false;
                    let rhs = null; // rhs of RecvStmt, or nil
                    const comm = clause.comm;
                    switch (comm?.kind) {
                        case undefined:
                        case "SendStmt":
                            valid = true;
                            break;
                        case "AssignStmt":
                            if (comm.rhs.length === 1) {
                                rhs = comm.rhs[0];
                            }
                            break;
                        case "ExprStmt":
                            rhs = comm.expr;
                            break;
                    }
                    // if present, rhs must be a receive operation
                    if (rhs !== null) {
                        const x = Unparen(rhs);
                        if (x.kind === "UnaryExpr" && x.op === TokenKind.Arrow) {
                            valid = true;
                        }
                    }
                    if (!valid) {
                        this.error(clause.comm, "InvalidSelectCase", "select case must be send or receive (possibly with assignment)");
                        continue;
                    }
                    this.openScope(clause, "case");
                    try {
                        if (clause.comm !== undefined) {
                            this.stmt(inner, clause.comm);
                        }
                        this.stmtList(inner, clause.body);
                    }
                    finally {
                        this.closeScope();
                    }
                }
                break;
            case "ForStmt":
                inner |= breakOk | continueOk;
                this.openScope(s, "for");
                try {
                    this.simpleStmt(s.init);
                    if (s.condition !== undefined) {
                        const x = new operand();
                        this.expr(null, x, s.condition);
                        if (x.isValid() && !allBoolean(x.typ())) {
                            this.error(s.condition, "InvalidCond", "non-boolean condition in for statement");
                        }
                    }
                    this.simpleStmt(s.post);
                    // spec: "The init statement may be a short variable
                    // declaration, but the post statement must not."
                    if (s.post?.kind === "AssignStmt" && s.post.token === TokenKind.Define) {
                        this.softErrorf(s.post, "InvalidPostDecl", "cannot declare in post statement");
                        // Don't call useLHS here because we want to use the lhs in
                        // this erroneous statement so that we don't get errors about
                        // these lhs variables being declared and not used.
                        this.use(...s.post.lhs); // avoid follow-up errors
                    }
                    this.stmt(inner, s.body);
                }
                finally {
                    this.closeScope();
                }
                break;
            case "RangeStmt":
                inner |= breakOk | continueOk;
                this.rangeStmt(inner, s, new atPos(PosOf(s)), s.key ?? null, s.value ?? null, null, s.source, s.token === TokenKind.Define);
                break;
            default:
                this.error(s, "InvalidSyntaxTree", "invalid statement");
                break;
        }
    }
    finally {
        this.processDelayed(top);
        if (debug) {
            assert(scope === this.scope);
        }
    }
};
function caseValueKey(x) {
    if (typeof x === "bigint") {
        return `i:${x}`;
    }
    if (typeof x === "number") {
        return `f:${Object.is(x, -0) ? "-0" : x}`;
    }
    return x;
}
