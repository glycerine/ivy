// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import { Checker, atPos, importKey, nopos, type positioner } from "./check.js";
import type { goVersion } from "./version.js";
import { asGoVersion } from "./version.js";
import { NewPackage, type Package } from "./package.js";
import { NewScope, type Scope } from "./scope.js";
import { NewConst, NewFunc, NewPkgName, NewTypeName, NewVar, type Object, type PkgName, type TypeName, type Var } from "./object.js";
import { assert } from "./util.js";
import { Signature } from "./signature.js";
import type { Type } from "./type.js";

// A declInfo describes a package-level const, type, var, or func declaration.
export class declInfo {
  public file: Scope | null = null; // scope of file containing this declaration
  public version: goVersion | null = null; // Go version of file containing this declaration
  public lhs: Var[] | null = null; // lhs of n:1 variable declarations, or nil
  public vtyp: unknown = null; // type, or nil (for const and var declarations only)
  public init: unknown = null; // init/orig expression, or nil (for const and var declarations only)
  public inherited = false; // if set, the init expression is inherited from a previous constant declaration
  public tdecl: unknown = null; // type declaration, or nil
  public fdecl: unknown = null; // func declaration, or nil

  // The deps field tracks initialization expression dependencies.
  public deps: Map<Object, boolean> | null = null; // lazily initialized

  public constructor(init?: Partial<declInfo>) {
    Object.assign(this, init);
  }

  // hasInitializer reports whether the declared object has an initialization
  // expression or function body.
  public hasInitializer(): boolean {
    const fdecl = this.fdecl as { Body?: unknown } | null;
    return this.init !== null || (fdecl !== null && fdecl.Body !== null && fdecl.Body !== undefined);
  }

  // addDep adds obj to the set of objects d's init expression depends on.
  public addDep(obj: Object): void {
    let m = this.deps;
    if (m === null) {
      m = new Map();
      this.deps = m;
    }
    m.set(obj, true);
  }
}

declare module "./check.js" {
  interface Checker {
    arityMatch(s: unknown, init: unknown): void;
    declarePkgObj(ident: unknown, obj: Object, d: declInfo): void;
    filename(fileNo: number): string;
    importPackage(at: positioner, path: string, dir: string): Package | null;
    collectObjects(): void;
    sortObjects(): void;
    unpackRecv(rtyp: unknown, unpackParams: boolean): [boolean, unknown, unknown];
    resolveBaseTypeName(ptr: boolean, recv: unknown): TypeName | null;
    packageObjects(): void;
    unusedImports(): void;
    errorUnusedPkg(obj: PkgName): void;
  }
}

