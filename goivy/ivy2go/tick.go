package ivy2go

// tick.go mirrors ivy2cpp/tick.go. M7 emits a no-op Tick() method;
// progress / rely-decl based scheduling lands in a later milestone
// alongside any concurrency support.

// emitTickMethod writes a (*State).Tick() method into the runtime
// stream. The body is a no-op for M7; future milestones can append
// progress / rely scheduling logic here.
func (g *Generator) emitTickMethod(w *goWriter) {
	w.linef("// Tick advances any timed / progress-tracked state. M7 stub: no-op.")
	w.linef("func (s *%s) Tick() {}", g.StateTypeName)
	w.blank()
}
