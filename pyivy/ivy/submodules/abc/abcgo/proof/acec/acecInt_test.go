package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestAdderSigns(t *testing.T) {
	adds := vec.Vec_IntStart(12)
	for box := int32(0); box < 2; box++ {
		for bit := int32(0); bit < 16; bit++ {
			if Acec_SignBit(adds, box, bit) != 0 || Acec_SignBit2(adds, box, bit) != 0 {
				t.Fatal("unexpected initial sign")
			}
			Acec_SignSetBit(adds, box, bit, 0)
			Acec_SignSetBit2(adds, box, bit, 0)
			if Acec_SignBit(adds, box, bit) != 0 || Acec_SignBit2(adds, box, bit) != 0 {
				t.Fatal("zero set modified sign")
			}
			Acec_SignSetBit(adds, box, bit, 1)
			Acec_SignSetBit2(adds, box, bit, 1)
			if Acec_SignBit(adds, box, bit) != 1 || Acec_SignBit2(adds, box, bit) != 1 {
				t.Fatal("set sign missing", box, bit)
			}
		}
	}
}
