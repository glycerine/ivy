package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/opt/dau"
	"io"
	"strings"
)

func ifDsdPrimeCover(truth []word, n int32) ([]int32, int32) {
	input := make([]uint32, 2*len(truth))
	for i, w := range truth {
		input[2*i], input[2*i+1] = uint32(w), uint32(w>>32)
	}
	var cover []uint32
	complement := kit.Kit_TruthIsop(input, n, &cover, 1)
	cubes := make([]int32, len(cover))
	for i, c := range cover {
		cubes[i] = int32(c)
	}
	return cubes, complement
}
func If_DsdManComputeTruth_rec(p *If_DsdMan_t, lit int32, result []word, perm []byte, support *int32) {
	o := dsdObject(p, lit)
	switch o.Type {
	case IF_DSD_VAR:
		pin := 2 * *support
		if perm != nil {
			pin = int32(perm[*support])
		}
		*support++
		assert(*support <= p.nVars, "DSD truth support fits manager")
		for i := range result {
			result[i] = p.pTtElems[pin>>1][i]
			if (pin^lit)&1 != 0 {
				result[i] = ^result[i]
			}
		}
		return
	case IF_DSD_AND, IF_DSD_XOR:
		clear(result)
		if o.Type == IF_DSD_AND {
			for i := range result {
				result[i] = ^word(0)
			}
		}
		temp := make([]word, len(result))
		for _, fan := range o.pFans {
			If_DsdManComputeTruth_rec(p, fan, temp, perm, support)
			for i := range result {
				if o.Type == IF_DSD_AND {
					result[i] &= temp[i]
				} else {
					result[i] ^= temp[i]
				}
			}
		}
	case IF_DSD_MUX, IF_DSD_PRIME:
		fanins := make([][]word, o.nFans)
		for i, fan := range o.pFans {
			fanins[i] = make([]word, len(result))
			If_DsdManComputeTruth_rec(p, fan, fanins[i], perm, support)
		}
		if o.Type == IF_DSD_MUX {
			for i := range result {
				result[i] = fanins[0][i]&fanins[1][i] | ^fanins[0][i]&fanins[2][i]
			}
		} else {
			dau.Dau_DsdTruthCompose_rec(If_DsdObjTruth(p, o), fanins, result, o.nFans, int32(len(result)))
		}
	default:
		panic("if: invalid canonical DSD truth node")
	}
	if lit&1 != 0 {
		for i := range result {
			result[i] = ^result[i]
		}
	}
}
func If_DsdManComputeTruthPtr(p *If_DsdMan_t, lit int32, perm []byte, result []word) {
	result = result[:p.nWords]
	if lit < 2 {
		for i := range result {
			result[i] = 0
			if lit == 1 {
				result[i] = ^word(0)
			}
		}
		return
	}
	var support int32
	If_DsdManComputeTruth_rec(p, lit, result, perm, &support)
	assert(support == dsdObject(p, lit).nSupp, "DSD truth visits complete support")
}
func If_DsdManComputeTruth(p *If_DsdMan_t, lit int32, perm []byte) []word {
	result := p.pTtElems[p.nVars]
	If_DsdManComputeTruthPtr(p, lit, perm, result)
	return result
}
func ifDsdTextRec(p *If_DsdMan_t, lit int32, perm []byte, support *int32, out *strings.Builder) {
	if lit&1 != 0 {
		out.WriteByte('!')
	}
	o := dsdObject(p, lit)
	if o.Type == IF_DSD_CONST0 {
		out.WriteByte('0')
		return
	}
	if o.Type == IF_DSD_VAR {
		pin := 2 * *support
		if perm != nil {
			pin = int32(perm[*support])
		}
		*support++
		if pin&1 != 0 {
			out.WriteByte('!')
		}
		out.WriteByte(byte('a' + pin>>1))
		return
	}
	open, close := byte('('), byte(')')
	switch o.Type {
	case IF_DSD_XOR:
		open, close = '[', ']'
	case IF_DSD_MUX:
		open, close = '<', '>'
	case IF_DSD_PRIME:
		open, close = '{', '}'
		truth := If_DsdObjTruth(p, o)
		if o.nFans <= 6 {
			width := 1 << uint(max(o.nFans-2, 0))
			mask := ^word(0)
			if o.nFans < 6 {
				mask = word(1)<<uint(1<<uint(o.nFans)) - 1
			}
			fmt.Fprintf(out, "%0*X", width, truth[0]&mask)
		} else {
			for i := len(truth) - 1; i >= 0; i-- {
				fmt.Fprintf(out, "%016X", truth[i])
			}
		}
	}
	out.WriteByte(open)
	for _, fan := range o.pFans {
		ifDsdTextRec(p, fan, perm, support, out)
	}
	out.WriteByte(close)
}
func ifDsdText(p *If_DsdMan_t, lit int32, perm []byte) string {
	var result strings.Builder
	var support int32
	ifDsdTextRec(p, lit, perm, &support, &result)
	return result.String()
}

// If_DsdManPrintOne prints one canonical record in the original source format.
func If_DsdManPrintOne(out io.Writer, p *If_DsdMan_t, id int32, perm []byte, newLine int32) {
	o := If_DsdVecObj(&p.vObjs, id)
	var text strings.Builder
	var support int32
	ifDsdTextRec(p, id<<1, perm, &support, &text)
	assert(support == o.nSupp, "printed DSD visits complete support")
	fmt.Fprintf(out, "%6d : %2d %8d %d  %s", id, o.nSupp, o.Count, o.fMark, text.String())
	if newLine != 0 {
		fmt.Fprintln(out)
	}
}
