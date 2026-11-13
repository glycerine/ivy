package acec

// Carry-cone construction from src/proof/acec/acecCl.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// acecUniquePolarity checks all paths before building replacement nodes. A
// reconvergent leaf may occur in both polarities, in which case the source's
// carry rewrite cannot safely choose one.
func acecUniquePolarity(p *Gia_Man_t, node, leaf int32) (int32, bool) {
	if node < leaf {
		return -1, true
	}
	if node == leaf {
		return node << 1, true
	}
	o := gia.Gia_ManObj(p, node)
	if gia.Gia_ObjIsAnd(o) == 0 {
		return -1, true
	}
	a, ok := acecUniquePolarity(p, gia.Gia_ObjFaninId0(o, node), leaf)
	if !ok {
		return -1, false
	}
	b, ok := acecUniquePolarity(p, gia.Gia_ObjFaninId1(o, node), leaf)
	if !ok {
		return -1, false
	}
	if a != -1 {
		a ^= gia.Gia_ObjFaninC0(o)
	}
	if b != -1 {
		b ^= gia.Gia_ObjFaninC1(o)
	}
	if a == -1 {
		return b, true
	}
	if b == -1 {
		return a, true
	}
	if a != b {
		return -1, false
	}
	return a, true
}
func acecXorRootPolarities(p *Gia_Man_t, roots *Vec_Int_t) ([][3]int32, bool) {
	values := make([][3]int32, vec.Vec_IntSize(roots)/4)
	for i := range values {
		root := vec.Vec_IntEntry(roots, 4*i)
		for k := 0; k < 3; k++ {
			lit, ok := acecUniquePolarity(p, root, vec.Vec_IntEntry(roots, 4*i+k+1))
			if !ok || lit < 0 {
				return nil, false
			}
			values[i][k] = lit
		}
	}
	return values, true
}

// Acec_DetectXorBuildNew preserves the original graph when a reconvergent
// XOR input has no unique path polarity for the source carry rewrite.
func Acec_DetectXorBuildNew(p *Gia_Man_t, roots *Vec_Int_t) *Gia_Man_t {
	polarities, ok := acecXorRootPolarities(p, roots)
	if !ok {
		return gia.Gia_ManDup(p)
	}
	mirrors := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i, lits := range polarities {
		or := gia.Gia_ManAppendOr(p, lits[1], lits[2])
		a := gia.Gia_ManAppendAnd(p, lits[0], or)
		b := gia.Gia_ManAppendAnd(p, lits[1], lits[2])
		carry := gia.Gia_ManAppendOr(p, a, b)
		vec.Vec_IntWriteEntry(mirrors, int(vec.Vec_IntEntry(roots, 4*i+1)), carry)
	}
	vec.Vec_IntFillExtra(mirrors, int(gia.Gia_ManObjNum(p)), -1)
	return Acec_ManDerive(p, mirrors)
}

func Acec_DetectAdditional(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	start := time.Now()
	roots := Acec_CollectXorTops(p)
	var dst *Gia_Man_t
	count := 0
	if roots != nil {
		count = vec.Vec_IntSize(roots) / 4
		if _, ok := acecXorRootPolarities(p, roots); !ok {
			return gia.Gia_ManDup(p)
		}
		Acec_DetectComputeSupports(p, roots)
		dst = Acec_DetectXorBuildNew(p, roots)
	} else {
		dst = gia.Gia_ManDup(p)
	}
	fmt.Printf("Detected %d top XORs.  Time = %v\n", count, time.Since(start))
	return dst
}
