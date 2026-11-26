package abci

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/extra"
)

// aigSignature lists a combinational AIG as AIGER-style literals: each AND's
// two fanins in object order, then each PO's driver. Equal signatures mean
// identical structure, fanin order and polarity.
func aigSignature(t *testing.T, n *abc.Network) []int {
	t.Helper()
	index := map[*abc.Object]int{}
	next := 1
	for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
		index[abc.Abc_NtkPi(n, i)] = next
		next++
	}
	lit := func(o *abc.Object, compl int32) int {
		if o.Kind == abc.ObjectConst1 {
			return 1 ^ int(compl)
		}
		v, ok := index[o]
		if !ok {
			t.Fatalf("object %d used before definition", o.ID)
		}
		return 2*v + int(compl)
	}
	var sig []int
	for i := 0; i < int(abc.Abc_NtkObjNumMax(n)); i++ {
		o := n.Object(i)
		if o == nil || abc.Abc_ObjIsNode(o) == 0 {
			continue
		}
		sig = append(sig, lit(abc.Abc_ObjFanin0(o), abc.Abc_ObjFaninC0(o)), lit(abc.Abc_ObjFanin1(o), abc.Abc_ObjFaninC1(o)))
		index[o] = next
		next++
	}
	for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
		po := abc.Abc_NtkPo(n, i)
		sig = append(sig, lit(abc.Abc_ObjFanin0(po), abc.Abc_ObjFaninC0(po)))
	}
	return sig
}

func readSignature(t *testing.T, path string) []int {
	t.Helper()
	n, err := abcio.ReadAiger(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer abc.Abc_NtkDelete(n)
	return aigSignature(t, n)
}

// The goldens were written by native ABC (with the CBUGS_FIXED.md #1 NPN fix)
// using test_vectors/i10_balance_rewrite_stages.c. Each Go stage is written
// and read back like the native one, so both pass through the same encoding.
func TestI10BalanceRewriteMatchesNativeABC(t *testing.T) {
	extra.Extra_ProgressBarSetMode(false, true)
	defer extra.Extra_ProgressBarSetMode(true, false)
	n, err := abcio.ReadAiger("../../../i10.aig", true)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		balanced := Abc_NtkBalance(n, 0, 0, 1)
		abc.Abc_NtkDelete(n)
		n = balanced
	}
	defer func() { abc.Abc_NtkDelete(n) }()
	dir := t.TempDir()
	check := func(stage string, nodes, levels int32) {
		t.Helper()
		if got, gotLevels := abc.Abc_NtkNodeNum(n), abc.Abc_AigLevel(n); got != nodes || gotLevels != levels {
			t.Fatalf("%s: %d ANDs at depth %d, native ABC has %d at %d", stage, got, gotLevels, nodes, levels)
		}
		path := filepath.Join(dir, stage+".aig")
		abcio.Io_WriteAiger(n, path, 0, 0, 0)
		if !slices.Equal(readSignature(t, path), readSignature(t, filepath.Join("test_vectors", "i10_c_"+stage+".aig"))) {
			t.Fatalf("%s: structure differs from native ABC", stage)
		}
	}
	check("bal2", 2394, 37)
	if Abc_NtkRewrite(n, 0, 0, 0, 0, 0) == 0 {
		t.Fatal("rewrite failed")
	}
	check("rw", 2020, 36)
}
