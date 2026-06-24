// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ----------------------------------------------------------------------------
// API

import { NewIdent, PosOf, type BasicLit, type Expr, type Ident, type StructType } from "../../front/ast.js";
import { TokenKind } from "../../front/token.js";
import { Basic, Invalid, UnsafePointer } from "./basic.js";
import { Checker, atPos , registerCheckerMethod } from "./check.js";
import { deref } from "./lookup.js";
import { Interface } from "./interface.js";
import { Pointer } from "./pointer.js";
import type { Type } from "./type.js";
import { TypeString } from "./typestring.js";
import { NewField, type Object, type Var } from "./object.js";
import { objset } from "./objset.js";
import { isTypeParam, isValid } from "./predicates.js";
import { Typ } from "./universe.js";
import { makeFromLiteral } from "./util.js";

declare module "./check.js" {
  interface Checker {
    declareInSet(oset: objset, pos: number, obj: Object): boolean;
    tag(t: BasicLit | undefined): string;
  }
}

// A Struct represents a struct type.
export class Struct implements Type {
  public fields: Var[] | null = null; // fields != nil indicates the struct is set up (possibly with len(fields) == 0)
  public tags: string[] | null = null; // field tags; nil if there are no tags

  public constructor(fields: Var[] | null = null, tags: string[] | null = null) {
    this.fields = fields;
    this.tags = tags;
  }

  // NumFields returns the number of fields in the struct (including blank and embedded fields).
  public NumFields(): number { return this.fields?.length ?? 0; }

  // Field returns the i'th field for 0 <= i < NumFields().
  public Field(i: number): Var { return this.fields![i]!; }

  // Tag returns the i'th field tag for 0 <= i < NumFields().
  public Tag(i: number): string {
    if (this.tags !== null && i < this.tags.length) {
      return this.tags[i]!;
    }
    return "";
  }

  public Underlying(): Type { return this; }
  public String(): string { return TypeString(this, null); }

  // ----------------------------------------------------------------------------
  // Implementation

  public markComplete(): void {
    if (this.fields === null) {
      this.fields = [];
    }
  }
}

// NewStruct returns a new struct with the given fields and corresponding field tags.
// If a field with index i has a tag, tags[i] must be that tag, but len(tags) may be
// only as long as required to hold the tag with the largest index i. Consequently,
// if no field has a tag, tags may be nil.
export function NewStruct(fields: Var[], tags: string[] | null): Struct {
  const fset = new objset();
  for (const f of fields) {
    if (f.name !== "_" && fset.insert(f) !== null) {
      throw new Error("multiple fields with the same name");
    }
  }
  if (tags !== null && tags.length > fields.length) {
    throw new Error("more tags than fields");
  }
  const s = new Struct(fields, tags);
  s.markComplete();
  return s;
}

