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

class AsyncGoCancellationScope {
  private readonly callbacks = new Set<(error: unknown) => void>();
  private cancelled = false;
  private cancellationError: unknown;

  public register(callback: (error: unknown) => void): () => void {
    if (this.cancelled) {
      callback(this.cancellationError);
      return () => undefined;
    }
    this.callbacks.add(callback);
    return () => {
      this.callbacks.delete(callback);
    };
  }

  public cancel(error: unknown): void {
    if (this.cancelled) return;
    this.cancelled = true;
    this.cancellationError = error;
    const callbacks = [...this.callbacks];
    this.callbacks.clear();
    for (const callback of callbacks) callback(error);
  }
}

export class AsyncGoScheduler {
  private readonly prng: DeterministicPrng;
  private liveGoroutines = 0;
  private blockedGoroutines = 0;
  private readonly errors: unknown[] = [];
  private readonly signalWaiters: Array<() => void> = [];
  private currentCancellationScope: AsyncGoCancellationScope | undefined;

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

  public async runRoot<T>(fn: () => Promise<T> | T): Promise<T> {
    let done = false;
    let value: T | undefined;
    let failure: unknown;
    const previousCancellationScope = this.currentCancellationScope;
    const cancellationScope = new AsyncGoCancellationScope();
    this.currentCancellationScope = cancellationScope;

    try {
      this.liveGoroutines += 1;
      queueMicrotask(() => {
        Promise.resolve()
          .then(fn)
          .then((result) => {
            value = result;
          }, (error: unknown) => {
            failure = error;
          })
          .finally(() => {
            done = true;
            this.liveGoroutines -= 1;
            this.signal();
          });
      });
      this.signal();

      while (!done) {
        await Promise.resolve();
        this.throwFirstError();
        if (this.blockedGoroutines >= this.liveGoroutines) {
          cancellationScope.cancel(new AsyncGoDeadlockError("all goroutines are asleep - deadlock"));
        }
        await this.waitForSignal();
      }

      this.throwFirstError();
      if (failure !== undefined) throw failure;
      return value as T;
    } finally {
      this.currentCancellationScope = previousCancellationScope;
    }
  }

  public blockOn<T>(promise: Promise<T>, cancel?: (error: unknown) => void): Promise<T> {
    this.blockedGoroutines += 1;
    this.signal();
    const unregister = cancel && this.currentCancellationScope
      ? this.currentCancellationScope.register(cancel)
      : undefined;
    return promise.finally(() => {
      unregister?.();
      this.blockedGoroutines -= 1;
      this.signal();
    });
  }

  public blockForever<T>(): Promise<T> {
    let cancel: ((error: unknown) => void) | undefined;
    const pending = new Promise<T>((_resolve, reject) => {
      cancel = reject;
    });
    return this.blockOn(pending, (error) => cancel?.(error));
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
    let waiter: SendWaiter<T> | undefined;
    const pending = new Promise<void>((resolve, reject) => {
      waiter = { value, resolve, reject };
      this.sendWaiters.push(waiter);
      this.notifyChange();
    });
    return this.scheduler.blockOn(pending, (error) => {
      if (waiter && removeWaiter(this.sendWaiters, waiter)) {
        waiter.reject(error);
        this.notifyChange();
      }
    });
  }

  public receive(): Promise<[T, boolean]> {
    const ready = this.tryReceive();
    if (ready) return Promise.resolve(ready);
    let waiter: ReceiveWaiter<T> | undefined;
    const pending = new Promise<[T, boolean]>((resolve, reject) => {
      waiter = { resolve, reject };
      this.receiveWaiters.push(waiter);
      this.notifyChange();
    });
    return this.scheduler.blockOn(pending, (error) => {
      if (waiter && removeWaiter(this.receiveWaiters, waiter)) {
        waiter.reject(error);
        this.notifyChange();
      }
    });
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

    const wait = waitForAnyChannelChange(watchedChannels);
    await scheduler.blockOn(wait.promise, wait.cancel);
  }
}

function waitForAnyChannelChange<T>(channels: Iterable<AsyncGoChannel<T>>): {
  promise: Promise<void>;
  cancel: (error: unknown) => void;
} {
  const unsubscribers: Array<() => void> = [];
  let settle: (() => void) | undefined;
  let rejectWait: ((error: unknown) => void) | undefined;
  const cleanup = () => {
    for (const unsubscribe of unsubscribers.splice(0)) unsubscribe();
  };
  const promise = new Promise<void>((resolve, reject) => {
    settle = resolve;
    rejectWait = reject;
  });
  const done = () => {
    cleanup();
    settle?.();
  };
  for (const channel of channels) {
    unsubscribers.push(channel.onChange(done));
  }
  return {
    promise,
    cancel(error) {
      cleanup();
      rejectWait?.(error);
    }
  };
}

function removeWaiter<T>(items: T[], item: T): boolean {
  const index = items.indexOf(item);
  if (index < 0) return false;
  items.splice(index, 1);
  return true;
}
