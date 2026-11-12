package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"time"
)

type Opa_Man_t struct {
	pGia     *Gia_Man_t
	vFront   *Vec_Int_t
	pvParts  []*Vec_Int_t
	pId2Part []int32
	nParts   int32
}

func Opa_ManStart(g *Gia_Man_t) *Opa_Man_t {
	p := &Opa_Man_t{pGia: g, vFront: vec.Vec_IntAlloc(100), pvParts: make([]*Vec_Int_t, Gia_ManPoNum(g)), pId2Part: make([]int32, g.nObjs), nParts: Gia_ManPoNum(g)}
	for i := range p.pId2Part {
		p.pId2Part[i] = -1
	}
	for i := int32(0); i < Gia_ManPoNum(g); i++ {
		id := Gia_ObjId(g, Gia_ManPo(g, i))
		p.pvParts[i] = vec.Vec_IntAlloc(1)
		vec.Vec_IntPush(p.pvParts[i], id)
		p.pId2Part[id] = i
		vec.Vec_IntPush(p.vFront, id)
	}
	return p
}
func Opa_ManStop(p *Opa_Man_t) {
	if p != nil {
		*p = Opa_Man_t{}
	}
}
func Opa_ManPrint(p *Opa_Man_t) {
	fmt.Println("Groups:")
	for i, part := range p.pvParts {
		if vec.Vec_IntSize(part) == 0 {
			continue
		}
		fmt.Printf("%3d : ", i)
		for _, id := range vec.Vec_IntArray(part) {
			fmt.Printf("%d ", id)
		}
		fmt.Println()
	}
}
func Opa_ManPrint2(p *Opa_Man_t) {
	fmt.Printf("Groups %d: ", p.nParts)
	for _, part := range p.pvParts {
		if vec.Vec_IntSize(part) == 0 {
			continue
		}
		count := int32(0)
		for _, id := range vec.Vec_IntArray(part) {
			count += Gia_ObjIsPo(p.pGia, Gia_ManObj(p.pGia, id))
		}
		fmt.Printf("%d ", count)
	}
	fmt.Println()
}
func Opa_ManMoveOne(p *Opa_Man_t, obj, fanin *Gia_Obj_t) {
	id, fid := Gia_ObjId(p.pGia, obj), Gia_ObjId(p.pGia, fanin)
	if fid == 0 {
		return
	}
	group := p.pId2Part[id]
	if group < 0 {
		panic("gia: unassigned partition object")
	}
	if p.pId2Part[fid] < 0 {
		p.pId2Part[fid] = group
		vec.Vec_IntPush(p.pvParts[group], fid)
		if Gia_ObjIsAnd(fanin) != 0 {
			vec.Vec_IntPush(p.vFront, fid)
		} else if Gia_ObjIsRo(p.pGia, fanin) != 0 {
			ri := Gia_ObjId(p.pGia, Gia_ObjRoToRi(p.pGia, fanin))
			if p.pId2Part[ri] >= 0 {
				panic("gia: duplicate register input in partition")
			}
			p.pId2Part[ri] = group
			vec.Vec_IntPush(p.pvParts[group], ri)
			vec.Vec_IntPush(p.vFront, ri)
		}
	} else if target := p.pId2Part[fid]; target != group {
		for _, member := range vec.Vec_IntArray(p.pvParts[group]) {
			vec.Vec_IntPush(p.pvParts[target], member)
			p.pId2Part[member] = target
		}
		vec.Vec_IntClear(p.pvParts[group])
		p.nParts--
	}
}
func Opa_ManPerform(g *Gia_Man_t) {
	p := Opa_ManStart(g)
	defer Opa_ManStop(p)
	limit := vec.Vec_IntSize(p.vFront)
	count := 0
	for i := 0; i < vec.Vec_IntSize(p.vFront); i++ {
		if i == limit {
			count++
			fmt.Printf("%6d : %6d -> %6d\n", count, i, p.nParts)
			limit = vec.Vec_IntSize(p.vFront)
			if count > 1 {
				Opa_ManPrint2(p)
			}
		}
		obj := Gia_ManObj(g, vec.Vec_IntEntry(p.vFront, i))
		for _, id := range giaObjectFanins(g, obj) {
			Opa_ManMoveOne(p, obj, Gia_ManObj(g, id))
		}
		if p.nParts == 1 || count == 5 {
			break
		}
	}
	fmt.Println()
}
func Gia_ManConeMark_rec(p *Gia_Man_t, obj *Gia_Obj_t, roots *Vec_Int_t, limit int32) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsRo(p, obj) != 0 {
		vec.Vec_IntPush(roots, Gia_ObjId(p, Gia_ObjRoToRi(p, obj)))
	} else {
		for _, id := range giaObjectFanins(p, obj) {
			if Gia_ManConeMark_rec(p, Gia_ManObj(p, id), roots, limit) != 0 {
				return 1
			}
		}
	}
	return giaBool(int32(vec.Vec_IntSize(roots)) > limit)
}
func giaConeCount(p *Gia_Man_t, outs []int32, limit int32) int32 {
	roots := vec.Vec_IntAlloc(len(outs))
	for _, out := range outs {
		vec.Vec_IntPush(roots, Gia_ObjId(p, Gia_ManPo(p, out)))
	}
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		if Gia_ManConeMark_rec(p, Gia_ManObj(p, vec.Vec_IntEntry(roots, i)), roots, limit) != 0 {
			break
		}
	}
	return int32(vec.Vec_IntSize(roots) - len(outs))
}
func Gia_ManConeMark(p *Gia_Man_t, out, limit int32) int32 {
	return giaConeCount(p, []int32{out}, limit)
}
func Gia_ManCountFlops(p *Gia_Man_t, outs *Vec_Int_t) int32 {
	return giaConeCount(p, vec.Vec_IntArray(outs), int32(^uint32(0)>>1))
}

