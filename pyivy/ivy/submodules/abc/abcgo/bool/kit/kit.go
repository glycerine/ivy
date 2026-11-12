// Package kit implements ABC's Boolean computation kit: truth tables, covers,
// decomposition graphs, and local BDDs. Bit zero is the all-zero assignment.
package kit

type Kit_Sop_t struct {
	nLits  int32
	failed bool
	pCubes []uint32
}
type Kit_Dsd_t = uint32

const (
	KIT_DSD_NONE Kit_Dsd_t = iota
	KIT_DSD_CONST1
	KIT_DSD_VAR
	KIT_DSD_AND
	KIT_DSD_XOR
	KIT_DSD_PRIME
)

type Kit_DsdObj_t struct {
	Id, Type, fMark, Offset, nRefs, nFans uint32
	pFans                                 []uint16
	truth                                 []uint32
}
type Kit_DsdNtk_t struct {
	nVars, Root  uint16
	pMem, pSupps []uint32
	pNodes       []*Kit_DsdObj_t
}
type Kit_DsdMan_t struct {
	nVars              int32
	vTtElems, vTtNodes [][]uint32
	dd                 *CloudManager
	vTtBdds            []*CloudNode
	vNodes             []int32
}

func Kit_DsdObjOffset(nFans int32) uint32 { return uint32((nFans + 1) / 2) }
func Kit_DsdObjTruth(obj *Kit_DsdObj_t) []uint32 {
	if obj.Type == KIT_DSD_PRIME {
		return obj.truth
	}
	return nil
}
func Kit_DsdNtkObjNum(ntk *Kit_DsdNtk_t) int32 { return int32(ntk.nVars) + int32(len(ntk.pNodes)) }
func Kit_DsdNtkObj(ntk *Kit_DsdNtk_t, id int32) *Kit_DsdObj_t {
	if id < 0 || id >= Kit_DsdNtkObjNum(ntk) {
		panic("kit: DSD object index out of bounds")
	}
	if id < int32(ntk.nVars) {
		return nil
	}
	return ntk.pNodes[id-int32(ntk.nVars)]
}
func Kit_DsdNtkRoot(ntk *Kit_DsdNtk_t) *Kit_DsdObj_t { return Kit_DsdNtkObj(ntk, int32(ntk.Root>>1)) }
func Kit_DsdLitIsLeaf(ntk *Kit_DsdNtk_t, lit int32) int32 {
	id := lit >> 1
	if id < 0 || id >= Kit_DsdNtkObjNum(ntk) {
		panic("kit: DSD literal out of bounds")
	}
	if id < int32(ntk.nVars) {
		return 1
	}
	return 0
}
func Kit_DsdLitSupport(ntk *Kit_DsdNtk_t, lit int32) uint32 {
	if len(ntk.pSupps) == 0 {
		return 0
	}
	id := lit >> 1
	if Kit_DsdLitIsLeaf(ntk, lit) != 0 {
		return 1 << uint(id)
	}
	return ntk.pSupps[id-int32(ntk.nVars)]
}

const KIT_INFINITY = 100000000

func Kit_CubeHasLit(c uint32, i int32) int32 {
	if c&(1<<uint(i)) != 0 {
		return 1
	}
	return 0
}
func Kit_CubeSetLit(c uint32, i int32) uint32 { return c | 1<<uint(i) }
func Kit_CubeXorLit(c uint32, i int32) uint32 { return c ^ 1<<uint(i) }
func Kit_CubeRemLit(c uint32, i int32) uint32 { return c &^ (1 << uint(i)) }
func Kit_CubeContains(large, small uint32) int32 {
	if large&small == small {
		return 1
	}
	return 0
}
func Kit_CubeSharp(c, mask uint32) uint32      { return c &^ mask }
func Kit_CubeMask(n int32) uint32              { return Kit_BitMask(n) }
func Kit_CubeIsMarked(c uint32) int32          { return Kit_CubeHasLit(c, 31) }
func Kit_CubeMark(c uint32) uint32             { return Kit_CubeSetLit(c, 31) }
func Kit_CubeUnmark(c uint32) uint32           { return Kit_CubeRemLit(c, 31) }
func Kit_SopCubeNum(s *Kit_Sop_t) int32        { return s.cubeNum() }
func Kit_SopCube(s *Kit_Sop_t, i int32) uint32 { return s.pCubes[i] }
func Kit_SopShrink(s *Kit_Sop_t, n int32) {
	if n < 0 || n > s.cubeNum() {
		panic("kit: invalid cover shrink")
	}
	s.pCubes = s.pCubes[:n]
}
func Kit_SopPushCube(s *Kit_Sop_t, c uint32)           { s.pCubes = append(s.pCubes[:s.cubeNum()], c) }
func Kit_SopWriteCube(s *Kit_Sop_t, c uint32, i int32) { s.pCubes[i] = c }

// cubeNum preserves the public -1 allocation-failure status without caching a
// second length that can disagree with the authoritative cube slice.
func (s *Kit_Sop_t) cubeNum() int32 {
	if s.failed {
		return -1
	}
	return int32(len(s.pCubes))
}
