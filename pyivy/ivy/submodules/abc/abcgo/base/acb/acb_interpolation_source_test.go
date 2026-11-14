package acb

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestInterOneRejectsInconsistentPartitionAtInsertion(t *testing.T) {
	on, off := gia.Gia_ManStart(4), gia.Gia_ManStart(4)
	defer gia.Gia_ManStop(on)
	defer gia.Gia_ManStop(off)
	gia.Gia_ManAppendCo(on, 0)
	gia.Gia_ManAppendCo(off, 1)
	// abcDar.c -> aigInter.c exits when adding A's clauses is inconsistent.
	if got := Gia_ManInterOne(on, off, 0); got != nil {
		gia.Gia_ManStop(got)
		t.Fatal("interpolation substituted a graph for the source insertion failure")
	}
}
