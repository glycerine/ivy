package acec

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Truth table operations from src/proof/acec/acec2Mult.c. The configured
// maximum cut size is six, so each truth table occupies one uint64.
func Sdb_CutTruth(p *Sdb_Sto_t, c *Sdb_Cut_t) uint64 {
	return vec.Vec_MemReadEntry(p.vTtMem, int(c.iFunc>>1))[0]
}
func Sdb_CutComputeTruth6(p *Sdb_Sto_t, a, b *Sdb_Cut_t, complA, complB int32, out *Sdb_Cut_t, isXor int32) int32 {
	old := out.nLeaves
	t0, t1 := Sdb_CutTruth(p, a), Sdb_CutTruth(p, b)
	if (a.iFunc&1)^complA != 0 {
		t0 = ^t0
	}
	if (b.iFunc&1)^complB != 0 {
		t1 = ^t1
	}
	t0 = util.Abc_Tt6Expand(t0, a.pLeaves[:], int32(a.nLeaves), out.pLeaves[:], int32(out.nLeaves))
	t1 = util.Abc_Tt6Expand(t1, b.pLeaves[:], int32(b.nLeaves), out.pLeaves[:], int32(out.nLeaves))
	t := t0 & t1
	if isXor != 0 {
		t = t0 ^ t1
	}
	compl := int32(t & 1)
	if compl != 0 {
		t = ^t
	}
	if p.fTruthMin != 0 {
		out.nLeaves = uint32(util.Abc_Tt6MinBase(&t, out.pLeaves[:], int32(out.nLeaves)))
	}
	if t&1 != 0 || out.nLeaves > old {
		panic("acec: invalid minimized truth table")
	}
	id := vec.Vec_MemHashInsert(p.vTtMem, []uint64{t})
	out.iFunc = int32(id)*2 + compl
	if out.nLeaves < old {
		return 1
	}
	return 0
}
func Sdb_CutComputeTruth(p *Sdb_Sto_t, a, b *Sdb_Cut_t, complA, complB int32, out *Sdb_Cut_t, isXor int32) int32 {
	if p.nCutSize > 6 {
		panic("acec: cut size exceeds configured maximum")
	}
	return Sdb_CutComputeTruth6(p, a, b, complA, complB, out, isXor)
}
