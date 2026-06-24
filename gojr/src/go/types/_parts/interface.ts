// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ----------------------------------------------------------------------------
// API

import { PosOf } from "../../front/ast.js";
import { registerCheckerMethod, type Checker } from "./check.js";
import type { Pos } from "./token.js";
import type { Type } from "./type.js";
import { TypeString } from "./typestring.js";
import { asNamed } from "./alias.js";
import { Named, setNamedInterfaceConstructor } from "./named.js";
import { Signature } from "./signature.js";
import { NewFunc, newVar, setObjectEmptyInterface, TypeName, VarKind, type Func } from "./object.js";
import { _TypeSet, topTypeSet, computeInterfaceTypeSet, sortMethods } from "./typeset.js";
import { parseUnion } from "./union.js";

declare module "./check.js" {
  interface Checker {
    newInterface(): Interface;
    interfaceType(ityp: Interface, iface: unknown, def: TypeName | null): void;
    typ(e: unknown): Type;
    error(at: unknown, code: unknown, msg: string): void;
    errorf(at: unknown, code: unknown, format: string, ...args: unknown[]): void;
    recordDef(id: unknown, obj: Func): void;
  }
}

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

setNamedInterfaceConstructor(Interface);

registerCheckerMethod("newInterface", function newInterfaceMethod(): Interface {
  const typ = new Interface(this);
  this.needsCleanup(typ);
  return typ;
});

registerCheckerMethod("interfaceType", function interfaceType(ityp: Interface, iface: unknown, def: TypeName | null): void {
  const addEmbedded = (pos: Pos, typ: Type): void => {
    ityp.embeddeds.push(typ);
    if (ityp.embedPos === null) {
      ityp.embedPos = [];
    }
    ityp.embedPos.push(pos);
  };

  for (const f of interfaceFields(iface)) {
    const names = fieldNames(f);
    const typExpr = fieldType(f);
    if (names.length === 0) {
      addEmbedded(PosOf(typExpr as never), parseUnion(this, typExpr));
      continue;
    }

    const name = names[0]!;
    const nameText = identName(name);
    if (nameText === "_") {
      this.error(name, "BlankIfaceMethod", "methods must have a unique non-blank name");
      continue;
    }

    const typ = this.typ(typExpr);
    if (!(typ instanceof Signature)) {
      if ((typ as Type).String() !== "invalid type") {
        this.errorf(typExpr, "InvalidSyntaxTree", "%s is not a method signature", typ);
      }
      continue;
    }

    if (typ.tparams !== null) {
      this.error(typExpr, "InvalidSyntaxTree", "interface methods cannot have type parameters");
    }

    let recvTyp: Type = ityp;
    if (def !== null) {
      const named = asNamed(def.typ);
      if (named !== null) {
        recvTyp = named;
      }
    }
    typ.recv = newVar(VarKind.RecvVar, PosOf(name as never), this.pkg, "", recvTyp);

    const m = NewFunc(PosOf(name as never), this.pkg, nameText, typ);
    this.recordDef(name, m);
    ityp.methods.push(m);
  }

  ityp.complete = true;

  if (ityp.methods.length === 0 && ityp.embeddeds.length === 0) {
    ityp.tset = topTypeSet;
    return;
  }

  sortMethods(ityp.methods);

  this.later(() => {
    computeInterfaceTypeSet(this, PosOf(iface as never), ityp);
  }).describef(iface as never, "compute type set for %s", ityp);
});

// emptyInterface represents the empty (completed) interface
export const emptyInterface = new Interface(null);
emptyInterface.complete = true;
emptyInterface.tset = topTypeSet;
setObjectEmptyInterface(emptyInterface);

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

export { _TypeSet, topTypeSet, computeInterfaceTypeSet, sortMethods };

function interfaceFields(iface: unknown): unknown[] {
  const i = iface as { Methods?: { List?: unknown[] }; MethodList?: unknown[]; methods?: { fields?: unknown[] } } | null;
  return i?.Methods?.List ?? i?.MethodList ?? i?.methods?.fields ?? [];
}

function fieldNames(field: unknown): unknown[] {
  const f = field as { Names?: unknown[]; Name?: unknown; names?: unknown[] } | null;
  if (f?.Names !== undefined) return f.Names;
  if (f?.Name !== undefined && f.Name !== null) return [f.Name];
  return f?.names ?? [];
}

function fieldType(field: unknown): unknown {
  const f = field as { Type?: unknown; type?: unknown } | null;
  return f?.Type ?? f?.type ?? null;
}

function identName(ident: unknown): string {
  const id = ident as { Name?: string; Value?: string; name?: string } | null;
  return id?.Name ?? id?.Value ?? id?.name ?? "";
}
