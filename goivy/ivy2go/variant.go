package ivy2go

import "github.com/glycerine/ivy/goivy"

// variant.go mirrors ivy2cpp/variant.go. Full variant (tagged union)
// emission is scheduled for M8. M2 provides the hooks so emitSortDecls
// can compile and run; the hooks all return "no variant here".

func (g *Generator) variantSuperName(s goivy.Sort) (string, bool) {
	_ = s
	return "", false
}

func (g *Generator) variantSubtypeName(s goivy.Sort) (string, bool) {
	_ = s
	return "", false
}

func (g *Generator) isVariantSuperName(name string) bool {
	_ = name
	return false
}

func (g *Generator) isVariantSubtypeName(name string) bool {
	_ = name
	return false
}

func (g *Generator) isPlainVariantSubtypeName(name string) bool {
	_ = name
	return false
}

// emitVariantSuperStruct and emitVariantLeafStruct are M2 stubs that
// emit nothing. M8 fills them in.
func (g *Generator) emitVariantSuperStruct(w *goWriter, name string) {
	_ = w
	_ = name
}

func (g *Generator) emitVariantLeafStruct(w *goWriter, name string) {
	_ = w
	_ = name
}

// Silence unused-import / dead-code complaints during M2.
var _ goivy.Sort
