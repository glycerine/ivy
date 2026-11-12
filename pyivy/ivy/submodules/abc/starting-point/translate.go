// Package startingpoint produces source-preserving drafts, not verified translations.
package startingpoint

import (
	"crypto/sha256"
	"fmt"
	"go/token"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type Edit struct {
	Start, End int
	Text       string
}

// ApplyEdits rejects conflicts instead of depending on substitution order.
func ApplyEdits(s string, edits []Edit) (string, error) {
	e := append([]Edit(nil), edits...)
	sort.SliceStable(e, func(i, j int) bool { return e[i].Start < e[j].Start })
	var b strings.Builder
	pos := 0
	for _, x := range e {
		if x.Start < pos || x.End < x.Start || x.End > len(s) {
			return "", fmt.Errorf("invalid or overlapping edit: %+v", x)
		}
		b.WriteString(s[pos:x.Start])
		b.WriteString(x.Text)
		pos = x.End
	}
	b.WriteString(s[pos:])
	return b.String(), nil
}

type Function struct {
	Name                  string
	Start, End, BodyStart int
	Original, Draft       string
}
type Issue struct {
	ID, Category, Symbol, Status, Message string
	Start, End                            int
}
type Result struct {
	InlineMacros  bool             `json:",omitempty"`
	Configuration map[string]int64 `json:",omitempty"`
	Schema        int
	Path, SHA256  string
	Library       []LibraryChange     `json:",omitempty"`
	Conditionals  []ConditionalReview `json:",omitempty"`
	Functions     []Function
	Issues        []Issue
	Draft         string
}

var signature = regexp.MustCompile(`\b([A-Za-z_][A-Za-z_0-9 \t*]*?)\s*\b([A-Za-z_]\w*)\s*\(([^{};]*)\)\s*\{`)
var osFileReference = regexp.MustCompile(`\bos\.File\b`)
var declaration = regexp.MustCompile(`\b([A-Za-z_]\w*)(\s*\*+|\s+)\s*([A-Za-z_]\w*)\s*(=|;)`)
var namedType = regexp.MustCompile(`^\s*([A-Za-z_]\w*(?:::[A-Za-z_]\w*)*)\s*(\**&?)\s*([A-Za-z_]\w*)((?:\s*\[[^\[\]]*\])*)\s*$`)

func goType(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "::", ".")
	if strings.HasSuffix(s, "&") {
		return "*" + goType(strings.TrimSuffix(s, "&"))
	}
	if strings.HasSuffix(s, "*") {
		base := strings.TrimSpace(strings.TrimSuffix(s, "*"))
		if base == "FILE" {
			return "*os.File"
		}
		if base == "char" {
			return "string"
		}
		if base == "void" {
			return "any"
		}
		return "*" + goType(base)
	}
	switch s {
	case "clock_t", "time_t":
		return "int64"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "char":
		return "byte"
	case "unsigned":
		return "uint32"
	case "int":
		return "int32"
	case "short":
		return "int16"
	case "void":
		return ""
	case "sdbl_t":
		return "uint64"
	}
	return s
}

// Analyze currently recognizes ordinary C function definitions. Everything
// else remains verbatim in Draft and in the preserved original input. It does
// not evaluate preprocessor branches or claim to translate C++ yet.
func Analyze(path, source string) (Result, error) {
	return AnalyzeWithOptions(path, source, Options{})
}

// ResolvePackage resolves one package for a source or a batch's first source.
// Batch callers must reuse the result for every file in an output directory.
func ResolvePackage(path, override string) (string, error) {
	name := override
	if name == "" {
		dir, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return "", fmt.Errorf("resolve source directory: %w", err)
		}
		name = filepath.Base(dir)
		name = strings.Map(func(r rune) rune {
			if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return '_'
		}, name)
		if name == "" || name == "_" {
			name = "pkg"
		}
		if !token.IsIdentifier(name) {
			if token.Lookup(name).IsKeyword() {
				name += "_"
			} else {
				name = "pkg_" + name
			}
		}
	}
	if !token.IsIdentifier(name) || name == "_" {
		return "", fmt.Errorf("invalid Go package name %q; supply an explicit package override", name)
	}
	return name, nil
}

