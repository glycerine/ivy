package startingpoint

import (
	"strings"
)

// packageVariableEdits routes simple file-scope variables through the existing
// declaration lowering without applying body transforms to types or prototypes.
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
		for typeIndex < len(ts) && (ts[typeIndex].Text == "static" || ts[typeIndex].Text == "const") {
			typeIndex++
		}
		if typeIndex >= len(ts) {
			continue
		}
		typ := ts[typeIndex].Text
		if !types[typ] && !strings.HasSuffix(typ, "_t") {
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
