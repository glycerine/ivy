// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import { Checker, atPos, debug, nopos, tracePos, environment, type positioner , registerCheckerMethod } from "./check.js";
import type { Object } from "./object.js";
import { Builtin, Const, FieldVar, Func, Label, LocalVar, NewConst, NewTypeName, NewVar, TypeName, Var, newVar, packagePrefix } from "./object.js";
import { Typ } from "./universe.js";
import { Invalid } from "./basic.js";
import { Alias, asNamed, unalias } from "./alias.js";
import { Named } from "./named.js";
import { Interface } from "./interface.js";
import { TypeParam, TypeParamList } from "./typeparam.js";
import type { Type } from "./type.js";
import { Signature } from "./signature.js";
import { NewTuple } from "./tuple.js";
import { bindTParams } from "./typelists.js";
import { objset } from "./objset.js";
import { Struct } from "./struct.js";
import { assert } from "./util.js";
import { cmpPos } from "./check.js";
import { declInfo } from "./resolver.js";
import type { Scope } from "./scope.js";
import { go1_18, go1_23, go1_9 } from "./version.js";
import { operand } from "./operand.js";
import { fieldListNumFields, funcDeclBody, funcDeclName, funcDeclRecv, funcDeclType, genDeclSpecs, genDeclTok, identName, nodeEnd, nodeKind, nodePos, specName, specNames, specType, specValues } from "./astcompat.js";
import { TypeString } from "./typestring.js";

declare module "./check.js" {
  interface Checker {
    declare(scope: Scope, id: unknown, obj: Object, pos: number): void;
    objDecl(obj: Object): void;
    validCycle(obj: Object): boolean;
    cycleError(cycle: Object[], start: number): void;
    walkDecls(decls: unknown[], f: (d: decl) => void): void;
    walkDecl(d: unknown, f: (d: decl) => void): void;
    constDecl(obj: Const, typ: unknown, init: unknown, inherited: boolean): void;
    varDecl(obj: Var, lhs: Var[] | null, typ: unknown, init: unknown): void;
    isImportedConstraint(typ: Type): boolean;
    typeDecl(obj: TypeName, tdecl: unknown): void;
    collectTypeParams(dst: { value: TypeParamList | null }, list: unknown): void;
    bound(x: unknown): Type;
    declareTypeParam(name: unknown, scopePos: number): TypeParam;
    collectMethods(obj: TypeName): void;
    checkFieldUniqueness(base: Named): void;
    funcDecl(obj: Func, decl: declInfo): void;
    declStmt(d: unknown): void;
  }
}

registerCheckerMethod("declare", function declare(scope: Scope, id: unknown, obj: Object, pos: number): void {
  // spec: "The blank identifier, represented by the underscore
  // character _, may be used in a declaration like any other
  // identifier but the declaration does not introduce a new
  // binding."
  if (obj.Name() !== "_") {
    const alt = scope.Insert(obj);
    if (alt !== null) {
      const err = (this as unknown as { newError: (code: unknown) => { addf: (...args: unknown[]) => void; addAltDecl: (obj: Object) => void; report: () => void } }).newError("DuplicateDecl");
      err.addf(obj, "%s redeclared in this block", obj.Name());
      err.addAltDecl(alt);
      err.report();
      return;
    }
    obj.setScopePos(pos);
  }
  if (id !== null && id !== undefined) {
    (this as unknown as { recordDef: (id: unknown, obj: Object) => void }).recordDef(id, obj);
  }
});

// pathString returns a string of the form a->b-> ... ->g for a path [a, b, ... g].
export function pathString(path: Object[]): string {
  let s = "";
  for (let i = 0; i < path.length; i++) {
    const p = path[i]!;
    if (i > 0) {
      s += "->";
    }
    s += p.Name();
  }
  return s;
}

