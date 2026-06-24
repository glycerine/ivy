// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file defines a hash function for Types.

import { Alias, Unalias } from "./alias.js";
import { Array as ArrayType } from "./array.js";
import { Basic } from "./basic.js";
import { Chan } from "./chan.js";
import { Interface } from "./interface.js";
import { Map as MapType } from "./map.js";
import { Named } from "./named.js";
import type { TypeName } from "./object.js";
import { Pointer } from "./pointer.js";
import { Identical } from "./predicates.js";
import { IdenticalIgnoreTags } from "./api_predicates.js";
import { Signature } from "./signature.js";
import { Slice } from "./slice.js";
import { Struct } from "./struct.js";
import type { Type } from "./type.js";
import { Tuple } from "./tuple.js";
import { TypeParam } from "./typeparam.js";
import { Union } from "./union.js";

export class Hasher {
  // Hashers are stateless.
  public Hash(h: MapHash, t: Type): void {
    // The two hashers use essentially the same hash function,
    // which ignores tags; only the Equal methods vary.
    // But for future-proofing we gratuitously force them
    // to differ by one byte.
    h.WriteByte(0);
    new hasher(false).hash(h, t);
  }

  public Equal(x: Type, y: Type): boolean { return Identical(x, y); }
}

export class HasherIgnoreTags {
  public Hash(h: MapHash, t: Type): void {
    h.WriteByte(1);
    new hasher(false).hash(h, t);
  }

  public Equal(x: Type, y: Type): boolean { return IdenticalIgnoreTags(x, y); }
}

export const _ = [Hasher, HasherIgnoreTags];

// hasher holds the state of a single hash traversal, namely,
// whether we are inside the signature of a generic function.
// This is used to optimize [hasher.hashTypeParam].
export class hasher {
  public constructor(public inGenericSig: boolean) {}

  public hash(h: MapHash, t: Type | null): void {
    // See [Identical] for rationale.
    switch (true) {
      case t instanceof Alias:
        this.hash(h, Unalias(t));
        break;

      case t instanceof ArrayType:
        h.WriteByte("A");
        writeComparable(h, t.Len());
        this.hash(h, t.Elem());
        break;

      case t instanceof Basic:
        h.WriteByte("B");
        h.WriteByte(t.Kind());
        break;

      case t instanceof Chan:
        h.WriteByte("C");
        h.WriteByte(t.Dir());
        this.hash(h, t.Elem());
        break;

      case t instanceof Interface: {
        h.WriteByte("I");
        h.WriteByte(t.NumMethods());

        // Interfaces are identical if they have the same set of methods, with
        // identical names and types, and they have the same set of type
        // restrictions. See [Identical] for more details.

        // Hash the methods.
        //
        // Because [Identical] treats Methods as an unordered set,
        // we must either:
        // (a) sort the methods into some canonical order; or
        // (b) hash them each in parallel, combine them with a
        //     commutative operation such as + or ^, and then
        //     write this value into the primary hasher.
        // Since (a) requires allocation, we choose (b).
        let hash = 0n;
        for (let i = 0; i < t.NumMethods(); i++) {
          const m = t.Method(i);
          const subh = new MapHash();
          subh.SetSeed(h.Seed());
          // Ignore m.Pkg().
          // Use shallow hash on method signature to
          // avoid anonymous interface cycles.
          subh.WriteString(m.Name());
          this.shallowHash(subh, m.Type());
          hash ^= subh.Sum64();
        }
        writeComparable(h, hash);

        // Hash type restrictions.
        // TODO(adonovan): call (fork of) InterfaceTermSet from
        // golang.org/x/tools/internal/typeparams/normalize.go.
        // hr.hashTermSet(h, terms)
        break;
      }

      case t instanceof MapType:
        h.WriteByte("M");
        this.hash(h, t.Key());
        this.hash(h, t.Elem());
        break;

      case t instanceof Named:
        h.WriteByte("N");
        this.hashTypeName(h, t.Obj());
        for (const targ of t.TypeArgs()?.list() ?? []) {
          this.hash(h, targ);
        }
        break;

      case t instanceof Pointer:
        h.WriteByte("P");
        this.hash(h, t.Elem());
        break;

      case t instanceof Signature: {
        h.WriteByte("F");
        writeComparable(h, t.Variadic());
        const tparams = t.TypeParams();
        const n = tparams?.Len() ?? 0;
        const savedInGenericSig = this.inGenericSig;
        try {
          if (n > 0) {
            this.inGenericSig = true; // affects constraints, params, and results

            writeComparable(h, n);
            for (const tparam of tparams!.list()) {
              this.hash(h, tparam.Constraint());
            }
          }
          this.hashTuple(h, t.Params());
          this.hashTuple(h, t.Results());
        } finally {
          this.inGenericSig = savedInGenericSig;
        }
        break;
      }

      case t instanceof Slice:
        h.WriteByte("S");
        this.hash(h, t.Elem());
        break;

      case t instanceof Struct: {
        h.WriteByte("R"); // mnemonic: a struct is a record type
        const n = t.NumFields();
        h.WriteByte(n);
        for (let i = 0; i < n; i++) {
          const f = t.Field(i);
          writeComparable(h, f.Anonymous());
          // Ignore t.Tag(i), so that a single hash function
          // can be used with both [Identical] and [IdenticalIgnoreTags].
          h.WriteString(f.Name()); // (ignore f.Pkg)
          this.hash(h, f.Type());
        }
        break;
      }

      case t instanceof Tuple:
        this.hashTuple(h, t);
        break;

      case t instanceof TypeParam:
        this.hashTypeParam(h, t);
        break;

      case t instanceof Union:
        h.WriteByte("U");
        // TODO(adonovan): opt: call (fork of) UnionTermSet from
        // golang.org/x/tools/internal/typeparams/normalize.go.
        // hr.hashTermSet(h, terms)
        break;

      default:
        throw new Error(`${t?.constructor.name}: ${String(t)}`);
    }
  }

