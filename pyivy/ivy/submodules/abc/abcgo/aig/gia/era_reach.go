package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"os"
	"time"
)

type Gia_ObjEra_t struct {
	Num, Cond, iPrev, iNext int32
	pData                   []uint32
}
type Gia_ManEra_t struct {
	pAig                 *Gia_Man_t
	nWordsSim, nWordsDat int32
	pDataSim             []uint32
	vStates              []*Gia_ObjEra_t
	pStateNew            *Gia_ObjEra_t
	iCurState            int32
	vBugTrace, vStgDump  *Vec_Int_t
	nBins                int32
	pBins                []int32
}

func Gia_ManEraData(p *Gia_ManEra_t, id int32) []uint32 {
	start := id * p.nWordsSim
	return p.pDataSim[start : start+p.nWordsSim]
}
func Gia_ManEraState(p *Gia_ManEra_t, id int32) *Gia_ObjEra_t { return p.vStates[id] }
func Gia_ManEraCreate(g *Gia_Man_t) *Gia_ManEra_t {
	inputs := Gia_ManPiNum(g)
	if inputs < 0 || inputs > 12 || g.nRegs <= 0 {
		panic("gia: explicit reachability needs 1+ registers and at most 12 inputs")
	}
	words := max(int32(1), (int32(1)<<inputs)/32)
	p := &Gia_ManEra_t{pAig: g, nWordsSim: words, nWordsDat: (g.nRegs + 31) / 32, pDataSim: make([]uint32, words*g.nObjs), vStates: []*Gia_ObjEra_t{nil}, nBins: util.Abc_PrimeCudd(100000), vStgDump: vec.Vec_IntAlloc(1000)}
	p.pBins = make([]int32, p.nBins)
	for i := int32(0); i < inputs; i++ {
		data := Gia_ManEraData(p, Gia_ManPi(g, i).id)
		for bit := int32(0); bit < words*32; bit++ {
			if bit&(1<<i) != 0 {
				data[bit>>5] |= 1 << uint(bit&31)
			}
		}
	}
	return p
}
func Gia_ManEraFree(p *Gia_ManEra_t) {
	if p != nil {
		*p = Gia_ManEra_t{}
	}
}
func Gia_ManEraCreateState(p *Gia_ManEra_t) *Gia_ObjEra_t {
	state := &Gia_ObjEra_t{Num: int32(len(p.vStates)), pData: make([]uint32, p.nWordsDat)}
	p.vStates = append(p.vStates, state)
	return state
}

var giaEraPrimes = [128]uint32{1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147}

