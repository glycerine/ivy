import { describe, expect, test } from "./testHarness.js";
import { evaluateSource, formatReplValue, GoJuniorSession } from "../src/index.js";

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

  test("supports Go two-value map lookups for key presence", () => {
    const missing = expectRuns(`
var m map[int]int
a, ok := m[3]
return a, ok
`);
    expect(missing.values).toEqual([0n, false]);

    const present = expectRuns(`
var m map[int]string
m[3] = "hi"
a, ok := m[3]
return a, ok
`);
    expect(present.values).toEqual(["hi", true]);
  });

  test("supports standard Go make for maps and slices", () => {
    const mapResult = expectRuns(`
m := make(map[int]int)
missing, missingOK := m[3]
m[3] = 9
present, presentOK := m[3]
return missing, missingOK, present, presentOK
`);
    expect(mapResult.values).toEqual([0n, false, 9n, true]);

    const sliceResult = expectRuns(`
slc := make([]int, 20, 50)
slc = append(slc, 7)
slc2 := make([]string, 3)
return len(slc), cap(slc), slc[0], slc[20], len(slc2), cap(slc2), slc2[0]
`);
    expect(sliceResult.values).toEqual([21n, 50n, 0n, 7n, 3n, 3n, ""]);
  });

  test("reports ordinary runtime failures with GoJr-prefixed diagnostic codes", () => {
    const result = evaluateSource(`
missingName
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(result.diagnostics[0]?.message).toContain("missingName is not declared");
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
var floats map[int]float64
empty := fmt.Sprintf("%v", floats)
floats[39] = 3.2
return fmt.Sprintf("%v | %#v | %v | %v", counts, counts, empty, floats)
`);

    expect(result.value).toBe(`map[string]int64{a:1 b:2} | map[string]int64{string("a"): int64(1), string("b"): int64(2)} | map[int]float64{} | map[int]float64{39:3.2}`);
  });

  test("formats REPL map string values as Go literals", () => {
    const result = expectRuns(`
var m map[int]string
m[3] = "hi"
m[5] = "there"
var empty map[int]string
var multi map[int]string
multi[9] = "hello\\nthere"
var quoted map[int]string
quoted[1] = "he said \\"hi\\""
quoted[2] = "tick \` and \\"quote\\""
return empty, m, multi, quoted
`);

    expect(result.values?.map(formatReplValue)).toEqual([
      `map[int]string{}`,
      `map[int]string{3:"hi", 5:"there"}`,
      "map[int]string{9:`hello\nthere`}",
      "map[int]string{1:`he said \"hi\"`, 2:`tick \\` and \"quote\"`}"
    ]);
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

  test("represents interfaces as typed runtime values with dynamic nil state", () => {
    const nilInterface = expectRuns(`
type I interface {
  fun()
}

var i I
return i, i == nil
`);

    expect(nilInterface.values?.map(formatReplValue)).toEqual(["I(nil)", "true"]);

    const typedNilPointer = expectRuns(`
type I interface {
  fun()
}

type Box struct { X int }
func (b *Box) fun() {}

var p *Box
var i I = p
q, ok := i.(*Box)
return i == nil, p == nil, q == nil, ok, i
`);

    expect(typedNilPointer.values?.slice(0, 4)).toEqual([false, true, true, true]);
    expect(formatReplValue(typedNilPointer.values?.[4] ?? null)).toBe("*Box(nil)");

    const empty = expectRuns(`
type I interface {
  fun()
}

var i I
var x interface{} = i
return x, x == nil
`);

    expect(empty.values?.map(formatReplValue)).toEqual(["interface{}(nil)", "true"]);
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

  test("evaluates Go raw string literals", () => {
    const result = expectRuns("a := `hi\nthere`\nreturn a");

    expect(result.value).toBe("hi\nthere");
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

  test("keeps incomplete raw strings pending in REPL input", () => {
    const session = new GoJuniorSession();
    const incomplete = session.evaluate("a := ` hi there");

    expect(incomplete.incomplete).toBe(true);
    expect(incomplete.diagnostics.length).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.message).toContain("unterminated raw string");
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
      } else {
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

  test("supports Go conversions, numeric literals, rune literals, imaginary literals, and complex builtins", () => {
    const result = expectRuns(`
a := int(0b1010)
b := int64(0x10)
c := rune('A')
d := byte(0o7)
z := complex(float64(a), 2.5)
return a, b, c, d, real(z), imag(z), 3i + 2i
`);

    expect(result.values).toEqual([10n, 16n, 65n, 7n, 10, 2.5, { real: 0, imag: 5 }]);
  });

  test("supports bitwise, shift, unary complement, and compound assignment operators", () => {
    const result = expectRuns(`
x := 0b1010
x |= 0b0101
y := x & 0b1100
y ^= 0b0010
y &^= 0b0100
y <<= 2
y >>= 1
s := "hi"
s += "!"
return x, y, ^0, s
`);

    expect(result.values).toEqual([15n, 20n, -1n, "hi!"]);
  });

  test("supports if init statements and Go-style builtins new delete clear copy min max print println", () => {
    const result = expectRuns(`
xs := []int{1, 2, 3}
ys := make([]int, 3)
n := copy(ys, xs)
m := map[string]int{"a": 1, "b": 2}
delete(m, "a")
before := len(m)
clear(m)
p := new(int)
*p = max(3, min(7, 4))
if v := *p; v == 4 {
  print("v=", v)
  println(" ok")
}
return n, ys[0], ys[2], before, len(m), *p
`);

    expect(result.output).toEqual(["v=", "4", " ok\n"]);
    expect(result.values).toEqual([3n, 1n, 3n, 1n, 0n, 4n]);
  });

  test("supports methods on named non-struct types and two-value type assertions", () => {
    const result = expectRuns(`
type Duration int

func (d Duration) Double() Duration {
  return d + d
}

var x interface{} = Duration(5)
v, ok := x.(Duration)
bad, badOK := x.(string)
return ok, v.Double(), bad, badOK
`);

    expect(result.values).toEqual([true, 10n, "", false]);
  });

  test("supports address-of composite literals and three-index slicing capacity", () => {
    const result = expectRuns(`
type Point struct { X int }
p := &Point{X: 7}
xs := make([]int, 5, 8)
ys := xs[1:3:4]
return p.X, len(ys), cap(ys)
`);

    expect(result.values).toEqual([7n, 2n, 3n]);
  });

  test("enforces comparable map keys and supports array and struct comparability", () => {
    const ok = expectRuns(`
type Point struct { X int; Y string }
a := [2]int{1, 2}
b := [2]int{1, 2}
p := Point{X: 1, Y: "a"}
q := Point{X: 1, Y: "a"}
return a == b, p == q
`);
    expect(ok.values).toEqual([true, true]);

    const bad = evaluateSource(`
m := map[[]int]int{}
_ = m
`);
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.message).toContain("map key type []int is not comparable");
  });

  test("supports blank imports, dot imports, init functions, and range over integers and iterator functions", () => {
    const result = expectRuns(`
import . "fmt"
import _ "fmt"

var total int

func init() {
  total = 2
}

for i := range 4 {
  total += i
}

iter := func(yield func(int, string) bool) {
  if !yield(10, "a") {
    return
  }
  yield(20, "b")
}

out := ""
for k, v := range iter {
  out += Sprintf("%v:%v;", k, v)
}

return total, out
`);

    expect(result.values).toEqual([8n, "10:a;20:b;"]);
  });

  test("supports embedded fields, promoted methods, interface embedding, and struct tags", () => {
    const result = expectRuns(`
type Inner struct {
  X int
}

func (i Inner) Double() int {
  return i.X * 2
}

type Doubler interface {
  Double() int
}

type NamedDoubler interface {
  Doubler
}

type Outer struct {
  Inner \`json:"inner"\`
  Name string \`json:"name"\`
}

o := Outer{Inner: Inner{X: 3}, Name: "n"}
o.X = 4
var d NamedDoubler
d = o
return o.X, o.Double(), d.Double()
`);

    expect(result.values).toEqual([4n, 8n, 8n]);
  });

  test("supports pointer receivers on named scalar types", () => {
    const result = expectRuns(`
type Counter int

func (c *Counter) Inc() {
  *c = *c + 1
}

var c Counter
c.Inc()
c.Inc()
return c
`);

    expect(result.value).toBe(2n);
  });

  test("supports switch init statements and short redeclarations", () => {
    const result = expectRuns(`
x := 1
x, y := 2, 3
out := 0
switch z := x + y; z {
case 5:
  out = z
default:
  out = 99
}
return x, y, out
`);

    expect(result.values).toEqual([2n, 3n, 5n]);
  });

  test("rejects invalid short declarations, for posts, fallthrough, and gotos over variables", () => {
    const noNew = evaluateSource(`
x := 1
x := 2
`);
    expect(noNew.diagnostics).toHaveLength(1);
    expect(noNew.diagnostics[0]?.message).toContain("short declaration has no new variables");

    const badPost = evaluateSource(`
for i := 0; i < 2; i := i + 1 {
}
`);
    expect(badPost.diagnostics).toHaveLength(1);
    expect(badPost.diagnostics[0]?.message).toContain("for post");

    const badFallthroughMiddle = evaluateSource(`
switch 1 {
case 1:
  fallthrough
  fmt.Printf("nope")
default:
}
`);
    expect(badFallthroughMiddle.diagnostics).toHaveLength(1);
    expect(badFallthroughMiddle.diagnostics[0]?.message).toContain("fallthrough must be the final statement");

    const badFallthroughFinal = evaluateSource(`
switch 1 {
case 1:
  fallthrough
}
`);
    expect(badFallthroughFinal.diagnostics).toHaveLength(1);
    expect(badFallthroughFinal.diagnostics[0]?.message).toContain("final switch clause");

    const badGoto = evaluateSource(`
goto Done
x := 1
Done:
return x
`);
    expect(badGoto.diagnostics).toHaveLength(1);
    expect(badGoto.diagnostics[0]?.message).toContain("jumps over variable declaration");
  });

  test("decodes Go string and rune escapes", () => {
    const result = expectRuns(`
s := "\\x41\\101\\u0042\\U00000043"
r := '\\n'
return s, r
`);

    expect(result.values).toEqual(["AABC", 10n]);
  });

  test("enforces recursive map-key comparability for arrays and structs", () => {
    const ok = expectRuns(`
type Key struct { A [2]int; B string }
m := map[Key]int{}
m[Key{A: [2]int{1, 2}, B: "x"}] = 7
return m[Key{A: [2]int{1, 2}, B: "x"}]
`);
    expect(ok.value).toBe(7n);

    const badStruct = evaluateSource(`
type Bad struct { A []int }
_ = map[Bad]int{}
`);
    expect(badStruct.diagnostics).toHaveLength(1);
    expect(badStruct.diagnostics[0]?.message).toContain("map key type []int is not comparable");
  });

  test("reports unresolved goto labels", () => {
    const result = evaluateSource(`
goto Missing
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("unresolved goto label Missing");
  });
});