// objDecl type-checks the declaration of obj in its respective (file) environment.
registerCheckerMethod("objDecl", function objDecl(obj: Object): void {
  if (tracePos) {
    this.pushPos(new atPos(obj.Pos()));
  }
  try {
    if (this.conf._Trace && obj.Type() === null) {
      if (this.indent === 0) {
        console.log(); // empty line between top-level objects for readability
      }
      (this as unknown as { trace: (pos: number, format: string, ...args: unknown[]) => void }).trace(obj.Pos(), "-- checking %s (objPath = %s)", obj, pathString(this.objPath.filter((x): x is Object => x !== null)));
      this.indent++;
    }

    // Checking the declaration of an object means determining its type
    // (and also its value for constants). An object (and thus its type)
    // may be in 1 of 3 states:
    //
    // - not in Checker.objPathIdx and type == nil : type is not yet known (white)
    // -     in Checker.objPathIdx                 : type is pending       (grey)
    // - not in Checker.objPathIdx and type != nil : type is known         (black)
    if (this.objPathIdx?.has(obj)) {
      if (obj instanceof Const || obj instanceof Var) {
        if (!this.validCycle(obj) || obj.Type() === null) {
          obj.setType(Typ[Invalid]!);
        }
      } else if (obj instanceof TypeName) {
        if (!this.validCycle(obj)) {
          obj.setType(Typ[Invalid]!);
        }
      } else if (obj instanceof Func) {
        if (!this.validCycle(obj)) {
          // Don't set type to Typ[Invalid]; plenty of code asserts that
          // functions have a *Signature type. Instead, leave the type
          // as an empty signature, which makes it impossible to
          // initialize a variable with the function.
        }
      } else {
        throw new Error("unreachable");
      }

      assert(obj.Type() !== null);
      return;
    }

    if (obj.Type() !== null) { // black, meaning it's already type-checked
      return;
    }

    // white, meaning it must be type-checked
    this.push(obj); // mark as grey
    try {
      const d = this.objMap.get(obj);
      assert(d !== undefined);

      // save/restore current environment and set up object environment
      const saved = Object.assign(new environment(), this);
      this.scope = d!.file;
      this.version = d!.version;

      try {
        if (obj instanceof Const) {
          this.decl = d!; // new package-level const decl
          this.constDecl(obj, d!.vtyp, d!.init, d!.inherited);
        } else if (obj instanceof Var) {
          this.decl = d!; // new package-level var decl
          this.varDecl(obj, d!.lhs, d!.vtyp, d!.init);
        } else if (obj instanceof TypeName) {
          // invalid recursive types are detected via path
          this.typeDecl(obj, d!.tdecl);
          this.collectMethods(obj); // methods can only be added to top-level types
        } else if (obj instanceof Func) {
          // functions may be recursive - no need to track dependencies
          this.funcDecl(obj, d!);
        } else {
          throw new Error("unreachable");
        }
      } finally {
        this.decl = saved.decl;
        this.scope = saved.scope;
        this.version = saved.version;
        this.iota = saved.iota;
        this.errpos = saved.errpos;
        this.inTParamList = saved.inTParamList;
        this.sig = saved.sig;
        this.isPanic = saved.isPanic;
        this.hasLabel = saved.hasLabel;
        this.hasCallOrRecv = saved.hasCallOrRecv;
        this.exprPos = saved.exprPos;
      }
    } finally {
      this.pop();
    }
  } finally {
    if (tracePos) {
      this.popPos();
    }
  }
});

// validCycle checks if the cycle starting with obj is valid and
// reports an error if it is not.
registerCheckerMethod("validCycle", function validCycle(obj: Object): boolean {
  const start = this.objPathIdx?.get(obj);
  assert(start !== undefined);
  const cycle = this.objPath.slice(start).filter((x): x is Object => x !== null);
  let tparCycle = false; // if set, the cycle is through a type parameter list
  let nval = 0; // number of (constant or variable) values in the cycle
  let ndef = 0; // number of type definitions in the cycle

  for (const obj of cycle) {
    if (obj instanceof Const || obj instanceof Var) {
      nval++;
    } else if (obj instanceof TypeName) {
      // If we reach a generic type that is part of a cycle
      // and we are in a type parameter list, we have a cycle
      // through a type parameter list.
      if (this.inTParamList && isGeneric(obj.typ)) {
        tparCycle = true;
        break;
      }

      if (!obj.IsAlias()) {
        ndef++;
      }
    } else if (obj instanceof Func) {
      // ignored for now
    } else {
      throw new Error("unreachable");
    }
  }

  if (this.conf._Trace) {
    (this as unknown as { trace: (pos: number, format: string, ...args: unknown[]) => void }).trace(obj.Pos(), "## cycle detected: objPath = %s->%s (len = %d)", pathString(cycle), obj.Name(), cycle.length);
  }

  // Cycles through type parameter lists are ok (go.dev/issue/68162).
  if (tparCycle) {
    return true;
  }

  // A cycle involving only constants and variables is invalid but we
  // ignore them here because they are reported via the initialization
  // cycle check.
  if (nval === cycle.length) {
    return true;
  }

  // A cycle involving only types (and possibly functions) must have at least
  // one type definition to be permitted: If there is no type definition, we
  // have a sequence of alias type names which will expand ad infinitum.
  if (nval === 0 && ndef > 0) {
    return true;
  }

  this.cycleError(cycle, firstInSrc(cycle));
  return false;
});

