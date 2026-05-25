package ivy2go

// definitions.go mirrors ivy2cpp/definitions.go. Emits Go functions
// for definitional axioms (sym := expr) declared in the module.
//
// M4 ships only the dispatcher hook. Full per-definition emission
// lands alongside state in M5 once Generator has a stable expression
// receiver for `s.<field>` access.

// emitDefinitionDecls is the M4 stub.
func (g *Generator) emitDefinitionDecls(w *goWriter) {
	_ = w
}
