package gia

// nfCutStorage uses the original 16-bit page/offset handles, while retaining
// Go-owned pages. Handle zero remains reserved for objects without cuts.
type nfCutStorage struct {
	pages [][]int32
	next  int32
}

func newNfCutStorage() *nfCutStorage { return &nfCutStorage{next: 2} }
func (s *nfCutStorage) at(handle int32) []int32 {
	if handle <= 0 || int(uint32(handle)>>16) >= len(s.pages) {
		panic("gia: invalid Nf cut handle")
	}
	return s.pages[uint32(handle)>>16][uint32(handle)&0xffff:]
}
func (s *nfCutStorage) save(cuts []*Nf_Cut_t, usefulOnly bool) int32 {
	size, count := 1, 0
	for _, c := range cuts {
		if c.nLeaves > NF_LEAF_MAX {
			panic("gia: invalid Nf cut")
		}
		if !usefulOnly || c.Useless == 0 {
			size += int(c.nLeaves) + 1
			count++
		}
	}
	if size > 0xffff {
		panic("gia: Nf cut set exceeds page")
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
	dst[0] = int32(count)
	pos := 1
	for _, c := range cuts {
		if usefulOnly && c.Useless != 0 {
			continue
		}
		dst[pos] = Nf_CutSetBoth(int32(c.nLeaves), int32(c.iFunc))
		pos++
		pos += copy(dst[pos:], c.pLeaves[:c.nLeaves])
	}
	return handle
}
func Nf_CutSize(c []int32) int32               { return c[0] & NF_NO_LEAF }
func Nf_CutFunc(c []int32) int32               { return int32(uint32(c[0]) >> 5) }
func Nf_CutLeaves(c []int32) []int32           { return c[1 : 1+Nf_CutSize(c)] }
func Nf_CutSetBoth(size, function int32) int32 { return size | (function << 5) }
func Nf_CutIsTriv(c []int32, id int32) int32   { return giaBool(Nf_CutSize(c) == 1 && c[1] == id) }

func Nf_SetLastCutIsContained(cuts []*Nf_Cut_t, n int32) int32 {
	candidate := cuts[n]
	for _, c := range cuts[:n] {
		if c.nLeaves <= candidate.nLeaves && c.Sign&candidate.Sign == c.Sign && Nf_SetCutIsContainedOrder(candidate, c) != 0 {
			return 1
		}
	}
	return 0
}
func Nf_SetLastCutContainsArea(cuts []*Nf_Cut_t, n int32) int32 {
	candidate := cuts[n]
	write := 0
	for read := 0; read < int(n); read++ {
		c := cuts[read]
		if candidate.nLeaves < c.nLeaves && candidate.Sign&c.Sign == candidate.Sign && Nf_SetCutIsContainedOrder(c, candidate) != 0 {
			c.nLeaves = NF_NO_LEAF
			continue
		}
		cuts[write], cuts[read] = cuts[read], cuts[write]
		write++
	}
	cuts[write], cuts[n] = cuts[n], cuts[write]
	return int32(write)
}
func Nf_SetSortByArea(cuts []*Nf_Cut_t, n int32) {
	for i := n; i > 0; i-- {
		if Nf_CutCompareArea(cuts[i-1], cuts[i]) < 0 {
			return
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
}
func Nf_SetAddCut(cuts []*Nf_Cut_t, n, limit int32) int32 {
	if n == 0 {
		return 1
	}
	n = Nf_SetLastCutContainsArea(cuts, n)
	Nf_SetSortByArea(cuts, n)
	return min(n+1, limit-1)
}
