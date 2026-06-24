// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { atPos, importKey, nopos, registerCheckerMethod } from "./check.js";
import { asGoVersion, go1_18, go1_27 } from "./version.js";
import { NewPackage } from "./package.js";
import { NewScope } from "./scope.js";
import { NewConst, NewFunc, NewPkgName, NewTypeName, NewVar } from "./object.js";
import { assert } from "./util.js";
import { basicLitValue, fieldListNumFields, fileDecls, fileName, funcDeclBody, funcDeclName, funcDeclRecv, funcDeclType, identName, nodeEnd, nodePos, specName, specNames, specPath, specType, specValues } from "./astcompat.js";
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
registerCheckerMethod("arityMatch", function arityMatch(s, init) {
    const initSpec = init;
    const l = specNames(s).length;
    let r = specValues(s).length;
    if (initSpec !== null && initSpec !== undefined) {
        r = specValues(initSpec).length;
    }
    const code = "WrongAssignCount";
    switch (true) {
        case init === null && r === 0:
            // var decl w/o init expr
            if (specType(s) === null || specType(s) === undefined) {
                this.error(s, code, "missing type or init expr");
            }
            break;
        case l < r:
            if (l < specValues(s).length) {
                // init exprs from s
                const n = specValues(s)[l];
                this.errorf(n, code, "extra init expr %s", n);
                // TODO(gri) avoid declared and not used error here
            }
            else {
                // init exprs "inherited"
                this.errorf(s, code, "extra init expr at %s", String(nodePos(initSpec) ?? ""));
                // TODO(gri) avoid declared and not used error here
            }
            break;
        case l > r && (init !== null || r !== 1): {
            const n = specNames(s)[r];
            this.errorf(n, code, "missing init expr for %s", n);
            break;
        }
    }
});
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
registerCheckerMethod("declarePkgObj", function declarePkgObj(ident, obj, d) {
    const name = identName(ident);
    assert(name === obj.Name());
    // spec: "A package-scope or file-scope identifier with name init
    // may only be declared to be a function with this (func()) signature."
    if (name === "init") {
        this.error(ident, "InvalidInitDecl", "cannot declare init - must be func");
        return;
    }
    // spec: "The main package must have package name main and declare
    // a function main that takes no arguments and returns no value."
    if (name === "main" && this.pkg.name === "main") {
        this.error(ident, "InvalidMainDecl", "cannot declare main - must be func");
        return;
    }
    this.declare(this.pkg.scope, ident, obj, nopos);
    this.objMap.set(obj, d);
    obj.setOrder(this.objMap.size);
});
// filename returns a filename suitable for debugging output.
registerCheckerMethod("filename", function filename(fileNo) {
    const file = this.files?.[fileNo];
    const pos = file?.Pos?.() ?? 0;
    if (pos !== 0) {
        return file?.filename ?? `file[${fileNo}]`;
    }
    return `file[${fileNo}]`;
});
registerCheckerMethod("importPackage", function importPackage(at, path, dir_) {
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
});
// collectObjects collects all file and package objects and inserts them
// into their respective scopes. It also performs imports and associates
// methods with receiver base type names.
registerCheckerMethod("collectObjects", function collectObjects() {
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
        this.recordDef(fileName(file), null);
        // Use the actual source file extent rather than *ast.File extent since the
        // latter doesn't include comments which appear at the start or end of the file.
        // Be conservative and use the *ast.File extent if we don't have a *token.File.
        const pos = nodePos(file) || nopos;
        const end = nodeEnd(file) || nopos;
        const fileScope = NewScope(pkg.scope, pos, end, this.filename(fileNo));
        fileScopes[fileNo] = fileScope;
        this.recordScope(file, fileScope);
        // determine file directory, necessary to resolve imports
        // FileName may be "" (typically for tests) in which case
        // we get "." as the directory which is what we would want.
        const fileDir = dir("");
        this.walkDecls(fileDecls(file), (d) => {
            const kind = d.kind;
            switch (kind) {
                case "importDecl": {
                    const spec = d.spec;
                    const pathLit = specPath(spec);
                    const [path, err] = validatedImportPath(basicLitValue(pathLit));
                    if (err !== null) {
                        this.errorf(pathLit, "BadImportPath", "invalid import path (%s)", err);
                        return;
                    }
                    const imp = this.importPackage(new atPos(pos), path, fileDir);
                    if (imp === null) {
                        return;
                    }
                    const impPkg = imp;
                    if (!pkgImports.has(impPkg)) {
                        pkgImports.set(impPkg, true);
                        pkg.imports.push(impPkg);
                    }
                    const nameIdent = specName(spec);
                    const name = identName(nameIdent) || impPkg.name;
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
                        this.declare(fileScope, nameIdent ?? pathLit, pkgName, pos);
                    }
                    break;
                }
                case "constDecl": {
                    const cd = d;
                    const names = specNames(cd.spec);
                    for (let i = 0; i < names.length; i++) {
                        const name = names[i];
                        const obj = NewConst(nodePos(name) || pos, pkg, identName(name), null, cd.iota);
                        const init = i < cd.init.length ? cd.init[i] : null;
                        this.declarePkgObj(name, obj, new declInfo({ file: fileScope, version: this.version, vtyp: cd.typ, init, inherited: cd.inherited }));
                    }
                    break;
                }
                case "varDecl": {
                    const vd = d;
                    const names = specNames(vd.spec);
                    const values = specValues(vd.spec);
                    const lhs = names.map((name) => NewVar(nodePos(name) || pos, pkg, identName(name), null));
                    for (let i = 0; i < names.length; i++) {
                        const init = values.length === 1 ? values[0] : (i < values.length ? values[i] : null);
                        const sharedLhs = values.length === 1 ? lhs : null;
                        this.declarePkgObj(names[i], lhs[i], new declInfo({ file: fileScope, version: this.version, lhs: sharedLhs, vtyp: specType(vd.spec), init }));
                    }
                    break;
                }
                case "typeDecl": {
                    const td = d;
                    const name = specName(td.spec);
                    const obj = NewTypeName(nodePos(name) || pos, pkg, identName(name), null);
                    this.declarePkgObj(name, obj, new declInfo({ file: fileScope, version: this.version, tdecl: td.spec }));
                    break;
                }
                case "funcDecl": {
                    const fd = d.decl;
                    const name = funcDeclName(fd);
                    const nameText = identName(name);
                    const obj = NewFunc(nodePos(name) || pos, pkg, identName(name), null);
                    const info = new declInfo({ file: fileScope, version: this.version, fdecl: fd });
                    const recvList = funcDeclRecv(fd);
                    const ftyp = funcDeclType(fd);
                    const typeParams = ftyp?.TypeParams ?? ftyp?.typeParams;
                    const tparam0 = fieldListNumFields(typeParams) > 0 ? fieldListFields(typeParams)[0] : null;
                    const recvCount = fieldListNumFields(recvList);
                    if (recvList === null || recvList === undefined || recvCount === 0) {
                        // regular function
                        if (recvList !== null && recvList !== undefined) {
                            this.error(recvList, "BadRecv", "method has no receiver");
                        }
                        if (nameText === "init" || (nameText === "main" && pkg.name === "main")) {
                            // init and main functions must not declare type and ordinary parameters or results.
                            const code = nameText === "main" ? "InvalidMainDecl" : "InvalidInitDecl";
                            if (tparam0 !== null && tparam0 !== undefined) {
                                this.softErrorf(tparam0, code, "func %s must have no type parameters", nameText);
                            }
                            const params = ftyp?.Params ?? ftyp?.params;
                            const results = ftyp?.Results ?? ftyp?.results;
                            if (fieldListNumFields(params) !== 0 || fieldListNumFields(results) !== 0) {
                                this.softErrorf(name, code, "func %s must have no arguments and no return values", nameText);
                            }
                        }
                        else {
                            void (tparam0 !== null && tparam0 !== undefined && this.verifyVersionf(tparam0, go1_18, "type parameter"));
                        }
                        if (nameText === "init") {
                            // Don't declare init functions in the package scope: they are invisible.
                            obj.parent = pkg.scope;
                            this.recordDef(name, obj);
                            if (funcDeclBody(fd) === null || funcDeclBody(fd) === undefined) {
                                this.softErrorf(obj, "MissingInitBody", "func init must have a body");
                            }
                        }
                        else {
                            this.declare(pkg.scope, name, obj, nopos);
                        }
                    }
                    else {
                        const recvField = fieldListFields(recvList)[0];
                        const [ptr, base] = this.unpackRecv(fieldType(recvField), false);
                        if (isIdentNode(base) && identName(name) !== "_") {
                            methods.push(new methodInfo(obj, ptr, base));
                        }
                        void (tparam0 !== null && tparam0 !== undefined && this.verifyVersionf(tparam0, go1_27, "generic method"));
                        this.recordDef(name, obj);
                    }
                    this.objMap.set(obj, info);
                    obj.setOrder(this.objMap.size);
                    void funcDeclBody(fd);
                    break;
                }
            }
        });
    }
    if (methods.length > 0) {
        this.methods = new Map();
        for (const m of methods) {
            const base = this.resolveBaseTypeName(m.ptr, m.recv);
            if (base !== null) {
                m.obj.hasPtrRecv_ = m.ptr;
                const list = this.methods.get(base) ?? [];
                list.push(m.obj);
                this.methods.set(base, list);
            }
        }
    }
});
registerCheckerMethod("sortObjects", function sortObjects() {
    this.objList = Array.from(this.objMap.keys());
    this.objList.sort((a, b) => a.order() - b.order());
});
registerCheckerMethod("unpackRecv", function unpackRecv(rtyp, _unpackParams) {
    const expr = rtyp;
    if (expr?.kind === "StarExpr") {
        return [true, expr.X ?? expr.expr, null];
    }
    return [false, rtyp, null];
});
registerCheckerMethod("resolveBaseTypeName", function resolveBaseTypeName(_ptr, recv) {
    const id = recv;
    const obj = this.pkg.scope.Lookup(id.Name ?? id.Value ?? id.name ?? "");
    return obj instanceof NewTypeName(nopos, null, "", null).constructor ? obj : null;
});
registerCheckerMethod("packageObjects", function packageObjects() {
    for (const obj of this.objList) {
        this.objDecl(obj);
    }
});
registerCheckerMethod("unusedImports", function unusedImports() {
    for (const obj of this.imports ?? []) {
        if (!this.usedPkgNames.has(obj) && obj.name !== "_" && obj.name !== ".") {
            this.errorUnusedPkg(obj);
        }
    }
});
registerCheckerMethod("errorUnusedPkg", function errorUnusedPkg(obj) {
    this.softErrorf(obj, "UnusedImport", "%s imported and not used", obj.Name());
});
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
function fieldListFields(list) {
    const l = list;
    return l?.List ?? l?.fields ?? [];
}
function fieldType(field) {
    const f = field;
    return f?.Type ?? f?.type ?? null;
}
function isIdentNode(node) {
    const n = node;
    return n?.kind === "Ident" || n?.Name !== undefined || n?.Value !== undefined || n?.name !== undefined;
}
