import { describe, expect, test } from "./testHarness.js";
import {
  cellDependency,
  evaluatePackageSourceFiles,
  evaluateSource,
  evaluateSourceFiles,
  formatGoNode,
  formatGoSource,
  formatReplValue,
  GoJuniorSession,
  parseProgram,
  rangeDependency,
  testSource,
  testSourceFiles
} from "../src/index.js";

async function expectRuns(source: string, options = {}) {
  const result = await evaluateSource(source, options);
  expect(result.diagnostics).toEqual([]);
  return result;
}

describe("Go-junior runtime slice", () => {
  test("evaluates sheet arithmetic without number casts", async () => {
    const result = await expectRuns(`
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

  test("resolves absolute refs, cross-sheet namespaces, and fmt aliases", async () => {
    const result = await expectRuns(`
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

  test("evaluates spreadsheet ranges as row-major two-dimensional arrays", async () => {
    const result = await expectRuns(`
rows := sheet.A1:B2
return rows[0][0] + rows[1][1], rows
`, {
      sheet: {
        A1: 1n,
        B1: 2n,
        A2: 3n,
        B2: 4n
      }
    });

    expect(result.values?.[0]).toBe(5n);
    expect(result.values?.[1]).toEqual([
      [1n, 2n],
      [3n, 4n]
    ]);
  });

  test("records observed spreadsheet cell and range dependencies", async () => {
    const result = await expectRuns(`
_ = sheet.$A$1
_ = Budget.B2
return sheet.A1:B2
`, {
      sheet: {
        A1: 1n,
        B1: 2n,
        A2: 3n,
        B2: 4n
      },
      sheets: {
        Budget: {
          B2: 5n
        }
      }
    });

    expect(result.observedDeps).toEqual([
      cellDependency({ sheet: "sheet", cell: "A1" }),
      cellDependency({ sheet: "Budget", cell: "B2" }),
      rangeDependency("sheet", "A1", "B2")
    ]);
  });

  test("runs defers after the script body in reverse order", async () => {
    const result = await expectRuns(`
import "fmt"

defer fmt.Printf("third")
defer fmt.Printf(" second ")
fmt.Printf("first")
`);

    expect(result.output).toEqual(["first", " second ", "third"]);
  });

  test("keeps function defer stacks scoped and last-in-first-out", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`import "fmt"`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`
func inner() {
  defer fmt.Printf("inner defer\\n")
  fmt.Printf("inner body\\n")
}
`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`
func outer() {
  defer fmt.Printf("outer first\\n")
  defer fmt.Printf("outer second\\n")
  inner()
  fmt.Printf("after inner\\n")
}
`)).diagnostics).toEqual([]);

    const result = await session.evaluate("outer()");
    expect(result.diagnostics).toEqual([]);
    expect(result.output).toEqual([
      "inner body\n",
      "inner defer\n",
      "after inner\n",
      "outer second\n",
      "outer first\n"
    ]);
  });

  test("runs deferred closures before reading named return values", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate(`
func f() (x int) {
  defer func() {
    x = 2
  }()
  x = 1
  return
}
`);
    expect(define.diagnostics).toEqual([]);

    const result = await session.evaluate("f()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(2n);
  });

  test("runs function defers while panicking", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`import "fmt"`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`
func boom() {
  defer fmt.Printf("cleanup\\n")
  panic("bad")
}
`)).diagnostics).toEqual([]);

    const result = await session.evaluate("boom()");
    expect(result.output).toEqual(["cleanup\n"]);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_PANIC001");
  });

  test("supports Go recover only from direct deferred function calls", async () => {
    const recovered = await expectRuns(`
func f() (out string) {
  defer func() {
    if r := recover(); r != nil {
      out = r.(string)
    }
  }()
  panic("caught")
}
return f(), recover() == nil
`);

    expect(recovered.values).toEqual(["caught", true]);

    const helper = await evaluateSource(`
func helper() interface{} { return recover() }
func f() {
  defer func() { _ = helper() }()
  panic("boom")
}
f()
`);
    expect(helper.diagnostics).toHaveLength(1);
    expect(helper.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(helper.diagnostics[0]?.message).toContain("boom");

    const directDefer = await evaluateSource(`
func f() {
  defer recover()
  panic("boom")
}
f()
`);
    expect(directDefer.diagnostics).toHaveLength(1);
    expect(directDefer.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(directDefer.diagnostics[0]?.message).toContain("boom");
  });

  test("returns named values and zero unnamed values after recover", async () => {
    const result = await expectRuns(`
func named() (x int) {
  defer func() {
    recover()
    x = 7
  }()
  panic("named")
}

func unnamed() int {
  defer func() { recover() }()
  panic("unnamed")
}

return named(), unnamed()
`);

    expect(result.values).toEqual([7n, 0n]);
  });

  test("formats Go-junior values with fmt %#v", async () => {
    const result = await expectRuns(`
import "fmt"

fmt.Printf("i = %#v\\n", 7)
return fmt.Sprintf("%#v %#v %#v", "x", 1.5, true)
`);

    expect(result.output).toEqual(["i = int64(7)\n"]);
    expect(result.value).toBe(`string("x") float64(1.5) bool(true)`);
  });

  test("formats fmt.Sprint with Go operand spacing", async () => {
    const result = await expectRuns(`
return fmt.Sprint("signed ", 1), fmt.Sprint(1, 2), fmt.Sprint("a", "b"), fmt.Sprint(1, "a", 2), fmt.Sprintln("a", 1, "b")
`);

    expect(result.values).toEqual(["signed 1", "1 2", "ab", "1a2", "a 1 b\n"]);
  });

  test("formats Go source through the TypeScript go/format port", () => {
    const expected = "func f(a int) int {\n\treturn a + 1\n}";
    expect(formatGoSource("func f(a int)int{return a+1}").trimEnd()).toBe(expected);

    const parsed = parseProgram("func f(a int)int{return a+1}", "format-node.go");
    const declaration = parsed.parsed.file?.declarations[0];
    expect(declaration ? formatGoNode(declaration) : "").toBe(expected);
  });

  test("displays saved formatted source for Go-junior function values", async () => {
    const session = new GoJuniorSession();
    const define = await session.evaluate("func f(a int)int{return a+1}");
    expect(define.diagnostics).toEqual([]);

    const lookup = await session.evaluate("f");
    expect(lookup.diagnostics).toEqual([]);
    expect(formatReplValue(lookup.value ?? null)).toBe("func f(a int) int {\n\treturn a + 1\n}");

    const closure = await expectRuns("f := func(a,b int)int{return a+b}\nreturn f");
    expect(formatReplValue(closure.value ?? null)).toBe("func(a, b int) int {\n\treturn a + b\n}");
  });

  test("automatically imports fmt in scripts and REPL sessions", async () => {
    const script = await expectRuns(`
fmt.Printf("auto %v\\n", 7)
return fmt.Sprintf("ok %v", 8)
`);
    expect(script.output).toEqual(["auto 7\n"]);
    expect(script.value).toBe("ok 8");

    const session = new GoJuniorSession();
    const define = await session.evaluate(`f := func() { fmt.Printf("hiya\\n") }`);
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("f()");
    expect(call.diagnostics).toEqual([]);
    expect(call.output).toEqual(["hiya\n"]);
    expect(call.value).toBeNull();
  });

  test("supports importing the testing package", async () => {
    const script = await expectRuns(`
import "testing"

return testing.Short(), testing.Verbose()
`);
    expect(script.values).toEqual([false, false]);

    const session = new GoJuniorSession();
    const define = await session.evaluate(`
import "testing"

func TestThing(t *testing.T) {
  t.Helper()
  t.Logf("x=%v", 1)
  if testing.Short() {
    t.SkipNow()
  }
}
`);
    expect(define.diagnostics).toEqual([]);
  });

  test("supports importing os.Exit without ambient process authority", async () => {
    const script = await expectRuns(`
import "os"

if false {
  os.Exit(1)
}
return 7
`);
    expect(script.value).toBe(7n);

    const exit = await evaluateSource(`
import "os"

os.Exit(3)
`);
    expect(exit.diagnostics).toHaveLength(1);
    expect(exit.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(exit.diagnostics[0]?.message).toBe("os.Exit(3)");
  });

  test("supports explicit deterministic os.Getenv bindings", async () => {
    const empty = await expectRuns(`
import "os"
return os.Getenv("MISSING")
`);
    expect(empty.value).toBe("");

    const configured = await expectRuns(`
import "os"
return os.Getenv("GOJR_MODE")
`, {
      env: {
        GOJR_MODE: "test"
      }
    });
    expect(configured.value).toBe("test");
  });

  test("supports importing math.NaN for float map keys and clear", async () => {
    const script = await expectRuns(`
import "math"

m := make(map[float64]int)
m[math.NaN()] = 1
m[math.NaN()] = 2
before := len(m)
clear(m)
return before, len(m)
`);

    expect(script.values).toEqual([2n, 0n]);
  });

  test("supports strconv.Itoa and math float bit helpers", async () => {
    const script = await expectRuns(`
import "math"
import "strconv"

f32 := math.Float32frombits(1 << 31)
f64 := math.Float64frombits(1 << 63)
return strconv.Itoa(-12), strconv.Itoa(34), math.Float32bits(f32), math.Float64bits(f64), math.IsNaN(math.NaN())
`);

    expect(script.values).toEqual(["-12", "34", 2147483648n, 9223372036854775808n, true]);
  });

  test("matches Go float map-key semantics for signed zero and NaN", async () => {
    const script = await expectRuns(`
import "math"

positiveZero := 0.0
negativeZero := math.Float64frombits(1 << 63)
m := map[float64]string{positiveZero: "+0"}
before := m[negativeZero]
m[negativeZero] = "-0"

nanA := math.NaN()
nanB := math.Float64frombits(math.Float64bits(nanA) ^ 2)
m[nanA] = "nan-a"
m[nanB] = "nan-b"
_, okA := m[nanA]
_, okB := m[nanB]
return before, m[positiveZero], len(m), okA, okB, math.IsNaN(nanB)
`);

    expect(script.values).toEqual(["+0", "-0", 3n, false, false, true]);
  });

  test("converts map lookup and delete keys to the declared key type", async () => {
    const script = await expectRuns(`
mf := map[float64]string{0: "zero", 1.0: "one"}
before := mf[0]
_, okBefore := mf[1]
delete(mf, 1)
_, okAfter := mf[1.0]

mi := map[int]string{0.0: "int-zero"}
return before, okBefore, okAfter, mi[0.0]
`);

    expect(script.values).toEqual(["zero", true, false, "int-zero"]);
  });

  test("compares NaN values with Go ordered-comparison semantics", async () => {
    const script = await expectRuns(`
import "math"

nan := math.NaN()
f := 1.0
return nan == nan, nan != nan, nan < nan, nan <= nan, nan > nan, nan >= nan,
  f < nan, f <= nan, f > nan, f >= nan,
  nan < f, nan <= f, nan > f, nan >= f
`);

    expect(script.values).toEqual([
      false, true, false, false, false, false,
      false, false, false, false,
      false, false, false, false
    ]);
  });

  test("evaluates source packages and lets formulas import their exported runtime values", async () => {
    const pkg = await evaluatePackageSourceFiles([{
      filename: "counter.go",
      source: `package counter

var Count int

func init() {
  Count = 40
}

func Next() int {
  Count++
  return Count
}
`
    }], { importPath: "example.com/counter" });

    expect(pkg.diagnostics).toEqual([]);
    expect(pkg.package).toBeDefined();

    const first = await expectRuns(`
import counter "example.com/counter"
return counter.Next()
`, {
      packages: {
        "example.com/counter": pkg.package ?? {}
      }
    });
    const second = await expectRuns(`
import counter "example.com/counter"
return counter.Next()
`, {
      packages: {
        "example.com/counter": pkg.package ?? {}
      }
    });

    expect(first.value).toBe(41n);
    expect(second.value).toBe(42n);
  });

  test("typechecks and evaluates source packages that import other source packages", async () => {
    const lib = await evaluatePackageSourceFiles([{
      filename: "lib.go",
      source: `package lib

func One() int {
  return 1
}
`
    }], { importPath: "example.com/lib" });
    expect(lib.diagnostics).toEqual([]);
    expect(lib.package).toBeDefined();
    expect(lib.packageInfo).toBeDefined();

    const app = await evaluatePackageSourceFiles([{
      filename: "app.go",
      source: `package app

import lib "example.com/lib"

func Two() int {
  return lib.One() + 1
}
`
    }], {
      importPath: "example.com/app",
      packages: {
        "example.com/lib": lib.package ?? {}
      },
      packageInfos: {
        "example.com/lib": lib.packageInfo!
      }
    });
    expect(app.diagnostics).toEqual([]);
    expect(app.package).toBeDefined();

    const result = await expectRuns(`
import app "example.com/app"
return app.Two()
`, {
      packages: {
        "example.com/app": app.package ?? {},
        "example.com/lib": lib.package ?? {}
      }
    });
    expect(result.value).toBe(2n);
  });

  test("runs Go-junior tests that import source package metadata", async () => {
    const app = await evaluatePackageSourceFiles([{
      filename: "app.go",
      source: `package app

func Two() int {
  return 2
}
`
    }], { importPath: "example.com/app" });
    expect(app.diagnostics).toEqual([]);
    expect(app.package).toBeDefined();
    expect(app.packageInfo).toBeDefined();

    const result = await testSourceFiles([{
      filename: "use_app_test.go",
      source: `package useapp

import (
  app "example.com/app"
  "testing"
)

func TestTwo(t *testing.T) {
  if app.Two() != 2 {
    t.Fatalf("bad Two")
  }
}
`
    }], {
      packages: {
        "example.com/app": app.package ?? {}
      },
      packageInfos: {
        "example.com/app": app.packageInfo!
      }
    });
    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("PASS");
  });

  test("runs Go-junior tests with testing.T", async () => {
    const result = await testSource(`
import "testing"

func Add(a, b int) int { return a + b }

func TestAdd(t *testing.T) {
  t.Helper()
  t.Logf("sum=%v", Add(2, 3))
  if Add(2, 3) != 5 {
    t.Fatalf("bad sum")
  }
}

func TestNoArg() {}
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.output.join("")).toContain("=== RUN   TestAdd\n");
    expect(result.output.join("")).toContain("    sum=5\n");
    expect(result.output.join("")).toContain("--- PASS: TestAdd\n");
    expect(result.output.join("")).toContain("--- PASS: TestNoArg\n");
    expect(result.output.join("")).toContain("PASS\n");
  });

  test("reports failing and skipped Go-junior tests", async () => {
    const result = await testSource(`
import "testing"

func TestFail(t *testing.T) {
  t.Errorf("bad %v", 3)
}

func TestSkip(t *testing.T) {
  t.Skipf("skip %v", 4)
}
`);

    const output = result.output.join("");
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_TEST001");
    expect(output).toContain("=== RUN   TestFail\n");
    expect(output).toContain("    bad 3\n");
    expect(output).toContain("--- FAIL: TestFail\n");
    expect(output).toContain("=== RUN   TestSkip\n");
    expect(output).toContain("    skip 4\n");
    expect(output).toContain("--- SKIP: TestSkip\n");
    expect(output).toContain("FAIL\n");
  });

  test("validates Go-junior test signatures", async () => {
    const result = await testSource(`
func TestBad(t int) {}
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_TEST001");
    expect(result.diagnostics[0]?.message).toContain("func TestX(t *testing.T)");
    expect(result.output.join("")).toContain("--- FAIL: TestBad\n");
  });

  test("supports declared maps with typed string and integer keys", async () => {
    const stringKeyed = await expectRuns(`
var m map[string]int
m["hi"] = 3
return m["hi"]
`);
    expect(stringKeyed.value).toBe(3n);

    const intKeyed = await expectRuns(`
var mm map[int]string
mm[3] = "hi"
return mm[3]
`);
    expect(intKeyed.value).toBe("hi");
  });

  test("supports Go two-value map lookups for key presence", async () => {
    const missing = await expectRuns(`
var m map[int]int
a, ok := m[3]
return a, ok
`);
    expect(missing.values).toEqual([0n, false]);

    const present = await expectRuns(`
var m map[int]string
m[3] = "hi"
a, ok := m[3]
return a, ok
`);
    expect(present.values).toEqual(["hi", true]);
  });

  test("supports standard Go make for maps and slices", async () => {
    const mapResult = await expectRuns(`
m := make(map[int]int)
missing, missingOK := m[3]
m[3] = 9
present, presentOK := m[3]
return missing, missingOK, present, presentOK
`);
    expect(mapResult.values).toEqual([0n, false, 9n, true]);

    const sliceResult = await expectRuns(`
slc := make([]int, 20, 50)
slc = append(slc, 7)
slc2 := make([]string, 3)
return len(slc), cap(slc), slc[0], slc[20], len(slc2), cap(slc2), slc2[0]
`);
    expect(sliceResult.values).toEqual([21n, 50n, 0n, 7n, 3n, 3n, ""]);
  });

  test("reports ordinary runtime failures with GoJr-prefixed diagnostic codes", async () => {
    const result = await evaluateSource(`
missingName
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(result.diagnostics[0]?.message).toContain("missingName is not declared");
  });

  test("keeps filenames on source-file diagnostics", async () => {
    const result = await evaluateSourceFiles([
      {
        filename: "pkg/bad.go",
        source: `
package pkg

func Bad() {
  return )
}
`
      }
    ]);

    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.filename).toBe("pkg/bad.go");
    expect(result.diagnostics[0]?.span?.filename).toBe("pkg/bad.go");
    expect(result.diagnostics[0]?.span?.line).toBe(5);
  });

  test("predeclares top-level types before variable initializers", async () => {
    const result = await expectRuns(`
var p = S1{
  F1: complex(float64(2.5), float64(-0.25)),
  F2: S2{F1: 9},
  F3: 103050709,
}

var d Derived = Box{X: 5}

type S1 struct {
  F1 complex128
  F2 S2
  F3 uint64
}

type S2 struct {
  F1 uint64
  F2 empty
}

type empty struct{}

type Derived interface { Base }
type Base interface { String() string }

type Box struct { X int }
func (b Box) String() string { return fmt.Sprintf("Box(%v)", b.X) }

return p.F2.F1, p.F3, d.String()
`);

    expect(result.values).toEqual([9n, 103050709n, "Box(5)"]);
  });

  test("enforces declared and inferred local variable types", async () => {
    const declared = await evaluateSource(`
var x int
x = "bad"
`);
    expect(declared.diagnostics).toHaveLength(1);
    expect(declared.diagnostics[0]?.message).toContain("variable x bad is not assignable to int");

    const inferred = await evaluateSource(`
x := 1
x = "bad"
`);
    expect(inferred.diagnostics).toHaveLength(1);
    expect(inferred.diagnostics[0]?.message).toContain("variable x bad is not assignable to int64");
  });

  test("rejects mixed string and numeric addition", async () => {
    const direct = await evaluateSource(`
d := \` hi there\`
a := 10
return d + a
`);
    expect(direct.diagnostics).toHaveLength(1);
    expect(direct.diagnostics[0]?.code).toBe("GOJR_RUNTIME001");
    expect(direct.diagnostics[0]?.message).toContain("invalid operation: string + int64");

    const session = new GoJuniorSession();
    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate("d := ` hi there`")).diagnostics).toEqual([]);
    const mixed = await session.evaluate("d + a");
    expect(mixed.diagnostics).toHaveLength(1);
    expect(mixed.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(mixed.diagnostics[0]?.message).toContain("invalid operation: string + int64");

    const strings = await expectRuns(`
s := "hi"
return s + " there"
`);
    expect(strings.value).toBe("hi there");
  });

  test("compares huge untyped integer constants without losing precision", async () => {
    const result = await expectRuns(`
const (
  chuge = 1 << 100
  chuge_1 = chuge - 1
)
return chuge > chuge_1, chuge == chuge_1, chuge_1 + 1 == chuge
`);

    expect(result.values).toEqual([true, false, true]);
  });

  test("supports Go-style const groups with iota and repeated expressions", async () => {
    const result = await expectRuns(`
const Single = iota
const (
  A = iota
  B
  C int = iota
  D
)
const (
  F float32 = 2 * iota
  G complex128 = iota
)
const (
  abit, amask = 1 << iota, 1<<iota - 1
  bbit, bmask
)
const (
  PackageX = 2
)
func shadowConst() (int, int, int, int, int, int) {
  const (
    First = iota
    iota = iota
    ShadowedA
    ShadowedB
  )
  const (
    PackageX = PackageX + PackageX
    LocalY
    LocalZ = iota
  )
  return First, ShadowedA, ShadowedB, PackageX, LocalY, LocalZ
}
First, ShadowedA, ShadowedB, LocalX, LocalY, LocalZ := shadowConst()
return Single, A, B, C, D, F, G, abit, amask, bbit, bmask, First, ShadowedA, ShadowedB, LocalX, LocalY, LocalZ
`);

    expect(result.values).toEqual([0n, 0n, 1n, 2n, 3n, 0, { real: 1, imag: 0 }, 1n, 0n, 2n, 1n, 0n, 1n, 1n, 4n, 8n, 1n]);

    const immutable = await evaluateSource(`
const X = 1
X = 2
`);
    expect(immutable.diagnostics).toHaveLength(1);
    expect(immutable.diagnostics[0]?.message).toContain("X is const");
  });

  test("enforces function parameter and typed collection assignments", async () => {
    const badParam = await evaluateSource(`
func f(x int) int { return x }
return f("bad")
`);
    expect(badParam.diagnostics).toHaveLength(1);
    expect(badParam.diagnostics[0]?.message).toContain("bad is not assignable to int");

    const badArray = await evaluateSource(`
var xs []int
xs = []string{"bad"}
`);
    expect(badArray.diagnostics).toHaveLength(1);
    expect(badArray.diagnostics[0]?.message).toContain("variable xs element bad is not assignable to int");

    const badMap = await evaluateSource(`
var m map[string]int
m = map[int]string{1: "bad"}
`);
    expect(badMap.diagnostics).toHaveLength(1);
    expect(badMap.diagnostics[0]?.message).toContain("variable m");
  });

  test("reports map key and value type mismatches without numeric-index coercion", async () => {
    const badKey = await evaluateSource(`
var mm map[int]string
mm["hi"] = 3
`);
    expect(badKey.diagnostics).toHaveLength(1);
    expect(badKey.diagnostics[0]?.message).toContain("map key hi is not assignable to int");
    expect(badKey.diagnostics[0]?.message).not.toContain("numeric");

    const badValue = await evaluateSource(`
var m map[string]int
m["hi"] = "three"
`);
    expect(badValue.diagnostics).toHaveLength(1);
    expect(badValue.diagnostics[0]?.message).toContain("map value three is not assignable to int");
  });

  test("evaluates map literals, missing-key zero values, and insertion-order range", async () => {
    const result = await expectRuns(`
counts := map[string]int64{"a": 1, "b": 2}
out := ""
for k, v := range counts {
  out = out + k + ":" + fmt.Sprintf("%v", v) + ";"
}
return counts["a"] + counts["b"] + counts["missing"], out
`);

    expect(result.values).toEqual([3n, "a:1;b:2;"]);
  });

  test("evaluates named map composite literals", async () => {
    const result = await expectRuns(`
type M map[int]int
m := M{0: 10, 1: 20}
m[2] = 30
v, ok := m[1]
missing, missingOK := m[3]
return m[0] + v + m[2] + missing, ok, missingOK, len(m)
`);

    expect(result.values).toEqual([60n, true, false, 3n]);
  });

  test("formats typed maps with fmt verbs", async () => {
    const result = await expectRuns(`
import "fmt"

counts := map[string]int64{"a": 1, "b": 2}
var floats map[int]float64
empty := fmt.Sprintf("%v", floats)
floats[39] = 3.2
return fmt.Sprintf("%v | %#v | %v | %v", counts, counts, empty, floats)
`);

    expect(result.value).toBe(`map[string]int64{a:1 b:2} | map[string]int64{string("a"): int64(1), string("b"): int64(2)} | map[int]float64{} | map[int]float64{39:3.2}`);
  });

  test("formats REPL map string values as Go literals", async () => {
    const result = await expectRuns(`
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

  test("evaluates struct literals, zero values, field mutation, and fmt verbs", async () => {
    const result = await expectRuns(`
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

  test("evaluates elided composite literals in typed array and map literals", async () => {
    const result = await expectRuns(`
type Point struct{ X, Y int }
rows := []struct {
  Name string
  Pos Point
}{
  {"a", Point{1, 2}},
  {"b", {Y: 4}},
}
lookup := map[Point]Point{
  {X: 1}: {Y: 2},
}
return rows[0].Name, rows[1].Pos.Y, lookup[Point{X: 1}].Y
`);

    expect(result.values).toEqual(["a", 4n, 2n]);
  });

  test("supports value and pointer receiver methods with Go selector syntax", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`
type Point struct {
  X, Y int
}
`)).diagnostics).toEqual([]);

    const valueMethod = await session.evaluate(`
func (p Point) Sum() int {
  return p.X + p.Y
}
`);
    expect(valueMethod.diagnostics).toEqual([]);

    const pointerMethod = await session.evaluate(`
func (p *Point) Scale(k int) {
  p.X = p.X * k
  p.Y = p.Y * k
}
`);
    expect(pointerMethod.diagnostics).toEqual([]);

    const call = await session.evaluate(`
p := Point{2, 3}
before := p.Sum()
p.Scale(4)
return before, p.X, p.Y, p.Sum()
`);
    expect(call.diagnostics).toEqual([]);
    expect(call.values).toEqual([5n, 8n, 12n, 20n]);
  });

  test("supports Go method expressions on named receiver types", async () => {
    const result = await expectRuns(`
type T []int

func (t T) Len() int {
  return len(t)
}

type Counter int

func (c *Counter) IncBy(k int) {
  *c = *c + Counter(k)
}

var t T = T{0, 1, 2, 3, 4}
var c Counter
f := T.Len
g := (*T).Len
h := (*Counter).IncBy
h(&c, 3)
return T.Len(t), f(t), g(&t), c
`);

    expect(result.values).toEqual([5n, 5n, 5n, 3n]);
  });

  test("supports interface and promoted method expressions", async () => {
    const result = await expectRuns(`
got := ""

type I interface {
  m()
}

type S struct{}

func (S) m() {
  got += "m;"
}

func (S) m1(s string) {
  got += "m1(" + s + ");"
}

type T int

func (T) m2() {
  got += "m2;"
}

type Outer struct { *Inner }
type Inner struct { s string }

func (i Inner) M() string {
  return i.s
}

I.m(S{})
f := interface{ m1(string) }.m1
f(S{}, "a")
interface{ m1(string) }.m1(S{}, "b")
g := struct{ T }.m2
g(struct{ T }{})
h := (*Outer).M
return got, h(&Outer{&Inner{"hello"}})
`);

    expect(result.values).toEqual(["m;m1(a);m1(b);m2;", "hello"]);
  });

  test("supports promoted pointer receiver method expressions", async () => {
    const result = await expectRuns(`
type Scalar int

func (s *Scalar) M(a int, x [2]int, b float64, y [2]float64) (Scalar, int, [2]int, float64, [2]float64) {
  return *s, a, x, b, y
}

type Wrapper struct {
  Scalar
}

var scalar Scalar = 42
var wrapper = &Wrapper{Scalar: scalar}
fn := (*Wrapper).M
s1, a1, x1, b1, y1 := fn(wrapper, 123, [2]int{456, 789}, 1.2, [2]float64{3.4, 5.6})
return s1 == scalar && a1 == 123 && x1 == [2]int{456, 789} && b1 == 1.2 && y1 == [2]float64{3.4, 5.6}
`);

    expect(result.value).toBe(true);
  });

  test("REPL sessions typecheck interface method expressions from earlier declarations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`got := ""`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`type I interface { m() }`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`type S struct{}`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`func (S) m() { got += "m" }`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`I.m(S{})`)).diagnostics).toEqual([]);

    const result = await session.evaluate("got");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe("m");
  });

  test("keeps value selectors ahead of type method expressions when a type name is shadowed", async () => {
    const result = await expectRuns(`
type T struct { X int }

func (t T) XPlus(k int) int {
  return t.X + k
}

T := T{X: 7}
return T.X
`);

    expect(result.value).toBe(7n);
  });

  test("supports address-of and dereference assignment for structs and fields", async () => {
    const result = await expectRuns(`
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

  test("executes Go type switches over structs, pointers, nil, and basic values", async () => {
    const result = await expectRuns(`
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

  test("supports typed nil pointer conversions", async () => {
    const result = await expectRuns(`
var p *int
var five = 5
accept := func(p *int) bool { return p != nil }
return (*int)(nil) == p, fmt.Sprintf("%#v", (*int)(nil)), accept(&five)
`);

    expect(result.values).toEqual([true, "*int(nil)", true]);
  });

  test("allows methods with pointer receivers on typed nil pointers", async () => {
    const result = await expectRuns(`
type T []T

func (*T) Sum(args ...int) int {
  s := 0
  for _, v := range args {
    s += v
  }
  return s
}

return ((*T)(nil)).Sum(1, 3, 5, 7), (*T).Sum(nil, 1, 3, 5, 6)
`);

    expect(result.values).toEqual([16n, 15n]);
  });

  test("preserves concrete dynamic numeric types inside interfaces", async () => {
    const result = await expectRuns(`
type Duration int

describe := func(x interface{}) string {
  switch x.(type) {
  case int:
    return "int"
  case int64:
    return "int64"
  case uint:
    return "uint"
  case Duration:
    return "duration"
  default:
    return "other"
  }
}

var d interface{} = Duration(5)
sameDuration, okDuration := d.(Duration)
_, okInt := d.(int)

return describe(1), describe(int64(1)), describe(uint(1)), describe(Duration(1)), okDuration, int(sameDuration), okInt
`);

    expect(result.values).toEqual(["int", "int64", "uint", "duration", true, 5n, false]);
  });

  test("preserves interface dynamic types through index assignments", async () => {
    const result = await expectRuns(`
type Duration int

var xs [3]interface{}
xs[0] = 1
xs[1] = int64(2)
xs[2] = Duration(3)
_, xs0Int := xs[0].(int)
_, xs0Int64 := xs[0].(int64)
_, xs1Int64 := xs[1].(int64)
_, xs2Duration := xs[2].(Duration)

slc := make([]interface{}, 1)
slc[0] = 4
_, slcInt := slc[0].(int)

var m map[string]interface{}
m["x"] = 5
m["y"] = int64(6)
_, mapInt := m["x"].(int)
_, mapInt64 := m["y"].(int64)

return xs0Int, xs0Int64, xs1Int64, xs2Duration, slcInt, mapInt, mapInt64
`);

    expect(result.values).toEqual([true, false, true, true, true, true, true]);
  });

  test("treats byte and rune as aliases in interface type switches", async () => {
    const result = await expectRuns(`
var x interface{}
x = byte(1)
byteIsUint8 := false
switch x.(type) {
case uint8:
  byteIsUint8 = true
}
x = uint8(2)
uint8IsByte := false
switch x.(type) {
case byte:
  uint8IsByte = true
}
x = rune(3)
runeIsInt32 := false
switch x.(type) {
case int32:
  runeIsInt32 = true
}
x = int32(4)
int32IsRune := false
switch x.(type) {
case rune:
  int32IsRune = true
}
return byteIsUint8, uint8IsByte, runeIsInt32, int32IsRune
`);

    expect(result.values).toEqual([true, true, true, true]);
  });

  test("executes unbound type switches and rejects fallthrough", async () => {
    const ok = await expectRuns(`
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

    const bad = await evaluateSource(`
var x interface{} = 1
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

  test("runs switch default only after checking non-default cases", async () => {
    const result = await expectRuns(`
value := ""
switch 2 {
default:
  value = "default"
case 2:
  value = "case"
}

typed := ""
var x interface{} = 1
switch x.(type) {
default:
  typed = "default"
case int:
  typed = "int"
}

return value, typed
`);

    expect(result.values).toEqual(["case", "int"]);
  });

  test("evaluates Go type assertions and reports mismatches", async () => {
    const ok = await expectRuns(`
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

    const bad = await evaluateSource(`
x := "hello"
return x.(int)
`);
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.message).toContain("does not have dynamic type int");
  });

  test("enforces named interface method sets on typed variables", async () => {
    const ok = await expectRuns(`
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

    const pointerOnly = await evaluateSource(`
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

    const pointerOk = await expectRuns(`
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

    const promotedPointerOk = await expectRuns(`
type Summable interface {
  Sum(...int) int
}

type T []T
func (*T) Sum(args ...int) int {
  total := 0
  for _, v := range args {
    total += v
  }
  return total
}

type U struct { *T }

var u U
var s Summable = u
var holder struct { Summable }
holder.Summable = &u
return s.Sum(2, 3, 5, 6), holder.Sum(2, 3, 5, 8)
`);
    expect(promotedPointerOk.values).toEqual([16n, 18n]);
  });

  test("represents interfaces as typed runtime values with dynamic nil state", async () => {
    const nilInterface = await expectRuns(`
type I interface {
  fun()
}

var i I
return i, i == nil
`);

    expect(nilInterface.values?.map(formatReplValue)).toEqual(["I(nil)", "true"]);

    const typedNilPointer = await expectRuns(`
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

    const empty = await expectRuns(`
type I interface {
  fun()
}

var i I
var x interface{} = i
return x, x == nil
`);

    expect(empty.values?.map(formatReplValue)).toEqual(["interface{}(nil)", "true"]);
  });

  test("evaluates array and slice literals with indexing, slicing, and range", async () => {
    const result = await expectRuns(`
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

  test("supports len, cap, append, and default slice/array declarations", async () => {
    const result = await expectRuns(`
var xs []int
var ys [3]string
xs = append(xs, 1, 2)
more := []int{3, 4}
xs = append(xs, more...)
return xs, len(xs), cap(xs), len(ys), ys[0]
`);

    expect(result.values).toEqual([[1n, 2n, 3n, 4n], 4n, 4n, 3n, ""]);
  });

  test("supports reslicing make slices up to capacity with zero-filled backing storage", async () => {
    const result = await expectRuns(`
s := make([]int, 2, 5)
s[0] = 7
grown := s[0:5]
grown[3] = 11
again := s[0:5]
return len(s), cap(s), len(grown), cap(grown), grown[0], grown[1], grown[3], again[3]
`);

    expect(result.values).toEqual([2n, 5n, 5n, 5n, 7n, 0n, 11n, 11n]);
  });

  test("supports len and cap on pointers to arrays", async () => {
    const result = await expectRuns(`
p := new([4]int)
var nilp *[3]string
return len(p), cap(p), len(nilp), cap(nilp)
`);

    expect(result.values).toEqual([4n, 4n, 3n, 3n]);
  });

  test("supports constant identifiers in array lengths", async () => {
    const result = await expectRuns(`
const size = 4
var a [size]byte
for k := range a {
  a[k] = byte(k + 1)
}
return len(a), a[0], a[3]
`);

    expect(result.values).toEqual([4n, 1n, 4n]);
  });

  test("reports typed array and slice literal element mismatches", async () => {
    const result = await evaluateSource(`
xs := []int{1, "bad"}
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("array element bad is not assignable to int");
  });

  test("supports len on strings and maps", async () => {
    const result = await expectRuns(`
m := map[string]int{"a": 1, "b": 2}
return len("hiya"), len(m)
`);

    expect(result.values).toEqual([4n, 2n]);
  });

  test("reports panicOn failures as runtime diagnostics", async () => {
    const result = await evaluateSource(`
panicOn("bad")
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_PANIC001");
  });

  test("keeps REPL session locals across eager evaluations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    const result = await session.evaluate("a + 2");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(12n);
  });

  test("checks REPL const groups with scoped iota", async () => {
    const session = new GoJuniorSession();

    const result = await session.evaluate(`
const (
  abit, amask = 1 << iota, 1<<iota - 1
  bbit, bmask
)
return abit, amask, bbit, bmask
`);

    expect(result.diagnostics).toEqual([]);
    expect(result.values).toEqual([1n, 0n, 2n, 1n]);
  });

  test("updates REPL sheet type environment when sheet data changes", async () => {
    const session = new GoJuniorSession();

    session.setSheet({ A1: 40n });
    const result = await session.evaluate("sheet.A1 + 2");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(42n);
    expect(result.observedDeps).toEqual([
      cellDependency({ sheet: "sheet", cell: "A1" })
    ]);

    const next = await session.evaluate("sheet.A1:B1");
    expect(next.diagnostics).toEqual([]);
    expect(next.observedDeps).toEqual([
      rangeDependency("sheet", "A1", "B1")
    ]);
  });

  test("evaluates Go raw string literals", async () => {
    const result = await expectRuns("a := `hi\nthere`\nreturn a");

    expect(result.value).toBe("hi\nthere");
  });

  test("supports Go-style increment and decrement statements in sessions", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate("a++")).diagnostics).toEqual([]);
    expect((await session.evaluate("a")).value).toBe(11n);
    expect((await session.evaluate("a--")).diagnostics).toEqual([]);
    expect((await session.evaluate("a")).value).toBe(10n);
  });

  test("accepts top-level semicolons in REPL session input", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("a := 10;")).diagnostics).toEqual([]);
    const result = await session.evaluate("a");

    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(10n);
  });

  test("accepts import-only REPL input and does not replay output", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate(`import "fmt"`)).diagnostics).toEqual([]);
    expect((await session.evaluate(`fmt.Printf("hi")`)).output).toEqual(["hi"]);
    expect((await session.evaluate("a := 1")).output).toEqual([]);
  });

  test("defines and calls functions with grouped names in REPL sessions", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate("func f(a, b, c int) (x, y, z int) { return a, b, c }");
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("f(1, 2, 3)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toEqual([1n, 2n, 3n]);
  });

  test("supports named result variables and naked returns in REPL functions", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate(`
func swap(a, b int) (left, right int) {
  left = b
  right = a
  return
}
`);
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("swap(1, 2)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toEqual([2n, 1n]);
  });

  test("supports variadic parameters and spread calls in REPL functions", async () => {
    const session = new GoJuniorSession({
      sheet: {
        A1: [4n, 5n, 6n]
      }
    });

    const define = await session.evaluate(`
func sum(vals ...int) int {
  total := 0
  for _, v := range vals {
    total = total + v
  }
  return total
}
`);
    expect(define.diagnostics).toEqual([]);

    const direct = await session.evaluate("sum(1, 2, 3)");
    expect(direct.diagnostics).toEqual([]);
    expect(direct.value).toBe(6n);

    const spread = await session.evaluate("sum(sheet.A1...)");
    expect(spread.diagnostics).toEqual([]);
    expect(spread.value).toBe(15n);
  });

  test("supports multiple short declarations and assignments", async () => {
    const result = await expectRuns(`
a, b := 1, 2
a, b = b, a
return a, b
`);

    expect(result.values).toEqual([2n, 1n]);
  });

  test("supports destructuring multiple function returns", async () => {
    const result = await expectRuns(`
func divmod(x, y int) (int, int) {
  return x / y, x % y
}

q, r := divmod(17, 5)
_, onlyR := divmod(19, 5)
return q, r, onlyR
`);

    expect(result.values).toEqual([3n, 2n, 4n]);
  });

  test("supports multi-assignment to index targets after evaluating rhs", async () => {
    const result = await expectRuns(`
xs := []int{1, 2}
xs[0], xs[1] = xs[1], xs[0]
return xs
`);

    expect(result.value).toEqual([2n, 1n]);
  });

  test("defines and calls function literals with grouped names in REPL sessions", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate("f := func(a, b, c int) (d, e, f int) { return b, c, a }");
    expect(define.diagnostics).toEqual([]);

    const call = await session.evaluate("f(1, 2, 3)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toEqual([2n, 3n, 1n]);
  });

  test("closures capture lexical variables by reference", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("base := 10")).diagnostics).toEqual([]);
    expect((await session.evaluate("addBase := func(x int) int { return base + x }")).diagnostics).toEqual([]);
    expect((await session.evaluate("base = 20")).diagnostics).toEqual([]);

    const call = await session.evaluate("addBase(2)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toBe(22n);
  });

  test("keeps REPL top-level channel variables visible to later function declarations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("func f() { for i := range 5 { c <- i } }")).diagnostics).toEqual([]);
    expect((await session.evaluate("go f()")).diagnostics).toEqual([]);

    const first = await session.evaluate("<-c");
    const second = await session.evaluate("<-c");

    expect(first.diagnostics).toEqual([]);
    expect(second.diagnostics).toEqual([]);
    expect(first.value).toBe(0n);
    expect(second.value).toBe(1n);
  });

  test("returned closures keep their defining function scope alive", async () => {
    const session = new GoJuniorSession();

    const define = await session.evaluate(`
func makeAdder(base int) func(int) int {
  return func(x int) int {
    return base + x
  }
}
`);
    expect(define.diagnostics).toEqual([]);
    expect((await session.evaluate("add5 := makeAdder(5)")).diagnostics).toEqual([]);

    const call = await session.evaluate("add5(3)");
    expect(call.diagnostics).toEqual([]);
    expect(call.value).toBe(8n);
  });

  test("function literals support variadic parameters and spread calls", async () => {
    const session = new GoJuniorSession({
      sheet: {
        A1: [7n, 8n, 9n]
      }
    });

    const define = await session.evaluate(`
sum := func(vals ...int) int {
  total := 0
  for _, v := range vals {
    total = total + v
  }
  return total
}
`);
    expect(define.diagnostics).toEqual([]);

    const direct = await session.evaluate("sum(1, 2, 3)");
    expect(direct.diagnostics).toEqual([]);
    expect(direct.value).toBe(6n);

    const spread = await session.evaluate("sum(sheet.A1...)");
    expect(spread.diagnostics).toEqual([]);
    expect(spread.value).toBe(24n);
  });

  test("marks incomplete REPL input without executing it", async () => {
    const session = new GoJuniorSession();
    const incomplete = await session.evaluate("if true {");

    expect(incomplete.incomplete).toBe(true);
    expect(incomplete.diagnostics.length).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.span?.line).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.span?.column).toBeGreaterThan(0);

    const complete = await session.evaluate(`
if true {
  return 1
}
`);

    expect(complete.diagnostics).toEqual([]);
    expect(complete.value).toBe(1n);
  });

  test("keeps incomplete raw strings pending in REPL input", async () => {
    const session = new GoJuniorSession();
    const incomplete = await session.evaluate("a := ` hi there");

    expect(incomplete.incomplete).toBe(true);
    expect(incomplete.diagnostics.length).toBeGreaterThan(0);
    expect(incomplete.diagnostics[0]?.message).toContain("unterminated raw string");
  });

  test("keeps for blocks with increment statements pending until closed", async () => {
    const session = new GoJuniorSession();
    const result = await session.evaluate("for {\n  a++");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.span?.line).not.toBeNaN();
    expect(result.diagnostics[0]?.span?.column).not.toBeNaN();
  });

  test("keeps Go-style for clauses pending until the block is closed", async () => {
    const session = new GoJuniorSession();
    const result = await session.evaluate("for b := 0; b < 10; b++ {");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
    expect(result.diagnostics[0]?.message).not.toContain("':='");
  });

  test("keeps bare labels pending until their statement is entered", async () => {
    const session = new GoJuniorSession();
    const result = await session.evaluate("top:");

    expect(result.incomplete).toBe(true);
    expect(result.diagnostics.length).toBeGreaterThan(0);
  });

  test("executes Go-style for init condition and post clauses", async () => {
    const result = await expectRuns(`
sum := 0
for b := 0; b < 10; b++ {
  sum = sum + b
}
return sum
`);

    expect(result.value).toBe(45n);
  });

  test("executes Go-style for clauses with omitted init statements", async () => {
    const result = await expectRuns(`
sum := 0
b := 0
for ; b < 5; b++ {
  if b == 2 {
    continue
  }
  sum += b
}
for ; ; b-- {
  if b == 0 {
    break
  }
  sum++
}
return sum
`);

    expect(result.value).toBe(13n);
  });

  test("runs for post clause after continue", async () => {
    const result = await expectRuns(`
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

  test("supports REPL entry of labeled nested loops with labeled break", async () => {
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
      const result = await session.evaluate(source);
      if (index < lines.length - 1) {
        expect(result.incomplete).toBe(true);
      } else {
        expect(result.diagnostics).toEqual([]);
        expect(result.output).toEqual(["i=0 j=0\n"]);
      }
    }
  });

  test("supports goto to forward and backward labels", async () => {
    const forward = await expectRuns(`
i := 0
goto Done
i = 99
Done:
return i
`);
    expect(forward.value).toBe(0n);

    const backward = await expectRuns(`
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

  test("supports labeled break out of nested loops", async () => {
    const result = await expectRuns(`
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

  test("supports labeled continue from inside a switch to an outer for", async () => {
    const result = await expectRuns(`
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

  test("keeps unlabeled break scoped to switch inside for", async () => {
    const result = await expectRuns(`
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

  test("lets unlabeled continue inside switch continue the containing for", async () => {
    const result = await expectRuns(`
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

  test("supports Go conversions, numeric literals, rune literals, imaginary literals, and complex builtins", async () => {
    const result = await expectRuns(`
a := int(0b1010)
b := int64(0x10)
c := rune('A')
d := byte(0o7)
z := complex(float64(a), 2.5)
return a, b, c, d, real(z), imag(z), 3i + 2i
`);

    expect(result.values).toEqual([10n, 16n, 65n, 7n, 10, 2.5, { real: 0, imag: 5 }]);
  });

  test("rounds float32 assignments, conversions, and expression results like Go", async () => {
    const result = await expectRuns(`
func f32(v float64) float32 { return float32(v) }

var f09 float32 = 1e-10
var f10 float32 = 1e+10
var f13 float32 = .1e-10
var f14 float32 = .1e+10
var c64 complex64 = complex(1.1, .1e-10)
return f13 == f09/10.0, f14 == f10/10.0, float64(float32(1.1)) == float64(f32(1.1)), real(c64) == f32(1.1)
`);

    expect(result.values).toEqual([true, true, true, true]);
  });

  test("wraps typed integer expression results like Go", async () => {
    const result = await expectRuns(`
func f8(x, y int8) (int8, int8) {
  return x / y, x % y
}
func f16(x, y int16) (int16, int16) {
  return x / y, x % y
}

q8, r8 := f8(-1<<7, -1)
q16, r16 := f16(-1<<15, -1)
return q8, r8, q16, r16
`);

    expect(result.values).toEqual([-128n, 0n, -32768n, 0n]);
  });

  test("preserves named numeric expression result identity", async () => {
    const result = await expectRuns(`
type A int
var a A = 1

_, plusA := interface{}(+a).(A)
_, addA := interface{}(a + 0).(A)
_, addInt := interface{}(a + 0).(int)
return plusA, addA, addInt
`);

    expect(result.values).toEqual([true, true, false]);
  });

  test("wraps explicit integer conversions like Go", async () => {
    const result = await expectRuns(`
a := int8(-32668)
b := uint8(-1)
c := int16(65535)
d := int32(uint32(0xffffffff))
e := uint64(-1)
return a, b, c, d, e
`);

    expect(result.values).toEqual([100n, 255n, -1n, -1n, 18446744073709551615n]);

    const assignment = await evaluateSource(`
var x int8 = 128
`);
    expect(assignment.diagnostics).toHaveLength(1);
    expect(assignment.diagnostics[0]?.message).toContain("not assignable to int8");
  });

  test("supports Go string conversions from byte and rune slices", async () => {
    const result = await expectRuns(`
bs := []byte{0xe1, 0x88, 0xb4}
rs := []rune{'a', '\\u1234', 'c'}
p := new([3]byte)
p[0] = 'x'
p[1] = 'y'
p[2] = 'z'
return string(bs), string(rs), string(p[0:])
`);

    expect(result.values).toEqual(["\u1234", "a\u1234c", "xyz"]);
  });

  test("treats Go strings as byte sequences for len index slice and escapes", async () => {
    const result = await expectRuns(`
s := "aä本☺"
largest := string(0x10ffff)
encoded := "\\xf4\\x8f\\xbf\\xbf"
bad := "\\xff\\xff"
return len(s), s[1], s[1:3], largest == encoded, []rune(bad), string([]byte(bad))
`);

    expect(result.values).toEqual([9n, 195n, "ä", true, [65533n, 65533n], "\ufffd\ufffd"]);
  });

  test("supports Go byte and rune slice conversions from strings", async () => {
    const result = await expectRuns(`
type Bytes []byte
type Runes []rune
s := "aä本☺"
bs := []byte(s)
rs := []rune(s)
nbs := Bytes(s)
nrs := Runes(s)
return bs, rs, string(bs), string(rs), string(nbs), string(nrs)
`);

    expect(result.values).toEqual([
      [97n, 195n, 164n, 230n, 156n, 172n, 226n, 152n, 186n],
      [97n, 228n, 26412n, 9786n],
      "aä本☺",
      "aä本☺",
      "aä本☺",
      "aä本☺"
    ]);
  });

  test("supports typed nil slice conversions with Go len and cap", async () => {
    const result = await expectRuns(`
type Ints []int
s := []int(nil)
named := Ints(nil)
return len(s), cap(s), len(named), cap(named), fmt.Sprintf("%v", s), fmt.Sprintf("%#v", named)
`);

    expect(result.values).toEqual([0n, 0n, 0n, 0n, "<nil>", "Ints(nil)"]);
  });

  test("keeps inferred var declaration types for later conversions", async () => {
    const result = await expectRuns(`
var bs = make([]uint8, 3)
bs[0] = 'x'
bs[1] = 'y'
bs[2] = 'z'
return string(bs)
`);

    expect(result.value).toBe("xyz");
  });

  test("supports conversions to named slice types with matching underlying type", async () => {
    const result = await expectRuns(`
type Bytes []uint8
var bs = make([]uint8, 2)
named := Bytes(bs)
named[0] = 'o'
named[1] = 'k'
return string(bs), string(named)
`);

    expect(result.values).toEqual(["ok", "ok"]);
  });

  test("slices share backing storage for index assignment and copy", async () => {
    const result = await expectRuns(`
xs := []int{1, 2, 3, 4}
ys := xs[1:3]
ys[0] = 20
n := copy(xs[2:], []int{30, 40})
return xs, ys[0], n
`);

    expect(result.values).toEqual([[1n, 20n, 30n, 40n], 20n, 2n]);
  });

  test("copy from strings writes UTF-8 bytes into byte slices", async () => {
    const result = await expectRuns(`
buf := make([]uint8, 4)
n := copy(buf, "abc")
return n, buf, string(buf[:3])
`);

    expect(result.values).toEqual([3n, [97n, 98n, 99n, 0n], "abc"]);
  });

  test("supports bitwise, shift, unary complement, and compound assignment operators", async () => {
    const result = await expectRuns(`
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

  test("supports if init statements and Go-style builtins new delete clear copy min max print println", async () => {
    const result = await expectRuns(`
xs := []int{1, 2, 3}
ys := make([]int, 3)
n := copy(ys, xs)
clear(ys[1:3])
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
return n, ys[0], ys[1], ys[2], before, len(m), *p
`);

    expect(result.output).toEqual(["v=", "4", " ok\n"]);
    expect(result.values).toEqual([3n, 1n, 0n, 0n, 1n, 0n, 4n]);
  });

  test("supports Go 1.26 new with expression arguments", async () => {
    const result = await expectRuns(`
p := new(123)
x := [2]int{123, 456}
q := new(x)
x[0] = 999
i := 0
next := func() int { i++; return i }
r := new(next())
b := new(i > 10)
return *p, (*q)[0], (*q)[1], *r, i, *b
`);

    expect(result.values).toEqual([123n, 123n, 456n, 1n, 1n, false]);
  });

  test("REPL typechecker treats star expressions as dereferences when the operand is a value", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("x := 42")).diagnostics).toEqual([]);
    expect((await session.evaluate("p := new(x)")).diagnostics).toEqual([]);
    let result = await session.evaluate("*p");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(42n);

    expect((await session.evaluate("arr := [2]int{1, 2}")).diagnostics).toEqual([]);
    expect((await session.evaluate("q := new(arr)")).diagnostics).toEqual([]);
    expect((await session.evaluate("arr[0] = 9")).diagnostics).toEqual([]);
    result = await session.evaluate("(*q)[0]");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("supports buffered channels, close, len cap, and receive ok values", async () => {
    const result = await expectRuns(`
ch := make(chan int, 2)
ch <- 7
ch <- 8
a := <-ch
b, ok := <-ch
close(ch)
c, ok2 := <-ch
return a, b, ok, c, ok2, len(ch), cap(ch)
`);

    expect(result.values).toEqual([7n, 8n, true, 0n, false, 0n, 2n]);
  });

  test("keeps goroutine channel sends parked across REPL evaluations", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("go func() { c <- 1 }()")).diagnostics).toEqual([]);
    expect((await session.evaluate("a := <-c")).diagnostics).toEqual([]);

    const result = await session.evaluate("a");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("runs for range loops inside REPL goroutine function literals", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("go func() { for i := range 5 { c <- i } }()")).diagnostics).toEqual([]);

    for (const expected of [0n, 1n, 2n, 3n, 4n]) {
      const result = await session.evaluate("<-c");
      expect(result.diagnostics).toEqual([]);
      expect(result.value).toBe(expected);
    }
  });

  test("lets later REPL function declarations use earlier top-level variables", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("func f() { for i := range 5 { c <- i } }")).diagnostics).toEqual([]);
    expect((await session.evaluate("go f()")).diagnostics).toEqual([]);

    for (const expected of [0n, 1n, 2n, 3n, 4n]) {
      const result = await session.evaluate("<-c");
      expect(result.diagnostics).toEqual([]);
      expect(result.value).toBe(expected);
    }
  });

  test("compiled REPL functions observe later top-level variable value changes", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("x := 1")).diagnostics).toEqual([]);
    expect((await session.evaluate("func readX() int { return x }")).diagnostics).toEqual([]);

    let result = await session.evaluate("readX()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);

    expect((await session.evaluate("x = 7")).diagnostics).toEqual([]);
    result = await session.evaluate("readX()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(7n);
  });

  test("compatible REPL function redefinition updates existing callers through the function slot", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("func f() int { return 1 }")).diagnostics).toEqual([]);
    expect((await session.evaluate("func g() int { return f() }")).diagnostics).toEqual([]);

    let result = await session.evaluate("g()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);

    expect((await session.evaluate("func f() int { return 2 }")).diagnostics).toEqual([]);
    result = await session.evaluate("g()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(2n);
  });

  test("incompatible REPL function redefinition is rejected and keeps the old slot value", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("func f() int { return 1 }")).diagnostics).toEqual([]);

    const replacement = await session.evaluate("func f(x int) int { return x }");
    expect(replacement.diagnostics).toHaveLength(1);
    expect(replacement.diagnostics[0]?.code).toBe("GOJR_TYPE001");
    expect(replacement.diagnostics[0]?.message).toContain("cannot redeclare f with different signature");

    const result = await session.evaluate("f()");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(1n);
  });

  test("reports REPL deadlocks and keeps the session usable without zombie receives", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("c := make(chan int)")).diagnostics).toEqual([]);
    expect((await session.evaluate("r := 0")).diagnostics).toEqual([]);

    let result = await session.evaluate("r = <-c");
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(result.diagnostics[0]?.message).toContain("receive from channel would block");

    expect((await session.evaluate("go func() { c <- 2 }()")).diagnostics).toEqual([]);

    result = await session.evaluate("r");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(0n);

    expect((await session.evaluate("r = <-c")).diagnostics).toEqual([]);
    result = await session.evaluate("r");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(2n);
  });

  test("supports deterministic select choice from the configured runtime seed", async () => {
    const source = `
ch1 := make(chan int, 2)
ch2 := make(chan int, 2)
ch1 <- 1
ch1 <- 3
ch2 <- 2
ch2 <- 4
a := 0
b := 0
select {
case a = <-ch1:
case a = <-ch2:
}
select {
case b = <-ch1:
case b = <-ch2:
}
return a, b
`;

    const first = await expectRuns(source, { randomSeed: "gojr-select-seed" });
    const second = await expectRuns(source, { randomSeed: "gojr-select-seed" });
    const other = await expectRuns(source, { randomSeed: "gojr-other-select-seed" });

    expect(first.values).toEqual(second.values);
    expect(first.values).toEqual([2n, 1n]);
    expect(other.values).toEqual([1n, 2n]);
  });

  test("supports select default and reports would-block channel operations", async () => {
    const result = await expectRuns(`
ch := make(chan int, 1)
out := 3
select {
case out = <-ch:
default:
  out = 9
}
return out
`);
    expect(result.value).toBe(9n);

    const send = await evaluateSource(`
ch := make(chan int)
ch <- 1
`);
    expect(send.diagnostics).toHaveLength(1);
    expect(send.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(send.diagnostics[0]?.message).toContain("send on channel would block");

    const receive = await evaluateSource(`
ch := make(chan int)
return <-ch
`);
    expect(receive.diagnostics).toHaveLength(1);
    expect(receive.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(receive.diagnostics[0]?.message).toContain("receive from channel would block");
  });

  test("supports nil channel blocking and disabled nil-channel select cases", async () => {
    const selectedDefault = await expectRuns(`
var ch chan int
side := 0
next := func() int {
  side++
  return 1
}
select {
case ch <- next():
  side = 100
case v := <-ch:
  side = 200 + v
default:
  side += 10
}
return side
`);
    expect(selectedDefault.value).toBe(11n);

    const send = await evaluateSource(`
var ch chan int
ch <- 1
`);
    expect(send.diagnostics).toHaveLength(1);
    expect(send.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(send.diagnostics[0]?.message).toContain("send on nil channel would block");

    const receive = await evaluateSource(`
var ch chan int
return <-ch
`);
    expect(receive.diagnostics).toHaveLength(1);
    expect(receive.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(receive.diagnostics[0]?.message).toContain("receive from nil channel would block");

    const blockedSelect = await evaluateSource(`
var ch chan int
select {
case <-ch:
}
`);
    expect(blockedSelect.diagnostics).toHaveLength(1);
    expect(blockedSelect.diagnostics[0]?.code).toBe("GOJR_DEADLOCK001");
    expect(blockedSelect.diagnostics[0]?.message).toContain("select would block");

    const closeNil = await evaluateSource(`
var ch chan int
close(ch)
`);
    expect(closeNil.diagnostics).toHaveLength(1);
    expect(closeNil.diagnostics[0]?.code).toBe("GOJR_PANIC001");
    expect(closeNil.diagnostics[0]?.message).toContain("close of nil channel");
  });

  test("supports methods on named non-struct types and two-value type assertions", async () => {
    const result = await expectRuns(`
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

  test("supports erased generic functions and generic struct declarations", async () => {
    const result = await expectRuns(`
type Box[T any] struct {
  Value T
}

type Number interface {
  ~int | ~float64
}

func Identity[T any](value T) T {
  var copy T = value
  return copy
}

func Pick[T Number](value T) T {
  return value
}

func Add[T Number](left, right T) T {
  return left + right
}

func Unbox[T any](box Box[T]) T {
  return box.Value
}

a := Identity[int](42)
b := Identity[string]("hi")
c := Unbox[int](Box[int]{Value: 7})
d := Pick[float64](2.5)
e := Add[int](3, 4)
f := Add[float64](1.25, 2.5)
return a, b, c, d, e, f
`);

    expect(result.values).toEqual([42n, "hi", 7n, 2.5, 7n, 3.75]);
  });

  test("supports erased generic functions with multiple type arguments", async () => {
    const result = await expectRuns(`
type Pair[A, B any] struct {
  A A
  B B
}

func Make[A, B any](a A, b B) Pair[A, B] {
  return Pair[A, B]{A: a, B: b}
}

p := Make[int, string](7, "seven")
return p.A, p.B
`);

    expect(result.values).toEqual([7n, "seven"]);
  });

  test("REPL checker accepts keyed generic struct literals", async () => {
    const session = new GoJuniorSession();

    expect((await session.evaluate("type Box[T any] struct { Value T }")).diagnostics).toEqual([]);
    expect((await session.evaluate("func Unbox[T any](box Box[T]) T { return box.Value }")).diagnostics).toEqual([]);

    const result = await session.evaluate("Unbox[int](Box[int]{Value: 7})");
    expect(result.diagnostics).toEqual([]);
    expect(result.value).toBe(7n);
  });

  test("supports address-of composite literals and three-index slicing capacity", async () => {
    const result = await expectRuns(`
type Point struct { X int }
p := &Point{X: 7}
xs := make([]int, 5, 8)
ys := xs[1:3:4]
return p.X, len(ys), cap(ys)
`);

    expect(result.values).toEqual([7n, 2n, 3n]);
  });

  test("enforces comparable map keys and supports array and struct comparability", async () => {
    const ok = await expectRuns(`
type Point struct { X int; Y string }
a := [2]int{1, 2}
b := [2]int{1, 2}
p := Point{X: 1, Y: "a"}
q := Point{X: 1, Y: "a"}
return a == b, p == q
`);
    expect(ok.values).toEqual([true, true]);

    const bad = await evaluateSource(`
m := map[[]int]int{}
_ = m
`);
    expect(bad.diagnostics).toHaveLength(1);
    expect(bad.diagnostics[0]?.message).toContain("map key type []int is not comparable");
  });

  test("supports anonymous empty struct zero values, literals, and array range assignment", async () => {
    const result = await expectRuns(`
var xs [3]struct{}
for i := range xs {
  xs[i] = struct{}{}
}
return xs[0] == struct{}{}, xs == [3]struct{}{}
`);

    expect(result.values).toEqual([true, true]);
  });

  test("supports non-empty anonymous struct zero values and pointer selectors", async () => {
    const result = await expectRuns(`
type x2 struct { a, b, c int; d int }
var g1 x2
var g2 struct { a, b, c int; d x2 }

s1 := &g1
s2 := &g2
s1.a = 1
s1.b = 2
s1.c = 3
s1.d = 5
s2.a = 7
s2.b = 11
s2.c = 13
s2.d.a = 17
s2.d.b = 19
s2.d.c = 23
s2.d.d = 20
return s2.d.c, g2.d.c, s1.a + s1.b + s1.c + s1.d + s2.a + s2.b + s2.c + s2.d.a + s2.d.b + s2.d.c + s2.d.d
`);

    expect(result.values).toEqual([23n, 23n, 121n]);
  });

  test("supports blank imports, dot imports, init functions, and range over integers and iterator functions", async () => {
    const result = await expectRuns(`
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

xs := [2]int{}
q := 0
for xs[func() int {
  q++
  return 0
}()] = range [2]int{} {
}

yieldOnce := func(yield func(int) bool) {
  yield(1)
}

for _ = range yieldOnce {
  total++
}

neg := 0
for i := range -1 {
  neg += i
  total += 1000
}

runeCount := 0
var last rune
for i := range 'a' {
  var _ *rune = &i
  last = i
  runeCount++
}

return total, out, q, xs[0], neg, runeCount, last
`);

    expect(result.values).toEqual([9n, "10:a;20:b;", 2n, 1n, 0n, 97n, 96n]);
  });

  test("supports embedded fields, promoted methods, interface embedding, and struct tags", async () => {
    const result = await expectRuns(`
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

  test("supports pointer receivers on named scalar types", async () => {
    const result = await expectRuns(`
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

  test("supports switch init statements and short redeclarations", async () => {
    const result = await expectRuns(`
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

  test("scopes statement init and clause short declarations like Go", async () => {
    const result = await expectRuns(`
total := 0
for i := 0; i < 2; i++ {
  total += i
}
for i := 0; i < 2; i++ {
  total += i * 10
}

if _, ok := map[int]int{}[1]; !ok {
  total += 100
}
if _, ok := map[int]int{1: 1}[1]; ok {
  total += 1000
}

switch x := 1; x {
case 1:
  y := 2
  total += y
case 2:
  y := 3
  total += y
}

ch := make(chan int)
select {
case v := <-ch:
  total += v
default:
  v := 7
  total += v
}
select {
case v := <-ch:
  total += v
default:
  v := 8
  total += v
}

return total
`);

    expect(result.value).toBe(1128n);
  });

  test("rejects invalid short declarations, for posts, fallthrough, and gotos over variables", async () => {
    const noNew = await evaluateSource(`
x := 1
x := 2
`);
    expect(noNew.diagnostics).toHaveLength(1);
    expect(noNew.diagnostics[0]?.message).toContain("short declaration has no new variables");

    const badPost = await evaluateSource(`
for i := 0; i < 2; i := i + 1 {
}
`);
    expect(badPost.diagnostics).toHaveLength(1);
    expect(badPost.diagnostics[0]?.message).toContain("for post");

    const badFallthroughMiddle = await evaluateSource(`
switch 1 {
case 1:
  fallthrough
  fmt.Printf("nope")
default:
}
`);
    expect(badFallthroughMiddle.diagnostics).toHaveLength(1);
    expect(badFallthroughMiddle.diagnostics[0]?.message).toContain("fallthrough must be the final statement");

    const badFallthroughFinal = await evaluateSource(`
switch 1 {
case 1:
  fallthrough
}
`);
    expect(badFallthroughFinal.diagnostics).toHaveLength(1);
    expect(badFallthroughFinal.diagnostics[0]?.message).toContain("final switch clause");

    const badGoto = await evaluateSource(`
goto Done
x := 1
Done:
return x
`);
    expect(badGoto.diagnostics).toHaveLength(1);
    expect(badGoto.diagnostics[0]?.message).toContain("jumps over variable declaration");
  });

  test("decodes Go string and rune escapes", async () => {
    const result = await expectRuns(`
s := "\\x41\\101\\u0042\\U00000043"
r := '\\n'
return s, r
`);

    expect(result.values).toEqual(["AABC", 10n]);
  });

  test("enforces recursive map-key comparability for arrays and structs", async () => {
    const ok = await expectRuns(`
type Key struct { A [2]int; B string }
m := map[Key]int{}
m[Key{A: [2]int{1, 2}, B: "x"}] = 7
return m[Key{A: [2]int{1, 2}, B: "x"}]
`);
    expect(ok.value).toBe(7n);

    const badStruct = await evaluateSource(`
type Bad struct { A []int }
_ = map[Bad]int{}
`);
    expect(badStruct.diagnostics).toHaveLength(1);
    expect(badStruct.diagnostics[0]?.message).toContain("map key type []int is not comparable");
  });

  test("reports unresolved goto labels", async () => {
    const result = await evaluateSource(`
goto Missing
`);

    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]?.message).toContain("unresolved goto label Missing");
  });
});
