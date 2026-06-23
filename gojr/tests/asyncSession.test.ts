import { describe, expect, test } from "./testHarness.js";
import { AsyncGoJuniorSession } from "../src/index.js";

describe("Go-junior async REPL session", () => {
  test("keeps goroutine channel sends parked across REPL evaluations", async () => {
    const session = new AsyncGoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("go func() { c <- 1 }()")).diagnostics).toEqual([]);
    expect((await session.evaluate("a := <-c")).diagnostics).toEqual([]);

    const result = await session.evaluate("a");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("returns last expression values from persistent session state", async () => {
    const session = new AsyncGoJuniorSession();

    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    const result = await session.evaluate("a + 5");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(15n);
  });
});
