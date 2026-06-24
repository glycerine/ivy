// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import type { Type } from "./type.js";
import type { Package } from "./package.js";
import { NewPackage } from "./package.js";
import type { Object, PkgName, Var } from "./object.js";
import type { Scope } from "./scope.js";
import type { Selection } from "./selection.js";
import type { TypeList } from "./typelists.js";
import { init as initUniverse, Typ } from "./universe.js";
import { UntypedNil } from "./basic.js";
import { operandMode, novalue, typexpr, builtin, constant_, variable, mapindex, value, commaok, commaerr } from "./operand.js";
import type { Sizes } from "./sizes.js";
import type { Context } from "./context.js";
import { NewChecker } from "./check.js";
import { WriteExpr } from "./exprstring.js";

type configCtor = { prototype: any };
type configMethodRecord = [string, Function];

function configMethodQueue(): configMethodRecord[] {
  const g = globalThis as typeof globalThis & { __gojrPendingConfigMethods?: configMethodRecord[] };
  if (g.__gojrPendingConfigMethods === undefined) {
    g.__gojrPendingConfigMethods = [];
  }
  return g.__gojrPendingConfigMethods;
}

function installPendingConfigMethods(ctor: configCtor): void {
  const g = globalThis as typeof globalThis & { __gojrConfigCtor?: configCtor };
  g.__gojrConfigCtor = ctor;
  for (const [name, fn] of configMethodQueue()) {
    ctor.prototype[name] = fn;
  }
  configMethodQueue().length = 0;
}

// An Error describes a type-checking error; it implements the error interface.
export class Error extends globalThis.Error {
  public constructor(
    public Fset: { Position(pos: unknown): string } | null,
    public Pos: unknown,
    public Msg: string,
    public Soft = false,
    public go116code: unknown = null,
    public go116start: unknown = null,
    public go116end: unknown = null
  ) {
    super(Msg);
  }

  // Error returns an error string formatted as follows:
  // filename:line:column: message
  public Error(): string {
    return `${this.Fset?.Position(this.Pos) ?? String(this.Pos)}: ${this.Msg}`;
  }
}

// An ArgumentError holds an error associated with an argument index.
export class ArgumentError extends globalThis.Error {
  public constructor(
    public Index: number,
    public Err: globalThis.Error
  ) {
    super(Err.message);
  }

  public Error(): string { return this.Err.message; }
  public Unwrap(): globalThis.Error { return this.Err; }
}

// An Importer resolves import paths to Packages.
export interface Importer {
  // Import returns the imported package for the given import path.
  Import(path: string): [Package | null, unknown];
}

// ImportMode is reserved for future use.
export type ImportMode = number;

// An ImporterFrom resolves import paths to packages; it
// supports vendoring per https://golang.org/s/go15vendor.
export interface ImporterFrom extends Importer {
  ImportFrom(path: string, dir: string, mode: ImportMode): [Package | null, unknown];
}

// A Config specifies the configuration for type checking.
// The zero value for Config is a ready-to-use default configuration.
export class Config {
  public Context: Context | null = null;
  public GoVersion = "";
  public IgnoreFuncBodies = false;
  public FakeImportC = false;
  public go115UsesCgo = false;
  public _Trace = false;
  public Error: ((err: unknown) => void) | null = null;
  public Importer: Importer | null = null;
  public Sizes: Sizes | null = null;
  public DisableUnusedImportCheck = false;
  public _ErrorURL = "";

  // Check type-checks a package and returns the resulting package object and
  // the first error if any. Additionally, if info != nil, Check populates each
  // of the non-nil maps in the [Info] struct.
  public Check(path: string, fset: unknown, files: unknown[], info: Info | null): [Package, unknown] {
    initUniverse();
    const pkg = NewPackage(path, "");
    return [pkg, NewChecker(this, fset, pkg, info).Files(files)];
  }
}

installPendingConfigMethods(Config);

