// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// ----------------------------------------------------------------------------
// API
import { PosOf } from "../../front/ast.js";
import { registerCheckerMethod } from "./check.js";
import { TypeString } from "./typestring.js";
import { asNamed } from "./alias.js";
import { Named, setNamedInterfaceConstructor } from "./named.js";
import { Signature } from "./signature.js";
import { NewFunc, newVar, setObjectEmptyInterface, VarKind } from "./object.js";
import { _TypeSet, topTypeSet, computeInterfaceTypeSet, sortMethods } from "./typeset.js";
import { parseUnion } from "./union.js";
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
setNamedInterfaceConstructor(Interface);
registerCheckerMethod("newInterface", function newInterfaceMethod() {
    const typ = new Interface(this);
    this.needsCleanup(typ);
    return typ;
});
registerCheckerMethod("interfaceType", function interfaceType(ityp, iface, def) {
    const addEmbedded = (pos, typ) => {
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
            addEmbedded(PosOf(typExpr), parseUnion(this, typExpr));
            continue;
        }
        const name = names[0];
        const nameText = identName(name);
        if (nameText === "_") {
            this.error(name, "BlankIfaceMethod", "methods must have a unique non-blank name");
            continue;
        }
        const typ = this.typ(typExpr);
        if (!(typ instanceof Signature)) {
            if (typ.String() !== "invalid type") {
                this.errorf(typExpr, "InvalidSyntaxTree", "%s is not a method signature", typ);
            }
            continue;
        }
        if (typ.tparams !== null) {
            this.error(typExpr, "InvalidSyntaxTree", "interface methods cannot have type parameters");
        }
        let recvTyp = ityp;
        if (def !== null) {
            const named = asNamed(def.typ);
            if (named !== null) {
                recvTyp = named;
            }
        }
        typ.recv = newVar(VarKind.RecvVar, PosOf(name), this.pkg, "", recvTyp);
        const m = NewFunc(PosOf(name), this.pkg, nameText, typ);
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
        computeInterfaceTypeSet(this, PosOf(iface), ityp);
    }).describef(iface, "compute type set for %s", ityp);
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
function interfaceFields(iface) {
    const i = iface;
    return i?.Methods?.List ?? i?.MethodList ?? i?.methods?.fields ?? [];
}
function fieldNames(field) {
    const f = field;
    if (f?.Names !== undefined)
        return f.Names;
    if (f?.Name !== undefined && f.Name !== null)
        return [f.Name];
    return f?.names ?? [];
}
function fieldType(field) {
    const f = field;
    return f?.Type ?? f?.type ?? null;
}
function identName(ident) {
    const id = ident;
    return id?.Name ?? id?.Value ?? id?.name ?? "";
}
