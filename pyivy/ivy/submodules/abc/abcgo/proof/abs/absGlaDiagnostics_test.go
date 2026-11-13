package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"strings"
	"testing"
)

func TestGlaPrintPseudoInputsInDescendingOrder(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(g, x>>1), 1)
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(g, y>>1), 1)
	p := &Ga2_Man_t{pGia: g, vIds: vec.Vec_IntStart(int(gia.Gia_ManObjNum(g))), vCnfs: make([]*vec.Vec_Int_t, 4), vValues: vec.Vec_IntAllocArrayCopy([]int32{0, x >> 1, y >> 1}, 3)}
	vec.Vec_IntWriteEntry(p.vIds, int(x>>1), 0)
	vec.Vec_IntWriteEntry(p.vIds, int(y>>1), 1)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = write
	Ga2_ManRefinePrintPPis(p)
	os.Stdout = previous
	write.Close()
	data, err := io.ReadAll(read)
	read.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); !strings.Contains(got, "Current PPIs (2): 2 1 ") {
		t.Fatalf("diagnostic %q", got)
	}
}
