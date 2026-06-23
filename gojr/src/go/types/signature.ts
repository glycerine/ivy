// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import type { Scope } from "./scope.js";
import type { Type } from "./type.js";
import { TypeString } from "./typestring.js";
import type { Tuple } from "./tuple.js";
import type { TypeParam, TypeParamList } from "./typeparam.js";
import type { Var } from "./object.js";

// ----------------------------------------------------------------------------
// API

// A Signature represents a (non-builtin) function or method type.
// The receiver is ignored when comparing signatures for identity.
export class Signature implements Type {
  // We need to keep the scope in Signature (rather than passing it around
  // and store it in the Func Object) because when type-checking a function
  // literal we call the general type checker which returns a general Type.
  // We then unpack the *Signature and use the scope for the literal body.
  public rparams: TypeParamList | null = null; // receiver type parameters from left to right, or nil
  public tparams: TypeParamList | null = null; // type parameters from left to right, or nil
  public scope: Scope | null = null; // function scope for package-local and non-instantiated signatures; nil otherwise
  public recv: Var | null = null; // nil if not a method
  public recvold: Var | null = null; // receiver dropped via method selection; or nil
  public params: Tuple | null = null; // (incoming) parameters from left to right; or nil
  public results: Tuple | null = null; // (outgoing) results from left to right; or nil
  public variadic = false; // true if the last parameter's type is of the form ...T

  public constructor(recv: Var | null = null, params: Tuple | null = null, results: Tuple | null = null, variadic = false) {
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
  public Recv(): Var | null { return this.recv; }

  // TypeParams returns the type parameters of signature s, or nil.
  public TypeParams(): TypeParamList | null { return this.tparams; }

  // RecvTypeParams returns the receiver type parameters of signature s, or nil.
  public RecvTypeParams(): TypeParamList | null { return this.rparams; }

  // Params returns the parameters of signature s, or nil.
  // See [NewSignatureType] for details of variadic functions.
  public Params(): Tuple | null { return this.params; }

  // Results returns the results of signature s, or nil.
  public Results(): Tuple | null { return this.results; }

  // Variadic reports whether the signature s is variadic.
  public Variadic(): boolean { return this.variadic; }

  public Underlying(): Type { return this; }
  public String(): string { return TypeString(this, null); }
}

// sentinel value for detecting method expressions
export const methodExprSentinel = {} as Var;

// NewSignature returns a new function type for the given receiver, parameters,
// and results, either of which may be nil. If variadic is set, the function
// is variadic, it must have at least one parameter, and the last parameter
// must be of unnamed slice type.
//
// Deprecated: Use [NewSignatureType] instead which allows for type parameters.
//
//go:fix inline
export function NewSignature(recv: Var | null, params: Tuple | null, results: Tuple | null, variadic: boolean): Signature {
  return NewSignatureType(recv, null, null, params, results, variadic);
}

// NewSignatureType creates a new function type for the given receiver,
// receiver type parameters, type parameters, parameters, and results.
export function NewSignatureType(recv: Var | null, recvTypeParams: TypeParam[] | null, typeParams: TypeParam[] | null, params: Tuple | null, results: Tuple | null, variadic: boolean): Signature {
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
