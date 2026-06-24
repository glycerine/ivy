// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { TypeString } from "./typestring.js";
// ----------------------------------------------------------------------------
// API
// A Signature represents a (non-builtin) function or method type.
// The receiver is ignored when comparing signatures for identity.
export class Signature {
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
import { bindTParams } from "./typelists.js";
