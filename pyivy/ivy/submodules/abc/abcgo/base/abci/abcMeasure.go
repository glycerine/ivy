package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
)

func Abc_NtkPrintMeasures(pTruth []uint32, nVars int32) {
	var cof [10][32]uint32
	count := func(x []uint32) int32 { return kit.Kit_TruthCountOnes(x, nVars) }
	// The C routine counts an uninitialized stack buffer here. Use the input
	// truth table, which is the only deterministic interpretation of the total.
	nOnes := count(pTruth)
	// Small Kit tables repeat across a full word; count assignments once.
	if nVars < 5 {
		nOnes >>= uint(5 - nVars)
	}
	fmt.Printf("Total = %d.\n", nOnes*((1<<nVars)-nOnes))
	for i := int32(0); i < nVars; i++ {
		kit.Kit_TruthUniqueNew(cof[0][:], pTruth, nVars, i)
		fmt.Printf("%7d ", count(cof[0][:]))
	}
	fmt.Printf("\n")
	for i := int32(0); i < nVars; i++ {
		for k := int32(0); k < nVars; k++ {
			if i == k {
				fmt.Printf("        ")
				continue
			}
			kit.Kit_TruthCofactor0New(cof[0][:], pTruth, nVars, i)
			kit.Kit_TruthCofactor1New(cof[1][:], pTruth, nVars, i)
			kit.Kit_TruthCofactor0New(cof[2][:], cof[0][:], nVars, k)
			kit.Kit_TruthCofactor1New(cof[3][:], cof[0][:], nVars, k)
			kit.Kit_TruthCofactor0New(cof[4][:], cof[1][:], nVars, k)
			kit.Kit_TruthCofactor1New(cof[5][:], cof[1][:], nVars, k)
			kit.Kit_TruthAndPhase(cof[6][:], cof[2][:], cof[5][:], nVars, 0, 1)
			kit.Kit_TruthAndPhase(cof[7][:], cof[2][:], cof[5][:], nVars, 1, 0)
			kit.Kit_TruthAndPhase(cof[8][:], cof[3][:], cof[4][:], nVars, 0, 1)
			kit.Kit_TruthAndPhase(cof[9][:], cof[3][:], cof[4][:], nVars, 1, 0)
			nOnes = count(cof[6][:]) + count(cof[7][:]) + count(cof[8][:]) + count(cof[9][:])
			fmt.Printf("%7d ", nOnes)
			if k == nVars-1 {
				fmt.Printf("\n")
			}
		}
	}
	fmt.Printf("\n")
}
func Abc_Ntk4VarObjPrint_rec(pObj *Abc_Obj_t) {
	if pObj == nil {
		return
	}
	if pObj.Kind == abc.ObjectConst1 {
		fmt.Print("1")
		return
	}
	if pObj.Kind == abc.ObjectPI {
		fmt.Printf("%c", rune(pObj.ID-1+'a'))
		return
	}
	fanins := pObj.Fanins()
	if len(fanins) < 2 {
		return
	}
	fmt.Print("(")
	Abc_Ntk4VarObjPrint_rec(fanins[0].Object)
	if fanins[0].Complement {
		fmt.Print("'")
	}
	Abc_Ntk4VarObjPrint_rec(fanins[1].Object)
	if fanins[1].Complement {
		fmt.Print("'")
	}
	fmt.Print(")")
}
func Abc_Ntk4VarObj(vNodes []*Abc_Obj_t) uint32 {
	var truth uint32
	for _, obj := range vNodes {
		f := obj.Fanins()
		if len(f) < 2 {
			continue
		}
		a, ok := f[0].Object.CopyData.(uint32)
		assert(ok, "fanin 0 truth table is available")
		b, ok := f[1].Object.CopyData.(uint32)
		assert(ok, "fanin 1 truth table is available")
		if f[0].Complement {
			a = ^a
		}
		if f[1].Complement {
			b = ^b
		}
		truth = a & b
		obj.CopyData = truth
	}
	return truth
}

