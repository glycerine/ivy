// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { TypeString } from "./typestring.js";
import { term } from "./typeterm.js";
import { debug } from "./check.js";
import { assert } from "./util.js";
import { Identical, isTypeParam, isValid, IsInterface } from "./predicates.js";
import { Interface } from "./interface.js";
import { Typ, universeComparable } from "./universe.js";
import { BasicKind } from "./basic.js";
import { typexpr } from "./operand.js";
// A Union represents a union of terms embedded in an interface.
export class Union {
    terms;
    constructor(terms // list of syntactical terms (not a canonicalized termlist)
    ) {
        this.terms = terms;
    }
    Len() { return this.terms.length; }
    Term(i) { return this.terms[i]; }
    Underlying() { return this; }
    String() { return TypeString(this, null); }
}
// NewUnion returns a new [Union] type with the given terms.
// It is an error to create an empty union; they are syntactically not possible.
export function NewUnion(terms) {
    if (terms.length === 0) {
        throw new Error("empty union");
    }
    return new Union(terms);
}
// A Term represents a term in a [Union].
export class Term extends term {
    constructor(tilde, typ) {
        super(tilde, typ);
    }
    Tilde() { return this.tilde; }
    Type() { return this.typ; }
    String() { return super.String(); }
}
// NewTerm returns a new union term.
export function NewTerm(tilde, typ) { return new Term(tilde, typ); }
// Avoid excessive type-checking times due to quadratic termlist operations.
export const maxTermCount = 100;
// parseUnion parses uexpr as a union of expressions.
// The result is a Union type, or Typ[Invalid] for some errors.
export function parseUnion(check, uexpr) {
    const [blist, tlist] = flattenUnion([], uexpr);
    assert(blist.length === tlist.length - 1);
    const terms = [];
    let u = Typ[BasicKind.Invalid];
    for (let i = 0; i < tlist.length; i++) {
        const x = tlist[i];
        const trm = parseTilde(check, x);
        if (tlist.length === 1 && !trm.tilde) {
            // Single type. Ok to return early because all relevant
            // checks have been performed in parseTilde (no need to
            // run through term validity check below).
            return trm.typ; // typ already recorded through check.typ in parseTilde
        }
        if (terms.length >= maxTermCount) {
            if (isValid(u)) {
                check.errorf(x, "InvalidUnion", "cannot handle more than %d union terms (implementation limitation)", maxTermCount);
                u = Typ[BasicKind.Invalid];
            }
        }
        else {
            terms.push(trm);
            u = new Union(terms);
        }
        if (i > 0) {
            check.recordTypeAndValue(blist[i - 1], typexpr, u, null);
        }
    }
    if (!isValid(u)) {
        return u;
    }
    // Check validity of terms.
    // Do this check later because it requires types to be set up.
    // Note: This is a quadratic algorithm, but unions tend to be short.
    check.later(() => {
        for (let i = 0; i < terms.length; i++) {
            const t = terms[i];
            if (!isValid(t.typ)) {
                continue;
            }
            const u = t.typ.Underlying();
            const f = u instanceof Interface ? u : null;
            if (t.tilde) {
                if (f !== null) {
                    check.errorf(tlist[i], "InvalidUnion", "invalid use of ~ (%s is an interface)", t.typ);
                    continue; // don't report another error for t
                }
                if (!Identical(u, t.typ)) {
                    check.errorf(tlist[i], "InvalidUnion", "invalid use of ~ (underlying type of %s is %s)", t.typ, u);
                    continue;
                }
            }
            // Stand-alone embedded interfaces are ok and are handled by the single-type case
            // in the beginning. Embedded interfaces with tilde are excluded above. If we reach
            // here, we must have at least two terms in the syntactic term list (but not necessarily
            // in the term list of the union's type set).
            if (f !== null) {
                const tset = f.typeSet();
                switch (true) {
                    case tset.NumMethods() !== 0:
                        check.errorf(tlist[i], "InvalidUnion", "cannot use %s in union (%s contains methods)", t, t);
                        break;
                    case t.typ === universeComparable.Type():
                        check.error(tlist[i], "InvalidUnion", "cannot use comparable in union");
                        break;
                    case tset.comparable:
                        check.errorf(tlist[i], "InvalidUnion", "cannot use %s in union (%s embeds comparable)", t, t);
                        break;
                }
                continue; // terms with interface types are not subject to the no-overlap rule
            }
            // Report overlapping (non-disjoint) terms such as
            // a|a, a|~a, ~a|~a, and ~a|A (where under(A) == a).
            const j = overlappingTerm(terms.slice(0, i), t);
            if (j >= 0) {
                check.softErrorf(tlist[i], "InvalidUnion", "overlapping terms %s and %s", t, terms[j]);
            }
        }
    }).describef(uexpr, "check term validity %s", uexpr);
    return u;
}
export function parseTilde(check, tx) {
    let x = tx;
    let tilde = false;
    const op = x;
    if (op !== null && op !== undefined && op.Op === "~") {
        x = op.X;
        tilde = true;
    }
    let typ = check.typ(x);
    // Embedding stand-alone type parameters is not permitted (go.dev/issue/47127).
    // We don't need this restriction anymore if we make the underlying type of a type
    // parameter its constraint interface: if we embed a lone type parameter, we will
    // simply use its underlying type (like we do for other named, embedded interfaces),
    // and since the underlying type is an interface the embedding is well defined.
    if (isTypeParam(typ)) {
        if (tilde) {
            check.errorf(x, "MisplacedTypeParam", "type in term %s cannot be a type parameter", tx);
        }
        else {
            check.error(x, "MisplacedTypeParam", "term cannot be a type parameter");
        }
        typ = Typ[BasicKind.Invalid];
    }
    const trm = NewTerm(tilde, typ);
    if (tilde) {
        check.recordTypeAndValue(tx, typexpr, new Union([trm]), null);
    }
    return trm;
}
// overlappingTerm reports the index of the term x in terms which is
// overlapping (not disjoint) from y. The result is < 0 if there is no
// such term. The type of term y must not be an interface, and terms
// with an interface type are ignored in the terms list.
export function overlappingTerm(terms, y) {
    assert(!IsInterface(y.typ));
    for (let i = 0; i < terms.length; i++) {
        const x = terms[i];
        if (IsInterface(x.typ)) {
            continue;
        }
        // disjoint requires non-nil, non-top arguments,
        // and non-interface types as term types.
        if (debug) {
            if (x === null || x.typ === null || y === null || y.typ === null) {
                throw new Error("empty or top union term");
            }
        }
        if (!x.disjoint(y)) {
            return i;
        }
    }
    return -1;
}
// flattenUnion walks a union type expression of the form A | B | C | ...,
// extracting both the binary exprs (blist) and leaf types (tlist).
export function flattenUnion(list, x) {
    let blist = [];
    let tlist = list;
    const o = x;
    if (o !== null && o !== undefined && o.Op === "|") {
        [blist, tlist] = flattenUnion(list, o.X);
        blist.push(o);
        x = o.Y;
    }
    tlist.push(x);
    return [blist, tlist];
}
