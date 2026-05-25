package ivy2go

// native_thunk.go mirrors ivy2cpp/native_thunk.go. Wraps user-supplied
// native action blocks in thunks so they can be passed around the
// runtime like ordinary actions. M8 ships only the hook; M10 (native
// blocks) fills it in.

// emitNativeThunkDecls is the M8 stub.
func (g *Generator) emitNativeThunkDecls(w *goWriter) {
	_ = w
}
