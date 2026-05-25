package ivy2go

// constructors.go mirrors ivy2cpp/constructors.go. Sort constructor
// emission lands in M5/M8. M2 provides no-op stubs so the rest of the
// package can refer to the function names without forward references.

func (g *Generator) emitConstructorDecls(w *goWriter) {
	_ = w
}

// emitNativeTypeDecl is the M2 stub for native-typed sort declarations
// (e.g., `<<< go ... type Foo bytes32 ... >>>`). Real handling lands
// in M10.
func (g *Generator) emitNativeTypeDecl(w *goWriter, name, nativeType string) {
	w.linef("// native type %s = %s (deferred to M10)", goExportedName(name), nativeType)
}
