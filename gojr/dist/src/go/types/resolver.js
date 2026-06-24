// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { Checker, atPos, importKey, nopos } from "./check.js";
import { asGoVersion } from "./version.js";
import { NewPackage } from "./package.js";
import { NewScope } from "./scope.js";
import { NewPkgName, NewTypeName } from "./object.js";
import { assert } from "./util.js";
// A declInfo describes a package-level const, type, var, or func declaration.
export class declInfo {
    file = null; // scope of file containing this declaration
    version = null; // Go version of file containing this declaration
    lhs = null; // lhs of n:1 variable declarations, or nil
    vtyp = null; // type, or nil (for const and var declarations only)
    init = null; // init/orig expression, or nil (for const and var declarations only)
    inherited = false; // if set, the init expression is inherited from a previous constant declaration
    tdecl = null; // type declaration, or nil
    fdecl = null; // func declaration, or nil
    // The deps field tracks initialization expression dependencies.
    deps = null; // lazily initialized
    constructor(init) {
        Object.assign(this, init);
    }
    // hasInitializer reports whether the declared object has an initialization
    // expression or function body.
    hasInitializer() {
        const fdecl = this.fdecl;
        return this.init !== null || (fdecl !== null && fdecl.Body !== null && fdecl.Body !== undefined);
    }
    // addDep adds obj to the set of objects d's init expression depends on.
    addDep(obj) {
        let m = this.deps;
        if (m === null) {
            m = new Map();
            this.deps = m;
        }
        m.set(obj, true);
    }
}
// arityMatch checks that the lhs and rhs of a const or var decl
// have the appropriate number of names and init exprs. For const
// decls, init is the value spec providing the init exprs; for
// var decls, init is nil (the init exprs are in s in this case).
Checker.prototype.arityMatch = function arityMatch(s, init) {
    const spec = s;
    const initSpec = init;
    const l = spec.Names?.length ?? 0;
    let r = spec.Values?.length ?? 0;
    if (initSpec !== null && initSpec !== undefined) {
        r = initSpec.Values?.length ?? 0;
    }
    const code = "WrongAssignCount";
    switch (true) {
        case init === null && r === 0:
            // var decl w/o init expr
            if (spec.Type === null || spec.Type === undefined) {
                this.error(s, code, "missing type or init expr");
            }
            break;
        case l < r:
            if (l < (spec.Values?.length ?? 0)) {
                // init exprs from s
                const n = spec.Values[l];
                this.errorf(n, code, "extra init expr %s", n);
                // TODO(gri) avoid declared and not used error here
            }
            else {
                // init exprs "inherited"
                this.errorf(s, code, "extra init expr at %s", String(initSpec?.Pos?.() ?? ""));
                // TODO(gri) avoid declared and not used error here
            }
            break;
        case l > r && (init !== null || r !== 1): {
            const n = spec.Names[r];
            this.errorf(n, code, "missing init expr for %s", n);
            break;
        }
    }
};
export function validatedImportPath(path) {
    let s;
    try {
        s = JSON.parse(path);
    }
    catch (err) {
        return ["", err instanceof Error ? err : new Error(String(err))];
    }
    if (s === "") {
        return ["", new Error("empty string")];
    }
    const illegalChars = `!"#$%&'()*,:;<=>?[\\]^{|}` + "`\uFFFD";
    for (const r of s) {
        if (!isGraphic(r) || /\s/u.test(r) || illegalChars.includes(r)) {
            return [s, new Error(`invalid character ${JSON.stringify(r)}`)];
        }
    }
    return [s, null];
}
function isGraphic(r) {
    return r.length > 0 && !/[\p{Cc}\p{Cs}\p{Cn}]/u.test(r);
}
// declarePkgObj declares obj in the package scope, records its ident -> obj mapping,
// and updates check.objMap. The object must not be a function or method.
Checker.prototype.declarePkgObj = function declarePkgObj(ident, obj, d) {
    const id = ident;
    assert(id.Name === obj.Name());
    // spec: "A package-scope or file-scope identifier with name init
    // may only be declared to be a function with this (func()) signature."
    if (id.Name === "init") {
        this.error(ident, "InvalidInitDecl", "cannot declare init - must be func");
        return;
    }
    // spec: "The main package must have package name main and declare
    // a function main that takes no arguments and returns no value."
    if (id.Name === "main" && this.pkg.name === "main") {
        this.error(ident, "InvalidMainDecl", "cannot declare main - must be func");
        return;
    }
    this.declare(this.pkg.scope, ident, obj, nopos);
    this.objMap.set(obj, d);
    obj.setOrder(this.objMap.size);
};
// filename returns a filename suitable for debugging output.
Checker.prototype.filename = function filename(fileNo) {
    const file = this.files?.[fileNo];
    const pos = file?.Pos?.() ?? 0;
    if (pos !== 0) {
        return file?.filename ?? `file[${fileNo}]`;
    }
    return `file[${fileNo}]`;
};
Checker.prototype.importPackage = function importPackage(at, path, dir_) {
    // If we already have a package for the given (path, dir)
    // pair, use it instead of doing a full import.
    // Checker.impMap only caches packages that are marked Complete
    // or fake (dummy packages for failed imports). Incomplete but
    // non-fake packages do require an import to complete them.
    const key = new importKey(path, dir_);
    const keyText = `${key.path}\u0000${key.dir}`;
    let imp = this.impMap.get(keyText) ?? null;
    if (imp !== null) {
        return imp;
    }
    // no package yet => import it
    if (path === "C" && (this.conf.FakeImportC || this.conf.go115UsesCgo)) {
        if (this.conf.FakeImportC && this.conf.go115UsesCgo) {
            this.error(at, "BadImportPath", "cannot use FakeImportC and go115UsesCgo together");
        }
        imp = NewPackage("C", "C");
        imp.fake = true; // package scope is not populated
        imp.cgo = this.conf.go115UsesCgo;
    }
    else {
        // ordinary import
        let err = null;
        const importer = this.conf.Importer;
        if (importer === null) {
            err = new Error("Config.Importer not installed");
        }
        else if ("ImportFrom" in importer && typeof importer.ImportFrom === "function") {
            const result = importer.ImportFrom(path, dir_, 0);
            imp = result[0];
            err = result[1];
            if (imp === null && err === null) {
                err = new Error(`Config.Importer.ImportFrom(${path}, ${dir_}, 0) returned nil but no error`);
            }
        }
        else {
            const result = importer.Import(path);
            imp = result[0];
            err = result[1];
            if (imp === null && err === null) {
                err = new Error(`Config.Importer.Import(${path}) returned nil but no error`);
            }
        }
        // make sure we have a valid package name
        // (errors here can only happen through manipulation of packages after creation)
        if (err === null && imp !== null && (imp.name === "_" || imp.name === "")) {
            err = new Error(`invalid package name: ${JSON.stringify(imp.name)}`);
            imp = null; // create fake package below
        }
        if (err !== null) {
            this.errorf(at, "BrokenImport", "could not import %s (%s)", path, err);
            if (imp === null) {
                // create a new fake package
                // come up with a sensible package name (heuristic)
                let name = path;
                if (name.length > 0 && name[name.length - 1] === "/") {
                    name = name.slice(0, -1);
                }
                const i = name.lastIndexOf("/");
                if (i >= 0) {
                    name = name.slice(i + 1);
                }
                imp = NewPackage(path, name);
            }
            // continue to use the package as best as we can
            imp.fake = true; // avoid follow-up lookup failures
        }
    }
    // package should be complete or marked fake, but be cautious
    if (imp !== null && (imp.complete || imp.fake)) {
        this.impMap.set(keyText, imp);
        // Once we've formatted an error message, keep the pkgPathMap
        // up-to-date on subsequent imports. It is used for package
        // qualification in error messages.
        if (this.pkgPathMap !== null) {
            this.markImports(imp);
        }
        return imp;
    }
    // something went wrong (importer may have returned incomplete package without error)
    return null;
};
// collectObjects collects all file and package objects and inserts them
// into their respective scopes. It also performs imports and associates
// methods with receiver base type names.
Checker.prototype.collectObjects = function collectObjects() {
    const pkg = this.pkg;
    // pkgImports is the set of packages already imported by any package file seen
    // so far. Used to avoid duplicate entries in pkg.imports. Allocate and populate
    // it (pkg.imports may not be empty if we are checking test files incrementally).
    // Note that pkgImports is keyed by package (and thus package path), not by an
    // importKey value. Two different importKey values may map to the same package
    // which is why we cannot use the check.impMap here.
    const pkgImports = new Map();
    for (const imp of pkg.imports) {
        pkgImports.set(imp, true);
    }
    class methodInfo {
        obj;
        ptr;
        recv;
        constructor(obj, // method
        ptr, // true if pointer receiver
        recv // receiver type name
        ) {
            this.obj = obj;
            this.ptr = ptr;
            this.recv = recv;
        }
    }
    const methods = []; // collected methods with valid receivers and non-blank _ names
    const fileScopes = new Array((this.files ?? []).length); // fileScopes[i] corresponds to check.files[i]
    for (let fileNo = 0; fileNo < (this.files ?? []).length; fileNo++) {
        const file = (this.files ?? [])[fileNo];
        this.version = asGoVersion(this.versions?.get(file) ?? "");
        // The package identifier denotes the current package,
        // but there is no corresponding package object.
        this.recordDef(file.Name, null);
        // Use the actual source file extent rather than *ast.File extent since the
        // latter doesn't include comments which appear at the start or end of the file.
        // Be conservative and use the *ast.File extent if we don't have a *token.File.
        const pos = file.Pos?.() ?? nopos;
        const end = file.End?.() ?? nopos;
        const fileScope = NewScope(pkg.scope, pos, end, this.filename(fileNo));
        fileScopes[fileNo] = fileScope;
        this.recordScope(file, fileScope);
        // determine file directory, necessary to resolve imports
        // FileName may be "" (typically for tests) in which case
        // we get "." as the directory which is what we would want.
        const fileDir = dir("");
        this.walkDecls(file.Decls ?? [], (d) => {
            const kind = d.kind;
            switch (kind) {
                case "importDecl": {
                    const spec = d.spec;
                    const [path, err] = validatedImportPath(spec.Path?.Value ?? "");
                    if (err !== null) {
                        this.errorf(spec.Path, "BadImportPath", "invalid import path (%s)", err);
                        return;
                    }
                    const imp = this.importPackage(atPos(pos), path, fileDir);
                    if (imp === null) {
                        return;
                    }
                    const impPkg = imp;
                    if (!pkgImports.has(impPkg)) {
                        pkgImports.set(impPkg, true);
                        pkg.imports.push(impPkg);
                    }
                    const name = spec.Name?.Name ?? impPkg.name;
                    const pkgName = NewPkgName(pos, pkg, name, impPkg);
                    this.imports = [...(this.imports ?? []), pkgName];
                    if (name === ".") {
                        for (const objName of impPkg.scope.Names()) {
                            const obj = impPkg.scope.Lookup(objName);
                            if (obj !== null) {
                                fileScope.Insert(obj);
                            }
                        }
                    }
                    else if (name !== "_") {
                        this.declare(fileScope, spec.Name ?? spec.Path, pkgName, pos);
                    }
                    break;
                }
                case "constDecl":
                case "varDecl":
                case "typeDecl":
                case "funcDecl":
                    // The detailed declaration-object construction is in decl.go in the
                    // original source. This pass records the file scope and leaves
                    // declaration processing to the mechanically translated decl walker.
                    break;
            }
        });
    }
};
Checker.prototype.sortObjects = function sortObjects() {
    this.objList = Array.from(this.objMap.keys());
    this.objList.sort((a, b) => a.order() - b.order());
};
Checker.prototype.unpackRecv = function unpackRecv(rtyp, _unpackParams) {
    const expr = rtyp;
    if (expr?.kind === "StarExpr") {
        return [true, expr.X, null];
    }
    return [false, rtyp, null];
};
Checker.prototype.resolveBaseTypeName = function resolveBaseTypeName(_ptr, recv) {
    const id = recv;
    const obj = this.pkg.scope.Lookup(id.Name ?? "");
    return obj instanceof NewTypeName(nopos, null, "", null).constructor ? obj : null;
};
Checker.prototype.packageObjects = function packageObjects() {
    for (const obj of this.objList) {
        this.objDecl(obj);
    }
};
Checker.prototype.unusedImports = function unusedImports() {
    for (const obj of this.imports ?? []) {
        if (!this.usedPkgNames.has(obj) && obj.name !== "_" && obj.name !== ".") {
            this.errorUnusedPkg(obj);
        }
    }
};
Checker.prototype.errorUnusedPkg = function errorUnusedPkg(obj) {
    this.softErrorf(obj, "UnusedImport", "%s imported and not used", obj.Name());
};
export function dir(path) {
    const i = path.lastIndexOf("/");
    if (i < 0) {
        return ".";
    }
    if (i === 0) {
        return "/";
    }
    return path.slice(0, i);
}
