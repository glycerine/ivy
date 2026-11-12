package sbd

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
)

type Sbd_Srv_t_ struct {
	nLutSize, nCutSize, nCutNum, fVerbose int32
	pGia                                  *Network
	vMirrors, vLutLevs, vLevs, vRefs      *vec.Vec_Int_t
	vCut0, vCut, vCutTop, vCutBot         *vec.Vec_Int_t
}
type Sbd_Srv_t = Sbd_Srv_t_

func Sbd_ManCutServerStart(network *Network, mirrors, lutLevels, levels, refs *vec.Vec_Int_t, lutSize, cutSize, cutNum, verbose int32) *Sbd_Srv_t {
	if lutSize < 1 || lutSize > cutSize || cutSize < 2 || cutSize > 10 || cutNum < 2 || cutNum >= 501 {
		panic("invalid cut-server dimensions")
	}
	return &Sbd_Srv_t{nLutSize: lutSize, nCutSize: cutSize, nCutNum: cutNum, fVerbose: verbose, pGia: network, vMirrors: mirrors, vLutLevs: lutLevels, vLevs: levels, vRefs: refs, vCut0: vec.Vec_IntAlloc(100), vCut: vec.Vec_IntAlloc(100), vCutTop: vec.Vec_IntAlloc(100), vCutBot: vec.Vec_IntAlloc(100)}
}
func Sbd_ManCutServerStop(p *Sbd_Srv_t) {
	if p != nil {
		*p = Sbd_Srv_t{}
	}
}
func sbdMirrorObject(mirrors *vec.Vec_Int_t, id int32) int32 {
	if mirrors != nil {
		if lit := vec.Vec_IntEntry(mirrors, int(id)); lit >= 0 {
			return lit >> 1
		}
	}
	return id
}
func sbdCutTopo(network *Network, mirrors *vec.Vec_Int_t, id int32, seen map[int32]bool) bool {
	id = sbdMirrorObject(mirrors, id)
	if id == 0 || seen[id] {
		return true
	}
	seen[id] = true
	node := network.object(id)
	if node.Kind == aig.Input {
		return false
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("topology traversal requires gate")
	}
	a := sbdCutTopo(network, mirrors, int32(node.Fanins[0].NodeID()), seen)
	b := sbdCutTopo(network, mirrors, int32(node.Fanins[1].NodeID()), seen)
	return a && b
}
func Sbd_ManCutIsTopo_rec(network *Network, mirrors *vec.Vec_Int_t, id int32) int32 {
	if sbdCutTopo(network, mirrors, id, map[int32]bool{}) {
		return 1
	}
	return 0
}
func Sbd_ManCutIsTopo(network *Network, mirrors, cut *vec.Vec_Int_t, id int32) int32 {
	seen := map[int32]bool{}
	for _, leaf := range vec.Vec_IntArray(cut) {
		seen[leaf] = true
	}
	if !sbdCutTopo(network, mirrors, id, seen) {
		panic("cut is not topological")
	}
	return 1
}
func sbdCutFanins(network *Network, mirrors *vec.Vec_Int_t, id int32) (int32, int32) {
	node := network.object(id)
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("cut expansion requires binary gate")
	}
	return sbdMirrorObject(mirrors, int32(node.Fanins[0].NodeID())), sbdMirrorObject(mirrors, int32(node.Fanins[1].NodeID()))
}
func sbdPushOrdered(cut *vec.Vec_Int_t, id int32) {
	if id != 0 {
		vec.Vec_IntPushUniqueOrder(cut, id)
	}
}
func Sbd_ManCutExpandOne(network *Network, mirrors, levels, cut *vec.Vec_Int_t, place, id int32) int32 {
	if network.object(id).Kind == aig.Input {
		return 0
	}
	a, b := sbdCutFanins(network, mirrors, id)
	level := vec.Vec_IntEntry(levels, int(id))
	hasA, hasB := slices.Contains(vec.Vec_IntArray(cut), a), slices.Contains(vec.Vec_IntArray(cut), b)
	if !hasA && !hasB || vec.Vec_IntEntry(levels, int(a)) > level || vec.Vec_IntEntry(levels, int(b)) > level {
		return 0
	}
	vec.Vec_IntDrop(cut, int(place))
	if !hasA {
		sbdPushOrdered(cut, a)
	}
	if !hasB {
		sbdPushOrdered(cut, b)
	}
	return 1
}
func Vec_IntIsOrdered(cut *vec.Vec_Int_t) {
	data := vec.Vec_IntArray(cut)
	for i := 1; i < len(data); i++ {
		if data[i-1] >= data[i] {
			panic("unordered cut")
		}
	}
}
func Sbd_ManCutExpand(network *Network, mirrors, levels, cut *vec.Vec_Int_t) {
	for {
		changed := false
		for i, id := range vec.Vec_IntArray(cut) {
			if Sbd_ManCutExpandOne(network, mirrors, levels, cut, int32(i), id) != 0 {
				changed = true
				break
			}
		}
		if !changed {
			return
		}
	}
}
func Sbd_ManCutReload(mirrors, levels *vec.Vec_Int_t, stop int32, cut, top, bottom *vec.Vec_Int_t) {
	vec.Vec_IntClear(top)
	vec.Vec_IntClear(bottom)
	for _, id := range vec.Vec_IntArray(cut) {
		if id == 0 || sbdMirrorObject(mirrors, id) != id || vec.Vec_IntEntry(levels, int(id)) > stop {
			panic("invalid cut reload")
		}
		if vec.Vec_IntEntry(levels, int(id)) == stop {
			vec.Vec_IntPush(top, id)
		} else {
			vec.Vec_IntPush(bottom, id)
		}
	}
	Vec_IntIsOrdered(cut)
}
func sbdCollectCut(network *Network, mirrors *vec.Vec_Int_t, id, stop int32, levels, cut *vec.Vec_Int_t, seen map[int32]bool) bool {
	id = sbdMirrorObject(mirrors, id)
	if id == 0 || seen[id] {
		return true
	}
	seen[id] = true
	node := network.object(id)
	level := vec.Vec_IntEntry(levels, int(id))
	if node.Kind == aig.Input || level <= stop {
		vec.Vec_IntPush(cut, id)
		return level <= stop
	}
	a, b := sbdCutFanins(network, mirrors, id)
	ra := sbdCollectCut(network, mirrors, a, stop, levels, cut, seen)
	rb := sbdCollectCut(network, mirrors, b, stop, levels, cut, seen)
	return ra && rb
}
func Sbd_ManCutCollect_rec(network *Network, mirrors *vec.Vec_Int_t, id, stop int32, levels, cut *vec.Vec_Int_t) int32 {
	if sbdCollectCut(network, mirrors, id, stop, levels, cut, map[int32]bool{}) {
		return 1
	}
	return 0
}
func Sbd_ManCutReduceTop(network *Network, mirrors *vec.Vec_Int_t, id int32, levels, cut, top *vec.Vec_Int_t, _ int32) int32 {
	stop := vec.Vec_IntEntry(levels, int(id)) - 2
	Vec_IntIsOrdered(cut)
	data := vec.Vec_IntArray(top)
	for i := len(data) - 1; i >= 0; i-- {
		leaf := data[i]
		if network.object(leaf).Kind == aig.Input {
			continue
		}
		if vec.Vec_IntEntry(levels, int(leaf)) != stop {
			panic("cut top level mismatch")
		}
		a, b := sbdCutFanins(network, mirrors, leaf)
		la, lb := vec.Vec_IntEntry(levels, int(a)), vec.Vec_IntEntry(levels, int(b))
		if la > stop || lb > stop || la == stop && lb == stop {
			continue
		}
		vec.Vec_IntRemove(cut, leaf)
		sbdPushOrdered(cut, a)
		sbdPushOrdered(cut, b)
		return 1
	}
	return 0
}
func Sbd_ManCutServerFirst(p *Sbd_Srv_t, id int32, out []int32) int32 {
	stop := vec.Vec_IntEntry(p.vLutLevs, int(id)) - 2
	collect := func(bound int32) bool {
		vec.Vec_IntClear(p.vCut)
		if Sbd_ManCutCollect_rec(p.pGia, p.vMirrors, id, bound, p.vLutLevs, p.vCut) == 0 {
			return false
		}
		slices.Sort(vec.Vec_IntArray(p.vCut))
		return true
	}
	expand := func() {
		Sbd_ManCutExpand(p.pGia, p.vMirrors, p.vLutLevs, p.vCut)
		Sbd_ManCutReload(p.vMirrors, p.vLutLevs, stop, p.vCut, p.vCutTop, p.vCutBot)
	}
	fits := func() bool {
		return vec.Vec_IntSize(p.vCut) <= int(p.nCutSize) && vec.Vec_IntSize(p.vCutTop) <= int(p.nLutSize-1)
	}
	save := func() int32 { copy(out, vec.Vec_IntArray(p.vCut)); return int32(vec.Vec_IntSize(p.vCut)) }
	if !collect(stop) {
		return -1
	}
	expand()
	if fits() {
		return save()
	}
	vec.Vec_IntClear(p.vCut0)
	for _, v := range vec.Vec_IntArray(p.vCut) {
		vec.Vec_IntPush(p.vCut0, v)
	}
	// The source unrolls exactly four successful top-reduction attempts.
	for attempts := 0; attempts < 4 && vec.Vec_IntSize(p.vCut) < int(p.nCutSize); attempts++ {
		if Sbd_ManCutReduceTop(p.pGia, p.vMirrors, id, p.vLutLevs, p.vCut, p.vCutTop, p.nCutSize) == 0 {
			break
		}
		expand()
		if vec.Vec_IntSize(p.vCut) > int(p.nCutSize) {
			panic("expanded cut exceeds bound")
		}
		if fits() {
			return save()
		}
	}
	if !collect(stop - 1) {
		return -1
	}
	expand()
	if fits() {
		return save()
	}
	return -1
}
