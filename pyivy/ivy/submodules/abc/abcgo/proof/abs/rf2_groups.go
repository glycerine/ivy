package abs

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rf2_ManCountPpis(p *Rf2_Man_t) int32 {
	var count int32
	for _, id := range vec.Vec_IntArray(p.vMap) {
		if gia.Gia_ObjIsPi(p.pGia, gia.Gia_ManObj(p.pGia, id)) == 0 {
			count++
		}
	}
	return count
}
func Rf2_ManPrintVector(p *vec.Vec_Int_t, num int32) {
	for _, entry := range vec.Vec_IntArray(p) {
		for k := int32(0); k < num; k++ {
			fmt.Printf("%d", (entry>>k)&1)
		}
		fmt.Println()
	}
}
func Rf2_ManAssignJustIds(p *Rf2_Man_t) int32 {
	count := Rf2_ManCountPpis(p)
	groupSize := (count + 29) / 30
	vec.Vec_VecClear(p.vGrp2Ppi)
	if count == 0 {
		fmt.Println("Considering 0 PPIs combined into 0 groups of size 0.")
		return 0
	}
	var k int32
	for i, id := range vec.Vec_IntArray(p.vMap) {
		if gia.Gia_ObjIsPi(p.pGia, gia.Gia_ManObj(p.pGia, id)) == 0 {
			vec.Vec_VecPushInt(p.vGrp2Ppi, int(k/groupSize), int32(i))
			k++
		}
	}
	groups := (k-1)/groupSize + 1
	fmt.Printf("Considering %d PPIs combined into %d groups of size %d.\n", k, groups, groupSize)
	return groups
}
func Rf2_ManPrintVectorSpecial(p *Rf2_Man_t, masks *vec.Vec_Int_t) {
	groupSize := (Rf2_ManCountPpis(p) + 29) / 30
	for _, mask := range vec.Vec_IntArray(masks) {
		var k, count int32
		for _, id := range vec.Vec_IntArray(p.vMap) {
			if gia.Gia_ObjIsPi(p.pGia, gia.Gia_ManObj(p.pGia, id)) != 0 {
				fmt.Print("-")
				continue
			}
			if mask>>(k/groupSize)&1 != 0 {
				fmt.Print("1")
				count++
			} else {
				fmt.Print("0")
			}
			k++
		}
		fmt.Printf(" %3d \n", count)
	}
}
