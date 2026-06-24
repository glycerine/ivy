// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { Id } from "./object.js";
import { Selection, SelectionKind } from "./selection.js";
import { asNamed } from "./alias.js";
import { Interface } from "./interface.js";
import { Struct } from "./struct.js";
import { concat, consolidateMultiples, deref, embeddedType, instanceLookup, isPointer } from "./lookup.js";
import { IsInterface } from "./predicates.js";
// A MethodSet is an ordered set of concrete or abstract (interface) methods;
// a method is a [MethodVal] selection, and they are ordered by ascending m.Obj().Id().
// The zero value for a MethodSet is a ready-to-use empty method set.
export class MethodSet {
    list;
    constructor(list = []) {
        this.list = list;
    }
    String() {
        if (this.Len() === 0) {
            return "MethodSet {}";
        }
        const buf = [];
        buf.push("MethodSet {\n");
        for (const f of this.list) {
            buf.push("\t");
            buf.push(f.String());
            buf.push("\n");
        }
        buf.push("}\n");
        return buf.join("");
    }
    // Len returns the number of methods in s.
    Len() { return this.list.length; }
    // At returns the i'th method in s for 0 <= i < s.Len().
    At(i) { return this.list[i]; }
    // Lookup returns the method with matching package and name, or nil if not found.
    Lookup(pkg, name) {
        if (this.Len() === 0) {
            return null;
        }
        const key = Id(pkg, name);
        let lo = 0;
        let hi = this.list.length;
        while (lo < hi) {
            const mid = Math.floor((lo + hi) / 2);
            const m = this.list[mid];
            if (m.obj.Id() >= key) {
                hi = mid;
            }
            else {
                lo = mid + 1;
            }
        }
        const i = lo;
        if (i < this.list.length) {
            const m = this.list[i];
            if (m.obj.Id() === key) {
                return m;
            }
        }
        return null;
    }
}
// Shared empty method set.
export const emptyMethodSet = new MethodSet();
// NewMethodSet returns the method set for the given type T.
// It always returns a non-nil method set, even if it is empty.
export function NewMethodSet(T) {
    // WARNING: The code in this function is extremely subtle - do not modify casually!
    //          This function and lookupFieldOrMethod should be kept in sync.
    // TODO(rfindley) confirm that this code is in sync with lookupFieldOrMethod
    //                with respect to type params.
    // Methods cannot be associated with a named pointer type.
    // (spec: "The type denoted by T is called the receiver base type;
    // it must not be a pointer or interface type and it must be declared
    // in the same package as the method.").
    const namedT = asNamed(T);
    if (namedT !== null && isPointer(namedT)) {
        return emptyMethodSet;
    }
    // method set up to the current depth, allocated lazily
    let base = new methodSet();
    const [typ0, isPtr] = deref(T);
    // *typ where typ is an interface has no methods.
    if (isPtr && IsInterface(typ0)) {
        return emptyMethodSet;
    }
    // Start with typ as single entry at shallowest depth.
    let current = [new embeddedType(typ0, null, isPtr, false)];
    // seen tracks named types that we have seen already, allocated lazily.
    const seen = new instanceLookup();
    // collect methods at current depth
    while (current.length > 0) {
        let next = []; // embedded types found at current depth
        // field and method sets at current depth, indexed by names (Id's), and allocated lazily
        const fset = new Map(); // we only care about the field names
        let mset = new methodSet();
        for (const e of current) {
            const typ = e.typ;
            // If we have a named type, we may have associated methods.
            // Look for those first.
            const named = asNamed(typ);
            if (named !== null) {
                const alt = seen.lookup(named);
                if (alt !== null) {
                    // We have seen this type before, at a more shallow depth
                    // (note that multiples of this type at the current depth
                    // were consolidated before). The type at that depth shadows
                    // this same type at the current depth, so we can ignore
                    // this one.
                    continue;
                }
                seen.add(named);
                for (let i = 0; i < named.NumMethods(); i++) {
                    mset = mset.addOne(named.Method(i), concat(e.index, i), e.indirect, e.multiples);
                }
            }
            const u = typ.Underlying();
            if (u instanceof Struct) {
                for (let i = 0; i < (u.fields?.length ?? 0); i++) {
                    const f = u.fields[i];
                    fset.set(f.Id(), true);
                    // Embedded fields are always of the form T or *T where
                    // T is a type name. If typ appeared multiple times at
                    // this depth, f.Type appears multiple times at the next
                    // depth.
                    if (f.embedded) {
                        const [typ2, isPtr2] = deref(f.typ);
                        // TODO(gri) optimization: ignore types that can't
                        // have fields or methods (only Named, Struct, and
                        // Interface types need to be considered).
                        next.push(new embeddedType(typ2, concat(e.index, i), e.indirect || isPtr2, e.multiples));
                    }
                }
            }
            else if (u instanceof Interface) {
                mset = mset.add(u.typeSet().methods, e.index, true, e.multiples);
            }
        }
        // Add methods and collisions at this depth to base if no entries with matching
        // names exist already.
        for (const [k, mm] of mset.entries()) {
            let m = mm;
            if (!base.has(k)) {
                // Fields collide with methods of the same name at this depth.
                if (fset.get(k)) {
                    m = null; // collision
                }
                base.set(k, m);
            }
        }
        // Add all (remaining) fields at this depth as collisions (since they will
        // hide any method further down) if no entries with matching names exist already.
        for (const k of fset.keys()) {
            if (!base.has(k)) {
                base.set(k, null); // collision
            }
        }
        current = consolidateMultiples(next);
    }
    if (base.size === 0) {
        return emptyMethodSet;
    }
    // collect methods
    const list = [];
    for (const m of base.values()) {
        if (m !== null) {
            m.recv = T;
            list.push(m);
        }
    }
    // sort by unique name
    list.sort((a, b) => a.obj.Id().localeCompare(b.obj.Id()));
    return new MethodSet(list);
}
// A methodSet is a set of methods and name collisions.
// A collision indicates that multiple methods with the
// same unique id, or a field with that id appeared.
export class methodSet extends Map {
    // Add adds all functions in list to the method set s.
    // If multiples is set, every function in list appears multiple times
    // and is treated as a collision.
    add(list, index, indirect, multiples) {
        if (list.length === 0) {
            return this;
        }
        for (let i = 0; i < list.length; i++) {
            const f = list[i];
            this.addOne(f, concat(index, i), indirect, multiples);
        }
        return this;
    }
    addOne(f, index, indirect, multiples) {
        const key = f.Id();
        // if f is not in the set, add it
        if (!multiples) {
            // TODO(gri) A found method may not be added because it's not in the method set
            // (!indirect && f.hasPtrRecv()). A 2nd method on the same level may be in the method
            // set and may not collide with the first one, thus leading to a false positive.
            // Is that possible? Investigate.
            if (!this.has(key) && (indirect || !f.hasPtrRecv())) {
                this.set(key, new Selection(SelectionKind.MethodVal, null, f, index, indirect));
                return this;
            }
        }
        this.set(key, null); // collision
        return this;
    }
}
