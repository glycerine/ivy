package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestOrdinaryFunctionBodies(t *testing.T) {
	source := `void f(int n)
{
 for (i = 0; i < n; i++)
  use(i);
 if (n) // keep condition comment
 {
  use(n);
 }
}
static int g(void) { return 1; }
inline int h(void) { return 2; }
extern int k(void) { return 3; }
`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Functions) != 4 {
		t.Fatalf("functions: %d\n%s", len(r.Functions), r.Draft)
	}
	for _, want := range []string{"func f(n int32) {", "for i = 0; i < n; i++ {", "if n != 0 { // keep condition comment", "func g() int32 {", "func h() int32 {", "func k() int32 {"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestInt2BmcOrdinaryFunctions(t *testing.T) {
	source, err := os.ReadFile("../src/proof/int2/int2Bmc.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("int2Bmc.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Functions) != 8 {
		t.Fatalf("recognized %d functions, want 8", len(r.Functions))
	}
	tokens, err := Lex(r.Draft)
	if err != nil {
		t.Fatal(err)
	}
	var significant []Token
	for _, tok := range tokens {
		if tok.Kind != "space" && tok.Kind != "comment" {
			significant = append(significant, tok)
		}
	}
	pairs, err := MatchDelimiters(significant)
	if err != nil {
		t.Fatal(err)
	}
	for i, tok := range significant {
		if tok.Text == "if" {
			brace := ifBodyIndex(significant, pairs, i)
			if brace < 0 {
				t.Fatalf("unbraced if at %d", tok.Start)
			}
			if strings.Contains(r.Draft[significant[brace-1].End:significant[brace].Start], "\n") {
				t.Fatalf("brace on following line at %d", tok.Start)
			}
		}
	}
	for _, want := range []string{"var i, Lit, RetValue int32;", "var pSat *sat_solver;", "func Int2_ManCreateFrames(p *Int2_Man_t, iFrame int32, vPrefCos *Vec_Int_t) {", "for  i = Vec_PtrSize(p.vMapFrames); i <= iFrame; i++  {", "if f == 0 {"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestJoinExistingControlBraces(t *testing.T) {
	source := "if (x)\n{\n for (i=0;i<n;i++)\n { use(i); }\n} else\n{ use(x); }"
	got, err := braceStatements(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"if (x) {", "for (i=0;i<n;i++) {", "else {"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q: %s", want, got)
		}
	}
}

func TestMacroInsideOrdinaryFunction(t *testing.T) {
	r, err := Analyze("x.c", "void f(void) {\n#define STEP(x) use(x)\n STEP(1);\n}\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "func f() {") || strings.Contains(r.Draft, "func STEP(") || !strings.Contains(r.Draft, "use(1);") {
		t.Fatal(r.Draft)
	}
}

func TestIfInsideIteratorMacroBlock(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) {
  Each(p, i) {
   if (i) continue;
   if (i == 2) return;
  }
 }`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"if i { continue; }", "if i == 2 { return; }"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q: %s", want, r.Draft)
		}
	}
}

func TestIteratorTraversalDoesNotDuplicateEdits(t *testing.T) {
	source, err := os.ReadFile("../src/proof/acec/acecPo.c")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Analyze("acecPo.c", string(source)); err != nil {
		t.Fatal(err)
	}
}
