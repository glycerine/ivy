package startingpoint

import "fmt"

// Report representation decisions at original source offsets rather than
// manufacturing an apparently equivalent Go layout or ownership model.
func mechanicalReviewIssues(path, source string) []Issue {
	ts, err := Lex(source)
	if err != nil {
		return nil
	}
	var issues []Issue
	notes := map[string]string{
		"union":  "Union members share storage in C. Choose a tagged representation or an explicit shared-storage model; independent Go fields are not equivalent.",
		"sizeof": "C object size/layout may differ from Go. Whole known struct zeroing is handled separately; otherwise establish the representation before replacing byte counts.",
		"enum":   "Ordinary enum definitions and references are lowered. Retained C++ scoped enums and trailing object declarators require type and name decisions.",
		"static": "Local-static storage may be lifted when declarations and names are unambiguous. Review retained complex static initializers and shadowed names; deleting static changes lifetime.",
		"do":     "Simple do/while bodies are lowered. Retained loops with continue need control-flow repair so continue still evaluates the condition.",
	}
	depth := 0
	for _, t := range ts {
		if t.Text == "{" {
			depth++
		}
		if t.Text == "}" {
			depth--
		}
		note, ok := notes[t.Text]
		if t.Kind != "word" || !ok {
			continue
		}
		if t.Text == "static" && depth == 0 {
			continue
		} // ordinary function/storage qualifiers are already handled
		issues = append(issues, Issue{ID: fmt.Sprintf("%s::mechanical::%s::%d", path, t.Text, t.Start), Category: "mechanical-review", Status: "open", Message: note, Start: t.Start, End: t.End})
	}
	significant, _, err := significantTokens(source)
	if err == nil {
		for i, t := range significant {
			if t.Text != "(" || i+3 >= len(significant) || significant[i+1].Kind != "word" || significant[i+2].Text != "*" {
				continue
			}
			end := i + 2
			for end < len(significant) && significant[end].Text == "*" {
				end++
			}
			if end < len(significant) && significant[end].Text == ")" {
				issues = append(issues, Issue{ID: fmt.Sprintf("%s::pointer-cast::%d", path, t.Start), Category: "representation-review", Status: "open", Message: "Pointer cast retained: decide pointer/slice/string representation, aliasing and any tagged-pointer bits before translating.", Start: t.Start, End: significant[end].End})
			}
		}
	}
	return issues
}
