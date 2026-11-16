package gia

import (
	"fmt"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand/v2"
	"time"
)

func Gia_IsoTestOld(p *Gia_Man_t, verbose int32) {
	start := time.Now()
	groups := Gia_IsoDeriveEquivPos(p, 0, verbose)
	count := 1
	if groups != nil {
		count = vec.Vec_PtrSize(groups)
	}
	fmt.Printf("Reduced %d outputs to %d.  Time = %.2f sec\n", Gia_ManPoNum(p), count, time.Since(start).Seconds())
	if verbose != 0 && groups != nil && int(Gia_ManPoNum(p)) != count {
		fmt.Println("Nontrivial classes:")
	}
}

// Keep C's full-range swap loop with Go's portable random source.
func Gia_IsoTestGenPerm(n int32) *Vec_Int_t {
	if n < 0 {
		panic("gia: negative permutation size")
	}
	result := vec.Vec_IntStartNatural(int(n))
	values := vec.Vec_IntArray(result)
	for i := range values {
		j := rand.IntN(len(values))
		values[i], values[j] = values[j], values[i]
	}
	return result
}
func Gia_IsoTest(p *Gia_Man_t, cex *Abc_Cex_t, verbose int32) {
	if Gia_ManPoNum(p) != 1 || Gia_ManRegNum(p) == 0 {
		panic("gia: isomorphism CEX test needs one output and registers")
	}
	permutation := Gia_IsoTestGenPerm(Gia_ManPiNum(p))
	fmt.Println("Considering random permutation of the primary inputs of the AIG:")
	vec.Vec_IntPrint(permutation)
	permuted := Gia_ManDupPerm(p, permutation)
	defer Gia_ManStop(permuted)
	doubled := Gia_ManDupAppendNew(p, permuted)
	defer Gia_ManStop(doubled)
	var groups, perms *Vec_Ptr_t
	reduced := Gia_ManIsoReduce(doubled, &groups, &perms, 0, 0, 0, 0)
	if reduced == nil {
		panic("gia: isomorphism reduction failed")
	}
	defer Gia_ManStop(reduced)
	old, ok0 := vec.Vec_PtrEntry(perms, 0).(*Vec_Int_t)
	next, ok1 := vec.Vec_PtrEntry(perms, 1).(*Vec_Int_t)
	if !ok0 || !ok1 {
		panic("gia: output permutations unavailable")
	}
	mapped := abcutil.Abc_CexPermuteTwo(cex, isoCexPIOrder(p, old), isoCexPIOrder(permuted, next))
	if Gia_ManVerifyCex(p, cex, 0) != 0 {
		fmt.Println("CEX for the init AIG is valid.")
	} else {
		fmt.Println("CEX for the init AIG is not valid.")
	}
	if Gia_ManVerifyCex(permuted, mapped, 0) != 0 {
		fmt.Println("CEX for the perm AIG is valid.")
	} else {
		fmt.Println("CEX for the perm AIG is not valid.")
	}
}

// isoCexPIOrder lifts cone-local canonical indices to the full PI interface.
// Pair unused inputs in increasing order; they cannot affect this output cone.
func isoCexPIOrder(graph *Gia_Man_t, local *Vec_Int_t) []int32 {
	count := int(Gia_ManPiNum(graph))
	if vec.Vec_IntSize(local) == count {
		return vec.Vec_IntArray(local)
	}
	inSupport := make([]bool, count)
	for _, id := range vec.Vec_IntArray(Gia_ManCollectSeq(graph, []int32{0}, 1)) {
		obj := Gia_ManObj(graph, id)
		if Gia_ObjIsPi(graph, obj) != 0 {
			inSupport[Gia_ObjCioId(obj)] = true
		}
	}
	support := make([]int32, 0, vec.Vec_IntSize(local))
	for pi, used := range inSupport {
		if used {
			support = append(support, int32(pi))
		}
	}
	if len(support) != vec.Vec_IntSize(local) {
		panic("gia: canonical PI order does not match sequential support")
	}
	result := make([]int32, 0, count)
	for _, index := range vec.Vec_IntArray(local) {
		result = append(result, support[index])
	}
	for pi, used := range inSupport {
		if !used {
			result = append(result, int32(pi))
		}
	}
	return result
}