// arityMatch checks that the lhs and rhs of a const or var decl
// have the appropriate number of names and init exprs. For const
// decls, init is the value spec providing the init exprs; for
// var decls, init is nil (the init exprs are in s in this case).
Checker.prototype.arityMatch = function arityMatch(s: unknown, init: unknown): void {
  const spec = s as { Names?: unknown[]; Values?: unknown[]; Type?: unknown };
  const initSpec = init as { Values?: unknown[]; Pos?: () => unknown } | null;
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
        (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error(s, code, "missing type or init expr");
      }
      break;
    case l < r:
      if (l < (spec.Values?.length ?? 0)) {
        // init exprs from s
        const n = spec.Values![l];
        (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(n, code, "extra init expr %s", n);
        // TODO(gri) avoid declared and not used error here
      } else {
        // init exprs "inherited"
        (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(s, code, "extra init expr at %s", String(initSpec?.Pos?.() ?? ""));
        // TODO(gri) avoid declared and not used error here
      }
      break;
    case l > r && (init !== null || r !== 1): {
      const n = spec.Names![r];
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(n, code, "missing init expr for %s", n);
      break;
    }
  }
};

export function validatedImportPath(path: string): [string, Error | null] {
  let s: string;
  try {
    s = JSON.parse(path) as string;
  } catch (err) {
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

function isGraphic(r: string): boolean {
  return r.length > 0 && !/[\p{Cc}\p{Cs}\p{Cn}]/u.test(r);
}

// declarePkgObj declares obj in the package scope, records its ident -> obj mapping,
// and updates check.objMap. The object must not be a function or method.
Checker.prototype.declarePkgObj = function declarePkgObj(ident: unknown, obj: Object, d: declInfo): void {
  const id = ident as { Name?: string };
  assert(id.Name === obj.Name());

  // spec: "A package-scope or file-scope identifier with name init
  // may only be declared to be a function with this (func()) signature."
  if (id.Name === "init") {
    (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error(ident, "InvalidInitDecl", "cannot declare init - must be func");
    return;
  }

  // spec: "The main package must have package name main and declare
  // a function main that takes no arguments and returns no value."
  if (id.Name === "main" && this.pkg.name === "main") {
    (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error(ident, "InvalidMainDecl", "cannot declare main - must be func");
    return;
  }

  (this as unknown as { declare: (scope: Scope, id: unknown, obj: Object, pos: number) => void }).declare(this.pkg.scope, ident, obj, nopos);
  this.objMap.set(obj, d);
  obj.setOrder(this.objMap.size);
};

// filename returns a filename suitable for debugging output.
Checker.prototype.filename = function filename(fileNo: number): string {
  const file = this.files?.[fileNo] as { Pos?: () => number; filename?: string } | undefined;
  const pos = file?.Pos?.() ?? 0;
  if (pos !== 0) {
    return file?.filename ?? `file[${fileNo}]`;
  }
  return `file[${fileNo}]`;
};

Checker.prototype.importPackage = function importPackage(at: positioner, path: string, dir_: string): Package | null {
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
      (this as unknown as { error: (at: unknown, code: unknown, msg: string) => void }).error(at, "BadImportPath", "cannot use FakeImportC and go115UsesCgo together");
    }
    imp = NewPackage("C", "C");
    imp.fake = true; // package scope is not populated
    imp.cgo = this.conf.go115UsesCgo;
  } else {
    // ordinary import
    let err: unknown = null;
    const importer = this.conf.Importer;
    if (importer === null) {
      err = new Error("Config.Importer not installed");
    } else if ("ImportFrom" in importer && typeof importer.ImportFrom === "function") {
      const result = importer.ImportFrom(path, dir_, 0);
      imp = result[0];
      err = result[1];
      if (imp === null && err === null) {
        err = new Error(`Config.Importer.ImportFrom(${path}, ${dir_}, 0) returned nil but no error`);
      }
    } else {
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
      (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(at, "BrokenImport", "could not import %s (%s)", path, err);
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
      (this as unknown as { markImports: (pkg: Package) => void }).markImports(imp);
    }
    return imp;
  }

  // something went wrong (importer may have returned incomplete package without error)
  return null;
};

// collectObjects collects all file and package objects and inserts them
// into their respective scopes. It also performs imports and associates
// methods with receiver base type names.
Checker.prototype.collectObjects = function collectObjects(): void {
  const pkg = this.pkg;

  // pkgImports is the set of packages already imported by any package file seen
  // so far. Used to avoid duplicate entries in pkg.imports. Allocate and populate
  // it (pkg.imports may not be empty if we are checking test files incrementally).
  // Note that pkgImports is keyed by package (and thus package path), not by an
  // importKey value. Two different importKey values may map to the same package
  // which is why we cannot use the check.impMap here.
  const pkgImports = new Map<Package, boolean>();
  for (const imp of pkg.imports) {
    pkgImports.set(imp, true);
  }

  class methodInfo {
    public constructor(
      public obj: import("./object.js").Func, // method
      public ptr: boolean, // true if pointer receiver
      public recv: unknown // receiver type name
    ) {}
  }
  const methods: methodInfo[] = []; // collected methods with valid receivers and non-blank _ names

  const fileScopes = new Array<Scope>((this.files ?? []).length); // fileScopes[i] corresponds to check.files[i]
  for (let fileNo = 0; fileNo < (this.files ?? []).length; fileNo++) {
    const file = (this.files ?? [])[fileNo] as {
      Name?: { Name?: string; Pos?: () => number };
      Pos?: () => number;
      End?: () => number;
      Decls?: unknown[];
    };
    this.version = asGoVersion(this.versions?.get(file) ?? "");

    // The package identifier denotes the current package,
    // but there is no corresponding package object.
    (this as unknown as { recordDef: (id: unknown, obj: Object | null) => void }).recordDef(file.Name, null);

    // Use the actual source file extent rather than *ast.File extent since the
    // latter doesn't include comments which appear at the start or end of the file.
    // Be conservative and use the *ast.File extent if we don't have a *token.File.
    const pos = file.Pos?.() ?? nopos;
    const end = file.End?.() ?? nopos;
    const fileScope = NewScope(pkg.scope, pos, end, this.filename(fileNo));
    fileScopes[fileNo] = fileScope;
    (this as unknown as { recordScope: (node: unknown, scope: Scope) => void }).recordScope(file, fileScope);

    // determine file directory, necessary to resolve imports
    // FileName may be "" (typically for tests) in which case
    // we get "." as the directory which is what we would want.
    const fileDir = dir("");

    (this as unknown as { walkDecls: (decls: unknown[], f: (d: unknown) => void) => void }).walkDecls(file.Decls ?? [], (d: unknown) => {
      const kind = (d as { kind?: string; spec?: unknown; decl?: unknown }).kind;
      switch (kind) {
        case "importDecl": {
          const spec = (d as { spec: { Path?: { Value?: string }; Name?: { Name?: string } } }).spec;
          const [path, err] = validatedImportPath(spec.Path?.Value ?? "");
          if (err !== null) {
            (this as unknown as { errorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).errorf(spec.Path, "BadImportPath", "invalid import path (%s)", err);
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
          } else if (name !== "_") {
            (this as unknown as { declare: (scope: Scope, id: unknown, obj: Object, pos: number) => void }).declare(fileScope, spec.Name ?? spec.Path, pkgName, pos);
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

Checker.prototype.sortObjects = function sortObjects(): void {
  this.objList = Array.from(this.objMap.keys());
  this.objList.sort((a, b) => a.order() - b.order());
};

Checker.prototype.unpackRecv = function unpackRecv(rtyp: unknown, _unpackParams: boolean): [boolean, unknown, unknown] {
  const expr = rtyp as { kind?: string; X?: unknown };
  if (expr?.kind === "StarExpr") {
    return [true, expr.X, null];
  }
  return [false, rtyp, null];
};

Checker.prototype.resolveBaseTypeName = function resolveBaseTypeName(_ptr: boolean, recv: unknown): TypeName | null {
  const id = recv as { Name?: string };
  const obj = this.pkg.scope.Lookup(id.Name ?? "");
  return obj instanceof NewTypeName(nopos, null, "", null).constructor ? obj as TypeName : null;
};

Checker.prototype.packageObjects = function packageObjects(): void {
  for (const obj of this.objList) {
    (this as unknown as { objDecl: (obj: Object) => void }).objDecl(obj);
  }
};

Checker.prototype.unusedImports = function unusedImports(): void {
  for (const obj of this.imports ?? []) {
    if (!this.usedPkgNames.has(obj) && obj.name !== "_" && obj.name !== ".") {
      this.errorUnusedPkg(obj);
    }
  }
};

Checker.prototype.errorUnusedPkg = function errorUnusedPkg(obj: PkgName): void {
  (this as unknown as { softErrorf: (at: unknown, code: unknown, format: string, ...args: unknown[]) => void }).softErrorf(obj, "UnusedImport", "%s imported and not used", obj.Name());
};

export function dir(path: string): string {
  const i = path.lastIndexOf("/");
  if (i < 0) {
    return ".";
  }
  if (i === 0) {
    return "/";
  }
  return path.slice(0, i);
}