// The two exploratory entry points are empty in ABC's source.
func Gia_ManFindPoPartition3(p *Gia_Man_t, out, delta, minOut, maxOut, verbose int32, parts **Vec_Ptr_t) *Gia_Man_t {
	return nil
}
func Gia_ManFindPoPartition2(p *Gia_Man_t, start, delta, minOut, maxOut, largest, verbose int32, parts **Vec_Ptr_t) *Gia_Man_t {
	return nil
}

func Gia_ManFindPivots(p *Gia_Man_t, shift, onlyCis, verbose int32) *Vec_Int_t {
	counts := make([]int32, p.nObjs)
	for _, obj := range p.pObjs {
		for _, id := range giaObjectFanins(p, obj) {
			counts[id]++
		}
	}
	type pivot struct{ id, weight int32 }
	pivots := make([]pivot, 0)
	for id, count := range counts {
		if count >= 2 && (onlyCis == 0 || Gia_ObjIsCi(p.pObjs[id]) != 0) {
			pivots = append(pivots, pivot{int32(id), count})
		}
	}
	if verbose != 0 {
		fmt.Printf("Selected %d pivots with more than one fanout (out of %d CIs and ANDs).\n", len(pivots), Gia_ManCiNum(p)+Gia_ManAndNum(p))
	}
	Gia_ManRandom(1)
	Gia_ManRandom(0)
	for i := range pivots {
		j := int(Gia_ManRandom(0)>>1) % len(pivots)
		pivots[i], pivots[j] = pivots[j], pivots[i]
	}
	if shift == 0 {
		sort.SliceStable(pivots, func(i, j int) bool { return pivots[i].weight > pivots[j].weight })
	}
	result := vec.Vec_IntAlloc(min(64, len(pivots)))
	for i := 0; i < min(64, len(pivots)); i++ {
		j := (int64(i) + int64(shift)) % int64(len(pivots))
		if j < 0 {
			j += int64(len(pivots))
		}
		v := pivots[j]
		if verbose != 0 {
			fmt.Printf("%2d : Pivot =%7d  Fanout =%7d\n", j, v.id, v.weight)
		}
		vec.Vec_IntPush(result, v.id)
	}
	return result
}
func Gia_ManDeriveSigns(p *Gia_Man_t, pivots *Vec_Int_t, verbose int32) *Vec_Wrd_t {
	if vec.Vec_IntSize(pivots) > 64 {
		panic("gia: too many signature pivots")
	}
	signs := vec.Vec_WrdStart(int(p.nObjs))
	values := vec.Vec_WrdArray(signs)
	Gia_ManFillValue(p)
	for i, id := range vec.Vec_IntArray(pivots) {
		if id < 0 || id >= p.nObjs {
			panic("gia: invalid pivot")
		}
		p.pObjs[id].Value = uint32(i)
		values[id] |= uint64(1) << uint(i)
	}
	if verbose != 0 {
		fmt.Print("Signature propagation: ")
	}
	for {
		for _, obj := range p.pObjs {
			for _, id := range giaObjectFanins(p, obj) {
				values[obj.id] |= values[id]
			}
		}
		changed := 0
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			ro, ri := Gia_ManCi(p, Gia_ManPiNum(p)+i), Gia_ManCo(p, Gia_ManPoNum(p)+i)
			old := values[ro.id]
			values[ro.id] |= values[ri.id]
			if old != values[ro.id] {
				changed++
			}
		}
		if verbose != 0 {
			fmt.Printf("%d ", changed)
		}
		if changed == 0 {
			break
		}
	}
	if verbose != 0 {
		fmt.Println()
	}
	return signs
}
func Gia_ManHashOutputs(p *Gia_Man_t, signs *Vec_Wrd_t, verbose int32) *Vec_Ptr_t {
	bins := vec.Vec_PtrAlloc(0)
	index := make(map[uint64]*Vec_Int_t)
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		signature := vec.Vec_WrdEntry(signs, int(Gia_ObjId(p, Gia_ManPo(p, i))))
		bin := index[signature]
		if bin == nil {
			bin = vec.Vec_IntAlloc(1)
			index[signature] = bin
			vec.Vec_PtrPush(bins, bin)
		}
		vec.Vec_IntPush(bin, i)
	}
	entries := vec.Vec_PtrArray(bins)
	sort.SliceStable(entries, func(i, j int) bool {
		return vec.Vec_IntSize(entries[i].(*Vec_Int_t)) > vec.Vec_IntSize(entries[j].(*Vec_Int_t))
	})
	if verbose != 0 {
		fmt.Printf("Computed %d partitions:\n", len(entries))
	} else {
		fmt.Println("Listing partitions with more than 100 outputs:")
	}
	for i, entry := range entries {
		bin := entry.(*Vec_Int_t)
		if verbose == 0 && vec.Vec_IntSize(bin) <= 100 {
			continue
		}
		first := vec.Vec_IntEntry(bin, 0)
		signature := vec.Vec_WrdEntry(signs, int(Gia_ObjId(p, Gia_ManPo(p, first))))
		fmt.Printf("%3d %064b  PO =%7d  FF =%7d\n", i, signature, vec.Vec_IntSize(bin), Gia_ManCountFlops(p, bin))
	}
	return bins
}
func Gia_ManFindPoPartition(p *Gia_Man_t, shift, onlyCis, largest, verbose int32, parts **Vec_Ptr_t) *Gia_Man_t {
	start := time.Now()
	pivots := Gia_ManFindPivots(p, shift, onlyCis, verbose)
	signs := Gia_ManDeriveSigns(p, pivots, verbose)
	bins := Gia_ManHashOutputs(p, signs, verbose)
	var result *Gia_Man_t
	if largest != 0 && vec.Vec_PtrSize(bins) > 0 {
		result = giaDupPartitionCones(p, vec.Vec_IntArray(vec.Vec_PtrEntry(bins, 0).(*Vec_Int_t)))
	}
	if parts != nil {
		*parts = bins
		fmt.Printf("The algorithm divided %d POs into %d partitions. Time: %s\n", Gia_ManPoNum(p), vec.Vec_PtrSize(bins), time.Since(start))
	}
	return result
}

