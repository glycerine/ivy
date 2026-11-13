package acec

// Adder pool box collection from src/proof/acec/acecPool.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"time"
)

func Acec_ManCompareTwo(a, b []int32) int32 {
	if a[1] < b[1] {
		return -1
	}
	if a[1] > b[1] {
		return 1
	}
	return 0
}
func acecSortRankPairs(v *Vec_Int_t) {
	data := vec.Vec_IntArray(v)
	pairs := make([][2]int32, len(data)/2)
	for i := range pairs {
		pairs[i] = [2]int32{data[2*i], data[2*i+1]}
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i][1] < pairs[j][1] })
	for i, p := range pairs {
		data[2*i], data[2*i+1] = p[0], p[1]
	}
}
func Acec_ManCollectInsOuts(p *Gia_Man_t, adds, boxes, ranks *Vec_Int_t, ins, outs *Vec_Bit_t, resultIns, resultOuts *Vec_Int_t) {
	for _, box := range vec.Vec_IntArray(boxes) {
		off := int(6 * box)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(ins, int(vec.Vec_IntEntry(adds, off+k)), 1)
		}
		for k := 3; k < 5; k++ {
			vec.Vec_BitWriteEntry(outs, int(vec.Vec_IntEntry(adds, off+k)), 1)
		}
	}
	for _, box := range vec.Vec_IntArray(boxes) {
		off := int(6 * box)
		for k := 0; k < 3; k++ {
			id := vec.Vec_IntEntry(adds, off+k)
			if vec.Vec_BitEntry(outs, int(id)) == 0 {
				vec.Vec_IntPushTwo(resultIns, id, vec.Vec_IntEntry(ranks, int(box)))
			}
		}
		for k := 3; k < 5; k++ {
			id := vec.Vec_IntEntry(adds, off+k)
			if id != 0 && vec.Vec_BitEntry(ins, int(id)) == 0 {
				rank := vec.Vec_IntEntry(ranks, int(box))
				if k == 4 {
					rank--
				}
				vec.Vec_IntPushTwo(resultOuts, id, rank)
			}
		}
	}
	for _, box := range vec.Vec_IntArray(boxes) {
		off := int(6 * box)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(ins, int(vec.Vec_IntEntry(adds, off+k)), 0)
		}
		for k := 3; k < 5; k++ {
			vec.Vec_BitWriteEntry(outs, int(vec.Vec_IntEntry(adds, off+k)), 0)
		}
	}
	var max int32
	for _, v := range []*Vec_Int_t{resultIns, resultOuts} {
		for i := 1; i < vec.Vec_IntSize(v); i += 2 {
			if rank := vec.Vec_IntEntry(v, i); rank > max {
				max = rank
			}
		}
	}
	for _, v := range []*Vec_Int_t{resultIns, resultOuts} {
		for i := 1; i < vec.Vec_IntSize(v); i += 2 {
			vec.Vec_IntWriteEntry(v, i, 1+max-vec.Vec_IntEntry(v, i))
		}
		acecSortRankPairs(v)
	}
}
func Acec_ManCollectBoxSets_rec(p *Gia_Man_t, carry, rank int32, adds, m, boxes, ranks *Vec_Int_t) {
	box := vec.Vec_IntEntry(m, int(carry))
	if box == -1 {
		return
	}
	for k := 0; k < 3; k++ {
		id := vec.Vec_IntEntry(adds, int(box*6)+k)
		if k < 2 || id != 0 {
			Acec_ManCollectBoxSets_rec(p, id, rank+1, adds, m, boxes, ranks)
		}
	}
	vec.Vec_IntPush(boxes, box)
	vec.Vec_IntWriteEntry(ranks, int(box), rank)
}
func Acec_ManCollectBoxSets(p *Gia_Man_t, adds, xors *Vec_Int_t) *Vec_Wec_t {
	m := Acec_ManCreateCarryMap(p, adds)
	sets := Acec_ManCollectCarryRootSets(p, adds, m, xors, Acec_ManCollectXorRoots(p, xors), Acec_ManCollectCarryRoots(p, adds))
	ranks := vec.Vec_IntStart(vec.Vec_IntSize(adds) / 6)
	ins, outs := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; 5*i < vec.Vec_WecSize(sets); i++ {
		roots, boxes := vec.Vec_WecEntry(sets, 5*i+1), vec.Vec_WecEntry(sets, 5*i+2)
		for _, root := range vec.Vec_IntArray(roots) {
			Acec_ManCollectBoxSets_rec(p, root, 1, adds, m, boxes, ranks)
		}
		Acec_ManCollectInsOuts(p, adds, boxes, ranks, ins, outs, vec.Vec_WecEntry(sets, 5*i+3), vec.Vec_WecEntry(sets, 5*i+4))
	}
	return sets
}
func Acec_ManPrintRanks2(pairs *Vec_Int_t) {
	for i := 1; i < vec.Vec_IntSize(pairs); i += 2 {
		fmt.Printf("%d ", vec.Vec_IntEntry(pairs, i))
	}
	fmt.Println()
}
func Acec_ManPrintRanks(pairs *Vec_Int_t) {
	counts := vec.Vec_IntStart(100)
	max := 0
	for i := 1; i < vec.Vec_IntSize(pairs); i += 2 {
		rank := int(vec.Vec_IntEntry(pairs, i))
		vec.Vec_IntFillExtra(counts, rank+1, 0)
		vec.Vec_IntAddToEntry(counts, rank, 1)
		if rank > max {
			max = rank
		}
	}
	for i := 1; i <= max; i++ {
		fmt.Printf("%2d=%2d ", i, vec.Vec_IntEntry(counts, i))
	}
	fmt.Println()
}
func Acec_ManProfile(p *Gia_Man_t, verbose int32) {
	start := time.Now()
	var xors *Vec_Int_t
	adds := Ree_ManComputeCuts(p, &xors, verbose)
	n := Ree_ManCountFadds(adds)
	fmt.Printf("Detected %d full-adders and %d half-adders. Found %d XOR-cuts. Time = %v\n", n, int32(vec.Vec_IntSize(adds)/6)-n, vec.Vec_IntSize(xors)/4, time.Since(start))
	start = time.Now()
	sets := Acec_ManCollectBoxSets(p, adds, xors)
	fmt.Printf("Detected %d adder-trees. Time = %v\n", vec.Vec_WecSize(sets)/5, time.Since(start))
	if verbose != 0 {
		for i := 0; 5*i < vec.Vec_WecSize(sets); i++ {
			fmt.Printf("Tree %3d : Xor = %4d Root = %4d Adder = %4d In = %4d Out = %4d\n", i, vec.Vec_IntSize(vec.Vec_WecEntry(sets, 5*i)), vec.Vec_IntSize(vec.Vec_WecEntry(sets, 5*i+1)), vec.Vec_IntSize(vec.Vec_WecEntry(sets, 5*i+2)), vec.Vec_IntSize(vec.Vec_WecEntry(sets, 5*i+3))/2, vec.Vec_IntSize(vec.Vec_WecEntry(sets, 5*i+4))/2)
			Acec_ManPrintRanks(vec.Vec_WecEntry(sets, 5*i+3))
			Acec_ManPrintRanks(vec.Vec_WecEntry(sets, 5*i+4))
		}
	}
}
func Acec_ManPool(p *Gia_Man_t) {
	start := time.Now()
	adds := Ree_ManComputeCuts(p, nil, 1)
	n := Ree_ManCountFadds(adds)
	fmt.Printf("Detected %d FAs and %d HAs. Time = %v\n", n, int32(vec.Vec_IntSize(adds)/6)-n, time.Since(start))
	tops := Acec_ManPoolTopMost(p, adds)
	fmt.Printf("Detected %d topmost adders.\n", vec.Vec_IntSize(tops))
	trees := Gia_PolynCoreOrderArray(p, adds, tops)
	for i := 0; i < vec.Vec_WecSize(trees); i++ {
		fmt.Printf("Adder %5d : Tree with %5d nodes.\n", vec.Vec_IntEntry(tops, i), vec.Vec_IntSize(vec.Vec_WecEntry(trees, i)))
	}
}
