import {
  NewPackage,
  type Package as GoTypesPackage,
  type ScopeTransaction as GoTypesScopeTransaction
} from "./go/types/index.js";
import type { PackageRuntime } from "./runtime.js";

export type CodebaseTxnKind = "view" | "update";

interface PackageUpdate {
  readonly importPath: string;
  readonly target: GoTypesPackage;
  readonly candidate: GoTypesPackage;
  readonly scopeTransaction: GoTypesScopeTransaction;
  commit(): GoTypesPackage;
  rollback(): void;
}

export class Codebase {
  private readonly packageInfos = new Map<string, GoTypesPackage>();
  private readonly packageRuntimes = new Map<string, PackageRuntime>();

  public NewViewTxn(): CodebaseViewTxn {
    return new CodebaseViewTxn(this);
  }

  public NewUpdateTxn(): CodebaseUpdateTxn {
    return new CodebaseUpdateTxn(this);
  }

  public PackageInfo(txn: CodebaseTxn, importPath: string): GoTypesPackage | undefined {
    this.assertTxn(txn);
    return txn.packageInfo(importPath);
  }

  public PackageRuntime(txn: CodebaseTxn, importPath: string): PackageRuntime | undefined {
    this.assertTxn(txn);
    return txn.packageRuntime(importPath);
  }

  public PackageInfosRecord(txn: CodebaseTxn): Record<string, GoTypesPackage> {
    this.assertTxn(txn);
    return Object.fromEntries(txn.packageInfoEntries());
  }

  public PackageRuntimesRecord(txn: CodebaseTxn): Record<string, PackageRuntime> {
    this.assertTxn(txn);
    return Object.fromEntries(txn.packageRuntimeEntries());
  }

  public SetPackageInfo(txn: CodebaseUpdateTxn, importPath: string, pkg: GoTypesPackage): void {
    this.assertUpdateTxn(txn);
    txn.setPackageInfo(importPath, pkg);
  }

  public SetPackageRuntime(txn: CodebaseUpdateTxn, importPath: string, runtime: PackageRuntime): void {
    this.assertUpdateTxn(txn);
    txn.setPackageRuntime(importPath, runtime);
  }

  public BeginPackageUpdate(txn: CodebaseUpdateTxn, importPath: string, packageName: string): GoTypesPackage {
    this.assertUpdateTxn(txn);
    return txn.beginPackageUpdate(importPath, packageName);
  }

  public RollbackOnErrors(txn: CodebaseTxn | undefined, diagnostics: { severity: string }[]): void {
    if (!txn || txn.Kind() !== "update" || txn.IsClosed()) return;
    if (diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
      txn.Rollback();
    }
  }

  internalPackageInfo(importPath: string): GoTypesPackage | undefined {
    return this.packageInfos.get(importPath);
  }

  internalPackageRuntime(importPath: string): PackageRuntime | undefined {
    return this.packageRuntimes.get(importPath);
  }

  internalPackageInfoEntries(): [string, GoTypesPackage][] {
    return [...this.packageInfos.entries()];
  }

  internalPackageRuntimeEntries(): [string, PackageRuntime][] {
    return [...this.packageRuntimes.entries()];
  }

  internalCommitPackageInfo(importPath: string, pkg: GoTypesPackage): void {
    this.packageInfos.set(importPath, pkg);
  }

  internalCommitPackageRuntime(importPath: string, runtime: PackageRuntime): void {
    this.packageRuntimes.set(importPath, runtime);
  }

  private assertTxn(txn: CodebaseTxn): void {
    if (txn.Codebase() !== this) {
      throw new Error("Codebase transaction belongs to a different Codebase");
    }
    txn.assertOpen();
  }

  private assertUpdateTxn(txn: CodebaseUpdateTxn): void {
    this.assertTxn(txn);
    if (txn.Kind() !== "update") {
      throw new Error("Codebase update requires an update transaction");
    }
  }
}

export abstract class CodebaseTxn {
  protected closed = false;

  protected constructor(
    protected readonly codebase: Codebase,
    protected readonly kind: CodebaseTxnKind
  ) {}

  public Codebase(): Codebase {
    return this.codebase;
  }

  public Kind(): CodebaseTxnKind {
    return this.kind;
  }

  public IsClosed(): boolean {
    return this.closed;
  }

  public PackageInfo(importPath: string): GoTypesPackage | undefined {
    return this.codebase.PackageInfo(this, importPath);
  }

  public PackageRuntime(importPath: string): PackageRuntime | undefined {
    return this.codebase.PackageRuntime(this, importPath);
  }

  public PackageInfosRecord(): Record<string, GoTypesPackage> {
    return this.codebase.PackageInfosRecord(this);
  }

  public PackageRuntimesRecord(): Record<string, PackageRuntime> {
    return this.codebase.PackageRuntimesRecord(this);
  }

  public Rollback(): void {
    this.assertOpen();
    this.close();
  }

  public assertOpen(): void {
    if (this.closed) throw new Error("Codebase transaction already closed");
  }

  abstract packageInfo(importPath: string): GoTypesPackage | undefined;
  abstract packageRuntime(importPath: string): PackageRuntime | undefined;
  abstract packageInfoEntries(): [string, GoTypesPackage][];
  abstract packageRuntimeEntries(): [string, PackageRuntime][];

  protected close(): void {
    this.closed = true;
  }
}

export class CodebaseViewTxn extends CodebaseTxn {
  public constructor(codebase: Codebase) {
    super(codebase, "view");
  }