// giaDupPartitionCones retains selected POs and the transitive register inputs
// needed by their sequential cones. Primary inputs are trimmed to their support.
func giaDupPartitionCones(p *Gia_Man_t, outputs []int32) *Gia_Man_t {
	return giaDupPartitionConesMode(p, outputs, true)
}

func giaDupPartitionConesMode(p *Gia_Man_t, outputs []int32, trim bool) *Gia_Man_t {
	roots := make([]int32, 0, len(outputs))
	for _, i := range outputs {
		roots = append(roots, Gia_ObjId(p, Gia_ManPo(p, i)))
	}
	seen := make([]bool, p.nObjs)
	seen[0] = true
	nodes := make([]int32, 0)
	inputs := make([]int32, 0)
	var visit func(int32)
	visit = func(id int32) {
		if seen[id] {
			return
		}
		seen[id] = true
		obj := p.pObjs[id]
		if Gia_ObjIsPi(p, obj) != 0 {
			inputs = append(inputs, id)
			return
		}
		if Gia_ObjIsRo(p, obj) != 0 {
			roots = append(roots, Gia_ObjId(p, Gia_ObjRoToRi(p, obj)))
			return
		}
		for _, fanin := range giaObjectFanins(p, obj) {
			visit(fanin)
		}
		if Gia_ObjIsAnd(obj) != 0 {
			nodes = append(nodes, id)
		}
	}
	for i := 0; i < len(roots); i++ {
		visit(roots[i])
	}
	sort.Slice(inputs, func(i, j int) bool { return Gia_ObjCioId(p.pObjs[inputs[i]]) < Gia_ObjCioId(p.pObjs[inputs[j]]) })
	if !trim {
		inputs = inputs[:0]
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			inputs = append(inputs, Gia_ManPi(p, i).id)
		}
	}
	out := Gia_ManStart(int32(1 + len(inputs) + len(nodes) + len(roots)*2))
	out.CopyIdentityFrom(p)
	mapping := make([]int32, p.nObjs)
	for _, id := range inputs {
		mapping[id] = Gia_ManAppendCi(out)
	}
	for _, id := range roots[len(outputs):] {
		ri := p.pObjs[id]
		ro := Gia_ObjRiToRo(p, ri)
		mapping[ro.id] = Gia_ManAppendCi(out)
	}
	read := func(lit int32) int32 { return mapping[lit>>1] ^ (lit & 1) }
	for _, id := range nodes {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			mapping[id] = Gia_ManAppendAnd(out, read(a), read(b))
		case 3:
			mapping[id] = Gia_ManAppendXor(out, read(a), read(b))
		case 4:
			mapping[id] = Gia_ManAppendMux(out, read(a), read(b), read(c))
		case 5:
			mapping[id] = Gia_ManAppendBuf(out, read(a))
		default:
			panic("gia: invalid partition cone gate")
		}
	}
	for _, id := range roots {
		a, _, _, _ := p.Gate(id)
		Gia_ManAppendCo(out, read(a))
	}
	Gia_ManSetRegNum(out, int32(len(roots)-len(outputs)))
	return out
}
