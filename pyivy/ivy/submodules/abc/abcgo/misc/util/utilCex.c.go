package util

import (
	"fmt"
	"math/bits"
)

func Abc_CexAlloc(regs, pis, frames int32) *Abc_Cex_t {
	if regs < 0 || pis < 0 || frames < 0 {
		panic("negative counterexample dimensions")
	}
	total := int64(regs) + int64(pis)*int64(frames)
	if total > 1<<31-1 {
		panic("counterexample dimensions overflow int32")
	}
	n := int32(total)
	return &Abc_Cex_t{NRegs: regs, NPis: pis, NBits: n, PData: make([]uint32, Abc_BitWordNum(n))}
}
func Abc_CexAllocFull(regs, pis, frames int32) *Abc_Cex_t {
	p := Abc_CexAlloc(regs, pis, frames)
	p.IPo = -1
	p.IFrame = -1
	for i := range p.PData {
		p.PData[i] = ^uint32(0)
	}
	return p
}
func Abc_CexMakeTriv(regs, pis, pos, frameOut int32) *Abc_Cex_t {
	// A combinational graph has no registers but still has a valid frame-zero
	// counterexample; the C helper's nRegs > 0 assertion excludes that case.
	assert(regs >= 0 && pos > 0 && frameOut >= 0, "trivial CEX dimensions")
	p := Abc_CexAlloc(regs, pis, frameOut/pos+1)
	p.IPo = frameOut % pos
	p.IFrame = frameOut / pos
	return p
}
func Abc_CexCreate(regs, pis int32, array []int32, frame, po, skipRegs int32) *Abc_Cex_t {
	p := Abc_CexAlloc(regs, pis, frame+1)
	p.IPo = po
	p.IFrame = frame
	if array == nil {
		return p
	}
	start := int32(0)
	if skipRegs != 0 {
		start = regs
	}
	for i := start; i < p.NBits; i++ {
		if array[i-start] != 0 {
			Abc_InfoSetBit(p.PData, i)
		}
	}
	return p
}
func Abc_CexDup(p *Abc_Cex_t, newRegs int32) *Abc_Cex_t {
	if p == AbcCexSentinel {
		return p
	}
	if newRegs == -1 {
		newRegs = p.NRegs
	}
	q := Abc_CexAlloc(newRegs, p.NPis, p.IFrame+1)
	q.IPo = p.IPo
	q.IFrame = p.IFrame
	for i := p.NRegs; i < p.NBits; i++ {
		if Abc_InfoHasBit(p.PData, i) {
			Abc_InfoSetBit(q.PData, q.NRegs+i-p.NRegs)
		}
	}
	return q
}
func Abc_CexDeriveFromCombModel(model []int32, pis, regs, po int32) *Abc_Cex_t {
	p := Abc_CexAlloc(regs, pis, 1)
	p.IPo = po
	for i := int32(0); i < pis; i++ {
		if model[i] != 0 {
			Abc_InfoSetBit(p.PData, i)
		}
	}
	return p
}
func Abc_CexMerge(p, part *Abc_Cex_t, beg, end int32) *Abc_Cex_t {
	if beg < 0 || end < 0 || beg > p.IFrame || end > p.IFrame || beg > end {
		return nil
	}
	assert(p.NPis == part.NPis && end-beg+part.IPo >= part.IFrame, "CEX merge dimensions")
	gain := end - beg + part.IPo - part.IFrame
	q := Abc_CexAlloc(p.NRegs, p.NPis, p.IFrame+1-gain)
	q.IPo = p.IPo
	q.IFrame = p.IFrame - gain
	dst := int32(0)
	copyBit := func(src *Abc_Cex_t, index int32) {
		if Abc_InfoHasBit(src.PData, index) {
			Abc_InfoSetBit(q.PData, dst)
		}
		dst++
	}
	for ; dst < p.NRegs; dst++ {
		if Abc_InfoHasBit(p.PData, dst) {
			Abc_InfoSetBit(q.PData, dst)
		}
	}
	for f := int32(0); f < beg; f++ {
		for i := int32(0); i < p.NPis; i++ {
			copyBit(p, p.NRegs+p.NPis*f+i)
		}
	}
	for f := int32(0); f < part.IFrame; f++ {
		for i := int32(0); i < p.NPis; i++ {
			copyBit(part, part.NRegs+p.NPis*f+i)
		}
	}
	for f := end; f <= p.IFrame; f++ {
		for i := int32(0); i < p.NPis; i++ {
			copyBit(p, p.NRegs+p.NPis*f+i)
		}
	}
	assert(dst == q.NBits, "merged bit count")
	return q
}
func Abc_CexPrintStats(p *Abc_Cex_t) {
	if p == nil {
		fmt.Printf("The counter example is NULL.\n")
		return
	}
	if p == AbcCexSentinel {
		fmt.Printf("The counter example is present but not available (pointer has value \"1\").\n")
		return
	}
	var counter int32
	for k := int32(0); k < p.NBits; k++ {
		if Abc_InfoHasBit(p.PData, k) {
			counter++
		}
	}
	fmt.Printf("CEX: Po =%4d  Frame =%4d  FF = %d  PI = %d  Bit =%8d  1s =%8d (%5.2f %%)\n",
		p.IPo, p.IFrame, p.NRegs, p.NPis, p.NBits, counter, 100.0*float64(counter)/float64(p.NBits-p.NRegs))
}

