package startingpoint

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// ConditionalReview identifies an original directive needing a branch decision.
type ConditionalReview struct {
	Location  SourceLocation
	Directive string
}

type opaqueRegion struct {
	start, end                   int
	projectedStart, projectedEnd int
	begin, finish, original      string
	effect                       string
}
type conditionalNode struct {
	start, end int
	branches   [][]Token
	otherwise  bool
}

// AnalyzeWithOptions selects explicitly configured preprocessing choices and
// isolates unresolved choices without guessing their values.
// A delimiter-only projection is used for parsing; original islands are restored
// after all rewriting, and all reported offsets are mapped to original bytes.
func AnalyzeWithOptions(path, source string, opts Options) (Result, error) {
	if !opts.inlinePrepared {
		return analyzeInlineMacros(path, source, opts)
	}
	originalSource := source
	selected, err := selectConfiguredBranches(source, opts.Defines)
	if err != nil {
		return Result{}, sourceError(source, err)
	}
	source = selected
	projected, regions, err := projectConditionals(source)
	if err != nil {
		return Result{}, sourceError(source, err)
	}
	opts.opaque = regions
	r, err := analyzeCore(path, projected, opts)
	r.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(originalSource)))
	if err != nil {
		var diagnostic *SourceError
		if errors.As(err, &diagnostic) {
			err = sourceError(source, &syntaxError{diagnostic.Message, originalOffset(diagnostic.Location.ByteOffset, regions)})
		}
		return r, err
	}
	// Report every preserved conditional opener, including nested ones. Recovery
	// regions can start before #if; use directive offsets, not region offsets.
	originalTokens, _ := Lex(source)
	for _, token := range originalTokens {
		name := directiveName(token)
		if name != "if" && name != "ifdef" && name != "ifndef" {
			continue
		}
		for _, region := range regions {
			if token.Start >= region.start && token.End <= region.end {
				prefix := source[:token.Start]
				r.Conditionals = append(r.Conditionals, ConditionalReview{
					Location:  SourceLocation{Line: strings.Count(prefix, "\n") + 1, Column: token.Start - strings.LastIndex(prefix, "\n"), ByteOffset: token.Start},
					Directive: strings.TrimSpace(token.Text),
				})
				break
			}
		}
	}
	formatted, library, err := rewriteStandardLibrary(r.Draft, true)
	if err != nil {
		return r, transformedDiagnostic(path, "library-call rewrite", r.Draft, err)
	}
	r.Library = library
	r.Draft = formatted
	libraryContext, err := newLibraryEnvironment(r.Draft)
	if err != nil {
		return r, transformedDiagnostic(path, "library metadata", r.Draft, err)
	}
	r.Draft = restoreRegions(r.Draft, regions)
	r.Draft, err = commentUsingNamespaces(r.Draft)
	if err != nil {
		return r, err
	}
	r.Draft, err = commentPthreadIncludes(r.Draft)
	if err != nil {
		return r, err
	}
	for i := range r.Functions {
		f := &r.Functions[i]
		f.Start = originalOffset(f.Start, regions)
		f.End = originalOffset(f.End, regions)
		f.BodyStart = originalOffset(f.BodyStart, regions)
		f.Original = originalSource[f.Start:f.End]
		if emitted, ok := libraryContext.functions[f.Name]; ok {
			formatted = emitted
		} else {
			formatted, _, err = rewriteStandardLibraryContext(f.Draft, false, libraryContext)
		}
		if err != nil {
			return r, transformedDiagnostic(path, "library-call rewrite of function "+f.Name, f.Draft, err)
		}
		f.Draft = formatted
		f.Draft = restoreRegions(f.Draft, regions)
		f.Draft, err = commentUsingNamespaces(f.Draft)
		if err != nil {
			return r, err
		}
		f.Draft, err = commentPthreadIncludes(f.Draft)
		if err != nil {
			return r, err
		}
	}
	for i := range r.Issues {
		r.Issues[i].Start = originalOffset(r.Issues[i].Start, regions)
		r.Issues[i].End = originalOffset(r.Issues[i].End, regions)
	}
	for i, region := range regions {
		r.Issues = append(r.Issues, Issue{ID: fmt.Sprintf("%s::conditional::%d", path, i), Category: "conditional-review", Status: "open", Message: "Preprocessor region preserved without selecting a branch. Delimiter-only scaffolding was used during analysis; surrounding code may also be retained when branch structures disagree.", Start: region.start, End: region.end})
	}
	return r, nil
}

func directiveName(t Token) string {
	if t.Kind != "directive" {
		return ""
	}
	m := directiveParts.FindStringSubmatch(t.Text)
	if m == nil {
		return ""
	}
	return m[1]
}

// delimiterEffect is the external stack interface: unmatched closing delimiters
// followed by unmatched openings. All balanced internal syntax cancels out.
func delimiterEffect(tokens []Token) (string, bool) {
	stack := ""
	closes := ""
	for _, t := range tokens {
		if t.Kind != "punct" {
			continue
		}
		c := t.Text
		if strings.Contains("([{", c) {
			stack += c
			continue
		}
		if !strings.Contains(")]}", c) {
			continue
		}
		if len(stack) == 0 {
			closes += c
			continue
		}
		last := stack[len(stack)-1:]
		if strings.Index("([{", last) != strings.Index(")]}", c) {
			return "", false
		}
		stack = stack[:len(stack)-1]
	}
	return closes + stack, true
}
func effectTokens(effect string) []Token {
	var out []Token
	for _, c := range effect {
		out = append(out, Token{Kind: "punct", Text: string(c)})
	}
	return out
}

