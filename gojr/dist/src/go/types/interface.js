// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { TypeString } from "./typestring.js";
import { Named } from "./named.js";
import { Signature } from "./signature.js";
import { newVar, VarKind } from "./object.js";
import { _TypeSet, topTypeSet, computeInterfaceTypeSet, sortMethods } from "./typeset.js";
// An Interface represents an interface type.
export class Interface {
    check;
    methods = []; // ordered list of explicitly declared methods
    embeddeds = []; // ordered list of explicitly embedded elements
    embedPos = null; // positions of embedded elements; or nil (for error messages) - use pointer to save space
    implicit = false; // interface is wrapper for type set literal (non-interface T, ~T, or A|B)
    complete = false; // indicates that obj, methods, and embeddeds are set and type set can be computed
    tset = null; // type set described by this interface, computed lazily
    constructor(check = null // for error reporting; nil once type set is computed
    ) {
        this.check = check;
    }
    // typeSet returns the type set for interface t.
    typeSet() { return computeInterfaceTypeSet(this.check, 0, this); }
    // MarkImplicit marks the interface t as implicit, meaning this interface
    // corresponds to a constraint literal such as ~T or A|B without explicit
    // interface embedding. MarkImplicit should be called before any concurrent use
    // of implicit interfaces.
    MarkImplicit() {
        this.implicit = true;
    }
    // NumExplicitMethods returns the number of explicitly declared methods of interface t.
    NumExplicitMethods() { return this.methods.length; }
    // ExplicitMethod returns the i'th explicitly declared method of interface t for 0 <= i < t.NumExplicitMethods().
    // The methods are ordered by their unique [Id].
    ExplicitMethod(i) { return this.methods[i]; }
    // NumEmbeddeds returns the number of embedded types in t.
    NumEmbeddeds() { return this.embeddeds.length; }
    // Embedded returns the i'th embedded defined (*[Named]) type of interface t for 0 <= i < t.NumEmbeddeds().
    // The result is nil if the i'th embedded type is not a defined type.
    //
    // Deprecated: Use [Interface.EmbeddedType] which is not restricted to defined (*[Named]) types.
    Embedded(i) { return this.embeddeds[i] instanceof Named ? this.embeddeds[i] : null; }
    // EmbeddedType returns the i'th embedded type of interface t for 0 <= i < t.NumEmbeddeds().
    EmbeddedType(i) { return this.embeddeds[i]; }
    // NumMethods returns the total number of methods of interface t.
    NumMethods() { return this.typeSet().NumMethods(); }
    // Method returns the i'th method of interface t for 0 <= i < t.NumMethods().
    // The methods are ordered by their unique Id.
    Method(i) { return this.typeSet().Method(i); }
    // Empty reports whether t is the empty interface.
    Empty() { return this.typeSet().IsAll(); }
    // IsComparable reports whether each type in interface t's type set is comparable.
    IsComparable() { return this.typeSet().IsComparable(null); }
    // IsMethodSet reports whether the interface t is fully described by its method
    // set.
    IsMethodSet() { return this.typeSet().IsMethodSet(); }
    // IsImplicit reports whether the interface t is a wrapper for a type set literal.
    IsImplicit() { return this.implicit; }
    // Complete computes the interface's type set. It must be called by users of
    // [NewInterfaceType] and [NewInterface] after the interface's embedded types are
    // fully defined and before using the interface type in any way other than to
    // form other types. The interface must not contain duplicate methods or a
    // panic occurs. Complete returns the receiver.
    //
    // Interface types that have been completed are safe for concurrent use.
    Complete() {
        if (!this.complete) {
            this.complete = true;
        }
        this.typeSet(); // checks if t.tset is already set
        return this;
    }
    Underlying() { return this; }
    String() { return TypeString(this, null); }
    // ----------------------------------------------------------------------------
    // Implementation
    cleanup() {
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
export function NewInterface(methods, embeddeds) {
    const tnames = [];
    for (const t of embeddeds ?? []) {
        tnames.push(t);
    }
    return NewInterfaceType(methods, tnames);
}
// NewInterfaceType returns a new interface for the given methods and embedded
// types. NewInterfaceType takes ownership of the provided methods and may
// modify their types by setting missing receivers.
export function NewInterfaceType(methods, embeddeds) {
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
