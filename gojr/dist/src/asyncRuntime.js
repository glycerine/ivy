import { DeterministicPrng } from "./prng.js";
export class AsyncGoDeadlockError extends Error {
    constructor(message) {
        super(message);
        this.name = "AsyncGoDeadlockError";
    }
}
export class AsyncGoPanic extends Error {
    value;
    constructor(message, value = message) {
        super(message);
        this.value = value;
        this.name = "AsyncGoPanic";
    }
}
class AsyncGoCancellationScope {
    callbacks = new Set();
    cancelled = false;
    cancellationError;
    register(callback) {
        if (this.cancelled) {
            callback(this.cancellationError);
            return () => undefined;
        }
        this.callbacks.add(callback);
        return () => {
            this.callbacks.delete(callback);
        };
    }
    cancel(error) {
        if (this.cancelled)
            return;
        this.cancelled = true;
        this.cancellationError = error;
        const callbacks = [...this.callbacks];
        this.callbacks.clear();
        for (const callback of callbacks)
            callback(error);
    }
}
export class AsyncGoScheduler {
    prng;
    liveGoroutines = 0;
    blockedGoroutines = 0;
    errors = [];
    signalWaiters = [];
    currentCancellationScope;
    constructor(options = {}) {
        this.prng = new DeterministicPrng(options.randomSeed);
    }
    go(fn) {
        this.liveGoroutines += 1;
        queueMicrotask(() => {
            Promise.resolve()
                .then(fn)
                .catch((error) => {
                this.errors.push(error);
            })
                .finally(() => {
                this.liveGoroutines -= 1;
                this.signal();
            });
        });
        this.signal();
    }
    async run() {
        while (true) {
            await Promise.resolve();
            this.throwFirstError();
            if (this.liveGoroutines === 0)
                return;
            if (this.blockedGoroutines >= this.liveGoroutines) {
                throw new AsyncGoDeadlockError("all goroutines are asleep - deadlock");
            }
            await this.waitForSignal();
        }
    }
    async runRoot(fn) {
        let done = false;
        let value;
        let failure;
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
                }, (error) => {
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
            if (failure !== undefined)
                throw failure;
            return value;
        }
        finally {
            this.currentCancellationScope = previousCancellationScope;
        }
    }
    blockOn(promise, cancel) {
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
    blockForever() {
        let cancel;
        const pending = new Promise((_resolve, reject) => {
            cancel = reject;
        });
        return this.blockOn(pending, (error) => cancel?.(error));
    }
    randomIndex(length) {
        return this.prng.nextIndex(length);
    }
    throwFirstError() {
        if (this.errors.length === 0)
            return;
        throw this.errors[0];
    }
    waitForSignal() {
        return new Promise((resolve) => {
            this.signalWaiters.push(resolve);
        });
    }
    signal() {
        const waiters = this.signalWaiters.splice(0);
        for (const resolve of waiters)
            queueMicrotask(resolve);
    }
}
export class AsyncGoChannel {
    scheduler;
    capacityValue;
    zeroValue;
    buffer = [];
    sendWaiters = [];
    receiveWaiters = [];
    changeWaiters = new Set();
    closed = false;
    constructor(scheduler, capacityValue = 0, zeroValue) {
        this.scheduler = scheduler;
        this.capacityValue = capacityValue;
        this.zeroValue = zeroValue;
        if (!Number.isInteger(capacityValue) || capacityValue < 0) {
            throw new RangeError("channel capacity must be a non-negative integer");
        }
    }
    send(value) {
        if (this.trySend(value))
            return Promise.resolve();
        let waiter;
        const pending = new Promise((resolve, reject) => {
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
    receive() {
        const ready = this.tryReceive();
        if (ready)
            return Promise.resolve(ready);
        let waiter;
        const pending = new Promise((resolve, reject) => {
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
    close() {
        if (this.closed)
            throw new AsyncGoPanic("close of closed channel");
        this.closed = true;
        while (this.sendWaiters.length > 0) {
            this.sendWaiters.shift()?.reject(new AsyncGoPanic("send on closed channel"));
        }
        this.pump();
        this.notifyChange();
    }
    canSendNow() {
        return this.closed || this.receiveWaiters.length > 0 || this.buffer.length < this.capacityValue;
    }
    canReceiveNow() {
        return this.buffer.length > 0 || this.sendWaiters.length > 0 || this.closed;
    }
    trySend(value) {
        if (this.closed)
            throw new AsyncGoPanic("send on closed channel");
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
    tryReceive() {
        if (this.buffer.length > 0) {
            const value = this.buffer.shift();
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
        if (this.closed)
            return [this.zeroValue(), false];
        return undefined;
    }
    len() {
        return this.buffer.length;
    }
    cap() {
        return this.capacityValue;
    }
    onChange(callback) {
        this.changeWaiters.add(callback);
        return () => {
            this.changeWaiters.delete(callback);
        };
    }
    pump() {
        let changed = false;
        while (this.receiveWaiters.length > 0 && this.buffer.length > 0) {
            const receiver = this.receiveWaiters.shift();
            const value = this.buffer.shift();
            receiver?.resolve([value, true]);
            changed = true;
        }
        while (this.receiveWaiters.length > 0 && this.sendWaiters.length > 0) {
            const receiver = this.receiveWaiters.shift();
            const sender = this.sendWaiters.shift();
            if (!receiver || !sender)
                break;
            sender.resolve();
            receiver.resolve([sender.value, true]);
            changed = true;
        }
        while (!this.closed && this.sendWaiters.length > 0 && this.buffer.length < this.capacityValue) {
            const sender = this.sendWaiters.shift();
            if (!sender)
                break;
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
        if (changed)
            this.notifyChange();
    }
    notifyChange() {
        for (const callback of [...this.changeWaiters])
            queueMicrotask(callback);
    }
}
export async function asyncSelect(scheduler, cases) {
    while (true) {
        const ready = [];
        let defaultIndex;
        const watchedChannels = new Set();
        for (const [index, item] of cases.entries()) {
            if (item.op === "default") {
                defaultIndex = index;
                continue;
            }
            const channel = item.channel;
            if (!channel)
                continue;
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
                    if (!received)
                        throw new AsyncGoDeadlockError("select receive case was no longer ready");
                    return { index, op: "receive", value: received[0], ok: received[1] };
                });
            }
        }
        if (ready.length > 0)
            return ready[scheduler.randomIndex(ready.length)]();
        if (defaultIndex !== undefined)
            return { index: defaultIndex, op: "default" };
        if (watchedChannels.size === 0)
            return scheduler.blockForever();
        const wait = waitForAnyChannelChange(watchedChannels);
        await scheduler.blockOn(wait.promise, wait.cancel);
    }
}
function waitForAnyChannelChange(channels) {
    const unsubscribers = [];
    let settle;
    let rejectWait;
    const cleanup = () => {
        for (const unsubscribe of unsubscribers.splice(0))
            unsubscribe();
    };
    const promise = new Promise((resolve, reject) => {
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
function removeWaiter(items, item) {
    const index = items.indexOf(item);
    if (index < 0)
        return false;
    items.splice(index, 1);
    return true;
}
