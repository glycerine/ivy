import { describe, expect, test } from "vitest";
import { evaluateSource, GoJuniorSession } from "../src/index.js";
function expectRuns(source, options = {}) {
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
});
