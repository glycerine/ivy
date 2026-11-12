package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func giaRelationDimensions(rel *Vec_Wrd_t, outputs int32) (int32, int32) {
	if outputs < 0 || outputs > 20 {
		panic("gia: invalid relation output count")
	}
	mints := int32(1) << uint(outputs)
	if vec.Vec_WrdSize(rel)%int(mints) != 0 {
		panic("gia: invalid relation size")
	}
	return mints, int32(vec.Vec_WrdSize(rel)) / mints
}
func Gia_ManSimRelCheckFuncs(p *Gia_Man_t, rel *Vec_Wrd_t, outputs int32, funcs *Vec_Wrd_t) {
	mints, words := giaRelationDimensions(rel, outputs)
	if vec.Vec_WrdSize(funcs) != int(2*outputs*words) {
		panic("gia: invalid relation function size")
	}
	data, functions := vec.Vec_WrdArray(rel), vec.Vec_WrdArray(funcs)
	var errors int32
	for pattern := int32(0); pattern < 64*words; pattern++ {
		for mint := int32(0); mint < mints; mint++ {
			allowed := true
			for k := int32(0); k < outputs; k++ {
				base := 2 * k * words
				zero := giaSimBit(functions[base:base+words], pattern)
				one := giaSimBit(functions[base+words:base+2*words], pattern)
				if zero != 0 && one != 0 {
					panic("gia: contradictory function outputs")
				}
				if zero != 0 && mint&(1<<uint(k)) != 0 || one != 0 && mint&(1<<uint(k)) == 0 {
					allowed = false
					break
				}
			}
			if allowed && giaSimBit(data, pattern*mints+mint) == 0 {
				if errors == 0 {
					fmt.Printf("For pattern %d, minterm %d produced by function is not in the relation.\n", pattern, mint)
				}
				errors++
			}
		}
	}
	if errors != 0 {
		fmt.Printf("Total number of similar errors = %d.\n", errors)
	} else {
		fmt.Println("The function agrees with the relation.")
	}
}
func giaRelationDerive(p *Gia_Man_t, rel *Vec_Wrd_t, outputs int32, allowDC bool) *Vec_Wrd_t {
	mints, words := giaRelationDimensions(rel, outputs)
	funcs := vec.Vec_WrdStart(int(2 * outputs * words))
	data, out := vec.Vec_WrdArray(rel), vec.Vec_WrdArray(funcs)
	dcs := make([]int32, outputs)
	var missing int32
	for pattern := int32(0); pattern < 64*words; pattern++ {
		mint := int32(0)
		for ; mint < mints; mint++ {
			if giaSimBit(data, pattern*mints+mint) != 0 {
				break
			}
		}
		if mint == mints {
			missing++
			continue
		}
		for k := int32(0); k < outputs; k++ {
			if allowDC && giaSimBit(data, pattern*mints+(mint^(1<<uint(k)))) != 0 {
				dcs[k]++
				continue
			}
			phase := (mint >> uint(k)) & 1
			bit := int32(2)*k*words + phase*words + (pattern >> 6)
			out[bit] |= uint64(1) << uint(pattern&63)
		}
	}
	if missing != 0 {
		fmt.Printf("The relation is not well-defined for %d (out of %d) patterns.\n", missing, 64*words)
	} else if allowDC {
		fmt.Printf("The relation was successfully determized with don't-cares for %d patterns.\n", 64*words)
		for k := int32(0); k < outputs; k++ {
			var zero, one int
			for w := int32(0); w < words; w++ {
				zero += bits.OnesCount64(out[(2*k)*words+w])
				one += bits.OnesCount64(out[(2*k+1)*words+w])
			}
			fmt.Printf("%4d : Off = %6d  On = %6d  Dc = %6d (%6.2f %%)\n", k, zero, one, dcs[k], 100*float64(dcs[k])/float64(64*words))
		}
	} else {
		fmt.Printf("The relation was successfully determized without don't-cares for %d patterns.\n", 64*words)
	}
	Gia_ManSimRelCheckFuncs(p, rel, outputs, funcs)
	return funcs
}
func Gia_ManSimRelDeriveFuncs(p *Gia_Man_t, rel *Vec_Wrd_t, outputs int32) *Vec_Wrd_t {
	return giaRelationDerive(p, rel, outputs, false)
}
func Gia_ManSimRelDeriveFuncs2(p *Gia_Man_t, rel *Vec_Wrd_t, outputs int32) *Vec_Wrd_t {
	return giaRelationDerive(p, rel, outputs, true)
}
