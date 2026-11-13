package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
	"time"
)

func TestAbsProverWithSswSimplification(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManAppendAnd(g, x, q)
	gia.Gia_ManAppendCo(g, out)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(q>>1), 1)
	vec.Vec_IntWriteEntry(classes, int(out>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	Gia_GlaProveCancel(0)
	Gia_GlaProveAbsracted(g, 1, 0)
	defer Gia_GlaProveCancel(0)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if Gia_GlaProveCheck(0) != 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("simplified safe abstraction was not proved")
}
