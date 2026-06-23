import { DeterministicPrng } from "./prng.js";

export interface AsyncGoSchedulerOptions {
  randomSeed?: number | string | bigint;
}

export class AsyncGoDeadlockError extends Error {
  public constructor(message: string) {
    super(message);
    this.name = "AsyncGoDeadlockError";
  }
}

export class AsyncGoPanic extends Error {
  public constructor(
    message: string,
    public readonly value: unknown = message
  ) {
    super(message);
    this.name = "AsyncGoPanic";
  }
}

export class AsyncGoScheduler {
  private readonly prng: DeterministicPrng;
  private liveGoroutines = 0;
  private blockedGoroutines = 0;
  private readonly errors: unknown[] = [];
  private readonly signalWaiters: Array<() => void> = [];

  public constructor(options: AsyncGoSchedulerOptions = {}) {
    this.prng = new DeterministicPrng(options.randomSeed);
  }

  public go(fn: () => Promise<void> | void): void {
    this.liveGoroutines += 1;
    queueMicrotask(() => {
      Promise.resolve()
        .then(fn)
        .catch((error: unknown) => {
          this.errors.push(error);
        })
        .finally(() => {
          this.liveGoroutines -= 1;
          this.signal();
        });
    });
    this.signal();
  }

  public async run(): Promise<void> {
    while (true) {
      await Promise.resolve();
      this.throwFirstError();
      if (this.liveGoroutines === 0) return;
      if (this.blockedGoroutines >= this.liveGoroutines) {
        throw new AsyncGoDeadlockError("all goroutines are asleep - deadlock");
      }
      await this.waitForSignal();
    }
  }

  public blockOn<T>(promise: Promise<T>): Promise<T> {
    this.blockedGoroutines += 1;
    this.signal();
    return promise.finally(() => {
      this.blockedGoroutines -= 1;
      this.signal();
    });
  }

  public blockForever<T>(): Promise<T> {
    return this.blockOn(new Promise<T>(() => undefined));
  }

  public randomIndex(length: number): number {
    return this.prng.nextIndex(length);
  }

  private throwFirstError(): void {
    if (this.errors.length === 0) return;
    throw this.errors[0];
  }

  private waitForSignal(): Promise<void> {
    return new Promise((resolve) => {
      this.signalWaiters.push(resolve);
    });
  }

  private signal(): void {
    const waiters = this.signalWaiters.splice(0);
    for (const resolve of waiters) queueMicrotask(resolve);
  }
}

interface SendWaiter<T> {
  value: T;
  resolve: () => void;
  reject: (error: unknown) => void;
}

interface ReceiveWaiter<T> {
  resolve: (result: [T, boolean]) => void;
  reject: (error: unknown) => void;
}

export class AsyncGoChannel<T> {
  private readonly buffer: T[] = [];
  private readonly sendWaiters: Array<SendWaiter<T>> = [];
  private readonly receiveWaiters: Array<ReceiveWaiter<T>> = [];
  private readonly changeWaiters = new Set<() => void>();
  private closed = false;

  public constructor(
    private readonly scheduler: AsyncGoScheduler,
    private readonly capacityValue = 0,
    private readonly zeroValue: () => T
  ) {
    if (!Number.isInteger(capacityValue) || capacityValue < 0) {
      throw new RangeError("channel capacity must be a non-negative integer");
    }
  }

  public send(value: T): Promise<void> {
    if (this.trySend(value)) return Promise.resolve();
    const pending = new Promise<void>((resolve, reject) => {
      this.sendWaiters.push({ value, resolve, reject });
      this.notifyChange();
    });
    return this.scheduler.blockOn(pending);
  }

  public receive(): Promise<[T, boolean]> {
    const ready = this.tryReceive();
    if (ready) return Promise.resolve(ready);
    const pending = new Promise<[T, boolean]>((resolve, reject) => {
      this.receiveWaiters.push({ resolve, reject });
      this.notifyChange();
    });
    return this.scheduler.blockOn(pending);
  }

  public close(): void {
    if (this.closed) throw new AsyncGoPanic("close of closed channel");
    this.closed = true;
    while (this.sendWaiters.length > 0) {
      this.sendWaiters.shift()?.reject(new AsyncGoPanic("send on closed channel"));
    }
    this.pump();
    this.notifyChange();
  }

  public canSendNow(): boolean {
    return this.closed || this.receiveWaiters.length > 0 || this.buffer.length < this.capacityValue;
  }

  public canReceiveNow(): boolean {
    return this.buffer.length > 0 || this.sendWaiters.length > 0 || this.closed;
  }