// Abc_CexPrintStatsInputs is utilCex.c's, separating the 1s of the real
// and pseudo inputs.
func Abc_CexPrintStatsInputs(p *Abc_Cex_t, nRealPis int32) {
	if p == nil {
		fmt.Printf("The counter example is NULL.\n")
		return
	}
	if p == AbcCexSentinel {
		fmt.Printf("The counter example is present but not available (pointer has value \"1\").\n")
		return
	}
	assert(nRealPis <= p.NPis, "real input count")
	var counter, counterPi, counterPpi int32
	for k := int32(0); k < p.NBits; k++ {
		v := int32(0)
		if Abc_InfoHasBit(p.PData, k) {
			v = 1
		}
		counter += v
		if nRealPis == p.NPis {
			continue
		}
		if (k-p.NRegs)%p.NPis < nRealPis {
			counterPi += v
		} else {
			counterPpi += v
		}
	}
	fmt.Printf("CEX: Po =%4d  Fr =%4d  FF = %d  PI = %d  Bit =%7d  1 =%8d (%5.2f %%)",
		p.IPo, p.IFrame, p.NRegs, p.NPis, p.NBits, counter, 100.0*float64(counter)/float64((p.IFrame+1)*p.NPis))
	if nRealPis < p.NPis {
		fmt.Printf(" 1pi =%8d (%5.2f %%) 1ppi =%8d (%5.2f %%)",
			counterPi, 100.0*float64(counterPi)/float64((p.IFrame+1)*nRealPis),
			counterPpi, 100.0*float64(counterPpi)/float64((p.IFrame+1)*(p.NPis-nRealPis)))
	}
	fmt.Printf("\n")
}
func Abc_CexPrint(p *Abc_Cex_t) {
	if p == nil {
		return
	}
	fmt.Print("Init : ")
	for i := int32(0); i < p.NRegs; i++ {
		if Abc_InfoHasBit(p.PData, i) {
			fmt.Print(1)
		} else {
			fmt.Print(0)
		}
	}
	fmt.Println()
	for f := int32(0); f <= p.IFrame; f++ {
		fmt.Printf("Frame %d : ", f)
		for i := int32(0); i < p.NPis; i++ {
			if Abc_InfoHasBit(p.PData, p.NRegs+f*p.NPis+i) {
				fmt.Print(1)
			} else {
				fmt.Print(0)
			}
		}
		fmt.Println()
	}
}
func Abc_CexFree(p *Abc_Cex_t) {
	if p != nil && p != AbcCexSentinel {
		p.PData = nil
	}
}
func Abc_CexFreeP(p **Abc_Cex_t) {
	if p != nil {
		Abc_CexFree(*p)
		*p = nil
	}
}
func Abc_CexTransformPhase(p *Abc_Cex_t, pis, pos, regs int32) *Abc_Cex_t {
	assert(pis > 0 && p.NPis%pis == 0, "phase input dimensions")
	frames := p.NPis / pis
	assert(p.IPo < pos*frames, "phase output index")
	q := Abc_CexDup(p, regs)
	q.NPis = pis
	q.IPo = -1
	q.IFrame = (p.IFrame+1)*frames - 1
	// Dup already accounts for the replacement register prefix.
	return q
}
func Abc_CexTransformTempor(p *Abc_Cex_t, pis, pos, regs int32) *Abc_Cex_t {
	assert(p.IFrame > 0 && pis > 0 && p.NPis%pis == 0, "temporal dimensions")
	frames := p.NPis/pis - 1
	q := Abc_CexAlloc(regs, pis, frames+p.IFrame+1)
	q.IPo = p.IPo
	q.IFrame = frames + p.IFrame
	dst := regs
	copyFrame := func(index int32) {
		for k := int32(0); k < pis; k++ {
			if Abc_InfoHasBit(p.PData, index+k) {
				Abc_InfoSetBit(q.PData, dst)
			}
			dst++
		}
	}
	for i := int32(0); i < frames; i++ {
		copyFrame(p.NRegs + (i+1)*pis)
	}
	for i := int32(0); i <= p.IFrame; i++ {
		copyFrame(p.NRegs + i*p.NPis)
	}
	assert(dst == q.NBits, "temporal bit count")
	return q
}
func Abc_CexTransformUndc(p *Abc_Cex_t, init string) *Abc_Cex_t {
	added := int32(0)
	for _, c := range init {
		if c == 'x' || c == 'X' {
			added++
		}
	}
	regs := int32(len(init))
	q := Abc_CexAlloc(regs, p.NPis-added, p.IFrame+1)
	q.IPo = p.IPo
	q.IFrame = p.IFrame
	extra := int32(0)
	for i, c := range []byte(init) {
		set := c == '1'
		if c == 'x' || c == 'X' {
			set = Abc_InfoHasBit(p.PData, p.NRegs+p.NPis-added+extra)
			extra++
		}
		if set {
			Abc_InfoSetBit(q.PData, int32(i))
		}
	}
	for f := int32(0); f <= p.IFrame; f++ {
		for i := int32(0); i < q.NPis; i++ {
			if Abc_InfoHasBit(p.PData, p.NRegs+p.NPis*f+i) {
				Abc_InfoSetBit(q.PData, q.NRegs+q.NPis*f+i)
			}
		}
	}
	return q
}
func Abc_CexPermute(p *Abc_Cex_t, mapping []int32) *Abc_Cex_t {
	assert(int32(len(mapping)) == p.NPis, "permutation size")
	q := Abc_CexAlloc(p.NRegs, p.NPis, p.IFrame+1)
	q.IPo = p.IPo
	q.IFrame = p.IFrame
	// Only PIs are permuted; keep the register initialization unchanged.
	words := p.NRegs / 32
	copy(q.PData[:words], p.PData[:words])
	if bits := p.NRegs % 32; bits != 0 {
		q.PData[words] = p.PData[words] & ((uint32(1) << uint(bits)) - 1)
	}
	for i := p.NRegs; i < p.NBits; i++ {
		if Abc_InfoHasBit(p.PData, i) {
			j := p.NRegs + p.NPis*((i-p.NRegs)/p.NPis) + mapping[(i-p.NRegs)%p.NPis]
			Abc_InfoSetBit(q.PData, j)
		}
	}
	return q
}
func Abc_CexPermuteTwo(p *Abc_Cex_t, old, new []int32) *Abc_Cex_t {
	assert(len(old) == int(p.NPis) && len(new) == len(old), "permutation size")
	mapping := make([]int32, p.NPis)
	for i := range old {
		mapping[old[i]] = new[i]
	}
	return Abc_CexPermute(p, mapping)
}
func Abc_CexOnes32(i uint32) int32 { return int32(bits.OnesCount32(i)) }
func Abc_CexCountOnes(p *Abc_Cex_t) int32 {
	var count int32
	for _, w := range p.PData[:Abc_BitWordNum(p.NBits)] {
		count += int32(bits.OnesCount32(w))
	}
	return count
}
