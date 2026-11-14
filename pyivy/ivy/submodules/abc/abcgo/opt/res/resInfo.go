package res

import "github.com/glycerine/abcgo/misc/util"

// These operations follow the 32-bit unsigned-word helpers in abc.h.
func Abc_InfoClear(p []uint32, n int32) { clear(p[:n]) }
func Abc_InfoFill(p []uint32, n int32) {
	for i := int32(0); i < n; i++ {
		p[i] = ^uint32(0)
	}
}
func Abc_InfoOr(p, q []uint32, n int32) {
	for i := int32(0); i < n; i++ {
		p[i] |= q[i]
	}
}
func Abc_InfoIsOne(p []uint32, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if p[i] != ^uint32(0) {
			return 0
		}
	}
	return 1
}
func Abc_InfoIsOrOne(p, q []uint32, n int32) bool {
	for i := int32(0); i < n; i++ {
		if p[i]|q[i] != ^uint32(0) {
			return false
		}
	}
	return true
}
func Abc_InfoIsOrOne3(p, q, r []uint32, n int32) bool {
	for i := int32(0); i < n; i++ {
		if p[i]|q[i]|r[i] != ^uint32(0) {
			return false
		}
	}
	return true
}
func Abc_InfoHasBit(p []uint32, i int32) bool { return util.Abc_InfoHasBit(p, i) }
func Abc_InfoSetBit(p []uint32, i int32)      { util.Abc_InfoSetBit(p, i) }
