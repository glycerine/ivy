import { describe, expect, test } from "./testHarness.js";
import { Codebase } from "../src/codebase.js";
import { checkGoJuniorSourceFiles } from "../src/typecheck.js";
import {
  Int64,
  NewPackage,
  NewVar,
  NoPos,
  String as GoTypesString,
  Typ,
  ensureUniverseInitialized
} from "../src/go/types/index.js";

describe("Codebase transactions", () => {
  test("view transactions read accepted packages but cannot update", () => {
    ensureUniverseInitialized();
    const codebase = new Codebase();
    const update = codebase.NewUpdateTxn();
    const pkg = NewPackage("example.com/p", "p");
    codebase.SetPackageInfo(update, "example.com/p", pkg);
    update.Commit();

    const view = codebase.NewViewTxn();

    expect(codebase.PackageInfo(view, "example.com/p")).toBe(pkg);
    expect(() => codebase.SetPackageInfo(view as never, "example.com/q", NewPackage("example.com/q", "q"))).toThrow(/update transaction/);

    view.Rollback();
    expect(() => codebase.PackageInfo(view, "example.com/p")).toThrow(/already closed/);
  });

  test("update transactions shadow package scope changes until commit", () => {
    ensureUniverseInitialized();
    const codebase = new Codebase();
    const seed = codebase.NewUpdateTxn();
    const accepted = codebase.BeginPackageUpdate(seed, "example.com/p", "p");
    const original = NewVar(NoPos, accepted, "Value", Typ[Int64]!);
    expect(accepted.Scope().Insert(original)).toBeNull();
    seed.Commit();

    const update = codebase.NewUpdateTxn();
    const candidate = codebase.BeginPackageUpdate(update, "example.com/p", "p");
    const replacement = NewVar(NoPos, candidate, "Value", Typ[GoTypesString]!);
    expect(candidate.Scope().Insert(replacement)).toBeNull();

    const viewBeforeCommit = codebase.NewViewTxn();
    expect(codebase.PackageInfo(viewBeforeCommit, "example.com/p")?.Scope().Lookup("Value")).toBe(original);
    viewBeforeCommit.Rollback();
    expect(candidate.Scope().Lookup("Value")).toBe(replacement);

    update.Commit();

    const viewAfterCommit = codebase.NewViewTxn();
    expect(codebase.PackageInfo(viewAfterCommit, "example.com/p")?.Scope().Lookup("Value")).toBe(replacement);
  });

  test("rollback discards all package updates in the transaction", () => {
    ensureUniverseInitialized();
    const codebase = new Codebase();
    const seed = codebase.NewUpdateTxn();
    const accepted = codebase.BeginPackageUpdate(seed, "example.com/p", "p");
    const original = NewVar(NoPos, accepted, "Value", Typ[Int64]!);
    expect(accepted.Scope().Insert(original)).toBeNull();
    seed.Commit();

    const update = codebase.NewUpdateTxn();
    const first = codebase.BeginPackageUpdate(update, "example.com/p", "p");
    const replacement = NewVar(NoPos, first, "Value", Typ[GoTypesString]!);
    expect(first.Scope().Insert(replacement)).toBeNull();
    const second = codebase.BeginPackageUpdate(update, "example.com/q", "q");
    const added = NewVar(NoPos, second, "Added", Typ[Int64]!);
    expect(second.Scope().Insert(added)).toBeNull();

    update.Rollback();

    const view = codebase.NewViewTxn();
    expect(codebase.PackageInfo(view, "example.com/p")?.Scope().Lookup("Value")).toBe(original);
    expect(codebase.PackageInfo(view, "example.com/q")).toBeUndefined();
  });

  test("rejects transactions from another codebase", () => {
    const first = new Codebase();
    const second = new Codebase();
    const tx = first.NewViewTxn();

    expect(() => second.PackageInfo(tx, "example.com/p")).toThrow(/different Codebase/);
  });

  test("typecheck errors rollback the active update transaction", () => {
    const codebase = new Codebase();
    const goodTxn = codebase.NewUpdateTxn();
    const good = checkGoJuniorSourceFiles([{
      filename: "p.go",
      source: "package p\nfunc F() int { return 1 }\n"
    }], {
      codebaseTxn: goodTxn,
      packagePath: "example.com/p",
      packageName: "p",
      autoImportFmt: false
    });
    expect(good.diagnostics).toEqual([]);
    expect(goodTxn.IsClosed()).toBe(false);
    goodTxn.Commit();

    const badTxn = codebase.NewUpdateTxn();
    const bad = checkGoJuniorSourceFiles([{
      filename: "p.go",
      source: "package p\nfunc Bad() { return 1 }\n"
    }], {
      codebaseTxn: badTxn,
      packagePath: "example.com/p",
      packageName: "p",
      autoImportFmt: false
    });
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(badTxn.IsClosed()).toBe(true);

    const view = codebase.NewViewTxn();
    const accepted = codebase.PackageInfo(view, "example.com/p");
    expect(accepted?.Scope().Lookup("F")).toBeDefined();
    expect(accepted?.Scope().Lookup("Bad")).toBeNull();
  });
});
