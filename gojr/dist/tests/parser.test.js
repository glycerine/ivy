import { describe, expect, test } from "vitest";
import { parseProgram } from "../src/index.js";
function expectParses(source) {
    const result = parseProgram(source);
    expect(result.diagnostics).toEqual([]);
    expect(result.ast).toBeDefined();
    return result;
}
describe("Go-junior parser", () => {
    test("parses cell imports with aliases and fmt diagnostics", () => {
        const result = expectParses(`
import (
  f "fmt"
  stats "workbook/stats"
)

f.Printf("mean starts at %v\\n", sheet.A1)
return stats.Mean(sheet.A1:A10)
`);
        expect(result.ast?.imports).toEqual([
            { alias: "f", path: "fmt" },
            { alias: "stats", path: "workbook/stats" }
        ]);
    });
    test("parses pointer receiver methods with Go selector syntax", () => {
        expectParses(`
func (p *Point) Scale(k float64) {
  p.X = p.X * k
  p.Y = p.Y * k
}
`);
    });
    test("rejects C and C++ style pointer selector syntax", () => {
        const result = parseProgram(`
p->Scale(2.0)
`);
        expect(result.diagnostics.length).toBeGreaterThan(0);
    });
    test("parses panic and panicOn calls", () => {
        expectParses(`
panicOn(err)
panic("boom")
`);
    });
    test("parses Go-junior function cell source", () => {
        const result = expectParses(`
import "fmt"

func Double(x float64) float64 {
  fmt.Printf("double %v\\n", x)
  return x * 2
}
`);
        expect(result.ast?.kind).toBe("function");
        expect(result.ast?.imports).toEqual([{ path: "fmt" }]);
    });
});
