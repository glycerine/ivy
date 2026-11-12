package extra

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

/**CFile****************************************************************

  FileName    [extraUtilCube.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Permutation computation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilCube.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

func Abc_StatePush(data *vec.Vec_Int_t, state []byte, k int32) {
	state = state[:24]
	for i := 0; i < 6; i++ {
		vec.Vec_IntWriteEntry(data, 6*int(k)+i, int32(binary.LittleEndian.Uint32(state[4*i:])))
	}
}
func Abc_StatePerm(state, permutation, result []byte) {
	state = state[:24]
	permutation = permutation[:24]
	result = result[:24]
	var next [24]byte
	for i, v := range permutation {
		next[i] = state[v]
	}
	copy(result, next[:])
}
func Abc_StatePrint(state []byte) {
	for _, value := range state[:24] {
		fmt.Printf(" %2d", value)
	}
	fmt.Println()
}
func Abc_DataHasBit(data []uint64, bit uint64) int32 {
	if data[bit>>6]&(uint64(1)<<uint(bit&63)) != 0 {
		return 1
	}
	return 0
}
func Abc_DataXorBit(data []uint64, bit uint64) { data[bit>>6] ^= uint64(1) << uint(bit&63) }
func Abc_DataGetCube(word uint64, i int32) int32 {
	if i < 0 || i > 12 {
		panic("cube field outside packed word")
	}
	return int32(word >> uint(5*i) & 31)
}
func Abc_DataXorCube(word uint64, i, c int32) uint64 {
	if i < 0 || i > 12 || c < 0 || c > 31 {
		panic("invalid cube field")
	}
	return word ^ uint64(c)<<uint(5*i)
}
func Abc_CubeGenerateSign(state []byte) uint64 {
	var corners = [24]int32{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5, 5, -1, -1, -1, 6, 6, 6}
	positions := [7]int{0, 3, 6, 9, 12, 15, 21}
	state = state[:24]
	var result uint64
	for field, pos := range positions {
		value := state[pos]
		if value >= 24 || corners[value] < 0 {
			panic("fixed cube corner moved into a variable corner")
		}
		code := corners[value]<<2 | int32(value%3)
		result = Abc_DataXorCube(result, int32(field), code)
	}
	return result
}
func cubeQuarterMoves() [9][24]byte {
	var result [9][24]byte
	for axis, swaps := range zddCubeTurns {
		var state [24]byte
		for i := range state {
			state[i] = byte(i)
		}
		for power := 0; power < 3; power++ {
			for _, pair := range swaps {
				a, b := pair[0]-1, pair[1]-1
				state[a], state[b] = state[b], state[a]
			}
			result[3*power+axis] = state
		}
	}
	return result
}

type cubeSeenStates struct {
	full       map[[24]byte]struct{}
	signatures map[uint64]struct{}
}

func (s *cubeSeenStates) add(state [24]byte) bool {
	if s.signatures != nil {
		key := Abc_CubeGenerateSign(state[:])
		if _, seen := s.signatures[key]; seen {
			return false
		}
		s.signatures[key] = struct{}{}
		return true
	}
	if _, seen := s.full[state]; seen {
		return false
	}
	s.full[state] = struct{}{}
	return true
}

// Enumerate by complete BFS frontiers, using either full permutations or the
// original seven-corner signature. Sparse maps replace raw hash-pointer/dense
// 4GB bit-array storage; the state order and generated moves are unchanged.
func cubeEnumerate(rounds, limit int, signature, verbose bool) (counts []int, converged bool) {
	if rounds < 0 || limit < 1 {
		panic("invalid cube enumeration limits")
	}
	seen := cubeSeenStates{}
	if signature {
		seen.signatures = map[uint64]struct{}{}
	} else {
		seen.full = map[[24]byte]struct{}{}
	}
	var identity [24]byte
	for i := range identity {
		identity[i] = byte(i)
	}
	states := [][24]byte{identity}
	seen.add(identity)
	counts = []int{1}
	beg, end := 0, 1
	moves := cubeQuarterMoves()
	started := time.Now()
	if verbose {
		fmt.Println("Enumerating states of 2x2x2 cube.")
		fmt.Printf("Iter %2d -> %8d   Time = %v\n", 0, 1, time.Since(started))
	}
	for round := 1; round <= rounds; round++ {
		if beg == end {
			return counts, true
		}
		for k := beg; k < end; k++ {
			for _, move := range moves {
				var next [24]byte
				Abc_StatePerm(states[k][:], move[:], next[:])
				if !seen.add(next) {
					continue
				}
				if len(states) == limit {
					if verbose {
						fmt.Println("Did not converge.")
					}
					return counts, false
				}
				states = append(states, next)
			}
		}
		beg, end = end, len(states)
		counts = append(counts, end)
		if verbose {
			fmt.Printf("Iter %2d -> %8d   Time = %v\n", round, end, time.Since(started))
		}
	}
	return counts, beg == end
}
func Abc_EnumerateCubeStates2() { cubeEnumerate(100, 1<<22, false, true) }
func Abc_EnumerateCubeStates()  { cubeEnumerate(100, 1<<22, true, true) }
