package amap

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
)

// Amap_LibParseEquations parses every usable gate function into a Hop graph
// and stores the resulting truth table on its gate. Gates that exceed the
// mapper's supported input limit are retained in the library but skipped for
// matching, as in the C implementation.
func Amap_LibParseEquations(lib *Amap_Lib_t, verbose int32) int32 {
	if lib == nil || lib.vGates == nil {
		return 0
	}
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	for i := int32(0); i < Vec_PtrSize(lib.vGates); i++ {
		gate, ok := Vec_PtrEntry(lib.vGates, i).(*Amap_Gat_t)
		if !ok || gate == nil {
			return 0
		}
		pins := int(gate.nPins)
		if pins == 0 {
			switch gate.pForm {
			case AMAP_STRING_CONST0:
				gate.pFunc = []uint32{0}
			case AMAP_STRING_CONST1:
				gate.pFunc = []uint32{^uint32(0)}
			default:
				if verbose != 0 {
					fmt.Printf("Cannot parse formula %q of gate %q with no pins.\n", gate.pForm, gate.pName)
				}
				return 0
			}
			continue
		}
		if pins > AMAP_MAXINS {
			continue
		}
		if len(gate.Pins) < pins {
			if verbose != 0 {
				fmt.Printf("Gate %q has %d declared pins but only %d pin names.\n", gate.pName, pins, len(gate.Pins))
			}
			return 0
		}
		pinNames := make([]string, pins)
		for j := range pinNames {
			pinNames[j] = gate.Pins[j].pName
		}
		root, err := Amap_ParseFormulaToHop(nil, gate.pForm, pinNames, manager, gate.pName)
		if err != nil {
			if verbose != 0 {
				fmt.Println(err)
			}
			return 0
		}
		truth := hop.Hop_ManConvertAigToTruth(manager, root, int32(pins), nil, 0)
		if !amapTruthUsesEveryPin(truth, pins) {
			if verbose != 0 {
				fmt.Printf("Skipping gate %q because its output does not depend on every input pin.\n", gate.pName)
			}
			gate.pFunc = nil
			continue
		}
		gate.pFunc = append(gate.pFunc[:0], truth...)
	}
	return 1
}

func amapTruthUsesEveryPin(truth []uint32, nPins int) bool {
	if nPins <= 0 {
		return true
	}
	for pin := 0; pin < nPins; pin++ {
		mask := uint32(1) << uint(pin)
		depends := false
		for assignment := 0; assignment < 1<<nPins && !depends; assignment++ {
			if assignment&int(mask) != 0 {
				continue
			}
			other := assignment | int(mask)
			if truthValue(truth, assignment) != truthValue(truth, other) {
				depends = true
			}
		}
		if !depends {
			return false
		}
	}
	return true
}

func truthValue(words []uint32, assignment int) bool {
	word := assignment >> 5
	if word >= len(words) {
		return false
	}
	return words[word]&(uint32(1)<<uint(assignment&31)) != 0
}
