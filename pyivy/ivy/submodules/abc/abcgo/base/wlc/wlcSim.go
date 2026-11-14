package wlc

import "fmt"

func Wlc_NtkDeleteSim(result *Vec_Ptr_t) {
	if result == nil {
		return
	}
	for i := int32(0); i < Vec_PtrSize(result); i++ {
		inner, ok := Vec_PtrEntry(result, i).(*Vec_Ptr_t)
		if !ok || inner == nil {
			continue
		}
		Vec_PtrFree(inner)
	}
	Vec_PtrFree(result)
}

func Wlc_NtkSimulate(p *Wlc_Ntk_t, nodes *Vec_Int_t, nWords, nFrames int32) *Vec_Ptr_t {
	gia := Wlc_NtkBitBlast(p, nil)
	if gia == nil {
		return nil
	}
	if nodes == nil || nWords <= 0 || nFrames <= 0 {
		Gia_ManStop(gia)
		return nil
	}
	Gia_ManClearSimData(gia)
	Gia_ManSetSimData(gia, Vec_WrdStart(Gia_ManObjNum(gia)*nWords), nWords)
	result := Vec_PtrAlloc(Vec_IntSize(nodes))
	for i := int32(0); i < Vec_IntSize(nodes); i++ {
		obj := Wlc_NtkObj(p, Vec_IntEntry(nodes, i))
		inner := Vec_PtrAlloc(Wlc_ObjRange(obj))
		for bit := int32(0); bit < Wlc_ObjRange(obj); bit++ {
			Vec_PtrPush(inner, make([]word, int(nWords*nFrames)))
		}
		Vec_PtrPush(result, inner)
	}
	Gia_ManRandomW(1)
	var dangling int32
	for frame := int32(0); frame < nFrames; frame++ {
		for id := int32(0); id < Gia_ManObjNum(gia); id++ {
			obj := Gia_ManObj(gia, id)
			if Gia_ObjIsAnd(obj) != 0 {
				Wlc_ObjSimAnd(gia, id)
			} else if Gia_ObjIsCo(obj) != 0 {
				Wlc_ObjSimCo(gia, id)
			} else if Gia_ObjIsPi(gia, obj) != 0 {
				wlcNtkSimPi(gia, id)
			} else if Gia_ObjIsRo(gia, obj) != 0 {
				Wlc_ObjSimRo(gia, id)
			}
		}
		for i := int32(0); i < Vec_IntSize(nodes); i++ {
			obj := Wlc_NtkObj(p, Vec_IntEntry(nodes, i))
			first := Vec_IntEntry(&p.vCopies, Wlc_ObjId(p, obj))
			inner := Vec_PtrEntry(result, i).(*Vec_Ptr_t)
			for bit := int32(0); bit < Wlc_ObjRange(obj); bit++ {
				lit := Vec_IntEntry(&p.vBits, first+bit)
				data := Vec_PtrEntry(inner, bit).([]word)
				if lit == -1 {
					dangling++
					clear(data[int(frame*nWords):int((frame+1)*nWords)])
					continue
				}
				packed := Wlc_ObjSim(gia, Abc_Lit2Var(lit))
				for w := int32(0); w < nWords; w++ {
					value := packed[w]
					if Abc_LitIsCompl(lit) != 0 {
						value = ^value
					}
					data[int(frame*nWords+w)] = value
				}
			}
		}
		if frame == 0 && dangling > 0 {
			fmt.Printf("Replaced %d dangling internal bits with constant 0.\n", dangling)
		}
	}
	Gia_ManClearSimData(gia)
	Gia_ManStop(gia)
	return result
}

// wlcNtkSimPi mirrors wlcSim.c. The Wlc_ObjSimPi helper used by graft code
// intentionally shifts its first word, while this simulator does not.
func wlcNtkSimPi(gia *Gia_Man_t, id int32) {
	words := Gia_ManSimObjectData(gia, id)
	for i := range words {
		words[i] = Gia_ManRandomW(0)
	}
}

func Wlc_NtkSimulatePrint(p *Wlc_Ntk_t, nodes *Vec_Int_t, result *Vec_Ptr_t, nWords, nFrames int32) {
	if result == nil {
		return
	}
	for frame := int32(0); frame < nFrames; frame++ {
		for w := int32(0); w < nWords; w++ {
			for bit := int32(0); bit < 64; bit++ {
				pattern := frame*nWords*64 + w*64 + bit
				for i := int32(0); i < Vec_IntSize(nodes); i++ {
					obj := Wlc_NtkObj(p, Vec_IntEntry(nodes, i))
					inner := Vec_PtrEntry(result, i).(*Vec_Ptr_t)
					for k := Wlc_ObjRange(obj) - 1; k >= 0; k-- {
						data := Vec_PtrEntry(inner, k).([]word)
						wordIndex, bitIndex := pattern/64, uint(pattern%64)
						if int(wordIndex) >= len(data) || data[wordIndex]&(word(1)<<bitIndex) == 0 {
							fmt.Print("0")
						} else {
							fmt.Print("1")
						}
					}
					fmt.Print(" ")
				}
				fmt.Println()
			}
		}
	}
}

func Wlc_NtkSimulateTest(p *Wlc_Ntk_t) {
	nodes := Vec_IntAlloc(3)
	Vec_IntPush(nodes, 1)
	Vec_IntPush(nodes, 2)
	Vec_IntPush(nodes, 3)
	result := Wlc_NtkSimulate(p, nodes, 2, 2)
	Wlc_NtkSimulatePrint(p, nodes, result, 2, 2)
	Wlc_NtkDeleteSim(result)
	Vec_IntFree(nodes)
}