// cycleError reports a declaration cycle starting with the object at cycle[start].
registerCheckerMethod("cycleError", function cycleError(cycle: Object[], start: number): void {
  // name returns the (possibly qualified) object name.
  // This is needed because with generic types, cycles
  // may refer to imported types. See go.dev/issue/50788.
  const name = (obj: Object): string => {
    // include any type arguments in the reported error message
    const n = asNamed(obj.Type());
    if (n !== null && n.inst !== null) {
      return TypeString(n, null);
    }
    return packagePrefix(obj.Pkg(), null) + obj.Name();
  };

  // If obj is a type alias, mark it as valid (not broken) in order to avoid follow-on errors.
  let obj = cycle[start]!;
  const tname = obj instanceof TypeName ? obj : null;
  if (tname !== null) {
    const a = tname.Type();
    if (a instanceof Alias) {
      a.fromRHS = Typ[Invalid]!;
    }
  }

  // report a more concise error for self references
  if (cycle.length === 1) {
    if (tname !== null) {
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(obj, "InvalidDeclCycle", "invalid recursive type: %s refers to itself", name(obj));
    } else {
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(obj, "InvalidDeclCycle", "invalid cycle in declaration: %s refers to itself", name(obj));
    }
    return;
  }

  const err = (this as unknown as { newError: (code: unknown) => { addf: (...args: unknown[]) => void; report: () => void } }).newError("InvalidDeclCycle");
  if (tname !== null) {
    err.addf(obj, "invalid recursive type %s", name(obj));
  } else {
    err.addf(obj, "invalid cycle in declaration of %s", name(obj));
  }
  // "cycle[i] refers to cycle[j]" for (i,j) = (s,s+1), (s+1,s+2), ..., (n-1,0), (0,1), ..., (s-1,s) for len(cycle) = n, s = start.
  for (let i = 0; i < cycle.length; i++) {
    const next = cycle[(start + i + 1) % cycle.length]!;
    err.addf(obj, "%s refers to %s", name(obj), name(next));
    obj = next;
  }
  err.report();
});

// firstInSrc reports the index of the object with the "smallest"
// source position in path. path must not be empty.
export function firstInSrc(path: Object[]): number {
  let fst = 0;
  let pos = path[0]!.Pos();
  for (let i = 0; i < path.slice(1).length; i++) {
    const t = path.slice(1)[i]!;
    if (cmpPos(t.Pos(), pos) < 0) {
      fst = i + 1;
      pos = t.Pos();
    }
  }
  return fst;
}

export interface decl { node(): unknown; }

export class importDecl implements decl {
  public readonly kind = "importDecl";
  public constructor(public spec: unknown) {}
  public node(): unknown { return this.spec; }
}

export class constDecl implements decl {
  public readonly kind = "constDecl";
  public constructor(
    public spec: unknown,
    public iota: number,
    public typ: unknown,
    public init: unknown[],
    public inherited: boolean
  ) {}
  public node(): unknown { return this.spec; }
}

export class varDecl implements decl {
  public readonly kind = "varDecl";
  public constructor(public spec: unknown) {}
  public node(): unknown { return this.spec; }
}

export class typeDecl implements decl {
  public readonly kind = "typeDecl";
  public constructor(public spec: unknown) {}
  public node(): unknown { return this.spec; }
}

export class funcDecl implements decl {
  public readonly kind = "funcDecl";
  public constructor(public decl: unknown) {}
  public node(): unknown { return this.decl; }
}

registerCheckerMethod("walkDecls", function walkDecls(decls: unknown[], f: (d: decl) => void): void {
  for (const d of decls) {
    this.walkDecl(d, f);
  }
});

