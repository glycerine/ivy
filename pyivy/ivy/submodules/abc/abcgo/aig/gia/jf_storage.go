package gia

// jfCutMemory owns one mutable packed cut list per handle. A separate slice
// per list preserves references to previously allocated lists as it grows.
type jfCutMemory struct {
	lists [][]int32
	words int
}

func (m *jfCutMemory) append(data []int32, size int) int32 {
	if size < 0 || size > len(data) {
		panic("JF invalid cut list size")
	}
	id := int32(len(m.lists))
	m.lists = append(m.lists, append([]int32(nil), data[:size]...))
	m.words += size
	return id
}
func (m *jfCutMemory) get(handle int32) []int32 { return m.lists[handle] }
func (m *jfCutMemory) reset()                   { *m = jfCutMemory{} }
func (m *jfCutMemory) memory() float64          { return float64(m.words * 4) }

// Jf_ObjSetBestCut moves one cut to the front while retaining the relative
// order of all other cuts in the packed list.
func Jf_ObjSetBestCut(cuts, selected []int32, _ *Vec_Int_t) {
	if len(cuts) < 2 || len(selected) == 0 {
		panic("JF invalid best cut")
	}
	offset := 1
	for i := int32(0); i < cuts[0]; i++ {
		if offset >= len(cuts) {
			panic("JF malformed cut list")
		}
		size := int(Jf_CutSize(cuts[offset:])) + 1
		if offset+size > len(cuts) {
			panic("JF malformed cut")
		}
		if &cuts[offset] == &selected[0] {
			if offset == 1 {
				return
			}
			current := append([]int32(nil), cuts[offset:offset+size]...)
			copy(cuts[1+size:offset+size], cuts[1:offset])
			copy(cuts[1:], current)
			return
		}
		offset += size
	}
	panic("JF best cut is not in list")
}
