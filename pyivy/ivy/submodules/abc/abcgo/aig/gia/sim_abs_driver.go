package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_SimAbsInit(p *Gia_SimAbsMan_t) {
	off, on := Gia_SimAbsFind(p.vValues, 0), Gia_SimAbsFind(p.vValues, 1)
	vec.Vec_IntClear(p.vPatPairs)
	n0, n1 := vec.Vec_IntSize(off), vec.Vec_IntSize(on)
	fmt.Printf("There are %d offset and %d onset minterms (%d pairs) and %d divisors.\n", n0, n1, n0*n1, p.nCands)
	if n0 == 0 || n1 == 0 {
		panic("gia: abstraction needs onset and offset patterns")
	}
	util.Abc_Random(1)
	for i := 0; i < 64; i++ {
		a := int(util.Abc_Random(0) % uint32(n0))
		b := int(util.Abc_Random(0) % uint32(n1))
		vec.Vec_IntPushTwo(p.vPatPairs, vec.Vec_IntEntry(off, a), vec.Vec_IntEntry(on, b))
	}
}

// Gia_SimAbsCheckSolution prints all one-, two-, or three-candidate covers.
func Gia_SimAbsCheckSolution(p *Gia_SimAbsMan_t) {
	if vec.Vec_WrdSize(p.vCoverTable) != int(p.nWordsTable*(p.nCands+1)) {
		panic("gia: invalid abstraction cover table")
	}
	table := vec.Vec_WrdArray(p.vCoverTable)
	target := make([]uint64, p.nWordsTable)
	for i := int32(0); i < int32(vec.Vec_IntSize(p.vPatPairs)/2); i++ {
		target[i>>6] |= uint64(1) << uint(i&63)
	}
	covers := func(ids ...int32) bool {
		for w := int32(0); w < p.nWordsTable; w++ {
			var union uint64
			for _, id := range ids {
				union |= table[id*p.nWordsTable+w]
			}
			if union != target[w] {
				return false
			}
		}
		return true
	}
	found := false
	for x := int32(0); x < p.nCands; x++ {
		if covers(x) {
			fmt.Printf("Found solution { %d }\n", x)
			found = true
		}
	}
	if found {
		return
	}
	for x := int32(0); x < p.nCands; x++ {
		for y := int32(0); y < x; y++ {
			if covers(x, y) {
				fmt.Printf("Found solution { %d %d }\n", y, x)
				found = true
			}
		}
	}
	if found {
		return
	}
	for x := int32(0); x < p.nCands; x++ {
		for y := int32(0); y < x; y++ {
			for z := int32(0); z < y; z++ {
				if covers(x, y, z) {
					fmt.Printf("Found solution { %d %d %d }\n", z, y, x)
				}
			}
		}
	}
}
func Gia_SimAbsPerformOne(g *Gia_Man_t, off, on []uint64, sims *Vec_Wrd_t, words, verbose int32) *Vec_Int_t {
	start := time.Now()
	result := vec.Vec_IntAlloc(10)
	p := Gia_SimAbsAlloc(g, off, on, sims, words, result, verbose)
	defer Gia_SimAbsFree(p)
	Gia_SimAbsInit(p)
	for {
		Gia_SimAbsSolve(p)
		if Gia_SimAbsRefine(p) == 0 {
			break
		}
	}
	fmt.Printf("Resubstitution time: %s\n", time.Since(start))
	return result
}
