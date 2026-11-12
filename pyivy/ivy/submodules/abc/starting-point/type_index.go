package startingpoint

import "os"

// CollectTypeNames inventories typedef names without preprocessing includes or
// inferring memory representation. Names are hints for declaration recognition.
func CollectTypeNames(paths []string) (map[string]bool, error) {
	names := map[string]bool{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		all, err := Lex(string(data))
		if err != nil {
			continue
		}
		var ts []Token
		for _, t := range all {
			if t.Kind != "space" && t.Kind != "comment" {
				ts = append(ts, t)
			}
		}
		for i, t := range ts {
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
	}
	return names, nil
}
