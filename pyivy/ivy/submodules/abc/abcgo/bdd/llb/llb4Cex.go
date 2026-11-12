package llb

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func llbCexSet(data []uint32, bit int)      { data[bit>>5] |= uint32(1) << uint(bit&31) }
func llbCexGet(data []uint32, bit int) bool { return data[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 }
func llbStateBits(value any, registers int) ([]bool, bool) {
	words, ok := value.([]uint32)
	if !ok {
		return nil, false
	}
	bits := make([]bool, registers)
	for i := range bits {
		bits[i] = i/32 < len(words) && words[i/32]&(uint32(1)<<uint(i%32)) != 0
	}
	return bits, true
}

// Llb4_Nonlin4TransformCex reconstructs a sequential trace by enumerating PI
// valuations for each transition. It is intended for the compact trace sizes
// produced by the reachability engine.
func Llb4_Nonlin4TransformCex(g *Aig_Man_t, states *Vec_Ptr_t, failedOutput, verbose int32) *Abc_Cex_t {
	if g == nil || states == nil || vec.Vec_PtrSize(states) == 0 {
		return nil
	}
	registers := int(g.RegisterCount())
	inputs, outputs := g.Inputs(), g.Outputs()
	pis := len(inputs) - registers
	properties := len(outputs) - registers
	if registers < 0 || pis < 0 || properties <= 0 || pis > 20 {
		return nil
	}
	frames := vec.Vec_PtrSize(states)
	cex := util.Abc_CexAlloc(int32(registers), int32(pis), int32(frames))
	if cex == nil {
		return nil
	}
	cex.IFrame, cex.IPo = int32(frames-1), -1
	current, ok := llbStateBits(vec.Vec_PtrEntry(states, 0), registers)
	if !ok {
		return nil
	}
	for i, value := range current {
		if value {
			llbCexSet(cex.PData, i)
		}
	}
	for frame := 0; frame < frames; frame++ {
		var wantedNext []bool
		if frame+1 < frames {
			wantedNext, ok = llbStateBits(vec.Vec_PtrEntry(states, frame+1), registers)
			if !ok {
				return nil
			}
		}
		matched := false
		selectedOutput := int32(-1)
		for mask := 0; mask < 1<<pis; mask++ {
			assignment := make([]bool, len(inputs))
			for i := 0; i < pis; i++ {
				assignment[i] = mask&(1<<i) != 0
			}
			copy(assignment[pis:], current)
			values, err := g.Evaluate(assignment)
			if err != nil {
				return nil
			}
			if frame+1 < frames {
				good := true
				for i := 0; i < registers; i++ {
					if values[properties+i] != wantedNext[i] {
						good = false
						break
					}
				}
				if !good {
					continue
				}
			} else {
				if failedOutput >= 0 && int(failedOutput) < properties {
					if !values[failedOutput] {
						continue
					}
					selectedOutput = failedOutput
				} else {
					selectedOutput = -1
					for i := 0; i < properties; i++ {
						if values[i] {
							selectedOutput = int32(i)
							break
						}
					}
					if selectedOutput < 0 {
						continue
					}
				}
			}
			for i := 0; i < pis; i++ {
				if assignment[i] {
					llbCexSet(cex.PData, registers+frame*pis+i)
				}
			}
			matched = true
			break
		}
		if !matched {
			return nil
		}
		if frame+1 == frames {
			cex.IPo = selectedOutput
		}
		if frame+1 < frames {
			current = wantedNext
		}
	}
	return cex
}

// Llb4_Nonlin4VerifyCex simulates each frame and returns the state snapshots.
func Llb4_Nonlin4VerifyCex(g *Aig_Man_t, cex *Abc_Cex_t) *Vec_Ptr_t {
	if g == nil || cex == nil {
		return nil
	}
	registers, pis := int(cex.NRegs), int(cex.NPis)
	if registers != int(g.RegisterCount()) || pis != len(g.Inputs())-registers {
		return nil
	}
	states := vec.Vec_PtrAlloc(int(cex.IFrame) + 1)
	current := make([]bool, registers)
	for i := 0; i < registers; i++ {
		current[i] = llbCexGet(cex.PData, i)
	}
	for frame := 0; frame <= int(cex.IFrame); frame++ {
		words := make([]uint32, (registers+31)/32)
		for i, v := range current {
			if v {
				words[i/32] |= 1 << uint(i%32)
			}
		}
		vec.Vec_PtrPush(states, words)
		assignment := make([]bool, pis+registers)
		for i := 0; i < pis; i++ {
			assignment[i] = llbCexGet(cex.PData, registers+frame*pis+i)
		}
		copy(assignment[pis:], current)
		values, err := g.Evaluate(assignment)
		if err != nil {
			return nil
		}
		for i := 0; i < registers; i++ {
			current[i] = values[len(values)-registers+i]
		}
	}
	return states
}

// Llb4_Nonlin4NormalizeCex replays the supplied trace against the original AIG
// and accepts it only if every frame remains valid.
func Llb4_Nonlin4NormalizeCex(original, reduced *Aig_Man_t, cex *Abc_Cex_t) *Abc_Cex_t {
	if original == nil || reduced == nil || cex == nil {
		return nil
	}
	if !llbCexValidForGraph(reduced, cex) {
		return nil
	}
	if !llbCexValidForGraph(original, cex) {
		return nil
	}
	return cex
}

func llbCexValidForGraph(g *Aig_Man_t, cex *Abc_Cex_t) bool {
	if g == nil || cex == nil || int(g.RegisterCount()) != int(cex.NRegs) || len(g.Inputs())-int(g.RegisterCount()) != int(cex.NPis) || cex.IFrame < 0 {
		return false
	}
	registers, pis := int(cex.NRegs), int(cex.NPis)
	current := make([]bool, registers)
	for i := range current {
		current[i] = llbCexGet(cex.PData, i)
	}
	properties := len(g.Outputs()) - registers
	for frame := 0; frame <= int(cex.IFrame); frame++ {
		assignment := make([]bool, pis+registers)
		for i := 0; i < pis; i++ {
			assignment[i] = llbCexGet(cex.PData, registers+frame*pis+i)
		}
		copy(assignment[pis:], current)
		values, err := g.Evaluate(assignment)
		if err != nil {
			return false
		}
		if frame == int(cex.IFrame) {
			if cex.IPo < 0 || int(cex.IPo) >= properties || !values[cex.IPo] {
				return false
			}
		}
		for i := 0; i < registers; i++ {
			current[i] = values[properties+i]
		}
	}
	return true
}

func Llb4_Nonlin4PrintCex(cex *Abc_Cex_t) {
	if cex != nil {
		fmt.Printf("counterexample: frame %d, output %d\n", cex.IFrame, cex.IPo)
	}
}