// Linkname for use from srcimporter.
//go:linkname srcimporter_setUsesCgo
export function srcimporter_setUsesCgo(conf: Config): void {
  conf.go115UsesCgo = true;
}

// Info holds result type information for a type-checked package.
// Only the information for which a map is provided is collected.
// If the package has type errors, the collected information may
// be incomplete.
export class Info {
  public Types: Map<unknown, TypeAndValue> | null = null;
  public Instances: Map<unknown, Instance> | null = null;
  public Defs: Map<unknown, Object | null> | null = null;
  public Uses: Map<unknown, Object> | null = null;
  public Implicits: Map<unknown, Object> | null = null;
  public Selections: Map<unknown, Selection> | null = null;
  public Scopes: Map<unknown, Scope> | null = null;
  public InitOrder: Initializer[] | null = null;
  public FileVersions: Map<unknown, string> | null = null;

  public recordTypes(): boolean {
    return this.Types !== null;
  }

  // TypeOf returns the type of expression e, or nil if not found.
  // Precondition: the Types, Uses and Defs maps are populated.
  public TypeOf(e: unknown): Type | null {
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
  public ObjectOf(id: unknown): Object | null {
    const obj = this.Defs?.get(id);
    if (obj !== undefined && obj !== null) {
      return obj;
    }
    return this.Uses?.get(id) ?? null;
  }

  // PkgNameOf returns the local package name defined by the import,
  // or nil if not found.
  public PkgNameOf(imp: unknown): PkgName | null {
    const name = (imp as { Name?: unknown } | null)?.Name;
    let obj: Object | null | undefined;
    if (name !== undefined && name !== null) {
      obj = this.Defs?.get(name);
    } else {
      obj = this.Implicits?.get(imp);
    }
    return obj !== undefined && obj !== null && obj.constructor.name === "PkgName" ? obj as PkgName : null;
  }
}

// TypeAndValue reports the type and value (for constants)
// of the corresponding expression.
export class TypeAndValue {
  public constructor(
    public mode: operandMode,
    public Type: Type | null,
    public Value: unknown = null
  ) {}

  // IsVoid reports whether the corresponding expression
  // is a function call without results.
  public IsVoid(): boolean { return this.mode === novalue; }

  // IsType reports whether the corresponding expression specifies a type.
  public IsType(): boolean { return this.mode === typexpr; }

  // IsBuiltin reports whether the corresponding expression denotes
  // a (possibly parenthesized) built-in function.
  public IsBuiltin(): boolean { return this.mode === builtin; }

  // IsValue reports whether the corresponding expression is a value.
  // Builtins are not considered values. Constant values have a non-
  // nil Value.
  public IsValue(): boolean {
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
  public IsNil(): boolean { return this.mode === value && this.Type === Typ[UntypedNil]; }

  // Addressable reports whether the corresponding expression
  // is addressable (https://golang.org/ref/spec#Address_operators).
  public Addressable(): boolean { return this.mode === variable; }

  // Assignable reports whether the corresponding expression
  // is assignable to (provided a value of the right type).
  public Assignable(): boolean { return this.mode === variable || this.mode === mapindex; }

  // HasOk reports whether the corresponding expression may be
  // used on the rhs of a comma-ok assignment.
  public HasOk(): boolean { return this.mode === commaok || this.mode === mapindex; }
}

// Instance reports the type arguments and instantiated type for type and
// function instantiations.
export class Instance {
  public constructor(
    public TypeArgs: TypeList | null,
    public Type: Type | null
  ) {}

  public String(): string {
    return `${this.TypeArgs?.String() ?? ""}${this.Type?.String() ?? "<nil>"}`;
  }
}

// An Initializer describes a package-level variable, or a list of variables in case
// of a multi-valued initialization expression, and the corresponding initialization
// expression.
export class Initializer {
  public constructor(
    public Lhs: Var[],
    public Rhs: unknown
  ) {}

  public String(): string {
    const buf: string[] = [];
    for (let i = 0; i < this.Lhs.length; i++) {
      const lhs = this.Lhs[i]!;
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