  public hashTuple(h: MapHash, t: Tuple | null): void {
    h.WriteByte("T");
    h.WriteByte(t?.Len() ?? 0);
    for (const v of t?.vars ?? []) {
      this.hash(h, v.Type());
    }
  }

  // hashTypeParam encodes a type parameter into hasher h.
  public hashTypeParam(h: MapHash, t: TypeParam): void {
    h.WriteByte("P");
    // Within the signature of a generic function, TypeParams are
    // identical if they have the same index and constraint, so we
    // hash them based on index.
    //
    // When we are outside a generic function, free TypeParams are
    // identical iff they are the same object, so we can use a
    // more discriminating hash consistent with object identity.
    // This optimization saves [Map] about 4% when hashing all the
    // Info.Types in the forward closure of net/http.
    if (!this.inGenericSig) {
      // Optimization: outside a generic function signature,
      // use a more discrimating hash consistent with object identity.
      this.hashTypeName(h, t.Obj());
    } else {
      h.WriteByte(t.Index());
    }
  }

  // hashTypeName hashes the pointer of tname.
  public hashTypeName(h: MapHash, tname: TypeName): void {
    h.WriteByte("N");
    // Since Identical uses == to compare TypeNames,
    // the hash function uses maphash.Comparable.
    writeComparable(h, tname);
  }

