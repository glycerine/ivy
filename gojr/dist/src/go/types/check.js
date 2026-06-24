// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// This file implements the Check function, which drives type-checking.
import { NoPos } from "./token.js";
import { asGoVersion, go_current, go1_21 } from "./version.js";
import { assert } from "./util.js";
import { NoPos as zeroPos } from "./token.js";
import { fileGoVersion, fileName, identName, nodePos } from "./astcompat.js";
function checkerMethodQueue() {
    const g = globalThis;
    if (g.__gojrPendingCheckerMethods === undefined) {
        g.__gojrPendingCheckerMethods = [];
    }
    return g.__gojrPendingCheckerMethods;
}
export function registerCheckerMethod(name, fn) {
    const g = globalThis;
    if (g.__gojrCheckerCtor !== undefined) {
        g.__gojrCheckerCtor.prototype[name] = fn;
        return;
    }
    checkerMethodQueue().push([name, fn]);
}
function installPendingCheckerMethods(ctor) {
    const g = globalThis;
    g.__gojrCheckerCtor = ctor;
    for (const [name, fn] of checkerMethodQueue()) {
        ctor.prototype[name] = fn;
    }
    checkerMethodQueue().length = 0;
}
// atPos wraps a token.Pos to implement the positioner interface.
export class atPos {
    s;
    constructor(s) {
        this.s = s;
    }
    Pos() {
        return this.s;
    }
}
function posIsValid(pos) {
    return pos !== NoPos;
}
export function cmpPos(p, q) { return p - q; }
// nopos, noposn indicate an unknown position
export const nopos = zeroPos;
export const noposn = new atPos(nopos);
// debugging/development support
export const debug = false; // leave on during development
// position tracing for panics during type checking
export const tracePos = true;
// exprInfo stores information about an untyped expression.
export class exprInfo {
    isLhs;
    mode;
    typ;
    val;
    constructor(isLhs, // expression is lhs operand of a shift with delayed type-check
    mode, typ, val // constant value; or nil (if not a constant)
    ) {
        this.isLhs = isLhs;
        this.mode = mode;
        this.typ = typ;
        this.val = val;
    }
}
// An environment represents the environment within which an object is
// type-checked.
export class environment {
    decl = null; // package-level declaration whose init expression/function body is checked
    scope = null; // top-most scope for lookups
    version = null; // current accepted language version; changes across files
    iota = null; // value of iota in a constant declaration; nil otherwise
    errpos = null; // if set, identifier position of a constant with inherited initializer
    inTParamList = false; // set if inside a type parameter list
    sig = null; // function signature if inside a function; nil otherwise
    isPanic = null; // set of panic call expressions (used for termination check)
    hasLabel = false; // set if a function makes use of labels (only ~1% of functions); unused outside functions
    hasCallOrRecv = false; // set if an expression contains a function call or channel receive operation
    // go/types only
    exprPos = nopos; // if valid, identifiers are looked up as if at position pos (used by CheckExpr, Eval)
    // lookupScope looks up name in the current environment and if an object is
    // found it returns the scope containing the object and the object.
    // Otherwise it returns (nil, nil).
    //
    // Note that obj.Parent() may be different from the returned scope if the
    // object was inserted into the scope and already had a parent at that
    // time (see Scope.Insert). This can only happen for dot-imported objects
    // whose parent is the scope of the package that exported them.
    lookupScope(name) {
        for (let s = this.scope; s !== null; s = s.parent) {
            const obj = s.Lookup(name);
            if (obj !== null && (!posIsValid(this.exprPos) || cmpPos(obj.scopePos(), this.exprPos) <= 0)) {
                return [s, obj];
            }
        }
        return [null, null];
    }
    // lookup is like lookupScope but it only returns the object (or nil).
    lookup(name) {
        const [, obj] = this.lookupScope(name);
        return obj;
    }
}
// An importKey identifies an imported package by import path and source directory
// (directory containing the file containing the import). In practice, the directory
// may always be the same, or may not matter. Given an (import path, directory), an
// importer must always return the same package (but given two different import paths,
// an importer may still return the same package by mapping them to the same package
// paths).
export class importKey {
    path;
    dir;
    constructor(path, dir) {
        this.path = path;
        this.dir = dir;
    }
}
// A dotImportKey describes a dot-imported object in the given scope.
export class dotImportKey {
    scope;
    name;
    constructor(scope, name) {
        this.scope = scope;
        this.name = name;
    }
}
// An action describes a (delayed) action.
export class action {
    version;
    f;
    desc;
    constructor(version, // applicable language version
    f, // action to be executed
    desc = null // action description; may be nil, requires debug to be set
    ) {
        this.version = version;
        this.f = f;
        this.desc = desc;
    }
    // If debug is set, describef sets a printf-formatted description for action a.
    // Otherwise, it is a no-op.
    describef(pos, format, ...args) {
        if (debug) {
            this.desc = new actionDesc(pos, format, args);
        }
    }
}
// An actionDesc provides information on an action.
// For debugging only.
export class actionDesc {
    pos;
    format;
    args;
    constructor(pos, format, args) {
        this.pos = pos;
        this.format = format;
        this.args = args;
    }
}
// A bailout panic is used for early termination.
export class bailout {
}
// A Checker maintains the state of the type checker.
// It must be created with [NewChecker].
export class Checker extends environment {
    // package information
    // (initialized by NewChecker, valid for the life-time of checker)
    conf;
    ctxt;
    fset;
    pkg;
    Info;
    nextID = 0; // unique Id for type parameters (first valid Id is 1)
    objMap = new Map(); // maps package-level objects and (non-interface) methods to declaration info
    objList = []; // source-ordered keys of objMap
    impMap = new Map(); // maps (import path, source directory) to (complete or fake) package
    // pkgPathMap maps package names to the set of distinct import paths we've
    // seen for that name, anywhere in the import graph. It is used for
    // disambiguating package names in error messages.
    pkgPathMap = null;
    seenPkgMap = null;
    // information collected during type-checking of a set of package files
    // (initialized by Files, valid only for the duration of check.Files;
    // maps and lists are allocated on demand)
    files = null; // package files
    versions = null; // maps files to goVersion strings
    imports = null; // list of imported packages
    dotImportMap = null; // maps dot-imported objects to the package they were dot-imported through
    brokenAliases = null; // set of aliases with broken (not yet determined) types
    unionTypeSets = null; // computed type sets for union types
    usedVars = new Map(); // set of used variables
    usedPkgNames = new Map(); // set of used package names
    mono = null; // graph for detecting non-monomorphizable instantiation loops
    firstErr = null; // first error encountered
    methods = null; // maps package scope type names to associated non-blank methods
    untyped = null; // map of expressions without final type
    delayed = []; // stack of delayed action segments; segments are processed in FIFO order
    objPath = []; // path of object dependencies during type-checking (for cycle reporting)
    objPathIdx = null; // map of object to object path index during type-checking
    cleaners = []; // list of types that may need a final cleanup at the end of type-checking
    // debugging
    posStack = []; // stack of source positions seen; used for panic tracing
    indent = 0; // indentation for tracing
    constructor(conf, fset, pkg, info) {
        super();
        this.conf = conf;
        this.ctxt = conf.Context;
        this.fset = fset;
        this.pkg = pkg;
        this.Info = info;
        this.usedVars = new Map();
        this.usedPkgNames = new Map();
    }
    // addDeclDep adds the dependency edge (check.decl -> to) if check.decl exists
    addDeclDep(to) {
        const from = this.decl;
        if (from === null) {
            return; // not in a package-level init expression
        }
        if (!this.objMap.has(to)) {
            return; // to is not a package-level object
        }
        from.addDep(to);
    }
    rememberUntyped(e, lhs, mode, typ, val) {
        let m = this.untyped;
        if (m === null) {
            m = new Map();
            this.untyped = m;
        }
        m.set(e, new exprInfo(lhs, mode, typ, val));
    }
    // later pushes f on to the stack of actions that will be processed later;
    // either at the end of the current statement, or in case of a local constant
    // or variable declaration, before the constant or variable is in scope
    // (so that f still sees the scope before any new declarations).
    // later returns the pushed action so one can provide a description
    // via action.describef for debugging, if desired.
    later(f) {
        const i = this.delayed.length;
        this.delayed.push(new action(this.version, f));
        return this.delayed[i];
    }
    // push pushes obj onto the object path and records its index in the path index map.
    push(obj) {
        if (this.objPathIdx === null) {
            this.objPathIdx = new Map();
        }
        this.objPathIdx.set(obj, this.objPath.length);
        this.objPath.push(obj);
    }
    // pop pops an object from the object path and removes it from the path index map.
    pop() {
        const i = this.objPath.length - 1;
        const obj = this.objPath[i];
        this.objPath[i] = null; // help the garbage collector
        this.objPath = this.objPath.slice(0, i);
        if (obj !== null && obj !== undefined) {
            this.objPathIdx?.delete(obj);
        }
    }
    // needsCleanup records objects/types that implement the cleanup method
    // which will be called at the end of type-checking.
    needsCleanup(c) {
        this.cleaners.push(c);
    }
    // initFiles initializes the files-specific portion of checker.
    // The provided files must all belong to the same package.
    initFiles(files) {
        // start with a clean slate (check.Files may be called multiple times)
        // TODO(gri): what determines which fields are zeroed out here, vs at the end
        // of checkFiles?
        this.files = null;
        this.imports = null;
        this.dotImportMap = null;
        this.firstErr = null;
        this.methods = null;
        this.untyped = null;
        this.delayed = [];
        this.objPath = [];
        this.objPathIdx = null;
        this.cleaners = [];
        // We must initialize usedVars and usedPkgNames both here and in NewChecker,
        // because initFiles is not called in the CheckExpr or Eval codepaths, yet we
        // want to free this memory at the end of Files ('used' predicates are
        // only needed in the context of a given file).
        this.usedVars = new Map();
        this.usedPkgNames = new Map();
        // determine package name and collect valid files
        const pkg = this.pkg;
        for (const file of files) {
            const nameIdent = fileName(file);
            const name = identName(nameIdent);
            switch (pkg.name) {
                case "":
                    if (name !== "_") {
                        pkg.name = name;
                    }
                    else {
                        this.error(nameIdent, "BlankPkgName", "invalid package name _");
                    }
                    this.files = [...(this.files ?? []), file];
                    break;
                case name:
                    this.files = [...(this.files ?? []), file];
                    break;
                default:
                    this.errorf(new atPos(nodePos(nameIdent) || nopos), "MismatchedPkgName", "package %s; expected package %s", name, pkg.name);
                // ignore this file
            }
        }
        // reuse Info.FileVersions if provided
        let versions = this.Info.FileVersions;
        if (versions === null) {
            versions = new Map();
        }
        this.versions = versions;
        const pkgVersion = asGoVersion(this.conf.GoVersion);
        if (pkgVersion.isValid() && files.length > 0 && pkgVersion.cmp(go_current) > 0) {
            this.errorf(files[0], "TooNew", "package requires newer Go version %v (application built with %v)", pkgVersion, go_current);
        }
        // determine Go version for each file
        for (const file of this.files ?? []) {
            const nameIdent = fileName(file);
            // use unaltered Config.GoVersion by default
            // (This version string may contain dot-release numbers as in go1.20.1,
            // unlike file versions which are Go language versions only, if valid.)
            let v = this.conf.GoVersion;
            // If the file specifies a version, use max(fileVersion, go1.21).
            const fileVersion = asGoVersion(fileGoVersion(file));
            if (fileVersion.isValid()) {
                // Go 1.21 introduced the feature of setting the go.mod
                // go line to an early version of Go and allowing //go:build lines
                // to set the Go version in a given file. Versions Go 1.21 and later
                // can be set backwards compatibly as that was the first version
                // files with go1.21 or later build tags could be built with.
                //
                // Set the version to max(fileVersion, go1.21): That will allow a
                // downgrade to a version before go1.22, where the for loop semantics
                // change was made, while being backwards compatible with versions of
                // go before the new //go:build semantics were introduced.
                v = versionMax(fileVersion, go1_21).String();
                // Report a specific error for each tagged file that's too new.
                // (Normally the build system will have filtered files by version,
                // but clients can present arbitrary files to the type checker.)
                if (fileVersion.cmp(go_current) > 0) {
                    // Use position of 'package [p]' for types/types2 consistency.
                    // (Ideally we would use the //build tag itself.)
                    this.errorf(nameIdent, "TooNew", "file requires newer Go version %v (application built with %v)", fileVersion, go_current);
                }
            }
            versions.set(file, v);
        }
    }
    // pushPos pushes pos onto the pos stack.
    pushPos(pos) {
        this.posStack.push(pos);
    }
    // popPos pops from the pos stack.
    popPos() {
        this.posStack = this.posStack.slice(0, this.posStack.length - 1);
    }
    handleBailout(err) {
        err.value = this.firstErr;
    }
    // Files checks the provided files as part of the checker's package.
    Files(files) {
        if (this.pkg.path === "unsafe" && this.pkg.name === "unsafe") {
            // Defensive handling for Unsafe, which cannot be type checked, and must
            // not be mutated. See https://go.dev/issue/61212 for an example of where
            // Unsafe is passed to NewChecker.
            return null;
        }
        // Avoid early returns here! Nearly all errors can be
        // localized to a piece of syntax and needn't prevent
        // type-checking of the rest of the package.
        const err = { value: null };
        try {
            this.checkFiles(files);
        }
        catch (p) {
            if (p instanceof bailout) {
                this.handleBailout(err);
            }
            else {
                throw p;
            }
        }
        return err.value;
    }
    // checkFiles type-checks the specified files. Errors are reported as
    // a side effect, not by returning early, to ensure that well-formed
    // syntax is properly type annotated even in a package containing
    // errors.
    checkFiles(files) {
        const print = (msg) => {
            if (this.conf._Trace) {
                console.log();
                console.log(msg);
            }
        };
        print("== initFiles ==");
        this.initFiles(files);
        print("== collectObjects ==");
        this.collectObjects();
        print("== sortObjects ==");
        this.sortObjects();
        print("== directCycles ==");
        this.directCycles();
        print("== packageObjects ==");
        this.packageObjects();
        print("== processDelayed ==");
        this.processDelayed(0); // incl. all functions
        print("== cleanup ==");
        this.cleanup();
        print("== initOrder ==");
        this.initOrder();
        if (!this.conf.DisableUnusedImportCheck) {
            print("== unusedImports ==");
            this.unusedImports();
        }
        print("== recordUntyped ==");
        this.recordUntyped();
        if (this.firstErr === null) {
            // TODO(mdempsky): Ensure monomorph is safe when errors exist.
            this.monomorph();
        }
        this.pkg.goVersion = this.conf.GoVersion;
        this.pkg.complete = true;
        // no longer needed - release memory
        this.imports = null;
        this.dotImportMap = null;
        this.pkgPathMap = null;
        this.seenPkgMap = null;
        this.brokenAliases = null;
        this.unionTypeSets = null;
        this.usedVars = new Map();
        this.usedPkgNames = new Map();
        this.ctxt = null;
        // TODO(gri): shouldn't the cleanup above occur after the bailout?
        // TODO(gri) There's more memory we should release at this point.
    }
    // processDelayed processes all delayed actions pushed after top.
    processDelayed(top) {
        // If each delayed action pushes a new action, the
        // stack will continue to grow during this loop.
        // However, it is only processing functions (which
        // are processed in a delayed fashion) that may
        // add more actions (such as nested functions), so
        // this is a sufficiently bounded process.
        const savedVersion = this.version;
        for (let i = top; i < this.delayed.length; i++) {
            const a = this.delayed[i];
            if (this.conf._Trace) {
                if (a.desc !== null) {
                    this.trace(a.desc.pos.Pos(), "-- " + a.desc.format, ...a.desc.args);
                }
                else {
                    this.trace(nopos, "-- delayed %p", a.f);
                }
            }
            this.version = a.version; // reestablish the effective Go version captured earlier
            a.f(); // may append to check.delayed
            if (this.conf._Trace) {
                console.log();
            }
        }
        assert(top <= this.delayed.length); // stack must not have shrunk
        this.delayed = this.delayed.slice(0, top);
        this.version = savedVersion;
    }
    // cleanup runs cleanup for all collected cleaners.
    cleanup() {
        // Don't use a range clause since Named.cleanup may add more cleaners.
        for (let i = 0; i < this.cleaners.length; i++) {
            this.cleaners[i].cleanup();
        }
        this.cleaners = [];
    }
    // go/types doesn't support recording of types directly in the AST.
    // dummy function to match types2 code.
    recordTypeAndValueInSyntax(_x, _mode, _typ, _val) {
        // nothing to do
    }
    // go/types doesn't support recording of types directly in the AST.
    // dummy function to match types2 code.
    recordCommaOkTypesInSyntax(_x, _t0, _t1) {
        // nothing to do
    }
    // allowVersion reports whether the current effective Go version
    // (which may vary from one file to another) is allowed to use the
    // feature version (want).
    allowVersion(want) {
        return this.version === null || !this.version.isValid() || this.version.cmp(want) >= 0;
    }
    // verifyVersionf is like allowVersion but also accepts a format string and arguments
    // which are used to report a version error if allowVersion returns false.
    verifyVersionf(at, v, format, ...args) {
        if (!this.allowVersion(v)) {
            this.versionErrorf(at, v, format, ...args);
            return false;
        }
        return true;
    }
}
installPendingCheckerMethods(Checker);
// NewChecker returns a new [Checker] instance for a given package.
// [Package] files may be added incrementally via checker.Files.
export function NewChecker(conf, fset, pkg, info) {
    // make sure we have a configuration
    if (conf === null) {
        conf = newDefaultConfig();
    }
    // make sure we have an info struct
    if (info === null) {
        info = newDefaultInfo();
    }
    // Note: clients may call NewChecker with the Unsafe package, which is
    // globally shared and must not be mutated. Therefore NewChecker must not
    // mutate *pkg.
    //
    // (previously, pkg.goVersion was mutated here: go.dev/issue/61212)
    return new Checker(conf, fset, pkg, info);
}
function newDefaultConfig() {
    return {
        Context: null,
        GoVersion: "",
        IgnoreFuncBodies: false,
        FakeImportC: false,
        go115UsesCgo: false,
        _Trace: false,
        Error: null,
        Importer: null,
        Sizes: null,
        DisableUnusedImportCheck: false,
        _ErrorURL: "",
        Check(path, fset, files, info) {
            const pkg = thisPackage(path);
            return [pkg, NewChecker(this, fset, pkg, info).Files(files)];
        },
        alignof(T) { return this.Sizes.Alignof(T); },
        offsetsof(T) { return this.Sizes.Offsetsof(T.fields ?? []); },
        offsetof(_T, _index) { return -1; },
        sizeof(T) { return this.Sizes.Sizeof(T); }
    };
}
function newDefaultInfo() {
    return {
        Types: null,
        Instances: null,
        Defs: null,
        Uses: null,
        Implicits: null,
        Selections: null,
        Scopes: null,
        InitOrder: null,
        FileVersions: null,
        recordTypes() { return this.Types !== null; },
        TypeOf(e) {
            const t = this.Types?.get(e);
            if (t !== undefined) {
                return t.Type ?? null;
            }
            const obj = this.ObjectOf(e);
            if (obj !== null) {
                return obj.Type();
            }
            return null;
        },
        ObjectOf(id) {
            const obj = this.Defs?.get(id);
            if (obj !== undefined && obj !== null) {
                return obj;
            }
            return this.Uses?.get(id) ?? null;
        },
        PkgNameOf(_imp) { return null; }
    };
}
function thisPackage(path) {
    return {
        path,
        name: "",
        scope: null,
        imports: [],
        complete: false,
        fake: false,
        cgo: false,
        goVersion: "",
        Path() { return this.path; },
        Name() { return this.name; },
        SetName(name) { this.name = name; },
        GoVersion() { return this.goVersion; },
        Scope() { return this.scope; },
        Complete() { return this.complete; },
        MarkComplete() { this.complete = true; },
        Imports() { return this.imports; },
        SetImports(list) { this.imports = list; },
        String() { return `package ${this.name} (${JSON.stringify(this.path)})`; }
    };
}
export function versionMax(a, b) {
    if (a.cmp(b) < 0) {
        return b;
    }
    return a;
}
// instantiatedIdent determines the identifier of the type instantiated in expr.
// Helper function for recordInstance in recording.go.
export function instantiatedIdent(expr) {
    let selOrIdent = null;
    const e = expr;
    switch (e.kind) {
        case "IndexExpr":
            selOrIdent = e.X ?? e.object;
            break;
        case "IndexListExpr": // only exists in go/ast, not syntax
            selOrIdent = e.X ?? e.object;
            break;
        case "SelectorExpr":
        case "Ident":
            selOrIdent = e;
            break;
    }
    const x = selOrIdent;
    switch (x?.kind) {
        case "Ident":
            return x;
        case "SelectorExpr":
            return x.Sel ?? x.selector;
    }
    // extra debugging of go.dev/issue/63933
    throw new Error(`instantiated ident not found; please report: ${String(expr)}`);
}