var spStatic_Abc_Ntk4VarTable_u4VarTruths [4]uint32 = [4]uint32{0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00}
var spStatic_Abc_Ntk4VarTable_u4VarTts [222]uint32 = [222]uint32{
	0x0000, 0x0001, 0x0003, 0x0006, 0x0007, 0x000f, 0x0016, 0x0017, 0x0018, 0x0019,
	0x001b, 0x001e, 0x001f, 0x003c, 0x003d, 0x003f, 0x0069, 0x006b, 0x006f, 0x007e,
	0x007f, 0x00ff, 0x0116, 0x0117, 0x0118, 0x0119, 0x011a, 0x011b, 0x011e, 0x011f,
	0x012c, 0x012d, 0x012f, 0x013c, 0x013d, 0x013e, 0x013f, 0x0168, 0x0169, 0x016a,
	0x016b, 0x016e, 0x016f, 0x017e, 0x017f, 0x0180, 0x0181, 0x0182, 0x0183, 0x0186,
	0x0187, 0x0189, 0x018b, 0x018f, 0x0196, 0x0197, 0x0198, 0x0199, 0x019a, 0x019b,
	0x019e, 0x019f, 0x01a8, 0x01a9, 0x01aa, 0x01ab, 0x01ac, 0x01ad, 0x01ae, 0x01af,
	0x01bc, 0x01bd, 0x01be, 0x01bf, 0x01e8, 0x01e9, 0x01ea, 0x01eb, 0x01ee, 0x01ef,
	0x01fe, 0x033c, 0x033d, 0x033f, 0x0356, 0x0357, 0x0358, 0x0359, 0x035a, 0x035b,
	0x035e, 0x035f, 0x0368, 0x0369, 0x036a, 0x036b, 0x036c, 0x036d, 0x036e, 0x036f,
	0x037c, 0x037d, 0x037e, 0x03c0, 0x03c1, 0x03c3, 0x03c5, 0x03c6, 0x03c7, 0x03cf,
	0x03d4, 0x03d5, 0x03d6, 0x03d7, 0x03d8, 0x03d9, 0x03db, 0x03dc, 0x03dd, 0x03de,
	0x03fc, 0x0660, 0x0661, 0x0662, 0x0663, 0x0666, 0x0667, 0x0669, 0x066b, 0x066f,
	0x0672, 0x0673, 0x0676, 0x0678, 0x0679, 0x067a, 0x067b, 0x067e, 0x0690, 0x0691,
	0x0693, 0x0696, 0x0697, 0x069f, 0x06b0, 0x06b1, 0x06b2, 0x06b3, 0x06b4, 0x06b5,
	0x06b6, 0x06b7, 0x06b9, 0x06bd, 0x06f0, 0x06f1, 0x06f2, 0x06f6, 0x06f9, 0x0776,
	0x0778, 0x0779, 0x077a, 0x077e, 0x07b0, 0x07b1, 0x07b4, 0x07b5, 0x07b6, 0x07bc,
	0x07e0, 0x07e1, 0x07e2, 0x07e3, 0x07e6, 0x07e9, 0x07f0, 0x07f1, 0x07f2, 0x07f8,
	0x0ff0, 0x1668, 0x1669, 0x166a, 0x166b, 0x166e, 0x167e, 0x1681, 0x1683, 0x1686,
	0x1687, 0x1689, 0x168b, 0x168e, 0x1696, 0x1697, 0x1698, 0x1699, 0x169a, 0x169b,
	0x169e, 0x16a9, 0x16ac, 0x16ad, 0x16bc, 0x16e9, 0x177e, 0x178e, 0x1796, 0x1798,
	0x179a, 0x17ac, 0x17e8, 0x18e7, 0x19e1, 0x19e3, 0x19e6, 0x1bd8, 0x1be4, 0x1ee1,
	0x3cc3, 0x6996,
}

