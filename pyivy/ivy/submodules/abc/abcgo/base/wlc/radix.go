package wlc

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Wlc_ComputeSum adds little-endian base-radix digit arrays in place.
func Wlc_ComputeSum(result, add []byte, nBits, radix int32) {
	carry := int32(0)
	for i := int32(0); i < nBits; i++ {
		sum := int32(result[i]) + int32(add[i]) + carry
		if sum >= radix {
			sum -= radix
			carry = 1
		} else {
			carry = 0
		}
		if sum < 0 || sum >= radix {
			panic("invalid radix digit")
		}
		result[i] = byte(sum)
	}
	if carry != 0 {
		panic("radix sum overflow")
	}
}

// Wlc_ConvertToRadix converts the selected little-endian bit range to a
// NUL-terminated lowercase digit vector, as the C reporting code expects.
func Wlc_ConvertToRadix(bits []uint32, start, nBits, radix int32) *vec.Vec_Str_t {
	if radix < 2 || radix >= 36 {
		panic("invalid radix")
	}
	result := vec.Vec_StrStart(int(nBits))
	sum := vec.Vec_StrStart(int(nBits))
	digits, increment := vec.Vec_StrArray(result), vec.Vec_StrArray(sum)
	increment[0] = 1
	for i := int32(0); i < nBits; i++ {
		if util.Abc_InfoHasBit(bits, start+i) {
			Wlc_ComputeSum(digits, increment, nBits, radix)
		}
		if i < nBits-1 {
			Wlc_ComputeSum(increment, increment, nBits, radix)
		}
	}
	i := int(nBits) - 1
	for i >= 0 && digits[i] == 0 {
		i--
	}
	vec.Vec_StrShrink(result, i+1)
	for ; i >= 0; i-- {
		if digits[i] < 10 {
			digits[i] += '0'
		} else {
			digits[i] += 'a' - 10
		}
	}
	vec.Vec_StrReverseOrder(result)
	if vec.Vec_StrSize(result) == 0 {
		vec.Vec_StrPush(result, '0')
	}
	vec.Vec_StrPush(result, 0)
	return result
}
