package startingpoint

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// LibraryChange refers to the generated draft immediately before this pass,
// never to misleading original-source byte offsets. Original C is saved separately.
type LibraryChange struct {
	Name, Status, Message string
	Location              SourceLocation
	Coordinates           string
	SourceLine            string
}

type libraryPass struct {
	source   string
	ts       []Token
	pairs    map[int]int
	imports  *libraryImports
	rules    map[string]LibraryRule
	types    map[string]string
	shadowed map[string]bool
	changes  []LibraryChange
}
type libraryCall struct {
	name       string
	token, end int
	args       []string
	rawArgs    []string
	standalone bool
}

// RewriteStandardLibrary is usable independently on an imperfect Go draft.
// No compiler AST or C runtime dependency is required. Unknown/unsafe shapes are
// retained and reported rather than translated by a misleading name substitution.
func RewriteStandardLibrary(source string) (string, []LibraryChange, error) {
	return rewriteStandardLibrary(source, true)
}
func rewriteStandardLibrary(source string, imports bool) (string, []LibraryChange, error) {
	return rewriteStandardLibraryContext(source, imports, nil)
}

func rewriteStandardLibraryContext(source string, imports bool, context *libraryEnvironment) (string, []LibraryChange, error) {
	all, err := Lex(source)
	if err != nil {
		return "", nil, err
	}
	var ts []Token
	for _, t := range all {
		if t.Kind != "space" && t.Kind != "comment" {
			ts = append(ts, t)
		}
	}
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return "", nil, err
	}
	p := &libraryPass{source: source, ts: ts, pairs: pairs, imports: newLibraryImports(source, ts, pairs), rules: map[string]LibraryRule{}, types: map[string]string{}, shadowed: map[string]bool{}}
	for _, r := range libraryCatalog {
		p.rules[r.Name] = r
	}
	if context != nil {
		p.imports = cloneLibraryImports(context.imports)
		for k, v := range context.types {
			p.types[k] = v
		}
		for k, v := range context.shadowed {
			p.shadowed[k] = v
		}
	}
	p.collectTypes()
	for name := range p.types {
		if _, ok := p.rules[name]; ok {
			p.shadowed[name] = true
		}
		if name == "stdin" || name == "stdout" || name == "stderr" {
			p.shadowed[name] = true
		}
	}
	for _, t := range all {
		if t.Kind != "directive" {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t.Text), "#")))
		if len(fields) >= 2 && fields[0] == "define" {
			name := strings.Split(fields[1], "(")[0]
			p.shadowed[name] = true
		}
	}
	for i, t := range ts {
		if (t.Text == "var" || t.Text == "const") && i+1 < len(ts) {
			p.shadowed[ts[i+1].Text] = true
		}
	}
	// A source definition/macro with a library spelling must not be rewritten.
	for i, t := range ts {
		if t.Text == "func" && i+1 < len(ts) {
			p.shadowed[ts[i+1].Text] = true
		}
	}
	text, err := p.rewrite(0, len(ts), 0, len(source))
	if err != nil {
		return "", nil, err
	}
	if imports {
		// Import edits refer to original header offsets, which precede all body edits.
		text, err = ApplyEdits(text, p.imports.edits(ts))
	}
	sort.SliceStable(p.changes, func(i, j int) bool { return p.changes[i].Location.ByteOffset < p.changes[j].Location.ByteOffset })
	return text, p.changes, err
}
func (p *libraryPass) collectTypes() {
	set := func(name, typ string) {
		if old, ok := p.types[name]; ok && old != typ {
			p.types[name] = "ambiguous"
		} else {
			p.types[name] = typ
		}
	}
	for i, t := range p.ts {
		if t.Kind != "word" || i+1 >= len(p.ts) {
			continue
		}
		next := p.ts[i+1].Text
		if next == "string" {
			set(t.Text, "string")
		}
		if next == "[" {
			close, ok := p.pairs[i+1]
			if ok && close+1 < len(p.ts) && p.ts[close+1].Text == "byte" {
				set(t.Text, "bytes")
			}
		}
		if next == "*" && i+4 < len(p.ts) && p.ts[i+3].Text == "." && p.ts[i+4].Text == "File" {
			set(t.Text, "file")
		}
		if next == "int32" || next == "int64" || next == "int" || next == "uint32" {
			set(t.Text, next)
		}
	}
}
func (p *libraryPass) record(name, status, message string, offset int) {
	prefix := p.source[:offset]
	lineStart := strings.LastIndex(prefix, "\n") + 1
	lineEnd := strings.IndexByte(p.source[offset:], '\n')
	if lineEnd < 0 {
		lineEnd = len(p.source)
	} else {
		lineEnd += offset
	}
	p.changes = append(p.changes, LibraryChange{Name: name, Status: status, Message: message, Location: SourceLocation{strings.Count(prefix, "\n") + 1, offset - lineStart + 1, offset}, Coordinates: "pre-library Go draft", SourceLine: p.source[lineStart:lineEnd]})
}
func (p *libraryPass) rewrite(lo, hi, start, end int) (string, error) {
	var edits []Edit
	for i := lo; i < hi; i++ {
		t := p.ts[i]
		member := i > 0 && (p.ts[i-1].Text == "." || p.ts[i-1].Text == "->" || p.ts[i-1].Text == "::")
		if t.Kind != "word" || member {
			continue
		}
		if replacement, ok := map[string]string{"stdin": "Stdin", "stdout": "Stdout", "stderr": "Stderr"}[t.Text]; ok && !p.shadowed[t.Text] {
			// Declaration names are not the standard stream object.
			if i+1 < hi && (p.ts[i+1].Text == "*" || p.ts[i+1].Text == "string") {
				continue
			}
			edits = append(edits, Edit{t.Start - start, t.End - start, p.imports.use("os") + replacement})
			p.record(t.Text, "converted", "Standard stream object mapped to os.", t.Start)
			continue
		}
		if replacement, ok := map[string]string{"EOF": "-1", "EXIT_SUCCESS": "0", "EXIT_FAILURE": "1"}[t.Text]; ok && !p.shadowed[t.Text] {
			edits = append(edits, Edit{t.Start - start, t.End - start, replacement})
			p.record(t.Text, "converted", "Conventional C status constant mapped; check platform-specific overrides.", t.Start)
			continue
		}
		if t.Text == "CLOCKS_PER_SEC" && !p.shadowed[t.Text] {
			edits = append(edits, Edit{t.Start - start, t.End - start, "int64(" + p.imports.use("time") + "Second)"})
			p.record(t.Text, "converted", "Nanoseconds per second for the wall-clock clock() draft policy.", t.Start)
			continue
		}
		rule, known := p.rules[t.Text]
		if !known || p.shadowed[t.Text] || i+1 >= hi || p.ts[i+1].Text != "(" {
			continue
		}
		close, ok := p.pairs[i+1]
		if !ok || close >= hi {
			continue
		}
		if close+1 < hi && p.ts[close+1].Text == "{" {
			conditionCall := i > 0 && (p.ts[i-1].Text == "if" || p.ts[i-1].Text == "for" || p.ts[i-1].Text == "!" || p.ts[i-1].Text == "&" || p.ts[i-1].Text == "|")
			if !conditionCall {
				continue
			}
		}
		// Never rewrite signatures/prototypes: at top level only a call in an
		// initializer/expression can qualify. The prefix before a call is decisive.
		if i > 0 && p.ts[i-1].Kind == "word" && p.ts[i-1].Text != "return" && p.ts[i-1].Text != "if" && p.ts[i-1].Text != "for" && !p.statementBoundary(i) {
			continue
		}
		c := libraryCall{name: t.Text, token: i, end: close}
		c.standalone = p.statementBoundary(i) && (close+1 == len(p.ts) || p.ts[close+1].Text == ";" || strings.Contains(p.source[p.ts[close].End:p.ts[close+1].Start], "\n"))
		argToken, argStart := i+2, p.ts[i+1].End
		for j := i + 2; j <= close; j++ {
			if j < close && p.ts[j].Text != "," {
				if k, ok := p.pairs[j]; ok && k > j {
					j = k
				}
				continue
			}
			if argToken < j {
				raw := p.source[argStart:p.ts[j].Start]
				text, err := p.rewrite(argToken, j, argStart, p.ts[j].Start)
				if err != nil {
					return "", err
				}
				c.args = append(c.args, strings.TrimSpace(text))
				c.rawArgs = append(c.rawArgs, strings.TrimSpace(raw))
			}
			argToken, argStart = j+1, p.ts[j].End
		}
		before := cloneLibraryImports(p.imports)
		replacement, note, converted := p.convert(c)
		if converted {
			// Preserve every original comment. If a rewrite would drop or duplicate one,
			// retain the call and offer the rule as a review item instead.
			original := p.source[t.Start:p.ts[close].End]
			if !commentsRetained(original, replacement) {
				converted = false
				note = "Call retained because rewriting would discard or duplicate an argument comment."
			}
		}
		if !converted {
			p.imports = before
			// Nested conversions still survive an unsupported outer call.

			// Reconstruct with original separators/trivia via a second local edit set.
			var nested []Edit
			at := i + 2
			pos := p.ts[i+1].End
			arg := 0
			for j := i + 2; j <= close; j++ {
				if j < close && p.ts[j].Text != "," {
					if k, ok := p.pairs[j]; ok && k > j {
						j = k
					}
					continue
				}
				if at < j && arg < len(c.args) {
					raw := p.source[pos:p.ts[j].Start]
					trim := strings.TrimSpace(raw)
					lead := strings.Index(raw, trim)
					nested = append(nested, Edit{pos + lead - t.Start, pos + lead + len(trim) - t.Start, c.args[arg]})
					arg++
				}
				at, pos = j+1, p.ts[j].End
			}
			var err error
			replacement, err = ApplyEdits(p.source[t.Start:p.ts[close].End], nested)
			if err != nil {
				return "", err
			}
			if note == "" {
				note = rule.Notes
			}
			p.record(c.name, "review", note, t.Start)
		} else {
			p.record(c.name, "converted", note, t.Start)
		}
		indent := lineIndent(p.source, t.Start)
		replacement = strings.ReplaceAll(replacement, "\n", "\n"+indent)
		edits = append(edits, Edit{t.Start - start, p.ts[close].End - start, replacement})
		i = close
	}
	return ApplyEdits(p.source[start:end], edits)
}
func cloneLibraryImports(p *libraryImports) *libraryImports {
	q := &libraryImports{map[string]string{}, map[string]bool{}, map[string]Token{}, p.used}
	for k, v := range p.existing {
		q.existing[k] = v
	}
	for k, v := range p.required {
		q.required[k] = v
	}
	for k, v := range p.blank {
		q.blank[k] = v
	}
	return q
}
func commentsRetained(before, after string) bool {
	ts, err := Lex(before)
	if err != nil {
		return false
	}
	for _, t := range ts {
		if t.Kind == "comment" && strings.Count(before, t.Text) != strings.Count(after, t.Text) {
			return false
		}
	}
	return true
}
func lineIndent(source string, at int) string {
	line := strings.LastIndex(source[:at], "\n") + 1
	j := line
	for j < at && (source[j] == ' ' || source[j] == '\t') {
		j++
	}
	return source[line:j]
}
func iife(params, result, body string, args ...string) string {
	if result != "" {
		result = " " + result
	}
	return "func(" + params + ")" + result + " {\n\t" + strings.ReplaceAll(body, "\n", "\n\t") + "\n}(" + strings.Join(args, ", ") + ")"
}
func (p *libraryPass) kind(expr string) string {
	expr = strings.TrimSpace(expr)
	if _, err := strconv.Unquote(expr); err == nil && strings.HasPrefix(expr, `"`) {
		return "string"
	}
	if strings.HasSuffix(expr, "[:]") {
		base := strings.TrimSpace(strings.TrimSuffix(expr, "[:]"))
		if kind := p.types[base]; kind == "bytes" || kind == "string" {
			return kind
		}
	}
	if strings.HasPrefix(expr, "os.Args[") {
		return "string"
	}
	return p.types[expr]
}
func (p *libraryPass) bytes(expr string) (string, bool) {
	if p.kind(expr) != "bytes" {
		return "", false
	}
	if strings.HasSuffix(expr, "[:]") {
		return expr, true
	}
	return expr + "[:]", true
}
func (p *libraryPass) text(expr string) (string, bool) {
	if p.kind(expr) == "string" {
		return expr, true
	}
	if b, ok := p.bytes(expr); ok {
		return iife("spBytes []byte", "string", "if spEnd := "+p.imports.use("bytes")+"IndexByte(spBytes, 0); spEnd >= 0 { spBytes = spBytes[:spEnd] }\nreturn string(spBytes)", b), true
	}
	return "", false
}
func (p *libraryPass) convert(c libraryCall) (string, string, bool) {
	if result, note, ok := p.convertFormatted(c); ok || note != "" {
		return result, note, ok
	}
	if result, note, ok := p.convertTime(c); ok || note != "" {
		return result, note, ok
	}
	if result, note, ok := p.convertIO(c); ok || note != "" {
		return result, note, ok
	}
	if result, note, ok := p.convertText(c); ok || note != "" {
		return result, note, ok
	}
	if result, note, ok := p.convertNumeric(c); ok || note != "" {
		return result, note, ok
	}
	return "", p.rules[c.name].Notes, false
}

