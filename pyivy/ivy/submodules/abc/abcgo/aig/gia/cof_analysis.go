package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"time"
)

func Cof_ManInsertEntry_rec(nodes *Vec_Ptr_t, node *Cof_Obj_t, limit int32) {
	if limit <= 0 {
		return
	}
	all := append([]any(nil), vec.Vec_PtrArray(nodes)...)
	all = append(all, node)
	sort.SliceStable(all, func(i, j int) bool {
		return Cof_ObjFanoutNum(all[i].(*Cof_Obj_t)) > Cof_ObjFanoutNum(all[j].(*Cof_Obj_t))
	})
	if len(all) > int(limit) {
		all = all[:limit]
	}
	vec.Vec_PtrClear(nodes)
	for _, entry := range all {
		vec.Vec_PtrPush(nodes, entry)
	}
}
func Cof_ManCollectHighFanout(p *Cof_Man_t, limit int32) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(int(max(limit, 0)))
	for _, obj := range p.objects {
		if Cof_ObjIsCi(obj) != 0 || Cof_ObjIsNode(obj) != 0 {
			Cof_ManInsertEntry_rec(result, obj, limit)
		}
	}
	return result
}

// Cof_ManCountRemoved propagates substitutions that simplify to an existing
// AND node or literal. Object IDs are topological, so one forward pass suffices.
func Cof_ManCountRemoved(p *Cof_Man_t, root *Cof_Obj_t, one int32) int32 {
	if one != 0 && one != 1 {
		panic("gia: invalid cofactor constant")
	}
	Cof_ManIncrementTravId(p)
	Cof_ObjSetTravIdCurrent(p, root)
	root.iLit = one
	var count int32
	for _, node := range p.objects[root.Id+1:] {
		if Cof_ObjIsNode(node) == 0 || len(node.fanins) != 2 {
			continue
		}
		_, _, _, kind := p.pGia.Gate(node.Id)
		if kind != 2 {
			continue
		}
		changed := false
		lits := [2]int32{}
		for i, fanin := range node.fanins {
			lits[i] = fanin.Id << 1
			if Cof_ObjIsTravIdCurrent(p, fanin) != 0 {
				lits[i] = fanin.iLit
				changed = true
			}
			if node.complements[i] {
				lits[i] ^= 1
			}
		}
		if !changed {
			continue
		}
		lit := Gia_ManHashAndTry(p.pGia, lits[0], lits[1])
		if lit < 0 {
			continue
		}
		node.iLit = lit
		Cof_ObjSetTravIdCurrent(p, node)
		count++
	}
	return count
}
func Cof_ManPrintHighFanoutOne(p *Cof_Man_t, obj *Cof_Obj_t) {
	nodes := []*Cof_Obj_t{obj}
	fmt.Printf("%7d : i/o/c =%2d %5d %5d  l =%4d  s =%5d  TFI =%7d  TFO =%7d  C0 =%6d  C1 =%6d\n", obj.Id, Cof_ObjFaninNum(obj), Cof_ObjFanoutNum(obj), 2*obj.nFanoutsM, Cof_ObjLevel(p, obj), Cof_ManSuppSize(p, nodes, 1), Cof_ManTfiSize(p, nodes, 1), Cof_ManTfoSize(p, nodes, 1), Cof_ManCountRemoved(p, obj, 0), Cof_ManCountRemoved(p, obj, 1))
}
func Cof_ManPrintHighFanout(p *Cof_Man_t, n int32) {
	nodes := Cof_ManCollectHighFanout(p, n)
	for _, entry := range vec.Vec_PtrArray(nodes) {
		Cof_ManPrintHighFanoutOne(p, entry.(*Cof_Obj_t))
	}
}
func Cof_NodeDeref_rec(node *Cof_Obj_t) int32 {
	if len(node.fanins) == 0 {
		return 0
	}
	node.refs--
	if node.refs > 0 {
		return 0
	}
	count := int32(1)
	for _, fanin := range node.fanins {
		count += Cof_NodeDeref_rec(fanin)
	}
	return count
}
func Cof_NodeRef_rec(node *Cof_Obj_t) int32 {
	if len(node.fanins) == 0 {
		return 0
	}
	old := node.refs
	node.refs++
	if old > 0 {
		return 0
	}
	count := int32(1)
	for _, fanin := range node.fanins {
		count += Cof_NodeRef_rec(fanin)
	}
	return count
}
func Cof_ObjMffcSize(node *Cof_Obj_t) int32 {
	old := node.refs
	node.refs = 1
	count := Cof_NodeDeref_rec(node)
	restored := Cof_NodeRef_rec(node)
	node.refs = old
	if count != restored {
		panic("gia: inconsistent MFFC references")
	}
	return count
}
func Cof_ManPrintFanio(p *Cof_Man_t) {
	type row struct{ fanins, fanouts, mffc int32 }
	histogram := map[int32]*row{}
	var maxIn, maxOut, maxMffc, totalIn, totalOut, totalMffc, mffcNodes int32
	for _, node := range p.objects {
		if Cof_ObjIsNode(node) == 0 {
			continue
		}
		in, out := Cof_ObjFaninNum(node), Cof_ObjFanoutNum(node)
		var mffc int32
		if out > 1 {
			mffc = Cof_ObjMffcSize(node)
			mffcNodes++
		}
		totalIn += in
		totalOut += out
		totalMffc += mffc
		maxIn = max(maxIn, in)
		maxOut = max(maxOut, out)
		maxMffc = max(maxMffc, mffc)
		for i, value := range []int32{in, out, mffc} {
			if i == 2 && value == 0 {
				continue
			}
			r := histogram[value]
			if r == nil {
				r = &row{}
				histogram[value] = r
			}
			switch i {
			case 0:
				r.fanins++
			case 1:
				r.fanouts++
			case 2:
				r.mffc++
			}
		}
	}
	keys := make([]int, 0, len(histogram))
	for k := range histogram {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	fmt.Println("The distribution of fanins, fanouts, and MFFCs in the network:")
	fmt.Println("         Number    Nodes with fanin   Nodes with fanout   Nodes with MFFC")
	for _, key := range keys {
		r := histogram[int32(key)]
		fmt.Printf("%15d : %11d       %12d    %12d\n", key, r.fanins, r.fanouts, r.mffc)
	}
	average := func(a, b int32) float64 {
		if b == 0 {
			return 0
		}
		return float64(a) / float64(b)
	}
	fmt.Printf("Fanins: Max = %d. Ave = %.2f.  Fanouts: Max = %d. Ave = %.2f.  MFFCs: Max = %d. Ave = %.2f.\n", maxIn, average(totalIn, p.nNodes), maxOut, average(totalOut, p.nNodes), maxMffc, average(totalMffc, mffcNodes))
}
func Gia_ManPrintFanio(g *Gia_Man_t, n int32) {
	started := time.Now()
	p := Cof_ManCreateLogicSimple(g)
	defer Cof_ManStop(p)
	Gia_ManLevelNum(g)
	Cof_ManPrintFanio(p)
	if n > 0 {
		Cof_ManResetTravId(p)
		Gia_ManHashStart(g)
		Cof_ManPrintHighFanout(p, n)
		Gia_ManHashStop(g)
		fmt.Printf("Time: %s\n", time.Since(started))
	}
}
