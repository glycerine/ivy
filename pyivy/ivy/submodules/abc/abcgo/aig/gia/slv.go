package gia

import "fmt"

// Slv_Man_t retains the experimental solver's literal and fanout indexing.
// Explicit kinds disambiguate a constant-driven CO from the zero/zero CI marker.
type Slv_Man_t struct {
	vCis, vCos, vFanins, vFanoutN, vFanout0, vCopies []int32
	vValues                                          []byte
	kinds                                            []graphKind
	nRegs                                            int32
}

func Slv_ManObjNum(p *Slv_Man_t) int32                   { return int32(len(p.vFanins) / 2) }
func Slv_ObjFaninLit(p *Slv_Man_t, id, f int32) int32    { return p.vFanins[2*id+f] }
func Slv_ObjFanin(p *Slv_Man_t, id, f int32) int32       { return Slv_ObjFaninLit(p, id, f) >> 1 }
func Slv_ObjFaninC(p *Slv_Man_t, id, f int32) int32      { return Slv_ObjFaninLit(p, id, f) & 1 }
func Slv_ObjIsCi(p *Slv_Man_t, id int32) int32           { return giaBool(p.kinds[id] == graphInput) }
func Slv_ObjIsCo(p *Slv_Man_t, id int32) int32           { return giaBool(p.kinds[id] == graphOutput) }
func Slv_ObjIsAnd(p *Slv_Man_t, id int32) int32          { return giaBool(p.kinds[id] == graphAnd) }
func Slv_ObjFanout0(p *Slv_Man_t, id int32) int32        { return p.vFanout0[id] }
func Slv_ObjSetFanout0(p *Slv_Man_t, id, lit int32)      { p.vFanout0[id] = lit }
func Slv_ObjNextFanout(p *Slv_Man_t, lit int32) int32    { return p.vFanoutN[lit] }
func Slv_ObjSetNextFanout(p *Slv_Man_t, lit, next int32) { p.vFanoutN[lit] = next }
func Slv_ObjCopyLit(p *Slv_Man_t, id int32) int32        { return p.vCopies[id] }
func Slv_ObjSetCopyLit(p *Slv_Man_t, id, lit int32)      { p.vCopies[id] = lit }

const graphOutput graphKind = 3

func Slv_ManAlloc(capacity int32) *Slv_Man_t {
	if capacity < 1 {
		capacity = 1
	}
	return &Slv_Man_t{vFanins: make([]int32, 2, 2*int(capacity)), vFanoutN: make([]int32, 2, 2*int(capacity)), vFanout0: make([]int32, 1, int(capacity)), vValues: []byte{0}, kinds: []graphKind{graphConstant}}
}
func Slv_ManFree(p *Slv_Man_t) {
	if p != nil {
		*p = Slv_Man_t{}
	}
}
func Slv_ManPrintFanouts(p *Slv_Man_t) {
	for id := int32(1); id < Slv_ManObjNum(p); id++ {
		fmt.Printf("Fanouts of node %d: ", id)
		for lit := Slv_ObjFanout0(p, id); lit != 0; lit = Slv_ObjNextFanout(p, lit) {
			fmt.Printf("%d ", lit>>1)
		}
		fmt.Println()
	}
}
func slvAppend(p *Slv_Man_t, a, b int32, kind graphKind) int32 {
	id := Slv_ManObjNum(p)
	if a < 0 || b < 0 || a>>1 >= id || b>>1 >= id {
		panic("gia: invalid solver fanins")
	}
	p.vValues = append(p.vValues, 0)
	p.vFanout0 = append(p.vFanout0, 0)
	p.vFanoutN = append(p.vFanoutN, 0, 0)
	p.kinds = append(p.kinds, kind)
	switch kind {
	case graphInput:
		b = int32(len(p.vCis)) + 1
		p.vCis = append(p.vCis, id)
	case graphOutput:
		p.vCos = append(p.vCos, id)
	case graphAnd:
		p.vFanoutN[id<<1] = p.vFanout0[a>>1]
		p.vFanout0[a>>1] = id << 1
		p.vFanoutN[id<<1|1] = p.vFanout0[b>>1]
		p.vFanout0[b>>1] = id<<1 | 1
	}
	p.vFanins = append(p.vFanins, a, b)
	return id << 1
}
func Slv_ManAppendObj(p *Slv_Man_t, a, b int32) int32 {
	kind := graphAnd
	if a == 0 {
		kind = graphInput
	} else if b == 0 {
		kind = graphOutput
	}
	return slvAppend(p, a, b, kind)
}
func Slv_ManFromGia(p *Gia_Man_t) *Slv_Man_t {
	out := Slv_ManAlloc(p.nObjs)
	out.nRegs = p.nRegs
	copies := make([]int32, p.nObjs)
	for id := int32(1); id < p.nObjs; id++ {
		obj := p.pObjs[id]
		edge := func(diff, compl uint32) int32 { return copies[id-int32(diff)] ^ int32(compl) }
		switch {
		case Gia_ObjIsCi(obj) != 0:
			copies[id] = slvAppend(out, 0, 0, graphInput)
		case Gia_ObjIsCo(obj) != 0:
			copies[id] = slvAppend(out, edge(obj.iDiff0, obj.fCompl0), 0, graphOutput)
		case Gia_ObjIsAnd(obj) != 0:
			if Gia_ObjIsXor(obj) != 0 || (int(id) < len(p.PMuxes) && p.PMuxes[id] != 0) {
				panic("gia: normalize solver input XOR/MUX nodes")
			}
			copies[id] = slvAppend(out, edge(obj.iDiff0, obj.fCompl0), edge(obj.iDiff1, obj.fCompl1), graphAnd)
		default:
			panic("gia: unsupported solver object")
		}
		obj.Value = uint32(copies[id])
	}
	p.pObjs[0].Value = 0
	return out
}
func Slv_ManToGia(p *Slv_Man_t) *Gia_Man_t {
	out := Gia_ManStart(Slv_ManObjNum(p))
	out.fGiaSimple = 1
	p.vCopies = make([]int32, Slv_ManObjNum(p))
	for id := int32(1); id < Slv_ManObjNum(p); id++ {
		edge := func(f int32) int32 { lit := Slv_ObjFaninLit(p, id, f); return p.vCopies[lit>>1] ^ (lit & 1) }
		switch p.kinds[id] {
		case graphInput:
			p.vCopies[id] = Gia_ManAppendCi(out)
		case graphOutput:
			p.vCopies[id] = Gia_ManAppendCo(out, edge(0))
		case graphAnd:
			p.vCopies[id] = Gia_ManAppendAnd(out, edge(0), edge(1))
		default:
			panic("gia: unsupported solver object")
		}
	}
	out.nRegs = p.nRegs
	out.fGiaSimple = 0
	return out
}
func Slv_ManToAig(p *Gia_Man_t) *Gia_Man_t {
	solver := Slv_ManFromGia(p)
	out := Slv_ManToGia(solver)
	out.pName = p.pName
	out.PSpec = p.PSpec
	Slv_ManPrintFanouts(solver)
	Slv_ManFree(solver)
	return out
}
