// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ----------------------------------------------------------------------------
// API

import type { Checker } from "./check.js";
import type { Pos } from "./token.js";
import type { Type } from "./type.js";
import { TypeString } from "./typestring.js";
import { Named } from "./named.js";
import { Signature } from "./signature.js";
import { newVar, VarKind, type Func } from "./object.js";

export class _TypeSet {
  public constructor(
    public methods: Func[] = [],
    public all = true,
    public comparable = false,
    public methodSet = true
  ) {}

  public NumMethods(): number { return this.methods.length; }
  public Method(i: number): Func { return this.methods[i]!; }
  public IsAll(): boolean { return this.all; }
  public IsComparable(_seen: unknown): boolean { return this.comparable; }
  public IsMethodSet(): boolean { return this.methodSet; }
}

export const topTypeSet = new _TypeSet([], true, false, true);

// An Interface represents an interface type.
export class Interface implements Type {
  public methods: Func[] = []; // ordered list of explicitly declared methods
  public embeddeds: Type[] = []; // ordered list of explicitly embedded elements
  public embedPos: Pos[] | null = null; // positions of embedded elements; or nil (for error messages) - use pointer to save space
  public implicit = false; // interface is wrapper for type set literal (non-interface T, ~T, or A|B)
  public complete = false; // indicates that obj, methods, and embeddeds are set and type set can be computed

  public tset: _TypeSet | null = null; // type set described by this interface, computed lazily

  public constructor(
    public check: Checker | null = null // for error reporting; nil once type set is computed
  ) {}

  // typeSet returns the type set for interface t.
  public typeSet(): _TypeSet { return computeInterfaceTypeSet(this.check, 0, this); }

  // MarkImplicit marks the interface t as implicit, meaning this interface
  // corresponds to a constraint literal such as ~T or A|B without explicit
  // interface embedding. MarkImplicit should be called before any concurrent use
  // of implicit interfaces.
  public MarkImplicit(): void {
    this.implicit = true;
  }

  // NumExplicitMethods returns the number of explicitly declared methods of interface t.
  public NumExplicitMethods(): number { return this.methods.length; }

  // ExplicitMethod returns the i'th explicitly declared method of interface t for 0 <= i < t.NumExplicitMethods().
  // The methods are ordered by their unique [Id].
  public ExplicitMethod(i: number): Func { return this.methods[i]!; }

  // NumEmbeddeds returns the number of embedded types in t.
  public NumEmbeddeds(): number { return this.embeddeds.length; }

  // Embedded returns the i'th embedded defined (*[Named]) type of interface t for 0 <= i < t.NumEmbeddeds().
  // The result is nil if the i'th embedded type is not a defined type.
  //
  // Deprecated: Use [Interface.EmbeddedType] which is not restricted to defined (*[Named]) types.
  public Embedded(i: number): Named | null { return this.embeddeds[i] instanceof Named ? this.embeddeds[i] : null; }

  // EmbeddedType returns the i'th embedded type of interface t for 0 <= i < t.NumEmbeddeds().
  public EmbeddedType(i: number): Type { return this.embeddeds[i]!; }

  // NumMethods returns the total number of methods of interface t.
  public NumMethods(): number { return this.typeSet().NumMethods(); }

  // Method returns the i'th method of interface t for 0 <= i < t.NumMethods().
  // The methods are ordered by their unique Id.
  public Method(i: number): Func { return this.typeSet().Method(i); }

  // Empty reports whether t is the empty interface.
  public Empty(): boolean { return this.typeSet().IsAll(); }

  // IsComparable reports whether each type in interface t's type set is comparable.
  public IsComparable(): boolean { return this.typeSet().IsComparable(null); }

  // IsMethodSet reports whether the interface t is fully described by its method
  // set.
  public IsMethodSet(): boolean { return this.typeSet().IsMethodSet(); }

  // IsImplicit reports whether the interface t is a wrapper for a type set literal.
  public IsImplicit(): boolean { return this.implicit; }

  // Complete computes the interface's type set. It must be called by users of
  // [NewInterfaceType] and [NewInterface] after the interface's embedded types are
  // fully defined and before using the interface type in any way other than to
  // form other types. The interface must not contain duplicate methods or a
  // panic occurs. Complete returns the receiver.
  //
  // Interface types that have been completed are safe for concurrent use.
  public Complete(): Interface {
    if (!this.complete) {
      this.complete = true;
    }
    this.typeSet(); // checks if t.tset is already set
    return this;
  }

  public Underlying(): Type { return this; }
  public String(): string { return TypeString(this, null); }

  // ----------------------------------------------------------------------------
  // Implementation

  public cleanup(): void {
    this.typeSet(); // any interface that escapes type checking must be safe for concurrent use
    this.check = null;
    this.embedPos = null;
  }
}

// emptyInterface represents the empty (completed) interface
export const emptyInterface = new Interface(null);
emptyInterface.complete = true;
emptyInterface.tset = topTypeSet;

// NewInterface returns a new interface for the given methods and embedded types.
// NewInterface takes ownership of the provided methods and may modify their types
// by setting missing receivers.
//
// Deprecated: Use NewInterfaceType instead which allows arbitrary embedded types.
export function NewInterface(methods: Func[] | null, embeddeds: Named[] | null): Interface {
  const tnames: Type[] = [];
  for (const t of embeddeds ?? []) {
    tnames.push(t);
  }
  return NewInterfaceType(methods, tnames);
}

// NewInterfaceType returns a new interface for the given methods and embedded
// types. NewInterfaceType takes ownership of the provided methods and may
// modify their types by setting missing receivers.
export function NewInterfaceType(methods: Func[] | null, embeddeds: Type[] | null): Interface {
  if ((methods?.length ?? 0) === 0 && (embeddeds?.length ?? 0) === 0) {
    return emptyInterface;
  }

  // set method receivers if necessary
  const typ = new Interface(null);
  for (const m of methods ?? []) {
    if (m.typ instanceof Signature && m.typ.recv === null) {
      m.typ.recv = newVar(VarKind.RecvVar, m.pos, m.pkg, "", typ);
    }
  }

  // sort for API stability
  sortMethods(methods ?? []);

  typ.methods = methods ?? [];
  typ.embeddeds = embeddeds ?? [];
  typ.complete = true;

  return typ;
}

export function computeInterfaceTypeSet(_check: Checker | null, _pos: Pos, ityp: Interface): _TypeSet {
  if (ityp.tset === null) {
    ityp.tset = new _TypeSet(ityp.methods, ityp.embeddeds.length === 0, false, true);
  }
  return ityp.tset;
}

export function sortMethods(methods: Func[]): void {
  methods.sort((a, b) => a.Id().localeCompare(b.Id()));
}