func analyzeCore(path, source string, opts Options) (result Result, resultErr error) {
	sourceChecked := false
	defer func() {
		if sourceChecked && resultErr != nil {
			resultErr = fmt.Errorf("transformed draft for %s (not original-source coordinates): %w", path, resultErr)
		}
	}()
	discoveredTypes := discoverTypeNames(source)
	normalized, normalizeErr := normalizeEnumReferences(source)
	if normalizeErr != nil {
		return Result{}, sourceError(source, normalizeErr)
	}
	source = normalized
	r := Result{Configuration: opts.Defines, Schema: 2, Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(source)))}
	var err error
	opts.Package, err = ResolvePackage(path, opts.Package)
	if err != nil {
		return r, err
	}
	ts, err := Lex(source)
	if err != nil {
		return r, sourceError(source, err)
	}
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return r, sourceError(source, err)
	}
	sourceChecked = true
	mask := Mask(ts)
	ends := map[int]int{}
	for a, b := range pairs {
		ends[ts[a].Start] = ts[b].End
	}
	edits, structIssues := structEdits(source, ts, ends, path)
	edits = append(edits, headerGuardEdits(ts)...)
	r.Issues = append(r.Issues, structIssues...)
	r.Issues = append(r.Issues, mechanicalReviewIssues(path, source)...)
	for _, t := range ts {
		if t.Kind != "directive" {
			continue
		}
		name, draft, message, ok := macroDraft(t)
		if !ok {
			continue
		}
		edits = append(edits, Edit{t.Start, t.End, draft})
		r.Issues = append(r.Issues, Issue{path + "::macro::" + name, "macro-review", name, "open", message, t.Start, t.End})
	}
	depthAt := make(map[int]int)
	depth := 0
	for _, t := range ts {
		depthAt[t.Start] = depth
		if t.Kind == "punct" && t.Text == "{" {
			depth++
		}
		if t.Kind == "punct" && t.Text == "}" {
			depth--
		}
	}
	functionSignatures := signature.FindAllStringSubmatchIndex(mask, -1)
	var topLevelSignatures [][]int
	for _, loc := range functionSignatures {
		if depthAt[loc[0]] == 0 {
			topLevelSignatures = append(topLevelSignatures, loc)
		}
	}
	localTypes := declarationTypes(mask, topLevelSignatures)
	for name := range opts.TypeNames {
		localTypes[name] = true
	}
	for name := range discoveredTypes {
		localTypes[name] = true
	}
	globalEdits, err := packageVariableEdits(source, ts, localTypes, opts.opaque)
	if err != nil {
		return r, err
	}
	edits = append(edits, globalEdits...)
	needsOS := false
	for _, loc := range functionSignatures {
		start, open := loc[0], loc[1]-1
		end, ok := ends[open]
		if !ok || depthAt[start] != 0 || insideOpaque(start, opts.opaque) || insideOpaque(open, opts.opaque) || insideOpaque(end-1, opts.opaque) {
			continue
		}
		ret, name, args := mask[loc[2]:loc[3]], mask[loc[4]:loc[5]], mask[loc[6]:loc[7]]
		retWords := strings.Fields(ret)
		for len(retWords) > 0 && (retWords[0] == "static" || retWords[0] == "inline" || retWords[0] == "extern") {
			retWords = retWords[1:]
		}
		ret = strings.Join(retWords, " ")
		// Normalize qualifiers before parsing the signature, rather than only
		// removing them from the final draft after a failed parse.
		args, err = stripConstQualifiers(args)
		if err != nil {
			return r, err
		}
		ret, err = stripConstQualifiers(ret)
		if err != nil {
			return r, err
		}
		var params []string
		valid := true
		if strings.TrimSpace(args) != "" && strings.TrimSpace(args) != "void" {
			arguments, _ := splitExpressions(args)
			for _, arg := range arguments {
				param, ok := goParameter(arg)
				if !ok {
					valid = false
					break
				}
				params = append(params, param)
			}
		}
		// A comment inside a signature must not disappear through regeneration.
		for _, t := range ts {
			if t.Start >= start && t.End <= open && t.Kind == "comment" {
				valid = false
			}
		}
		body := source[open:end]
		// Directives and local structs inside a function must be composed with
		// its body rewrite, not applied again as overlapping whole-file edits.
		var bodyEdits, outerEdits []Edit
		for _, edit := range edits {
			if edit.Start >= open && edit.End <= end {
				bodyEdits = append(bodyEdits, Edit{edit.Start - open, edit.End - open, edit.Text})
			} else {
				outerEdits = append(outerEdits, edit)
			}
		}
		body, err = ApplyEdits(body, bodyEdits)
		if err != nil {
			return r, err
		}
		edits = outerEdits
		converted, err := rewriteBody(body, localTypes)
		if err != nil {
			return r, err
		}
		resultType := goType(ret)
		if name == "main" && valid && (len(params) == 0 || len(params) == 2) {
			var usesOS bool
			converted, usesOS, err = rewriteMainBody(converted, params)
			if err != nil {
				return r, err
			}
			needsOS = needsOS || usesOS
			params = nil
			resultType = ""
		}
		resultSuffix := ""
		if resultType != "" {
			resultSuffix = " " + resultType
		}
		draft := "func " + name + "(" + strings.Join(params, ", ") + ")" + resultSuffix + " " + converted
		if strings.Contains(strings.Join(params, " "), " any") || resultType == "any" {
			r.Issues = append(r.Issues, Issue{path + "::void-pointer::" + name, "void-pointer-review", name, "open", "C void pointers use any in the draft; recover concrete types and repair casts/dereferences from call sites.", start, end})
		}
		message := "Review C truth values, casts, control-flow braces, allocation, macros and numeric semantics; draft is not executable Go."
		if !valid {
			draft = source[start:open] + converted
			message = "Unsupported parameter declarator or signature comment; signature retained verbatim, with mechanical body transformations preserved."
		}
		draft, err = stripConstQualifiers(draft)
		if err != nil {
			return r, err
		}
		draft, err = SanitizeCommentUTF8(draft)
		if err != nil {
			return r, err
		}
		check, _ := Lex(draft)
		if strings.Contains(Mask(check), "?") {
			r.Issues = append(r.Issues, Issue{path + "::ternary::" + name, "ternary-review", name, "open", "Embedded conditional expression requires typed temporary or structural expression lowering; retained verbatim. Check C truth-value conversion in lowered conditions too.", start, end})
		}
		r.Functions = append(r.Functions, Function{name, start, end, open, source[start:end], draft})
		r.Issues = append(r.Issues, Issue{path + "::" + name, "semantic-review", name, "open", message, start, end})
		edits = append(edits, Edit{start, end, draft})
	}
	r.Issues = append(r.Issues, Issue{path + "::file", "file-review", "", "open", "Preprocessor configuration, types, macros and declarations outside recognized functions are preserved verbatim and require review.", 0, len(source)})
	r.Draft, err = ApplyEdits(source, edits)
	if err == nil {
		r.Draft, err = stripConstQualifiers(r.Draft)
		if err != nil {
			return r, err
		}
		r.Draft, err = mechanicalDraft(r.Draft)
		if err != nil {
			return r, err
		}
		r.Draft, err = lowerEnums(r.Draft)
		if err != nil {
			return r, err
		}
		r.Draft, err = lowerFloatCasts(r.Draft)
		if err != nil {
			return r, err
		}
		r.Draft, err = lowerIntegerSuffixes(r.Draft)
		if err != nil {
			return r, err
		}
		r.Draft, err = integerDefines(r.Draft)
		if err != nil {
			return r, err
		}
		r.Draft, err = commentScaffolding(r.Draft)
		if err == nil {
			r.Draft, err = SanitizeCommentUTF8(r.Draft)
		}
		if err == nil {
			r.Draft = "package " + opts.Package + "\n\n" + r.Draft
			draftTokens, lexErr := Lex(r.Draft)
			if lexErr != nil {
				return r, lexErr
			}
			if needsOS || osFileReference.MatchString(Mask(draftTokens)) {
				r.Draft, err = ensureOSImport(r.Draft)
			}
		}
	}
	if err == nil {
		for i := range r.Functions {
			r.Functions[i].Draft, err = mechanicalDraft(r.Functions[i].Draft)
			if err != nil {
				return r, err
			}
			r.Functions[i].Draft, err = lowerEnums(r.Functions[i].Draft)
			if err != nil {
				return r, err
			}
			r.Functions[i].Draft, err = lowerFloatCasts(r.Functions[i].Draft)
			if err != nil {
				return r, err
			}
			r.Functions[i].Draft, err = lowerIntegerSuffixes(r.Functions[i].Draft)
			if err != nil {
				return r, err
			}
			r.Functions[i].Draft, err = integerDefines(r.Functions[i].Draft)
			if err != nil {
				return r, err
			}
		}
	}
	return r, err
}

