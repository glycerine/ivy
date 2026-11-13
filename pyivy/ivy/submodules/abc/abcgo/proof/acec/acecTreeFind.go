package acec

// Connected adder-tree discovery from src/proof/acec/acecTree.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Acec_TreeAddInOutPoint(m *Vec_Int_t, obj, box, out int32) {
	index := int(obj<<1 | out)
	old := vec.Vec_IntEntry(m, index)
	if old == -1 {
		vec.Vec_IntWriteEntry(m, index, box)
	} else if old >= 0 {
		vec.Vec_IntWriteEntry(m, index, -2)
	}
}
func Acec_TreeFindPoints(p *Gia_Man_t, adds *Vec_Int_t, ignore *Vec_Bit_t) *Vec_Int_t {
	m := vec.Vec_IntStartFull(2 * int(gia.Gia_ManObjNum(p)))
	for i := 0; 6*i < vec.Vec_IntSize(adds); i++ {
		if ignore != nil && (vec.Vec_BitEntry(ignore, int(vec.Vec_IntEntry(adds, 6*i+3))) != 0 || vec.Vec_BitEntry(ignore, int(vec.Vec_IntEntry(adds, 6*i+4))) != 0) {
			continue
		}
		for k := 0; k < 5; k++ {
			out := int32(0)
			if k >= 3 {
				out = 1
			}
			Acec_TreeAddInOutPoint(m, vec.Vec_IntEntry(adds, 6*i+k), int32(i), out)
		}
	}
	return m
}
func Acec_TreeWhichPoint(adds *Vec_Int_t, box, obj int32) int32 {
	for k := int32(0); k < 5; k++ {
		if vec.Vec_IntEntry(adds, int(6*box+k)) == obj {
			return k
		}
	}
	panic("acec: node absent from box")
}
func Acec_TreeFindTrees2_rec(adds, m *Vec_Int_t, box, rank int32, tree *Vec_Int_t, found *Vec_Bit_t) {
	if vec.Vec_BitEntry(found, int(box)) != 0 {
		return
	}
	vec.Vec_BitWriteEntry(found, int(box), 1)
	vec.Vec_IntPushTwo(tree, box, rank)
	for k := int32(0); k < 5; k++ {
		next := rank
		if k == 4 {
			next++
		}
		Acec_TreeFindTrees_rec(adds, m, vec.Vec_IntEntry(adds, int(6*box+k)), next, tree, found)
	}
}
func Acec_TreeFindTrees_rec(adds, m *Vec_Int_t, obj, rank int32, tree *Vec_Int_t, found *Vec_Bit_t) {
	in, out := vec.Vec_IntEntry(m, int(obj<<1|1)), vec.Vec_IntEntry(m, int(obj<<1))
	if in < 0 || out < 0 {
		return
	}
	inRank := rank
	if Acec_TreeWhichPoint(adds, in, obj) == 4 {
		inRank--
	}
	Acec_TreeFindTrees2_rec(adds, m, in, inRank, tree, found)
	Acec_TreeFindTrees2_rec(adds, m, out, rank, tree, found)
}
func Acec_TreeFindTrees(p *Gia_Man_t, adds *Vec_Int_t, ignore *Vec_Bit_t, filterIn, filterOut int32) *Vec_Wec_t {
	trees := vec.Vec_WecAlloc(10)
	m := Acec_TreeFindPoints(p, adds, ignore)
	found := vec.Vec_BitStart(vec.Vec_IntSize(adds) / 6)
	for i := 0; i < vec.Vec_IntSize(m); i += 2 {
		in, out := vec.Vec_IntEntry(m, i), vec.Vec_IntEntry(m, i+1)
		if in < 0 || out < 0 {
			continue
		}
		if vec.Vec_BitEntry(found, int(in)) != vec.Vec_BitEntry(found, int(out)) {
			panic("acec: inconsistent tree marks")
		}
		if vec.Vec_BitEntry(found, int(in)) != 0 {
			continue
		}
		tree := vec.Vec_WecPushLevel(trees)
		Acec_TreeFindTrees_rec(adds, m, int32(i/2), 0, tree, found)
		min := int32(1 << 30)
		for j := 1; j < vec.Vec_IntSize(tree); j += 2 {
			if r := vec.Vec_IntEntry(tree, j); r < min {
				min = r
			}
		}
		for j := 1; j < vec.Vec_IntSize(tree); j += 2 {
			vec.Vec_IntWriteEntry(tree, j, vec.Vec_IntEntry(tree, j)-min)
		}
	}
	if filterIn != 0 {
		Acec_TreeFilterTrees2(p, adds, trees)
	} else if filterOut != 0 {
		Acec_TreeFilterTrees(p, adds, trees)
	}
	vec.Vec_WecSort(trees, 1)
	return trees
}
func Acec_TreeFindTreesTest(p *Gia_Man_t) {
	start := time.Now()
	adds := Ree_ManComputeCuts(p, nil, 1)
	n := Ree_ManCountFadds(adds)
	fmt.Printf("Detected %d adders (%d FAs and %d HAs). Time = %v\n", vec.Vec_IntSize(adds)/6, n, int32(vec.Vec_IntSize(adds)/6)-n, time.Since(start))
	start = time.Now()
	trees := Acec_TreeFindTrees(p, adds, nil, 0, 0)
	fmt.Printf("Collected %d trees with %d adders in them. Time = %v\n", vec.Vec_WecSize(trees), vec.Vec_WecSizeSize(trees)/2, time.Since(start))
	vec.Vec_WecPrint(trees, 0)
}
