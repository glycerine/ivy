package abc

// Byte-lane lookup counting from opt/dau/dauCount.c:
// Table generation and Abc_TtCountOnesInCofsQuick_rec/Quick.

import (
	"fmt"
	"math/bits"
	"strings"
)

// truthQuickByteCounts packs three low-input counts, four byte-index input
// counts and a total into byte lanes. Generated once, then read-only.
var truthQuickByteCounts = func() [256][32]uint64 {
	var table [256][32]uint64
	masks := [3]uint8{0x55, 0x33, 0x0f}
	for value := 0; value < 256; value++ {
		count := uint64(bits.OnesCount8(uint8(value)))
		for index := 0; index < 32; index++ {
			word := count << 56
			for v, mask := range masks {
				word |= uint64(bits.OnesCount8(uint8(value)&mask)) << uint(v*8)
			}
			for v := 3; v < 7; v++ {
				if index&(1<<uint(v-3)) == 0 {
					word |= count << uint(v*8)
				}
			}
			table[value][index] = word
		}
	}
	return table
}()

// TruthQuickByteCountTable returns an owned value snapshot of the generated
// 256-by-32 lookup. It replaces Abc_TtCountGenerate's process-global stdout
// emission while retaining direct inspection and reproducible regeneration.
func TruthQuickByteCountTable() [256][32]uint64 {
	return truthQuickByteCounts
}

// FormatTruthQuickByteCountTable returns a stable, language-neutral table: one
// row per byte value and 32 uppercase fixed-width hexadecimal words per row.
func FormatTruthQuickByteCountTable() string {
	var output strings.Builder
	for value, row := range truthQuickByteCounts {
		fmt.Fprintf(&output, "%02X:", value)
		for _, word := range row {
			fmt.Fprintf(&output, " %016X", word)
		}
		output.WriteByte('\n')
	}
	return output.String()
}

// TruthQuickCofactorPopulations follows Quick's packed convention for one or
// more inputs (including small-table repetitions). Zero inputs count the one
// logical constant bit, as C does. uint64 counts avoid signed overflow.
type TruthQuickCofactorPopulations struct {
	Ones      uint64
	Zero, One []uint64
}

func truthQuickCofactorCounts(words []uint64, variables int, zero []uint64) uint64 {
	if variables > 8 {
		half := len(words) / 2
		a := truthQuickCofactorCounts(words[:half], variables-1, zero)
		b := truthQuickCofactorCounts(words[half:], variables-1, zero)
		// C assigns here, losing earlier sibling contributions to this input.
		zero[variables-1] += a
		return a + b
	}
	bytes := 8
	if variables > 6 {
		bytes = 1 << uint(variables-3)
	}
	var sums [2]uint64
	for i := 0; i < bytes; i++ {
		value := uint8(words[i/8] >> uint((i%8)*8))
		sums[i/(bytes/2)] += truthQuickByteCounts[value][i]
	}
	// Each half contains at most 128 minterms, so its packed lanes cannot
	// overflow. Decode before combining to fix C's 256-total byte overflow.
	for v := 0; v < min(variables, 7); v++ {
		zero[v] += ((sums[0] >> uint(v*8)) & 255) + ((sums[1] >> uint(v*8)) & 255)
	}
	if variables == 8 {
		zero[7] += sums[0] >> 56
	}
	return (sums[0] >> 56) + (sums[1] >> 56)
}

// CofactorPopulationsQuick ports C's lookup/recursive algorithm, correcting
// full-eight-input total overflow and overwritten recursive cofactor sums.
// The table uses explicit numeric byte extraction, independent of host byte
// order and pointer aliasing. Results own storage; source is unchanged.
func (t *TruthTable) CofactorPopulationsQuick() (TruthQuickCofactorPopulations, error) {
	if e := t.Check(); e != nil {
		return TruthQuickCofactorPopulations{}, e
	}
	r := TruthQuickCofactorPopulations{Zero: make([]uint64, t.variables), One: make([]uint64, t.variables)}
	if t.variables == 0 {
		r.Ones = t.words[0] & 1
		return r, nil
	}
	r.Ones = truthQuickCofactorCounts(t.words, t.variables, r.Zero)
	for i, count := range r.Zero {
		r.One[i] = r.Ones - count
	}
	return r, nil
}