  public trySend(value: T): boolean {
    if (this.closed) throw new AsyncGoPanic("send on closed channel");
    const receiver = this.receiveWaiters.shift();
    if (receiver) {
      receiver.resolve([value, true]);
      this.notifyChange();
      return true;
    }
    if (this.buffer.length < this.capacityValue) {
      this.buffer.push(value);
      this.notifyChange();
      return true;
    }
    return false;
  }

  public tryReceive(): [T, boolean] | undefined {
    if (this.buffer.length > 0) {
      const value = this.buffer.shift() as T;
      this.pump();
      this.notifyChange();
      return [value, true];
    }
    const sender = this.sendWaiters.shift();
    if (sender) {
      sender.resolve();
      this.notifyChange();
      return [sender.value, true];
    }
    if (this.closed) return [this.zeroValue(), false];
    return undefined;
  }

  public len(): number {
    return this.buffer.length;
  }

  public cap(): number {
    return this.capacityValue;
  }

  public onChange(callback: () => void): () => void {
    this.changeWaiters.add(callback);
    return () => {
      this.changeWaiters.delete(callback);
    };
  }

  private pump(): void {
    let changed = false;

    while (this.receiveWaiters.length > 0 && this.buffer.length > 0) {
      const receiver = this.receiveWaiters.shift();
      const value = this.buffer.shift() as T;
      receiver?.resolve([value, true]);
      changed = true;
    }

    while (this.receiveWaiters.length > 0 && this.sendWaiters.length > 0) {
      const receiver = this.receiveWaiters.shift();
      const sender = this.sendWaiters.shift();
      if (!receiver || !sender) break;
      sender.resolve();
      receiver.resolve([sender.value, true]);
      changed = true;
    }

    while (!this.closed && this.sendWaiters.length > 0 && this.buffer.length < this.capacityValue) {
      const sender = this.sendWaiters.shift();
      if (!sender) break;
      this.buffer.push(sender.value);
      sender.resolve();
      changed = true;
    }

    if (this.closed && this.buffer.length === 0) {
      while (this.receiveWaiters.length > 0) {
        this.receiveWaiters.shift()?.resolve([this.zeroValue(), false]);
        changed = true;
      }
    }

    if (changed) this.notifyChange();
  }

  private notifyChange(): void {
    for (const callback of [...this.changeWaiters]) queueMicrotask(callback);
  }
}

export type AsyncSelectCase<T> =
  | { op: "send"; channel: AsyncGoChannel<T> | null | undefined; value: T }
  | { op: "receive"; channel: AsyncGoChannel<T> | null | undefined }
  | { op: "default" };

export type AsyncSelectResult<T> =
  | { index: number; op: "send" }
  | { index: number; op: "receive"; value: T; ok: boolean }
  | { index: number; op: "default" };

export async function asyncSelect<T>(
  scheduler: AsyncGoScheduler,
  cases: Array<AsyncSelectCase<T>>
): Promise<AsyncSelectResult<T>> {
  while (true) {
    const ready: Array<() => AsyncSelectResult<T>> = [];
    let defaultIndex: number | undefined;
    const watchedChannels = new Set<AsyncGoChannel<T>>();

    for (const [index, item] of cases.entries()) {
      if (item.op === "default") {
        defaultIndex = index;
        continue;
      }
      const channel = item.channel;
      if (!channel) continue;
      watchedChannels.add(channel);
      if (item.op === "send" && channel.canSendNow()) {
        ready.push(() => {
          channel.trySend(item.value);
          return { index, op: "send" };
        });
      }
      if (item.op === "receive" && channel.canReceiveNow()) {
        ready.push(() => {
          const received = channel.tryReceive();
          if (!received) throw new AsyncGoDeadlockError("select receive case was no longer ready");
          return { index, op: "receive", value: received[0], ok: received[1] };
        });
      }
    }

    if (ready.length > 0) return ready[scheduler.randomIndex(ready.length)]!();
    if (defaultIndex !== undefined) return { index: defaultIndex, op: "default" };
    if (watchedChannels.size === 0) return scheduler.blockForever();

    await scheduler.blockOn(waitForAnyChannelChange(watchedChannels));
  }
}

function waitForAnyChannelChange<T>(channels: Iterable<AsyncGoChannel<T>>): Promise<void> {
  return new Promise((resolve) => {
    const unsubscribers: Array<() => void> = [];
    const done = () => {
      for (const unsubscribe of unsubscribers.splice(0)) unsubscribe();
      resolve();
    };
    for (const channel of channels) {
      unsubscribers.push(channel.onChange(done));
    }
  });
}
