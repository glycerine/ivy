package ivy2go

// extensional.go mirrors ivy2cpp/extensional.go. Extensional relation
// iteration is part of the quantifier / large-assignment lowering;
// both are deferred. M4 ships the hook only.

// extensionalRelations mirrors ivy2cpp Generator.extensionalRelations.
// Stub returns no extensional relations until M5.
func (g *Generator) extensionalRelations() map[string]bool {
	if g.extRel == nil {
		g.extRel = map[string]bool{}
	}
	return g.extRel
}

// emitExtensionalDecls is the M4 stub.
func (g *Generator) emitExtensionalDecls(w *goWriter) {
	_ = w
}
