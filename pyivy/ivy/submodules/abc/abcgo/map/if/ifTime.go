// Arrival and required-time propagation from ifTime.c.
package if_

import (
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func If_CutSortInputPins(p *If_Man_t, c *If_Cut_t, permutation []int32, delays []float32) {
	if len(permutation) < int(c.nLeaves) || len(delays) < int(c.nLeaves) {
		panic("short cut pin sorting buffers")
	}
	for pin := int32(0); pin < int32(c.nLeaves); pin++ {
		permutation[pin] = pin
		delays[pin] = If_ObjCutBest(If_CutLeaf(p, c, pin)).Delay
	}
	sort.SliceStable(permutation[:c.nLeaves], func(i, j int) bool { return delays[permutation[i]] > delays[permutation[j]] })
}
func ifTimingPinDelays(p *If_Man_t, c *If_Cut_t, reconstruct bool) []byte {
	if !reconstruct {
		return If_CutPerm(c)
	}
	var function func(*If_Man_t, *If_Cut_t, []byte) int32
	switch {
	case p.pPars.fDelayOpt != 0:
		function = If_CutSopBalancePinDelays
	case p.pPars.fDelayOptLut != 0:
		function = If_CutLutBalancePinDelays
	case p.pPars.fDsdBalance != 0:
		function = If_CutDsdBalancePinDelays
	default:
		return If_CutPerm(c)
	}
	permutation := make([]byte, c.nLeaves)
	delay := function(p, c, permutation)
	if delay < 0 || delay != int32(c.Delay) {
		panic("reconstructed pin delay disagrees with cut arrival")
	}
	return permutation
}
func ifTimingPinDelay(permutation []byte, pin int) float32 {
	if permutation == nil {
		return 1
	}
	if pin >= len(permutation) {
		panic("short user-cut pin profile")
	}
	if permutation[pin] == IF_BIG_CHAR {
		return -float32(IF_BIG_CHAR)
	}
	// C's pPerm field is `char *`; values such as Renode's -100 sentinel
	// are stored in the same byte array but read back as signed chars.
	return float32(int8(permutation[pin]))
}
func If_CutDelay(p *If_Man_t, _ *If_Obj_t, c *If_Cut_t) float32 {
	maximum := float32(-IF_FLOAT_LARGE)
	if c.fAndCut != 0 {
		for pin := int32(0); pin < int32(c.nLeaves); pin++ {
			maximum = max(maximum, If_ObjCutBest(If_CutLeaf(p, c, pin)).Delay+float32(p.pPars.nAndDelay))
		}
		return maximum
	}
	if library := p.pPars.pLutLib; library != nil {
		if p.pPars.fLiftLeaves != 0 {
			panic("LUT libraries do not support lifted cut leaves")
		}
		if c.nLeaves > uint32(library.LutMax) {
			panic("cut exceeds LUT library")
		}
		if library.fVarPinDelays != 0 {
			permutation, delays := make([]int32, c.nLeaves), make([]float32, c.nLeaves)
			If_CutSortInputPins(p, c, permutation, delays)
			for pin, index := range permutation {
				maximum = max(maximum, delays[index]+library.pLutDelays[c.nLeaves][pin])
			}
		} else {
			for pin := int32(0); pin < int32(c.nLeaves); pin++ {
				maximum = max(maximum, If_ObjCutBest(If_CutLeaf(p, c, pin)).Delay+library.pLutDelays[c.nLeaves][0])
			}
		}
		return maximum
	}
	permutation := If_CutPerm(c)
	for pin := int32(0); pin < int32(c.nLeaves); pin++ {
		arrival := If_ObjCutBest(If_CutLeaf(p, c, pin)).Delay
		delay := float32(1)
		if c.fUser != 0 {
			if p.pPars.fLiftLeaves != 0 {
				panic("user cuts do not support lifted leaves")
			}
			delay = ifTimingPinDelay(permutation, int(pin))
		} else if p.pPars.fLiftLeaves != 0 {
			arrival -= float32(c.pLeaves[pin]&255) * float32(p.Period)
		}
		maximum = max(maximum, arrival+delay)
	}
	return maximum
}
func If_CutPropagateRequired(p *If_Man_t, _ *If_Obj_t, c *If_Cut_t, required float32) {
	if p.pPars.fLiftLeaves != 0 {
		panic("required-time propagation does not support lifted leaves")
	}
	if c.fAndCut != 0 {
		for pin := int32(0); pin < int32(c.nLeaves); pin++ {
			leaf := If_CutLeaf(p, c, pin)
			leaf.Required = min(leaf.Required, required-float32(p.pPars.nAndDelay))
		}
		return
	}
	if library := p.pPars.pLutLib; library != nil {
		if c.nLeaves > uint32(library.LutMax) {
			panic("cut exceeds LUT library")
		}
		if library.fVarPinDelays != 0 {
			permutation, delays := make([]int32, c.nLeaves), make([]float32, c.nLeaves)
			If_CutSortInputPins(p, c, permutation, delays)
			for pin, index := range permutation {
				leaf := If_CutLeaf(p, c, index)
				leaf.Required = min(leaf.Required, required-library.pLutDelays[c.nLeaves][pin])
			}
		} else {
			for pin := int32(0); pin < int32(c.nLeaves); pin++ {
				leaf := If_CutLeaf(p, c, pin)
				leaf.Required = min(leaf.Required, required-library.pLutDelays[c.nLeaves][0])
			}
		}
		return
	}
	var permutation []byte
	if c.fUser != 0 {
		permutation = ifTimingPinDelays(p, c, true)
	}
	for pin := int32(0); pin < int32(c.nLeaves); pin++ {
		delay := float32(1)
		if c.fUser != 0 {
			delay = ifTimingPinDelay(permutation, int(pin))
		}
		leaf := If_CutLeaf(p, c, pin)
		leaf.Required = min(leaf.Required, required-delay)
	}
}
func ifTimingOutputs(p *If_Man_t, primaryOnly, latchesOnly bool) []*If_Obj_t {
	all := vec.Vec_PtrArray(p.vCos)
	start, end := 0, len(all)
	if primaryOnly {
		start = int(p.pPars.nLatchesCoBox)
		end -= int(p.pPars.nLatchesCo)
	} else if latchesOnly {
		start = len(all) - int(p.pPars.nLatchesCo)
	}
	if start < 0 || end < start || end > len(all) {
		panic("invalid sequential output counts")
	}
	outputs := make([]*If_Obj_t, 0, end-start)
	for _, entry := range all[start:end] {
		outputs = append(outputs, entry.(*If_Obj_t))
	}
	return outputs
}
func If_ManDelayMax(p *If_Man_t, sequential int32) float32 {
	if p.pPars.fLatchPaths != 0 && (p.pPars.nLatchesCi == 0 || p.pPars.nLatchesCo == 0) {
		fmt.Println("Delay optimization of latch paths disabled: no latches.")
		p.pPars.fLatchPaths = 0
	}
	if sequential != 0 && p.pPars.nLatchesCi <= 0 {
		panic("sequential delay requires latch inputs")
	}
	maximum := float32(-IF_FLOAT_LARGE)
	for _, output := range ifTimingOutputs(p, sequential != 0, sequential == 0 && p.pPars.fLatchPaths != 0) {
		maximum = max(maximum, If_ObjArrTime(If_ObjFanin0(output)))
	}
	return maximum
}
func ifTimingApplyTarget(p *If_Man_t) {
	pars := p.pPars
	if pars.nRelaxRatio != 0 && pars.DelayTargetNew == 0 {
		pars.DelayTargetNew = p.RequiredGlo * (100 + float32(pars.nRelaxRatio)) / 100
	}
	if pars.DelayTarget != -1 {
		if p.RequiredGlo > pars.DelayTarget+p.fEpsilon {
			if p.fNextRound == 0 {
				p.fNextRound = 1
				fmt.Printf("Cannot meet target required time %.2f; mapping continues.\n", pars.DelayTarget)
			}
		} else if p.RequiredGlo < pars.DelayTarget-p.fEpsilon {
			p.fNextRound = 1
			p.RequiredGlo = pars.DelayTarget
		}
	} else if pars.DelayTargetNew > 0 {
		p.RequiredGlo = pars.DelayTargetNew
	}
}
func ifTimingOutputRequired(p *If_Man_t, output *If_Obj_t) float32 {
	if p.pPars.fDoAverage != 0 {
		return If_ObjArrTime(If_ObjFanin0(output)) * (100 + float32(p.pPars.nRelaxRatio)) / 100
	}
	return p.RequiredGlo
}
func If_ManComputeRequired(p *If_Man_t) {
	If_ManMarkMapping(p)
	p.RequiredGlo = If_ManDelayMax(p, 0)
	outputs := ifTimingOutputs(p, false, false)
	if p.pManTim == nil {
		if p.pPars.pTimesReq != nil && p.pPars.fAreaOnly == 0 {
			if len(p.pPars.pTimesReq) < len(outputs) {
				panic("short output required-time array")
			}
			violations := 0
			for index, output := range outputs {
				driver := If_ObjFanin0(output)
				arrival := If_ObjArrTime(driver)
				required := p.pPars.pTimesReq[index]
				if arrival > required+p.fEpsilon {
					required = arrival
					violations++
				}
				driver.Required = min(driver.Required, required)
			}
			if violations > 0 && p.fReqTimeWarn == 0 {
				fmt.Printf("Required times exceeded at %d outputs; using earliest arrivals.\n", violations)
				p.fReqTimeWarn = 1
			}
		} else {
			ifTimingApplyTarget(p)
			if p.pPars.fAreaOnly != 0 {
				return
			}
			selected := outputs
			if p.pPars.fDoAverage == 0 && p.pPars.fLatchPaths != 0 {
				selected = ifTimingOutputs(p, false, true)
			}
			for _, output := range selected {
				driver := If_ObjFanin0(output)
				driver.Required = min(driver.Required, ifTimingOutputRequired(p, output))
			}
		}
		objects := vec.Vec_PtrArray(p.vObjs)
		for index := len(objects) - 1; index >= 0; index-- {
			object := objects[index].(*If_Obj_t)
			if object.Type == IF_AND && object.nRefs != 0 {
				If_CutPropagateRequired(p, object, If_ObjCutBest(object), object.Required)
			}
		}
		return
	}
	ifTimingApplyTarget(p)
	if p.pPars.fAreaOnly != 0 {
		return
	}
	tim.Tim_ManIncrementTravId(p.pManTim)
	switch {
	case p.vCoAttrs != nil:
		if vec.Vec_IntSize(p.vCoAttrs) != len(outputs) {
			panic("output timing-attribute count mismatch")
		}
		for index, output := range outputs {
			switch vec.Vec_IntEntry(p.vCoAttrs, index) {
			case -1:
				continue
			case 0:
				tim.Tim_ManSetCoRequired(p.pManTim, index, p.RequiredGlo)
			case 1:
				tim.Tim_ManSetCoRequired(p.pManTim, index, If_ObjArrTime(If_ObjFanin0(output)))
			case 2:
				tim.Tim_ManSetCoRequired(p.pManTim, index, IF_FLOAT_LARGE)
			default:
				panic("invalid output timing attribute")
			}
		}
	case p.pPars.fDoAverage != 0:
		for index, output := range outputs {
			tim.Tim_ManSetCoRequired(p.pManTim, index, ifTimingOutputRequired(p, output))
		}
	case p.pPars.fLatchPaths != 0:
		start, end := int(p.pPars.nLatchesCoBox), len(outputs)-int(p.pPars.nLatchesCo)
		for index := start; index < end; index++ {
			tim.Tim_ManSetCoRequired(p.pManTim, index, IF_FLOAT_LARGE)
		}
		for index := end; index < len(outputs); index++ {
			tim.Tim_ManSetCoRequired(p.pManTim, index, p.RequiredGlo)
		}
	default:
		tim.Tim_ManInitPoRequiredAll(p.pManTim, p.RequiredGlo)
	}
	objects := vec.Vec_PtrArray(p.vObjs)
	for index := len(objects) - 1; index >= 0; index-- {
		object := objects[index].(*If_Obj_t)
		switch object.Type {
		case IF_AND:
			if object.nRefs != 0 {
				If_CutPropagateRequired(p, object, If_ObjCutBest(object), object.Required)
			}
		case IF_CI:
			tim.Tim_ManSetCiRequired(p.pManTim, int(object.IdPio), object.Required)
		case IF_CO:
			required := tim.Tim_ManGetCoRequired(p.pManTim, int(object.IdPio))
			driver := If_ObjFanin0(object)
			driver.Required = min(driver.Required, required)
		case IF_CONST1:
		default:
			panic("unknown timing object type")
		}
	}
}
