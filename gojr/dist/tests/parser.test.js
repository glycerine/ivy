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
    test("parses map types and literals", () => {
        expectParses(`
var m map[string]int
m = map[string]int{"a": 1, "b": 2}
`);
    });
    test("parses grouped parameter names and variadic parameters", () => {
        const grouped = expectParses(`
func F(a, b, c int) (x, y, z int) {
  return a, b, c
}
`);
        expect(grouped.ast?.functions[0]?.signature.parameters.map((param) => param.name)).toEqual(["a", "b", "c"]);
        expect(grouped.ast?.functions[0]?.signature.parameters.map((param) => param.type.text)).toEqual(["int", "int", "int"]);
        expect(grouped.ast?.functions[0]?.signature.results.map((param) => param.name)).toEqual(["x", "y", "z"]);
        expect(grouped.ast?.functions[0]?.signature.results.map((param) => param.type.text)).toEqual(["int", "int", "int"]);
        const variadic = expectParses(`
func Sum(prefix string, vals ...int) int {
  return 0
}
`);
        expect(variadic.ast?.functions[0]?.signature.parameters.map((param) => param.name)).toEqual(["prefix", "vals"]);
        expect(variadic.ast?.functions[0]?.signature.parameters.map((param) => param.variadic)).toEqual([false, true]);
    });
    test("parses function literals as expressions", () => {
        const result = expectParses(`
f := func(a, b, c int) (d, e, f int) { return b, c, a }
`);
        expect(result.ast?.body[0]?.kind).toBe("ShortVarStatement");
    });
    test("keeps comma-separated unnamed type parameters distinct", () => {
        const result = expectParses(`
func F(int, string, bool) {
}
`);
        expect(result.ast?.functions[0]?.signature.parameters).toEqual([
            { type: { text: "int", span: expect.any(Object) }, variadic: false },
            { type: { text: "string", span: expect.any(Object) }, variadic: false },
            { type: { text: "bool", span: expect.any(Object) }, variadic: false }
        ]);
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
