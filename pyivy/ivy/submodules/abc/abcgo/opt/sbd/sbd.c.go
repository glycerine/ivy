package sbd

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/glycerine/abcgo/misc/vec"
)

func Sbd_CountConfigVars(set *vec.Vec_Int_t, vars, degree int32) int32 {
	values := vec.Vec_IntArray(set)
	if len(values) == 0 {
		return 0
	}
	if vars < 0 || degree < 1 {
		panic("invalid topology dimensions")
	}
	previous := int64(values[0])
	count, sum := int64(0), int64(0)
	for _, v := range values {
		if v < 1 {
			panic("invalid topology level")
		}
		sum += int64(v)
	}
	for i, entry := range values[1:] {
		if int64(degree)*previous < int64(entry) {
			panic("topology exceeds fanin capacity")
		}
		if int64(degree)*previous == int64(entry) {
			previous = int64(entry)
			continue
		}
		below := int64(vars)
		for _, tail := range values[i+1:] {
			below += int64(tail)
		}
		factor := int64(degree)*previous - 1
		if sum >= int64(vars)-1 {
			if below != 0 && factor > (math.MaxInt32-count)/below {
				panic("topology parameter count overflow")
			}
			count += below * factor
		}
		previous = int64(entry)
	}
	if sum < int64(vars)-1 {
		return 0
	}
	factor := int64(degree) * previous
	if vars != 0 && factor > (math.MaxInt32-count)/int64(vars) {
		panic("topology parameter count overflow")
	}
	count += int64(vars) * factor
	if count > math.MaxInt32 {
		panic("topology parameter count overflow")
	}
	return int32(count)
}
func sbdTopologyKey(values []int32) string {
	key := make([]byte, 4*len(values))
	for i, v := range values {
		binary.LittleEndian.PutUint32(key[4*i:], uint32(v))
	}
	return string(key)
}
func sbdCountTopologies(nodes, degree int32) [][][]int32 {
	if nodes < 0 || degree < 1 {
		panic("invalid topology dimensions")
	}
	states := [][]int32{nil}
	seen := map[string]bool{"": true}
	result := make([][][]int32, nodes)
	start := 0
	for size := int32(0); size < nodes; size++ {
		stop := len(states)
		for e := start; e < stop; e++ {
			state := append([]int32(nil), states[e]...)
			add := func(candidate []int32) {
				key := sbdTopologyKey(candidate)
				if !seen[key] {
					seen[key] = true
					states = append(states, append([]int32(nil), candidate...))
				}
			}
			for k := range state {
				if k == 0 && state[0] > 0 || k > 0 && state[k] == degree*state[k-1] {
					continue
				}
				state[k]++
				add(state)
				state[k]--
			}
			add(append(state, 1))
		}
		result[size] = append([][]int32(nil), states[stop:]...)
		start = stop
	}
	return result
}
func Sbd_CountTopos() {
	fmt.Println("Counting topologies for 9 inputs and 10 degree-3 nodes.")
	all := 1
	for size, states := range sbdCountTopologies(10, 3) {
		all += len(states)
		fmt.Printf("Nodes = %2d : This = %8d  All = %8d\n", size+1, len(states), all)
		for _, state := range states {
			set := vec.Vec_IntAlloc(len(state))
			for _, v := range state {
				vec.Vec_IntPush(set, v)
			}
			fmt.Printf("Params = %3d.  %v\n", Sbd_CountConfigVars(set, 9, 3), state)
		}
		fmt.Println()
	}
}
