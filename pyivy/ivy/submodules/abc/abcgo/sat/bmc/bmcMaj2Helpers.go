package bmc

import (
	"fmt"
	"math/bits"
	"math/rand"
)

// Maj_ManValueV2 returns the majority value of the low nVars bits.
func Maj_ManValueV2(minterm, nVars int32) int32 {
	if nVars < 0 || nVars > 32 {
		panic("bmc: invalid majority input count")
	}
	var mask uint32 = ^uint32(0)
	if nVars < 32 {
		mask = (uint32(1) << uint(nVars)) - 1
	}
	if bits.OnesCount32(uint32(minterm)&mask) > int(nVars/2) {
		return 1
	}
	return 0
}

// Maj_ManConnect creates forward edges between majority nodes. The original
// generator retries impossible requests forever, so this checks capacity.
func Maj_ManConnect(connections [][3]int32, nVars, nObjs, nRands, verbose int32) {
	if nVars < 0 || nObjs < 0 || int(nObjs) > len(connections) || nRands < 0 {
		panic("bmc: invalid topology")
	}
	first := nVars + 2
	for node := nObjs - 2; node >= first; node-- {
		var choices [][2]int32
		for dst := node + 1; dst < nObjs; dst++ {
			for slot := int32(2); slot >= 0; slot-- {
				if connections[dst][slot] == 0 {
					choices = append(choices, [2]int32{dst, slot})
					break
				}
			}
		}
		if len(choices) == 0 {
			panic("bmc: topology has no free fanin")
		}
		c := choices[rand.Intn(len(choices))]
		connections[c[0]][c[1]] = node
		if verbose != 0 {
			fmt.Printf("%d -> %d  ", node, c[0])
		}
	}
	for r := int32(0); r < nRands; r++ {
		var choices [][3]int32
		for node := first; node < nObjs-1; node++ {
			for dst := node + 1; dst < nObjs; dst++ {
				slot := int32(-1)
				duplicate := false
				for k := int32(2); k >= 0; k-- {
					if connections[dst][k] == node {
						duplicate = true
					}
					if connections[dst][k] == 0 && slot < 0 {
						slot = k
					}
				}
				if !duplicate && slot >= 0 {
					choices = append(choices, [3]int32{node, dst, slot})
				}
			}
		}
		if len(choices) == 0 {
			panic("bmc: topology has insufficient free connections")
		}
		c := choices[rand.Intn(len(choices))]
		connections[c[1]][c[2]] = c[0]
		if verbose != 0 {
			fmt.Printf("+%d -> %d  ", c[0], c[1])
		}
	}
	if verbose != 0 {
		fmt.Println()
	}
}
func Maj_ManConnect2(c [][3]int32, nVars, nObjs, nRands int32) {
	c[10][2] = 9
	c[12][2] = 11
	c[13][2] = 9
	c[13][1] = 10
	c[14][2] = 11
	c[14][1] = 12
	c[15][2] = 13
	c[15][1] = 14
}
