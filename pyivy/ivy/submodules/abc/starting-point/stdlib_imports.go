package startingpoint

import (
	"path"
	"sort"
	"strconv"
	"strings"
)

func splitWords(s string) []string { return strings.Fields(s) }

type libraryImports struct {
	existing map[string]string
	required map[string]bool
	blank    map[string]Token
	used     map[string]bool
}

func newLibraryImports(source string, ts []Token, pairs map[int]int) *libraryImports {
	p := &libraryImports{map[string]string{}, map[string]bool{}, map[string]Token{}, map[string]bool{}}
	for i, t := range ts {
		if t.Kind == "word" {
			if t.Text == "time" && i+1 < len(ts) && ts[i+1].Text == "(" && (i == 0 || ts[i-1].Text != "func") {
				continue
			}
			p.used[t.Text] = true
		}
	}
	for i, t := range ts {
		if t.Text != "import" || i+1 >= len(ts) {
			continue
		}
		start, end := i+1, i+2
		if ts[start].Text == "(" {
			end = pairs[start]
			start++
		} else if ts[start].Kind != "string" {
			end++
		}
		if end > len(ts) {
			end = len(ts)
		}
		for j := start; j < end; j++ {
			if ts[j].Kind != "string" {
				continue
			}
			name, err := strconv.Unquote(ts[j].Text)
			if err != nil {
				continue
			}
			alias := path.Base(name)
			if j > start && (ts[j-1].Kind == "word" || ts[j-1].Text == ".") {
				alias = ts[j-1].Text
			}
			p.existing[name] = alias
			if alias == "_" {
				p.blank[name] = ts[j-1]
			}
		}
	}
	return p
}
func (p *libraryImports) use(name string) string {
	p.required[name] = true
	if alias, ok := p.existing[name]; ok && alias != "_" {
		if alias == "." {
			return ""
		}
		return alias + "."
	}
	base := path.Base(name)
	alias := base
	// A generated import must not shadow a source local with the same spelling.
	// Existing imports already own their names and are reused above.
	if p.aliasTaken(alias) {
		alias = "sp" + strings.ToUpper(base[:1]) + base[1:]
		for p.aliasTaken(alias) {
			alias += "_"
		}
	}
	p.existing[name] = alias
	return alias + "."
}
func (p *libraryImports) edits(ts []Token) []Edit {
	var names []string
	for name := range p.required {
		names = append(names, name)
	}
	sort.Strings(names)
	var edits []Edit
	var imports []string
	for _, name := range names {
		if blank, ok := p.blank[name]; ok {
			edits = append(edits, Edit{blank.Start, blank.End, p.existing[name]})
			continue
		}
		// Source imports have no pending marker; identify them by tokens.
		found := false
		for i, t := range ts {
			if t.Text != "import" {
				continue
			}
			for j := i + 1; j < len(ts); j++ {
				if ts[j].Text == ")" || ts[j].Text == "func" || ts[j].Text == "var" || ts[j].Text == "type" {
					break
				}
				if ts[j].Kind == "string" {
					value, _ := strconv.Unquote(ts[j].Text)
					if value == name {
						found = true
					}
					if i+1 < len(ts) && ts[i+1].Text != "(" {
						break
					}
				}
			}
		}
		if found {
			continue
		}
		spec := strconv.Quote(name)
		if alias := p.existing[name]; alias != path.Base(name) {
			spec = alias + " " + spec
		}
		imports = append(imports, spec)
	}
	if len(imports) > 0 {
		for i, t := range ts {
			if t.Text == "package" && i+1 < len(ts) {
				text := "\n\nimport " + imports[0]
				if len(imports) > 1 {
					text = "\n\nimport (\n\t" + strings.Join(imports, "\n\t") + "\n)"
				}
				edits = append(edits, Edit{ts[i+1].End, ts[i+1].End, text})
				break
			}
		}
	}
	return edits
}

func (p *libraryImports) aliasTaken(alias string) bool {
	if p.used[alias] {
		return true
	}
	for _, existing := range p.existing {
		if existing == alias {
			return true
		}
	}
	return false
}
