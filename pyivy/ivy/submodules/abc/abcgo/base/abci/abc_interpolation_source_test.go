package abci

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGiaInterOnePreservesSourcePartitionFailure(t *testing.T) {
	on, off := gia.Gia_ManStart(4), gia.Gia_ManStart(4)
	defer gia.Gia_ManStop(on)
	defer gia.Gia_ManStop(off)
	gia.Gia_ManAppendCo(on, 0)
	gia.Gia_ManAppendCo(off, 1)
	if result := Gia_ManInterOne(on, off, 0); result != nil {
		gia.Gia_ManStop(result)
		t.Fatal("source returns nil when a CNF partition is inconsistent during insertion")
	}
}
