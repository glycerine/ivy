package extrab

import (
	"bytes"
	"hash/fnv"
	"strings"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestOriginalCKMapFixture(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	hash := fnv.New64a()
	for table := uint64(0); table < 256; table++ {
		f := autoTruth(dd, table, 3)
		var output bytes.Buffer
		Extra_PrintKMapRelation(&output, dd, f, cudd.Cudd_Not(f), 1, 2, dd.Vars[2:], dd.Vars[:2])
		hash.Write(output.Bytes())
	}
	if got := hash.Sum64(); got != 0xaa69523ca10d5418 {
		t.Fatalf("original C K-map text fixture: %016x", got)
	}
}
func TestKMapSelectedSupportNamesAndWriter(t *testing.T) {
	dd := cudd.NewBDDManager(5)
	f := cudd.Cudd_bddXor(dd, dd.Vars[4], dd.Vars[2])
	var output bytes.Buffer
	Extra_PrintKMap(&output, dd, f, cudd.Cudd_Not(f), 2, []*DdNode{dd.Vars[4], dd.Vars[2]}, 0, []string{"late", "middle"})
	text := output.String()
	if !strings.HasPrefix(text, "Truth table: 6\n") || !strings.Contains(text, " middle \\ late ") {
		t.Fatal("selected support diagnostic/names", text)
	}
	output.Reset()
	Extra_PrintKMap(&output, dd, f, cudd.Cudd_Not(f), 1, nil, 2, nil)
	if !strings.HasPrefix(output.String(), "Truth table: 6\n") {
		t.Fatal("automatic support", output.String())
	}
	output.Reset()
	Extra_PrintKMap(&output, dd, dd.One, dd.LogicZero, 0, nil, 0, nil)
	if output.String() != "Function is constant 1.\n" {
		t.Fatal("constant writer", output.String())
	}
}
func TestKMapDontCaresOverlapAndInvalidSupport(t *testing.T) {
	dd := cudd.NewBDDManager(2)
	var output bytes.Buffer
	Extra_PrintKMapRelation(&output, dd, dd.Vars[0], dd.LogicZero, 1, 1, dd.Vars[1:], dd.Vars[:1])
	if !strings.Contains(output.String(), " - |") || !strings.Contains(output.String(), " 1 |") {
		t.Fatal("don't-care cells", output.String())
	}
	output.Reset()
	Extra_PrintKMapRelation(&output, dd, dd.One, dd.One, 1, 1, dd.Vars[1:], dd.Vars[:1])
	if !strings.Contains(output.String(), "overlap") {
		t.Fatal("overlap check")
	}
	output.Reset()
	Extra_PrintKMap(&output, dd, dd.Vars[1], cudd.Cudd_Not(dd.Vars[1]), 1, dd.Vars[:1], 0, nil)
	if !strings.Contains(output.String(), "incomplete support") {
		t.Fatal("incomplete support accepted")
	}
}
func TestGrayCodeRoundTrip(t *testing.T) {
	for binary := 0; binary < 4096; binary++ {
		if BinCode(GrayCode(binary)) != binary {
			t.Fatal(binary)
		}
	}
	if BinCode(-1) != -1 {
		t.Fatal("negative inverse gray code")
	}
}
func TestKMapRelationDimensionsDoNotOverflow(t *testing.T) {
	dd := cudd.NewBDDManager(1)
	var output strings.Builder
	Extra_PrintKMapRelation(&output, dd, dd.Vars[0], cudd.Cudd_Not(dd.Vars[0]), int(^uint(0)>>1), 1, nil, nil)
	if !strings.Contains(output.String(), "more than") {
		t.Fatal("overflowing dimensions not rejected", output.String())
	}
}
