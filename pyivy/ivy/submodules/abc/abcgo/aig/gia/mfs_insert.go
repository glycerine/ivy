package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func giaMfsTopo(d *giaMfsNetworkData, groups [][]int, groupMap []int, allBoxes bool) ([]int, []int) {
	seen := make([]bool, len(d.fanins))
	var nodes, boxes []int
	var visit func(int)
	visit = func(id int) {
		if id < 0 || id >= len(seen) {
			panic("gia: invalid MFS node")
		}
		if id < d.pis || seen[id] {
			return
		}
		group := -1
		if groupMap[id] >= 0 {
			group = groupMap[id] >> 1
		}
		if group >= 0 {
			members := groups[group]
			for _, v := range members {
				seen[v] = true
			}
			for _, v := range members {
				for _, f := range d.fanins[v] {
					visit(f)
				}
			}
			nodes = append(nodes, members...)
			boxes = append(boxes, group)
		} else {
			seen[id] = true
			for _, f := range d.fanins[id] {
				visit(f)
			}
			nodes = append(nodes, id)
		}
	}
	if allBoxes {
		for _, g := range groups {
			if len(g) > 0 {
				visit(g[0])
			}
		}
	}
	for id := len(d.fanins) - d.pos; id < len(d.fanins); id++ {
		if len(d.fanins[id]) != 1 {
			panic("gia: invalid MFS output")
		}
		visit(d.fanins[id][0])
	}
	return nodes, boxes
}
func giaMfsInsert(p *Gia_Man_t, d *giaMfsNetworkData, allBoxes bool) *Gia_Man_t {
	boxes := int(Gia_ManBoxNum(p))
	realPis, realPos := int(Gia_ManPiNum(p)), int(Gia_ManPoNum(p))
	var blackIns, blackOuts int
	var timing *tim.Tim_Man_t
	if p.PManTime != nil {
		timing = p.PManTime.(*tim.Tim_Man_t)
		tim.Tim_ManBlackBoxIoNum(timing, &blackIns, &blackOuts)
		if boxes != 0 {
			realPis, realPos = tim.Tim_ManPiNum(timing), tim.Tim_ManPoNum(timing)
		}
	}
	copies := vec.Vec_IntArray(&p.vCopies)
	groups, groupMap := giaMfsBoxGroups(p, copies, len(d.fanins))
	topo, remaining := giaMfsTopo(d, groups, groupMap, allBoxes)
	old := make([]int32, len(d.fanins))
	vals := make([]int32, len(d.fanins))
	for i := range old {
		old[i], vals[i] = -1, -1
	}
	for _, obj := range p.pObjs {
		node := copies[obj.id]
		if node > 0 {
			old[node] = obj.id
		}
	}
	dst := Gia_ManStart(p.nObjs)
	dst.CopyIdentityFrom(p)
	vals[copies[0]] = 0
	for i, id := range vec.Vec_IntArray(p.vCis) {
		if i < realPis {
			vals[copies[id]] = Gia_ManAppendCi(dst)
		}
	}
	// vMapping is indexed by new object; vMapping2 holds the LUT records.
	mapping := vec.Vec_IntStart(int(p.nObjs))
	records := vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(mapping, 0, int32(vec.Vec_IntSize(records)))
	vec.Vec_IntPush(records, 0)
	vec.Vec_IntPush(records, 0)
	leaves := vec.Vec_IntAlloc(6)
	cover := vec.Vec_IntAlloc(1 << 16)
	for _, node := range topo {
		fanins := d.fanins[node]
		group := groupMap[node]
		var lit int32
		if len(fanins) == 1 && fanins[0] < blackOuts {
			if group < 0 || group&1 == 0 {
				panic("gia: invalid black box output")
			}
			vals[node] = Gia_ManAppendCi(dst)
			continue
		}
		vec.Vec_IntClear(leaves)
		for _, f := range fanins {
			if vals[f] < 0 {
				panic("gia: MFS fanin not available")
			}
			vec.Vec_IntPush(leaves, vals[f])
		}
		if group < 0 {
			truth := d.truths[node]
			if old[node] >= 0 && Gia_ObjLutIsMux(p, old[node]) != 0 {
				size := vec.Vec_IntSize(records)
				var nVarsNew int32
				util.Abc_TtSimplify(truth, vec.Vec_IntArray(leaves), int32(vec.Vec_IntSize(leaves)), &nVarsNew)
				vec.Vec_IntShrink(leaves, int(nVarsNew))
				lit = Gia_ManFromIfLogicCreateLut(dst, truth, leaves, cover, mapping, records)
				if size < vec.Vec_IntSize(records) {
					vec.Vec_IntWriteEntry(records, vec.Vec_IntSize(records)-1, -(lit >> 1))
				}
			} else {
				lit = Gia_ManFromIfLogicCreateLut(dst, truth, leaves, cover, mapping, records)
			}
		} else if group&1 != 0 {
			lit = Gia_ManAppendCi(dst)
		} else {
			truth := d.truths[node][0]
			if truth != 0xaaaaaaaaaaaaaaaa && truth != 0x5555555555555555 {
				panic("gia: invalid box buffer")
			}
			lit = Gia_ManAppendCo(dst, vec.Vec_IntEntry(leaves, 0)^b2i32(truth == 0x5555555555555555))
		}
		vals[node] = lit
	}
	for i, id := range vec.Vec_IntArray(p.vCos) {
		if i < int(Gia_ManCoNum(p))-realPos {
			continue
		}
		obj := Gia_ManObj(p, id)
		lit := vals[copies[Gia_ObjFanin0(obj).id]]
		if lit < 0 {
			panic("gia: missing MFS output")
		}
		Gia_ManAppendCo(dst, lit^Gia_ObjFaninC0(obj))
	}
	if vec.Vec_IntSize(mapping) > int(dst.nObjs) {
		vec.Vec_IntShrink(mapping, int(dst.nObjs))
	} else {
		vec.Vec_IntFillExtra(mapping, int(dst.nObjs), 0)
	}
	for i, offset := range vec.Vec_IntArray(mapping) {
		if offset > 0 {
			vec.Vec_IntWriteEntry(mapping, i, offset+dst.nObjs)
		}
	}
	vec.Vec_IntAppend(mapping, records)
	dst.VMapping = mapping
	selected := vec.Vec_IntAlloc(len(remaining))
	for _, b := range remaining {
		vec.Vec_IntPush(selected, int32(b))
	}
	if timing != nil {
		dst.PManTime = Gia_ManUpdateTimMan2(p, selected, 0)
	}
	if p.PAigExtra != nil {
		dst.PAigExtra = Gia_ManUpdateExtraAig2(p.PManTime, p.PAigExtra, selected)
	}
	if p.vRegClasses != nil {
		dst.vRegClasses = vec.Vec_IntDup(p.vRegClasses)
	}
	if p.vRegInits != nil {
		dst.vRegInits = vec.Vec_IntDup(p.vRegInits)
	}
	dst.nAnd2Delay = p.nAnd2Delay
	return dst
}

