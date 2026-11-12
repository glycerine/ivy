// Placement database types, ported from Aaron P. Hurst's place_base.h.
package place

type Rect struct{ x, y, w, h float32 }
type AbstractCell struct {
	m_label           string
	m_width, m_height float32
	m_pad             bool
}
type ConcreteCell struct {
	m_id     int32
	m_label  string
	m_parent *AbstractCell
	m_fixed  bool
	m_x, m_y float32
	m_data   int32
}
type ConcreteNet struct {
	m_id       int32
	m_numTerms int32
	m_terms    []*ConcreteCell
	m_weight   float32
	m_data     int32
}

func NewCellType(name string, width, height float32, pad bool) *AbstractCell {
	if width < 0 || height < 0 {
		panic("negative cell dimensions")
	}
	return &AbstractCell{m_label: name, m_width: width, m_height: height, m_pad: pad}
}
func NewCell(id int32, name string, kind *AbstractCell, x, y float32) *ConcreteCell {
	if id < 0 || kind == nil {
		panic("invalid placement cell")
	}
	return &ConcreteCell{m_id: id, m_label: name, m_parent: kind, m_x: x, m_y: y}
}
func NewNet(id int32, terms []*ConcreteCell, weight float32) *ConcreteNet {
	if id < 0 {
		panic("negative net id")
	}
	for _, cell := range terms {
		if cell == nil {
			panic("nil net terminal")
		}
	}
	return &ConcreteNet{m_id: id, m_terms: append([]*ConcreteCell(nil), terms...), m_numTerms: int32(len(terms)), m_weight: weight}
}
func (c *ConcreteCell) Position() (float32, float32)      { return c.m_x, c.m_y }
func (c *ConcreteCell) SetPosition(x, y float32)          { c.m_x, c.m_y = x, y }
func (c *ConcreteCell) ID() int32                         { return c.m_id }
func (n *ConcreteNet) ID() int32                          { return n.m_id }
func (r Rect) Coordinates() (x, y, width, height float32) { return r.x, r.y, r.w, r.h }
