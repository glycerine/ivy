package gia

import (
	"reflect"
	"testing"
)

func TestNfCutStorageAndDominance(t *testing.T) {
	makeCut := func(flow float32, leaves ...int32) *Nf_Cut_t {
		c := &Nf_Cut_t{Flow: flow, iFunc: 12, nLeaves: uint32(len(leaves)), Sign: Nf_CutGetSign(leaves)}
		copy(c.pLeaves[:], leaves)
		return c
	}
	large, small, other := makeCut(3, 1, 2, 3), makeCut(1, 1, 2), makeCut(2, 4, 5)
	cuts := []*Nf_Cut_t{large, other, small}
	if Nf_SetLastCutIsContained(cuts, 2) != 0 {
		t.Fatal("smaller cut rejected")
	}
	count := Nf_SetAddCut(cuts, 2, 5)
	if count != 2 || cuts[0] != small || cuts[1] != other {
		t.Fatal("dominance ordering", count, cuts)
	}
	storage := newNfCutStorage()
	other.Useless = 1
	h := storage.save(cuts[:count], true)
	row := storage.at(h)
	if row[0] != 1 || Nf_CutSize(row[1:]) != 2 || Nf_CutFunc(row[1:]) != 12 || !reflect.DeepEqual(Nf_CutLeaves(row[1:]), []int32{1, 2}) {
		t.Fatal("cut record", row[:4])
	}
	storage.next = 0xffff
	if next := storage.save(cuts[:count], false); next != 1<<16 || storage.at(next)[0] != 2 {
		t.Fatal("page rollover", next)
	}
}
