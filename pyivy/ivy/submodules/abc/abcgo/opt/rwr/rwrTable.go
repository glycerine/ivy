package rwr

// The built-in library has chains for only 134 of the 65,536 truth tables.
// Allocate a page on its first insertion and keep chain heads addressable.
const rwrPageSize = 64

type rwrTruthTable [65536 / rwrPageSize]*[rwrPageSize]*Rwr_Node_t

func (t *rwrTruthTable) head(index int32) *Rwr_Node_t {
	key := uint32(index)
	page := t[key/rwrPageSize]
	if page == nil {
		return nil
	}
	return page[key%rwrPageSize]
}

func (t *rwrTruthTable) slot(truth uint16) **Rwr_Node_t {
	page := t[truth/rwrPageSize]
	if page == nil {
		page = new([rwrPageSize]*Rwr_Node_t)
		t[truth/rwrPageSize] = page
	}
	return &page[truth%rwrPageSize]
}
