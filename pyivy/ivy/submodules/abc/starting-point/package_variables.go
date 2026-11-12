package startingpoint

import (
	"strings"
)

// packageVariableEdits routes simple file-scope variables through the existing
// declaration lowering, and comments out function prototypes without bodies.
func packageVariableEdits(source string, all []Token, types map[string]bool, opaque []opaqueRegion) ([]Edit, error) {
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return nil, err
	}
	var edits []Edit
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		if t.Text == "{" || t.Text == "(" || t.Text == "[" {
			i = pairs[i]
			continue
		}
		if t.Kind != "word" {
			continue
		}
		if i > 0 && ts[i-1].Text != ";" && ts[i-1].Text != "}" && ts[i-1].Kind != "directive" {
			line := strings.LastIndex(source[:t.Start], "\n") + 1
			if strings.TrimSpace(source[line:t.Start]) != "" {
				continue
			}
		}
		typeIndex := i
		for typeIndex < len(ts) && (ts[typeIndex].Text == "static" || ts[typeIndex].Text == "const" || ts[typeIndex].Text == "extern" || ts[typeIndex].Text == "inline") {
			typeIndex++
		}
		if typeIndex >= len(ts) {
			continue
		}
		typ := ts[typeIndex].Text
		if typ != "void" && !isDeclarationType(typ, types) {
			continue
		}
		end := typeIndex + 1
		initialized := false
		for ; end < len(ts); end++ {
			if ts[end].Text == ";" || ts[end].Kind == "directive" {
				break
			}
			if ts[end].Text == "=" {
				initialized = true
			}
			if ts[end].Text == "{" && !initialized {
				break
			}
			if close, ok := pairs[end]; ok && close > end {
				end = close
			}
		}
		if end >= len(ts) || ts[end].Text != ";" {
			continue
		}
		overlap := false
		for _, region := range opaque {
			if t.Start < region.projectedEnd && ts[end].End > region.projectedStart {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		if functionPrototype(ts, pairs, typeIndex, end) {
			original := source[t.Start:ts[end].End]
			commented := "// " + strings.ReplaceAll(original, "\n", "\n// ")
			// Do not swallow a declaration or definition following on the same line.
			commented += "\n"
			edits = append(edits, Edit{t.Start, ts[end].End, commented})
			i = end
			continue
		}
		// Retain comments in qualified declarations instead of erasing them.
		if strings.Contains(source[t.Start:ts[typeIndex].Start], "/*") || strings.Contains(source[t.Start:ts[typeIndex].Start], "//") {
			continue
		}
		fragment := source[ts[typeIndex].Start:ts[end].End]
		wrapped := "{" + fragment + "}"
		converted, err := lowerDeclarations(wrapped, types, true)
		if err != nil {
			return nil, transformedDiagnostic("package variable", "declaration lowering", wrapped, err)
		}
		if converted == wrapped {
			continue
		}
		// Compose ordinary expression rewrites (NULL, pointer-member access) too.
		converted, err = rewriteBody(converted, types)
		if err != nil {
			return nil, err
		}
		edits = append(edits, Edit{t.Start, ts[end].End, converted[1 : len(converted)-1]})
		i = end
	}
	return edits, nil
}

// A direct function declarator has a name immediately before its argument list,
// and no initializer or body. Pointer variables and function-pointer variables
// are deliberately excluded.
func functionPrototype(ts []Token, pairs map[int]int, start, end int) bool {
	for j := start + 1; j < end; j++ {
		if ts[j].Text == "(" {
			return j > start+1 && ts[j-1].Kind == "word" && pairs[j] == end-1
		}
		if ts[j].Kind != "word" && ts[j].Text != "*" {
			return false
		}
	}
	return false
}
