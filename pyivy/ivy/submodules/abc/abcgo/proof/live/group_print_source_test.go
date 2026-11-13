package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceGroupPrintersAppendNamesAndEmptyGroup(t *testing.T) {
	p := aig.Aig_ManStart(10)
	for i := 0; i < 4; i++ {
		aig.Aig_ObjCreateCo(p, 0)
	}
	n := &Network{Graph: p, OutputNames: []string{"p_hint_one", "hint_two", "p_csLevel1Stabil_x", "csLevel1Stabil_y"}}
	path := filepath.Join(t.TempDir(), "groups.txt")
	printAllIntVectors([][]int32{{0, 1}}, n, path)
	printAllIntVectorsStabil([][]int32{{2, 3}}, n, path)
	printAllIntVectors([][]int32{{}}, n, path)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "( hint_one || hint_two )\n( csLevel1Stabil_x || csLevel1Stabil_y )\n( "
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
