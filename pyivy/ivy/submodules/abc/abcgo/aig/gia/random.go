package gia

import "sync"

const NUMBER1 uint32 = 3716960521
const NUMBER2 uint32 = 2174103536

var giaRandom = struct {
	sync.Mutex
	z, w uint32
}{z: NUMBER1, w: NUMBER2}

func giaRandomStep(reset int32) uint32 {
	if reset != 0 {
		giaRandom.z = NUMBER1
		giaRandom.w = NUMBER2
	}
	giaRandom.z = 36969*(giaRandom.z&65535) + (giaRandom.z >> 16)
	giaRandom.w = 18000*(giaRandom.w&65535) + (giaRandom.w >> 16)
	return giaRandom.z<<16 + giaRandom.w
}
func Gia_ManRandom(reset int32) uint32 {
	giaRandom.Lock()
	defer giaRandom.Unlock()
	return giaRandomStep(reset)
}
func Gia_ManRandomW(reset int32) uint64 {
	giaRandom.Lock()
	defer giaRandom.Unlock()
	high := giaRandomStep(reset)
	low := giaRandomStep(reset)
	return uint64(high)<<32 | uint64(low)
}
func Gia_ManRandomInfo(info *Vec_Ptr_t, inputStart, wordStart, wordStop int32) {
	if inputStart < 0 || int(inputStart) > Vec_PtrSize(info) || wordStart < 0 || wordStop < wordStart {
		panic("gia: invalid random information range")
	}
	giaRandom.Lock()
	defer giaRandom.Unlock()
	for i := int(inputStart); i < Vec_PtrSize(info); i++ {
		words, ok := Vec_PtrEntry(info, i).([]uint32)
		if !ok || int(wordStop) > len(words) {
			panic("gia: invalid random information buffer")
		}
		for w := wordStart; w < wordStop; w++ {
			words[w] = giaRandomStep(0)
		}
	}
}
