import { describe, expect, test } from "./testHarness.js";
import { analyzeEffects, parseProgram } from "../src/index.js";
import type { ProgramAst } from "../src/index.js";

function parse(source: string): ProgramAst {
  const result = parseProgram(source, "effects-test.go");
  expect(result.diagnostics).toEqual([]);
  if (!result.ast) throw new Error("missing AST");
  return result.ast;
}

describe("Go-junior effect analysis", () => {
  test("marks ordinary functions and top-level code as direct", () => {
    const ast = parse(`
func add(a, b int) int {
  return a + b
}

func twice(x int) int {
  return add(x, x)
}

x := twice(3)
return x
`);

    const effects = analyzeEffects(ast);

    expect(effects.functions.add?.direct).toBe(true);
    expect(effects.functions.twice?.direct).toBe(true);
    expect(effects.functions.twice?.calls).toEqual(["add"]);
    expect(effects.topLevel.direct).toBe(true);
    expect(effects.topLevel.calls).toEqual(["twice"]);
  });

  test("propagates may-suspend effects through known function calls", () => {
    const ast = parse(`
func recv(ch chan int) int {
  return <-ch
}

func wrap(ch chan int) int {
  return recv(ch)
}

func top(ch chan int) int {
  return wrap(ch)
}
`);

    const effects = analyzeEffects(ast);

    expect(effects.functions.recv?.maySuspend).toBe(true);
    expect(effects.functions.recv?.reasons).toEqual([{ kind: "channel-receive" }]);
    expect(effects.functions.wrap?.maySuspend).toBe(true);
    expect(effects.functions.wrap?.reasons).toEqual([{ kind: "call", target: "recv" }]);
    expect(effects.functions.top?.maySuspend).toBe(true);
    expect(effects.functions.top?.reasons).toEqual([{ kind: "call", target: "wrap" }]);
  });

  test("detects go statements, channel sends, select statements, and immediate function literals", () => {
    const ast = parse(`
func worker(ch chan int) {
  ch <- 1
}

func launch(ch chan int) {
  go worker(ch)
}

func choose(ch chan int) int {
  return func() int {
    select {
    case v := <-ch:
      return v
    default:
      return 0
    }
  }()
}
`);

    const effects = analyzeEffects(ast);

    expect(effects.functions.worker?.reasons).toEqual([{ kind: "channel-send" }]);
    expect(effects.functions.launch?.reasons).toEqual([{ kind: "go" }, { kind: "call", target: "worker" }]);
    expect(effects.functions.choose?.maySuspend).toBe(true);
    expect(effects.functions.choose?.reasons).toEqual([
      { kind: "select" },
      { kind: "channel-receive" }
    ]);
  });
});