registerCheckerMethod("structType", function structType(styp: Struct, e: StructType): void {
  const list = e.fields;
  if (list === undefined || list.fields.length === 0) {
    styp.markComplete();
    return;
  }

  // struct fields and tags
  const fields: Var[] = [];
  let tags: string[] | null = null;

  // for double-declaration checks
  const fset = new objset();

  // current field typ and tag
  let typ: Type = Typ[Invalid]!;
  let tag = "";
  const add = (ident: Ident, embedded: boolean): void => {
    if (tag !== "" && tags === null) {
      tags = new Array<string>(fields.length).fill("");
    }
    if (tags !== null) {
      tags.push(tag);
    }

    const pos = PosOf(ident);
    const name = ident.name;
    const fld = NewField(pos, this.pkg, name, typ, embedded);
    // spec: "Within a struct, non-blank field names must be unique."
    if (name === "_" || this.declareInSet(fset, pos, fld)) {
      fields.push(fld);
      this.recordDef(ident, fld);
    }
  };

  // addInvalid adds an embedded field of invalid type to the struct for
  // fields with errors; this keeps the number of struct fields in sync
  // with the source as long as the fields are _ or have different names
  // (go.dev/issue/25627).
  const addInvalid = (ident: Ident): void => {
    typ = Typ[Invalid]!;
    tag = "";
    add(ident, true);
  };

  for (const f of list.fields) {
    typ = this.varType(f.type);
    tag = this.tag(f.tag);
    if (f.names.length > 0) {
      // named fields
      for (const name of f.names) {
        add(name, false);
      }
    } else {
      // embedded field
      // spec: "An embedded type must be specified as a type name T or as a
      // pointer to a non-interface type name *T, and T itself may not be a
      // pointer type."
      const pos = PosOf(f.type); // position of type, for errors
      let name = embeddedFieldIdent(f.type);
      if (name === null) {
        this.errorf(new atPos(PosOf(f.type)), "InvalidSyntaxTree", "embedded field type %s has no name", f.type);
        name = NewIdent("_");
        if (f.type.span !== undefined) {
          name.span = f.type.span;
        }
        addInvalid(name);
        continue;
      }
      add(name, true); // struct{p.T} field has position of T

      // Because we have a name, typ must be of the form T or *T, where T is the name
      // of a (named or alias) type, and t (= deref(typ)) must be the type of T.
      // We must delay this check to the end because we don't want to instantiate
      // (via t.Underlying()) a possibly incomplete type.

      // for use in the closure below
      const embeddedTyp = typ;
      const embeddedPos = f.type;

      this.later(() => {
        const [t, isPtr] = deref(embeddedTyp);
        const u = t.Underlying();
        if (u instanceof Basic) {
          if (!isValid(t)) {
            // error was reported before
            return;
          }
          // unsafe.Pointer is treated like a regular pointer
          if (u.kind === UnsafePointer) {
            this.error(new atPos(PosOf(embeddedPos)), "InvalidPtrEmbed", "embedded field type cannot be unsafe.Pointer");
          }
        } else if (u instanceof Pointer) {
          this.error(new atPos(PosOf(embeddedPos)), "InvalidPtrEmbed", "embedded field type cannot be a pointer");
        } else if (u instanceof Interface) {
          if (isTypeParam(t)) {
            // The error code here is inconsistent with other error codes for
            // invalid embedding, because this restriction may be relaxed in the
            // future, and so it did not warrant a new error code.
            this.error(new atPos(PosOf(embeddedPos)), "MisplacedTypeParam", "embedded field type cannot be a (pointer to a) type parameter");
          } else if (isPtr) {
            this.error(new atPos(PosOf(embeddedPos)), "InvalidPtrEmbed", "embedded field type cannot be a pointer to an interface");
          }
        }
      }).describef(new atPos(PosOf(embeddedPos)), "check embedded type %s", embeddedTyp);
      void pos;
    }
  }

  styp.fields = fields;
  styp.tags = tags;
  styp.markComplete();
});

export function embeddedFieldIdent(e: Expr): Ident | null {
  switch (e.kind) {
    case "Ident":
      return e;
    case "StarExpr":
      // *T is valid, but **T is not
      if (e.expr.kind !== "StarExpr") {
        return embeddedFieldIdent(e.expr);
      }
      break;
    case "SelectorExpr":
      return e.selector;
    case "IndexExpr":
      return embeddedFieldIdent(e.object);
    case "IndexListExpr":
      return embeddedFieldIdent(e.object);
  }
  return null; // invalid embedded field
}

registerCheckerMethod("declareInSet", function declareInSet(oset: objset, pos: number, obj: Object): boolean {
  const alt = oset.insert(obj);
  if (alt !== null) {
    const err = this.newError("DuplicateDecl");
    err.addf(new atPos(pos), "%s redeclared", obj.Name());
    err.addAltDecl(alt);
    err.report();
    return false;
  }
  return true;
});

registerCheckerMethod("tag", function tag(t: BasicLit | undefined): string {
  if (t !== undefined) {
    if (t.token === TokenKind.StringLiteral) {
      const val = makeFromLiteral(t.value, t.token);
      if (typeof val === "string") {
        return val;
      }
    }
    this.errorf(new atPos(PosOf(t)), "InvalidSyntaxTree", "incorrect tag syntax: %q", t.value);
  }
  return "";
});
