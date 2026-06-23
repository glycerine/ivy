import { describe, expect, test } from "vitest";
import { evaluateSource, GoJuniorSession } from "../src/index.js";

function expectRuns(source: string, options = {}) {
  const result = evaluateSource(source, options);
  expect(result.diagnostics).toEqual([]);
  return result;
}

describe("Go-junior runtime slice", () => {
  test("evaluates sheet arithmetic without number casts", () => {
    const result = expectRuns(`
var x = sheet.A1 + sheet.B1 * 2
if x > 10 {
  return x
}
return x * 2
`, {
      sheet: {
        A1: 4n,
        B1: 3n
      }
    });

    expect(result.value).toBe(20n);
    expect(result.values).toEqual([20n]);
  });

  test("resolves absolute refs, cross-sheet namespaces, and fmt aliases", () => {
    const result = expectRuns(`
import f "fmt"

f.Printf("A1=%v\\n", sheet.$A$1)
return Budget.B2 + sheet.$A$1
`, {
      sheet: {
        A1: 2n
      },
      sheets: {
        Budget: {
          B2: 5n
        }
      }
    });

    expect(result.output).toEqual(["A1=2\n"]);
    expect(result.value).toBe(7n);
  });

  test("evaluates spreadsheet ranges as row-major two-dimensional arrays", () => {
    const result = expectRuns(`
return sheet.A1:B2
`, {
      sheet: {
        A1: 1n,
        B1: 2n,
        A2: 3n,
        B2: 4n
      }
    });

    expect(result.value).toEqual([
      [1n, 2n],
      [3n, 4n]
    ]);
  });

  test("runs defers after the script body in reverse order", () => {
    const result = expectRuns(`
import "fmt"

defer fmt.Printf("third")
defer fmt.Printf(" second ")
fmt.Printf("first")
`);

    expect(result.output).toEqual(["first", " second ", "third"]);
  });

  test("formats Go-junior values with fmt %#v", () => {
    const result = expectRuns(`
import "fmt"

fmt.Printf("i = %#v\\n", 7)
return fmt.Sprintf("%#v %#v %#v", "x", 1.5, true)
`);

    expect(result.output).toEqual(["i = int64(7)\n"]);
    expect(result.value).toBe(`string("x") float64(1.5) bool(true)`);
  });

  test("supports declared maps with typed string and integer keys", () => {
    const stringKeyed = expectRuns(`
var m map[string]int
m["hi"] = 3
return m["hi"]
`);
    expect(stringKeyed.value).toBe(3n);

    const intKeyed = expectRuns(`
var mm map[int]string
mm[3] = "hi"
return mm[3]
`);
    expect(intKeyed.value).toBe("hi");
  });

  test("reports map key and value type mismatches without numeric-index coercion", () => {
    const badKey = evaluateSource(`
var mm map[int]string
mm["hi"] = 3
`);
    expect(badKey.diagnostics).toHaveLength(1);
    expect(badKey.diagnostics[0]?.message).toContain("map key hi is not assignable to int");
    expect(badKey.diagnostics[0]?.message).not.toContain("numeric");

    const badValue = evaluateSource(`
var m map[string]int
m["hi"] = "three"
`);
    expect(badValue.diagnostics).toHaveLength(1);
    expect(badValue.diagnostics[0]?.message).toContain("map value three is not assignable to int");
  });

  test("evaluates map literals, missing-key zero values, and insertion-order range", () => {
    const result = expectRuns(`
counts := map[string]int64{"a": 1, "b": 2}
out := ""
for k, v := range counts {
  out = out + k + ":" + v + ";"
}
return counts["a"] + counts["b"] + counts["missing"], out
`);

    expect(result.values).toEqual([3n, "a:1;b:2;"]);
  });

  test("formats typed maps with fmt verbs", () => {
    const result = expectRuns(`
import "fmt"

counts := map[string]int64{"a": 1, "b": 2}
return fmt.Sprintf("%v | %#v", counts, counts)
`);

    expect(result.value).toBe(`map[a:1 b:2] | map[string]int64{string("a"): int64(1), string("b"): int64(2)}`);
  });

  test("reports panicOn failures as runtime diagnostics", () => {
    const result = evaluateSource(`
panicOn("bad")
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GJPANIC001");
  });

  test("keeps REPL session locals across eager evaluations", () => {
    const session = new GoJuniorSession();

    expect(session.evaluate("a := 10").diagnostics).toEqual([]);
    const result = session.evaluate("a + 2");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(12n);
  });

  test("supports Go-style increment and decrement statements in sessions", () => {
    const session = new GoJuniorSession();

    expect(session.evaluate("a := 10").diagnostics).toEqual([]);
    expect(session.evaluate("a++").diagnostics).toEqual([]);
    expect(session.evaluate("a").value).toBe(11n);
    expect(session.evaluate("a--").diagnostics).toEqual([]);
    expect(session.evaluate("a").value).toBe(10n);
  });

  test("accepts top-level semicolons in REPL session input", () => {
    const session = new GoJuniorSession();

    expect(session.evaluate("a := 10;").diagnostics).toEqual([]);
    const result = session.evaluate("a");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(10n);
  });

  test("accepts import-only REPL input and does not replay output", () => {
    const session = new GoJuniorSession();

    expect(session.evaluate(`import "fmt"`).diagnostics).toEqual([]);
    expect(session.evaluate(`fmt.Printf("hi")`).output).toEqual(["hi"]);
    expect(session.evaluate("a := 1").output).toEqual([]);
  });

  test("marks incomplete REPL input without executing it", () => {
    const session = new GoJuniorSession();
    const incomplete = session.evaluate("if true {");

    expect(incomplete.incomplete).toBe(true);
    expect(incomplete.diagnostics.length).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.span?.line).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.span?.column).toBeGreaterThan(0);

    const complete = session.evaluate(`
if true {
  return 1
}
`);

    expect(complete.diagnostics).toEqual([]);
    expect(complete.value).toBe(1n);
  });

  test("keeps for blocks with increment statements pending until closed", () => {
    const session = new GoJuniorSession();
    const result = session.evaluate("for {\n  a++");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.span?.line).not.toBeNaN();
    expect(result.diagnostics[0]?.span?.column).not.toBeNaN();
  });

  test("keeps Go-style for clauses pending until the block is closed", () => {
    const session = new GoJuniorSession();
    const result = session.evaluate("for b := 0; b < 10; b++ {");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.message).not.toContain("':='");
  });

  test("executes Go-style for init condition and post clauses", () => {
    const result = expectRuns(`
sum := 0
for b := 0; b < 10; b++ {
  sum = sum + b
}
return sum
`);

    expect(result.value).toBe(45n);
  });

  test("runs for post clause after continue", () => {
    const result = expectRuns(`
sum := 0
for b := 0; b < 5; b++ {
  if b == 2 {
    continue
  }
  sum = sum + b
}
return sum
`);

    expect(result.value).toBe(8n);
  });

  test("supports goto to forward and backward labels", () => {
    const forward = expectRuns(`
i := 0
goto Done
i = 99
Done:
return i
`);
    expect(forward.value).toBe(0n);

    const backward = expectRuns(`
i := 0
Loop:
i++
if i < 3 {
  goto Loop
}
return i
`);
    expect(backward.value).toBe(3n);
  });

  test("supports labeled break out of nested loops", () => {
    const result = expectRuns(`
count := 0
Outer:
for i := 0; i < 3; i++ {
  for j := 0; j < 3; j++ {
    count = count + 1
    break Outer
  }
}
return count
`);

    expect(result.value).toBe(1n);
  });

  test("supports labeled continue from inside a switch to an outer for", () => {
    const result = expectRuns(`
sum := 0
Outer:
for i := 0; i < 4; i++ {
  switch i {
  case 2:
    continue Outer
  }
  sum = sum + i
}
return sum
`);

    expect(result.value).toBe(4n);
  });

  test("keeps unlabeled break scoped to switch inside for", () => {
    const result = expectRuns(`
sum := 0
for i := 0; i < 3; i++ {
  switch i {
  case 1:
    break
  }
  sum = sum + 1
}
return sum
`);

    expect(result.value).toBe(3n);
  });

  test("lets unlabeled continue inside switch continue the containing for", () => {
    const result = expectRuns(`
sum := 0
for i := 0; i < 3; i++ {
  switch i {
  case 1:
    continue
  }
  sum = sum + 1
}
return sum
`);

    expect(result.value).toBe(2n);
  });

  test("reports unresolved goto labels", () => {
    const result = evaluateSource(`
goto Missing
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("unresolved goto label Missing");
  });
});
