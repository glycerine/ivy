package startingpoint

import (
	"regexp"
	"strconv"
	"strings"
)

// Local C definitions distinguish presence from numeric truth: #define X 0
// satisfies #ifdef X, but not #if X. Explicit configuration keeps its documented
// zero-means-disabled policy and overrides local inference.
type localPreprocessor struct {
	present map[string]bool
	numbers map[string]int64
}

func newLocalPreprocessor() *localPreprocessor {
	return &localPreprocessor{map[string]bool{}, map[string]int64{}}
}
func (p *localPreprocessor) forget(d Token) {
	kind, name := macroDirective(d.Text)
	if kind == "include" {
		p.present = map[string]bool{}
		p.numbers = map[string]int64{}
	} else if kind == "define" || kind == "undef" {
		delete(p.present, name)
		delete(p.numbers, name)
	}
}
func (p *localPreprocessor) observe(d Token) {
	kind, name := macroDirective(d.Text)
	if kind == "include" {
		p.forget(d)
		return
	}
	if kind != "define" && kind != "undef" {
		return
	}
	delete(p.numbers, name)
	p.present[name] = kind == "define"
	if kind == "undef" {
		p.numbers[name] = 0
		return
	}
	if _, function := parseInlineMacro(d.Text); function {
		return
	}
	parts := directiveParts.FindStringSubmatch(d.Text)
	if parts == nil {
		return
	}
	value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[2]), name))
	ts, err := Lex(value)
	if err != nil {
		return
	}
	value = strings.TrimSpace(Mask(ts))
	value = strings.TrimRight(value, "uUlL")
	if n, err := strconv.ParseInt(value, 0, 64); err == nil {
		p.numbers[name] = n
	}
}

var localDefined = regexp.MustCompile(`\bdefined\s*\(\s*([A-Za-z_]\w*)\s*\)`)

func (p *localPreprocessor) condition(d Token, configured map[string]int64) (bool, bool) {
	values := map[string]int64{}
	for name, value := range p.numbers {
		values[name] = value
	}
	for name, value := range configured {
		values[name] = value
	}
	kind, name := macroDirective(d.Text)
	parts := directiveParts.FindStringSubmatch(d.Text)
	if parts == nil {
		return false, false
	}
	if kind == "ifdef" || kind == "ifndef" {
		if _, explicit := configured[name]; !explicit {
			if present, known := p.present[name]; known {
				if kind == "ifndef" {
					return !present, true
				}
				return present, true
			}
		}
		return configCondition(d, values)
	}
	expr := strings.ReplaceAll(strings.ReplaceAll(parts[2], "\\\r\n", ""), "\\\n", "")
	tokens, err := Lex(expr)
	if err != nil {
		return false, false
	}
	expr = Mask(tokens)
	expr = definedWithoutParens.ReplaceAllString(expr, "defined($1)")
	expr = localDefined.ReplaceAllStringFunc(expr, func(s string) string {
		name := localDefined.FindStringSubmatch(s)[1]
		if _, explicit := configured[name]; explicit {
			return s
		}
		if present, known := p.present[name]; known {
			if present {
				return "1"
			}
			return "0"
		}
		return s
	})
	d.Text = "#" + kind + " " + expr
	return configCondition(d, values)
}
