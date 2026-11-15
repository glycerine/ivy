package gia

import (
	"fmt"
	"time"
)

// Gia_ManUnr_t owns the source dependency tents and rolling literal histories.
// pNew is returned to the caller and survives stopping this manager.
type Gia_ManUnr_t struct {
	pAig, pOrder, pNew *Gia_Man_t
	pPars              *Gia_ParFra_t
	limits, delays     []int32
	first              []int
	store              []int32
	frame, LastLit     int32
}

func (p *Gia_ManUnr_t) edgeWidth() int {
	if p.pOrder.nMuxes != 0 {
		return 3
	}
	return 2
}

// Gia_ManUnrStart assigns source ranks, edge delays and bounded history rows.
func Gia_ManUnrStart(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_ManUnr_t {
	if src == nil || params == nil {
		return nil
	}
	started := time.Now()
	limits := Vec_IntAlloc(0)
	p := &Gia_ManUnr_t{pAig: src, pPars: params}
	p.pOrder = Gia_ManUnrollDup(src, limits)
	p.limits = Vec_IntArray(limits)
	n := int(p.pOrder.nObjs)
	rank, degree := make([]int32, n), make([]int32, n)
	r := 0
	for i := range rank {
		if r < len(p.limits) && p.limits[r] == int32(i) {
			r++
		}
		rank[i] = int32(r)
	}
	width := p.edgeWidth()
	p.delays = make([]int32, width*n)
	for _, obj := range p.pOrder.pObjs[1:] {
		// Logical inputs are in stored data-pin order, followed by control.
		fanins, count := giaQuantFanins(p.pOrder, obj)
		for pin, id := range fanins[:count] {
			delay := rank[obj.id] - rank[id]
			p.delays[width*int(obj.id)+pin] = delay
			degree[id] = max(degree[id], delay)
		}
	}
	p.first = make([]int, n)
	total := 0
	for i, d := range degree {
		p.first[i] = total
		total += 2 + int(d)
	}
	p.store = make([]int32, total)
	for i := range p.store {
		p.store[i] = -1
	}
	for i, d := range degree {
		p.store[p.first[i]] = 1 + d
	}
	if params.FVerbose != 0 {
		fmt.Printf("Convergence = %d.  Dangling objects = %d.  Average degree = %.3f   Time =%9.2f sec\n", len(p.limits)-1, src.nObjs-p.pOrder.nObjs, float64(len(p.store))/float64(n)-1, time.Since(started).Seconds())
	}
	return p
}

func (p *Gia_ManUnr_t) write(id, entry int32) {
	start := p.first[id]
	count := int(p.store[start])
	copy(p.store[start+2:start+count+1], p.store[start+1:start+count])
	p.store[start+1] = entry
}
func (p *Gia_ManUnr_t) read(id, delay int32) int32 {
	if id == 0 {
		return 0
	}
	start := p.first[id]
	if delay < 0 || delay >= p.store[start] {
		panic("gia: invalid unrolling history delay")
	}
	if delay > 0 {
		delay--
	}
	return p.store[start+int(delay)+1]
}
func (p *Gia_ManUnr_t) readCopy(obj *Gia_Obj_t, pin int) int32 {
	var lit int32
	switch pin {
	case 0:
		lit = Gia_ObjFaninLit0(obj, obj.id)
	case 1:
		lit = Gia_ObjFaninLit1(obj, obj.id)
	case 2:
		lit = int32(p.pOrder.PMuxes[obj.id])
	default:
		panic("gia: invalid unrolling pin")
	}
	return p.read(lit>>1, p.delays[p.edgeWidth()*int(obj.id)+pin]) ^ (lit & 1)
}
func (p *Gia_ManUnr_t) readCI(obj *Gia_Obj_t, frame int32) int32 {
	original := p.pAig.pObjs[obj.Value]
	nPis := Gia_ManPiNum(p.pAig)
	index := Gia_ObjCioId(original)
	if Gia_ObjIsPi(p.pAig, original) != 0 {
		index += nPis * frame
		// C's unconditional RegNum offset is wrong in initialized save mode:
		// those symbolic initial-state inputs were never allocated.
		if p.pPars.fSaveLastLit != 0 && p.pPars.FInit == 0 {
			index += p.pAig.nRegs
		}
		return Gia_ManCiLit(p.pNew, index)
	}
	if frame == 0 {
		if p.pPars.FInit != 0 {
			return 0
		}
		index -= nPis
		if p.pPars.fSaveLastLit == 0 {
			index += nPis * p.pPars.NFrames
		}
		return Gia_ManCiLit(p.pNew, index)
	}
	ri := Gia_ObjRoToRi(p.pAig, original)
	return p.read(int32(ri.Value>>1), 0)
}

func Gia_ManUnrollStart(src *Gia_Man_t, params *Gia_ParFra_t) any {
	if src == nil || params == nil || params.NFrames < 0 {
		return nil
	}
	reserve := int64(max(1, params.NFrames))
	if params.fSaveLastLit != 0 {
		reserve = 1
	}
	capacity := int64(src.nObjs) * reserve
	if capacity > int64(objectNoFanin) {
		return nil
	}
	p := Gia_ManUnrStart(src, params)
	p.pNew = Gia_ManStart(int32(max(int64(1), capacity)))
	p.pNew.CopyIdentityFrom(src)
	Gia_ManHashAlloc(p.pNew)
	if params.fSaveLastLit == 0 {
		for i := int32(0); i < params.NFrames*Gia_ManPiNum(src); i++ {
			Gia_ManAppendCi(p.pNew)
		}
	}
	if params.FInit == 0 {
		for i := int32(0); i < src.nRegs; i++ {
			Gia_ManAppendCi(p.pNew)
		}
	}
	return p
}

func (p *Gia_ManUnr_t) addStep(horizon int32) {
	if p.pPars.fSaveLastLit != 0 {
		for i := int32(0); i < Gia_ManPiNum(p.pAig); i++ {
			Gia_ManAppendCi(p.pNew)
		}
	}
	for frame := int32(0); frame < horizon; frame++ {
		tent := int(horizon - frame)
		if tent >= len(p.limits) {
			continue
		}
		for id := p.limits[tent-1]; id < p.limits[tent]; id++ {
			obj := p.pOrder.pObjs[id]
			var lit int32
			switch {
			case Gia_ObjIsAnd(obj) != 0:
				a := p.readCopy(obj, 0)
				switch {
				case Gia_ObjIsMux(p.pOrder, obj) != 0:
					lit = Gia_ManHashMux(p.pNew, p.readCopy(obj, 2), p.readCopy(obj, 1), a)
				case Gia_ObjIsXor(obj) != 0:
					lit = Gia_ManHashXor(p.pNew, a, p.readCopy(obj, 1))
				case Gia_ObjIsBuf(obj) != 0:
					lit = a
				default:
					lit = Gia_ManHashAnd(p.pNew, a, p.readCopy(obj, 1))
				}
			case Gia_ObjIsCo(obj) != 0:
				lit = p.readCopy(obj, 0)
				if frame == horizon-1 {
					if p.pPars.fSaveLastLit != 0 {
						p.LastLit = lit
					} else {
						Gia_ManAppendCo(p.pNew, lit)
					}
				}
			case Gia_ObjIsCi(obj) != 0:
				lit = p.readCI(obj, frame)
			default:
				panic("gia: invalid unrolling object")
			}
			if lit < 0 {
				panic("gia: unavailable unrolling history")
			}
			p.write(id, lit)
		}
	}
}

// Gia_ManUnrollAdd executes missing source scheduler steps in increasing order.
func Gia_ManUnrollAdd(raw any, horizon int32) any {
	p, ok := raw.(*Gia_ManUnr_t)
	if !ok || p == nil || p.pAig == nil || p.pNew == nil || horizon < p.frame || (p.pPars.NFrames > 0 && horizon > p.pPars.NFrames) {
		return nil
	}
	for p.frame < horizon {
		p.addStep(p.frame + 1)
		p.frame++
	}
	return p.pNew
}

// Gia_ManUnrollLastLit reads save mode's last encountered property literal.
func Gia_ManUnrollLastLit(raw any) int32 {
	p, ok := raw.(*Gia_ManUnr_t)
	if !ok || p == nil || p.pAig == nil {
		return -1
	}
	return p.LastLit
}

func Gia_ManUnrollStop(raw any) {
	p, ok := raw.(*Gia_ManUnr_t)
	if !ok || p == nil || p.pAig == nil {
		return
	}
	Gia_ManStop(p.pOrder)
	p.pOrder, p.pAig, p.pPars, p.pNew = nil, nil, nil, nil
	p.limits, p.delays, p.first, p.store = nil, nil, nil, nil
}

func Gia_ManUnroll(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	raw := Gia_ManUnrollStart(src, params)
	if raw == nil {
		return nil
	}
	p := raw.(*Gia_ManUnr_t)
	Gia_ManUnrollAdd(p, params.NFrames)
	Gia_ManHashStop(p.pNew)
	Gia_ManSetRegNum(p.pNew, 0)
	out := Gia_ManCleanup(p.pNew)
	Gia_ManStop(p.pNew)
	Gia_ManUnrollStop(p)
	return out
}

func Gia_ManFrames2(src *Gia_Man_t, params *Gia_ParFra_t) *Gia_Man_t {
	if src == nil || params == nil {
		return nil
	}
	started := time.Now()
	out := Gia_ManUnroll(src, params)
	if params.FVerbose != 0 {
		fmt.Printf("Time =%9.2f sec\n", time.Since(started).Seconds())
	}
	return out
}
