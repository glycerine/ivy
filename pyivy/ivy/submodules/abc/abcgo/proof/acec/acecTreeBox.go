package acec

// Arithmetic box construction from src/proof/acec/acecTree.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Acec_PrintAdders(boxes *Vec_Wec_t, adds *Vec_Int_t) {
	for i := 0; i < vec.Vec_WecSize(boxes); i++ {
		level := vec.Vec_WecEntry(boxes, i)
		fmt.Printf(" %4d : %2d  {", i, vec.Vec_IntSize(level))
		for _, box := range vec.Vec_IntArray(level) {
			mark := ""
			if vec.Vec_IntEntry(adds, int(6*box+2)) == 0 {
				mark = "*"
			}
			fmt.Printf(" %s%d=(%d,%d)", mark, box, vec.Vec_IntEntry(adds, int(6*box+3)), vec.Vec_IntEntry(adds, int(6*box+4)))
		}
		fmt.Println(" }")
	}
}
func Acec_TreePrintBox(box *Acec_Box_t, adds *Vec_Int_t) {
	fmt.Println("Adders:")
	Acec_PrintAdders(box.vAdds, adds)
	fmt.Println("Inputs:")
	vec.Vec_WecPrintLits(box.vLeafLits)
	fmt.Println("Outputs:")
	vec.Vec_WecPrintLits(box.vRootLits)
}
func Acec_CreateBoxMaxRank(tree *Vec_Int_t) int32 {
	var max int32
	for i := 1; i < vec.Vec_IntSize(tree); i += 2 {
		if rank := vec.Vec_IntEntry(tree, i); rank > max {
			max = rank
		}
	}
	return max
}
func Acec_CreateBox(p *Gia_Man_t, adds, tree *Vec_Int_t) *Acec_Box_t {
	max := Acec_CreateBoxMaxRank(tree)
	visited := vec.Vec_BitStart(vec.Vec_IntSize(adds) / 6)
	leaves, roots := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	box := &Acec_Box_t{pGia: p, vAdds: vec.Vec_WecStart(int(max + 1)), vLeafLits: vec.Vec_WecStart(int(max + 1)), vRootLits: vec.Vec_WecStart(int(max + 2))}
	for i := 0; i < vec.Vec_IntSize(tree); i += 2 {
		b, rank := vec.Vec_IntEntry(tree, i), vec.Vec_IntEntry(tree, i+1)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(leaves, int(vec.Vec_IntEntry(adds, int(6*b)+k)), 1)
		}
		for k := 3; k < 5; k++ {
			vec.Vec_BitWriteEntry(roots, int(vec.Vec_IntEntry(adds, int(6*b)+k)), 1)
		}
		vec.Vec_WecPush(box.vAdds, int(rank), b)
	}
	for i := 0; i < vec.Vec_WecSize(box.vAdds); i++ {
		vec.Vec_IntSort(vec.Vec_WecEntry(box.vAdds, i), 0)
	}
	m := Acec_TreeCarryMap(p, adds, box.vAdds)
	for i := vec.Vec_WecSize(box.vAdds) - 1; i >= 0; i-- {
		for _, b := range vec.Vec_IntArray(vec.Vec_WecEntry(box.vAdds, i)) {
			carry := vec.Vec_IntEntry(adds, int(6*b+4))
			if vec.Vec_BitEntry(leaves, int(carry)) == 0 {
				phase := int32(0)
				if vec.Vec_IntEntry(adds, int(6*b+2)) != 0 {
					phase = 1
				}
				Acec_TreePhases_rec(p, adds, m, carry, phase, visited)
			}
		}
	}
	Acec_TreeVerifyPhases(p, adds, box.vAdds)
	Acec_TreeVerifyPhases2(p, adds, box.vAdds)
	vec.Vec_BitWriteEntry(roots, 0, 1)
	for i := 0; i < vec.Vec_WecSize(box.vAdds); i++ {
		for _, b := range vec.Vec_IntArray(vec.Vec_WecEntry(box.vAdds, i)) {
			for k := int32(0); k < 3; k++ {
				id := vec.Vec_IntEntry(adds, int(6*b+k))
				if vec.Vec_BitEntry(roots, int(id)) == 0 {
					vec.Vec_WecPush(box.vLeafLits, i, id<<1|Acec_SignBit2(adds, b, k))
				}
			}
			for k := int32(3); k < 5; k++ {
				id := vec.Vec_IntEntry(adds, int(6*b+k))
				if vec.Vec_BitEntry(leaves, int(id)) == 0 {
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
	for i := 0; i < vec.Vec_WecSize(box.vLeafLits); i++ {
		vec.Vec_IntSort(vec.Vec_WecEntry(box.vLeafLits, i), 0)
	}
	for i := 0; i < vec.Vec_WecSize(box.vRootLits); i++ {
		vec.Vec_IntSort(vec.Vec_WecEntry(box.vRootLits, i), 1)
	}
	return box
}
func Acec_CreateBoxTest(p *Gia_Man_t) {
	start := time.Now()
	adds := Ree_ManComputeCuts(p, nil, 1)
	n := Ree_ManCountFadds(adds)
	fmt.Printf("Detected %d adders (%d FAs and %d HAs). Time = %v\n", vec.Vec_IntSize(adds)/6, n, int32(vec.Vec_IntSize(adds)/6)-n, time.Since(start))
	start = time.Now()
	trees := Acec_TreeFindTrees(p, adds, nil, 0, 0)
	fmt.Printf("Collected %d trees with %d adders in them. Time = %v\n", vec.Vec_WecSize(trees), vec.Vec_WecSizeSize(trees)/2, time.Since(start))
	for i := 0; i < vec.Vec_WecSize(trees); i++ {
		box := Acec_CreateBox(p, adds, vec.Vec_WecEntry(trees, i))
		fmt.Printf("Processing tree %d: Ranks = %d. Adders = %d. Leaves = %d. Roots = %d.\n", i, vec.Vec_WecSize(box.vAdds), vec.Vec_WecSizeSize(box.vAdds), vec.Vec_WecSizeSize(box.vLeafLits), vec.Vec_WecSizeSize(box.vRootLits))
		Acec_TreePrintBox(box, adds)
		Acec_BoxFree(box)
	}
}
func Acec_DeriveBox(p *Gia_Man_t, ignore *Vec_Bit_t, filterIn, filterOut, verbose int32) *Acec_Box_t {

	adds := Ree_ManComputeCuts(p, nil, verbose)
	trees := Acec_TreeFindTrees(p, adds, ignore, filterIn, filterOut)
	var box *Acec_Box_t
	if vec.Vec_WecSize(trees) > 0 {
		box = Acec_CreateBox(p, adds, vec.Vec_WecEntry(trees, 0))
		Acec_VerifyBoxLeaves(box, ignore)
	}
	if box != nil {
		fmt.Printf("Processing tree 0: Ranks = %d. Adders = %d. Leaves = %d. Roots = %d.\n", vec.Vec_WecSize(box.vAdds), vec.Vec_WecSizeSize(box.vAdds), vec.Vec_WecSizeSize(box.vLeafLits), vec.Vec_WecSizeSize(box.vRootLits))
		if verbose != 0 {
			Acec_TreePrintBox(box, adds)
		}
	}
	return box
}
