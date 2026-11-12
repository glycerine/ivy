package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"math/rand"
)

// Gia_ManAutomSimulate prints the outputs for a sequence of alphabet symbols.
func Gia_ManAutomSimulate(p *Gia_Man_t, alpha *Vec_Int_t, sequence string) {
	if Gia_ManPiNum(p) != int32(vec.Vec_IntSize(alpha)) || Gia_ManPoNum(p) < 1 {
		panic("gia: incompatible automaton alphabet")
	}
	fmt.Printf("Simulating string %q:\n", sequence)
	state := make([]bool, Gia_ManRegNum(p))
	input := make([]bool, Gia_ManCiNum(p))
	for frame, c := range []byte(sequence) {
		clear(input)
		if index := vec.Vec_IntFind(alpha, int32(c)); index >= 0 {
			input[index] = true
		}
		copy(input[Gia_ManPiNum(p):], state)
		outputs, err := p.Evaluate(input)
		if err != nil {
			panic(err)
		}
		copy(state, outputs[Gia_ManPoNum(p):])
		value := 0
		if outputs[0] {
			value = 1
		}
		fmt.Printf("Frame %d : %c %d\n", frame, c, value)
	}
}

// Gia_ManAutomInit identifies complemented register inputs in the graph.
func Gia_ManAutomInit(p *Gia_Man_t) word {
	var initial word
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) == 0 {
			continue
		}
		a, b, _, _ := p.Gate(id)
		for _, lit := range []int32{a, b} {
			if lit&1 == 0 {
				continue
			}
			obj := Gia_ManObj(p, lit>>1)
			if Gia_ObjIsCi(obj) == 0 {
				continue
			}
			index := Gia_ObjCioId(obj) - Gia_ManPiNum(p)
			if index >= 0 && index < 64 {
				initial |= word(1) << (63 - index)
			}
		}
	}
	return initial
}

func Gia_ManAutomWalkOne(p *Gia_Man_t, steps int32, states *Vec_Wrd_t, counts *Vec_Int_t, temp *Vec_Wrd_t, initial word) {
	var state word
	var next [64]word
	for i := int32(0); i < steps; i++ {
		output := Gia_ManAutomStep(p, state, next[:], temp)
		selected, index, best := -1, -1, int32(math.MaxInt32)
		for k := int32(0); k < Gia_ManPiNum(p); k++ {
			if next[k] == initial {
				continue
			}
			idx := vec.Vec_WrdFind(states, next[k])
			count := int32(0)
			if idx >= 0 {
				count = vec.Vec_IntEntry(counts, idx)
			}
			if best > count || (best != math.MaxInt32 && count != 0 && float64(best)/float64(count) > rand.Float64()) {
				selected, index, best = int(k), idx, count
			}
			if best == 0 {
				break
			}
		}
		if selected < 0 {
			for k := int32(0); k < Gia_ManPiNum(p); k++ {
				if output&(word(1)<<(63-k)) != 0 {
					fmt.Printf("%c!", 'a'+k)
				}
			}
			break
		}
		if index < 0 {
			index = vec.Vec_IntSize(counts)
			vec.Vec_IntPush(counts, 0)
			vec.Vec_WrdPush(states, next[selected])
		}
		vec.Vec_IntAddToEntry(counts, index, 1)
		state = next[selected]
		fmt.Printf("%c", 'a'+selected)
		if output&(word(1)<<(63-selected)) != 0 {
			fmt.Print("!")
		}
	}
	fmt.Println()
}
func Gia_ManAutomWalk(p *Gia_Man_t, steps, walks, verbose int32) {
	if Gia_ManPoNum(p) != 1 {
		fmt.Println("AIG should have one primary output.")
		return
	}
	if Gia_ManPiNum(p) > 64 {
		fmt.Println("Cannot simulate an automaton with more than 64 inputs.")
		return
	}
	if Gia_ManRegNum(p) > 64 {
		fmt.Println("Cannot simulate an automaton with more than 64 states.")
		return
	}
	temp := vec.Vec_WrdStart(int(p.nObjs))
	states := vec.Vec_WrdAlloc(1000)
	counts := vec.Vec_IntAlloc(1000)
	vec.Vec_WrdPush(states, 0)
	vec.Vec_IntPush(counts, 1)
	initial := Gia_ManAutomInit(p)
	for i := int32(0); i < walks; i++ {
		Gia_ManAutomWalkOne(p, steps, states, counts, temp, initial)
	}
	if verbose != 0 {
		for i, state := range vec.Vec_WrdArray(states) {
			fmt.Printf("%3d : %064b %d\n", i, state^initial, vec.Vec_IntEntry(counts, i))
		}
		fmt.Println()
	}
}
