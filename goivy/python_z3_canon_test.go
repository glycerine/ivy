package goivy

import (
	"strings"
	"testing"
)

func TestPythonCanonZ3BitVecExprDoesNotTreatVariableAsNumeral(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
from ivy import ivy_solver
z3 = ivy_solver.z3
x = z3.BitVec('x', 2)
s = z3.Solver()
s.add(x == z3.BitVecVal(1, 2))
print(ivy_solver._canon_z3_assertions(s))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python z3 canon for bit-vector expression failed: %v\n%s", err, out)
	}
	got := string(out)
	if strings.Contains(got, "AttributeError") {
		t.Fatalf("bit-vector variable should not be treated as a numeral; got:\n%s", got)
	}
	if !strings.Contains(got, "(c x bv)") || !strings.Contains(got, "(n 1)") {
		t.Fatalf("unexpected bit-vector canon:\n%s", got)
	}
}