func projectConditionals(source string) (string, []opaqueRegion, error) {
	ts, err := Lex(source)
	if err != nil {
		return "", nil, err
	}
	guards := map[int]bool{}
	for _, e := range headerGuardEdits(ts) {
		guards[e.Start] = true
	}
	var stack []*conditionalNode
	var regions []opaqueRegion
	unsafeStart := -1
	// Track completed top-level boundaries for conservative recovery when branch
	// interfaces disagree. Preserve the current declaration and remaining suffix.
	topStack := ""
	boundary := 0
	for _, t := range ts {
		if guards[t.Start] {
			continue
		}
		kind := directiveName(t)
		switch kind {
		case "if", "ifdef", "ifndef":
			stack = append(stack, &conditionalNode{start: t.Start, branches: [][]Token{{}}})
			continue
		case "elif", "elseif", "else":
			if len(stack) == 0 {
				return "", nil, &syntaxError{"unmatched #" + kind, t.Start}
			}
			node := stack[len(stack)-1]
			node.branches = append(node.branches, []Token{})
			if kind == "else" {
				node.otherwise = true
			}
			continue
		case "endif":
			if len(stack) == 0 {
				return "", nil, &syntaxError{"unmatched #endif", t.Start}
			}
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			node.end = t.End
			common, ok := delimiterEffect(node.branches[0])
			for _, branch := range node.branches[1:] {
				effect, valid := delimiterEffect(branch)
				if !valid || effect != common {
					ok = false
				}
			}
			if !node.otherwise && common != "" {
				ok = false
			}
			if !ok {
				unsafeStart = boundary
				break
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				n := len(parent.branches) - 1
				parent.branches[n] = append(parent.branches[n], effectTokens(common)...)
			} else {
				regions = append(regions, opaqueRegion{start: node.start, end: node.end, effect: common})
				for _, synthetic := range effectTokens(common) {
					topStack = trackTop(topStack, synthetic)
				}
			}
			continue
		}
		if unsafeStart >= 0 {
			break
		}
		if len(stack) > 0 {
			node := stack[len(stack)-1]
			n := len(node.branches) - 1
			node.branches[n] = append(node.branches[n], t)
			continue
		}
		topStack = trackTop(topStack, t)
		if topStack == "" && (t.Text == ";" || t.Text == "}") {
			boundary = t.End
		}
	}
	if len(stack) > 0 && unsafeStart < 0 {
		unsafeStart = boundary
	}
	if unsafeStart >= 0 {
		// No branch guessing: retain the whole unresolved suffix. The prefix can
		// still be translated, and the report makes the larger preserved span clear.
		keep := regions[:0]
		for _, r := range regions {
			if r.end <= unsafeStart {
				keep = append(keep, r)
			}
		}
		regions = append(keep, opaqueRegion{start: unsafeStart, end: len(source)})
	}
	var out strings.Builder
	pos := 0
	nonce := fmt.Sprintf("%x", sha256.Sum256([]byte(source)))[:16]
	for i := range regions {
		r := &regions[i]
		r.original = source[r.start:r.end]
		r.begin = fmt.Sprintf("/*SP_OPAQUE_%s_%d_BEGIN*/", nonce, i)
		r.finish = fmt.Sprintf("/*SP_OPAQUE_%s_%d_END*/", nonce, i)
		out.WriteString(source[pos:r.start])
		r.projectedStart = out.Len()
		out.WriteString(r.begin)
		out.WriteString(r.effect)
		out.WriteString(r.finish)
		r.projectedEnd = out.Len()
		pos = r.end
	}
	out.WriteString(source[pos:])
	return out.String(), regions, nil
}
func trackTop(stack string, t Token) string {
	if t.Kind != "punct" {
		return stack
	}
	if strings.Contains("([{", t.Text) {
		return stack + t.Text
	}
	if strings.Contains(")]}", t.Text) && len(stack) > 0 {
		return stack[:len(stack)-1]
	}
	return stack
}
func restoreRegions(text string, regions []opaqueRegion) string {
	for _, r := range regions {
		start := strings.Index(text, r.begin)
		if start < 0 {
			continue
		}
		finish := strings.Index(text[start+len(r.begin):], r.finish)
		if finish < 0 {
			continue
		}
		end := start + len(r.begin) + finish + len(r.finish)
		text = text[:start] + r.original + text[end:]
	}
	return text
}
func originalOffset(offset int, regions []opaqueRegion) int {
	delta := 0
	for _, r := range regions {
		if offset < r.projectedStart {
			break
		}
		if offset < r.projectedEnd {
			return r.start
		}
		delta += (r.end - r.start) - (r.projectedEnd - r.projectedStart)
	}
	return offset + delta
}
func insideOpaque(offset int, regions []opaqueRegion) bool {
	for _, r := range regions {
		if offset >= r.projectedStart && offset < r.projectedEnd {
			return true
		}
	}
	return false
}
