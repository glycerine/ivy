package cut

type Cut_List_t struct {
	head [CUT_SIZE_MAX + 1]*Cut_Cut_t
	tail [CUT_SIZE_MAX + 1]**Cut_Cut_t
}

func Cut_ListStart(p *Cut_List_t) {
	*p = Cut_List_t{}
	for i := int32(1); i <= CUT_SIZE_MAX; i++ {
		p.tail[i] = &p.head[i]
	}
}
func Cut_ListAdd(p *Cut_List_t, c *Cut_Cut_t) {
	n := len(c.leaves)
	if n < 1 || n > int(CUT_SIZE_MAX) {
		panic("cut: invalid list size")
	}
	c.Next = nil
	*p.tail[n] = c
	p.tail[n] = &c.Next
}
func Cut_ListAdd2(p *Cut_List_t, c *Cut_Cut_t) {
	n := len(c.leaves)
	if n < 1 || n > int(CUT_SIZE_MAX) {
		panic("cut: invalid list size")
	}
	spot := &p.head[n]
	for *spot != nil && Cut_CutCompare(c, *spot) >= 0 {
		spot = &(*spot).Next
	}
	c.Next = *spot
	*spot = c
	if c.Next == nil {
		p.tail[n] = &c.Next
	}
}
func Cut_ListDerive(p *Cut_List_t, list *Cut_Cut_t) {
	Cut_ListStart(p)
	for list != nil {
		next := list.Next
		Cut_ListAdd(p, list)
		list = next
	}
}
func Cut_ListAddList(dst, src *Cut_List_t) {
	for i := int32(1); i <= CUT_SIZE_MAX; i++ {
		if src.head[i] != nil {
			*dst.tail[i] = src.head[i]
			dst.tail[i] = src.tail[i]
		}
	}
}
func Cut_ListFinish(p *Cut_List_t) *Cut_Cut_t {
	var head *Cut_Cut_t
	tail := &head
	for i := int32(1); i <= CUT_SIZE_MAX; i++ {
		if p.head[i] != nil {
			*tail = p.head[i]
			tail = p.tail[i]
		}
	}
	*tail = nil
	return head
}
