package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func strashTestNetwork(t *testing.T) *abc.Network {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	man := n.FunctionManager.(*abc.Abc_Aig_t)
	ab := abc.Abc_AigAnd(man, a, abc.Abc_ObjNot(b))
	root := abc.Abc_AigAnd(man, ab, c)
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("out")
	abc.Abc_ObjAddFanin(po, abc.Abc_ObjNot(root))
	abc.Abc_NtkOrderCisCos(n)
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("invalid test network")
	}
	return n
}

func TestRestrashPreservesComplementedTruthTableAndCleanup(t *testing.T) {
	original := strashTestNetwork(t)
	for _, cleanup := range []int32{0, 1} {
		got := Abc_NtkRestrash(original, cleanup)
		if got == nil {
			t.Fatalf("Abc_NtkRestrash(cleanup=%d) failed", cleanup)
		}
		if abc.Abc_NtkCheck(got) == 0 {
			t.Fatalf("restrashed network invalid with cleanup=%d", cleanup)
		}
		for bits := 0; bits < 8; bits++ {
			inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want, err := original.Evaluate(original.Outputs()[0].Fanins()[0], mapInputs(original.Inputs(), inputs))
			if err != nil {
				t.Fatal(err)
			}
			value, err := got.Evaluate(got.Outputs()[0].Fanins()[0], mapInputs(got.Inputs(), inputs))
			if err != nil {
				t.Fatal(err)
			}
			if value != want {
				t.Errorf("cleanup=%d inputs=%03b: got %v want %v", cleanup, bits, value, want)
			}
		}
		abc.Abc_NtkDelete(got)
	}
}

func TestTopmostAndBottommostPreserveCutOrderingAndComplement(t *testing.T) {
	n := strashTestNetwork(t)
	top := Abc_NtkTopmost(n, 1)
	if top == nil {
		t.Fatal("Abc_NtkTopmost failed")
	}
	if got := len(top.Inputs()); got != 2 {
		t.Fatalf("top cut input count=%d, want 2", got)
	}
	for mask := 0; mask < 4; mask++ {
		want := !(mask&1 != 0 && mask&2 != 0)
		got, err := top.Evaluate(top.Outputs()[0].Fanins()[0], mapInputs(top.Inputs(), []bool{mask&1 != 0, mask&2 != 0}))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("top cut mask=%02b: got %v want %v", mask, got, want)
		}
	}
	abc.Abc_NtkDelete(top)

	bottom := Abc_NtkBottommost(n, 1)
	if bottom == nil {
		t.Fatal("Abc_NtkBottommost failed")
	}
	if got := len(bottom.Inputs()); got != 3 {
		t.Fatalf("bottom cut input count=%d, want 3", got)
	}
	if got := len(bottom.Outputs()); got != 1 {
		t.Fatalf("bottom cut output count=%d, want 1", got)
	}
	for mask := 0; mask < 8; mask++ {
		want := mask&1 != 0 && mask&2 == 0
		got, err := bottom.Evaluate(bottom.Outputs()[0].Fanins()[0], mapInputs(bottom.Inputs(), []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("bottom cut mask=%03b: got %v want %v", mask, got, want)
		}
	}
	abc.Abc_NtkDelete(bottom)
}

func TestWriteAigUsesSourceLiteralOrderingAndComplementBits(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	and := abc.Abc_AigAnd(n.FunctionManager.(*abc.Abc_Aig_t), a, abc.Abc_ObjNot(b))
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("z")
	abc.Abc_ObjAddFanin(po, abc.Abc_ObjNot(and))
	abc.Abc_NtkOrderCisCos(n)
	path := filepath.Join(t.TempDir(), "network.aig.go")
	Abc_NtkWriteAig(n, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{
		`    "`,
		`{ "a","b",NULL },`,
		`{ "z",NULL },`,
		`{ 2,0 },`,
		`{ 5,0 },`,
		`{ 7,0 },`,
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("AIG serialization missing %q:\n%s", expected, text)
		}
	}
}

func mapInputs(objects []*abc.Object, values []bool) map[*abc.Object]bool {
	result := make(map[*abc.Object]bool, len(objects))
	for i, object := range objects {
		result[object] = values[i]
	}
	return result
}