func Gia_ManEraStateHash(data []uint32, nWords, nBins int32) int32 {
	var hash uint32
	for i := int32(0); i < nWords; i++ {
		hash ^= data[i] * giaEraPrimes[i&127]
	}
	return int32(hash % uint32(nBins))
}
func Gia_ManEraHashFind(p *Gia_ManEra_t, state *Gia_ObjEra_t, num *int32) *int32 {
	place := &p.pBins[Gia_ManEraStateHash(state.pData, p.nWordsDat, p.nBins)]
	for *place != 0 {
		other := Gia_ManEraState(p, *place)
		equal := true
		for i := range state.pData {
			if state.pData[i] != other.pData[i] {
				equal = false
				break
			}
		}
		if equal {
			if num != nil {
				*num = other.Num
			}
			return nil
		}
		place = &other.iNext
	}
	if num != nil {
		*num = -1
	}
	return place
}
func Gia_ManEraHashResize(p *Gia_ManEra_t) {
	p.nBins = util.Abc_PrimeCudd(uint32(3 * p.nBins))
	p.pBins = make([]int32, p.nBins)
	for _, state := range p.vStates[1:] {
		if state == p.pStateNew {
			continue
		}
		state.iNext = 0
		place := Gia_ManEraHashFind(p, state, nil)
		if place == nil {
			panic("gia: duplicate reachability state")
		}
		*place = state.Num
	}
}
func Gia_ManInsertState(p *Gia_ManEra_t, state *Gia_ObjEra_t) {
	for i := int32(0); i < p.pAig.nRegs; i++ {
		data := Gia_ManEraData(p, Gia_ManRo(p.pAig, i).id)
		var v uint32
		if state.pData[i>>5]&(1<<uint(i&31)) != 0 {
			v = ^uint32(0)
		}
		for j := range data {
			data[j] = v
		}
	}
}
func Gia_ManOutputAsserted(p *Gia_ManEra_t, obj *Gia_Obj_t) int32 {
	for i, w := range Gia_ManEraData(p, obj.id) {
		if w != 0 {
			return int32(i*32 + bits.TrailingZeros32(w))
		}
	}
	return -1
}
func Gia_ManSimulateCo(p *Gia_ManEra_t, obj *Gia_Obj_t) {
	dst, src := Gia_ManEraData(p, obj.id), Gia_ManEraData(p, Gia_ObjFaninId0(obj, obj.id))
	for i, x := range src {
		if obj.fCompl0 != 0 {
			x = ^x
		}
		dst[i] = x
	}
}
func Gia_ManSimulateNode(p *Gia_ManEra_t, obj *Gia_Obj_t) {
	dst, a, b := Gia_ManEraData(p, obj.id), Gia_ManEraData(p, Gia_ObjFaninId0(obj, obj.id)), Gia_ManEraData(p, Gia_ObjFaninId1(obj, obj.id))
	for i := range dst {
		x, y := a[i], b[i]
		if obj.fCompl0 != 0 {
			x = ^x
		}
		if obj.fCompl1 != 0 {
			y = ^y
		}
		dst[i] = x & y
	}
}
func Gia_ManPerformOneIter(p *Gia_ManEra_t) {
	for _, obj := range p.pAig.pObjs[1:] {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManSimulateNode(p, obj)
		} else if Gia_ObjIsCo(obj) != 0 {
			Gia_ManSimulateCo(p, obj)
		}
	}
}
func Gia_ManCollectBugTrace(p *Gia_ManEra_t, state *Gia_ObjEra_t, condition int32) *Vec_Int_t {
	trace := vec.Vec_IntAlloc(10)
	vec.Vec_IntPush(trace, condition)
	for state != nil {
		vec.Vec_IntPush(trace, state.Cond)
		if state.iPrev == 0 {
			break
		}
		state = Gia_ManEraState(p, state.iPrev)
	}
	vec.Vec_IntReverseOrder(trace)
	return trace
}
func Gia_ManCountDepth(p *Gia_ManEra_t) int32 {
	if len(p.vStates) < 2 {
		return 0
	}
	state := p.vStates[len(p.vStates)-1]
	if state.iPrev == 0 && len(p.vStates) > 3 {
		state = p.vStates[len(p.vStates)-2]
	}
	var n int32
	for state != nil {
		n++
		if state.iPrev == 0 {
			break
		}
		state = Gia_ManEraState(p, state.iPrev)
	}
	return n
}
func Gia_ManAnalyzeResult(p *Gia_ManEra_t, state *Gia_ObjEra_t, miter, dump int32) int32 {
	if miter != 0 {
		for i := int32(0); i < Gia_ManPoNum(p.pAig); i++ {
			if cond := Gia_ManOutputAsserted(p, Gia_ManPo(p.pAig, i)); cond >= 0 {
				p.vBugTrace = Gia_ManCollectBugTrace(p, state, cond)
				return 1
			}
		}
	}
	for condition := int32(0); condition < 1<<Gia_ManPiNum(p.pAig); condition++ {
		if p.pStateNew == nil {
			p.pStateNew = Gia_ManEraCreateState(p)
		}
		clear(p.pStateNew.pData)
		for i := int32(0); i < p.pAig.nRegs; i++ {
			data := Gia_ManEraData(p, Gia_ManRi(p.pAig, i).id)
			if data[condition>>5]&(1<<uint(condition&31)) != 0 {
				p.pStateNew.pData[i>>5] |= 1 << uint(i&31)
			}
		}
		var output uint32
		if dump != 0 {
			for i := int32(0); i < Gia_ManPoNum(p.pAig) && i < 32; i++ {
				data := Gia_ManEraData(p, Gia_ManPo(p.pAig, i).id)
				if data[condition>>5]&(1<<uint(condition&31)) != 0 {
					output |= 1 << uint(i)
				}
			}
		}
		next := int32(-1)
		place := Gia_ManEraHashFind(p, p.pStateNew, &next)
		if place != nil {
			next = p.pStateNew.Num
			*place = next
			p.pStateNew.Cond = condition
			p.pStateNew.iPrev = state.Num
			p.pStateNew = nil
			if int32(len(p.vStates)) > 2*p.nBins {
				Gia_ManEraHashResize(p)
			}
		}
		if dump != 0 {
			for _, x := range []int32{condition, state.Num, next, int32(output)} {
				vec.Vec_IntPush(p.vStgDump, x)
			}
		}
	}
	return 0
}
func Gia_ManCollectReachable(g *Gia_Man_t, maxStates, miter, dump, verbose int32) int32 {
	start := time.Now()
	p := Gia_ManEraCreate(g)
	state := Gia_ManEraCreateState(p)
	hash := Gia_ManEraStateHash(state.pData, p.nWordsDat, p.nBins)
	p.pBins[hash] = state.Num
	result := int32(1)
	for p.iCurState < int32(len(p.vStates))-1 {
		if int32(len(p.vStates)) >= maxStates {
			fmt.Printf("Reached the limit on states traversed (%d).  ", maxStates)
			result = -1
			break
		}
		p.iCurState++
		state = Gia_ManEraState(p, p.iCurState)
		if p.iCurState > 1 && state.iPrev == 0 {
			continue
		}
		Gia_ManInsertState(p, state)
		Gia_ManPerformOneIter(p)
		if Gia_ManAnalyzeResult(p, state, miter, dump) != 0 && miter != 0 {
			result = 0
			fmt.Printf("Miter failed in state %d after %d transitions.  ", p.iCurState, vec.Vec_IntSize(p.vBugTrace)-1)
			break
		}
		if verbose != 0 && p.iCurState%5000 == 0 {
			fmt.Printf("States =%d. Reached =%d. Depth =%d. Time = %.2f sec  ", p.iCurState, len(p.vStates), Gia_ManCountDepth(p), time.Since(start).Seconds())
		}
	}
	fmt.Printf("Reachability analysis traversed %d states with depth %d.  Time = %.2f sec\n", p.iCurState-1, Gia_ManCountDepth(p), time.Since(start).Seconds())
	if dump != 0 {
		f, err := os.Create("test.stg")
		if err != nil {
			fmt.Printf("Cannot open file %q for writing.\n", "test.stg")
		} else {
			lines := vec.Vec_IntArray(p.vStgDump)
			width := len(fmt.Sprint(p.iCurState - 1))
			for i := 0; i < len(lines); i += 4 {
				fmt.Fprintf(f, "%0*b %*d %*d %0*b\n", Gia_ManPiNum(g), uint32(lines[i]), width, lines[i+1], width, lines[i+2], Gia_ManPoNum(g), uint32(lines[i+3]))
			}
			f.Close()
			fmt.Println("Extracted STG was written into file \"test.stg\".")
		}
	}
	Gia_ManEraFree(p)
	return result
}
