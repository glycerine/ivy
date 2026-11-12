package startingpoint

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type inlineMacro struct {
	params       []string
	body, reason string
}
type macroInliner struct {
	opts          Options
	macros        map[string]inlineMacro
	visiting      map[string]bool
	once          map[string]bool
	issues        []Issue
	steps         int
	expandedBytes int
	annotate      bool
}

const macroExpansionLimit = 1 << 20

// Expansion coordinates map back to the invocation in the original source.
// Bytes outside replacements retain their exact original offsets.
func macroOriginalOffset(offset int, edits []Edit) int {
	delta := 0
	for _, e := range edits {
		start := e.Start + delta
		if offset < start {
			break
		}
		if offset < start+len(e.Text) {
			return e.Start
		}
		delta += len(e.Text) - (e.End - e.Start)
	}
	return offset - delta
}

func analyzeInlineMacros(path, source string, opts Options) (Result, error) {
	p := &macroInliner{annotate: true, opts: opts, macros: map[string]inlineMacro{}, visiting: map[string]bool{}, once: map[string]bool{}}
	expanded, edits, err := p.process(path, source, true)
	if err != nil {
		return Result{}, err
	}
	opts.inlinePrepared = true
	r, err := AnalyzeWithOptions(path, expanded, opts)
	r.InlineMacros = true
	r.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(source)))
	if err != nil {
		var diagnostic *SourceError
		if errors.As(err, &diagnostic) {
			err = sourceError(source, &syntaxError{diagnostic.Message, macroOriginalOffset(diagnostic.Location.ByteOffset, edits)})
		}
		return r, err
	}
	for i := range r.Functions {
		f := &r.Functions[i]
		f.Start = macroOriginalOffset(f.Start, edits)
		f.End = macroOriginalOffset(f.End, edits)
		f.BodyStart = macroOriginalOffset(f.BodyStart, edits)
		f.Original = source[f.Start:f.End]
	}
	for i := range r.Issues {
		r.Issues[i].Start = macroOriginalOffset(r.Issues[i].Start, edits)
		r.Issues[i].End = macroOriginalOffset(r.Issues[i].End, edits)
	}
	for i := range r.Conditionals {
		off := macroOriginalOffset(r.Conditionals[i].Location.ByteOffset, edits)
		prefix := source[:off]
		r.Conditionals[i].Location = SourceLocation{Line: strings.Count(prefix, "\n") + 1, Column: off - strings.LastIndex(prefix, "\n"), ByteOffset: off}
	}
	r.Issues = append(r.Issues, p.issues...)
	r.Draft, err = macroInliningComments(r.Draft)
	if err != nil {
		return r, err
	}
	for i := range r.Functions {
		r.Functions[i].Draft, err = macroInliningComments(r.Functions[i].Draft)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func (p *macroInliner) review(path, name, message string, start, end int) {
	p.issues = append(p.issues, Issue{ID: fmt.Sprintf("%s::macro-inline::%d", path, start), Category: "macro-inline-review", Symbol: name, Status: "open", Message: message, Start: start, End: end})
}

// Process directives in source order, following local includes only. Unresolved
// conditional regions are kept opaque and their macro definitions are blocked.
func (p *macroInliner) process(path, source string, emit bool) (string, []Edit, error) {
	absolute, _ := filepath.Abs(path)
	if p.visiting[absolute] || p.once[absolute] {
		return source, nil, nil
	}
	if len(p.visiting) > 128 {
		return source, nil, fmt.Errorf("macro include nesting limit at %s", path)
	}
	p.visiting[absolute] = true
	defer delete(p.visiting, absolute)
	selected, err := selectConfiguredBranches(source, p.opts.Defines)
	if err != nil {
		return "", nil, sourceError(source, err)
	}
	_, regions, err := projectConditionals(selected)
	if err != nil {
		return "", nil, sourceError(source, err)
	}
	ts, err := Lex(selected)
	if err != nil {
		return "", nil, sourceError(source, err)
	}
	if len(headerGuardEdits(ts)) > 0 {
		p.once[absolute] = true
	}
	var edits []Edit
	regionIndex := 0
	for i := 0; i < len(ts); {
		t := ts[i]
		for regionIndex < len(regions) && t.Start >= regions[regionIndex].end {
			regionIndex++
		}
		if regionIndex < len(regions) && t.Start >= regions[regionIndex].start && t.Start < regions[regionIndex].end {
			if t.Kind == "directive" {
				kind, name := macroDirective(t.Text)
				if kind == "define" || kind == "undef" {
					p.macros[name] = inlineMacro{reason: "Macro depends on an unresolved preprocessor choice."}
				}
				// A conditional include could redefine any known macro. Do not guess.
				if kind == "include" {
					for name := range p.macros {
						p.macros[name] = inlineMacro{reason: "Macro visibility depends on an unresolved conditional include."}
					}
				}
			}
			i++
			continue
		}
		if t.Kind == "directive" {
			kind, name := macroDirective(t.Text)
			switch kind {
			case "define":
				delete(p.macros, name)
				if m, ok := parseInlineMacro(t.Text); ok {
					p.macros[name] = m
				}
			case "undef":
				delete(p.macros, name)
			case "pragma":
				if strings.TrimSpace(name) == "once" {
					p.once[absolute] = true
				}
			case "include":
				target := p.resolveInclude(path, t.Text)
				if target != "" {
					data, readErr := os.ReadFile(target)
					if readErr == nil {
						_, _, readErr = p.process(target, string(data), false)
					}
					if readErr != nil && emit {
						p.review(path, "include", fmt.Sprintf("Cannot load macros from %s: %v", target, readErr), t.Start, t.End)
					}
				} else if emit && strings.Contains(t.Text, "\"") {
					p.review(path, "include", "Local include not resolved for macro expansion; supply Options.IncludePaths when using the library.", t.Start, t.End)
				}
			}
			i++
			continue
		}
		if !emit {
			i++
			continue
		}
		// Expand an ordinary span with the macro environment at this point.
		j := i + 1
		for j < len(ts) && ts[j].Kind != "directive" && (regionIndex >= len(regions) || ts[j].Start < regions[regionIndex].start) {
			j++
		}
		end := ts[j-1].End
		_, changes, notes := p.expand(selected[t.Start:end], map[string]bool{}, 0)
		for _, e := range changes {
			e.Start += t.Start
			e.End += t.Start
			edits = append(edits, e)
		}
		for _, n := range notes {
			p.review(path, n.name, n.message, t.Start+n.start, t.Start+n.end)
		}
		i = j
	}
	// Apply to original bytes: configured branches are selected by the normal
	// pipeline, which already preserves their coordinates and original source.
	output, err := ApplyEdits(source, edits)
	return output, edits, err
}

func macroDirective(text string) (kind, name string) {
	m := directiveParts.FindStringSubmatch(text)
	if m == nil {
		return
	}
	kind = m[1]
	rest := strings.TrimSpace(m[2])
	for _, c := range rest {
		if !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			break
		}
		name += string(c)
	}
	return
}
func parseInlineMacro(text string) (inlineMacro, bool) {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\\\r\n", ""), "\\\n", "")
	match := functionMacro.FindStringSubmatch(text)
	if match == nil {
		return inlineMacro{}, false
	}
	m := inlineMacro{body: strings.TrimRight(match[3], "\r\n")}
	if strings.TrimSpace(match[2]) != "" {
		m.params = strings.Split(match[2], ",")
	}
	seen := map[string]bool{}
	for i, param := range m.params {
		param = strings.TrimSpace(param)
		m.params[i] = param
		if !identifier.MatchString(param) || seen[param] {
			m.reason = "Variadic or invalid macro parameters require manual expansion."
		}
		seen[param] = true
	}
	tokens, err := Lex(m.body)
	if err != nil {
		m.reason = "Macro replacement could not be tokenized."
	}
	for _, t := range tokens {
		if t.Text == "#" || t.Text == "##" || t.Kind == "directive" {
			m.reason = "Stringification and token pasting require manual expansion."
		}
	}
	return m, true
}
func (p *macroInliner) resolveInclude(path, text string) string {
	m := directiveParts.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	rest := strings.TrimSpace(m[2])
	if len(rest) < 2 || (rest[0] != '"' && rest[0] != '<') {
		return ""
	}
	end := byte('"')
	if rest[0] == '<' {
		end = '>'
	}
	stop := strings.IndexByte(rest[1:], end)
	if stop < 0 {
		return ""
	}
	name := rest[1 : stop+1]
	dirs := append([]string{filepath.Dir(path)}, p.opts.IncludePaths...)
	for _, dir := range dirs {
		candidate := filepath.Join(dir, filepath.FromSlash(name))
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	return ""
}

type inlineNote struct {
	name, message string
	start, end    int
}

func (p *macroInliner) expand(source string, disabled map[string]bool, depth int) (string, []Edit, []inlineNote) {
	ts, err := Lex(source)
	if err != nil {
		return source, nil, nil
	}
	var edits []Edit
	var notes []inlineNote
	next := func(i int) int {
		for i < len(ts) && (ts[i].Kind == "space" || ts[i].Kind == "comment") {
			i++
		}
		return i
	}
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		if t.Kind != "word" {
			continue
		}
		m, exists := p.macros[t.Text]
		if !exists {
			continue
		}
		open := next(i + 1)
		if open >= len(ts) || ts[open].Text != "(" {
			continue
		}
		close, args, ok := macroArguments(source, ts, open)
		reason := m.reason
		if !ok {
			reason = "Unbalanced macro invocation retained."
			close = open
		}
		if disabled[t.Text] {
			reason = "Recursive macro invocation retained."
		}
		if depth >= 64 || p.steps >= 100000 {
			reason = "Macro expansion nesting or work limit reached."
		}
		if ok && len(m.params) == 0 && len(args) == 1 && strings.TrimSpace(args[0]) == "" {
			args = nil
		}
		if reason == "" && len(args) != len(m.params) {
			reason = "Macro argument count does not match its definition."
		}
		if reason != "" {
			notes = append(notes, inlineNote{t.Text, reason, t.Start, ts[close].End})
			i = close
			continue
		}
		p.steps++
		values := map[string]string{}
		var nestedNotes []inlineNote
		for j, arg := range args {
			value, _, ns := p.expand(arg, disabled, depth+1)
			values[m.params[j]] = value
			nestedNotes = append(nestedNotes, ns...)
		}
		replacementTokens, _ := Lex(m.body)
		var substitutions []Edit
		for _, tok := range replacementTokens {
			if tok.Kind == "word" {
				if value, ok := values[tok.Text]; ok {
					substitutions = append(substitutions, Edit{tok.Start, tok.End, value})
				}
			}
		}
		estimated := len(m.body)
		for _, e := range substitutions {
			estimated += len(e.Text) - (e.End - e.Start)
		}
		if estimated > macroExpansionLimit {
			notes = append(notes, inlineNote{t.Text, "Macro expansion size limit reached; invocation retained.", t.Start, ts[close].End})
			i = close
			continue
		}
		value, _ := ApplyEdits(m.body, substitutions)
		blocked := map[string]bool{}
		for name, v := range disabled {
			blocked[name] = v
		}
		blocked[t.Text] = true
		value, _, ns := p.expand(value, blocked, depth+1)
		nestedNotes = append(nestedNotes, ns...)
		// Preserve comments between the macro name and '(' and in unused arguments.
		// Do not inject new parentheses: the C replacement defines precedence.
		original := source[t.Start:ts[close].End]
		originalTokens, _ := Lex(original)
		for _, comment := range originalTokens {
			if comment.Kind == "comment" && !strings.Contains(value, comment.Text) {
				value = comment.Text + "\n" + value
			}
		}
		if len(value) > macroExpansionLimit || p.expandedBytes+len(value) > 8*macroExpansionLimit {
			notes = append(notes, inlineNote{t.Text, "Macro expansion size limit reached; invocation retained.", t.Start, ts[close].End})
			i = close
			continue
		}
		// A trailing // in a definition must not swallow the caller's semicolon.
		lastTokens, _ := Lex(value)
		for k := len(lastTokens) - 1; k >= 0; k-- {
			if lastTokens[k].Kind == "space" {
				continue
			}
			if lastTokens[k].Kind == "comment" && strings.HasPrefix(lastTokens[k].Text, "//") {
				value += "\n"
			}
			break
		}
		if p.annotate {
			value = "/*__starting_point_inline_macro:" + t.Text + "__*/" + value
		}
		p.expandedBytes += len(value)
		edits = append(edits, Edit{t.Start, ts[close].End, value})
		for _, n := range nestedNotes {
			notes = append(notes, inlineNote{n.name, n.message, t.Start, ts[close].End})
		}
		i = close
	}
	output, err := ApplyEdits(source, edits)
	if err != nil {
		return source, nil, notes
	}
	return output, edits, notes
}
func macroArguments(source string, ts []Token, open int) (int, []string, bool) {
	stack := []string{")"}
	start := ts[open].End
	var args []string
	for i := open + 1; i < len(ts); i++ {
		if ts[i].Kind != "punct" {
			continue
		}
		switch ts[i].Text {
		case "(":
			stack = append(stack, ")")
		case "[":
			stack = append(stack, "]")
		case "{":
			stack = append(stack, "}")
		case ")", "]", "}":
			if len(stack) == 0 || stack[len(stack)-1] != ts[i].Text {
				return i, nil, false
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				args = append(args, strings.TrimSpace(source[start:ts[i].Start]))
				return i, args, true
			}
		case ",":
			if len(stack) == 1 {
				args = append(args, strings.TrimSpace(source[start:ts[i].Start]))
				start = ts[i].End
			}
		}
	}
	return open, nil, false
}
