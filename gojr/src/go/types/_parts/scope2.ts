// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements go/types-specific scope methods.
// These methods do not exist in types2.

import { cmpPos } from "./check.js";
import type { Object } from "./object.js";
import { Scope, scopeUniverse } from "./scope.js";
import type { Pos } from "./token.js";

declare module "./scope.js" {
  interface Scope {
    LookupParent(name: string, pos: Pos): [Scope | null, Object | null];
    Pos(): Pos;
    End(): Pos;
    Contains(pos: Pos): boolean;
    Innermost(pos: Pos): Scope | null;
  }
}

// LookupParent follows the parent chain of scopes starting with s until
// it finds a scope where Lookup(name) returns a non-nil object, and then
// returns that scope and object. If a valid position pos is provided,
// only objects that were declared at or before pos are considered.
// If no such scope and object exists, the result is (nil, nil).
// The results are guaranteed to be valid only if the type-checked
// AST has complete position information.
//
// Note that obj.Parent() may be different from the returned scope if the
// object was inserted into the scope and already had a parent at that
// time (see Insert). This can only happen for dot-imported objects
// whose parent is the scope of the package that exported them.
Scope.prototype.LookupParent = function LookupParent(name: string, pos: Pos): [Scope | null, Object | null] {
  for (let s: Scope | null = this; s !== null; s = s.parent) {
    const obj = s.Lookup(name);
    if (obj !== null && (pos === 0 || cmpPos(obj.scopePos(), pos) <= 0)) {
      return [s, obj];
    }
  }
  return [null, null];
};

// Pos and End describe the scope's source code extent [pos, end).
// The results are guaranteed to be valid only if the type-checked
// AST has complete position information. The extent is undefined
// for Universe and package scopes.
Scope.prototype.Pos = function Pos_(): Pos { return this.pos; };
Scope.prototype.End = function End_(): Pos { return this.end; };

// Contains reports whether pos is within the scope's extent.
// The result is guaranteed to be valid only if the type-checked
// AST has complete position information.
Scope.prototype.Contains = function Contains(pos: Pos): boolean {
  return cmpPos(this.pos, pos) <= 0 && cmpPos(pos, this.end) < 0;
};

// Innermost returns the innermost (child) scope containing
// pos. If pos is not within any scope, the result is nil.
// The result is also nil for the Universe scope.
// The result is guaranteed to be valid only if the type-checked
// AST has complete position information.
Scope.prototype.Innermost = function Innermost(pos: Pos): Scope | null {
  // Package scopes do not have extents since they may be
  // discontiguous, so iterate over the package's files.
  if (this.parent === scopeUniverse()) {
    for (const s of this.children) {
      const inner = s.Innermost(pos);
      if (inner !== null) {
        return inner;
      }
    }
  }

  if (this.Contains(pos)) {
    for (const s of this.children) {
      if (s.Contains(pos)) {
        return s.Innermost(pos);
      }
    }
    return this;
  }
  return null;
};
