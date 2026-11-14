package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

// This file ports the built abcIfif.c experiment. Delay columns represent
// structure degree and are indexed by object ID, matching the source manager.
const IFIF_MAX_LEAVES = 6

type Abc_IffObj_t struct{ Delay [IFIF_MAX_LEAVES + 1]float32 }
type Abc_IffMan_t struct {
	pNtk  *Abc_Ntk_t
	pPars *Ifif_Par_t
	pObjs []Abc_IffObj_t
}

func Abc_IffObj(p *Abc_IffMan_t, i int32) *Abc_IffObj_t {
	assert(i >= 0 && int(i) < len(p.pObjs), "i >= 0 && i < len(p.pObjs)")
	return &p.pObjs[i]
}
func Abc_IffDelay(p *Abc_IffMan_t, obj *Abc_Obj_t, degree int32) float32 {
	return Abc_IffObj(p, abc.Abc_ObjId(obj)).Delay[degree]
}

func Abc_NtkIfifStart(ntk *Abc_Ntk_t, pars *Ifif_Par_t) *Abc_IffMan_t {
	return &Abc_IffMan_t{pNtk: ntk, pPars: pars, pObjs: make([]Abc_IffObj_t, abc.Abc_NtkObjNumMax(ntk))}
}
func Abc_NtkIfifStop(p *Abc_IffMan_t) {
	*p = Abc_IffMan_t{}
}

// Abc_ObjSortByDelay collects fanins in decreasing delay-plus-pin-delay order.
// CI fanins are appended without bubbling at that insertion, as in abcIfif.c.
func Abc_ObjSortByDelay(p *Abc_IffMan_t, obj *Abc_Obj_t, degree int32, nodes []*Abc_Obj_t) {
	count := int(abc.Abc_ObjFaninNum(obj))
	assert(count <= len(nodes) && count <= IFIF_MAX_LEAVES, "fanin count fits IFIF scratch array")
	for i := 0; i < count; i++ {
		fanin := abc.Abc_ObjFanin(obj, int32(i))
		nodes[i] = fanin
		if abc.Abc_ObjIsCi(fanin) != 0 {
			continue
		}
		for a := i; a > 0; a-- {
			left := Abc_IffDelay(p, nodes[a-1], degree) + ifpkg.IfifPinDelay(p.pPars, int32(a-1))
			right := Abc_IffDelay(p, nodes[a], degree) + ifpkg.IfifPinDelay(p.pPars, int32(a))
			if left < right {
				nodes[a-1], nodes[a] = nodes[a], nodes[a-1]
			}
		}
	}
}

// Abc_ObjDelay0 computes the node's standalone LUT delay.
func Abc_ObjDelay0(p *Abc_IffMan_t, obj *Abc_Obj_t) float32 {
	var delay float32
	var nodes [IFIF_MAX_LEAVES]*Abc_Obj_t
	Abc_ObjSortByDelay(p, obj, 1, nodes[:])
	for i := int32(0); i < abc.Abc_ObjFaninNum(obj); i++ {
		delay = max(delay, Abc_IffDelay(p, nodes[i], 1)+ifpkg.IfifPinDelay(p.pPars, i))
	}
	return delay
}

// Abc_ObjDelay1 computes the clustered structure delay from abcIfif.c.
func Abc_ObjDelay1(p *Abc_IffMan_t, obj *Abc_Obj_t) float32 {
	if abc.Abc_ObjFaninNum(obj) == 0 {
		return 0
	}
	var nodes [IFIF_MAX_LEAVES]*Abc_Obj_t
	Abc_ObjSortByDelay(p, obj, 1, nodes[:])
	params := ifpkg.IfifReadParams(p.pPars)
	var delay float32
	for i := int32(0); i < abc.Abc_ObjFaninNum(obj); i++ {
		pinDelay := ifpkg.IfifPinDelay(p.pPars, i)
		candidate := Abc_IffDelay(p, nodes[i], 1) + pinDelay
		if i < params.Degree {
			candidate = min(candidate, Abc_IffDelay(p, nodes[i], 0)+pinDelay-params.WireDelay)
		}
		delay = max(delay, candidate)
	}
	assert(delay > 0, "Delay1 > 0")
	return delay
}

