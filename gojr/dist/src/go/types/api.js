// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { NewPackage } from "./package.js";
import { init as initUniverse, Typ } from "./universe.js";
import { UntypedNil } from "./basic.js";
import { novalue, typexpr, builtin, constant_, variable, mapindex, value, commaok, commaerr } from "./operand.js";
import { NewChecker } from "./check.js";
import { WriteExpr } from "./exprstring.js";
function configMethodQueue() {
    const g = globalThis;
    if (g.__gojrPendingConfigMethods === undefined) {
        g.__gojrPendingConfigMethods = [];
    }
    return g.__gojrPendingConfigMethods;
}
function installPendingConfigMethods(ctor) {
    const g = globalThis;
    g.__gojrConfigCtor = ctor;
    for (const [name, fn] of configMethodQueue()) {
        ctor.prototype[name] = fn;
    }
    configMethodQueue().length = 0;
}
// An Error describes a type-checking error; it implements the error interface.
export class Error extends globalThis.Error {
    Fset;
    Pos;
    Msg;
    Soft;
    go116code;
    go116start;
    go116end;
    constructor(Fset, Pos, Msg, Soft = false, go116code = null, go116start = null, go116end = null) {
        super(Msg);
        this.Fset = Fset;
        this.Pos = Pos;
        this.Msg = Msg;
        this.Soft = Soft;
        this.go116code = go116code;
        this.go116start = go116start;
        this.go116end = go116end;
    }
    // Error returns an error string formatted as follows:
    // filename:line:column: message
    Error() {
        return `${this.Fset?.Position(this.Pos) ?? String(this.Pos)}: ${this.Msg}`;
    }
}
// An ArgumentError holds an error associated with an argument index.
export class ArgumentError extends globalThis.Error {
    Index;
    Err;
    constructor(Index, Err) {
        super(Err.message);
        this.Index = Index;
        this.Err = Err;
    }
    Error() { return this.Err.message; }
    Unwrap() { return this.Err; }
}
// A Config specifies the configuration for type checking.
// The zero value for Config is a ready-to-use default configuration.
export class Config {
    Context = null;
    GoVersion = "";
    IgnoreFuncBodies = false;
    FakeImportC = false;
    go115UsesCgo = false;
    _Trace = false;
    Error = null;
    Importer = null;
    Sizes = null;
    DisableUnusedImportCheck = false;
    _ErrorURL = "";
    // Check type-checks a package and returns the resulting package object and
    // the first error if any. Additionally, if info != nil, Check populates each
    // of the non-nil maps in the [Info] struct.
    Check(path, fset, files, info) {
        initUniverse();
        const pkg = NewPackage(path, "");
        return [pkg, NewChecker(this, fset, pkg, info).Files(files)];
    }
}
installPendingConfigMethods(Config);
// Linkname for use from srcimporter.
//go:linkname srcimporter_setUsesCgo
export function srcimporter_setUsesCgo(conf) {
    conf.go115UsesCgo = true;
}
// Info holds result type information for a type-checked package.
// Only the information for which a map is provided is collected.
// If the package has type errors, the collected information may
// be incomplete.
export class Info {
    Types = null;
    Instances = null;
    Defs = null;
    Uses = null;
    Implicits = null;
    Selections = null;
    Scopes = null;
    InitOrder = null;
    FileVersions = null;
    recordTypes() {
        return this.Types !== null;
    }
    // TypeOf returns the type of expression e, or nil if not found.
    // Precondition: the Types, Uses and Defs maps are populated.
    TypeOf(e) {
        const t = this.Types?.get(e);
        if (t !== undefined) {
            return t.Type;
        }
        const obj = this.ObjectOf(e);
        if (obj !== null) {
            return obj.Type();
        }
        return null;
    }
    // ObjectOf returns the object denoted by the specified id,
    // or nil if not found.
    ObjectOf(id) {
        const obj = this.Defs?.get(id);
        if (obj !== undefined && obj !== null) {
            return obj;
        }
        return this.Uses?.get(id) ?? null;
    }
    // PkgNameOf returns the local package name defined by the import,
    // or nil if not found.
    PkgNameOf(imp) {
        const name = imp?.Name;
        let obj;
        if (name !== undefined && name !== null) {
            obj = this.Defs?.get(name);
        }
        else {
            obj = this.Implicits?.get(imp);
        }
        return obj !== undefined && obj !== null && obj.constructor.name === "PkgName" ? obj : null;
    }
}
// TypeAndValue reports the type and value (for constants)
// of the corresponding expression.
export class TypeAndValue {
    mode;
    Type;
    Value;
    constructor(mode, Type, Value = null) {
        this.mode = mode;
        this.Type = Type;
        this.Value = Value;
    }
    // IsVoid reports whether the corresponding expression
    // is a function call without results.
    IsVoid() { return this.mode === novalue; }
    // IsType reports whether the corresponding expression specifies a type.
    IsType() { return this.mode === typexpr; }
    // IsBuiltin reports whether the corresponding expression denotes
    // a (possibly parenthesized) built-in function.
    IsBuiltin() { return this.mode === builtin; }
    // IsValue reports whether the corresponding expression is a value.
    // Builtins are not considered values. Constant values have a non-
    // nil Value.
    IsValue() {
        switch (this.mode) {
            case constant_:
            case variable:
            case mapindex:
            case value:
            case commaok:
            case commaerr:
                return true;
        }
        return false;
    }
    // IsNil reports whether the corresponding expression denotes the
    // predeclared value nil.
    IsNil() { return this.mode === value && this.Type === Typ[UntypedNil]; }
    // Addressable reports whether the corresponding expression
    // is addressable (https://golang.org/ref/spec#Address_operators).
    Addressable() { return this.mode === variable; }
    // Assignable reports whether the corresponding expression
    // is assignable to (provided a value of the right type).
    Assignable() { return this.mode === variable || this.mode === mapindex; }
    // HasOk reports whether the corresponding expression may be
    // used on the rhs of a comma-ok assignment.
    HasOk() { return this.mode === commaok || this.mode === mapindex; }
}
// Instance reports the type arguments and instantiated type for type and
// function instantiations.
export class Instance {
    TypeArgs;
    Type;
    constructor(TypeArgs, Type) {
        this.TypeArgs = TypeArgs;
        this.Type = Type;
    }
    String() {
        return `${this.TypeArgs?.String() ?? ""}${this.Type?.String() ?? "<nil>"}`;
    }
}
// An Initializer describes a package-level variable, or a list of variables in case
// of a multi-valued initialization expression, and the corresponding initialization
// expression.
export class Initializer {
    Lhs;
    Rhs;
    constructor(Lhs, Rhs) {
        this.Lhs = Lhs;
        this.Rhs = Rhs;
    }
    String() {
        const buf = [];
        for (let i = 0; i < this.Lhs.length; i++) {
            const lhs = this.Lhs[i];
            if (i > 0) {
                buf.push(", ");
            }
            buf.push(lhs.Name());
        }
        buf.push(" = ");
        WriteExpr(buf, this.Rhs);
        return buf.join("");
    }
}
