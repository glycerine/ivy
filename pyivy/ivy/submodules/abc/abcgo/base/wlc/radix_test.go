package wlc

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestConvertToRadixAgainstBigInt(t *testing.T) {
	rng := rand.New(rand.NewSource(104))
	for _, length := range []int32{1, 7, 32, 65, 125} {
		for _, offset := range []int32{0, 3, 31} {
			for base := int32(2); base < 36; base++ {
				words := make([]uint32, (int(offset+length)+31)/32)
				value := new(big.Int)
				for i := int32(0); i < length; i++ {
					if rng.Intn(2) != 0 {
						words[(offset+i)/32] |= 1 << uint((offset+i)%32)
						value.SetBit(value, int(i), 1)
					}
				}
				digits := vec.Vec_StrArray(Wlc_ConvertToRadix(words, offset, length, base))
				if len(digits) == 0 || digits[len(digits)-1] != 0 || string(digits[:len(digits)-1]) != value.Text(int(base)) {
					t.Fatalf("bits=%d offset=%d base=%d got %q want %q", length, offset, base, digits, value.Text(int(base)))
				}
			}
		}
	}
}
func TestComputeSumAllowsAliasedOperands(t *testing.T) {
	digits := []byte{7, 4, 0}
	Wlc_ComputeSum(digits, digits, 3, 10)
	if string(digits) != string([]byte{4, 9, 0}) {
		t.Fatalf("got %v", digits)
	}
}
