// Cut matching, area flow and bounded exact-area recovery from amapMatch.c.
package amap

import (
	"fmt"
	"math"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

func Amap_ManDupCut(p *Amap_Man_t, c *Amap_Cut_t) *Amap_Cut_t {
	if c == nil {
		return nil
	}
	copy := *c
	copy.Fans = append([]int32(nil), c.Fans...)
	copy.pNext = nil
	return &copy
}
func Amap_ManMatchStart(m *Amap_Mat_t, c *Amap_Cut_t, s *Amap_Set_t) {
	*m = Amap_Mat_t{pCut: c, pSet: s}
}
func amapMatchObjects(p *Amap_Man_t) []*Amap_Obj_t {
	result := make([]*Amap_Obj_t, 0, Vec_PtrSize(p.vObjs))
	for _, entry := range vec.Vec_PtrArray(p.vObjs) {
		if entry != nil {
			result = append(result, entry.(*Amap_Obj_t))
		}
	}
	return result
}
func Amap_ManCleanRefs(p *Amap_Man_t) {
	for _, o := range amapMatchObjects(p) {
		o.nFouts = [2]int32{}
	}
}
func Amap_ManCleanData(p *Amap_Man_t) {
	for _, o := range amapMatchObjects(p) {
		o.pData = nil
	}
}
func Amap_ManMaxDelay(p *Amap_Man_t) float32 {
	var delay float32
	for _, entry := range vec.Vec_PtrArray(p.vPos) {
		o := entry.(*Amap_Obj_t)
		delay = max(delay, Amap_Regular(Amap_ObjFanin0(p, o)).Best.Delay)
	}
	return delay
}

// The set permutes and complements the cut leaves into gate-pin order.
func amapMatchFanins(p *Amap_Man_t, m *Amap_Mat_t, visit func(*Amap_Obj_t, int32)) {
	if m.pCut == nil || m.pSet == nil {
		panic("AMAP: missing match")
	}
	gate := Amap_LibGate(p.pLib, int32(m.pSet.iGate))
	if gate.nPins != m.pCut.nFans || int(m.pCut.nFans) > len(m.pCut.Fans) {
		panic("AMAP: inconsistent match support")
	}
	for i := uint32(0); i < gate.nPins; i++ {
		selector := int32(m.pSet.Ins[i])
		index := selector >> 1
		if index < 0 || int(index) >= len(m.pCut.Fans) {
			panic("AMAP: invalid match permutation")
		}
		literal := m.pCut.Fans[index]
		visit(Amap_ManObj(p, literal>>1), (selector^literal)&1)
	}
}
func Amap_ManComputeMapping_rec(p *Amap_Man_t, o *Amap_Obj_t, polarity int32) float32 {
	o = Amap_Regular(o)
	old := Amap_ObjRefsTotal(o)
	o.nFouts[polarity]++
	if old > 0 || Amap_ObjIsNode(o) == 0 {
		return 0
	}
	area := float32(Amap_LibGate(p.pLib, int32(o.Best.pSet.iGate)).dArea)
	amapMatchFanins(p, &o.Best, func(f *Amap_Obj_t, c int32) { area += Amap_ManComputeMapping_rec(p, f, c) })
	return area
}
func Amap_ManComputeMapping(p *Amap_Man_t) float32 {
	Amap_ManCleanRefs(p)
	var area float32
	for _, entry := range vec.Vec_PtrArray(p.vPos) {
		o := entry.(*Amap_Obj_t)
		area += Amap_ManComputeMapping_rec(p, Amap_Regular(Amap_ObjFanin0(p, o)), Amap_ObjFaninC0(o))
	}
	return area
}
func Amap_ManCountInverters(p *Amap_Man_t) int32 {
	var count int32
	for _, o := range amapMatchObjects(p) {
		if o.nFouts[o.fPolar^1] > 0 {
			count++
		}
	}
	return count
}
func amapCompareMetric(a, b, epsilon float32) int32 {
	if a < b-epsilon {
		return -1
	}
	if a > b+epsilon {
		return 1
	}
	return 0
}
func amapCompareFanout(a, b, epsilon float32) int32 {
	// Preserve the original preference, including its asymmetric treatment of nearly tied matches.
	if a > b-epsilon {
		return -1
	}
	if a < b+epsilon {
		return 1
	}
	return 0
}
func Amap_CutCompareDelay(p *Amap_Man_t, a, b *Amap_Mat_t) int32 {
	e := p.pPars.FEpsilon
	if c := amapCompareMetric(a.Delay, b.Delay, e); c != 0 {
		return c
	}
	if c := amapCompareMetric(a.Area, b.Area, e); c != 0 {
		return c
	}
	if c := amapCompareFanout(a.AveFan, b.AveFan, e); c != 0 {
		return c
	}
	return 1
}
func Amap_CutCompareArea(p *Amap_Man_t, a, b *Amap_Mat_t) int32 {
	e := p.pPars.FEpsilon
	if c := amapCompareMetric(a.Area, b.Area, e); c != 0 {
		return c
	}
	if c := amapCompareFanout(a.AveFan, b.AveFan, e); c != 0 {
		return c
	}
	if c := amapCompareMetric(a.Delay, b.Delay, e); c != 0 {
		return c
	}
	return 1
}
func Amap_CutAreaDeref(p *Amap_Man_t, m *Amap_Mat_t) float32 {
	area := float32(Amap_LibGate(p.pLib, int32(m.pSet.iGate)).dArea)
	amapMatchFanins(p, m, func(f *Amap_Obj_t, c int32) {
		if Amap_ObjRefsTotal(f) <= 0 || f.nFouts[c] <= 0 {
			panic("AMAP: dereferencing unused polarity")
		}
		if int32(f.fPolar) != c && f.nFouts[c] == 1 {
			area += p.fAreaInv
		}
		f.nFouts[c]--
		if Amap_ObjRefsTotal(f) == 0 && Amap_ObjIsNode(f) != 0 {
			area += Amap_CutAreaDeref(p, &f.Best)
		}
	})
	return area
}
func Amap_CutAreaRef2(p *Amap_Man_t, m *Amap_Mat_t, temp *Vec_Ptr_t, limit int32) float32 {
	area := float32(Amap_LibGate(p.pLib, int32(m.pSet.iGate)).dArea)
	if limit == 0 {
		return area
	}
	amapMatchFanins(p, m, func(f *Amap_Obj_t, c int32) {
		vec.Vec_PtrPush(temp, &f.nFouts[c])
		if int32(f.fPolar) != c && f.nFouts[c] == 0 {
			area += p.fAreaInv
		}
		old := Amap_ObjRefsTotal(f)
		f.nFouts[c]++
		if old == 0 && Amap_ObjIsNode(f) != 0 {
			area += Amap_CutAreaRef2(p, &f.Best, temp, limit-1)
		}
	})
	return area
}
func amapOutputInverter(p *Amap_Man_t, o *Amap_Obj_t, m *Amap_Mat_t) float32 {
	c := m.pCut.fInv ^ m.pSet.fInv
	if o.nFouts[c] == 0 && o.nFouts[c^1] > 0 {
		return p.fAreaInv
	}
	return 0
}
func Amap_CutAreaDerefed2(p *Amap_Man_t, o *Amap_Obj_t, m *Amap_Mat_t) float32 {
	if p.vTempP == nil {
		p.vTempP = vec.Vec_PtrAlloc(16)
	}
	vec.Vec_PtrClear(p.vTempP)
	area := Amap_CutAreaRef2(p, m, p.vTempP, 8)
	for _, entry := range vec.Vec_PtrArray(p.vTempP) {
		counter := entry.(*int32)
		*counter--
	}
	vec.Vec_PtrClear(p.vTempP)
	return area + amapOutputInverter(p, o, m)
}
func Amap_CutAreaRef(p *Amap_Man_t, m *Amap_Mat_t) float32 {
	area := float32(Amap_LibGate(p.pLib, int32(m.pSet.iGate)).dArea)
	amapMatchFanins(p, m, func(f *Amap_Obj_t, c int32) {
		if int32(f.fPolar) != c && f.nFouts[c] == 0 {
			area += p.fAreaInv
		}
		old := Amap_ObjRefsTotal(f)
		f.nFouts[c]++
		if old == 0 && Amap_ObjIsNode(f) != 0 {
			area += Amap_CutAreaRef(p, &f.Best)
		}
	})
	return area
}
func amapCheckArea(p *Amap_Man_t, a, b float32) {
	if math.Abs(float64(a-b)) > float64(p.fEpsilonInternal) {
		panic("AMAP: reference/dereference area mismatch")
	}
}
func Amap_CutAreaDerefed(p *Amap_Man_t, o *Amap_Obj_t, m *Amap_Mat_t) float32 {
	a := Amap_CutAreaRef(p, m)
	b := Amap_CutAreaDeref(p, m)
	amapCheckArea(p, a, b)
	return b + amapOutputInverter(p, o, m)
}
func Amap_CutAreaTest(p *Amap_Man_t, o *Amap_Obj_t) {
	var a, b float32
	if Amap_ObjRefsTotal(o) == 0 {
		a = Amap_CutAreaRef(p, &o.Best)
		b = Amap_CutAreaDeref(p, &o.Best)
	} else {
		a = Amap_CutAreaDeref(p, &o.Best)
		b = Amap_CutAreaRef(p, &o.Best)
	}
	amapCheckArea(p, a, b)
}
func Amap_ManMatchGetFlows(p *Amap_Man_t, m *Amap_Mat_t) {
	gate := Amap_LibGate(p.pLib, int32(m.pSet.iGate))
	m.Area = float32(gate.dArea)
	m.AveFan = 0
	m.Delay = 0
	amapMatchFanins(p, m, func(f *Amap_Obj_t, c int32) {
		m.Delay = max(m.Delay, f.Best.Delay)
		m.AveFan += float32(Amap_ObjRefsTotal(f))
		add := f.Best.Area
		if Amap_ObjRefsTotal(f) > 0 {
			add /= f.EstRefs
		}
		if m.Area >= 1e32 || add >= 1e32 {
			m.Area = 1e32
		} else {
			m.Area += add
		}
	})
	if gate.nPins > 0 {
		m.AveFan /= float32(gate.nPins)
	}
	m.Delay++
}
func Amap_ManMatchGetExacts(p *Amap_Man_t, o *Amap_Obj_t, m *Amap_Mat_t) {
	gate := Amap_LibGate(p.pLib, int32(m.pSet.iGate))
	m.Delay = 0
	m.AveFan = 0
	amapMatchFanins(p, m, func(f *Amap_Obj_t, c int32) {
		m.Delay = max(m.Delay, f.Best.Delay)
		m.AveFan += float32(Amap_ObjRefsTotal(f))
	})
	if gate.nPins > 0 {
		m.AveFan /= float32(gate.nPins)
	}
	m.Delay++
	m.Area = Amap_CutAreaDerefed2(p, o, m)
}
func Amap_ManMatchNode(p *Amap_Man_t, o *Amap_Obj_t, flow, refs int32) {
	if refs != 0 {
		o.EstRefs = (2*o.EstRefs + float32(Amap_ObjRefsTotal(o))) / 3
	} else {
		o.EstRefs = float32(o.nRefs)
	}
	referenced := refs != 0 && Amap_ObjRefsTotal(o) > 0
	if referenced {
		Amap_CutAreaDeref(p, &o.Best)
	}
	var bestArea, bestDelay Amap_Mat_t
	cut, ok := o.pData.(*Amap_Cut_t)
	if !ok {
		panic("AMAP: node cuts are not a native cut chain")
	}
	for ; cut != nil; cut = cut.pNext {
		if cut.iMat == 0 {
			continue
		}
		for set := Amap_LibNod(p.pLib, int32(cut.iMat)).pSets; set != nil; set = set.pNext {
			m := Amap_Mat_t{pCut: cut, pSet: set}
			if flow != 0 {
				Amap_ManMatchGetFlows(p, &m)
			} else {
				Amap_ManMatchGetExacts(p, o, &m)
			}
			if bestDelay.pCut == nil || Amap_CutCompareDelay(p, &bestDelay, &m) == 1 {
				bestDelay = m
			}
			if bestArea.pCut == nil || Amap_CutCompareArea(p, &bestArea, &m) == 1 {
				bestArea = m
			}
		}
	}
	if bestDelay.pCut == nil {
		panic("AMAP: no implementable cut")
	}
	best := bestDelay
	areaGain := float32(math.Abs(float64(bestArea.Area-bestDelay.Area))) / bestDelay.Area
	delayGain := float32(math.Abs(float64(bestArea.Delay-bestDelay.Delay))) / bestArea.Delay
	if areaGain >= p.pPars.FADratio*delayGain {
		best = bestArea
	}
	o.fPolar = best.pCut.fInv ^ best.pSet.fInv
	o.Best = best
	o.Best.pCut = Amap_ManDupCut(p, best.pCut)
	if referenced {
		Amap_CutAreaRef(p, &o.Best)
	}
}
func Amap_ManMatch(p *Amap_Man_t, flow, refs int32) {
	start := time.Now()
	for _, o := range amapMatchObjects(p) {
		if Amap_ObjIsNode(o) != 0 && o.pData != nil {
			Amap_ManMatchNode(p, o, flow, refs)
		}
	}
	area := Amap_ManComputeMapping(p)
	inverters := Amap_ManCountInverters(p)
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Area =%9.2f. Gate =%9.2f. Inv =%9.2f. (%6d.) Delay =%6.2f. Time %s\n", area+float32(inverters)*p.fAreaInv, area, float32(inverters)*p.fAreaInv, inverters, Amap_ManMaxDelay(p), time.Since(start))
	}
}
func Amap_ManMap(p *Amap_Man_t) {
	Amap_ManMerge(p)
	for i := int32(0); i < p.pPars.NIterFlow; i++ {
		Amap_ManMatch(p, 1, amapBool(i > 0))
	}
	for i := int32(0); i < p.pPars.NIterArea; i++ {
		Amap_ManMatch(p, 0, amapBool(p.pPars.NIterFlow > 0 || i > 0))
	}
	Amap_ManCleanData(p)
}
