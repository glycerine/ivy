package acec

import "github.com/glycerine/abcgo/aig/gia"

// Acec_Normalize derives an adder box and replaces it with an equivalent tree.
// Ported from src/proof/acec/acecNorm.c; insertion helpers are in acec_adders.go.
func Acec_Normalize(pGia *Gia_Man_t, fBooth, fVerbose int32) *Gia_Man_t {
	var ignore *Vec_Bit_t
	if fBooth != 0 {
		ignore = Acec_BoothFindPPG(pGia)
	}
	box := Acec_DeriveBox(pGia, ignore, 0, 0, fVerbose)
	if box == nil {
		return gia.Gia_ManDup(pGia)
	}
	return Acec_InsertBox(box, 1)
}
