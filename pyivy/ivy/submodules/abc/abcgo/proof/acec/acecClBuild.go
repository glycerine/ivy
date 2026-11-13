package acec

// Carry-cone construction from src/proof/acec/acecCl.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Acec_DetectXorBuildNew(p *Gia_Man_t, roots *Vec_Int_t) *Gia_Man_t {
	// The source initializes the mirror table to zero, so unmapped nodes
	// are redirected to constant zero rather than preserved.
	mirrors := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_IntSize(roots)/4; i++ {
		var lits [3]int32
		root := vec.Vec_IntEntry(roots, 4*i)
		for k := range lits {
			lits[k] = Acec_DetectLitPolarity(p, root, vec.Vec_IntEntry(roots, 4*i+k+1))
		}
		or := gia.Gia_ManAppendOr(p, lits[1], lits[2])
		a := gia.Gia_ManAppendAnd(p, lits[0], or)
		b := gia.Gia_ManAppendAnd(p, lits[1], lits[2])
		carry := gia.Gia_ManAppendOr(p, a, b)
		vec.Vec_IntWriteEntry(mirrors, int(vec.Vec_IntEntry(roots, 4*i+1)), carry)
	}
	vec.Vec_IntFillExtra(mirrors, int(gia.Gia_ManObjNum(p)), 0)
	return Acec_ManDerive(p, mirrors)
}

func Acec_DetectAdditional(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	start := time.Now()
	roots := Acec_CollectXorTops(p)
	var dst *Gia_Man_t
	count := 0
	if roots != nil {
		count = vec.Vec_IntSize(roots) / 4
		Acec_DetectComputeSupports(p, roots)
		dst = Acec_DetectXorBuildNew(p, roots)
	} else {
		dst = gia.Gia_ManDup(p)
	}
	fmt.Printf("Detected %d top XORs.  Time = %v\n", count, time.Since(start))
	return dst
}