  // shallowHash computes a hash of t without looking at any of its
  // element Types, to avoid potential anonymous cycles in the types of
  // interface methods.
  //
  // When an unnamed non-empty interface type appears anywhere among the
  // arguments or results of an interface method, there is a potential
  // for endless recursion. Consider:
  //
  //	type X interface { m() []*interface { X } }
  //
  // The problem is that the Methods of the interface in m's result type
  // include m itself; there is no mention of the named type X that
  // might help us break the cycle.
  // (See comment in [Identical], case *Interface, for more.)
  public shallowHash(h: MapHash, t: Type | null): void {
    // t is the type of an interface method (Signature),
    // its params or results (Tuples), or their immediate
    // elements (mostly Slice, Pointer, Basic, Named),
    // so there's no need to optimize anything else.
    switch (true) {
      case t instanceof Alias:
        this.shallowHash(h, Unalias(t));
        break;

      case t instanceof ArrayType:
        h.WriteByte("A");
        writeComparable(h, t.Len());
        // ignore t.Elem()
        break;

      case t instanceof Basic:
        h.WriteByte("B");
        h.WriteByte(t.Kind());
        break;

      case t instanceof Chan:
        h.WriteByte("C");
        // ignore Dir(), Elem()
        break;

      case t instanceof Interface:
        h.WriteByte("I");
        // no recursion here
        break;

      case t instanceof MapType:
        h.WriteByte("M");
        // ignore Key(), Elem()
        break;

      case t instanceof Named:
        this.hashTypeName(h, t.Obj());
        break;

      case t instanceof Pointer:
        h.WriteByte("P");
        // ignore t.Elem()
        break;

      case t instanceof Signature:
        h.WriteByte(btoi(t.Variadic()));
        // The Signature/Tuple recursion is always
        // finite and invariably shallow.
        this.shallowHash(h, t.Params());
        this.shallowHash(h, t.Results());
        break;

      case t instanceof Slice:
        h.WriteByte("S");
        // ignore t.Elem()
        break;

      case t instanceof Struct:
        h.WriteByte("R"); // mnemonic: a struct is a record type
        h.WriteByte(t.NumFields());
        // ignore t.Fields()
        break;

      case t instanceof Tuple:
        h.WriteByte("T");
        h.WriteByte(t.Len());
        for (const v of t.vars) {
          this.shallowHash(h, v.Type());
        }
        break;

      case t instanceof TypeParam:
        this.hashTypeParam(h, t);
        break;

      case t instanceof Union:
        h.WriteByte("U");
        // ignore term set
        break;

      default:
        throw new Error(`shallowHash: ${t?.constructor.name}: ${String(t)}`);
    }
  }
}

export function btoi(b: boolean): number {
  if (b) {
    return 1;
  } else {
    return 0;
  }
}

export class MapHash {
  private seed = 0xcbf29ce484222325n;
  private sum = this.seed;

  public WriteByte(b: number | string): void {
    const n = typeof b === "string" ? b.charCodeAt(0) : b;
    this.writeUint(BigInt(n & 0xff));
  }

  public WriteString(s: string): void {
    for (let i = 0; i < s.length; i++) {
      this.WriteByte(s.charCodeAt(i));
    }
  }

  public SetSeed(seed: bigint): void {
    this.seed = seed;
    this.sum = seed;
  }

  public Seed(): bigint { return this.seed; }

  public Sum64(): bigint { return this.sum & 0xffffffffffffffffn; }

  private writeUint(x: bigint): void {
    this.sum ^= x & 0xffn;
    this.sum = (this.sum * 0x100000001b3n) & 0xffffffffffffffffn;
  }
}

const objectIDs = new WeakMap<object, number>();
let nextObjectID = 1;

function objectID(x: object): number {
  let id = objectIDs.get(x);
  if (id === undefined) {
    id = nextObjectID;
    nextObjectID++;
    objectIDs.set(x, id);
  }
  return id;
}

function writeComparable(h: MapHash, x: unknown): void {
  switch (typeof x) {
    case "boolean":
      h.WriteByte(btoi(x));
      return;
    case "number":
      h.WriteString(String(x));
      h.WriteByte(0);
      return;
    case "bigint":
      h.WriteString(x.toString());
      h.WriteByte(0);
      return;
    case "string":
      h.WriteString(x);
      h.WriteByte(0);
      return;
    case "object":
      if (x === null) {
        h.WriteString("<nil>");
      } else {
        h.WriteString(String(objectID(x)));
      }
      h.WriteByte(0);
      return;
    default:
      h.WriteString(String(x));
      h.WriteByte(0);
  }
}