registerCheckerMethod("walkDecl", function walkDecl(d: unknown, f: (d: decl) => void): void {
  switch (nodeKind(d)) {
    case "BadDecl":
      // ignore
      break;
    case "GenDecl": {
      let last: unknown | null = null; // last ValueSpec with type or init exprs seen
      const specs = genDeclSpecs(d);
      for (let iota = 0; iota < specs.length; iota++) {
        const s = specs[iota];
        switch (nodeKind(s)) {
          case "ImportSpec":
            f(new importDecl(s));
            break;
          case "ValueSpec":
            switch (genDeclTok(d)) {
              case "CONST": {
                // determine which initialization expressions to use
                let inherited = true;
                switch (true) {
                  case specType(s) !== null && specType(s) !== undefined || specValues(s).length > 0:
                    last = s;
                    inherited = false;
                    break;
                  case last === null:
                    last = {}; // make sure last exists
                    inherited = false;
                    break;
                }
                this.arityMatch(s, last);
                f(new constDecl(s, iota, specType(last), specValues(last), inherited));
                break;
              }
              case "VAR":
                this.arityMatch(s, null);
                f(new varDecl(s));
                break;
              default:
                (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(s, "InvalidSyntaxTree", "invalid token %s", genDeclTok(d));
            }
            break;
          case "TypeSpec":
            f(new typeDecl(s));
            break;
          default:
            (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(s, "InvalidSyntaxTree", "unknown ast.Spec node %T", s);
        }
      }
      break;
    }
    case "FuncDecl":
      f(new funcDecl(d));
      break;
    default:
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(d, "InvalidSyntaxTree", "unknown ast.Decl node %T", d);
  }
});

registerCheckerMethod("constDecl", function constDeclMethod(obj: Const, typ: unknown, init: unknown, inherited: boolean): void {
  assert(obj.typ === null);

  // use the correct value of iota
  const savedIota = this.iota;
  const savedErrpos = this.errpos;
  this.iota = obj.val;
  this.errpos = null;

  try {
    // provide valid constant value under all circumstances
    obj.val = { unknown: true };

    // determine type, if any
    if (typ !== null && typ !== undefined) {
      const t = (this as unknown as { typ: (x: unknown) => Type }).typ(typ);
      if (!isConstType(t)) {
        if (isValid(t.Underlying())) {
          (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(typ, "InvalidConstType", "invalid constant type %s", t);
        }
        obj.typ = Typ[Invalid]!;
        return;
      }
      obj.typ = t;
    }

    // check initialization
    const x = new operand();
    if (init !== null && init !== undefined) {
      if (inherited) {
        this.errpos = new atPos(obj.pos);
      }
      (this as unknown as { expr: (target: unknown, x: operand, init: unknown) => void }).expr(null, x, init);
    }
    (this as unknown as { initConst: (obj: Const, x: operand) => void }).initConst(obj, x);
  } finally {
    this.iota = savedIota;
    this.errpos = savedErrpos;
  }
});

registerCheckerMethod("varDecl", function varDeclMethod(obj: Var, lhs: Var[] | null, typ: unknown, init: unknown): void {
  assert(obj.typ === null);

  // determine type, if any
  if (typ !== null && typ !== undefined) {
    obj.typ = (this as unknown as { varType: (x: unknown) => Type }).varType(typ);
  }

  // check initialization
  if (init === null || init === undefined) {
    if (typ === null || typ === undefined) {
      obj.typ = Typ[Invalid]!;
    }
    return;
  }

  if (lhs === null || lhs.length === 1) {
    assert(lhs === null || lhs[0] === obj);
    const x = new operand();
    (this as unknown as { expr: (target: unknown, x: operand, init: unknown) => void }).expr(newTarget(obj.typ, obj.name), x, init);
    (this as unknown as { initVar: (obj: Var, x: operand, context: string) => void }).initVar(obj, x, "variable declaration");
    return;
  }

  if (debug) {
    if (!lhs.includes(obj)) {
      throw new Error("inconsistent lhs");
    }
  }

  if (typ !== null && typ !== undefined) {
    for (const lhsVar of lhs) {
      lhsVar.typ = obj.typ;
    }
  }

  (this as unknown as { initVars: (lhs: Var[], init: unknown[], context: unknown) => void }).initVars(lhs, [init], null);
});

// isImportedConstraint reports whether typ is an imported type constraint.
registerCheckerMethod("isImportedConstraint", function isImportedConstraint(typ: Type): boolean {
  const named = asNamed(typ);
  if (named === null || named.obj.pkg === this.pkg || named.obj.pkg === null) {
    return false;
  }
  const u = named.Underlying();
  return u instanceof Interface && !u.IsMethodSet();
});

registerCheckerMethod("typeDecl", function typeDeclMethod(obj: TypeName, tdecl: unknown): void {
  assert(obj.typ === null);

  let versionErr = false;
  let rhs: Type = Typ[Invalid]!;
  this.later(() => {
    const t = asNamed(obj.typ);
    if (t !== null) {
      (this as unknown as { validType: (t: Named) => void }).validType(t);
    }
    const td = tdecl as { Type?: unknown };
    void (!versionErr && this.isImportedConstraint(rhs) && this.verifyVersionf(td.Type as positioner, go1_18, "using type constraint %s", rhs));
  }).describef(obj, "validType(%s)", obj.Name());

  const td = tdecl as { TypeParams?: { NumFields?: () => number; List?: unknown[] }; Assign?: { IsValid?: () => boolean }; Type?: unknown };
  const tparam0 = (td.TypeParams?.NumFields?.() ?? 0) > 0 ? td.TypeParams?.List?.[0] : null;

  // alias declaration
  if (td.Assign?.IsValid?.()) {
    if (!versionErr && tparam0 !== null && tparam0 !== undefined && !this.verifyVersionf(tparam0 as positioner, go1_23, "generic type alias")) {
      versionErr = true;
    }
    if (!versionErr && !this.verifyVersionf(new atPos(nopos), go1_9, "type alias")) {
      versionErr = true;
    }

    const alias = this.newAlias(obj, null);

    try {
      if (tparam0 !== null && tparam0 !== undefined) {
        (this as unknown as { openScope: (node: unknown, comment: string) => void; closeScope: () => void }).openScope(tdecl, "type parameters");
        try {
          const dst = { value: alias.tparams };
          this.collectTypeParams(dst, td.TypeParams);
          alias.tparams = dst.value;
        } finally {
          (this as unknown as { closeScope: () => void }).closeScope();
        }
      }

      rhs = (this as unknown as { declaredType: (x: unknown, def: TypeName) => Type }).declaredType(td.Type, obj);
      assert(rhs !== null);
      alias.fromRHS = rhs;

      if (rhs instanceof TypeParam && alias.tparams !== null && alias.tparams.list().includes(rhs)) {
        (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error(td.Type, "MisplacedTypeParam", "cannot use type parameter declared in alias declaration as RHS");
        alias.fromRHS = Typ[Invalid]!;
      }
    } finally {
      if (alias.fromRHS === null) {
        alias.fromRHS = Typ[Invalid]!;
        unalias(alias);
      }
    }

    return;
  }

  // type definition or generic type declaration
  if (!versionErr && tparam0 !== null && tparam0 !== undefined && !this.verifyVersionf(tparam0 as positioner, go1_18, "type parameter")) {
    versionErr = true;
  }

  const named = this.newNamed(obj, null, null);
  if (td.TypeParams !== null && td.TypeParams !== undefined) {
    (this as unknown as { openScope: (node: unknown, comment: string) => void; closeScope: () => void }).openScope(tdecl, "type parameters");
    try {
      const dst = { value: named.tparams };
      this.collectTypeParams(dst, td.TypeParams);
      named.tparams = dst.value;
    } finally {
      (this as unknown as { closeScope: () => void }).closeScope();
    }
  }

  rhs = (this as unknown as { declaredType: (x: unknown, def: TypeName) => Type }).declaredType(td.Type, obj);
  assert(rhs !== null);
  named.fromRHS = rhs;

  if (isTypeParam(rhs)) {
    (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error(td.Type, "MisplacedTypeParam", "cannot use a type parameter as RHS in type declaration");
    named.fromRHS = Typ[Invalid]!;
  }
});

registerCheckerMethod("collectTypeParams", function collectTypeParams(dst: { value: TypeParamList | null }, list: unknown): void {
  const tparams: TypeParam[] = [];
  const fields = (list as { Pos?: () => number; List?: unknown[] }).List ?? [];
  const scopePos = (list as { Pos?: () => number }).Pos?.() ?? nopos;
  for (const f of fields) {
    for (const name of (f as { Names?: unknown[] }).Names ?? []) {
      tparams.push(this.declareTypeParam(name, scopePos));
    }
  }

  dst.value = bindTParams(tparams);

  assert(!this.inTParamList);
  this.inTParamList = true;
  try {
    let index = 0;
    for (const f of fields) {
      let bound: Type;
      if ((f as { Type?: unknown }).Type !== null && (f as { Type?: unknown }).Type !== undefined) {
        bound = this.bound((f as { Type?: unknown }).Type);
        if (isTypeParam(bound)) {
          (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error((f as { Type?: unknown }).Type, "MisplacedTypeParam", "cannot use a type parameter as constraint");
          bound = Typ[Invalid]!;
        }
      } else {
        bound = Typ[Invalid]!;
      }
      const names = (f as { Names?: unknown[] }).Names ?? [];
      for (let i = 0; i < names.length; i++) {
        tparams[index + i]!.bound = bound;
      }
      index += names.length;
    }
  } finally {
    this.inTParamList = false;
  }
});

registerCheckerMethod("bound", function bound(x: unknown): Type {
  let wrap = false;
  const expr = x as { kind?: string; Op?: string };
  switch (expr.kind) {
    case "UnaryExpr":
      wrap = expr.Op === "TILDE";
      break;
    case "BinaryExpr":
      wrap = expr.Op === "OR";
      break;
  }
  if (wrap) {
    const t = (this as unknown as { typ: (x: unknown) => Type }).typ({ kind: "InterfaceType", Methods: { List: [{ Type: x }] } });
    if (t instanceof Interface) {
      t.implicit = true;
    }
    return t;
  }
  return (this as unknown as { typ: (x: unknown) => Type }).typ(x);
});

registerCheckerMethod("declareTypeParam", function declareTypeParam(name: unknown, scopePos: number): TypeParam {
  const ident = name as { Pos?: () => number; Name?: string };
  const tname = NewTypeName(ident.Pos?.() ?? nopos, this.pkg, ident.Name ?? "", null);
  const tpar = this.newTypeParam(tname, Typ[Invalid]!); // assigns type to tname as a side-effect
  this.declare(this.scope!, name, tname, scopePos);
  return tpar;
});

registerCheckerMethod("collectMethods", function collectMethods(obj: TypeName): void {
  const methods = this.methods?.get(obj);
  if (methods === undefined) {
    return;
  }
  this.methods?.delete(obj);

  const mset = new objset();
  const base = asNamed(obj.typ);
  if (base !== null) {
    this.later(() => {
      this.checkFieldUniqueness(base);
    }).describef(obj, "verifying field uniqueness for %v", base);

    for (let i = 0; i < base.NumMethods(); i++) {
      const m = base.Method(i);
      assert(m.name !== "_");
      assert(mset.insert(m) === null);
    }
  }

  for (const m of methods) {
    assert(m.name !== "_");
    const alt = mset.insert(m);
    if (alt !== null) {
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(m, "DuplicateMethod", "method %s.%s already declared", obj.Name(), m.name);
      continue;
    }

    if (base !== null) {
      base.AddMethod(m);
    }
  }
});

registerCheckerMethod("checkFieldUniqueness", function checkFieldUniqueness(base: Named): void {
  const t = base.Underlying();
  if (t instanceof Struct) {
    const mset = new objset();
    for (let i = 0; i < base.NumMethods(); i++) {
      const m = base.Method(i);
      assert(m.name !== "_");
      assert(mset.insert(m) === null);
    }

    for (const fld of t.fields ?? []) {
      if (fld.name !== "_") {
        const alt = mset.insert(fld);
        if (alt !== null) {
          const err = (this as unknown as { newError: (code: unknown) => { addf: (...args: unknown[]) => void; addAltDecl: (obj: Object) => void; report: () => void } }).newError("DuplicateFieldAndMethod");
          err.addf(alt, "field and method with the same name %s", fld.name);
          err.addAltDecl(fld);
          err.report();
        }
      }
    }
  }
});

registerCheckerMethod("funcDecl", function funcDeclMethod(obj: Func, decl: declInfo): void {
  assert(obj.typ === null);
  assert(this.iota === null);

  const sig = new Signature();
  obj.typ = sig; // guard against cycles

  const fdecl = decl.fdecl;
  const recv = funcDeclRecv(fdecl);
  const ftyp = funcDeclType(fdecl);
  const body = funcDeclBody(fdecl);
  (this as unknown as { funcType: (sig: Signature, recv: unknown, typ: unknown) => void }).funcType(sig, recv, ftyp);

  if (sig.scope !== null) {
    sig.scope.pos = nodePos(fdecl) || nopos;
    sig.scope.end = nodeEnd(fdecl) || nopos;
  }

  const typeParams = (ftyp as { TypeParams?: unknown; typeParams?: unknown } | null)?.TypeParams ?? (ftyp as { TypeParams?: unknown; typeParams?: unknown } | null)?.typeParams;
  if (fieldListNumFields(typeParams) > 0 && (body === null || body === undefined)) {
    (this as unknown as { softErrorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).softErrorf(funcDeclName(fdecl), "BadDecl", "generic function is missing function body");
  }

  if (!this.conf.IgnoreFuncBodies && body !== null && body !== undefined) {
    this.later(() => {
      (this as unknown as { funcBody: (decl: declInfo, name: string, sig: Signature, body: unknown, iota: unknown) => void }).funcBody(decl, obj.name, sig, body, null);
    }).describef(obj, "func %s", obj.name);
  }
});

registerCheckerMethod("declStmt", function declStmt(d: unknown): void {
  const pkg = this.pkg;

  this.walkDecl(d, (d: decl) => {
    if (d instanceof constDecl) {
      const top = this.delayed.length;

      const names = specNames(d.spec);
      const lhs: Const[] = new Array(names.length);
      for (let i = 0; i < names.length; i++) {
        const name = names[i];
        const obj = NewConst(nodePos(name) || nopos, pkg, identName(name), null, d.iota);
        lhs[i] = obj;

        let init: unknown = null;
        if (i < d.init.length) {
          init = d.init[i];
        }

        this.constDecl(obj, d.typ, init, d.inherited);
      }

      this.processDelayed(top);

      const scopePos = nodeEnd(d.spec) || nopos;
      for (let i = 0; i < names.length; i++) {
        this.declare(this.scope!, names[i], lhs[i]!, scopePos);
      }
    } else if (d instanceof varDecl) {
      const top = this.delayed.length;
      const names = specNames(d.spec);
      const values = specValues(d.spec);
      const lhs0: Var[] = new Array(names.length);
      for (let i = 0; i < names.length; i++) {
        const name = names[i];
        lhs0[i] = newVar(LocalVar, nodePos(name) || nopos, pkg, identName(name), null);
      }

      for (let i = 0; i < lhs0.length; i++) {
        const obj = lhs0[i]!;
        let lhs: Var[] | null = null;
        let init: unknown = null;
        switch (values.length) {
          case names.length:
            init = values[i];
            break;
          case 1:
            lhs = lhs0;
            init = values[0];
            break;
          default:
            if (i < values.length) {
              init = values[i];
            }
        }
        this.varDecl(obj, lhs, specType(d.spec), init);
        if (values.length === 1) {
          break;
        }
      }

      this.processDelayed(top);

      const scopePos = nodeEnd(d.spec) || nopos;
      for (let i = 0; i < names.length; i++) {
        this.declare(this.scope!, names[i], lhs0[i]!, scopePos);
      }
    } else if (d instanceof typeDecl) {
      const name = specName(d.spec);
      const obj = NewTypeName(nodePos(name) || nopos, pkg, identName(name), null);
      const scopePos = nodePos(name) || nopos;
      this.declare(this.scope!, name, obj, scopePos);
      this.push(obj); // mark as grey
      this.typeDecl(obj, d.spec);
      this.pop();
    } else {
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(d.node(), "InvalidSyntaxTree", "unknown ast.Decl node %T", d.node());
    }
  });
});

function isGeneric(t: Type | null): boolean {
  return t instanceof Named && t.TypeParams() !== null && t.TypeParams()!.Len() > 0;
}

function isConstType(t: Type): boolean {
  return t !== null;
}

function isValid(t: Type): boolean {
  return t !== Typ[Invalid];
}

function isTypeParam(t: Type | null): t is TypeParam {
  return t instanceof TypeParam;
}

function newTarget(typ: Type | null, name: string): { typ: Type | null; name: string } {
  return { typ, name };
}
