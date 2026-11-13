package acec

// XOR-rooted arithmetic box construction from src/proof/acec/acecXor.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Acec_FindXorLeaves(p *Gia_Man_t, xors, adds, roots, ranks *Vec_Int_t, outBoxes **Vec_Wec_t) *Vec_Wec_t {
	xm := Acec_MapXorOuts2(p, xors, ranks)
	mm := Acec_MapMajOuts2(p, adds, ranks)
	leaves := vec.Vec_WecStart(vec.Vec_IntSize(roots))
	boxes := vec.Vec_WecStart(vec.Vec_IntSize(roots))
	for i := 0; 4*i < vec.Vec_IntSize(xors); i++ {
		id := vec.Vec_IntEntry(xors, 4*i)
		rank := vec.Vec_IntEntry(ranks, int(id))
		if rank == -1 {
			continue
		}
		for k := 1; k < 4; k++ {
			f := vec.Vec_IntEntry(xors, 4*i+k)
			if f == 0 {
				continue
			}
			if vec.Vec_BitEntry(xm, int(f)) != 0 {
				if rank != vec.Vec_IntEntry(ranks, int(f)) {
					panic("acec: XOR tree rank mismatch")
				}
				continue
			}
			box := vec.Vec_IntEntry(mm, int(f))
			if box == -1 {
				vec.Vec_WecPush(leaves, int(rank), f)
			} else if rank > 0 {
				vec.Vec_WecPush(boxes, int(rank-1), box)
			}
		}
	}
	if outBoxes != nil {
		*outBoxes = boxes
	}
	return leaves
}
func Acec_CheckBoothPPs(p *Gia_Man_t, levels *Vec_Wec_t) {
	marked := Acec_MultMarkPPs(p)
	for i := 0; i < vec.Vec_WecSize(levels); i++ {
		level := vec.Vec_WecEntry(levels, i)
		pi, booth, other := 0, 0, 0
		for _, lit := range vec.Vec_IntArray(level) {
			id := lit >> 1
			if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) == 0 {
				pi++
			} else if vec.Vec_BitEntry(marked, int(id)) != 0 {
				booth++
			} else {
				other++
			}
		}
		fmt.Printf("Rank %2d : Lits = %5d    PI = %d  Booth = %5d  Non-Booth = %5d\n", i, vec.Vec_IntSize(level), pi, booth, other)
	}
}
func Acec_FindBox(p *Gia_Man_t, adds *Vec_Int_t, boxes, leaves *Vec_Wec_t, xorRoots *Vec_Int_t) *Acec_Box_t {
	n := vec.Vec_WecSize(boxes)
	if n == 0 {
		return nil
	}
	if n != vec.Vec_WecSize(leaves) || n != vec.Vec_IntSize(xorRoots) {
		panic("acec: incompatible XOR box levels")
	}
	visited := vec.Vec_BitStart(vec.Vec_IntSize(adds) / 6)
	isLeaf, isRoot := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	box := &Acec_Box_t{pGia: p, vAdds: boxes, vLeafLits: vec.Vec_WecStart(n), vRootLits: vec.Vec_WecStart(n)}
	for i := 0; i < n; i++ {
		for _, b := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			for k := 0; k < 3; k++ {
				vec.Vec_BitWriteEntry(isLeaf, int(vec.Vec_IntEntry(adds, int(6*b)+k)), 1)
			}
			for k := 3; k < 5; k++ {
				vec.Vec_BitWriteEntry(isRoot, int(vec.Vec_IntEntry(adds, int(6*b)+k)), 1)
			}
		}
		vec.Vec_IntSort(vec.Vec_WecEntry(boxes, i), 0)
	}
	m := Acec_TreeCarryMap(p, adds, boxes)
	for i := n - 1; i >= 0; i-- {
		for _, b := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			carry := vec.Vec_IntEntry(adds, int(6*b+4))
			if vec.Vec_BitEntry(isLeaf, int(carry)) == 0 {
				phase := int32(0)
				if vec.Vec_IntEntry(adds, int(6*b+2)) != 0 {
					phase = 1
				}
				Acec_TreePhases_rec(p, adds, m, carry, phase, visited)
			}
		}
	}
	Acec_TreeVerifyPhases(p, adds, boxes)
	Acec_TreeVerifyPhases2(p, adds, boxes)
	vec.Vec_BitWriteEntry(isRoot, 0, 1)
	for i := 0; i < n; i++ {
		for _, b := range vec.Vec_IntArray(vec.Vec_WecEntry(boxes, i)) {
			for k := int32(0); k < 3; k++ {
				id := vec.Vec_IntEntry(adds, int(6*b+k))
				if vec.Vec_BitEntry(isRoot, int(id)) == 0 {
					vec.Vec_WecPush(box.vLeafLits, i, id<<1|Acec_SignBit2(adds, b, k))
				}
			}
			for k := int32(3); k < 5; k++ {
				id := vec.Vec_IntEntry(adds, int(6*b+k))
				if vec.Vec_BitEntry(isLeaf, int(id)) == 0 {
					rank := i
					if k == 4 {
						rank++
					}
					vec.Vec_WecPush(box.vRootLits, rank, id<<1|Acec_SignBit2(adds, b, k))
				}
			}
			if vec.Vec_IntEntry(adds, int(6*b+2)) == 0 && Acec_SignBit2(adds, b, 2) != 0 {
				vec.Vec_WecPush(box.vLeafLits, i, 1)
			}
		}
	}
	last := vec.Vec_WecEntry(box.vLeafLits, n-1)
	xl := vec.Vec_WecEntry(leaves, n-1)
	if vec.Vec_IntSize(last) == 0 && vec.Vec_IntSize(xl) > 0 {
		for _, id := range vec.Vec_IntArray(xl) {
			vec.Vec_IntPush(last, id<<1)
		}
	}
	vec.Vec_IntFill(vec.Vec_WecEntry(box.vRootLits, n-1), 1, vec.Vec_IntEntryLast(xorRoots)<<1)
	for i := 0; i < n; i++ {
		vec.Vec_IntSort(vec.Vec_WecEntry(box.vLeafLits, i), 0)
		vec.Vec_IntSort(vec.Vec_WecEntry(box.vRootLits, i), 1)
	}
	return box
}
func Acec_ProduceBox(p *Gia_Man_t, verbose int32) *Acec_Box_t {
	start := time.Now()
	if gia.Gia_ManCoNum(p) < 2 {
		return nil
	}
	var xors *Vec_Int_t
	adds := Ree_ManComputeCuts(p, &xors, 0)
	roots := Acec_FindXorRoots(p, xors)
	if vec.Vec_IntSize(roots) == 0 {
		return nil
	}
	ranks := Acec_RankTrees(p, xors, roots)
	gia.Gia_ManLevelNum(p)
	if verbose != 0 {
		n := Ree_ManCountFadds(adds)
		fmt.Printf("Detected %d full-adders and %d half-adders. Found %d XOR-cuts. Time = %v\n", n, int32(vec.Vec_IntSize(adds)/6)-n, vec.Vec_IntSize(xors)/4, time.Since(start))
	}
	roots = Acec_OrderTreeRoots(p, adds, roots, ranks)
	ranks = Acec_RankTrees(p, xors, roots)
	var boxes *Vec_Wec_t
	leaves := Acec_FindXorLeaves(p, xors, adds, roots, ranks, &boxes)
	if vec.Vec_WecSizeSize(boxes) == 0 {
		return nil
	}
	Acec_TreeVerifyConnections(p, adds, boxes)
	box := Acec_FindBox(p, adds, boxes, leaves, roots)
	if box != nil && verbose != 0 {
		Acec_TreePrintBox(box, adds)
	}
	return box
}
