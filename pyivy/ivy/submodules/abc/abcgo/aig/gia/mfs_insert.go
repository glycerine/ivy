package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
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
	records := []int32{0, 0, 0}
	offsets := map[int32]int32{0: 1}
	var scratch []uint32
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
		leaves := make([]int32, len(fanins))
		for i, f := range fanins {
			leaves[i] = vals[f]
			if leaves[i] < 0 {
				panic("gia: MFS fanin not available")
			}
		}
		if group < 0 {
			truth := d.truths[node]
			t32 := make([]uint32, 2*len(truth))
			for i, w := range truth {
				t32[2*i], t32[2*i+1] = uint32(w), uint32(w>>32)
			}
			lit = kit.Kit_TruthToGia(dst, t32, int32(len(leaves)), &scratch, leaves, 0)
			if lit>>1 != 0 && Gia_ObjIsAnd(Gia_ManObj(dst, lit>>1)) != 0 {
				offsets[lit>>1] = int32(len(records))
				records = append(records, int32(len(leaves)))
				for _, leaf := range leaves {
					records = append(records, leaf>>1)
				}
				root := lit >> 1
				if old[node] >= 0 && Gia_ObjLutIsMux(p, old[node]) != 0 {
					root = -root
				}
				records = append(records, root)
			}
		} else if group&1 != 0 {
			lit = Gia_ManAppendCi(dst)
		} else {
			if len(leaves) != 1 {
				panic("gia: invalid box input")
			}
			truth := d.truths[node][0]
			if truth == 0x5555555555555555 {
				leaves[0] ^= 1
			} else if truth != 0xaaaaaaaaaaaaaaaa {
				panic("gia: invalid box buffer")
			}
			lit = Gia_ManAppendCo(dst, leaves[0])
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
	mapping := vec.Vec_IntStart(int(dst.nObjs))
	for id, offset := range offsets {
		vec.Vec_IntWriteEntry(mapping, int(id), dst.nObjs+offset)
	}
	for _, v := range records {
		vec.Vec_IntPush(mapping, v)
	}
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