func b2i32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// Gia_ManFromIfLogicCreateLut (giaIf.c) builds a truth function over leaves
// and records it as a LUT unless the result is a constant or one of the
// leaves. A node reused through structural hashing gets its latest record.
func Gia_ManFromIfLogicCreateLut(dst *Gia_Man_t, truth []word, leaves, cover, mapping, records *vec.Vec_Int_t) int32 {
	n := vec.Vec_IntSize(leaves)
	words := make([]uint32, len(truth)*2)
	for i, value := range truth {
		words[2*i], words[2*i+1] = uint32(value), uint32(value>>32)
	}
	var scratch []uint32
	lit := kit.Kit_TruthToGia(dst, words, int32(n), &scratch, vec.Vec_IntArray(leaves), 0)
	id := lit >> 1
	if id == 0 {
		return lit
	}
	for _, leaf := range vec.Vec_IntArray(leaves) {
		if leaf>>1 == id {
			return lit
		}
	}
	vec.Vec_IntSetEntry(mapping, int(id), int32(vec.Vec_IntSize(records)))
	vec.Vec_IntPush(records, int32(n))
	for _, leaf := range vec.Vec_IntArray(leaves) {
		vec.Vec_IntPush(records, leaf>>1)
	}
	vec.Vec_IntPush(records, id)
	return lit
}
