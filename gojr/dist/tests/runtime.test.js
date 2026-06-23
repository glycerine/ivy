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
    test("keeps function defer stacks scoped and last-in-first-out", () => {
        const session = new GoJuniorSession();
        expect(session.evaluate(`import "fmt"`).diagnostics).toEqual([]);
        expect(session.evaluate(`
func inner() {
  defer fmt.Printf("inner defer\\n")
  fmt.Printf("inner body\\n")
}
`).diagnostics).toEqual([]);
        expect(session.evaluate(`
func outer() {
  defer fmt.Printf("outer first\\n")
  defer fmt.Printf("outer second\\n")
  inner()
  fmt.Printf("after inner\\n")
}
`).diagnostics).toEqual([]);
        const result = session.evaluate("outer()");
        expect(result.diagnostics).toEqual([]);
        expect(result.output).toEqual([
            "inner body\n",
            "inner defer\n",
            "after inner\n",
            "outer second\n",
            "outer first\n"
        ]);
    });
    test("runs deferred closures before reading named return values", () => {
        const session = new GoJuniorSession();
        const define = session.evaluate(`
func f() (x int) {
  defer func() {
    x = 2
  }()
  x = 1
  return
}
`);
        expect(define.diagnostics).toEqual([]);
        const result = session.evaluate("f()");
        expect(result.diagnostics).toEqual([]);
        expect(result.value).toBe(2n);
    });
    test("runs function defers while panicking", () => {
        const session = new GoJuniorSession();
        expect(session.evaluate(`import "fmt"`).diagnostics).toEqual([]);
        expect(session.evaluate(`
func boom() {
  defer fmt.Printf("cleanup\\n")
  panic("bad")
}
`).diagnostics).toEqual([]);
        const result = session.evaluate("boom()");
        expect(result.output).toEqual(["cleanup\n"]);
        expect(result.diagnostics).toHaveLength(1);
        expect(result.diagnostics[0]?.code).toBe("GJPANIC001");
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
    test("automatically imports fmt in scripts and REPL sessions", () => {
        const script = expectRuns(`
fmt.Printf("auto %v\\n", 7)
return fmt.Sprintf("ok %v", 8)
`);
        expect(script.output).toEqual(["auto 7\n"]);
        expect(script.value).toBe("ok 8");
        const session = new GoJuniorSession();
        const define = session.evaluate(`f := func() { fmt.Printf("hiya\\n") }`);
        expect(define.diagnostics).toEqual([]);
        const call = session.evaluate("f()");
        expect(call.diagnostics).toEqual([]);
        expect(call.output).toEqual(["hiya\n"]);
        expect(call.value).toBeNull();
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
    test("enforces declared and inferred local variable types", () => {
        const declared = evaluateSource(`
var x int
x = "bad"
`);
        expect(declared.diagnostics).toHaveLength(1);
        expect(declared.diagnostics[0]?.message).toContain("variable x bad is not assignable to int");
        const inferred = evaluateSource(`
x := 1
x = "bad"
`);
        expect(inferred.diagnostics).toHaveLength(1);
        expect(inferred.diagnostics[0]?.message).toContain("variable x bad is not assignable to int64");
    });
    test("supports Go-style const groups with iota and repeated expressions", () => {
        const result = expectRuns(`
const Single = iota
const (
  A = iota
  B
  C int = iota
  D
)
return Single, A, B, C, D
`);
        expect(result.values).toEqual([0n, 0n, 1n, 2n, 3n]);
        const immutable = evaluateSource(`
const X = 1
X = 2
`);
        expect(immutable.diagnostics).toHaveLength(1);
        expect(immutable.diagnostics[0]?.message).toContain("X is const");
    });
    test("enforces function parameter and typed collection assignments", () => {
        const badParam = evaluateSource(`
func f(x int) int { return x }
return f("bad")
`);
        expect(badParam.diagnostics).toHaveLength(1);
        expect(badParam.diagnostics[0]?.message).toContain("variable x bad is not assignable to int");
        const badArray = evaluateSource(`
var xs []int
xs = []string{"bad"}
`);
        expect(badArray.diagnostics).toHaveLength(1);
        expect(badArray.diagnostics[0]?.message).toContain("variable xs element bad is not assignable to int");
        const badMap = evaluateSource(`
var m map[string]int
m = map[int]string{1: "bad"}
`);
        expect(badMap.diagnostics).toHaveLength(1);
        expect(badMap.diagnostics[0]?.message).toContain("variable m");
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
    test("evaluates struct literals, zero values, field mutation, and fmt verbs", () => {
        const result = expectRuns(`
type Point struct {
  X, Y int
  Name string
}

a := Point{X: 1, Y: 2, Name: "home"}
b := Point{3, 4, "away"}
c := Point{}
a.X = a.X + b.Y
return a.X, c.Y, fmt.Sprintf("%v | %#v", a, a)
`);
        expect(result.values).toEqual([
            5n,
            0n,
            `Point{X:5 Y:2 Name:home} | Point{X: int64(5), Y: int64(2), Name: string("home")}`
        ]);
    });
    test("supports value and pointer receiver methods with Go selector syntax", () => {
        const session = new GoJuniorSession();
        expect(session.evaluate(`
type Point struct {
  X, Y int
}
`).diagnostics).toEqual([]);
        const valueMethod = session.evaluate(`
func (p Point) Sum() int {
  return p.X + p.Y
}
`);
        expect(valueMethod.diagnostics).toEqual([]);
        const pointerMethod = session.evaluate(`
func (p *Point) Scale(k int) {
  p.X = p.X * k
  p.Y = p.Y * k
}
`);
        expect(pointerMethod.diagnostics).toEqual([]);
        const call = session.evaluate(`
p := Point{2, 3}
before := p.Sum()
p.Scale(4)
return before, p.X, p.Y, p.Sum()
`);
        expect(call.diagnostics).toEqual([]);
        expect(call.values).toEqual([5n, 8n, 12n, 20n]);
    });
    test("supports address-of and dereference assignment for structs and fields", () => {
        const result = expectRuns(`
type Box struct {
  X int
}

b := Box{X: 1}
p := &b
(*p).X = 9
q := &b.X
*q = *q + 1
return b.X, (*p).X
`);
        expect(result.values).toEqual([10n, 10n]);
    });
    test("executes Go type switches over structs, pointers, nil, and basic values", () => {
        const result = expectRuns(`
type Point struct {
  X int
}

describe := func(x interface{}) string {
  switch v := x.(type) {
  case nil:
    return "nil"
  case *Point:
    return fmt.Sprintf("ptr:%v", v.X)
  case Point:
    return fmt.Sprintf("point:%v", v.X)
  case int:
    return fmt.Sprintf("int:%v", v)
  default:
    return "other"
  }
}

p := Point{X: 7}
return describe(p), describe(&p), describe(nil), describe(3), describe("x")
`);
        expect(result.values).toEqual(["point:7", "ptr:7", "nil", "int:3", "other"]);
    });
    test("executes unbound type switches and rejects fallthrough", () => {
        const ok = expectRuns(`
x := "hello"
out := ""
switch x.(type) {
case string:
  out = "string"
default:
  out = "other"
}
return out
`);
        expect(ok.value).toBe("string");
        const bad = evaluateSource(`
x := 1
switch x.(type) {
case int:
  fallthrough
default:
  return "bad"
}
`);
        expect(bad.diagnostics).toHaveLength(1);
        expect(bad.diagnostics[0]?.message).toContain("fallthrough is not allowed in type switches");
    });
    test("evaluates Go type assertions and reports mismatches", () => {
        const ok = expectRuns(`
type Point struct {
  X int
}

asPoint := func(x interface{}) int {
  return x.(Point).X
}

p := Point{X: 11}
ptr := &p
return asPoint(p), ptr.(*Point).X
`);
        expect(ok.values).toEqual([11n, 11n]);
        const bad = evaluateSource(`
x := "hello"
return x.(int)
`);
        expect(bad.diagnostics).toHaveLength(1);
        expect(bad.diagnostics[0]?.message).toContain("does not have dynamic type int");
    });
    test("enforces named interface method sets on typed variables", () => {
        const ok = expectRuns(`
type Stringer interface {
  String() string
}

type Point struct {
  X int
}

func (p Point) String() string {
  return fmt.Sprintf("Point(%v)", p.X)
}

var s Stringer
s = Point{X: 5}
return s.String()
`);
        expect(ok.value).toBe("Point(5)");
        const pointerOnly = evaluateSource(`
type Mutator interface {
  Mutate()
}

type Box struct { X int }
func (b *Box) Mutate() { b.X++ }

var m Mutator
b := Box{X: 1}
m = b
`);
        expect(pointerOnly.diagnostics).toHaveLength(1);
        expect(pointerOnly.diagnostics[0]?.message).toContain("variable m");
        const pointerOk = expectRuns(`
type Mutator interface {
  Mutate()
}

type Box struct { X int }
func (b *Box) Mutate() { b.X++ }

var m Mutator
b := Box{X: 1}
m = &b
m.Mutate()
return b.X
`);
        expect(pointerOk.value).toBe(2n);
    });
    test("evaluates array and slice literals with indexing, slicing, and range", () => {
        const result = expectRuns(`
xs := []int{1, 2, 3}
ys := [4]int{4, 5}
zs := [...]string{"a", "b", "c"}
sum := 0
for _, v := range xs {
  sum = sum + v
}
return xs[1], xs[1:3], ys, len(zs), sum
`);
        expect(result.values).toEqual([
            2n,
            [2n, 3n],
            [4n, 5n, 0n, 0n],
            3n,
            6n
        ]);
    });
    test("supports len, cap, append, and default slice/array declarations", () => {
        const result = expectRuns(`
var xs []int
var ys [3]string
xs = append(xs, 1, 2)
more := []int{3, 4}
xs = append(xs, more...)
return xs, len(xs), cap(xs), len(ys), ys[0]
`);
        expect(result.values).toEqual([[1n, 2n, 3n, 4n], 4n, 4n, 3n, ""]);
    });
    test("reports typed array and slice literal element mismatches", () => {
        const result = evaluateSource(`
xs := []int{1, "bad"}
`);
        expect(result.diagnostics).toHaveLength(1);
        expect(result.diagnostics[0]?.message).toContain("array element bad is not assignable to int");
    });
    test("supports len on strings and maps", () => {
        const result = expectRuns(`
m := map[string]int{"a": 1, "b": 2}
return len("hiya"), len(m)
`);
        expect(result.values).toEqual([4n, 2n]);
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
    test("defines and calls functions with grouped names in REPL sessions", () => {
        const session = new GoJuniorSession();
        const define = session.evaluate("func f(a, b, c int) (x, y, z int) { return a, b, c }");
        expect(define.diagnostics).toEqual([]);
        const call = session.evaluate("f(1, 2, 3)");
        expect(call.diagnostics).toEqual([]);
        expect(call.value).toEqual([1n, 2n, 3n]);
    });
    test("supports named result variables and naked returns in REPL functions", () => {
        const session = new GoJuniorSession();
        const define = session.evaluate(`
func swap(a, b int) (left, right int) {
  left = b
  right = a
  return
}
`);
        expect(define.diagnostics).toEqual([]);
        const call = session.evaluate("swap(1, 2)");
        expect(call.diagnostics).toEqual([]);
        expect(call.value).toEqual([2n, 1n]);
    });
    test("supports variadic parameters and spread calls in REPL functions", () => {
        const session = new GoJuniorSession({
            sheet: {
                A1: [4n, 5n, 6n]
            }
        });
        const define = session.evaluate(`
func sum(vals ...int) int {
  total := 0
  for _, v := range vals {
    total = total + v
  }
  return total
}
`);
        expect(define.diagnostics).toEqual([]);
        const direct = session.evaluate("sum(1, 2, 3)");
        expect(direct.diagnostics).toEqual([]);
        expect(direct.value).toBe(6n);
        const spread = session.evaluate("sum(sheet.A1...)");
        expect(spread.diagnostics).toEqual([]);
        expect(spread.value).toBe(15n);
    });
    test("supports multiple short declarations and assignments", () => {
        const result = expectRuns(`
a, b := 1, 2
a, b = b, a
return a, b
`);
        expect(result.values).toEqual([2n, 1n]);
    });
    test("supports destructuring multiple function returns", () => {
        const result = expectRuns(`
func divmod(x, y int) (int, int) {
  return x / y, x % y
}

q, r := divmod(17, 5)
_, onlyR := divmod(19, 5)
return q, r, onlyR
`);
        expect(result.values).toEqual([3n, 2n, 4n]);
    });
    test("supports multi-assignment to index targets after evaluating rhs", () => {
        const result = expectRuns(`
xs := []int{1, 2}
xs[0], xs[1] = xs[1], xs[0]
return xs
`);
        expect(result.value).toEqual([2n, 1n]);
    });
    test("defines and calls function literals with grouped names in REPL sessions", () => {
        const session = new GoJuniorSession();
        const define = session.evaluate("f := func(a, b, c int) (d, e, f int) { return b, c, a }");
        expect(define.diagnostics).toEqual([]);
        const call = session.evaluate("f(1, 2, 3)");
        expect(call.diagnostics).toEqual([]);
        expect(call.value).toEqual([2n, 3n, 1n]);
    });
    test("closures capture lexical variables by reference", () => {
        const session = new GoJuniorSession();
        expect(session.evaluate("base := 10").diagnostics).toEqual([]);
        expect(session.evaluate("addBase := func(x int) int { return base + x }").diagnostics).toEqual([]);
        expect(session.evaluate("base = 20").diagnostics).toEqual([]);
        const call = session.evaluate("addBase(2)");
        expect(call.diagnostics).toEqual([]);
        expect(call.value).toBe(22n);
    });
    test("returned closures keep their defining function scope alive", () => {
        const session = new GoJuniorSession();
        const define = session.evaluate(`
func makeAdder(base int) func(int) int {
  return func(x int) int {
    return base + x
  }
}
`);
        expect(define.diagnostics).toEqual([]);
        expect(session.evaluate("add5 := makeAdder(5)").diagnostics).toEqual([]);
        const call = session.evaluate("add5(3)");
        expect(call.diagnostics).toEqual([]);
        expect(call.value).toBe(8n);
    });
    test("function literals support variadic parameters and spread calls", () => {
        const session = new GoJuniorSession({
            sheet: {
                A1: [7n, 8n, 9n]
            }
        });
        const define = session.evaluate(`
sum := func(vals ...int) int {
  total := 0
  for _, v := range vals {
    total = total + v
  }
  return total
}
`);
        expect(define.diagnostics).toEqual([]);
        const direct = session.evaluate("sum(1, 2, 3)");
        expect(direct.diagnostics).toEqual([]);
        expect(direct.value).toBe(6n);
        const spread = session.evaluate("sum(sheet.A1...)");
        expect(spread.diagnostics).toEqual([]);
        expect(spread.value).toBe(24n);
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
    test("keeps bare labels pending until their statement is entered", () => {
        const session = new GoJuniorSession();
        const result = session.evaluate("top:");
        expect(result.incomplete).toBe(true);
        expect(result.diagnostics.length).toBeGreaterThan(0);
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
    test("supports REPL entry of labeled nested loops with labeled break", () => {
        const session = new GoJuniorSession();
        const lines = [
            "top:",
            "for i := 0; i < 5; i++ {",
            "  inner:",
            "  for j := 0; j < 10; j++ {",
            "    fmt.Printf(\"i=%v j=%v\\n\", i, j)",
            "    break top",
            "  }",
            "}",
        ];
        let source = "";
        for (const [index, line] of lines.entries()) {
            source += `${line}\n`;
            const result = session.evaluate(source);
            if (index < lines.length - 1) {
                expect(result.incomplete).toBe(true);
            }
            else {
                expect(result.diagnostics).toEqual([]);
                expect(result.output).toEqual(["i=0 j=0\n"]);
            }
        }
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
