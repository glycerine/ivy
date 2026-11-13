package abs

import "github.com/glycerine/abcgo/misc/vec"

// Gia_VtaAbsToFrames unpacks the frame offsets and object IDs used by ABC's
// variable timeframe abstraction. Each returned frame owns its entries.
func Gia_VtaAbsToFrames(abs *vec.Vec_Int_t) *vec.Vec_Ptr_t {
	if abs == nil || vec.Vec_IntSize(abs) < 2 {
		panic("invalid variable timeframe abstraction")
	}
	data := vec.Vec_IntArray(abs)
	n := int(data[0])
	if n < 0 || n > len(data)-2 || int(data[1]) != n+2 || int(data[n+1]) != len(data) {
		panic("invalid variable timeframe abstraction offsets")
	}
	frames := vec.Vec_PtrAlloc(n)
	for i := 0; i < n; i++ {
		start, end := int(data[i+1]), int(data[i+2])
		if start < n+2 || end < start || end > len(data) {
			panic("invalid variable timeframe abstraction offsets")
		}
		vec.Vec_PtrPush(frames, vec.Vec_IntAllocArrayCopy(data[start:end], end-start))
	}
	return frames
}

// Gia_VtaFramesToAbs packs per-frame integer vectors using ABC's layout:
// frame count, frame offsets including an end sentinel, then object IDs.
func Gia_VtaFramesToAbs(frames *vec.Vec_Vec_t) *vec.Vec_Int_t {
	if frames == nil {
		panic("nil variable timeframe frames")
	}
	n := vec.Vec_VecSize(frames)
	data := make([]int32, 0, n+2+vec.Vec_VecSizeSize(frames))
	data = append(data, int32(n))
	end := n + 2
	for i := 0; i < n; i++ {
		frame := vec.Vec_VecEntryInt(frames, i)
		if frame == nil {
			panic("nil variable timeframe frame")
		}
		data = append(data, int32(end))
		end += vec.Vec_IntSize(frame)
	}
	data = append(data, int32(end))
	for i := 0; i < n; i++ {
		data = append(data, vec.Vec_IntArray(vec.Vec_VecEntryInt(frames, i))...)
	}
	return vec.Vec_IntAllocArray(data, len(data))
}

// Vta_ManDeriveAbsAll marks each object that appears in any frame word.
func Vta_ManDeriveAbsAll(words *vec.Vec_Int_t, nWords int32) *vec.Vec_Int_t {
	if words == nil || nWords <= 0 || vec.Vec_IntSize(words)%int(nWords) != 0 {
		panic("invalid variable timeframe word vector")
	}
	data := vec.Vec_IntArray(words)
	result := vec.Vec_IntAlloc(len(data) / int(nWords))
	for start := 0; start < len(data); start += int(nWords) {
		var seen int32
		for _, word := range data[start : start+int(nWords)] {
			if word != 0 {
				seen = 1
				break
			}
		}
		vec.Vec_IntPush(result, seen)
	}
	return result
}

// Vec_IntDoubleWidth doubles each object's word span, preserving its old
// low words and zero-initializing the new high words.
func Vec_IntDoubleWidth(words *vec.Vec_Int_t, nWords int32) int32 {
	if words == nil || nWords <= 0 || vec.Vec_IntSize(words)%int(nWords) != 0 {
		panic("invalid variable timeframe word vector")
	}
	data := vec.Vec_IntArray(words)
	newData := make([]int32, len(data)*2)
	width := int(nWords)
	for object := 0; object < len(data)/width; object++ {
		copy(newData[2*width*object:], data[width*object:width*(object+1)])
	}
	*vec.Vec_IntArrayP(words) = newData
	return 2 * nWords
}
