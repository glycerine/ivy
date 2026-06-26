export interface Transaction<K, V> {
  readonly environment: Environment<K, V>;
  commit(): void;
  rollback(): void;
}

type LookupResult<V> =
  | { found: true; value: V }
  | { found: false };

export class Environment<K, V> {
  private readonly values = new Map<K, V>();
  private readonly deleted = new Set<K>();
  private closed = false;

  public constructor(
    private readonly parent?: Environment<K, V>,
    private readonly transactional = false
  ) {}

  public static fromEntries<K, V>(entries: Iterable<readonly [K, V]>): Environment<K, V> {
    const environment = new Environment<K, V>();
    for (const [key, value] of entries) {
      environment.set(key, value);
    }
    return environment;
  }

  public beginTransaction(): Transaction<K, V> {
    return new EnvironmentTransaction(this, new Environment<K, V>(this, true));
  }

  public get(key: K): V | undefined {
    const result = this.lookup(key);
    return result.found ? result.value : undefined;
  }

  public has(key: K): boolean {
    return this.lookup(key).found;
  }

  public set(key: K, value: V): void {
    this.ensureOpen();
    this.deleted.delete(key);
    this.values.set(key, value);
  }

  public delete(key: K): void {
    this.ensureOpen();
    this.values.delete(key);
    if (this.parent) {
      this.deleted.add(key);
    }
  }

  private lookup(key: K): LookupResult<V> {
    if (this.deleted.has(key)) return { found: false };
    if (this.values.has(key)) return { found: true, value: this.values.get(key)! };
    return this.parent?.lookup(key) ?? { found: false };
  }

  public commitOverlayFromTransaction(overlay: Environment<K, V>): void {
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

  public closeFromTransaction(): void {
    this.closed = true;
    this.values.clear();
    this.deleted.clear();
  }

  private ensureOpen(): void {
    if (this.closed) {
      throw new Error(this.transactional ? "closed transaction environment" : "closed environment");
    }
  }
}

class EnvironmentTransaction<K, V> implements Transaction<K, V> {
  private closed = false;

  public constructor(
    private readonly parent: Environment<K, V>,
    public readonly environment: Environment<K, V>
  ) {}

  public commit(): void {
    this.ensureOpen();
    this.parent.commitOverlayFromTransaction(this.environment);
    this.closed = true;
  }

  public rollback(): void {
    this.ensureOpen();
    this.environment.closeFromTransaction();
    this.closed = true;
  }

  private ensureOpen(): void {
    if (this.closed) {
      throw new Error("transaction already closed");
    }
  }
}