func Abc_Ntk4VarTable(pNtk *Abc_Ntk_t) {
	var counters [222]int32
	assert(abc.Abc_NtkPiNum(pNtk) == 4, "Abc_NtkPiNum(pNtk) == 4")
	pis := pNtk.ObjectsOfKind(abc.ObjectPI)
	pos := pNtk.ObjectsOfKind(abc.ObjectPO)
	elementary := [4]uint32{0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00}
	for i, obj := range pis {
		obj.CopyData = elementary[i]
	}
	pNtk.One().Object.CopyData = uint32(0xffffffff)
	for _, po := range pos {
		nodes := abc.Abc_NtkDfsNodes(pNtk, []*abc.Object{po}, 1)
		var truth uint32
		if len(nodes) == 0 {
			fs := po.Fanins()
			if len(fs) > 0 {
				var ok bool
				truth, ok = fs[0].Object.CopyData.(uint32)
				assert(ok, "PO fanin truth table is available")
			}
		} else {
			truth = Abc_Ntk4VarObj(nodes)
		}
		if truth&0xffff < ^truth&0xffff {
			truth &= 0xffff
		} else {
			truth = ^truth & 0xffff
		}
		k := 0
		for k < 222 && spStatic_Abc_Ntk4VarTable_u4VarTts[k] != truth {
			k++
		}
		if k < 222 {
			counters[k]++
		}
	}
	total := int32(0)
	for k := 0; k < 222; k++ {
		fmt.Printf("%d/%x/%d ", k, spStatic_Abc_Ntk4VarTable_u4VarTts[k], counters[k])
		total += counters[k]
	}
	fmt.Printf(" Total = %d\n", total)
}
func Abc_NtkPrintOneDecompCheckCofList(cofs []uint32, n int32) int32 {
	ind := int32(-1)
	assert(n > 2, "nCofs > 2")
	for i := int32(1); i < n; i++ {
		if cofs[i] == cofs[0] {
			continue
		}
		if ind < 0 {
			ind = i
			continue
		}
		if cofs[i] == cofs[ind] {
			continue
		}
		return 0
	}
	return 1
}
func Abc_NtkPrintOneDecompCheck(cofs []uint32, n int32, mask uint32) int32 {
	var groups [32][32]uint32
	var counts, masks [32]int32
	ng := int32(0)
	for i := int32(0); i < n; i++ {
		k := int32(0)
		for k < ng && int32(uint32(i)&mask) != masks[k] {
			k++
		}
		if k == ng {
			masks[k] = int32(uint32(i) & mask)
			ng++
		}
		groups[k][counts[k]] = cofs[i]
		counts[k]++
		assert(counts[k] <= 32, "nCofNums[k] <= 32")
		assert(ng <= 32, "nGroups <= 32")
	}
	for i := int32(0); i < ng; i++ {
		if Abc_NtkPrintOneDecompCheckCofList(groups[i][:], counts[i]) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkPrintOneDecomp_rec(cofs []uint32, n, nVars int32, mask uint32, best *int32, bestMask *uint32) {
	last, counter := -1, int32(0)
	for v := int32(0); v < nVars; v++ {
		if mask&(1<<v) != 0 {
			last = int(v)
			counter++
		}
	}
	if counter > 3 {
		return
	}
	for v := int32(last + 1); v < nVars; v++ {
		next := mask | (1 << v)
		if Abc_NtkPrintOneDecompCheck(cofs, n, next) == 0 {
			continue
		}
		if *best < counter+1 {
			*best = counter + 1
			*bestMask = next
		}
		Abc_NtkPrintOneDecomp_rec(cofs, n, nVars, next, best, bestMask)
	}
}
func Abc_NtkPrintOneDecomp(truth []uint32, nVars int32) {
	bound := int32(6)
	var cofs [64]uint32
	var bestMask uint32
	best := int32(1)
	assert(nVars > bound, "nVars > BoundSet")
	assert(nVars <= bound+5, "nVars <= BoundSet + 5")
	nCofs := int32(1 << bound)
	nMints := int32(1 << (nVars - bound))
	mask := kit.Kit_CubeMask(nMints)
	shift := int32(0)
	for i := int32(0); i < nCofs; i++ {
		cofs[i] = (truth[shift/32] >> (shift % 32)) & mask
		shift += nMints
	}
	for i := int32(0); i < bound; i++ {
		Abc_NtkPrintOneDecomp_rec(cofs[:], nCofs, nVars, 1<<i, &best, &bestMask)
	}
	fmt.Printf("Best size = %d  Best mask = ", best)
	extra.Extra_PrintBinary(os.Stdout, []uint32{bestMask}, nVars)
	fmt.Printf("\n")
}
func Abc_NtkPrintOneDec(truth []uint32, nVars int32) {
	var cof [1 << 11]uint32
	var diffs, order [16]int32
	out, in := cof[:], truth
	counter := int32(0)
	for i := int32(0); i < nVars; i++ {
		kit.Kit_TruthUniqueNew(cof[:], truth, nVars, i)
		diffs[i] = kit.Kit_TruthCountOnes(cof[:], nVars)
		order[i] = i
	}
	for {
		changed := false
		for i := int32(0); i < nVars-1; i++ {
			if diffs[i] <= diffs[i+1] {
				continue
			}
			changed = true
			counter++
			diffs[i], diffs[i+1] = diffs[i+1], diffs[i]
			order[i], order[i+1] = order[i+1], order[i]
			extra.Extra_TruthSwapAdjacentVars(out, in, nVars, i)
			out, in = in, out
		}
		if !changed {
			break
		}
	}
	if counter&1 != 0 {
		extra.Extra_TruthCopy(out, in, nVars)
	}
	Abc_NtkPrintOneDecomp(truth, nVars)
}