func rewriteBody(body string, localTypes map[string]bool) (string, error) {
	body, err := lowerLocalDeclarations(body, localTypes)
	if err != nil {
		return "", err
	}
	ts, err := Lex(body)
	if err != nil {
		return "", err
	}
	mask := Mask(ts)
	var edits []Edit
	covered := map[int]bool{}
	statementStart := map[int]bool{}
	labelEnds := statementLabelEnds(ts)
	afterLabel := false
	previous, beforePrevious := "", ""
	for _, tok := range ts {
		if tok.Kind == "space" || tok.Kind == "comment" {
			continue
		}
		statementStart[tok.Start] = afterLabel || previous == "" || previous == "{" || previous == "}" || previous == ";" || (previous == "(" && beforePrevious == "for")
		afterLabel = labelEnds[tok.End]
		beforePrevious, previous = previous, tok.Text
	}
	for _, loc := range declaration.FindAllStringSubmatchIndex(mask, -1) {
		// A spelling alone is insufficient: a * b can be multiplication.
		if !statementStart[loc[0]] {
			continue
		}
		typeName := mask[loc[2]:loc[3]]
		if !localTypes[typeName] && !strings.HasSuffix(typeName, "_t") {
			continue
		}
		// Leave expression contexts and ambiguous declarations to the repair pass.
		typ := goType(mask[loc[2]:loc[3]] + mask[loc[4]:loc[5]])
		name := mask[loc[6]:loc[7]]
		op := mask[loc[8]:loc[9]]
		end := loc[7]
		text := "var " + name + " " + typ
		if op == "=" {
			text += " "
		}
		// Do not consume comments between declarator tokens.
		if strings.Contains(body[loc[0]:end], "/*") || strings.Contains(body[loc[0]:end], "//") {
			continue
		}
		edits = append(edits, Edit{loc[0], end, text})
		for i := loc[0]; i < end; i++ {
			covered[i] = true
		}
	}
	for _, t := range ts {
		if covered[t.Start] || t.Kind == "comment" || t.Kind == "string" || t.Kind == "directive" {
			continue
		}
		if t.Text == "->" {
			edits = append(edits, Edit{t.Start, t.End, "."})
		}
		if t.Text == "NULL" {
			edits = append(edits, Edit{t.Start, t.End, "nil"})
		}
	}
	converted, err := ApplyEdits(body, edits)
	if err != nil {
		return "", err
	}
	converted, err = emptyForBodies(converted, false)
	if err != nil {
		return "", err
	}
	converted, err = braceStatements(converted)
	if err != nil {
		return "", err
	}
	converted, err = lowerConditionals(converted)
	if err != nil {
		return "", err
	}
	return unwrapControlHeaders(converted)
}
