// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements (error and trace) message formatting support.

import { Checker , registerCheckerMethod } from "./check.js";
import type { Package } from "./package.js";
import type { Qualifier } from "./typestring.js";
import { TypeString } from "./typestring.js";
import type { Type } from "./type.js";
import { ObjectString, type Object } from "./object.js";
import { ExprString } from "./exprstring.js";
import { operand } from "./operand.js";

export function sprintf(_fset: unknown, qf: Qualifier | null, _tpSubscripts: boolean, format: string, ...args: unknown[]): string {
  for (let i = 0; i < args.length; i++) {
    let arg = args[i];
    if (arg === null || arg === undefined) {
      arg = "<nil>";
    } else if (arg instanceof operand) {
      arg = operandString(arg, qf);
    } else if (Array.isArray(arg) && arg.every((x) => x instanceof operand)) {
      arg = `[${arg.map((x) => operandString(x, qf)).join(", ")}]`;
    } else if (isObject(arg)) {
      arg = ObjectString(arg, qf);
    } else if (isType(arg)) {
      arg = TypeString(arg, qf);
    } else if (isExpr(arg)) {
      arg = ExprString(arg);
    }
    args[i] = arg;
  }
  return fmtSprintf(format, ...args);
}

function isObject(x: unknown): x is Object {
  return x !== null && typeof x === "object" && "Name" in x && "Type" in x && "Pkg" in x;
}

function isType(x: unknown): x is Type {
  return x !== null && typeof x === "object" && "Underlying" in x && "String" in x;
}

function isExpr(x: unknown): boolean {
  return x !== null && typeof x === "object" && ("kind" in x || "Kind" in x);
}

function operandString(x: operand, qf: Qualifier | null): string {
  if (x.typ() !== null) {
    return TypeString(x.typ()!, qf);
  }
  return String(x.val ?? x.mode());
}

function fmtSprintf(format: string, ...args: unknown[]): string {
  let i = 0;
  return format.replace(/%[#]?[vTsdqpu]/g, (verb) => {
    const arg = args[i++];
    switch (verb) {
      case "%q":
        return JSON.stringify(String(arg));
      case "%T":
        return arg === null || arg === undefined ? "<nil>" : (arg as object).constructor.name;
      case "%d":
        return String(Number(arg));
      default:
        return String(arg);
    }
  });
}

// check may be nil.
declare module "./check.js" {
  interface Checker {
    sprintf(format: string, ...args: unknown[]): string;
    trace(pos: number, format: string, ...args: unknown[]): void;
    dump(format: string, ...args: unknown[]): void;
    qualifier(pkg: Package): string;
    markImports(pkg: Package): void;
  }
}

registerCheckerMethod("sprintf", function sprintfMethod(format: string, ...args: unknown[]): string {
  let fset: unknown = null;
  let qf: Qualifier | null = null;
  if (this !== null) {
    fset = this.fset;
    qf = (pkg: unknown) => this.qualifier(pkg as Package);
  }
  return sprintf(fset, qf, false, format, ...args);
});

registerCheckerMethod("trace", function trace(_pos: number, format: string, ...args: unknown[]): void {
  console.log(`${".  ".repeat(this.indent)}${sprintf(this.fset, (pkg: unknown) => this.qualifier(pkg as Package), true, format, ...args)}`);
});

// ndigits returns the number of decimal digits in x.
// For x < 10, the result is always 1.
// For x > 100, the result is always 3.
export function ndigits(x: number): number {
  switch (true) {
    case x < 10:
      return 1;
    case x < 100:
      return 2;
    default:
      return 3;
  }
}

// dump is only needed for debugging
registerCheckerMethod("dump", function dump(format: string, ...args: unknown[]): void {
  console.log(sprintf(this.fset, (pkg: unknown) => this.qualifier(pkg as Package), true, format, ...args));
});

registerCheckerMethod("qualifier", function qualifier(pkg: Package): string {
  // Qualify the package unless it's the package being type-checked.
  if (pkg !== this.pkg) {
    if (this.pkgPathMap === null) {
      this.pkgPathMap = new Map();
      this.seenPkgMap = new Map();
      this.markImports(this.pkg);
    }
    // If the same package name was used by multiple packages, display the full path.
    if ((this.pkgPathMap.get(pkg.name)?.size ?? 0) > 1) {
      return JSON.stringify(pkg.path);
    }
    return pkg.name;
  }
  return "";
});

// markImports recursively walks pkg and its imports, to record unique import
// paths in pkgPathMap.
registerCheckerMethod("markImports", function markImports(pkg: Package): void {
  if (this.seenPkgMap?.get(pkg)) {
    return;
  }
  this.seenPkgMap?.set(pkg, true);

  let forName = this.pkgPathMap?.get(pkg.name);
  if (forName === undefined) {
    forName = new Map();
    this.pkgPathMap?.set(pkg.name, forName);
  }
  forName.set(pkg.path, true);

  for (const imp of pkg.imports) {
    this.markImports(imp);
  }
});

// stripAnnotations removes internal (type) annotations from s.
export function stripAnnotations(s: string): string {
  let buf = "";
  for (const r of s) {
    // strip #'s and subscript digits
    if (r < "₀" || "₀".codePointAt(0)! + 10 <= r.codePointAt(0)!) {
      buf += r;
    }
  }
  if (buf.length < s.length) {
    return buf;
  }
  return s;
}
