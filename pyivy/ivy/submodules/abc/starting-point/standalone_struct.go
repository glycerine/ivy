package startingpoint

import (
	"regexp"
	"strings"
)

var standaloneStruct = regexp.MustCompile(`\bstruct\s+([A-Za-z_]\w*)\s*\{`)

func standaloneStructEdits(source string, ts []Token, ends map[int]int, path string, existing []Edit) ([]Edit, []Issue) {
	mask := Mask(ts)
	var edits []Edit
	var issues []Issue
	covered := func(start, end int) bool {
		for _, edit := range existing {
			if edit.Start < end && edit.End > start {
				return true
			}
		}
		return false
	}
	depthAt := map[int]int{}
	previousAt := map[int]string{}
	depth, previous := 0, ""
	for _, t := range ts {
		if t.Kind == "space" || t.Kind == "comment" {
			continue
		}
		depthAt[t.Start], previousAt[t.Start] = depth, previous
		if t.Text == "{" {
			depth++
		}
		if t.Text == "}" {
			depth--
		}
		previous = t.Text
	}
	for _, loc := range standaloneStruct.FindAllStringSubmatchIndex(mask, -1) {
		start, open := loc[0], loc[1]-1
		if depthAt[start] != 0 || previousAt[start] == "typedef" {
			continue
		}
		end, ok := ends[open]
		if !ok || covered(start, end) {
			continue
		}
		tail := end
		for tail < len(mask) && strings.ContainsRune(" \t\r\n", rune(mask[tail])) {
			tail++
		}
		if tail >= len(mask) || mask[tail] != ';' {
			continue
		}
		blocked := false
		for _, tok := range ts {
			if tok.Start >= start && tok.End <= end && tok.Kind == "directive" {
				blocked = true
			}
		}
		if blocked {
			continue
		}
		tag := mask[loc[2]:loc[3]]
		edits = append(edits, Edit{start, start + len("struct"), "type"}, Edit{loc[2], loc[3], tag + " struct"}, Edit{tail, tail + 1, ""})
		edits = append(edits, structBraceEdits(source, loc[3], open)...)
		fe, fi := structFieldEdits(source, ts, open, end, tag, tag, path)
		edits = append(edits, fe...)
		issues = append(issues, fi...)
		issues = append(issues, Issue{ID: path + "::struct::" + tag, Category: "struct-review", Symbol: tag, Status: "open", Message: "Standalone struct translated using its tag; typedef aliases may be defined in another file. Review pointer representation and retained complex fields.", Start: start, End: tail + 1})
	}
	for _, loc := range structAlias.FindAllStringSubmatchIndex(mask, -1) {
		if depthAt[loc[0]] != 0 || covered(loc[0], loc[1]) {
			continue
		}
		// Edit code tokens only to retain any typedef comments.
		var code []Token
		for _, tok := range ts {
			if tok.Start >= loc[0] && tok.End <= loc[1] && tok.Kind != "space" && tok.Kind != "comment" {
				code = append(code, tok)
			}
		}
		if len(code) != 5 {
			continue
		}
		tag, alias := mask[loc[2]:loc[3]], mask[loc[4]:loc[5]]
		if tag == alias {
			for _, tok := range code {
				edits = append(edits, Edit{tok.Start, tok.End, ""})
			}
		} else {
			edits = append(edits, Edit{code[0].Start, code[0].End, "type"}, Edit{code[1].Start, code[1].End, alias}, Edit{code[2].Start, code[2].End, "="}, Edit{code[3].Start, code[3].End, tag})
		}
	}
	return edits, issues
}
