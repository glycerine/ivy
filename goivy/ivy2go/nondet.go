package ivy2go

// nondet.go mirrors ivy2cpp/nondet.go. The ivyChoose runtime helper is
// emitted by runtime.go (M3). Per-sort randomizers (havoc helpers) and
// the nondet symbol enumeration that feeds Init land in M5.

// emitNondetDecls is the M4 stub. M5 fleshes it out alongside Init.
func (g *Generator) emitNondetDecls(w *goWriter) {
	_ = w
}
