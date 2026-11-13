package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"math/rand"
	"testing"
)

// Source rarity2 scores complete groups from the last simulated frame, but
// transfers every register, including a final incomplete group.
func TestRarity2ScoresAndTransfersAgainstPatternOracle(t *testing.T) {
	r := rand.New(rand.NewSource(5112))
	for trial := 0; trial < 30; trial++ {
		g := aig.NewGraph()
		regs := make([]aig.Edge, 5)
		for i := range regs {
			regs[i] = g.NewInput()
		}
		g.AddOutput(g.Const0())
		outs := make([]aig.Edge, 5)
		for i := range outs {
			outs[i] = g.AddOutput(regs[i])
		}
		aig.Aig_ManSetRegNum(g, 5)
		p := sswRar2Start(g, 2, 2, 2, false)
		p.inits = vec.Vec_IntAlloc(10)
		states := make([][5]bool, 64)
		for i, out := range outs {
			words := Ssw_SmlSimInfo(p.sim, out)
			words[0] = r.Uint32()
			words[1] = r.Uint32()
			for pat := 0; pat < 64; pat++ {
				states[pat][i] = r.Intn(2) != 0
				if states[pat][i] {
					words[2+pat/32] |= 1 << uint(pat%32)
				}
			}
		}
		counts := [2][4]int{}
		for _, state := range states {
			for group := 0; group < 2; group++ {
				value := 0
				if state[2*group] {
					value |= 1
				}
				if state[2*group+1] {
					value |= 2
				}
				counts[group][value]++
			}
		}
		p.updateCounters()
		for group := 0; group < 2; group++ {
			for value := 0; value < 4; value++ {
				if got := p.rarity[group*4+value]; got != counts[group][value] {
					t.Fatalf("trial %d group %d value %d got %d want %d", trial, group, value, got, counts[group][value])
				}
			}
		}
		scores := make([]float64, 64)
		for pat, state := range states {
			for group := 0; group < 2; group++ {
				value := 0
				if state[2*group] {
					value |= 1
				}
				if state[2*group+1] {
					value |= 2
				}
				count := float64(counts[group][value])
				scores[pat] += 1 / (count * count)
			}
		}
		want := make([]int32, 0, 10)
		for i := 0; i < 2; i++ {
			best := 0
			for j := 1; j < 64; j++ {
				if scores[j] > scores[best] {
					best = j
				}
			}
			scores[best] = math.Inf(-1)
			for _, v := range states[best] {
				if v {
					want = append(want, 1)
				} else {
					want = append(want, 0)
				}
			}
		}
		p.transferPatterns()
		got := vec.Vec_IntArray(p.inits)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("trial %d transfer %v want %v", trial, got, want)
			}
		}
	}
}
