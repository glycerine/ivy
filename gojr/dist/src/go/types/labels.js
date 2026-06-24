// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { Checker, debug } from "./check.js";
import { NewLabel } from "./object.js";
import { NewScope, resolve } from "./scope.js";
import { assert } from "./util.js";
// labels checks correct label use in body.
Checker.prototype.labels = function labels(body) {
    const b = body;
    const all = NewScope(null, b.Pos?.() ?? 0, b.End?.() ?? 0, "label");
    const fwdJumps = this.blockBranches(all, null, null, b.List ?? []);
    for (const jmp of fwdJumps) {
        const j = jmp;
        let msg;
        let code;
        const name = j.Label?.Name ?? "";
        const alt = all.Lookup(name);
        if (alt !== null) {
            msg = "goto %s jumps into block";
            code = "JumpIntoBlock";
            alt.used = true;
        }
        else {
            msg = "label %s not declared";
            code = "UndeclaredLabel";
        }
        this.errorf(j.Label, code, msg, name);
    }
    for (const name of all.Names()) {
        const obj = resolve(name, all.elems.get(name) ?? null);
        const lbl = obj;
        if (lbl !== null && !lbl.used) {
            this.softErrorf(lbl, "UnusedLabel", "label %s declared and not used", lbl.name);
        }
    }
};
// A block tracks label declarations in a block and its enclosing blocks.
export class block {
    parent;
    lstmt;
    labels = null;
    constructor(parent, lstmt) {
        this.parent = parent;
        this.lstmt = lstmt;
    }
    // insert records a new label declaration for the current block.
    // The label must not have been declared before in any block.
    insert(s) {
        const name = s.Label?.Name ?? "";
        if (debug) {
            assert(this.gotoTarget(name) === null);
        }
        let labels = this.labels;
        if (labels === null) {
            labels = new Map();
            this.labels = labels;
        }
        labels.set(name, s);
    }
    // gotoTarget returns the labeled statement in the current
    // or an enclosing block with the given label name, or nil.
    gotoTarget(name) {
        for (let s = this; s !== null; s = s.parent) {
            const t = s.labels?.get(name);
            if (t !== undefined) {
                return t;
            }
        }
        return null;
    }
    // enclosingTarget returns the innermost enclosing labeled
    // statement with the given label name, or nil.
    enclosingTarget(name) {
        for (let s = this; s !== null; s = s.parent) {
            const t = s.lstmt;
            if (t !== null && t.Label?.Name === name) {
                return t;
            }
        }
        return null;
    }
}
// blockBranches processes a block's statement list and returns the set of outgoing forward jumps.
Checker.prototype.blockBranches = function blockBranches(all, parent, lstmt, list) {
    const b = new block(parent, lstmt);
    const fwdJumps = [];
    const stmtBranches = (lstmt, s) => {
        const stmt = s;
        switch (stmt.kind) {
            case "LabeledStmt": {
                const name = stmt.Label?.Name ?? "";
                if (name !== "_") {
                    const lbl = NewLabel(stmt.Label?.Pos?.() ?? 0, this.pkg, name);
                    const alt = all.Insert(lbl);
                    if (alt !== null) {
                        const err = this.newError("DuplicateLabel");
                        err.soft = true;
                        err.addf(lbl, "label %s already declared", name);
                        err.addAltDecl(alt);
                        err.report();
                    }
                    else {
                        b.insert(s);
                        this.recordDef(stmt.Label, lbl);
                    }
                    lstmt = s;
                }
                stmtBranches(lstmt, stmt.Stmt);
                break;
            }
            case "BranchStmt": {
                if (stmt.Label === null || stmt.Label === undefined) {
                    return;
                }
                const name = stmt.Label.Name ?? "";
                if (stmt.Tok === "GOTO" && b.gotoTarget(name) === null) {
                    fwdJumps.push(s);
                    return;
                }
                const obj = all.Lookup(name);
                if (obj !== null) {
                    obj.used = true;
                    this.recordUse(stmt.Label, obj);
                }
                break;
            }
            case "BlockStmt":
                fwdJumps.push(...this.blockBranches(all, b, lstmt, stmt.Body?.List ?? []));
                break;
            case "IfStmt":
                stmtBranches(lstmt, stmt.Body);
                if (stmt.Else !== null && stmt.Else !== undefined) {
                    stmtBranches(lstmt, stmt.Else);
                }
                break;
            default:
                if (stmt.Body?.List) {
                    fwdJumps.push(...this.blockBranches(all, b, null, stmt.Body.List));
                }
        }
    };
    for (const s of list) {
        stmtBranches(null, s);
    }
    return fwdJumps;
};
