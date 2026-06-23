import { describe, expect, test } from "./testHarness.js";
import {
  AsyncGoChannel,
  AsyncGoDeadlockError,
  AsyncGoPanic,
  AsyncGoScheduler,
  asyncSelect
} from "../src/index.js";

const zeroNumber = () => 0;

describe("Go-junior async scheduler runtime", () => {
  test("hands off unbuffered channel sends and receives between goroutines", async () => {
    const scheduler = new AsyncGoScheduler();
    const ch = new AsyncGoChannel<number>(scheduler, 0, zeroNumber);
    const events: string[] = [];

    scheduler.go(async () => {
      events.push("send:start");
      await ch.send(7);
      events.push("send:done");
    });
    scheduler.go(async () => {
      events.push("recv:start");
      const [value, ok] = await ch.receive();
      events.push(`recv:${value}:${ok}`);
    });

    await scheduler.run();

    expect(events.slice(0, 2)).toEqual(["send:start", "recv:start"]);
    expect(events.slice(2).sort()).toEqual(["recv:7:true", "send:done"]);
  });

  test("preserves buffered channel FIFO order and wakes blocked senders", async () => {
    const scheduler = new AsyncGoScheduler();
    const ch = new AsyncGoChannel<number>(scheduler, 2, zeroNumber);
    const values: number[] = [];

    scheduler.go(async () => {
      await ch.send(1);
      await ch.send(2);
      await ch.send(3);
      values.push(99);
    });
    scheduler.go(async () => {
      values.push((await ch.receive())[0]);
      values.push((await ch.receive())[0]);
      values.push((await ch.receive())[0]);
    });

    await scheduler.run();

    expect(values).toEqual([1, 2, 99, 3]);
    expect(ch.len()).toBe(0);
    expect(ch.cap()).toBe(2);
  });

  test("drains buffered values after close and then returns zero false", async () => {
    const scheduler = new AsyncGoScheduler();
    const ch = new AsyncGoChannel<number>(scheduler, 2, zeroNumber);

    await ch.send(4);
    await ch.send(5);
    ch.close();

    expect(await ch.receive()).toEqual([4, true]);
    expect(await ch.receive()).toEqual([5, true]);
    expect(await ch.receive()).toEqual([0, false]);
  });

  test("panics on send to closed channel and double close", async () => {
    const scheduler = new AsyncGoScheduler();
    const ch = new AsyncGoChannel<number>(scheduler, 1, zeroNumber);
    ch.close();

    let sendError: unknown;
    try {
      await ch.send(1);
    } catch (error) {
      sendError = error;
    }
    expect(sendError).toBeInstanceOf(AsyncGoPanic);

    let closeError: unknown;
    try {
      ch.close();
    } catch (error) {
      closeError = error;
    }
    expect(closeError).toBeInstanceOf(AsyncGoPanic);
  });

  test("select chooses among ready cases through the deterministic scheduler PRNG", async () => {
    const first = new AsyncGoScheduler({ randomSeed: "gojr-select-seed" });
    const ch1 = new AsyncGoChannel<number>(first, 1, zeroNumber);
    const ch2 = new AsyncGoChannel<number>(first, 1, zeroNumber);
    await ch1.send(1);
    await ch2.send(2);

    const selected = await asyncSelect(first, [
      { op: "receive", channel: ch1 },
      { op: "receive", channel: ch2 }
    ]);
    expect(selected).toEqual({ index: 1, op: "receive", value: 2, ok: true });

    const second = new AsyncGoScheduler({ randomSeed: "gojr-other-select-seed" });
    const other1 = new AsyncGoChannel<number>(second, 1, zeroNumber);
    const other2 = new AsyncGoChannel<number>(second, 1, zeroNumber);
    await other1.send(1);
    await other2.send(2);

    const otherSelected = await asyncSelect(second, [
      { op: "receive", channel: other1 },
      { op: "receive", channel: other2 }
    ]);
    expect(otherSelected).toEqual({ index: 0, op: "receive", value: 1, ok: true });
  });

  test("select disables nil channel cases and runs default immediately", async () => {
    const scheduler = new AsyncGoScheduler();
    const selected = await asyncSelect<number>(scheduler, [
      { op: "receive", channel: null },
      { op: "send", channel: undefined, value: 1 },
      { op: "default" }
    ]);

    expect(selected).toEqual({ index: 2, op: "default" });
  });

  test("detects channel deadlock when all goroutines are blocked", async () => {
    const scheduler = new AsyncGoScheduler();
    const ch = new AsyncGoChannel<number>(scheduler, 0, zeroNumber);

    scheduler.go(async () => {
      await ch.receive();
    });

    let error: unknown;
    try {
      await scheduler.run();
    } catch (caught) {
      error = caught;
    }

    expect(error).toBeInstanceOf(AsyncGoDeadlockError);
  });
});
