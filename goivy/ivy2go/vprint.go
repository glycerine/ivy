package ivy2go

// vprint.go mirrors ivy2cpp/vprint.go. Trace LHS/RHS printing wires
// into Config.Trace and the assignment emitter. M7 ships only the
// hook; M8 (alongside thunks) and M10 (alongside native blocks) fill
// it in.

// numberFormat returns the trace number format flag suffix. Mirrors
// ivy2cpp/vprint.go numberFormat. M7 returns the empty string; M8/M10
// may set "_hex" when the module attributes request hex tracing.
func (g *Generator) numberFormat() string {
	if g == nil {
		return ""
	}
	if g.numberFormatComputed {
		return g.numberFormatCache
	}
	g.numberFormatComputed = true
	g.numberFormatCache = ""
	return g.numberFormatCache
}
