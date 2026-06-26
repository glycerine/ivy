import { NewPackage } from "./go/types/index.js";
export class Codebase {
    packageInfos = new Map();
    packageRuntimes = new Map();
    NewViewTxn() {
        return new CodebaseViewTxn(this);
    }
    NewUpdateTxn() {
        return new CodebaseUpdateTxn(this);
    }
    PackageInfo(txn, importPath) {
        this.assertTxn(txn);
        return txn.packageInfo(importPath);
    }
    PackageRuntime(txn, importPath) {
        this.assertTxn(txn);
        return txn.packageRuntime(importPath);
    }
    PackageInfosRecord(txn) {
        this.assertTxn(txn);
        return Object.fromEntries(txn.packageInfoEntries());
    }
    PackageRuntimesRecord(txn) {
        this.assertTxn(txn);
        return Object.fromEntries(txn.packageRuntimeEntries());
    }
    SetPackageInfo(txn, importPath, pkg) {
        this.assertUpdateTxn(txn);
        txn.setPackageInfo(importPath, pkg);
    }
    SetPackageRuntime(txn, importPath, runtime) {
        this.assertUpdateTxn(txn);
        txn.setPackageRuntime(importPath, runtime);
    }
    BeginPackageUpdate(txn, importPath, packageName) {
        this.assertUpdateTxn(txn);
        return txn.beginPackageUpdate(importPath, packageName);
    }
    RollbackOnErrors(txn, diagnostics) {
        if (!txn || txn.Kind() !== "update" || txn.IsClosed())
            return;
        if (diagnostics.some((diagnostic) => diagnostic.severity === "error")) {
            txn.Rollback();
        }
    }
    internalPackageInfo(importPath) {
        return this.packageInfos.get(importPath);
    }
    internalPackageRuntime(importPath) {
        return this.packageRuntimes.get(importPath);
    }
    internalPackageInfoEntries() {
        return [...this.packageInfos.entries()];
    }
    internalPackageRuntimeEntries() {
        return [...this.packageRuntimes.entries()];
    }
    internalCommitPackageInfo(importPath, pkg) {
        this.packageInfos.set(importPath, pkg);
    }
    internalCommitPackageRuntime(importPath, runtime) {
        this.packageRuntimes.set(importPath, runtime);
    }
    assertTxn(txn) {
        if (txn.Codebase() !== this) {
            throw new Error("Codebase transaction belongs to a different Codebase");
        }
        txn.assertOpen();
    }
    assertUpdateTxn(txn) {
        this.assertTxn(txn);
        if (txn.Kind() !== "update") {
            throw new Error("Codebase update requires an update transaction");
        }
    }
}
export class CodebaseTxn {
    codebase;
    kind;
    closed = false;
    constructor(codebase, kind) {
        this.codebase = codebase;
        this.kind = kind;
    }
    Codebase() {
        return this.codebase;
    }
    Kind() {
        return this.kind;
    }
    IsClosed() {
        return this.closed;
    }
    PackageInfo(importPath) {
        return this.codebase.PackageInfo(this, importPath);
    }
    PackageRuntime(importPath) {
        return this.codebase.PackageRuntime(this, importPath);
    }
    PackageInfosRecord() {
        return this.codebase.PackageInfosRecord(this);
    }
    PackageRuntimesRecord() {
        return this.codebase.PackageRuntimesRecord(this);
    }
    Rollback() {
        this.assertOpen();
        this.close();
    }
    assertOpen() {
        if (this.closed)
            throw new Error("Codebase transaction already closed");
    }
    close() {
        this.closed = true;
    }
}
export class CodebaseViewTxn extends CodebaseTxn {
    constructor(codebase) {
        super(codebase, "view");
    }
    packageInfo(importPath) {
        this.assertOpen();
        return this.codebase.internalPackageInfo(importPath);
    }
    packageRuntime(importPath) {
        this.assertOpen();
        return this.codebase.internalPackageRuntime(importPath);
    }
    packageInfoEntries() {
        this.assertOpen();
        return this.codebase.internalPackageInfoEntries();
    }
    packageRuntimeEntries() {
        this.assertOpen();
        return this.codebase.internalPackageRuntimeEntries();
    }
}
export class CodebaseUpdateTxn extends CodebaseTxn {
    packageInfoOverlay = new Map();
    packageRuntimeOverlay = new Map();
    packageUpdates = new Map();
    constructor(codebase) {
        super(codebase, "update");
    }
    SetPackageInfo(importPath, pkg) {
        this.codebase.SetPackageInfo(this, importPath, pkg);
    }
    SetPackageRuntime(importPath, runtime) {
        this.codebase.SetPackageRuntime(this, importPath, runtime);
    }
    BeginPackageUpdate(importPath, packageName) {
        return this.codebase.BeginPackageUpdate(this, importPath, packageName);
    }
    packageInfo(importPath) {
        this.assertOpen();
        return this.packageInfoOverlay.get(importPath) ?? this.codebase.internalPackageInfo(importPath);
    }
    packageRuntime(importPath) {
        this.assertOpen();
        return this.packageRuntimeOverlay.get(importPath) ?? this.codebase.internalPackageRuntime(importPath);
    }
    packageInfoEntries() {
        this.assertOpen();
        return mergedEntries(this.codebase.internalPackageInfoEntries(), this.packageInfoOverlay);
    }
    packageRuntimeEntries() {
        this.assertOpen();
        return mergedEntries(this.codebase.internalPackageRuntimeEntries(), this.packageRuntimeOverlay);
    }
    setPackageInfo(importPath, pkg) {
        this.assertOpen();
        this.packageInfoOverlay.set(importPath, pkg);
    }
    setPackageRuntime(importPath, runtime) {
        this.assertOpen();
        this.packageRuntimeOverlay.set(importPath, runtime);
    }
    beginPackageUpdate(importPath, packageName) {
        this.assertOpen();
        const existing = this.packageUpdates.get(importPath);
        if (existing)
            return existing.candidate;
        const target = this.packageInfo(importPath) ?? NewPackage(importPath, packageName);
        const scopeTransaction = target.Scope().BeginTransaction();
        const candidate = NewPackage(target.Path(), target.Name() || packageName);
        copyPackageState(target, candidate);
        candidate.scope = scopeTransaction.Scope();
        if (!candidate.Name() && packageName)
            candidate.SetName(packageName);
        const update = packageUpdate(importPath, target, candidate, scopeTransaction);
        this.packageUpdates.set(importPath, update);
        this.packageInfoOverlay.set(importPath, candidate);
        return candidate;
    }
    Commit() {
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
    Rollback() {
        this.assertOpen();
        for (const update of this.packageUpdates.values())
            update.rollback();
        this.packageUpdates.clear();
        this.packageInfoOverlay.clear();
        this.packageRuntimeOverlay.clear();
        this.close();
    }
}
function packageUpdate(importPath, target, candidate, scopeTransaction) {
    let closed = false;
    return {
        importPath,
        target,
        candidate,
        scopeTransaction,
        commit() {
            if (closed)
                throw new Error("package transaction already closed");
            scopeTransaction.Commit();
            copyPackageState(candidate, target);
            closed = true;
            return target;
        },
        rollback() {
            if (closed)
                return;
            scopeTransaction.Rollback();
            closed = true;
        }
    };
}
function copyPackageState(source, target) {
    target.name = source.name;
    target.imports = [...source.imports];
    target.complete = source.complete;
    target.fake = source.fake;
    target.cgo = source.cgo;
    target.goVersion = source.goVersion;
}
function mergedEntries(parentEntries, overlay) {
    const merged = new Map(parentEntries);
    for (const [key, value] of overlay.entries())
        merged.set(key, value);
    return [...merged.entries()];
}
