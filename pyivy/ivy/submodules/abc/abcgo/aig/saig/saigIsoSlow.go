package saig

import (
	"fmt"
	"sort"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

// Port of src/aig/saig/saigIsoSlow.c: canonical numbering of CIs and nodes
// by iterated signature refinement.

// isoObj is Iso_Obj_t. The first three words (Level, with nFinNeg in its two
// top bits, always 0 here; FaninSig; FanoutSig) are the hashed and compared
// signature.
type isoObj struct {
	Level     uint32
	FaninSig  uint32
	FanoutSig uint32
	iNext     int32 // hash chain
	iClass    int32 // next one in the class
	Id        int32 // canonical number, 0 while unassigned
}

// isoMan is Iso_Man_t; objects are indexed by AIG object ID, 0 meaning none.
type isoMan struct {
	pAig                *aig.Graph
	pObjs               []isoObj
	nObjIds             int32
	nClasses, nEntries  int32
	nSingles            int32
	pBins               []int32
	vSingles, vClasses  []int32
	vTemp1, vTemp2      []int32
	timeHash, timeFout  time.Duration
	timeSort, timeOther time.Duration
	timeTotal           time.Duration
}

func (p *isoMan) obj(i int32) *isoObj {
	if i == 0 {
		return nil
	}
	return &p.pObjs[i]
}

func isoManStart(pAig *aig.Graph) *isoMan {
	n := aig.Aig_ManObjNumMax(pAig)
	return &isoMan{pAig: pAig, pObjs: make([]isoObj, n), pBins: make([]int32, util.Abc_PrimeCudd(uint32(n))), nObjIds: 1}
}

func (p *isoMan) stop(verbose int32) {
	if verbose != 0 {
		p.timeOther = p.timeTotal - p.timeHash - p.timeFout
		util.Abc_PrintTimeP(1, "Building ", p.timeFout.Nanoseconds(), p.timeTotal.Nanoseconds())
		util.Abc_PrintTimeP(1, "Hashing  ", (p.timeHash - p.timeSort).Nanoseconds(), p.timeTotal.Nanoseconds())
		util.Abc_PrintTimeP(1, "Sorting  ", p.timeSort.Nanoseconds(), p.timeTotal.Nanoseconds())
		util.Abc_PrintTimeP(1, "Other    ", p.timeOther.Nanoseconds(), p.timeTotal.Nanoseconds())
		util.Abc_PrintTimeP(1, "TOTAL    ", p.timeTotal.Nanoseconds(), p.timeTotal.Nanoseconds())
	}
}

// isoCompare is Iso_ObjCompare, -memcmp over the three signature words as
// little-endian bytes (a byte-wise, not numeric, order).
func isoCompare(a, b *isoObj) int {
	for _, w := range [3][2]uint32{{a.Level, b.Level}, {a.FaninSig, b.FaninSig}, {a.FanoutSig, b.FanoutSig}} {
		for k := uint(0); k < 32; k += 8 {
			x, y := byte(w[0]>>k), byte(w[1]>>k)
			if x != y {
				if x < y {
					return 1
				}
				return -1
			}
		}
	}
	return 0
}

func isoHash(o *isoObj, nBins int) int {
	const p0, p1, p2 = 12582917, 25165843, 50331653
	v := p0*o.Level ^ p1*o.FaninSig ^ p2*o.FanoutSig
	return int(v % uint32(nBins))
}

// hashAdd is Iso_ObjHashAdd: the object joins the class of an equal
// signature (right after its head) or starts a chain entry.
func (p *isoMan) hashAdd(id int32) bool {
	o := &p.pObjs[id]
	place := &p.pBins[isoHash(o, len(p.pBins))]
	p.nEntries++
	for this := p.obj(*place); this != nil; this = p.obj(*place) {
		if isoCompare(this, o) == 0 {
			if this.iClass == 0 {
				p.nClasses++
				p.nSingles--
			}
			o.iClass = this.iClass
			this.iClass = id
			return true
		}
		place = &this.iNext
	}
	*place = id
	p.nSingles++
	return false
}

func (p *isoMan) sortByCompare(ids []int32) {
	sort.Slice(ids, func(i, j int) bool { return isoCompare(&p.pObjs[ids[i]], &p.pObjs[ids[j]]) < 0 })
}

// collectClasses is Iso_ManCollectClasses.
func (p *isoMan) collectClasses() {
	p.vSingles, p.vClasses = p.vSingles[:0], p.vClasses[:0]
	for _, head := range p.pBins {
		for id := head; id != 0; id = p.pObjs[id].iNext {
			if p.pObjs[id].Id != 0 {
				panic("saig: iso object already numbered")
			}
			if p.pObjs[id].iClass != 0 {
				p.vClasses = append(p.vClasses, id)
			} else {
				p.vSingles = append(p.vSingles, id)
			}
		}
	}
	start := time.Now()
	p.sortByCompare(p.vSingles)
	p.sortByCompare(p.vClasses)
	p.timeSort += time.Since(start)
	if int32(len(p.vSingles)) != p.nSingles || int32(len(p.vClasses)) != p.nClasses {
		panic("saig: iso class counts disagree")
	}
	for _, id := range p.vSingles {
		if p.pObjs[id].Id == 0 {
			p.pObjs[id].Id = p.nObjIds
			p.nObjIds++
		}
	}
}

func isoLit(v uint32, c int32) uint32 { return 2*v + uint32(c) }

// objIDs lists the AIG's live object IDs in ID order (Aig_ManForEachObj).
func isoObjIDs(g *aig.Graph) []int32 {
	ids := make([]int32, 0, aig.Aig_ManObjNumMax(g))
	for i := int32(0); i < aig.Aig_ManObjNumMax(g); i++ {
		if !aig.Aig_ObjIsNone(g, aig.Aig_ManObj(g, i)) {
			ids = append(ids, i)
		}
	}
	return ids
}

// isoManCreate is Iso_ManCreate: level-weighted TFI and TFO signatures.
func isoManCreate(pAig *aig.Graph) *isoMan {
	p := isoManStart(pAig)
	ids := isoObjIDs(pAig)
	for _, i := range ids {
		e := aig.Aig_ManObj(pAig, i)
		if aig.Aig_ObjIsCo(pAig, e) {
			continue
		}
		o := &p.pObjs[i]
		o.Level = uint32(aig.Aig_ObjLevel(pAig, e))
		if aig.Aig_ObjIsNode(pAig, e) {
			f := &p.pObjs[aig.Aig_ObjFaninId0(pAig, e)]
			o.FaninSig += f.FaninSig
			o.FaninSig += o.Level * isoPrimes[isoLit(o.Level, aig.Aig_ObjFaninC0(pAig, e))&isoMask]
			f = &p.pObjs[aig.Aig_ObjFaninId1(pAig, e)]
			o.FaninSig += f.FaninSig
			o.FaninSig += o.Level * isoPrimes[isoLit(o.Level, aig.Aig_ObjFaninC1(pAig, e))&isoMask]
		}
	}
	for k := len(ids) - 1; k >= 0; k-- {
		i := ids[k]
		e := aig.Aig_ManObj(pAig, i)
		if aig.Aig_ObjIsCi(pAig, e) || aig.Aig_ObjIsConst1(pAig, e) {
			continue
		}
		o := &p.pObjs[i]
		if aig.Aig_ObjIsNode(pAig, e) {
			f := &p.pObjs[aig.Aig_ObjFaninId0(pAig, e)]
			f.FanoutSig += o.FanoutSig
			f.FanoutSig += o.Level * isoPrimes[isoLit(o.Level, aig.Aig_ObjFaninC0(pAig, e))&isoMask]
			f = &p.pObjs[aig.Aig_ObjFaninId1(pAig, e)]
			f.FanoutSig += o.FanoutSig
			f.FanoutSig += o.Level * isoPrimes[isoLit(o.Level, aig.Aig_ObjFaninC1(pAig, e))&isoMask]
		} else if aig.Aig_ObjIsCo(pAig, e) {
			f := &p.pObjs[aig.Aig_ObjFaninId0(pAig, e)]
			f.FanoutSig += o.FanoutSig
			f.FanoutSig += o.Level * isoPrimes[isoLit(o.Level, aig.Aig_ObjFaninC0(pAig, e))&isoMask]
		}
	}
	for r := int32(0); r < aig.Aig_ManRegNum(pAig); r++ {
		li, lo := Saig_ManLi(pAig, r), Saig_ManLo(pAig, r)
		if aig.Aig_ObjFaninId0(pAig, li) == 0 { // ignore constant
			continue
		}
		o, f := &p.pObjs[lo.NodeID()], &p.pObjs[aig.Aig_ObjFaninId0(pAig, li)]
		if o.FaninSig != 0 {
			panic("saig: flop output has a fanin signature")
		}
		o.FaninSig = f.FaninSig
		f.FanoutSig += o.FanoutSig
	}
	for _, i := range ids {
		e := aig.Aig_ManObj(pAig, i)
		if aig.Aig_ObjIsCi(pAig, e) || aig.Aig_ObjIsNode(pAig, e) {
			p.hashAdd(i)
		}
	}
	p.collectClasses()
	return p
}

// assignAdjacency is Iso_ManAssignAdjacency: signatures from the canonical
// numbers assigned so far.
func (p *isoMan) assignAdjacency() {
	g := p.pAig
	ids := isoObjIDs(g)
	for _, i := range ids {
		e := aig.Aig_ManObj(g, i)
		o := &p.pObjs[i]
		o.FaninSig, o.FanoutSig = 0, 0
		if aig.Aig_ObjIsCo(g, e) || !aig.Aig_ObjIsNode(g, e) {
			continue
		}
		for k, fid := range [2]int32{aig.Aig_ObjFaninId0(g, e), aig.Aig_ObjFaninId1(g, e)} {
			c := aig.Aig_ObjFaninC0(g, e)
			if k == 1 {
				c = aig.Aig_ObjFaninC1(g, e)
			}
			f := &p.pObjs[fid]
			o.FaninSig += f.FaninSig
			if f.Id != 0 {
				o.FaninSig += uint32(f.Id) * isoPrimes[isoLit(uint32(f.Id), c)&isoMask]
			}
		}
	}
	for k := len(ids) - 1; k >= 0; k-- {
		i := ids[k]
		e := aig.Aig_ManObj(g, i)
		if aig.Aig_ObjIsCi(g, e) || aig.Aig_ObjIsConst1(g, e) {
			continue
		}
		o := &p.pObjs[i]
		var fanins []int32
		var compl []int32
		if aig.Aig_ObjIsNode(g, e) {
			fanins = []int32{aig.Aig_ObjFaninId0(g, e), aig.Aig_ObjFaninId1(g, e)}
			compl = []int32{aig.Aig_ObjFaninC0(g, e), aig.Aig_ObjFaninC1(g, e)}
		} else if aig.Aig_ObjIsCo(g, e) {
			fanins = []int32{aig.Aig_ObjFaninId0(g, e)}
			compl = []int32{aig.Aig_ObjFaninC0(g, e)}
		}
		for j, fid := range fanins {
			f := &p.pObjs[fid]
			f.FanoutSig += o.FanoutSig
			if o.Id != 0 {
				f.FanoutSig += uint32(o.Id) * isoPrimes[isoLit(uint32(o.Id), compl[j])&isoMask]
			}
		}
	}
	for r := int32(0); r < aig.Aig_ManRegNum(g); r++ {
		li, lo := Saig_ManLi(g, r), Saig_ManLo(g, r)
		if aig.Aig_ObjFaninId0(g, li) == 0 { // ignore constant
			continue
		}
		o, f := &p.pObjs[lo.NodeID()], &p.pObjs[aig.Aig_ObjFaninId0(g, li)]
		c := aig.Aig_ObjFaninC0(g, li)
		o.FaninSig = f.FaninSig
		if f.Id != 0 {
			o.FaninSig += uint32(f.Id) * isoPrimes[isoLit(uint32(f.Id), c)&isoMask]
		}
		f.FanoutSig += o.FanoutSig
		if o.Id != 0 {
			f.FanoutSig += uint32(o.Id) * isoPrimes[isoLit(uint32(o.Id), c)&isoMask]
		}
	}
}

func (p *isoMan) printClasses(verbose int32) {
	if verbose != 0 {
		fmt.Printf("Total objects =%7d.  Entries =%7d.  Classes =%7d.  Singles =%7d.\n", len(p.pObjs), p.nEntries, p.nClasses, p.nSingles)
	}
}

// rehashClassNodes is Iso_ManRehashClassNodes: the unnumbered members of
// the classes are hashed again with their new signatures.
func (p *isoMan) rehashClassNodes() {
	p.vTemp1, p.vTemp2 = p.vTemp1[:0], p.vTemp2[:0]
	for _, head := range p.vClasses {
		for id := head; id != 0; id = p.pObjs[id].iClass {
			if p.pObjs[id].Id == 0 {
				p.vTemp1 = append(p.vTemp1, id)
			} else {
				p.vTemp2 = append(p.vTemp2, id)
			}
		}
	}
	p.nClasses, p.nEntries, p.nSingles = 0, 0, 0
	clear(p.pBins)
	for _, id := range p.vTemp1 {
		p.pObjs[id].iClass, p.pObjs[id].iNext = 0, 0
		p.hashAdd(id)
	}
	for _, id := range p.vTemp2 {
		p.pObjs[id].iClass, p.pObjs[id].iNext = 0, 0
	}
	p.collectClasses()
}

// breakTies is Iso_ManBreakTies: every member of the classes on the top
// signature level gets the next number.
func (p *isoMan) breakTies(verbose int32) {
	levelStart := int32(p.pObjs[p.vClasses[0]].Level)
	if verbose != 0 {
		fmt.Printf("Best level %d\n", levelStart)
	}
	for _, head := range p.vClasses {
		if int32(p.pObjs[head].Level) < levelStart {
			break
		}
		for id := head; id != 0; id = p.pObjs[id].iClass {
			if p.pObjs[id].Id != 0 {
				panic("saig: iso tie member already numbered")
			}
			p.pObjs[id].Id = p.nObjIds
			p.nObjIds++
		}
	}
}

// finalize is Iso_ManFinalize: the canonical number of every object
// (iData: 0 for the constant, -1 for COs) and the CIs, PIs then flops,
// in canonical order.
func (p *isoMan) finalize() (perm, iData []int32) {
	g := p.pAig
	if p.nClasses != 0 || len(p.vClasses) != 0 {
		panic("saig: iso classes remain")
	}
	iData = make([]int32, aig.Aig_ManObjNumMax(g))
	for _, i := range isoObjIDs(g) {
		e := aig.Aig_ManObj(g, i)
		if !aig.Aig_ObjIsCi(g, e) && !aig.Aig_ObjIsNode(g, e) {
			iData[i] = -1
			continue
		}
		iData[i] = p.pObjs[i].Id
	}
	iData[aig.Aig_ManConst1(g).NodeID()] = 0
	nPis := aig.Aig_ManCiNum(g) - aig.Aig_ManRegNum(g)
	var pis, flops []int32
	for i := int32(0); i < aig.Aig_ManCiNum(g); i++ {
		id := int32(aig.Aig_ManCi(g, i).NodeID())
		if iData[id] <= 0 {
			panic("saig: iso CI without a number")
		}
		if i >= nPis {
			flops = append(flops, i)
		} else {
			pis = append(pis, i)
		}
	}
	byData := func(cis []int32) {
		sort.Slice(cis, func(a, b int) bool {
			return iData[aig.Aig_ManCi(g, cis[a]).NodeID()] < iData[aig.Aig_ManCi(g, cis[b]).NodeID()]
		})
	}
	byData(pis)
	byData(flops)
	return append(pis, flops...), iData
}

// isoFindPerm is Saig_ManFindIsoPerm, also returning the objects' canonical
// numbers (C leaves them in pObj->iData).
func isoFindPerm(pAig *aig.Graph, verbose int32) (perm, iData []int32) {
	start := time.Now()
	p := isoManCreate(pAig)
	p.timeFout += time.Since(start)
	p.printClasses(verbose)
	refine := func() {
		clk := time.Now()
		p.assignAdjacency()
		p.timeFout += time.Since(clk)
		clk = time.Now()
		p.rehashClassNodes()
		p.timeHash += time.Since(clk)
		p.printClasses(verbose)
	}
	for p.nClasses != 0 {
		refine()
		for p.nSingles == 0 && p.nClasses != 0 {
			p.breakTies(verbose)
			refine()
		}
	}
	p.timeTotal = time.Since(start)
	if p.nObjIds != 1+aig.Aig_ManCiNum(pAig)+aig.Aig_ManNodeNum(pAig) {
		panic("saig: iso numbering incomplete")
	}
	perm, iData = p.finalize()
	p.stop(verbose)
	return perm, iData
}

// Saig_ManFindIsoPerm returns the CIs (PIs, then flops) in canonical order.
func Saig_ManFindIsoPerm(pAig *aig.Graph, fVerbose int32) []int32 {
	perm, _ := isoFindPerm(pAig, fVerbose)
	return perm
}

// Iso_ManObjCount counts the nodes in the cone of root (CIs excluded) and
// their complemented fanin edges.
func Iso_ManObjCount(p *aig.Graph, root aig.Edge) (nodes, complementedEdges int32) {
	if !aig.Aig_ObjIsNode(p, root) {
		panic("saig: Iso_ManObjCount needs a node")
	}
	aig.Aig_ManIncrementTravId(p)
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		if aig.Aig_ObjIsCi(p, e) || aig.Aig_ObjIsTravIdCurrent(p, e) {
			return
		}
		aig.Aig_ObjSetTravIdCurrent(p, e)
		visit(aig.Aig_ObjFanin0(p, e))
		visit(aig.Aig_ObjFanin1(p, e))
		complementedEdges += aig.Aig_ObjFaninC0(p, e) + aig.Aig_ObjFaninC1(p, e)
		nodes++
	}
	visit(aig.Aig_Regular(root))
	return
}
