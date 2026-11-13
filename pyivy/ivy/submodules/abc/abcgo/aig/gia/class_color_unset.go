package gia

// Gia_ObjUnsetColor clears one representative color without changing the
// other color or the object's proof and class metadata.
func Gia_ObjUnsetColor(p *Gia_Man_t, id, color int32) {
	if color != 0 {
		p.pReprs[id].fColorB = 0
	} else {
		p.pReprs[id].fColorA = 0
	}
}
