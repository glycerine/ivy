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
export class AsyncGoScheduler {
    prng;
    liveGoroutines = 0;
    blockedGoroutines = 0;
    errors = [];
    signalWaiters = [];
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
    blockOn(promise) {
        this.blockedGoroutines += 1;
        this.signal();
        return promise.finally(() => {
            this.blockedGoroutines -= 1;
            this.signal();
        });
    }
    blockForever() {
        return this.blockOn(new Promise(() => undefined));
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
        const pending = new Promise((resolve, reject) => {
            this.sendWaiters.push({ value, resolve, reject });
            this.notifyChange();
        });
        return this.scheduler.blockOn(pending);
    }
    receive() {
        const ready = this.tryReceive();
        if (ready)
            return Promise.resolve(ready);
        const pending = new Promise((resolve, reject) => {
            this.receiveWaiters.push({ resolve, reject });
            this.notifyChange();
        });
        return this.scheduler.blockOn(pending);
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
        await scheduler.blockOn(waitForAnyChannelChange(watchedChannels));
    }
}
function waitForAnyChannelChange(channels) {
    return new Promise((resolve) => {
        const unsubscribers = [];
        const done = () => {
            for (const unsubscribe of unsubscribers.splice(0))
                unsubscribe();
            resolve();
        };
        for (const channel of channels) {
            unsubscribers.push(channel.onChange(done));
        }
    });
}
