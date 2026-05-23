package ivy2cpp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func generateBVOperatorFixture(t *testing.T, width int, op string, arity int) *Output {
	t.Helper()
	src := fmt.Sprintf(`#lang ivy1.7
type word
interpret word -> bv[%d]
individual x : word
individual y : word
individual z : word
action step = {}
export step
`, width)
	mod := compileIvySource(t, src)
	word := mod.Sig.Sorts.Get("word")
	sorts := make([]goivy.Sort, 0, arity+1)
	args := make([]goivy.Expr, 0, arity)
	for i := 0; i < arity; i++ {
		sorts = append(sorts, word)
		name := "x"
		if i == 1 {
			name = "y"
		}
		args = append(args, goivy.NewConst(name, word))
	}
	sorts = append(sorts, word)
	fnSort, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	rhs := goivy.MustApply(goivy.NewConst(op, fnSort), args...)
	mod.Actions.Set("step", goivy.NewAssignAction(goivy.NewConst("z", word), rhs))
	out, err := Generate(mod, Config{ClassName: fmt.Sprintf("bvop%d", width)})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNoUnsupportedCPP(t, out)
	return out
}

func TestBVXorEmitsXor(t *testing.T) {
	out := generateBVOperatorFixture(t, 8, "bvxor", 2)
	if !strings.Contains(out.Impl, "z = (((x ^ y)) & 255);") {
		t.Fatalf("bvxor should lower to masked C++ xor:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestBVShiftLeftMasksWidthAndGuardsLargeShift(t *testing.T) {
	out := generateBVOperatorFixture(t, 8, "bvshl", 2)
	for _, want := range []string{
		"z = ([&]() -> unsigned",
		"unsigned __s = static_cast<unsigned>(__y);",
		"if (__s >= 8) return static_cast<unsigned>(0);",
		"return ((__x << __s) & 255);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("bvshl output missing %q:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestBVLogicalShiftRightUsesUnsignedGuard(t *testing.T) {
	out := generateBVOperatorFixture(t, 64, "bvlshr", 2)
	for _, want := range []string{
		"4294967295ULL",
		"if (__s >= 64) return static_cast<unsigned long long>(0);",
		"return ((__x >> __s) & 18446744073709551615ULL);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("bvlshr output missing %q:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestBVArithmeticShiftRightSignExtends(t *testing.T) {
	out := generateBVOperatorFixture(t, 8, "bvashr", 2)
	for _, want := range []string{
		"bool __sign = ((__x & (1U << 7)) != __zero);",
		"if (__s >= 8) return __sign ? __mask : __zero;",
		"__res = (__res | ((__mask << (8 - __s)) & __mask));",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("bvashr output missing %q:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestBVNegEmitsTwosComplement(t *testing.T) {
	out := generateBVOperatorFixture(t, 8, "bvneg", 1)
	if !strings.Contains(out.Impl, "z = (((-(x))) & 255);") {
		t.Fatalf("bvneg should lower to masked two's-complement negation:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestSymbolicShiftAliasesLowerAsBVShifts(t *testing.T) {
	left := generateBVOperatorFixture(t, 8, "<<", 2)
	if !strings.Contains(left.Impl, "return ((__x << __s) & 255);") {
		t.Fatalf("<< should lower as BV shift-left:\n%s", left.Impl)
	}
	right := generateBVOperatorFixture(t, 8, ">>", 2)
	if !strings.Contains(right.Impl, "return ((__x >> __s) & 255);") {
		t.Fatalf(">> should lower as BV logical shift-right:\n%s", right.Impl)
	}
	compileGeneratedCPP(t, left)
	compileGeneratedCPP(t, right)
}

func TestBVWideShiftUsesGeneralShiftAmountHelper(t *testing.T) {
	out := generateBVOperatorFixture(t, 256, "bvashr", 2)
	for _, want := range []string{
		`#include "ivy_wide_uint.hpp"`,
		"ivy_bv_shift_amount(__y)",
		"ivy_uint<256>(1) << 255",
		"ivy_uint<256>::mask()",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("wide bvashr output missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestUnknownBVOperatorReportsUnsupported(t *testing.T) {
	src := `#lang ivy1.7
type word
interpret word -> bv[8]
individual x : word
individual y : word
individual z : word
action step = {}
`
	mod := compileIvySource(t, src)
	word := mod.Sig.Sorts.Get("word")
	fnSort, err := goivy.NewFunctionSort(word, word, word)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	rhs := goivy.MustApply(goivy.NewConst("bvfoo", fnSort), goivy.NewConst("x", word), goivy.NewConst("y", word))
	mod.Actions.Set("step", goivy.NewAssignAction(goivy.NewConst("z", word), rhs))
	_, err = Generate(mod, Config{ClassName: "bvunknown"})
	if err == nil || !strings.Contains(err.Error(), "unknown BV operator bvfoo/2") {
		t.Fatalf("Generate error = %v, want unknown BV operator diagnostic", err)
	}
}

func TestStringOpAddEmitsPlus(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
individual a : text
individual b : text
individual c : text
action step = {
    c := a + b
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "stringadd"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "c = (a + b);") {
		t.Fatalf("string + should remain a C++ std::string plus expression:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestBVOperatorTableCoversPythonAndZ3Names(t *testing.T) {
	path := filepath.Join(repoRoot(t), "goivy", "ivy2cpp", "bv_expr.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(data)
	for _, op := range []string{
		"concat", "bvand", "bvor", "bvxor", "bvnot", "bvneg",
		"bvshl", "bvlshr", "bvashr", "<<", ">>",
		"bvadd", "bvsub", "bvmul", "bvudiv", "bvurem",
		"cast",
	} {
		if !strings.Contains(src, strconvQuote(op)) {
			t.Fatalf("bv_expr.go does not mention BV operator %q", op)
		}
	}
}

func strconvQuote(s string) string {
	return fmt.Sprintf("%q", s)
}
