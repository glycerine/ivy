package startingpoint

import "os"

// CollectTypeNames inventories typedef and enum-tag names across the input set.
func CollectTypeNames(paths []string) (map[string]bool, error) {
	names := map[string]bool{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for name := range discoverTypeNames(string(data)) {
			names[name] = true
		}
	}
	return names, nil
}
func discoverTypeNames(source string) map[string]bool {
	names := map[string]bool{}
	all, err := Lex(source)
	if err != nil {
		return names
	}
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	for i, t := range ts {
		if t.Text == "enum" && i+2 < len(ts) && ts[i+1].Kind == "word" && ts[i+2].Text == "{" {
			names[ts[i+1].Text] = true
		}
		if t.Text != "typedef" {
			continue
		}
		level := 0
		last := ""
		for j := i + 1; j < len(ts); j++ {
			v := ts[j]
			if v.Kind == "directive" {
				break
			}
			if v.Text == "{" || v.Text == "(" || v.Text == "[" {
				level++
			}
			if v.Text == "}" || v.Text == ")" || v.Text == "]" {
				level--
			}
			if level == 0 && v.Kind == "word" {
				last = v.Text
			}
			if level == 0 && (v.Text == "," || v.Text == ";") {
				if last != "" {
					names[last] = true
				}
				last = ""
				if v.Text == ";" {
					break
				}
			}
		}
	}
	return names
}

// Blank the redundant enum keyword in type references before signature and
// declaration recognition. Byte offsets and original function text stay intact.
func normalizeEnumReferences(source string) (string, error) {
	ts, _, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	for i, t := range ts {
		if t.Text != "enum" || i+2 >= len(ts) || ts[i+1].Kind != "word" || ts[i+1].Text == "class" || ts[i+1].Text == "struct" || ts[i+2].Text == "{" || ts[i+2].Text == ":" || (i > 0 && ts[i-1].Text == "typedef") {
			continue
		}
		edits = append(edits, Edit{t.Start, t.End, "    "})
	}
	return ApplyEdits(source, edits)
}
