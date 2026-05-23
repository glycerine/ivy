package ivy2cpp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBVWidth128GeneratesInt128TypedefAndSolverConversions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[128]
individual x : word
action set(v:word) returns(out:word) = {
    x := 300;
    x := x + v;
    out := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "wide128"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		`#include "ivy_wide_uint.hpp"`,
		"unsigned __int128 x;",
		`x = (ivy_uint128_from_string("300") & ivy_uint128_mask(128));`,
		"template <> void __from_solver<unsigned __int128>(gen &g, const z3::expr &expr, unsigned __int128 &out)",
		"template <> z3::expr __to_solver<unsigned __int128>(gen &g, const char *sort_name, const unsigned __int128 &value)",
		"template <> z3::expr __to_solver<unsigned __int128>(gen &g, const z3::expr &expr, const unsigned __int128 &value)",
		"template <> void __randomize<unsigned __int128>(gen &g, const z3::expr &expr, const std::string &range)",
		"return static_cast<unsigned __int128>(ivy_uint128_random(128));",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in bv[128] output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestBVWidth256GeneratesHelperTypeAndGeneralConversions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[256]
individual x : word
action set(v:word) returns(out:word) = {
    x := 300;
    x := x * v;
    out := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "wide256"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		`#include "ivy_wide_uint.hpp"`,
		"typedef ivy_uint<256> word;",
		"word x;",
		`x = (ivy_uint<256>("300") & ivy_uint<256>::mask());`,
		"template <> void __from_solver<ivy_uint<256>>(gen &g, const z3::expr &v, ivy_uint<256> &res)",
		"template <> z3::expr __to_solver<ivy_uint<256>>(gen &g, const char *sort_name, const ivy_uint<256> &val)",
		"template <> z3::expr __to_solver<ivy_uint<256>>(gen &g, const z3::expr &v, const ivy_uint<256> &val)",
		"template <> void __randomize<ivy_uint<256>>(gen &g, const z3::expr &apply_expr, const std::string &sort_name)",
		"template <> ivy_uint<256> _arg<ivy_uint<256>>(std::vector<ivy_value> &args, unsigned idx, long long bound)",
		"template <> void __ser<ivy_uint<256>>(ivy_ser &res, const ivy_uint<256> &inp)",
		"template <> void __deser<ivy_uint<256>>(ivy_deser &inp, ivy_uint<256> &res)",
		"return ivy_uint<256>::random();",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in bv[256] output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestBVWidth513IsHandledBySameHelperPath(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type digest
interpret digest -> bv[513]
individual x : digest
action set(v:digest) = {
    x := v;
    x := bvnot(x)
}
export set
`)
	out, err := Generate(mod, Config{ClassName: "wide513"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		"typedef ivy_uint<513> digest;",
		"digest x;",
		"ivy_uint<513>::mask()",
		"template <> ivy_uint<513> _arg<ivy_uint<513>>(std::vector<ivy_value> &args, unsigned idx, long long bound)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in bv[513] output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestBVWidth128OnMSVCRejected(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[128]
individual x : word
`)
	out, err := Generate(mod, Config{ClassName: "widecl", Compiler: "cl"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	_, err = BuildPlanFor(out, t.TempDir(), Config{Compiler: "cl", Target: "impl"})
	if err == nil {
		t.Fatal("BuildPlanFor compiler=cl unexpectedly accepted bv[128]")
	}
	if !strings.Contains(err.Error(), "cannot build bv widths greater than 64") || !strings.Contains(err.Error(), "MSVC") {
		t.Fatalf("MSVC wide-bv diagnostic did not explain the limitation: %v", err)
	}
}

func TestBVWidth256SupportHeaderExposesGeneralOperators(t *testing.T) {
	path := filepath.Join(repoRoot(t), "include2cpp", "ivy_wide_uint.hpp")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(data)
	for _, want := range []string{
		"friend ivy_uint operator+(ivy_uint a, const ivy_uint &b)",
		"friend ivy_uint operator-(ivy_uint a, const ivy_uint &b)",
		"friend ivy_uint operator*(const ivy_uint &a, const ivy_uint &b)",
		"friend ivy_uint operator&(ivy_uint a, const ivy_uint &b)",
		"friend ivy_uint operator|(ivy_uint a, const ivy_uint &b)",
		"friend ivy_uint operator^(ivy_uint a, const ivy_uint &b)",
		"friend ivy_uint operator<<(const ivy_uint &a, unsigned shift)",
		"friend ivy_uint operator>>(const ivy_uint &a, unsigned shift)",
		"friend bool operator==(const ivy_uint &a, const ivy_uint &b)",
		"friend bool operator<(const ivy_uint &a, const ivy_uint &b)",
		"friend bool operator<=(const ivy_uint &a, const ivy_uint &b)",
		"friend bool operator>=(const ivy_uint &a, const ivy_uint &b)",
		"friend bool operator>(const ivy_uint &a, const ivy_uint &b)",
		"friend std::ostream &operator<<(std::ostream &out, const ivy_uint &value)",
		"friend std::istream &operator>>(std::istream &in, ivy_uint &value)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("wide BV helper missing operator/support %q:\n%s", want, text)
		}
	}
}
