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
	Schema       int
	Path, SHA256 string
	Functions    []Function
	Issues       []Issue
	Draft        string
}

var signature = regexp.MustCompile(`\bstatic\s+inline\s+([A-Za-z_][A-Za-z_0-9 \t*]*?)\s*\b([A-Za-z_]\w*)\s*\(([^{};]*)\)\s*\{`)
var declaration = regexp.MustCompile(`\b(unsigned|int|sdbl_t|[A-Za-z_]\w*_t)\s+(\*?)\s*([A-Za-z_]\w*)\s*(=|;)`)
var namedType = regexp.MustCompile(`^\s*([A-Za-z_]\w*)\s*(\*?)\s*([A-Za-z_]\w*)\s*$`)

func goType(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "*") {
		base := strings.TrimSpace(strings.TrimSuffix(s, "*"))
		if base == "void" {
			return "any"
		}
		return "*" + goType(base)
	}
	switch s {
	case "unsigned":
		return "uint32"
	case "int":
		return "int32"
	case "void":
		return ""
	case "sdbl_t":
		return "uint64"
	}
	return s
}

// Analyze currently recognizes ordinary static-inline C functions. Everything
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

func AnalyzeWithOptions(path, source string, opts Options) (Result, error) {
	r := Result{Schema: 2, Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(source)))}
	var err error
	opts.Package, err = ResolvePackage(path, opts.Package)
	if err != nil {
		return r, err
	}
	ts, err := Lex(source)
	if err != nil {
		return r, err
	}
	pairs, err := MatchDelimiters(ts)
	if err != nil {
		return r, err
	}
	mask := Mask(ts)
	ends := map[int]int{}
	for a, b := range pairs {
		ends[ts[a].Start] = ts[b].End
	}
	edits, structIssues := structEdits(source, ts, ends, path)
	edits = append(edits, headerGuardEdits(ts)...)
	r.Issues = append(r.Issues, structIssues...)
	for _, t := range ts {
		if t.Kind != "directive" {
			continue
		}
		name, draft, message, ok := macroDraft(t, opts)
		if !ok {
			continue
		}
		edits = append(edits, Edit{t.Start, t.End, draft})
		r.Issues = append(r.Issues, Issue{path + "::macro::" + name, "macro-review", name, "open", message, t.Start, t.End})
	}
	for _, loc := range signature.FindAllStringSubmatchIndex(mask, -1) {
		start, open := loc[0], loc[1]-1
		end, ok := ends[open]
		if !ok {
			continue
		}
		ret, name, args := mask[loc[2]:loc[3]], mask[loc[4]:loc[5]], mask[loc[6]:loc[7]]
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
			for _, arg := range strings.Split(args, ",") {
				m := namedType.FindStringSubmatch(arg)
				if m == nil {
					valid = false
					break
				}
				params = append(params, m[3]+" "+goType(m[1]+m[2]))
			}
		}
		// A comment inside a signature must not disappear through regeneration.
		for _, t := range ts {
			if t.Start >= start && t.End <= open && t.Kind == "comment" {
				valid = false
			}
		}
		body := source[open:end]
		converted, err := rewriteBody(body)
		if err != nil {
			return r, err
		}
		resultType := goType(ret)
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
		r.Draft, err = commentScaffolding(r.Draft)
		if err == nil {
			r.Draft, err = SanitizeCommentUTF8(r.Draft)
		}
		if err == nil {
			r.Draft = "package " + opts.Package + "\n\n" + r.Draft
		}
	}
	return r, err
}

func rewriteBody(body string) (string, error) {
	ts, err := Lex(body)
	if err != nil {
		return "", err
	}
	mask := Mask(ts)
	var edits []Edit
	covered := map[int]bool{}
	for _, loc := range declaration.FindAllStringSubmatchIndex(mask, -1) {
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
	return unwrapForHeaders(converted)
}