  public packageInfo(importPath: string): GoTypesPackage | undefined {
    this.assertOpen();
    return this.codebase.internalPackageInfo(importPath);
  }

  public packageRuntime(importPath: string): PackageRuntime | undefined {
    this.assertOpen();
    return this.codebase.internalPackageRuntime(importPath);
  }

  public packageInfoEntries(): [string, GoTypesPackage][] {
    this.assertOpen();
    return this.codebase.internalPackageInfoEntries();
  }

  public packageRuntimeEntries(): [string, PackageRuntime][] {
    this.assertOpen();
    return this.codebase.internalPackageRuntimeEntries();
  }
}

export class CodebaseUpdateTxn extends CodebaseTxn {
  private readonly packageInfoOverlay = new Map<string, GoTypesPackage>();
  private readonly packageRuntimeOverlay = new Map<string, PackageRuntime>();
  private readonly packageUpdates = new Map<string, PackageUpdate>();

  public constructor(codebase: Codebase) {
    super(codebase, "update");
  }

  public SetPackageInfo(importPath: string, pkg: GoTypesPackage): void {
    this.codebase.SetPackageInfo(this, importPath, pkg);
  }

  public SetPackageRuntime(importPath: string, runtime: PackageRuntime): void {
    this.codebase.SetPackageRuntime(this, importPath, runtime);
  }

  public BeginPackageUpdate(importPath: string, packageName: string): GoTypesPackage {
    return this.codebase.BeginPackageUpdate(this, importPath, packageName);
  }

  public packageInfo(importPath: string): GoTypesPackage | undefined {
    this.assertOpen();
    return this.packageInfoOverlay.get(importPath) ?? this.codebase.internalPackageInfo(importPath);
  }

  public packageRuntime(importPath: string): PackageRuntime | undefined {
    this.assertOpen();
    return this.packageRuntimeOverlay.get(importPath) ?? this.codebase.internalPackageRuntime(importPath);
  }

  public packageInfoEntries(): [string, GoTypesPackage][] {
    this.assertOpen();
    return mergedEntries(this.codebase.internalPackageInfoEntries(), this.packageInfoOverlay);
  }

  public packageRuntimeEntries(): [string, PackageRuntime][] {
    this.assertOpen();
    return mergedEntries(this.codebase.internalPackageRuntimeEntries(), this.packageRuntimeOverlay);
  }

  public setPackageInfo(importPath: string, pkg: GoTypesPackage): void {
    this.assertOpen();
    this.packageInfoOverlay.set(importPath, pkg);
  }

  public setPackageRuntime(importPath: string, runtime: PackageRuntime): void {
    this.assertOpen();
    this.packageRuntimeOverlay.set(importPath, runtime);
  }

  public beginPackageUpdate(importPath: string, packageName: string): GoTypesPackage {
    this.assertOpen();
    const existing = this.packageUpdates.get(importPath);
    if (existing) return existing.candidate;

    const target = this.packageInfo(importPath) ?? NewPackage(importPath, packageName);
    const scopeTransaction = target.Scope().BeginTransaction();
    const candidate = NewPackage(target.Path(), target.Name() || packageName);
    copyPackageState(target, candidate);
    candidate.scope = scopeTransaction.Scope();
    if (!candidate.Name() && packageName) candidate.SetName(packageName);

    const update = packageUpdate(importPath, target, candidate, scopeTransaction);
    this.packageUpdates.set(importPath, update);
    this.packageInfoOverlay.set(importPath, candidate);
    return candidate;
  }

  public Commit(): void {
    this.assertOpen();
    for (const update of this.packageUpdates.values()) {
      this.packageInfoOverlay.set(update.importPath, update.commit());
    }
    for (const [importPath, pkg] of this.packageInfoOverlay.entries()) {
      this.codebase.internalCommitPackageInfo(importPath, pkg);
    }
    for (const [importPath, runtime] of this.packageRuntimeOverlay.entries()) {
      this.codebase.internalCommitPackageRuntime(importPath, runtime);
    }
    this.close();
  }

  public override Rollback(): void {
    this.assertOpen();
    for (const update of this.packageUpdates.values()) update.rollback();
    this.packageUpdates.clear();
    this.packageInfoOverlay.clear();
    this.packageRuntimeOverlay.clear();
    this.close();
  }
}

function packageUpdate(
  importPath: string,
  target: GoTypesPackage,
  candidate: GoTypesPackage,
  scopeTransaction: GoTypesScopeTransaction
): PackageUpdate {
  let closed = false;
  return {
    importPath,
    target,
    candidate,
    scopeTransaction,
    commit(): GoTypesPackage {
      if (closed) throw new Error("package transaction already closed");
      scopeTransaction.Commit();
      copyPackageState(candidate, target);
      closed = true;
      return target;
    },
    rollback(): void {
      if (closed) return;
      scopeTransaction.Rollback();
      closed = true;
    }
  };
}

function copyPackageState(source: GoTypesPackage, target: GoTypesPackage): void {
  target.name = source.name;
  target.imports = [...source.imports];
  target.complete = source.complete;
  target.fake = source.fake;
  target.cgo = source.cgo;
  target.goVersion = source.goVersion;
}

function mergedEntries<K, V>(parentEntries: [K, V][], overlay: Map<K, V>): [K, V][] {
  const merged = new Map<K, V>(parentEntries);
  for (const [key, value] of overlay.entries()) merged.set(key, value);
  return [...merged.entries()];
}