// Abc_ObjDelayDegree computes the cascade delay at degree d.
func Abc_ObjDelayDegree(p *Abc_IffMan_t, obj *Abc_Obj_t, d int32) float32 {
	params := ifpkg.IfifReadParams(p.pPars)
	assert(d >= 0 && d <= params.Degree, "d >= 0 && d <= p->pPars->nDegree")
	var nodes [IFIF_MAX_LEAVES]*Abc_Obj_t
	Abc_ObjSortByDelay(p, obj, params.Degree, nodes[:])
	var delay float32
	for i := int32(0); i < abc.Abc_ObjFaninNum(obj); i++ {
		pinDelay := ifpkg.IfifPinDelay(p.pPars, i)
		candidate := Abc_IffDelay(p, nodes[i], params.Degree) + pinDelay
		if i == 0 && d > 0 {
			candidate = min(candidate, Abc_IffDelay(p, nodes[i], d-1)+pinDelay-params.WireDelay)
		}
		delay = max(delay, candidate)
	}
	return delay
}

func Abc_NtkPerformIfif(ntk *Abc_Ntk_t, pars *Ifif_Par_t) {
	params := ifpkg.IfifReadParams(pars)
	lib := ifpkg.IfifReadLibrary(pars)
	assert(lib != nil && lib.LutMax >= 0 && lib.LutMax <= IFIF_MAX_LEAVES, "valid IFIF LUT library size")
	assert(params.LUTSize >= 0 && params.LUTSize <= IFIF_MAX_LEAVES, "valid IFIF LUT size")
	assert(params.Degree >= 0 && params.Degree <= IFIF_MAX_LEAVES, "valid IFIF degree")
	abc.Abc_NtkToAig(ntk)
	abc.Abc_NtkLevel(ntk)
	if params.Verbose != 0 {
		fmt.Printf("Running mapper into LUT structures with the following parameters:\nPin+Wire: {")
		for i := int32(0); i < lib.LutMax; i++ {
			fmt.Printf(" %3.2f", ifpkg.IfifPinDelay(pars, i))
		}
		fmt.Printf(" }  Wire %3.2f  Degree %d  Type: %s\n", params.WireDelay, params.Degree, ififType(params.Cascade))
	}
	manager := Abc_NtkIfifStart(ntk, pars)
	nodes := abc.Abc_NtkDfs(ntk, 0)
	for _, obj := range nodes {
		assert(abc.Abc_ObjIsNode(obj) != 0, "Abc_ObjIsNode(pObj)")
		entry := Abc_IffObj(manager, abc.Abc_ObjId(obj))
		if params.Cascade != 0 {
			for d := int32(0); d <= params.Degree; d++ {
				entry.Delay[d] = Abc_ObjDelayDegree(manager, obj, d)
			}
		} else {
			entry.Delay[0], entry.Delay[1] = Abc_ObjDelay0(manager, obj), Abc_ObjDelay1(manager, obj)
		}
	}
	degreeFinal := int32(1)
	if params.Cascade != 0 {
		degreeFinal = params.Degree
	}
	if params.VeryVerbose != 0 {
		for _, obj := range nodes {
			fmt.Printf("Node %3d : Lev =%3d   ", abc.Abc_ObjId(obj), abc.Abc_ObjLevel(obj))
			for d := int32(0); d <= degreeFinal; d++ {
				fmt.Printf("Del%d =%4.2f  ", d, Abc_IffDelay(manager, obj, d))
			}
			fmt.Println()
		}
	}
	delay := float32(0)
	for i := int32(0); i < abc.Abc_NtkCoNum(ntk); i++ {
		co := abc.Abc_NtkCo(ntk, i)
		delay = max(delay, Abc_IffDelay(manager, abc.Abc_ObjFanin0(co), degreeFinal))
	}
	delay10, count10 := float32(0.9*float64(delay)), int32(0)
	for i := int32(0); i < abc.Abc_NtkCoNum(ntk); i++ {
		co := abc.Abc_NtkCo(ntk, i)
		if Abc_IffDelay(manager, abc.Abc_ObjFanin0(co), degreeFinal) >= delay10 {
			count10++
		}
	}
	criticalPercent := 100.0 * float64(count10) / float64(abc.Abc_NtkCoNum(ntk))
	fmt.Printf("Critical delay %5.2f. Critical outputs %5.2f %%\n", delay, criticalPercent)
	Abc_NtkIfifStop(manager)
}
func ififType(cascade int32) string {
	if cascade != 0 {
		return "Cascade"
	}
	return "Cluster"
}
