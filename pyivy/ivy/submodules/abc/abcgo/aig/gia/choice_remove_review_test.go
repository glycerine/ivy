package gia

import "testing"

func TestRemoveChoicesUsedAsMuxControls(t *testing.T) {
	for name, remove := range map[string]func(*Gia_Man_t){"gia": Gia_ManRemoveBadChoices, "hcd": Hcd_ManRemoveBadChoices} {
		t.Run(name, func(t *testing.T) {
			p := Gia_ManStart(1)
			a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			ab := Gia_ManAppendAnd(p, a, b)
			head := Gia_ManAppendAnd(p, ab, c)
			ac := Gia_ManAppendAnd(p, a, c)
			used := Gia_ManAppendAnd(p, ac, b)
			bc := Gia_ManAppendAnd(p, b, c)
			unused := Gia_ManAppendAnd(p, bc, a)
			p.EnableMuxes()
			Gia_ManAppendCo(p, Gia_ManAppendMuxReal(p, used^1, c, d))
			p.StartRepresentatives()
			p.PNexts = make([]int32, p.nObjs)
			Gia_ObjSetRepr(p, used>>1, head>>1)
			Gia_ObjSetRepr(p, unused>>1, head>>1)
			Gia_ObjSetNext(p, head>>1, used>>1)
			Gia_ObjSetNext(p, used>>1, unused>>1)
			remove(p)
			if Gia_ObjRepr(p, used>>1) != GIA_VOID || Gia_ObjNext(p, used>>1) != 0 {
				t.Fatal("alternative with a MUX control fanout was kept as a choice")
			}
			if Gia_ObjNext(p, head>>1) != unused>>1 || Gia_ObjRepr(p, unused>>1) != head>>1 {
				t.Fatal("unused equivalent alternative was removed")
			}
			for _, obj := range p.pObjs {
				if obj.fMark0 != 0 {
					t.Fatal("bad-choice removal retained scratch marks")
				}
			}
		})
	}
}
