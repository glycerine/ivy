// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements error reporting.

import { Checker, bailout, nopos, noposn, type positioner } from "./check.js";
import { Error as TypesError } from "./api.js";
import type { Object } from "./object.js";
import { operand } from "./operand.js";
import { stripAnnotations } from "./format.js";
import type { goVersion } from "./version.js";

function assert(p: boolean): asserts p {
  if (!p) {
    let msg = "assertion failed";
    throw new Error(msg);
  }
}

// An errorDesc describes part of a type-checking error.
export class errorDesc {
  public constructor(
    public posn: positioner,
    public msg: string
  ) {}
}

// An error_ represents a type-checking error.
// A new error_ is created with Checker.newError.
// To report an error_, call error_.report.
export class error_ {
  public desc: errorDesc[] | null = null;
  public soft = false; // TODO(gri) eventually determine this from an error code

  public constructor(
    public check: Checker,
    public code: unknown
  ) {}

  // addf adds formatted error information to err.
  // It may be called multiple times to provide additional information.
  // The position of the first call to addf determines the position of the reported Error.
  // Subsequent calls to addf provide additional information in the form of additional lines
  // in the error message (types2) or continuation errors identified by a tab-indented error
  // message (go/types).
  public addf(at: positioner | Object | unknown, format: string, ...args: unknown[]): void {
    if (this.desc === null) {
      this.desc = [];
    }
    this.desc.push(new errorDesc(asPositioner(at), this.check.sprintf(format, ...args)));
  }

  // addAltDecl is a specialized form of addf reporting another declaration of obj.
  public addAltDecl(obj: Object): void {
    const pos = obj.Pos();
    if (pos !== nopos) {
      // We use "other" rather than "previous" here because
      // the first declaration seen may not be textually
      // earlier in the source.
      this.addf(obj, "other declaration of %s", obj.Name());
    }
  }

  public empty(): boolean {
    return this.desc === null;
  }

  public posn(): positioner {
    if (this.empty()) {
      return noposn;
    }
    return this.desc![0]!.posn;
  }

  // msg returns the formatted error message without the primary error position pos().
  public msg(): string {
    if (this.empty()) {
      return "no error";
    }

    const buf: string[] = [];
    for (let i = 0; i < this.desc!.length; i++) {
      const p = this.desc![i]!;
      if (i > 0) {
        buf.push("\n\t");
        if (p.posn.Pos() !== nopos) {
          buf.push(`${p.posn.Pos()}: `);
        }
      }
      buf.push(p.msg);
    }
    return buf.join("");
  }

  // report reports the error err, setting check.firstError if necessary.
  public report(): void {
    if (this.empty()) {
      throw new Error("no error");
    }

    const check = this.check;
    if (check.firstErr !== null) {
      const msg = this.desc![0]!.msg;
      if (msg.indexOf("invalid operand") > 0 || msg.indexOf("invalid type") > 0) {
        return;
      }
    }

    if (check.conf._Trace) {
      check.trace(this.posn().Pos(), "ERROR: %s (code = %d)", this.desc![0]!.msg, this.code);
    }

    let multiError = false;
    for (let i = 1; i < this.desc!.length; i++) {
      if (this.desc![i]!.posn.Pos() !== nopos) {
        multiError = true;
        break;
      }
    }

    if (multiError) {
      for (let i = 0; i < this.desc!.length; i++) {
        const p = this.desc![i]!;
        check.handleError(i, p.posn, this.code, p.msg, this.soft);
      }
    } else {
      check.handleError(0, this.posn(), this.code, this.msg(), this.soft);
    }

    this.desc = null;
  }
}

declare module "./check.js" {
  interface Checker {
    newError(code: unknown): error_;
    handleError(index: number, posn: positioner, code: unknown, msg: string, soft: boolean): void;
    error(at: positioner | Object | unknown, code: unknown, msg: string): void;
    errorf(at: positioner | Object | unknown, code: unknown, format: string, ...args: unknown[]): void;
    softErrorf(at: positioner | Object | unknown, code: unknown, format: string, ...args: unknown[]): void;
    versionErrorf(at: positioner | Object | unknown, v: goVersion, format: string, ...args: unknown[]): void;
  }
}

