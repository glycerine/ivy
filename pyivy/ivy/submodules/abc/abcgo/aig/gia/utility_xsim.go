package gia

import "fmt"

// Gia_ManPoXSim returns the first frame in which each primary output becomes
// unknown, or frames if it remains known throughout the simulation.
func Gia_ManPoXSim(p *Gia_Man_t, frames, verbose int32) *Vec_Int_t {
	if frames < 0 {
		panic("gia: negative simulation frame count")
	}
	result := Vec_IntAlloc(int(Gia_ManPoNum(p)))
	Vec_IntFill(result, int(Gia_ManPoNum(p)), frames)
	left := Gia_ManPoNum(p)
	Gia_ObjTerSimSet0(Gia_ManConst0(p))
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		Gia_ObjTerSimSet0(Gia_ManRi(p, i))
	}
	var frame int32
	for frame = 0; frame < frames; frame++ {
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			Gia_ObjTerSimSetX(Gia_ManPi(p, i))
		}
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			Gia_ObjTerSimRo(p, Gia_ManRo(p, i))
		}
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				Gia_ObjTerSimAnd(obj)
			}
		}
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			Gia_ObjTerSimCo(Gia_ManCo(p, i))
		}
		if verbose != 0 {
			for i := int32(0); i < Gia_ManPoNum(p); i++ {
				Gia_ObjTerSimPrint(Gia_ManPo(p, i))
			}
			fmt.Println()
		}
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			if Vec_IntEntry(result, int(i)) == frames && Gia_ObjTerSimGetX(Gia_ManPo(p, i)) != 0 {
				Vec_IntWriteEntry(result, int(i), frame)
				left--
			}
		}
		if left == 0 {
			break
		}
	}
	if verbose != 0 {
		if left == 0 {
			fmt.Printf("Simulation converged after %d frames.\n", frame+1)
		} else {
			fmt.Printf("Simulation terminated after %d frames.\n", frames)
		}
	}
	return result
}
