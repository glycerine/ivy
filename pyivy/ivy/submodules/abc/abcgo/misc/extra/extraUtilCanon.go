package extra

import (
	"fmt"
	"slices"
)

/**CFile****************************************************************

  FileName    [extraUtilMisc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Computing canonical forms of Boolean functions using truth tables.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilMisc.c,v 1.0 2003/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

type canonNegationResult struct {
	low, high uint32
	phases    []byte
}

// Extra_TruthCanonFastN minimizes a truth table over input negations only.
// It returns all minimizing input phases in the original recursive order.
// Results are owned by the caller, unlike the C function's shared static arrays.
// Truth tables use the same 32-bit repeated-small-function convention as ABC.
// As in C, real counts below three use the three-variable lookup domain.
func Extra_TruthCanonFastN(maxVars, real int32, truth []uint32) (canonical []uint32, phases []byte) {
	if maxVars < 0 || maxVars > 6 || real < 0 || real > maxVars {
		panic("invalid fast canonicalization dimensions")
	}
	value := uint64(truth[0])
	if real == 6 {
		value |= uint64(truth[1]) << 32
	}
	result := canonNegationRec(max(real, int32(3)), value)
	canonical = []uint32{result.low}
	if maxVars == 6 {
		high := result.high
		if real < 6 {
			high = result.low
		}
		canonical = append(canonical, high)
	}
	return canonical, result.phases
}
func canonHighPhases(input []byte, bit int32) []byte {
	result := slices.Clone(input)
	for i := range result {
		result[i] |= 1 << uint(bit)
	}
	return result
}
func canonPhase32(truth uint32, phase byte, n int32) uint32 {
	for v := int32(0); v < n; v++ {
		if phase&(1<<uint(v)) != 0 {
			mask := truthVariableMasks[v]
			shift := uint(1) << uint(v)
			truth = (truth&mask)>>shift | (truth&^mask)<<shift
		}
	}
	return truth
}
func canonNegation3(truth byte) canonNegationResult {
	best := uint32(255)
	var phases []byte
	for phase := byte(0); phase < 8; phase++ {
		value := canonPhase32(uint32(truth)*0x01010101, phase, 3) & 255
		if value < best {
			best = value
			phases = phases[:0]
		}
		if value == best {
			phases = append(phases, phase)
		}
	}
	return canonNegationResult{low: best * 0x01010101, phases: phases}
}
func canonNegationRec(n int32, truth uint64) canonNegationResult {
	if n == 3 {
		return canonNegation3(byte(truth))
	}
	child := n - 1
	var initial0, initial1 uint32
	switch child {
	case 3:
		initial0 = uint32(byte(truth)) * 0x01010101
		initial1 = uint32(byte(truth>>8)) * 0x01010101
	case 4:
		initial0 = uint32(uint16(truth)) * 0x00010001
		initial1 = uint32(uint16(truth>>16)) * 0x00010001
	case 5:
		initial0 = uint32(truth)
		initial1 = uint32(truth >> 32)
	default:
		panic("invalid recursive canonicalization level")
	}
	var low, high uint32
	var phases []byte
	switch {
	case initial0 == initial1 && (initial0 == 0 || initial0 == ^uint32(0)):
		result := canonNegationRec(child, uint64(initial0))
		low, high = result.low, result.low
		phases = append(slices.Clone(result.phases), canonHighPhases(result.phases, child)...)
	case initial1 == 0:
		result := canonNegationRec(child, uint64(initial0))
		low, high, phases = result.low, 0, result.phases
	case initial0 == 0:
		result := canonNegationRec(child, uint64(initial1))
		low, high, phases = result.low, 0, canonHighPhases(result.phases, child)
	case initial1 == ^uint32(0):
		result := canonNegationRec(child, uint64(initial0))
		low, high, phases = initial1, result.low, canonHighPhases(result.phases, child)
	case initial0 == ^uint32(0):
		result := canonNegationRec(child, uint64(initial1))
		low, high, phases = initial0, result.low, result.phases
	default:
		zero := canonNegationRec(child, uint64(initial0))
		one := canonNegationRec(child, uint64(initial1))
		if one.low == zero.low {
			phases = append(phases, one.phases...)
			phases = append(phases, canonHighPhases(zero.phases, child)...)
			low, high = canonPhase32(initial0, one.phases[0], child), zero.low
		} else {
			candidates, value, phaseBit := one.phases, initial0, byte(0)
			high = one.low
			if one.low > zero.low {
				candidates, value, phaseBit = zero.phases, initial1, 1<<uint(child)
				high = zero.low
			}
			low = ^uint32(0)
			for _, phase := range candidates {
				current := canonPhase32(value, phase, child)
				if current < low {
					low = current
					phases = phases[:0]
				}
				if current == low {
					phases = append(phases, phase|phaseBit)
				}
			}
		}
	}
	result := canonNegationResult{phases: phases}
	switch child {
	case 3:
		value := (high&255)<<8 | (low & 255)
		result.low = value<<16 | value
	case 4:
		result.low = (high&65535)<<16 | (low & 65535)
	case 5:
		result.low, result.high = low, high
	}
	return result
}
func Map_Var3Print() {
	for i := 0; i < 256; i++ {
		if i%8 == 0 {
			fmt.Println()
		}
		result := canonNegation3(byte(i))
		fmt.Printf("0x%08X, ", result.low)
	}
	fmt.Println()
	for i := 0; i < 256; i++ {
		phases := canonNegation3(byte(i)).phases
		fmt.Printf("%3d */  { %2d,   ", i, len(phases))
		for k, p := range phases {
			if k != 0 {
				fmt.Print(", ")
			}
			fmt.Print(p)
		}
		fmt.Println(" }")
	}
}
func Map_Var3Test() {
	for i := uint32(0); i < 256; i++ {
		Extra_TruthCanonFastN(5, 3, []uint32{i})
	}
}
func Map_Var4Test() {
	for i := uint32(0); i < 65536; i++ {
		Extra_TruthCanonFastN(5, 4, []uint32{i})
	}
}