// newError returns a new error_ with the given error code.
Checker.prototype.newError = function newError(code: unknown): error_ {
  if (code === 0 || code === null || code === undefined) {
    throw new Error("error code must not be 0");
  }
  return new error_(this, code);
};

// handleError should only be called by error_.report.
Checker.prototype.handleError = function handleError(index: number, posn: positioner, code: unknown, msg: string, soft: boolean): void {
  assert(code !== 0);

  if (index === 0) {
    if (this.errpos !== null && this.errpos.Pos() !== nopos) {
      assert(this.iota !== null);
      posn = this.errpos;
    }

    if (code === "InvalidSyntaxTree") {
      msg = "invalid syntax tree: " + msg;
    }

    if (this.conf._ErrorURL !== "") {
      const url = this.conf._ErrorURL.replace("%s", String(code));
      const i = msg.indexOf("\n");
      if (i >= 0) {
        msg = msg.slice(0, i) + url + msg.slice(i);
      } else {
        msg += url;
      }
    }
  } else {
    msg = "\t" + msg;
  }

  const span = spanOf(posn);
  const e = new TypesError(
    this.fset as { Position(pos: unknown): string } | null,
    span.pos,
    stripAnnotations(msg),
    soft,
    code,
    span.start,
    span.end
  );

  if (this.errpos !== null) {
    const span = spanOf(this.errpos);
    e.Pos = span.pos;
    e.go116start = span.start;
    e.go116end = span.end;
  }

  if (this.firstErr === null) {
    this.firstErr = e;
  }

  const f = this.conf.Error;
  if (f === null) {
    throw new bailout(); // record first error and exit
  }
  f(e);
};

export const invalidArg = "invalid argument: ";
export const invalidOp = "invalid operation: ";

Checker.prototype.error = function error(at: positioner | Object | unknown, code: unknown, msg: string): void {
  const err = this.newError(code);
  err.addf(at, "%s", msg);
  err.report();
};

Checker.prototype.errorf = function errorf(at: positioner | Object | unknown, code: unknown, format: string, ...args: unknown[]): void {
  const err = this.newError(code);
  err.addf(at, format, ...args);
  err.report();
};

Checker.prototype.softErrorf = function softErrorf(at: positioner | Object | unknown, code: unknown, format: string, ...args: unknown[]): void {
  const err = this.newError(code);
  err.addf(at, format, ...args);
  err.soft = true;
  err.report();
};

Checker.prototype.versionErrorf = function versionErrorf(at: positioner | Object | unknown, v: goVersion, format: string, ...args: unknown[]): void {
  const msg = this.sprintf(format, ...args);
  const err = this.newError("UnsupportedFeature");
  err.addf(at, "%s requires %s or later", msg, v);
  err.report();
};

// posSpan holds a position range along with a highlighted position within that
// range. This is used for positioning errors, with pos by convention being the
// first position in the source where the error is known to exist, and start
// and end defining the full span of syntax being considered when the error was
// detected. Invariant: start <= pos < end || start == pos == end.
export class posSpan implements positioner {
  public constructor(
    public start: number,
    public pos: number,
    public end: number
  ) {}

  public Pos(): number {
    return this.pos;
  }
}

// inNode creates a posSpan for the given node.
// Invariant: node.Pos() <= pos < node.End() (node.End() is the position of the
// first byte after node within the source).
export function inNode(node: { Pos: () => number; End: () => number }, pos: number): posSpan {
  const start = node.Pos();
  const end = node.End();
  if (false) {
    assert(start <= pos && pos < end);
  }
  return new posSpan(start, pos, end);
}

// spanOf extracts an error span from the given positioner. By default this is
// the trivial span starting and ending at pos, but this span is expanded when
// the argument naturally corresponds to a span of source code.
export function spanOf(at: positioner | null): posSpan {
  if (at === null) {
    throw new Error("nil positioner");
  }
  if (at instanceof posSpan) {
    return at;
  }
  if (at instanceof operand) {
    if (at.expr !== null) {
      const pos = at.Pos();
      return new posSpan(pos, pos, pos);
    }
    return new posSpan(nopos, nopos, nopos);
  }
  const pos = at.Pos();
  return new posSpan(pos, pos, pos);
}

function asPositioner(at: positioner | Object | unknown): positioner {
  if (at !== null && typeof at === "object" && "Pos" in at && typeof at.Pos === "function") {
    return at as positioner;
  }
  return noposn;
}
