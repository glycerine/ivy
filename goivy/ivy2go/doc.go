// Package ivy2go generates runnable Go source code from a goivy Module.
//
// It is the Go-emitting sibling of package ivy2cpp. ivy2cpp ports
// pyivy's ivy_to_cpp.py to emit C++; ivy2go ports ivy2cpp's structure
// file-for-file to emit Go instead.
//
// Pipeline (mirrors ivy2cpp):
//
//	CompileAndGenerate(filename, params, Config)
//	  -> goivy.SourceFile, isolate selection, Generate per isolate
//	  -> *Output { Files map[string]string, ExtraFiles, ... }
//	  -> WriteOutput(out, outDir)
//	  -> BuildOutput(out, outDir) running `go build`
//
// Emitted programs link against goivy directly so they can use the
// existing goivy.Solver / goivy.Translator facade for any runtime SMT
// queries (target=test, target=gen). The generator itself also uses
// goivy.Solver — never the lower-level goivy/smt package directly.
//
// See ARCHITECTURE_TODO.md in this directory for the full design and
// milestone plan; see CLAUDE.md for mechanical-port rules.
package ivy2go
