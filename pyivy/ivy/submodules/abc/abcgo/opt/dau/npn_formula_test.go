package dau

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNPNFormulaGraphAndTruthParity(t *testing.T) {
	Dau_ParseFormulaTest()
	graph := Dau_ParseFormulaAigTest()
	for pattern := 0; pattern < 32; pattern++ {
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = pattern>>i&1 != 0
		}
		out, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != (uint64(0x037d037d)>>uint(pattern)&1 != 0) {
			t.Fatal("formula graph mismatch", pattern)
		}
	}
	for _, bad := range []string{"", "a&", "~", "a&b^c", "(a", "g"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("invalid formula %q accepted", bad)
				}
			}()
			Dau_ParseFormula(bad)
		}()
	}
}
func TestNPNFormulaOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	sourceRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	source, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauNpn2.c"
int main(void){char s[4096];while(scanf("%4095s",s)==1)printf("%016llx\n",(unsigned long long)Dau_ParseFormula(s));return 0;}`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-ffunction-sections", "-fdata-sections", "-I"+sourceRoot, source, "-Wl,--gc-sections", "-o", exe).CombinedOutput(); err != nil {
		t.Fatalf("C parser oracle%v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(234))
	var formula func(int) string
	formula = func(depth int) string {
		if depth == 0 || rng.Intn(4) == 0 {
			return string(byte('a' + rng.Intn(6)))
		}
		s := "(" + formula(depth-1) + string("&^"[rng.Intn(2)]) + formula(depth-1) + ")"
		if rng.Intn(2) == 0 {
			s = "~" + s
		}
		return s
	}
	formulas := make([]string, 100)
	for i := range formulas {
		formulas[i] = formula(5)
	}
	command := exec.Command(exe)
	command.Stdin = strings.NewReader(strings.Join(formulas, "\n"))
	out, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	reader := strings.NewReader(string(out))
	for _, f := range formulas {
		var expected uint64
		if _, err := fmt.Fscanf(reader, "%x\n", &expected); err != nil {
			t.Fatal(err)
		}
		if Dau_ParseFormula(f) != expected {
			t.Fatal("original C formula truth mismatch", f)
		}
	}
}
