package acec

// Internal types and adder-sign accessors ported from acecInt.h.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Wec_t = vec.Vec_Wec_t
type Vec_Bit_t = vec.Vec_Bit_t

type Acec_Box_t struct {
	pGia      *Gia_Man_t
	vAdds     *Vec_Wec_t
	vLeafLits *Vec_Wec_t
	vRootLits *Vec_Wec_t
	vShared   *Vec_Wec_t
	vUnique   *Vec_Wec_t
}

func Acec_SignBit(adders *Vec_Int_t, box, bit int32) int32 {
	return (vec.Vec_IntEntry(adders, int(6*box+5)) >> bit) & 1
}
func Acec_SignBit2(adders *Vec_Int_t, box, bit int32) int32 {
	return (vec.Vec_IntEntry(adders, int(6*box+5)) >> (16 + bit)) & 1
}
func Acec_SignSetBit(adders *Vec_Int_t, box, bit, value int32) {
	if value != 0 {
		index := int(6*box + 5)
		vec.Vec_IntWriteEntry(adders, index, vec.Vec_IntEntry(adders, index)|(1<<bit))
	}
}
func Acec_SignSetBit2(adders *Vec_Int_t, box, bit, value int32) {
	if value != 0 {
		index := int(6*box + 5)
		vec.Vec_IntWriteEntry(adders, index, vec.Vec_IntEntry(adders, index)|(1<<(16+bit)))
	}
}
