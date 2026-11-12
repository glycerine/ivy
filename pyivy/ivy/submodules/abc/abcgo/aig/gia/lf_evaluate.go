package gia

import "github.com/glycerine/abcgo/misc/util"

func Lf_CutRef_rec(p *Lf_Man_t, cut *Lf_Cut_t) float32 {
	area := Lf_CutArea(p, cut)
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		if id == 0 {
			break
		}
		if Lf_ObjOff(p, id) < 0 {
			continue
		}
		if Lf_ObjMapRefInc(p, id) == 0 {
			area += Lf_CutRef_rec(p, Lf_ObjCutBestNew(p, id, new(Lf_Cut_t)))
		}
	}
	return area
}
func Lf_CutDeref_rec(p *Lf_Man_t, cut *Lf_Cut_t) float32 {
	area := Lf_CutArea(p, cut)
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		if id == 0 {
			break
		}
		if Lf_ObjOff(p, id) < 0 {
			continue
		}
		if Lf_ObjMapRefDec(p, id) == 0 {
			area += Lf_CutDeref_rec(p, Lf_ObjCutBestNew(p, id, new(Lf_Cut_t)))
		}
	}
	return area
}
func Lf_CutAreaDerefed(p *Lf_Man_t, cut *Lf_Cut_t) float32 {
	area := Lf_CutRef_rec(p, cut)
	Lf_CutDeref_rec(p, cut)
	return area
}
func Lf_CutRequired(p *Lf_Man_t, cut *Lf_Cut_t) int32 {
	var arrival, required int32
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		var a int32
		if Lf_ObjOff(p, id) < 0 {
			a = Lf_ObjArrival_rec(p, Gia_ManObj(p.pGia, id))
		} else {
			a = Lf_ObjReadBest(p, id).Delay[0]
		}
		arrival = max(arrival, a)
		r := Lf_ObjRequired(p, id)
		if r < util.ABC_INFINITY {
			required = max(required, r)
		}
	}
	return max(required+2, arrival+1)
}
func Lf_CutParams(p *Lf_Man_t, cut *Lf_Cut_t, required int32, refs float32, mux *Gia_Obj_t) {
	if cut.fMux7 != 0 && Gia_ObjIsMux(p.pGia, mux) == 0 {
		panic("LF expected MUX")
	}
	cut.fLate = 0
	cut.Delay = 0
	cut.Flow = 0
	if cut.nLeaves >= LF_NO_LEAF {
		panic("LF invalid cut")
	}
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		var delay int32
		if Lf_ObjOff(p, id) < 0 {
			delay = Lf_ObjArrival_rec(p, Gia_ManObj(p.pGia, id))
		} else {
			best := Lf_ObjReadBest(p, id)
			if best.Delay[0] > best.Delay[1] || best.Flow[0] < best.Flow[1] {
				panic("LF invalid best cut ordering")
			}
			var index int32
			if p.fUseEla != 0 {
				index = Lf_BestIndex(best)
			} else {
				if best.Delay[1]+1 <= required && required != util.ABC_INFINITY {
					index = 1
				}
				if cut.Flow >= 1e32 || best.Flow[index] >= 1e32 {
					cut.Flow = 1e32
				} else {
					cut.Flow = min(cut.Flow+best.Flow[index], 1e32)
				}
			}
			delay = best.Delay[index]
		}
		cut.Delay = max(cut.Delay, delay)
	}
	if cut.nLeaves > 1 {
		cut.Delay++
	}
	if cut.Delay > required {
		cut.fLate = 1
	}
	if p.fUseEla != 0 {
		cut.Flow = Lf_CutAreaDerefed(p, cut) / refs
	} else {
		cut.Flow = min((cut.Flow+Lf_CutArea(p, cut))/refs, 1e32)
	}
}
