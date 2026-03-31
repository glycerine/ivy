package gogen

import (
	"fmt"
	"strings"

	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// GoType maps an Ivy sort to its Go type string.
func GoType(s lg.Sort) string {
	switch st := s.(type) {
	case *lg.BooleanSort:
		return "bool"
	case *lg.EnumeratedSort:
		name := st.Name
		if name == "" {
			name = "Enum"
		}
		return goExportedName(name)
	case *lg.RangeSort:
		return "int"
	case *lg.UninterpretedSort:
		return "int"
	case *lg.FunctionSort:
		dom := st.Domain()
		rng := st.Range()
		if len(dom) == 0 {
			return GoType(rng)
		}
		// Single-argument function: map[K]V
		if len(dom) == 1 {
			return fmt.Sprintf("map[%s]%s", GoType(dom[0]), GoType(rng))
		}
		// Multi-argument: use struct key or nested map.
		// We use a flat map with array key for simplicity.
		keyParts := make([]string, len(dom))
		for i, d := range dom {
			keyParts[i] = GoType(d)
		}
		// If all domain types are the same, use [N]T as key.
		allSame := true
		for _, k := range keyParts {
			if k != keyParts[0] {
				allSame = false
				break
			}
		}
		if allSame {
			return fmt.Sprintf("map[[%d]%s]%s", len(dom), keyParts[0], GoType(rng))
		}
		// Mixed types: use a struct key type (represented as string for now).
		return fmt.Sprintf("map[%sKey]%s", goExportedName(keyParts[0]), GoType(rng))
	case *lg.TopSort:
		return "interface{}"
	default:
		return "interface{}"
	}
}

// StateFieldType determines the Go type for a state variable given its
// symbol name and sort from the module's Relations or Functions map.
//
// Nullary bool relation -> bool
// Unary relation (K -> bool) -> map[K]bool
// Binary relation (K1,K2 -> bool) -> map[[2]K]bool (or nested)
// Function (K -> V) -> map[K]V
// Individual / constant -> value type
func StateFieldType(name string, s lg.Sort) string {
	fs, ok := s.(*lg.FunctionSort)
	if !ok {
		// Non-function sort: individual/constant value.
		return GoType(s)
	}
	dom := fs.Domain()
	rng := fs.Range()
	if len(dom) == 0 {
		return GoType(rng)
	}
	_, isBool := rng.(*lg.BooleanSort)
	if len(dom) == 1 {
		if isBool {
			return fmt.Sprintf("map[%s]bool", GoType(dom[0]))
		}
		return fmt.Sprintf("map[%s]%s", GoType(dom[0]), GoType(rng))
	}
	// Multi-argument
	keyParts := make([]string, len(dom))
	for i, d := range dom {
		keyParts[i] = GoType(d)
	}
	allSame := true
	for _, k := range keyParts {
		if k != keyParts[0] {
			allSame = false
			break
		}
	}
	valType := GoType(rng)
	if allSame {
		return fmt.Sprintf("map[[%d]%s]%s", len(dom), keyParts[0], valType)
	}
	// Fallback: nested maps for two distinct types.
	if len(dom) == 2 {
		return fmt.Sprintf("map[%s]map[%s]%s", keyParts[0], keyParts[1], valType)
	}
	return fmt.Sprintf("map[interface{}]%s", valType)
}

// GoZeroValue returns the default zero value string for a Go type derived from an Ivy sort.
func GoZeroValue(s lg.Sort) string {
	switch s.(type) {
	case *lg.BooleanSort:
		return "false"
	case *lg.EnumeratedSort:
		return "0"
	case *lg.RangeSort:
		return "0"
	case *lg.UninterpretedSort:
		return "0"
	case *lg.FunctionSort:
		return "nil"
	default:
		return "nil"
	}
}

// GoSortValues returns all values for finite sorts, suitable for
// quantifier iteration loops. Returns nil for infinite/unknown sorts.
func GoSortValues(s lg.Sort) []string {
	switch st := s.(type) {
	case *lg.BooleanSort:
		return []string{"false", "true"}
	case *lg.EnumeratedSort:
		vals := make([]string, len(st.Extension))
		for i, ext := range st.Extension {
			vals[i] = goExportedName(ext)
		}
		return vals
	case *lg.RangeSort:
		// Range values are generated at runtime via a loop, not enumerated here.
		return nil
	case *lg.UninterpretedSort:
		// Uninterpreted sorts have a finite domain at runtime, but values
		// are not known at code-gen time.
		return nil
	default:
		return nil
	}
}

// EmitSortDecls emits all Go type declarations for the sorts in a module.
func EmitSortDecls(w *CodeWriter, mod *module.Module) {
	if mod.Sig == nil {
		return
	}
	// Emit enumerated sorts.
	for _, sortName := range mod.SortOrder {
		s, ok := mod.Sig.Sorts[sortName]
		if !ok {
			continue
		}
		if es, ok := s.(*lg.EnumeratedSort); ok {
			EmitEnumDecl(w, es)
			w.BlankLine()
		}
	}
}

// EmitEnumDecl emits a Go enum type declaration for an EnumeratedSort:
//
//	type Color int
//	const (
//	    Red Color = iota
//	    Green
//	    Blue
//	)
//	var allColor = [...]Color{Red, Green, Blue}
func EmitEnumDecl(w *CodeWriter, sort *lg.EnumeratedSort) {
	typeName := goExportedName(sort.Name)
	if typeName == "" {
		typeName = "Enum"
	}
	w.Linef("type %s int", typeName)
	w.BlankLine()
	if len(sort.Extension) > 0 {
		w.OpenBlock("const (")
		for i, val := range sort.Extension {
			name := goExportedName(val)
			if i == 0 {
				w.Linef("%s %s = iota", name, typeName)
			} else {
				w.Line(name)
			}
		}
		w.CloseBlock()
		w.BlankLine()
		// Emit the allX slice for quantifier iteration.
		vals := make([]string, len(sort.Extension))
		for i, v := range sort.Extension {
			vals[i] = goExportedName(v)
		}
		w.Linef("var all%s = [...]%s{%s}", typeName, typeName, strings.Join(vals, ", "))
	}
}

// EmitRangeHelpers emits helper constants/variables for a RangeSort,
// including lo/hi bounds and an iteration slice.
func EmitRangeHelpers(w *CodeWriter, sort *lg.RangeSort) {
	name := goExportedName(sort.Name)
	if name == "" {
		name = "Range"
	}
	w.Linef("const %sLo = %s", name, sort.LbString())
	w.Linef("const %sHi = %s", name, sort.UbString())
}

// goExportedName converts a name to a Go-exported identifier by
// uppercasing the first letter. Dots are replaced with underscores.
func goExportedName(name string) string {
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, ".", "_")
	// Uppercase first rune.
	runes := []rune(name)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 'a' + 'A'
	}
	return string(runes)
}

// goUnexportedName converts a name to a Go-unexported identifier by
// lowercasing the first letter. Dots are replaced with underscores.
func goUnexportedName(name string) string {
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, ".", "_")
	runes := []rune(name)
	if runes[0] >= 'A' && runes[0] <= 'Z' {
		runes[0] = runes[0] - 'A' + 'a'
	}
	return string(runes)
}
