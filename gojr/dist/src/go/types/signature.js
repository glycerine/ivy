// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { EndOf, PosOf, Unparen } from "../../front/ast.js";
import { TypeString } from "./typestring.js";
import { NewTuple } from "./tuple.js";
import { TypeParam } from "./typeparam.js";
import { Alias, Unalias } from "./alias.js";
import { Basic, UnsafePointer } from "./basic.js";
import { atPos, registerCheckerMethod } from "./check.js";
import { deref } from "./lookup.js";
import { Interface } from "./interface.js";
import { Named } from "./named.js";
import { NewPointer, Pointer } from "./pointer.js";
import { isValid } from "./predicates.js";
import { NewSlice } from "./slice.js";
import { makeRenameMap } from "./subst.js";
import { typexpr } from "./operand.js";
import { Typ } from "./universe.js";
import { Invalid } from "./basic.js";
import { bindTParams } from "./typelists.js";
import { go1_18 } from "./version.js";
import { measure } from "./assignments.js";
import { newVar, ParamVar, RecvVar, ResultVar, setObjectSignatureConstructor } from "./object.js";
// ----------------------------------------------------------------------------
// API
// A Signature represents a (non-builtin) function or method type.
// The receiver is ignored when comparing signatures for identity.
export class Signature {
    static {
        setObjectSignatureConstructor(Signature);
    }
    // We need to keep the scope in Signature (rather than passing it around
    // and store it in the Func Object) because when type-checking a function
    // literal we call the general type checker which returns a general Type.
    // We then unpack the *Signature and use the scope for the literal body.
    rparams = null; // receiver type parameters from left to right, or nil
    tparams = null; // type parameters from left to right, or nil
    scope = null; // function scope for package-local and non-instantiated signatures; nil otherwise
    recv = null; // nil if not a method
    recvold = null; // receiver dropped via method selection; or nil
    params = null; // (incoming) parameters from left to right; or nil
    results = null; // (outgoing) results from left to right; or nil
    variadic = false; // true if the last parameter's type is of the form ...T
    constructor(recv = null, params = null, results = null, variadic = false) {
        this.recv = recv;
        this.params = params;
        this.results = results;
        this.variadic = variadic;
    }
    // Recv returns the receiver of signature s (if a method), or nil if a
    // function. It is ignored when comparing signatures for identity.
    //
    // For an abstract method, Recv returns the enclosing interface either
    // as a *[Named] or an *[Interface]. Due to embedding, an interface may
    // contain methods whose receiver type is a different interface.
    Recv() { return this.recv; }
    // TypeParams returns the type parameters of signature s, or nil.
    TypeParams() { return this.tparams; }
    // RecvTypeParams returns the receiver type parameters of signature s, or nil.
    RecvTypeParams() { return this.rparams; }
    // Params returns the parameters of signature s, or nil.
    // See [NewSignatureType] for details of variadic functions.
    Params() { return this.params; }
    // Results returns the results of signature s, or nil.
    Results() { return this.results; }
    // Variadic reports whether the signature s is variadic.
    Variadic() { return this.variadic; }
    Underlying() { return this; }
    String() { return TypeString(this, null); }
}
// sentinel value for detecting method expressions
export const methodExprSentinel = {};
// NewSignature returns a new function type for the given receiver, parameters,
// and results, either of which may be nil. If variadic is set, the function
// is variadic, it must have at least one parameter, and the last parameter
// must be of unnamed slice type.
//
// Deprecated: Use [NewSignatureType] instead which allows for type parameters.
//
//go:fix inline
export function NewSignature(recv, params, results, variadic) {
    return NewSignatureType(recv, null, null, params, results, variadic);
}
// NewSignatureType creates a new function type for the given receiver,
// receiver type parameters, type parameters, parameters, and results.
export function NewSignatureType(recv, recvTypeParams, typeParams, params, results, variadic) {
    if (variadic) {
        const n = params?.Len() ?? 0;
        if (n === 0) {
            throw new Error("variadic function must have at least one parameter");
        }
    }
    const sig = new Signature(recv, params, results, variadic);
    if (recvTypeParams !== null && recvTypeParams.length !== 0) {
        if (recv === null) {
            throw new Error("function with receiver type parameters must have a receiver");
        }
        sig.rparams = bindTParams(recvTypeParams);
    }
    if (typeParams !== null && typeParams.length !== 0) {
        sig.tparams = bindTParams(typeParams);
    }
    return sig;
}
// ----------------------------------------------------------------------------
// Implementation
// funcType type-checks a function or method type.
registerCheckerMethod("funcType", function funcType(sig, recvPar, ftyp) {
    this.openScope(ftyp, "function");
    this.scope.isFunc = true;
    this.recordScope(ftyp, this.scope);
    sig.scope = this.scope;
    try {
        // collect method receiver, if any
        let recv = null;
        let rparams = null;
        if (recvPar !== undefined && recvPar.fields.length > 0) {
            // We have at least one receiver; make sure we don't have more than one.
            const n = recvPar.fields.length;
            if (n > 1) {
                this.error(new atPos(PosOf(recvPar.fields[n - 1])), "InvalidRecv", "method has multiple receivers");
                // continue with first one
            }
            // all type parameters' scopes start after the method name
            const scopePos = PosOf(ftyp);
            [recv, rparams] = this.collectRecv(recvPar.fields[0], scopePos);
        }
        // collect and declare function type parameters
        if (ftyp.typeParams !== undefined) {
            const dst = { value: sig.tparams };
            this.collectTypeParams(dst, ftyp.typeParams);
            sig.tparams = dst.value;
        }
        // collect ordinary and result parameters
        const [pnames, params, variadic] = this.collectParams(ParamVar, ftyp.params);
        const [rnames, results] = this.collectParams(ResultVar, ftyp.results);
        // declare named receiver, ordinary, and result parameters
        const scopePos = EndOf(ftyp); // all parameter's scopes start after the signature
        if (recv !== null && recv.name !== "") {
            this.declare(this.scope, recvPar.fields[0].names[0], recv, scopePos);
        }
        this.declareParams(pnames, params, scopePos);
        this.declareParams(rnames, results, scopePos);
        sig.recv = recv;
        sig.rparams = rparams;
        sig.params = NewTuple(...params);
        sig.results = NewTuple(...results);
        sig.variadic = variadic;
    }
    finally {
        this.closeScope();
    }
});
// collectRecv extracts the method receiver and its type parameters (if any) from rparam.
// It declares the type parameters (but not the receiver) in the current scope, and
// returns the receiver variable and its type parameter list (if any).
registerCheckerMethod("collectRecv", function collectRecv(rparam, scopePos) {
    // Unpack the receiver parameter which is of the form
    //
    //	"(" [rfield] ["*"] rbase ["[" rtparams "]"] ")"
    //
    // The receiver name rname, the pointer indirection, and the
    // receiver type parameters rtparams may not be present.
    const [rptr, rbase, rtparams] = unpackRecv(rparam.type, true);
    // Determine the receiver base type.
    let recvType = Typ[Invalid];
    let recvTParamsList = null;
    if (rtparams === null) {
        // If there are no type parameters, we can simply typecheck rparam.Type.
        // If that is a generic type, varType will complain.
        // Further receiver constraints will be checked later, with validRecv.
        // We use rparam.Type (rather than base) to correctly record pointer
        // and parentheses in types.Info (was bug, see go.dev/issue/68639).
        recvType = this.varType(rparam.type);
        // Defining new methods on instantiated (alias or defined) types is not permitted.
        // Follow literal pointer/alias type chain and check.
        // (Correct code permits at most one pointer indirection, but for this check it
        // doesn't matter if we have multiple pointers.)
        let a = unpointer(recvType) instanceof Alias ? unpointer(recvType) : null; // recvType is not generic per above
        for (; a !== null;) {
            const baseType = unpointer(a.fromRHS);
            const g = isGenericType(baseType) ? baseType : null;
            if (g !== null && g.TypeParams() !== null) {
                this.errorf(new atPos(PosOf(rbase)), "InvalidRecv", "cannot define new methods on instantiated type %s", g);
                recvType = Typ[Invalid]; // avoid follow-on errors by Checker.validRecv
                break;
            }
            a = baseType instanceof Alias ? baseType : null;
        }
    }
    else {
        // If there are type parameters, rbase must denote a generic base type.
        // Important: rbase must be resolved before declaring any receiver type
        // parameters (which may have the same name, see below).
        let baseType = null; // nil if not valid
        const cause = { value: "" };
        const t = this.genericType(rbase, cause);
        if (isValid(t)) {
            if (t instanceof Named) {
                baseType = t;
            }
            else if (t instanceof Alias) {
                // Methods on generic aliases are not permitted.
                // Only report an error if the alias type is valid.
                if (isValid(t)) {
                    this.errorf(new atPos(PosOf(rbase)), "InvalidRecv", "cannot define new methods on generic alias type %s", t);
                }
                // Ok to continue but do not set basetype in this case so that
                // recvType remains invalid (was bug, see go.dev/issue/70417).
            }
            else {
                throw new Error("unreachable");
            }
        }
        else {
            if (cause.value !== "") {
                this.errorf(new atPos(PosOf(rbase)), "InvalidRecv", "%s", cause.value);
            }
            // Ok to continue but do not set baseType (see comment above).
        }
        // Collect the type parameters declared by the receiver (see also
        // Checker.collectTypeParams). The scope of the type parameter T in
        // "func (r T[T]) f() {}" starts after f, not at r, so we declare it
        // after typechecking rbase (see go.dev/issue/52038).
        const recvTParams = new Array(rtparams.length);
        for (let i = 0; i < rtparams.length; i++) {
            const rparam = rtparams[i];
            const tpar = this.declareTypeParam(rparam, scopePos);
            recvTParams[i] = tpar;
            // For historic reasons, type parameters in receiver type expressions
            // are considered both definitions and uses and thus must be recorded
            // in the Info.Uses and Info.Types maps (see go.dev/issue/68670).
            this.recordUse(rparam, tpar.obj);
            this.recordTypeAndValue(rparam, typexpr, tpar, null);
        }
        recvTParamsList = bindTParams(recvTParams);
        // Get the type parameter bounds from the receiver base type
        // and set them for the respective (local) receiver type parameters.
        if (baseType !== null) {
            const baseTParams = baseType.TypeParams().list();
            if (recvTParams.length === baseTParams.length) {
                const smap = makeRenameMap(baseTParams, recvTParams);
                for (let i = 0; i < recvTParams.length; i++) {
                    const recvTPar = recvTParams[i];
                    const baseTPar = baseTParams[i];
                    this.mono?.recordCanon?.(recvTPar, baseTPar);
                    // baseTPar.bound is possibly parameterized by other type parameters
                    // defined by the generic base type. Substitute those parameters with
                    // the receiver type parameters declared by the current method.
                    recvTPar.bound = this.subst(recvTPar.obj.pos, baseTPar.bound, smap, null, this.context());
                }
            }
            else {
                const got = measure(recvTParams.length, "type parameter");
                this.errorf(new atPos(PosOf(rbase)), "BadRecv", "receiver declares %s, but receiver base type declares %d", got, baseTParams.length);
            }
            // The type parameters declared by the receiver also serve as
            // type arguments for the receiver type. Instantiate the receiver.
            this.verifyVersionf(new atPos(PosOf(rbase)), go1_18, "type instantiation");
            const targs = new Array(recvTParams.length);
            for (let i = 0; i < recvTParams.length; i++) {
                targs[i] = recvTParams[i];
            }
            recvType = this.instance(PosOf(rparam.type), baseType, targs, null, this.context());
            this.recordInstance(rbase, targs, recvType);
            // Reestablish pointerness if needed (but avoid a pointer to an invalid type).
            if (rptr && isValid(recvType)) {
                recvType = NewPointer(recvType);
            }
            this.recordParenthesizedRecvTypes(rparam.type, recvType);
        }
    }
    // Make sure we have no more than one receiver name.
    let rname = null;
    const n = rparam.names.length;
    if (n >= 1) {
        if (n > 1) {
            this.error(new atPos(PosOf(rparam.names[n - 1])), "InvalidRecv", "method has multiple receivers");
        }
        rname = rparam.names[0];
    }
    // Create the receiver parameter.
    // recvType is invalid if baseType was never set.
    let recv;
    if (rname !== null && rname.name !== "") {
        // named receiver
        recv = newVar(RecvVar, PosOf(rname), this.pkg, rname.name, recvType);
        // In this case, the receiver is declared by the caller
        // because it must be declared after any type parameters
        // (otherwise it might shadow one of them).
    }
    else {
        // anonymous receiver
        recv = newVar(RecvVar, PosOf(rparam), this.pkg, "", recvType);
        this.recordImplicit(rparam, recv);
    }
    // Delay validation of receiver type as it may cause premature expansion of types
    // the receiver type is dependent on (see go.dev/issue/51232, go.dev/issue/51233).
    this.later(() => {
        this.validRecv(new atPos(PosOf(rbase)), recv);
    }).describef(recv, "validRecv(%s)", recv);
    return [recv, recvTParamsList];
});
export function unpointer(t) {
    for (;;) {
        const p = t instanceof Pointer ? t : null;
        if (p === null) {
            return t;
        }
        t = p.base;
    }
}
// recordParenthesizedRecvTypes records parenthesized intermediate receiver type
// expressions that all map to the same type, by recursively unpacking expr and
// recording the corresponding type for it. Example:
//
//	expression  -->  type
//	----------------------
//	(*(T[P]))        *T[P]
//	 *(T[P])         *T[P]
//	  (T[P])          T[P]
//	   T[P]           T[P]
registerCheckerMethod("recordParenthesizedRecvTypes", function recordParenthesizedRecvTypes(expr, typ) {
    for (;;) {
        this.recordTypeAndValue(expr, typexpr, typ, null);
        switch (expr.kind) {
            case "ParenExpr":
                expr = expr.expr;
                break;
            case "StarExpr": {
                expr = expr.expr;
                // In a correct program, typ must be an unnamed
                // pointer type. But be careful and don't panic.
                const ptr = typ instanceof Pointer ? typ : null;
                if (ptr === null) {
                    return; // something is wrong
                }
                typ = ptr.base;
                break;
            }
            default:
                return; // cannot unpack any further
        }
    }
});
// collectParams collects (but does not declare) all parameter/result
// variables of list and returns the list of names and corresponding
// variables, and whether the (parameter) list is variadic.
// Anonymous parameters are recorded with nil names.
registerCheckerMethod("collectParams", function collectParams(kind, list) {
    const names = [];
    const params = [];
    let variadic = false;
    if (list === undefined) {
        return [names, params, variadic];
    }
    let named = false;
    let anonymous = false;
    for (let i = 0; i < list.fields.length; i++) {
        const field = list.fields[i];
        let ftype = field.type;
        if (ftype.kind === "Ellipsis") {
            ftype = ftype.element;
            if (kind === ParamVar && i === list.fields.length - 1 && field.names.length <= 1) {
                variadic = true;
            }
            else {
                this.softErrorf(new atPos(PosOf(field.type)), "InvalidSyntaxTree", "invalid use of ...");
                // ignore ... and continue
            }
        }
        const typ = this.varType(ftype);
        // The parser ensures that f.Tag is nil and we don't
        // care if a constructed AST contains a non-nil tag.
        if (field.names.length > 0) {
            // named parameter
            for (const name of field.names) {
                if (name.name === "") {
                    this.error(new atPos(PosOf(name)), "InvalidSyntaxTree", "anonymous parameter");
                    // ok to continue
                }
                const par = newVar(kind, PosOf(name), this.pkg, name.name, typ);
                // named parameter is declared by caller
                names.push(name);
                params.push(par);
            }
            named = true;
        }
        else {
            // anonymous parameter
            const par = newVar(kind, PosOf(ftype), this.pkg, "", typ);
            this.recordImplicit(field, par);
            names.push(null);
            params.push(par);
            anonymous = true;
        }
    }
    if (named && anonymous) {
        this.error(new atPos(PosOf(list)), "InvalidSyntaxTree", "list contains both named and anonymous parameters");
        // ok to continue
    }
    // For a variadic function, change the last parameter's type from T to []T.
    // Since we type-checked T rather than ...T, we also need to retro-actively
    // record the type for ...T.
    if (variadic) {
        const last = params[params.length - 1];
        last.typ = NewSlice(last.typ);
        this.recordTypeAndValue(list.fields[list.fields.length - 1].type, typexpr, last.typ, null);
    }
    return [names, params, variadic];
});
// declareParams declares each named parameter in the current scope.
registerCheckerMethod("declareParams", function declareParams(names, params, scopePos) {
    for (let i = 0; i < names.length; i++) {
        const name = names[i] ?? null;
        if (name !== null && name.name !== "") {
            this.declare(this.scope, name, params[i], scopePos);
        }
    }
});
// validRecv verifies that the receiver satisfies its respective spec requirements
// and reports an error otherwise.
registerCheckerMethod("validRecv", function validRecv(pos, recv) {
    // spec: "The receiver type must be of the form T or *T where T is a type name."
    const [rtyp] = deref(recv.typ);
    const unaliased = Unalias(rtyp) ?? rtyp;
    if (!isValid(unaliased)) {
        return; // error was reported before
    }
    // spec: "The type denoted by T is called the receiver base type; it must not
    // be a pointer or interface type and it must be declared in the same package
    // as the method."
    if (unaliased instanceof Named) {
        const T = unaliased;
        if (T.obj.pkg !== this.pkg || isCGoTypeObj(this.fset, T.obj)) {
            this.errorf(pos, "InvalidRecv", "cannot define new methods on non-local type %s", rtyp);
        }
        let cause = "";
        const u = T.Underlying();
        if (u instanceof Basic) {
            // unsafe.Pointer is treated like a regular pointer
            if (u.kind === UnsafePointer) {
                cause = "unsafe.Pointer";
            }
        }
        else if (u instanceof Pointer || u instanceof Interface) {
            cause = "pointer or interface type";
        }
        else if (u instanceof TypeParam) {
            throw new Error("unreachable");
        }
        if (cause !== "") {
            this.errorf(pos, "InvalidRecv", "invalid receiver type %s (%s)", rtyp, cause);
        }
    }
    else if (unaliased instanceof Basic) {
        this.errorf(pos, "InvalidRecv", "cannot define new methods on non-local type %s", rtyp);
    }
    else {
        this.errorf(pos, "InvalidRecv", "invalid receiver type %s", recv.typ);
    }
});
// isCGoTypeObj reports whether the given type name was created by cgo.
export function isCGoTypeObj(fset, obj) {
    const filename = fileNameForPos(fset, obj.pos);
    return obj.name.startsWith("_Ctype_") ||
        filename.split(/[\\/]/).pop()?.startsWith("_cgo_") === true;
}
function unpackRecv(rtyp, unpackParams) {
    let e = Unparen(rtyp);
    let ptr = false;
    if (e.kind === "StarExpr") {
        ptr = true;
        e = Unparen(e.expr);
    }
    if (!unpackParams) {
        return [ptr, e, null];
    }
    if (e.kind === "IndexExpr" || e.kind === "IndexListExpr") {
        const indices = e.kind === "IndexExpr" ? [e.index] : e.indices;
        const rtparams = [];
        for (const ix of indices) {
            const name = Unparen(ix);
            if (name.kind !== "Ident") {
                return [ptr, e.kind === "IndexExpr" || e.kind === "IndexListExpr" ? e.object : e, null];
            }
            rtparams.push(name);
        }
        return [ptr, e.object, rtparams];
    }
    return [ptr, e, null];
}
function isGenericType(t) {
    return typeof t.TypeParams === "function";
}
function fileNameForPos(fset, pos) {
    const file = fset?.File?.(pos);
    return file?.Name?.() ?? "";
}
