import { describe, expect, test } from "./testHarness.js";
import { emitAsyncJavaScript, parseProgram } from "../src/index.js";
import type { ProgramAst } from "../src/index.js";
import { AsyncGoChannel, AsyncGoDeadlockError, AsyncGoScheduler, asyncSelect } from "../src/asyncRuntime.js";

interface GeneratedModule {
  main: () => Promise<unknown>;
  run: () => Promise<unknown>;
}

function parse(source: string): ProgramAst {
  const result = parseProgram(source, "async-emitter-test.go");
  expect(result.diagnostics).toEqual([]);
  if (!result.ast) throw new Error("missing AST");
  return result.ast;
}

async function runGenerated(source: string, runtime?: unknown): Promise<unknown> {
  const module = generatedModule(source, runtime);
  return module.main();
}

async function runGeneratedProgram(source: string, runtime?: unknown): Promise<unknown> {
  const module = generatedModule(source, runtime);
  return module.run();
}

function generatedModule(source: string, runtime?: unknown): GeneratedModule {
  return new Function(source)(runtime) as GeneratedModule;
}

function runtime(seed?: string): unknown {
  const scheduler = new AsyncGoScheduler(seed ? { randomSeed: seed } : {});
  return { scheduler, AsyncGoChannel, asyncSelect };
}

describe("Go-junior async JavaScript emitter", () => {
  test("emits async functions and awaits Go-junior calls", async () => {
    const ast = parse(`
func add(a, b int) int {
  return a + b
}

func twice(x int) int {
  return add(x, x)
}

x := twice(21)
return x
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(emitted.effects.functions.add?.direct).toBe(true);
    expect(emitted.effects.functions.twice?.calls).toEqual(["add"]);
    expect(emitted.source).toContain("async function _fn_add");
    expect(emitted.source).toContain("async function _fn_twice");
    expect(emitted.source).toContain("await _fn_add");
    expect(await runGenerated(emitted.source)).toBe(42n);
  });

  test("emits top-level variables, assignment, if, and multiple return values", async () => {
    const ast = parse(`
func inc(x int) int {
  return x + 1
}

x := inc(2)
x = x + 3
if x == 6 {
  return x, "ok"
}
return 0, "bad"
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(await runGenerated(emitted.source)).toEqual([6n, "ok"]);
  });

  test("emits goroutine handoff over an unbuffered channel", async () => {
    const ast = parse(`
func send(ch chan int) {
  ch <- 9
}

ch := make(chan int)
go send(ch)
return <-ch
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(emitted.effects.functions.send?.maySuspend).toBe(true);
    expect(await runGeneratedProgram(emitted.source, runtime())).toBe(9n);
  });

  test("emits select default and receive cases", async () => {
    const ast = parse(`
ch := make(chan int, 1)
select {
case v := <-ch:
  return v
default:
}
ch <- 7
select {
case v := <-ch:
  return v
default:
  return 0
}
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(await runGeneratedProgram(emitted.source, runtime("select-smoke"))).toBe(7n);
  });

  test("emits two-value channel receives", async () => {
    const ast = parse(`
ch := make(chan int, 1)
ch <- 3
v, ok := <-ch
return v, ok
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(await runGeneratedProgram(emitted.source, runtime())).toEqual([3n, true]);
  });

  test("emits close and closed-channel zero receives", async () => {
    const ast = parse(`
ch := make(chan string)
close(ch)
v, ok := <-ch
return v, ok
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(await runGeneratedProgram(emitted.source, runtime())).toEqual(["", false]);
  });

  test("emits len and cap for channels", async () => {
    const ast = parse(`
ch := make(chan int, 2)
ch <- 1
return len(ch), cap(ch)
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(await runGeneratedProgram(emitted.source, runtime())).toEqual([1n, 2n]);
  });

  test("emits function literals used as goroutine roots", async () => {
    const ast = parse(`
ch := make(chan int)
go func() { ch <- 4 }()
return <-ch
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(await runGeneratedProgram(emitted.source, runtime())).toBe(4n);
  });

  test("generated run detects main goroutine channel deadlocks", async () => {
    const ast = parse(`
ch := make(chan int)
return <-ch
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    let thrown: unknown;
    try {
      await runGeneratedProgram(emitted.source, runtime());
    } catch (error) {
      thrown = error;
    }
    expect(thrown).toBeInstanceOf(AsyncGoDeadlockError);
  });

  test("emits channel receive expressions with effect metadata", async () => {
    const ast = parse(`
func recv(ch chan int) int {
  return <-ch
}
`);

    const emitted = emitAsyncJavaScript(ast);

    expect(emitted.diagnostics).toEqual([]);
    expect(emitted.source).toContain(".receive()");
    expect(emitted.effects.functions.recv?.maySuspend).toBe(true);
  });
});
