package gia

// ofCut is the unpacked cut used during LUT-structure enumeration.
type ofCut struct {
	sign           uint64
	delay, flow    int32
	function, size uint32
	leaves         [7]int32
}

func ofUnitCut(id int32) ofCut {
	c := ofCut{sign: uint64(1) << uint(id&63), function: 2, size: 1}
	c.leaves[0] = id
	return c
}

// ofCutStorage preserves the source's page/offset handles and reserves three
// words after each cut for delay and area-flow annotations.
type ofCutStorage struct {
	pages [][]int32
	next  int32
}

func newOfCutStorage() *ofCutStorage { return &ofCutStorage{next: 2} }
func (s *ofCutStorage) at(handle int32) []int32 {
	if handle <= 0 || int(uint32(handle)>>16) >= len(s.pages) {
		panic("gia: invalid Of cut handle")
	}
	return s.pages[uint32(handle)>>16][uint32(handle)&0xffff:]
}
func (s *ofCutStorage) save(cuts []*ofCut) int32 {
	size := 1
	for _, c := range cuts {
		if c.size > 6 || c.function > 0x7ffffff {
			panic("gia: invalid Of cut")
		}
		size += int(c.size) + 4
	}
	if size > 0xffff {
		panic("gia: Of cut set exceeds page")
	}
	if int(s.next&0xffff)+size > 0xffff {
		s.next = ((s.next >> 16) + 1) << 16
	}
	page := int(uint32(s.next) >> 16)
	for len(s.pages) <= page {
		s.pages = append(s.pages, make([]int32, 1<<16))
	}
	handle := s.next
	s.next += int32(size)
	dst := s.at(handle)
	dst[0] = int32(len(cuts))
	pos := 1
	for _, c := range cuts {
		dst[pos] = int32(c.size) | (int32(c.function) << 5)
		pos++
		pos += copy(dst[pos:], c.leaves[:c.size])
		dst[pos], dst[pos+1], dst[pos+2] = -1, -1, -1
		pos += 3
	}
	return handle
}
func ofStoredSize(c []int32) int32         { return c[0] & 31 }
func ofStoredFunction(c []int32) int32     { return int32(uint32(c[0]) >> 5) }
func ofStoredLeaves(c []int32) []int32     { return c[1 : 1+ofStoredSize(c)] }
func ofStoredDelay1(c []int32) int32       { return c[1+ofStoredSize(c)] }
func ofStoredDelay2(c []int32) int32       { return c[2+ofStoredSize(c)] }
func ofStoredFlow(c []int32) int32         { return c[3+ofStoredSize(c)] }
func ofStoredSetDelay1(c []int32, v int32) { c[1+ofStoredSize(c)] = v }
func ofStoredSetDelay2(c []int32, v int32) { c[2+ofStoredSize(c)] = v }
func ofStoredSetFlow(c []int32, v int32)   { c[3+ofStoredSize(c)] = v }
func ofStoredVar(c []int32, i int) int32   { return ofStoredLeaves(c)[i] >> 1 }
func ofStoredFlag(c []int32, i int) int32  { return ofStoredLeaves(c)[i] & 1 }
func ofStoredSetFlag(c []int32, i int, v int32) {
	leaves := ofStoredLeaves(c)
	leaves[i] = (leaves[i] &^ 1) | (v & 1)
}

func (s *ofCutStorage) lift(handle int32) {
	data := s.at(handle)
	pos := 1
	for i := int32(0); i < data[0]; i++ {
		c := data[pos:]
		n := ofStoredSize(c)
		for j := int32(0); j < n; j++ {
			c[1+j] *= 2
		}
		pos += int(n) + 4
	}
}
