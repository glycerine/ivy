package gia

// Tas_Sto_t stores clauses by integer offsets, so growth does not invalidate
// watch-list handles. The first sixteen words are reserved as in ABC.
type Tas_Sto_t struct {
	words []int32
}

type Tas_Cls_t struct {
	store  *Tas_Sto_t
	handle int32
}

func (s *Tas_Sto_t) reset() {
	if cap(s.words) < 16 {
		s.words = make([]int32, 16)
	} else {
		s.words = s.words[:16]
		clear(s.words)
	}
}

func (s *Tas_Sto_t) allocate(literals int) *Tas_Cls_t {
	if literals < 0 {
		panic("gia: negative clause size")
	}
	if len(s.words) == 0 {
		s.reset()
	}
	if int64(len(s.words))+int64(literals)+3 > int64(^uint32(0)>>1) {
		panic("gia: clause store overflow")
	}
	handle := int32(len(s.words))
	s.words = append(s.words, make([]int32, literals+3)...)
	s.words[handle+2] = int32(literals)
	return &Tas_Cls_t{store: s, handle: handle}
}

func (s *Tas_Sto_t) clause(handle int32) *Tas_Cls_t {
	if handle < 16 || int(handle)+3 > len(s.words) {
		panic("gia: invalid clause handle")
	}
	size := s.words[handle+2]
	if size < 0 || int64(handle)+3+int64(size) > int64(len(s.words)) {
		panic("gia: invalid clause size")
	}
	return &Tas_Cls_t{store: s, handle: handle}
}

func (c *Tas_Cls_t) literals() []int32 {
	start := c.handle + 3
	return c.store.words[start : start+c.store.words[c.handle+2]]
}
func (c *Tas_Cls_t) next(index int) int32 {
	if index < 0 || index > 1 {
		panic("gia: invalid watch index")
	}
	return c.store.words[int(c.handle)+index]
}
func (c *Tas_Cls_t) setNext(index int, next int32) {
	if index < 0 || index > 1 {
		panic("gia: invalid watch index")
	}
	c.store.words[int(c.handle)+index] = next
}
