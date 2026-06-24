// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import { Interface } from "./interface.js";
import { MethodSet } from "./methodset.js";
import { Named } from "./named.js";
import type { Func, Var } from "./object.js";
import { Scope } from "./scope.js";
import type { Selection } from "./selection.js";
import { Struct } from "./struct.js";
import type { Type } from "./type.js";
import { Tuple } from "./tuple.js";
import type { TypeParam } from "./typeparam.js";
import { TypeList, TypeParamList } from "./typelists.js";
import { Term, Union } from "./union.js";

export type Seq<T> = (yield_: (value: T) => boolean) => void;

declare module "./interface.js" {
  interface Interface {
    Methods(): Seq<Func>;
    ExplicitMethods(): Seq<Func>;
    EmbeddedTypes(): Seq<Type>;
  }
}

declare module "./named.js" {
  interface Named {
    Methods(): Seq<Func>;
  }
}

declare module "./scope.js" {
  interface Scope {
    Children(): Seq<Scope>;
  }
}

declare module "./struct.js" {
  interface Struct {
    Fields(): Seq<Var>;
  }
}

declare module "./tuple.js" {
  interface Tuple {
    Variables(): Seq<Var>;
  }
}

declare module "./methodset.js" {
  interface MethodSet {
    Methods(): Seq<Selection>;
  }
}

declare module "./union.js" {
  interface Union {
    Terms(): Seq<Term>;
  }
}

declare module "./typelists.js" {
  interface TypeParamList {
    TypeParams(): Seq<TypeParam>;
  }
  interface TypeList {
    Types(): Seq<Type>;
  }
}

// This file defines go1.23 iterator methods for a variety of data
// types. They are not mirrored to cmd/compile/internal/types2, as
// there is no point doing so until the bootstrap compiler it at least
// go1.23; therefore go1.23-style range statements should not be used
// in code common to types and types2, though clients of go/types are
// free to use them.

// Methods returns a go1.23 iterator over all the methods of an
// interface, ordered by Id.
//
// Example: for m := range t.Methods() { ... }
Interface.prototype.Methods = function Methods(): Seq<Func> {
  return (yield_: (m: Func) => boolean): void => {
    for (let i = 0; i < this.NumMethods(); i++) {
      if (!yield_(this.Method(i))) {
        break;
      }
    }
  };
};

// ExplicitMethods returns a go1.23 iterator over the explicit methods of
// an interface, ordered by Id.
//
// Example: for m := range t.ExplicitMethods() { ... }
Interface.prototype.ExplicitMethods = function ExplicitMethods(): Seq<Func> {
  return (yield_: (m: Func) => boolean): void => {
    for (let i = 0; i < this.NumExplicitMethods(); i++) {
      if (!yield_(this.ExplicitMethod(i))) {
        break;
      }
    }
  };
};

// EmbeddedTypes returns a go1.23 iterator over the types embedded within an interface.
//
// Example: for e := range t.EmbeddedTypes() { ... }
Interface.prototype.EmbeddedTypes = function EmbeddedTypes(): Seq<Type> {
  return (yield_: (e: Type) => boolean): void => {
    for (let i = 0; i < this.NumEmbeddeds(); i++) {
      if (!yield_(this.EmbeddedType(i))) {
        break;
      }
    }
  };
};

// Methods returns a go1.23 iterator over the declared methods of a named type.
//
// Example: for m := range t.Methods() { ... }
Named.prototype.Methods = function Methods(): Seq<Func> {
  return (yield_: (m: Func) => boolean): void => {
    for (let i = 0; i < this.NumMethods(); i++) {
      if (!yield_(this.Method(i))) {
        break;
      }
    }
  };
};

// Children returns a go1.23 iterator over the child scopes nested within scope s.
//
// Example: for child := range scope.Children() { ... }
Scope.prototype.Children = function Children(): Seq<Scope> {
  return (yield_: (child: Scope) => boolean): void => {
    for (let i = 0; i < this.NumChildren(); i++) {
      if (!yield_(this.Child(i))) {
        break;
      }
    }
  };
};

// Fields returns a go1.23 iterator over the fields of a struct type.
//
// Example: for field := range s.Fields() { ... }
Struct.prototype.Fields = function Fields(): Seq<Var> {
  return (yield_: (field: Var) => boolean): void => {
    for (let i = 0; i < this.NumFields(); i++) {
      if (!yield_(this.Field(i))) {
        break;
      }
    }
  };
};

// Variables returns a go1.23 iterator over the variables of a tuple type.
//
// Example: for v := range tuple.Variables() { ... }
Tuple.prototype.Variables = function Variables(): Seq<Var> {
  return (yield_: (v: Var) => boolean): void => {
    for (let i = 0; i < this.Len(); i++) {
      if (!yield_(this.At(i))) {
        break;
      }
    }
  };
};

// Methods returns a go1.23 iterator over the methods of a method set.
//
// Example: for method := range s.Methods() { ... }
MethodSet.prototype.Methods = function Methods(): Seq<Selection> {
  return (yield_: (method: Selection) => boolean): void => {
    for (let i = 0; i < this.Len(); i++) {
      if (!yield_(this.At(i))) {
        break;
      }
    }
  };
};

// Terms returns a go1.23 iterator over the terms of a union.
//
// Example: for term := range union.Terms() { ... }
Union.prototype.Terms = function Terms(): Seq<Term> {
  return (yield_: (term: Term) => boolean): void => {
    for (let i = 0; i < this.Len(); i++) {
      if (!yield_(this.Term(i))) {
        break;
      }
    }
  };
};

// TypeParams returns a go1.23 iterator over a list of type parameters.
//
// Example: for tparam := range l.TypeParams() { ... }
TypeParamList.prototype.TypeParams = function TypeParams(): Seq<TypeParam> {
  return (yield_: (tparam: TypeParam) => boolean): void => {
    for (let i = 0; i < this.Len(); i++) {
      if (!yield_(this.At(i))) {
        break;
      }
    }
  };
};

// Types returns a go1.23 iterator over the elements of a list of types.
//
// Example: for t := range l.Types() { ... }
TypeList.prototype.Types = function Types(): Seq<Type> {
  return (yield_: (t: Type) => boolean): void => {
    for (let i = 0; i < this.Len(); i++) {
      if (!yield_(this.At(i))) {
        break;
      }
    }
  };
};

