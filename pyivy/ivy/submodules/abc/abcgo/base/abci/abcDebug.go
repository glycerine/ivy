package abci

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	abcIO "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/ret"
)

// glibcRand reproduces the Linux/glibc rand() TYPE_3 sequence used by ABC's
// srand/rand pair. ABC's deterministic debug ordering depends on that sequence.
// Other libc implementations can use a different rand algorithm.
type glibcRand struct {
	state [31]uint32
	index uint64
}

func newGlibcRand(seed uint32) *glibcRand {
	if seed == 0 {
		seed = 1
	}
	seq := make([]uint32, 344)
	seq[0] = seed
	for i := 1; i < 31; i++ {
		// glibc initializes the additive generator with Park-Miller values.
		seq[i] = uint32((uint64(16807) * uint64(seq[i-1])) % 2147483647)
	}
	seq[31], seq[32], seq[33] = seq[0], seq[1], seq[2]
	for i := 34; i < len(seq); i++ {
		seq[i] = seq[i-31] + seq[i-3]
	}
	r := &glibcRand{index: 344}
	// Keep each value in the slot addressed by its absolute sequence index;
	// the recurrence below uses i modulo 31 just like glibc's front/rear
	// pointers. A plain copy loses that three-slot offset (313 % 31).
	for i := 313; i < 344; i++ {
		r.state[i%31] = seq[i]
	}
	return r
}

func (r *glibcRand) next() uint32 {
	i := r.index
	value := r.state[(i-31)%31] + r.state[(i-3)%31]
	r.state[i%31] = value
	r.index++
	return value >> 1
}

// Abc_NtkAutoDebug repeatedly simplifies a network while a caller supplied
// predicate continues to demonstrate a bug. The original network is never
// modified; the internal duplicate is always deleted before return.
func Abc_NtkAutoDebug(pNtk *Abc_Ntk_t, pFuncError func(*Abc_Ntk_t) int32) {
	assert(abc.Abc_NtkIsLogic(pNtk) != 0, "Abc_NtkIsLogic(pNtk)")
	rng := newGlibcRand(0x123123)
	pNtk = abc.Abc_NtkDup(pNtk)
	if pFuncError(pNtk) == 0 {
		fmt.Printf("The original network does not cause the bug. Quitting.\n")
		abc.Abc_NtkDelete(pNtk)
		return
	}
	for nIter := int32(0); ; nIter++ {
		clk := util.Abc_Clock()
		nSteps := 2 * Abc_NtkCountFaninsTotal(pNtk)
		randNum := int32(1) ^ int32(rng.next())
		i := int32(0)
		for ; i < nSteps; i++ {
			modNum := (i + randNum) % nSteps
			pNtkMod := Abc_NtkAutoDebugModify(pNtk, modNum/2, modNum%2)
			_ = abcIO.Io_WriteBlifLogic(pNtk, "bug_temp.blif", 1)
			if pFuncError(pNtkMod) != 0 {
				abc.Abc_NtkDelete(pNtk)
				pNtk = pNtkMod
				break
			}
			abc.Abc_NtkDelete(pNtkMod)
		}
		fmt.Printf("Iter %6d : Latches = %6d. Nodes = %6d. Steps = %6d. Error step = %3d.  ", nIter, abc.Abc_NtkLatchNum(pNtk), abc.Abc_NtkNodeNum(pNtk), nSteps, i)
		util.Abc_Print(1, "%s =%9.2f sec\n", "Time", float64(util.Abc_Clock()-clk)/float64(time.Second))
		if i == nSteps {
			break
		}
	}
	const filename = "bug_found.blif"
	_ = abcIO.Io_WriteBlifLogic(pNtk, filename, 1)
	fmt.Printf("Final network written into file \"%s\". ", filename)
	util.Abc_Print(1, "%s =%9.2f sec\n", "Total time", float64(util.Abc_Clock())/float64(time.Second))
	abc.Abc_NtkDelete(pNtk)
}

// Abc_NtkAutoDebugModify replaces the selected fanin with a constant in a
// duplicate network and performs the same cleanup/initial-state cycle as ABC.
func Abc_NtkAutoDebugModify(pNtkInit *Abc_Ntk_t, step, fConst1 int32) *Abc_Ntk_t {
	pNtk := abc.Abc_NtkDup(pNtkInit)
	assert(abc.Abc_NtkNodeNum(pNtk) == abc.Abc_NtkNodeNum(pNtkInit), "Abc_NtkNodeNum(pNtk) == Abc_NtkNodeNum(pNtkInit)")
	var pObj, pFanin *Abc_Obj_t
	if Abc_NtkFindGivenFanin(pNtk, step, &pObj, &pFanin) == 0 {
		abc.Abc_NtkDelete(pNtk)
		return nil
	}
	if abc.Abc_ObjIsPo(pObj) != 0 && abc.Abc_NodeIsConst(pFanin) != 0 {
		abc.Abc_NtkDeleteAll_rec(pObj)
		return pNtk
	}
	var pConst *Abc_Obj_t
	if fConst1 != 0 {
		pConst = abc.Abc_NtkCreateNodeConst1(pNtk)
	} else {
		pConst = abc.Abc_NtkCreateNodeConst0(pNtk)
	}
	abc.Abc_ObjTransferFanout(pFanin, pConst)
	abc.Abc_NtkDeleteAll_rec(pFanin)
	abc.Abc_NtkSweep(pNtk, 0)
	Abc_NtkCleanupSeq(pNtk, 0, 0, 0)
	abc.Abc_NtkToSop(pNtk, -1, 1_000_000_000)
	ret.Abc_NtkCycleInitStateSop(pNtk, 50, 0)
	return pNtk
}
