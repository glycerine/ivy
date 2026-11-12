package gia

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"

	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManMapMiniLut2MiniAig maps MiniLUT objects to equivalent MiniAIG
// literals, using representative classes in the combined GIA.
func Gia_ManMapMiniLut2MiniAig(combined, aig, lut *Gia_Man_t, aigMap, lutMap *Vec_Int_t) []int32 {
	result := make([]int32, vec.Vec_IntSize(lutMap))
	for i := range result {
		result[i] = -1
	}
	byRepr := make([]int32, combined.nObjs)
	for i := range byRepr {
		byRepr[i] = -1
	}
	Gia_ManSetPhase(aig)
	Gia_ManSetPhase(lut)
	for i, entry := range vec.Vec_IntArray(aigMap) {
		if entry < 0 {
			continue
		}
		obj := aig.pObjs[entry>>1]
		if obj.Value == ^uint32(0) {
			continue
		}
		phase := int32(obj.fPhase) ^ (entry & 1)
		repr := Gia_ObjReprSelf(combined, int32(obj.Value>>1))
		byRepr[repr] = int32(i)*2 + phase
	}
	for i, entry := range vec.Vec_IntArray(lutMap) {
		if entry < 0 {
			continue
		}
		obj := lut.pObjs[entry>>1]
		if obj.Value == ^uint32(0) {
			continue
		}
		repr := Gia_ObjReprSelf(combined, int32(obj.Value>>1))
		lit := byRepr[repr]
		if lit >= 0 {
			result[i] = lit ^ int32(obj.fPhase) ^ (entry & 1)
		}
	}
	outputs := make([]int32, Gia_ManCoNum(aig))
	for i := range outputs {
		outputs[i] = -1
	}
	for i, entry := range vec.Vec_IntArray(aigMap) {
		if entry >= 0 {
			obj := aig.pObjs[entry>>1]
			if Gia_ObjIsCo(obj) != 0 {
				outputs[Gia_ObjCioId(obj)] = int32(i)
			}
		}
	}
	for i, entry := range vec.Vec_IntArray(lutMap) {
		if entry >= 0 {
			obj := lut.pObjs[entry>>1]
			if Gia_ObjIsCo(obj) != 0 {
				index := Gia_ObjCioId(obj)
				if int(index) >= len(outputs) || outputs[index] < 0 {
					panic("gia: unmatched MiniLUT output")
				}
				result[i] = 2 * outputs[index]
			}
		}
	}
	return result
}

// Gia_ManNameMapVerify reports mismatched representatives and polarities.
func Gia_ManNameMapVerify(combined, aig, lut *Gia_Man_t, aigMap, lutMap *Vec_Int_t, mapping []int32) {
	Gia_ManSetPhase(aig)
	Gia_ManSetPhase(lut)
	for i, entry := range mapping {
		if entry < 0 {
			continue
		}
		a, b := vec.Vec_IntEntry(aigMap, int(entry>>1)), vec.Vec_IntEntry(lutMap, i)
		if a < 0 || b < 0 {
			continue
		}
		x, y := aig.pObjs[a>>1], lut.pObjs[b>>1]
		if Gia_ObjIsCo(x) != 0 || x.Value == ^uint32(0) || y.Value == ^uint32(0) {
			continue
		}
		if Gia_ObjReprSelf(combined, int32(x.Value>>1)) != Gia_ObjReprSelf(combined, int32(y.Value>>1)) {
			fmt.Printf("Found functional mismatch for LutId %d and AigId %d.\n", i, entry>>1)
		}
		if int32(y.fPhase)^(b&1) != int32(x.fPhase)^(a&1)^(entry&1) {
			fmt.Printf("Found phase mismatch for LutId %d and AigId %d.\n", i, entry>>1)
		}
	}
}

// Abc_FrameReadMiniLutNameMapping asks the installed bridge to compute names.
func Abc_FrameReadMiniLutNameMapping(frame *abcmain.Frame) ([]int32, error) {
	if frame == nil {
		return nil, fmt.Errorf("gia: missing frame")
	}
	result, err := frame.Export("mini-lut-name-mapping")
	if err != nil {
		return nil, err
	}
	mapping, ok := result.([]int32)
	if !ok {
		return nil, fmt.Errorf("gia: invalid MiniLUT name mapping")
	}
	return mapping, nil
}
