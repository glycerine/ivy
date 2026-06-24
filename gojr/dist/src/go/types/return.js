// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// This file implements isTerminating.
import { Unparen } from "../../front/ast.js";
import { TokenKind } from "../../front/token.js";
import { registerCheckerMethod } from "./check.js";
// isTerminating reports if s is a terminating statement.
// If s is labeled, label is the label name; otherwise s
// is "".
registerCheckerMethod("isTerminating", function isTerminating(s, label) {
    switch (s.kind) {
        default:
            throw new Error("unreachable");
        case "BadStmt":
        case "DeclStmt":
        case "EmptyStmt":
        case "SendStmt":
        case "IncDecStmt":
        case "AssignStmt":
        case "GoStmt":
        case "DeferStmt":
        case "RangeStmt":
            // no chance
            break;
        case "LabeledStmt":
            return this.isTerminating(s.stmt, s.label.name);
        case "ExprStmt": {
            // calling the predeclared (possibly parenthesized) panic() function is terminating
            const call = Unparen(s.expr);
            if (call.kind === "CallExpr" && this.isPanic?.get(call)) {
                return true;
            }
            break;
        }
        case "ReturnStmt":
            return true;
        case "BranchStmt":
            if (s.token === TokenKind.Goto || s.token === TokenKind.Fallthrough) {
                return true;
            }
            break;
        case "BlockStmt":
            return this.isTerminatingList(s.statements, "");
        case "IfStmt":
            if (s.else !== undefined &&
                this.isTerminating(s.body, "") &&
                this.isTerminating(s.else, "")) {
                return true;
            }
            break;
        case "SwitchStmt":
            return this.isTerminatingSwitch(s.body, label);
        case "TypeSwitchStmt":
            return this.isTerminatingSwitch(s.body, label);
        case "SelectStmt":
            for (const clause of s.body) {
                const cc = clause;
                if (!this.isTerminatingList(cc.body, "") || hasBreakList(cc.body, label, true)) {
                    return false;
                }
            }
            return true;
        case "ForStmt":
            if (s.condition === undefined && !hasBreak(s.body, label, true)) {
                return true;
            }
            break;
        case "CaseClause":
        case "CommClause":
        case "UnsupportedStmt":
            // no chance
            break;
    }
    return false;
});
registerCheckerMethod("isTerminatingList", function isTerminatingList(list, label) {
    // trailing empty statements are permitted - skip them
    for (let i = list.length - 1; i >= 0; i--) {
        if (list[i].kind !== "EmptyStmt") {
            return this.isTerminating(list[i], label);
        }
    }
    return false; // all statements are empty
});
registerCheckerMethod("isTerminatingSwitch", function isTerminatingSwitch(body, label) {
    let hasDefault = false;
    for (const s of body) {
        const cc = s;
        if (cc.list.length === 0 || cc.default) {
            hasDefault = true;
        }
        if (!this.isTerminatingList(cc.body, "") || hasBreakList(cc.body, label, true)) {
            return false;
        }
    }
    return hasDefault;
});
// TODO(gri) For nested breakable statements, the current implementation of hasBreak
// will traverse the same subtree repeatedly, once for each label. Replace
// with a single-pass label/break matching phase.
// hasBreak reports if s is or contains a break statement
// referring to the label-ed statement or implicit-ly the
// closest outer breakable statement.
export function hasBreak(s, label, implicit) {
    switch (s.kind) {
        default:
            throw new Error("unreachable");
        case "BadStmt":
        case "DeclStmt":
        case "EmptyStmt":
        case "ExprStmt":
        case "SendStmt":
        case "IncDecStmt":
        case "AssignStmt":
        case "GoStmt":
        case "DeferStmt":
        case "ReturnStmt":
            // no chance
            break;
        case "LabeledStmt":
            return hasBreak(s.stmt, label, implicit);
        case "BranchStmt":
            if (s.token === TokenKind.Break) {
                if (s.label === undefined) {
                    return implicit;
                }
                if (s.label.name === label) {
                    return true;
                }
            }
            break;
        case "BlockStmt":
            return hasBreakList(s.statements, label, implicit);
        case "IfStmt":
            if (hasBreak(s.body, label, implicit) ||
                s.else !== undefined && hasBreak(s.else, label, implicit)) {
                return true;
            }
            break;
        case "CaseClause":
            return hasBreakList(s.body, label, implicit);
        case "SwitchStmt":
            if (label !== "" && hasBreakList(s.body, label, false)) {
                return true;
            }
            break;
        case "TypeSwitchStmt":
            if (label !== "" && hasBreakList(s.body, label, false)) {
                return true;
            }
            break;
        case "CommClause":
            return hasBreakList(s.body, label, implicit);
        case "SelectStmt":
            if (label !== "" && hasBreakList(s.body, label, false)) {
                return true;
            }
            break;
        case "ForStmt":
            if (label !== "" && hasBreak(s.body, label, false)) {
                return true;
            }
            break;
        case "RangeStmt":
            if (label !== "" && hasBreak(s.body, label, false)) {
                return true;
            }
            break;
        case "UnsupportedStmt":
            // no chance
            break;
    }
    return false;
}
export function hasBreakList(list, label, implicit) {
    for (const s of list) {
        if (hasBreak(s, label, implicit)) {
            return true;
        }
    }
    return false;
}
