import { describe, expect, test } from "./testHarness.js";
import {
  Int64,
  NewScope,
  NewVar,
  NoPos,
  String as GoTypesString,
  Typ,
  ensureUniverseInitialized
} from "../src/go/types/index.js";

describe("go/types transactional Scope", () => {
  test("transaction overlay reads accepted bindings and shadows them locally", () => {
    ensureUniverseInitialized();
    const accepted = NewScope(null, NoPos, NoPos, "accepted package");
    const original = NewVar(NoPos, null, "f", Typ[Int64]!);
    const replacement = NewVar(NoPos, null, "f", Typ[GoTypesString]!);

    expect(accepted.Insert(original)).toBeNull();

    const tx = accepted.BeginTransaction();
    const candidate = tx.Scope();

    expect(candidate.Lookup("f")).toBe(original);
    expect(candidate.Insert(replacement)).toBeNull();
    expect(candidate.Lookup("f")).toBe(replacement);
    expect(accepted.Lookup("f")).toBe(original);

    tx.Rollback();

    expect(accepted.Lookup("f")).toBe(original);
  });

  test("commit applies overlay additions replacements and deletions", () => {
    ensureUniverseInitialized();
    const accepted = NewScope(null, NoPos, NoPos, "accepted package");
    const original = NewVar(NoPos, null, "f", Typ[Int64]!);
    const replacement = NewVar(NoPos, null, "f", Typ[GoTypesString]!);
    const added = NewVar(NoPos, null, "g", Typ[Int64]!);

    expect(accepted.Insert(original)).toBeNull();

    const tx = accepted.BeginTransaction();
    const candidate = tx.Scope();
    expect(candidate.Insert(replacement)).toBeNull();
    expect(candidate.Insert(added)).toBeNull();
    candidate.Delete("missing");

    tx.Commit();

    expect(accepted.Lookup("f")).toBe(replacement);
    expect(accepted.Lookup("g")).toBe(added);
  });

  test("rollback discards overlay additions replacements and deletions", () => {
    ensureUniverseInitialized();
    const accepted = NewScope(null, NoPos, NoPos, "accepted package");
    const original = NewVar(NoPos, null, "f", Typ[Int64]!);
    const replacement = NewVar(NoPos, null, "f", Typ[GoTypesString]!);
    const added = NewVar(NoPos, null, "g", Typ[Int64]!);

    expect(accepted.Insert(original)).toBeNull();

    const tx = accepted.BeginTransaction();
    const candidate = tx.Scope();
    expect(candidate.Insert(replacement)).toBeNull();
    expect(candidate.Insert(added)).toBeNull();
    candidate.Delete("f");

    expect(candidate.Lookup("f")).toBeNull();
    expect(candidate.Lookup("g")).toBe(added);

    tx.Rollback();

    expect(accepted.Lookup("f")).toBe(original);
    expect(accepted.Lookup("g")).toBeNull();
  });
});