// Keep messages compact enough for per-call JSON review records.
func arity(c libraryCall, n int) string {
	return fmt.Sprintf("%s requires %d arguments; got %d.", c.name, n, len(c.args))
}

func (p *libraryPass) statementBoundary(i int) bool {
	if i == 0 {
		return true
	}
	prev := p.ts[i-1]
	if prev.Text == "{" || prev.Text == "}" || prev.Text == ";" || prev.Text == ":" {
		return true
	}
	if !strings.Contains(p.source[prev.End:p.ts[i].Start], "\n") {
		return false
	}
	switch prev.Text {
	case "return", "int", "char", "void", "FILE", "unsigned", "static", "extern", "inline":
		return false
	}
	return prev.Kind == "word" || prev.Kind == "string" || prev.Text == ")" || prev.Text == "]"
}

// Reuse file-wide context for metadata rewrites instead of lexing the full
// expanded file once per function.
type libraryEnvironment struct {
	imports  *libraryImports
	types    map[string]string
	shadowed map[string]bool
}

func newLibraryEnvironment(source string) (*libraryEnvironment, error) {
	all, err := Lex(source)
	if err != nil {
		return nil, err
	}
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
	env := &libraryEnvironment{newLibraryImports(source, ts, pairs), map[string]string{}, map[string]bool{}}
	p := &libraryPass{ts: ts, pairs: pairs, types: env.types}
	p.collectTypes()
	for i, t := range ts {
		if t.Text == "func" && i+1 < len(ts) {
			env.shadowed[ts[i+1].Text] = true
		}
	}
	return env, nil
}
