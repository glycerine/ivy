// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { ParseExprFrom } from "../../front/parser.js";
import { Info, TypeAndValue } from "./api.js";
import { NewChecker, bailout, debug, nopos } from "./check.js";
import { operand, invalid } from "./operand.js";
import { NewPackage } from "./package.js";
import { Universe } from "./universe.js";
import "./recording.js";
import "./scope2.js";
// Eval returns the type and, if constant, the value for the
// expression expr, evaluated at position pos of package pkg,
// which must have been derived from type-checking an AST with
// complete position information relative to the provided file
// set.
//
// The meaning of the parameters fset, pkg, and pos is the
// same as in [CheckExpr]. An error is returned if expr cannot
// be parsed successfully, or the resulting expr AST cannot be
// type-checked.
export function Eval(fset, pkg, pos, expr) {
    // parse expressions
    const [node, err0] = ParseExprFrom(fset, "eval", expr, 0);
    if (err0 !== undefined) {
        return [new TypeAndValue(invalid, null, null), err0];
    }
    if (node === undefined) {
        return [new TypeAndValue(invalid, null, null), new Error("expression parse produced no node")];
    }
    const info = new Info();
    info.Types = new Map();
    const err = CheckExpr(fset, pkg, pos, node, info);
    return [info.Types.get(node) ?? new TypeAndValue(invalid, null, null), err];
}
// CheckExpr type checks the expression expr as if it had appeared at position
// pos of package pkg. [Type] information about the expression is recorded in
// info. The expression may be an identifier denoting an uninstantiated generic
// function or type.
//
// If pkg == nil, the [Universe] scope is used and the provided
// position pos is ignored. If pkg != nil, and pos is invalid,
// the package scope is used. Otherwise, pos must belong to the
// package.
//
// An error is returned if pos is not within the package or
// if the node cannot be type-checked.
//
// Note: [Eval] and CheckExpr should not be used instead of running Check
// to compute types and values, but in addition to Check, as these
// functions ignore the context in which an expression is used (e.g., an
// assignment). Thus, top-level untyped constants will return an
// untyped type rather than the respective context-specific type.
export function CheckExpr(fset, pkg, pos, expr, info) {
    // determine scope
    let scope = null;
    let checkPkg = pkg;
    if (pkg === null) {
        scope = Universe;
        pos = nopos;
        checkPkg = NewPackage("", "");
    }
    else if (pos === nopos) {
        scope = pkg.scope;
    }
    else {
        // The package scope extent (position information) may be
        // incorrect (files spread across a wide range of fset
        // positions) - ignore it and just consider its children
        // (file scopes).
        for (const fscope of pkg.scope.children) {
            scope = fscope.Innermost(pos);
            if (scope !== null) {
                break;
            }
        }
        if (scope === null || debug) {
            let s = scope;
            while (s !== null && s !== pkg.scope) {
                s = s.parent;
            }
            // s == nil || s == pkg.scope
            if (s === null) {
                return new Error(`no position ${pos} found in package ${pkg.name}`);
            }
        }
    }
    // initialize checker
    const check = NewChecker(null, fset, checkPkg, info);
    check.scope = scope;
    check.exprPos = pos;
    const err = { value: null };
    try {
        // evaluate node
        const x = new operand();
        check.rawExpr(null, x, expr, null, true); // allow generic expressions
        check.processDelayed(0); // incl. all functions
        check.recordUntyped();
    }
    catch (p) {
        if (p instanceof bailout) {
            check.handleBailout(err);
        }
        else {
            throw p;
        }
    }
    return err.value;
}
