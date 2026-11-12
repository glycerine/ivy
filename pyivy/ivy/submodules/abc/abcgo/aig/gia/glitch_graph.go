package gia

import "fmt"

// Gli_Obj_t uses stable object handles and Go-owned fanin/fanout slices.
type Gli_Obj_t struct {
	fTerm, fPhase, fPhase2, fMark         uint32
	nFanins, nFanouts                     int32
	Handle                                int32
	pTruth                                []word
	uSimInfo                              uint32
	iFanin, iFanout, nSwitches, nGlitches int32
	fanins, fanouts                       []*Gli_Obj_t
}
type Gli_Man_t struct {
	vCis, vCos, vCisChanged, vAffected, vFrontier *Vec_Int_t
	nObjs, nRegs, nTravIds, iObjData, nObjData    int32
	objects                                       []*Gli_Obj_t
	pSimInfoPrev                                  []uint32
}

func Gli_ManCiNum(p *Gli_Man_t) int32              { return int32(Vec_IntSize(p.vCis)) }
func Gli_ManCoNum(p *Gli_Man_t) int32              { return int32(Vec_IntSize(p.vCos)) }
func Gli_ManPiNum(p *Gli_Man_t) int32              { return Gli_ManCiNum(p) - p.nRegs }
func Gli_ManPoNum(p *Gli_Man_t) int32              { return Gli_ManCoNum(p) - p.nRegs }
func Gli_ManRegNum(p *Gli_Man_t) int32             { return p.nRegs }
func Gli_ManObjNum(p *Gli_Man_t) int32             { return p.nObjs }
func Gli_ManNodeNum(p *Gli_Man_t) int32            { return p.nObjs - Gli_ManCiNum(p) - Gli_ManCoNum(p) }
func Gli_ManObj(p *Gli_Man_t, id int32) *Gli_Obj_t { return p.objects[id] }
func Gli_ManCi(p *Gli_Man_t, id int32) *Gli_Obj_t {
	return Gli_ManObj(p, Vec_IntEntry(p.vCis, int(id)))
}
func Gli_ManCo(p *Gli_Man_t, id int32) *Gli_Obj_t {
	return Gli_ManObj(p, Vec_IntEntry(p.vCos, int(id)))
}
func Gli_ManPi(p *Gli_Man_t, id int32) *Gli_Obj_t {
	if id < 0 || id >= Gli_ManPiNum(p) {
		panic("glitch: invalid PI")
	}
	return Gli_ManCi(p, id)
}
func Gli_ManPo(p *Gli_Man_t, id int32) *Gli_Obj_t {
	if id < 0 || id >= Gli_ManPoNum(p) {
		panic("glitch: invalid PO")
	}
	return Gli_ManCo(p, id)
}
func Gli_ManRo(p *Gli_Man_t, id int32) *Gli_Obj_t {
	if id < 0 || id >= p.nRegs {
		panic("glitch: invalid RO")
	}
	return Gli_ManCi(p, Gli_ManPiNum(p)+id)
}
func Gli_ManRi(p *Gli_Man_t, id int32) *Gli_Obj_t {
	if id < 0 || id >= p.nRegs {
		panic("glitch: invalid RI")
	}
	return Gli_ManCo(p, Gli_ManPoNum(p)+id)
}
func Gli_ObjIsTerm(o *Gli_Obj_t) int32                { return int32(o.fTerm) }
func Gli_ObjIsCi(o *Gli_Obj_t) int32                  { return giaBool(o.fTerm != 0 && o.nFanins == 0) }
func Gli_ObjIsCo(o *Gli_Obj_t) int32                  { return giaBool(o.fTerm != 0 && o.nFanins == 1) }
func Gli_ObjIsNode(o *Gli_Obj_t) int32                { return giaBool(o.fTerm == 0) }
func Gli_ObjFaninNum(o *Gli_Obj_t) int32              { return o.nFanins }
func Gli_ObjFanoutNum(o *Gli_Obj_t) int32             { return o.nFanouts }
func Gli_ObjFanin(o *Gli_Obj_t, id int32) *Gli_Obj_t  { return o.fanins[id] }
func Gli_ObjFanout(o *Gli_Obj_t, id int32) *Gli_Obj_t { return o.fanouts[id] }
func Gli_ManAlloc(nObjs, nRegs, nFanioPairs int32) *Gli_Man_t {
	if nObjs < 0 || nRegs < 0 || nFanioPairs < 0 {
		panic("glitch: negative allocation")
	}
	return &Gli_Man_t{nRegs: nRegs, objects: make([]*Gli_Obj_t, 0, nObjs), vCis: Vec_IntAlloc(1000), vCos: Vec_IntAlloc(1000), vCisChanged: Vec_IntAlloc(1000), vAffected: Vec_IntAlloc(1000), vFrontier: Vec_IntAlloc(1000)}
}
func Gli_ManStop(p *Gli_Man_t) {
	if p != nil {
		*p = Gli_Man_t{}
	}
}
func Gli_ManPrintObjects(p *Gli_Man_t) {
	for _, o := range p.objects {
		fmt.Printf("Node %d\nFanins: ", o.Handle)
		for _, f := range o.fanins {
			fmt.Printf("%d ", f.Handle)
		}
		fmt.Print("\nFanouts: ")
		for _, f := range o.fanouts {
			fmt.Printf("%d ", f.Handle)
		}
		fmt.Println()
	}
}
func Gli_ManFinalize(p *Gli_Man_t) {
	for _, o := range p.objects {
		if int32(len(o.fanins)) != o.nFanins || int32(len(o.fanouts)) != o.nFanouts {
			panic("glitch: incomplete fanin/fanout links")
		}
		o.iFanin, o.iFanout = 0, 0
	}
}
func Gli_ObjAddFanin(o, fanin *Gli_Obj_t) {
	if int32(len(o.fanins)) >= o.nFanins || int32(len(fanin.fanouts)) >= fanin.nFanouts {
		panic("glitch: excess fanin/fanout")
	}
	o.fanins = append(o.fanins, fanin)
	fanin.fanouts = append(fanin.fanouts, o)
	o.iFanin++
	fanin.iFanout++
}
func Gli_ObjAlloc(p *Gli_Man_t, fanins, fanouts int32) *Gli_Obj_t {
	if fanins < 0 || fanouts < 0 {
		panic("glitch: negative fanin/fanout")
	}
	o := &Gli_Obj_t{Handle: p.nObjs, nFanins: fanins, nFanouts: fanouts, fanins: make([]*Gli_Obj_t, 0, fanins), fanouts: make([]*Gli_Obj_t, 0, fanouts)}
	p.objects = append(p.objects, o)
	p.nObjs++
	return o
}
func Gli_ManCreateCi(p *Gli_Man_t, fanouts int32) int32 {
	o := Gli_ObjAlloc(p, 0, fanouts)
	o.fTerm = 1
	Vec_IntPush(p.vCis, o.Handle)
	return o.Handle
}
func Gli_ManCreateCo(p *Gli_Man_t, fanin int32) int32 {
	o := Gli_ObjAlloc(p, 1, 0)
	o.fTerm = 1
	f := Gli_ManObj(p, fanin)
	Gli_ObjAddFanin(o, f)
	o.fPhase, o.fPhase2 = f.fPhase, f.fPhase
	Vec_IntPush(p.vCos, o.Handle)
	return o.Handle
}
func gliNodeValue(o *Gli_Obj_t, second bool) int32 {
	var phase uint32
	for i, f := range o.fanins {
		v := f.fPhase
		if second {
			v = f.fPhase2
		}
		phase |= v << uint(i)
	}
	if int(phase/64) >= len(o.pTruth) {
		panic("glitch: short truth table")
	}
	return int32(o.pTruth[phase/64] >> uint(phase%64) & 1)
}
func Gli_NodeComputeValue(o *Gli_Obj_t) int32  { return gliNodeValue(o, false) }
func Gli_NodeComputeValue2(o *Gli_Obj_t) int32 { return gliNodeValue(o, true) }
func Gli_ManCreateNode(p *Gli_Man_t, fanins *Vec_Int_t, fanouts int32, truth []word) int32 {
	if Vec_IntSize(fanins) > 16 {
		panic("glitch: too many fanins")
	}
	o := Gli_ObjAlloc(p, int32(Vec_IntSize(fanins)), fanouts)
	for _, id := range Vec_IntArray(fanins) {
		Gli_ObjAddFanin(o, Gli_ManObj(p, id))
	}
	o.pTruth = truth
	o.fPhase = uint32(Gli_NodeComputeValue(o))
	o.fPhase2 = o.fPhase
	return o.Handle
}
func Gli_ObjNumSwitches(p *Gli_Man_t, id int32) int32 { return Gli_ManObj(p, id).nSwitches }
func Gli_ObjNumGlitches(p *Gli_Man_t, id int32) int32 { return Gli_ManObj(p, id).nGlitches }
