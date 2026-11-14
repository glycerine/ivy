package gia

// Gia_ManCopyNameSpec copies the source metadata used by graph duplication.
func Gia_ManCopyNameSpec(dst, src *Gia_Man_t) {
	if dst == nil || src == nil {
		panic("gia: nil manager when copying name and spec")
	}
	dst.pName, dst.PSpec = src.pName, src.PSpec
}
