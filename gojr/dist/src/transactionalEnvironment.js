export class Environment {
    parent;
    transactional;
    values = new Map();
    deleted = new Set();
    closed = false;
    constructor(parent, transactional = false) {
        this.parent = parent;
        this.transactional = transactional;
    }
    static fromEntries(entries) {
        const environment = new Environment();
        for (const [key, value] of entries) {
            environment.set(key, value);
        }
        return environment;
    }
    beginTransaction() {
        return new EnvironmentTransaction(this, new Environment(this, true));
    }
    get(key) {
        const result = this.lookup(key);
        return result.found ? result.value : undefined;
    }
    has(key) {
        return this.lookup(key).found;
    }
    set(key, value) {
        this.ensureOpen();
        this.deleted.delete(key);
        this.values.set(key, value);
    }
    delete(key) {
        this.ensureOpen();
        this.values.delete(key);
        if (this.parent) {
            this.deleted.add(key);
        }
    }
    lookup(key) {
        if (this.deleted.has(key))
            return { found: false };
        if (this.values.has(key))
            return { found: true, value: this.values.get(key) };
        return this.parent?.lookup(key) ?? { found: false };
    }
    commitOverlayFromTransaction(overlay) {
        this.ensureOpen();
        overlay.ensureOpen();
        for (const key of overlay.deleted) {
            this.values.delete(key);
            this.deleted.delete(key);
        }
        for (const [key, value] of overlay.values) {
            this.deleted.delete(key);
            this.values.set(key, value);
        }
        overlay.closeFromTransaction();
    }
    closeFromTransaction() {
        this.closed = true;
        this.values.clear();
        this.deleted.clear();
    }
    ensureOpen() {
        if (this.closed) {
            throw new Error(this.transactional ? "closed transaction environment" : "closed environment");
        }
    }
}
class EnvironmentTransaction {
    parent;
    environment;
    closed = false;
    constructor(parent, environment) {
        this.parent = parent;
        this.environment = environment;
    }
    commit() {
        this.ensureOpen();
        this.parent.commitOverlayFromTransaction(this.environment);
        this.closed = true;
    }
    rollback() {
        this.ensureOpen();
        this.environment.closeFromTransaction();
        this.closed = true;
    }
    ensureOpen() {
        if (this.closed) {
            throw new Error("transaction already closed");
        }
    }
}
