// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ----------------------------------------------------------------------------
// API

import type { Type } from "./type.js";
import { TypeString } from "./typestring.js";
import type { Var } from "./object.js";
import { objset } from "./objset.js";

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
