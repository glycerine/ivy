package acb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestSourceAcbKeywordAndOperatorMappings(t *testing.T) {
	keywords := []string{"module", "endmodule", "input", "output", "wire", "buf", "not", "and", "nand", "or", "nor", "xor", "xnor", "_HMUX", "_DC"}
	for i, want := range keywords {
		if got := Acb_Num2Name(int32(i + 1)); got != want {
			t.Fatalf("keyword %d = %q, want %q", i+1, got, want)
		}
	}
	if got := Acb_Num2Name(0); got != "" {
		t.Fatalf("unknown keyword name = %q, want empty string (C NULL)", got)
	}

	operators := []struct {
		keyword int32
		oper    int32
		name    string
	}{
		{6, ABC_OPER_BIT_BUF, "buf"}, {7, ABC_OPER_BIT_INV, "not"},
		{8, ABC_OPER_BIT_AND, "and"}, {9, ABC_OPER_BIT_NAND, "nand"},
		{10, ABC_OPER_BIT_OR, "or"}, {11, ABC_OPER_BIT_NOR, "nor"},
		{12, ABC_OPER_BIT_XOR, "xor"}, {13, ABC_OPER_BIT_NXOR, "xnor"},
		{14, ABC_OPER_BIT_MUX, "mux"}, {15, ABC_OPER_TRI, "_DC"},
	}
	for _, tc := range operators {
		if got := Acb_Type2Oper(tc.keyword); got != tc.oper {
			t.Errorf("keyword %d operator = %d, want %d", tc.keyword, got, tc.oper)
		}
		if got := Acb_Oper2Name(tc.oper); got != tc.name {
			t.Errorf("operator %d name = %q, want %q", tc.oper, got, tc.name)
		}
	}
	if got := Acb_Oper2Name(ABC_OPER_CONST_F); got != "const0" {
		t.Errorf("false constant name = %q", got)
	}
	if got := Acb_Oper2Name(ABC_OPER_CONST_T); got != "const1" {
		t.Errorf("true constant name = %q", got)
	}
	if got := Acb_Oper2Name(ABC_OPER_CONST_X); got != "constX" {
		t.Errorf("unknown constant name = %q", got)
	}
}

func TestSourceRemapOneFunctionPermutesSupportColumns(t *testing.T) {
	supp := Vec_IntAlloc(2)
	Vec_IntPush(supp, 0)
	Vec_IntPush(supp, 1)
	defer Vec_IntFree(supp)
	mapping := Vec_IntAlloc(2)
	Vec_IntPush(mapping, 1)
	Vec_IntPush(mapping, 0)
	defer Vec_IntFree(mapping)
	got := Acb_RemapOneFunction("10 1\n01 0\n", supp, mapping, 2)
	if want := "01 1\n10 0\n"; got != want {
		t.Fatalf("remapped SOP = %q, want %q", got, want)
	}
}

func TestSourceWeightMapPairs(t *testing.T) {
	names := util.Abc_NamStart(8, 16)
	defer util.Abc_NamDeref(names)
	a := util.Abc_NamStrFindOrAdd(names, "a", nil)
	b := util.Abc_NamStrFindOrAdd(names, "b", nil)
	filename := filepath.Join(t.TempDir(), "weights.txt")
	if err := os.WriteFile(filename, []byte("a, 12;\nb (34)\n"), 0600); err != nil {
		t.Fatal(err)
	}
	weights := Acb_ReadWeightMap(filename, names)
	if weights == nil {
		t.Fatal("failed to read weight map")
	}
	if got := Vec_IntEntry(weights, a); got != 12 {
		t.Errorf("weight(a) = %d, want 12", got)
	}
	if got := Vec_IntEntry(weights, b); got != 34 {
		t.Errorf("weight(b) = %d, want 34", got)
	}
}

func TestSourceWireTargetPrefix(t *testing.T) {
	names := util.Abc_NamStart(8, 16)
	defer util.Abc_NamDeref(names)
	for _, tc := range []struct {
		name string
		want int32
	}{{"t_result", 1}, {"t_", 1}, {"target", 0}, {"x_t_result", 0}, {"t", 0}} {
		id := util.Abc_NamStrFindOrAdd(names, tc.name, nil)
		if got := Acb_WireIsTarget(id, names); got != tc.want {
			t.Errorf("target(%q) = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestSourceVecIntPermuteMatchesGlibcRand(t *testing.T) {
	values := Vec_IntStartNatural(5)
	var random acbCRand
	random.seed(0xabc)
	vecIntPermuteWithRand(values, random.next)
	want := []int32{3, 1, 0, 4, 2}
	for i, expected := range want {
		if got := Vec_IntEntry(values, int32(i)); got != expected {
			t.Fatalf("permuted entry %d = %d, want %d", i, got, expected)
		}
	}
}
