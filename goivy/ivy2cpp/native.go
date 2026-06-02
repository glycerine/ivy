package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Native code block handling. Ported from Python ivy_to_cpp.py:
//
//   native_split        (line 1378)  → splitNativeCode
//   native_type         (line 1385)  → tag field of nativeBlock
//   native_to_str       (line 1444)  → renderNativeTemplate
//   emit_native         (line 1456)  → emitClassMemberNatives
//   native_typeof       (line 1429)  → nativeTypeOf
//   native_z3name       (line 1436)  → nativeZ3Name
//   native_reference    (line 4032)  → nativeReference (incl. action thunks)
//   native_declaration  (line 1389)  → nativeReference fallthrough path
//
// Tag classification mirrors the validate-and-dispatch loop at
// ivy_to_cpp.py:2314-2328 plus the inline pass at 2395-2403.

type nativeTagKind int

const (
	nativeTagInvalid nativeTagKind = iota
	nativeTagHeader
	nativeTagMember
	nativeTagImpl
	nativeTagInit
	nativeTagInline
	nativeTagEncode
)

type nativeBlock struct {
	raw         goivy.Node    // original native node, for error context
	kind        nativeTagKind // classified tag
	encodedSort string        // sort name (encode tag only)
	rawTag      string        // tag text before classification (diagnostics)
	code        string        // body code (post-split, before antiquote)
	params      []goivy.Expr  // antiquote parameters
	loopVars    []goivy.Expr  // native label args; used by <<< init >>> loops
}

// classifyNativeTag maps the raw tag (already antiquote-substituted for
// the encode case) to a nativeTagKind. The first return is the kind;
// when kind == nativeTagEncode the second return is the encoded sort
// name (with "__" → "."). Mirrors Python ivy_to_cpp.py:2314-2326.
//
// The "once" tag is a Go-side alias for "header" that pre-existing
// tests depend on; Python does not emit "once" itself.
func classifyNativeTag(raw string) (nativeTagKind, string) {
	tag := strings.TrimSpace(raw)
	switch tag {
	case "":
		// native_split treats no-tag bodies as "member" already; reach
		// this only when the tag line is blank.
		return nativeTagMember, ""
	case "once":
		return nativeTagHeader, ""
	case "header":
		return nativeTagHeader, ""
	case "member":
		return nativeTagMember, ""
	case "impl":
		return nativeTagImpl, ""
	case "init":
		return nativeTagInit, ""
	case "inline":
		return nativeTagInline, ""
	}
	if strings.HasPrefix(tag, "encode") {
		rest := strings.TrimSpace(strings.TrimPrefix(tag, "encode"))
		// Python: tag = tag[6:].strip().replace('__','.')
		rest = strings.ReplaceAll(rest, "__", ".")
		return nativeTagEncode, rest
	}
	return nativeTagInvalid, ""
}

// buildNativeBlocks walks g.Mod.Natives once, splits the code into
// tag+body, classifies, and renders any antiquote substitutions needed
// to classify the tag (encode case). Errors during render attach to
// g.errs and skip the block.
func (g *Generator) buildNativeBlocks() ([]nativeBlock, error) {
	if g == nil || g.Mod == nil {
		return nil, nil
	}
	blocks := make([]nativeBlock, 0, len(g.Mod.Natives))
	for _, node := range g.Mod.Natives {
		args := node.Args()
		if len(args) < 2 {
			continue
		}
		codeNode, ok := args[1].(*goivy.NativeCode)
		if !ok {
			return nil, fmt.Errorf("ivy2cpp: native block has code %T", args[1])
		}
		rawTag, body := splitNativeCode(codeNode.Code)
		params := make([]goivy.Expr, 0, len(args)-2)
		for _, arg := range args[2:] {
			expr, err := nativeExpr(arg)
			if err != nil {
				return nil, err
			}
			params = append(params, expr)
		}
		loopVars := make([]goivy.Expr, 0)
		if args[0] != nil {
			label, err := nativeExpr(args[0])
			if err != nil {
				return nil, err
			}
			for _, arg := range label.Args() {
				expr, err := nativeExpr(arg)
				if err != nil {
					return nil, err
				}
				loopVars = append(loopVars, expr)
			}
		}
		// Python applies native_to_str(native, code=tag) before
		// classifying an encode tag, so antiquote refs in the tag
		// itself resolve. We do the same for ALL tags so that the
		// classifier sees the substituted form.
		renderedTag, err := g.renderNativeTemplate(rawTag, params)
		if err == nil {
			rawTag = renderedTag
		}
		kind, encoded := classifyNativeTag(rawTag)
		blocks = append(blocks, nativeBlock{
			raw:         node,
			kind:        kind,
			encodedSort: encoded,
			rawTag:      strings.TrimSpace(rawTag),
			code:        body,
			params:      params,
			loopVars:    loopVars,
		})
	}
	return blocks, nil
}

func splitNativeCode(code string) (string, string) {
	parts := strings.SplitN(code, "\n", 2)
	if len(parts) == 2 {
		tag := strings.TrimSpace(parts[0])
		if tag == "" {
			tag = "member"
		}
		return tag, parts[1]
	}
	return "member", code
}

func nativeExpr(node goivy.Node) (goivy.Expr, error) {
	switch n := node.(type) {
	case *goivy.CompiledNode:
		if expr, ok := n.Node.(goivy.Expr); ok {
			return expr, nil
		}
		if nested, ok := n.Node.(goivy.Node); ok {
			return nativeExpr(nested)
		}
		return nil, fmt.Errorf("ivy2cpp: native antiquote is compiled %T, not an expression", n.Node)
	case goivy.Expr:
		return n, nil
	default:
		return nil, fmt.Errorf("ivy2cpp: native antiquote %T is not an expression", node)
	}
}

// onceMemo returns the lazily-initialized dedup set for header/impl/
// inline/encode body emission. Python uses one `once_memo` across all
// four (ivy_to_cpp.py:1974).
func (g *Generator) onceMemo() map[string]bool {
	if g != nil && g.Ctx != nil {
		if g.Ctx.OnceGlobals == nil {
			g.Ctx.OnceGlobals = map[string]bool{}
		}
		return g.Ctx.OnceGlobals
	}
	if g.nativeOnceMemo == nil {
		g.nativeOnceMemo = map[string]bool{}
	}
	return g.nativeOnceMemo
}

func (g *Generator) encodedSet() map[string]bool {
	if g.encodedSorts == nil {
		g.encodedSorts = map[string]bool{}
	}
	return g.encodedSorts
}

// emitHeaderNatives emits `<<< header ... >>>` blocks (and the "once"
// alias) before the class declaration in the header file. Deduped via
// the shared once_memo.
func (g *Generator) emitHeaderNatives(w *cppWriter) error {
	blocks, err := g.buildNativeBlocks()
	if err != nil {
		return err
	}
	memo := g.onceMemo()
	for _, b := range blocks {
		if b.kind != nativeTagHeader {
			continue
		}
		rendered, err := g.renderNativeTemplate(b.code, b.params)
		if err != nil {
			return err
		}
		if memo[rendered] {
			continue
		}
		memo[rendered] = true
		emitNativeLines(w, rendered)
	}
	return nil
}

// emitClassMemberNatives emits `<<< member ... >>>` and untagged blocks
// inside the class declaration. NOT deduped (Python emit_native at line
// 1456 appends unconditionally). Also surfaces unknown-tag errors
// (Python IvyError at 2326), since this is the canonical validation
// pass in the Python code.
func (g *Generator) emitClassMemberNatives(w *cppWriter) error {
	blocks, err := g.buildNativeBlocks()
	if err != nil {
		return err
	}
	for _, b := range blocks {
		switch b.kind {
		case nativeTagMember:
			rendered, err := g.renderNativeTemplate(b.code, b.params)
			if err != nil {
				return err
			}
			emitNativeLines(w, rendered)
		case nativeTagInvalid:
			g.unsupported(w, "syntax error at token %s", b.rawTag)
		}
	}
	return nil
}

// emitImplNatives emits `<<< impl ... >>>` and `<<< encode <sort> ... >>>`
// blocks at file scope in the impl file, deduped via once_memo. Also
// validates that encoded sorts exist and are not duplicated (Python
// 2315-2323).
func (g *Generator) emitImplNatives(w *cppWriter) error {
	blocks, err := g.buildNativeBlocks()
	if err != nil {
		return err
	}
	memo := g.onceMemo()
	encoded := g.encodedSet()
	for _, b := range blocks {
		switch b.kind {
		case nativeTagImpl:
			// fall through
		case nativeTagEncode:
			name := b.encodedSort
			if _, ok := g.sortByName(name); !ok {
				g.unsupported(w, "%s is not a declared sort", name)
				continue
			}
			if encoded[name] {
				g.unsupported(w, "duplicate encoding for sort %s", name)
				continue
			}
			encoded[name] = true
		default:
			continue
		}
		rendered, err := g.renderNativeTemplateWithClass(b.code, b.params, g.ClassName)
		if err != nil {
			return err
		}
		if memo[rendered] {
			continue
		}
		memo[rendered] = true
		emitNativeLines(w, rendered)
	}
	return nil
}

// emitInitNatives emits `<<< init ... >>>` blocks inside the
// constructor body. NOT deduped (Python 2361-2370 always emits).
func (g *Generator) emitInitNatives(w *cppWriter) error {
	blocks, err := g.buildNativeBlocks()
	if err != nil {
		return err
	}
	for _, b := range blocks {
		if b.kind != nativeTagInit {
			continue
		}
		loops := 0
		loopFailed := false
		for _, v := range b.loopVars {
			header, err := g.loopHeaderForSort(v.NodeSort(), nativeLoopVarName(v))
			if err != nil {
				g.unsupported(w, "unsupported native initializer parameter %s:%s", nativeLoopVarName(v), err.Error())
				loopFailed = true
				break
			}
			w.open(header)
			loops++
		}
		if loopFailed {
			g.closeAssignmentLoops(w, loops)
			continue
		}
		rendered, err := g.renderNativeTemplate(b.code, b.params)
		if err != nil {
			return err
		}
		emitNativeLines(w, rendered)
		g.closeAssignmentLoops(w, loops)
	}
	return nil
}

// emitInlineNatives emits `<<< inline ... >>>` blocks in the header
// AFTER the class declaration's closing `};`. Deduped via the shared
// once_memo. Mirrors Python ivy_to_cpp.py:2395-2403.
func (g *Generator) emitInlineNatives(w *cppWriter) error {
	blocks, err := g.buildNativeBlocks()
	if err != nil {
		return err
	}
	memo := g.onceMemo()
	for _, b := range blocks {
		if b.kind != nativeTagInline {
			continue
		}
		rendered, err := g.renderNativeTemplateWithClass(b.code, b.params, g.ClassName)
		if err != nil {
			return err
		}
		if memo[rendered] {
			continue
		}
		memo[rendered] = true
		emitNativeLines(w, rendered)
	}
	return nil
}

func (g *Generator) renderNativeTemplate(code string, params []goivy.Expr) (string, error) {
	fields := strings.Split(code, "`")
	for i := 1; i < len(fields); i += 2 {
		idx, err := strconv.Atoi(fields[i])
		if err != nil {
			return "", fmt.Errorf("bad native antiquote index %q", fields[i])
		}
		if idx < 0 || idx >= len(params) {
			return "", fmt.Errorf("native antiquote index %d out of range", idx)
		}
		prev := fields[i-1]
		var repl string
		switch {
		case strings.HasSuffix(prev, "%"):
			repl, err = g.nativeTypeOf(params[idx])
		case strings.HasSuffix(prev, `"`):
			repl, err = g.nativeZ3Name(params[idx])
		default:
			repl, err = g.nativeReference(params[idx])
		}
		if err != nil {
			return "", err
		}
		fields[i] = repl
	}
	for i := 0; i < len(fields); i += 2 {
		if strings.HasSuffix(fields[i], "%") {
			fields[i] = strings.TrimSuffix(fields[i], "%")
		}
	}
	return strings.Join(fields, ""), nil
}

func (g *Generator) renderNativeTemplateWithClass(code string, params []goivy.Expr, className string) (string, error) {
	old := g.nativeClassName
	g.nativeClassName = className
	defer func() { g.nativeClassName = old }()
	return g.renderNativeTemplate(code, params)
}

func (g *Generator) nativeTypeFull(nt *goivy.NativeType) (string, error) {
	if nt == nil || len(nt.Elems) == 0 {
		return "", fmt.Errorf("ivy2cpp: empty native type")
	}
	code, err := nativeCodeText(nt.Elems[0])
	if err != nil {
		return "", err
	}
	fields := strings.Split(code, "`")
	for i := 1; i < len(fields); i += 2 {
		idx, err := strconv.Atoi(fields[i])
		if err != nil {
			return "", fmt.Errorf("bad native type antiquote index %q", fields[i])
		}
		if idx < 0 || idx+1 >= len(nt.Elems) {
			return "", fmt.Errorf("native type antiquote index %d out of range", idx)
		}
		repl, err := g.nativeReferenceInType(nt.Elems[idx+1])
		if err != nil {
			return "", err
		}
		fields[i] = repl
	}
	return strings.Join(fields, ""), nil
}

func nativeCodeText(node goivy.Node) (string, error) {
	switch n := node.(type) {
	case *goivy.NativeCode:
		return n.Code, nil
	case *goivy.Atom:
		return n.Rep, nil
	case *goivy.Symbol:
		return n.Rep, nil
	case *goivy.Const:
		return n.Name, nil
	default:
		return "", fmt.Errorf("ivy2cpp: native template has code %T", node)
	}
}

func (g *Generator) nativeReferenceInType(node goivy.Node) (string, error) {
	switch n := node.(type) {
	case *goivy.CompiledNode:
		if nested, ok := n.Node.(goivy.Node); ok {
			return g.nativeReferenceInType(nested)
		}
	case *goivy.Atom:
		if s, ok := g.sortByName(n.Rep); ok {
			return g.nativeCppType(s), nil
		}
		return varName(n.Rep), nil
	case *goivy.Symbol:
		if s, ok := g.sortByName(n.Rep); ok {
			return g.nativeCppType(s), nil
		}
		return varName(n.Rep), nil
	case *goivy.Const:
		if s, ok := g.sortByName(n.Name); ok {
			return g.nativeCppType(s), nil
		}
		return varName(n.Name), nil
	case *goivy.UninterpretedSort:
		return g.nativeCppType(n), nil
	case *goivy.LogicEnumeratedSort:
		return g.nativeCppType(n), nil
	case *goivy.RangeSort:
		return g.nativeCppType(n), nil
	}
	return "", fmt.Errorf("ivy2cpp: native type antiquote %T is not supported", node)
}

func (g *Generator) emitNativeExpr(n *goivy.LogicNativeExpr) (string, error) {
	if n == nil || len(n.CompiledChildren) == 0 {
		return "", fmt.Errorf("ivy2cpp: empty native expression")
	}
	code, err := nativeCodeText(n.CompiledChildren[0])
	if err != nil {
		return "", err
	}
	rendered, err := g.renderNativeTemplate(code, n.CompiledChildren[1:])
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(rendered), nil
}

func (g *Generator) nativeTypeForSort(name string) (*goivy.NativeType, bool) {
	if g == nil || g.Mod == nil || g.Mod.NativeTypes == nil {
		return nil, false
	}
	nt, ok := g.Mod.NativeTypes[name]
	return nt, ok && nt != nil
}

func (g *Generator) emitNativeTypeDecl(w *cppWriter, name string, nt *goivy.NativeType) {
	code, err := g.nativeTypeFull(nt)
	if err != nil {
		g.unsupported(w, "unsupported native type %s: %s", name, err.Error())
		return
	}
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "primitive ") {
		code = strings.TrimSpace(strings.TrimPrefix(code, "primitive "))
		w.linef("typedef %s %s;", code, varName(name))
		return
	}
	switch code {
	case "int", "bool":
		w.linef("typedef %s %s;", code, varName(name))
	case "std::vector<bool>":
		g.emitNativeClassTypeDecl(w, name, "std::vector<int>")
	default:
		g.emitNativeClassTypeDecl(w, name, code)
	}
}

func (g *Generator) emitNativeClassTypeDecl(w *cppWriter, name, base string) {
	w.open(fmt.Sprintf("class %s : public %s {", varName(name), base))
	w.line("public:")
	w.indent++
	w.linef("typedef %s ivy_native_base;", base)
	w.indent--
	w.close(";")
}

// isCallbackAction reports whether arg refers (via Relname) to an
// action in the current module. Used by both nativeReference (to emit
// a thunk constructor call) and the collectCallbackActions pass in
// native_thunk.go.
func (g *Generator) isCallbackAction(arg goivy.Node) (string, bool) {
	if g == nil {
		return "", false
	}
	return callbackActionName(g.Mod, arg)
}

func callbackActionName(mod *goivy.Module, arg goivy.Node) (string, bool) {
	if mod == nil || mod.Actions == nil {
		return "", false
	}
	rn, ok := arg.(interface{ Relname() string })
	if !ok {
		return "", false
	}
	name := rn.Relname()
	if _, ok := mod.Actions.Get2(name); !ok {
		return "", false
	}
	return name, true
}

// nativeArgName extracts a plain variable name from a child node of an
// action-reference Atom. Mirrors Python's
// `varname(arg.rep) for arg in atom.args` (ivy_to_cpp.py:4035).
func nativeArgName(child goivy.Node) string {
	switch c := child.(type) {
	case *goivy.Const:
		return varName(c.Name)
	case *goivy.LogicVariable:
		return varName(c.Name)
	case *goivy.Atom:
		return varName(c.Rep)
	case *goivy.Symbol:
		return varName(c.Rep)
	}
	if rn, ok := child.(interface{ Relname() string }); ok {
		return varName(rn.Relname())
	}
	return ""
}

func (g *Generator) nativeReference(arg goivy.Expr) (string, error) {
	// Action callback reference (Python native_reference,
	// ivy_to_cpp.py:4032-4036): atoms whose .rep is an action become
	// `thunk__name(this, argvar1, argvar2, ...)`.
	if name, ok := g.isCallbackAction(arg); ok {
		parts := []string{"this"}
		for _, child := range arg.Args() {
			n := nativeArgName(child)
			if n == "" {
				return "", fmt.Errorf("native callback %s has unsupported argument %T", name, child)
			}
			parts = append(parts, n)
		}
		return "thunk__" + varName(name) + "(" + strings.Join(parts, ", ") + ")", nil
	}
	switch a := arg.(type) {
	case *goivy.Const:
		if s, ok := g.sortByName(a.Name); ok {
			return g.nativeCppType(s), nil
		}
		return varName(a.Name), nil
	case *goivy.LogicVariable:
		return varName(a.Name), nil
	case *goivy.Apply:
		return g.emitExpr(a)
	case *goivy.UninterpretedSort:
		return g.nativeCppType(a), nil
	case *goivy.LogicEnumeratedSort:
		return g.nativeCppType(a), nil
	case *goivy.RangeSort:
		return g.nativeCppType(a), nil
	}
	if rn, ok := arg.(interface{ Relname() string }); ok {
		name := rn.Relname()
		if s, ok := g.sortByName(name); ok {
			return g.nativeCppType(s), nil
		}
		res := varName(name)
		for _, child := range arg.Args() {
			expr, ok := child.(goivy.Expr)
			if !ok {
				return "", fmt.Errorf("native reference %s has non-expression argument %T", name, child)
			}
			idx, err := g.nativeReference(expr)
			if err != nil {
				return "", err
			}
			res += "[" + idx + "]"
		}
		return res, nil
	}
	return g.emitExpr(arg)
}

func (g *Generator) nativeTypeOf(arg goivy.Expr) (string, error) {
	if name, ok := g.isCallbackAction(arg); ok {
		return "thunk__" + varName(name), nil
	}
	return g.nativeCppType(arg.NodeSort()), nil
}

func (g *Generator) nativeCppType(s goivy.Sort) string {
	if g == nil {
		return cppType(s)
	}
	if g != nil && g.nativeClassName != "" {
		return g.cppQualifiedType(s, g.nativeClassName)
	}
	return g.cppType(s)
}

func nativeLoopVarName(v goivy.Expr) string {
	switch x := v.(type) {
	case *goivy.Const:
		return varName(x.Name)
	case *goivy.LogicVariable:
		return varName(x.Name)
	}
	if rn, ok := v.(interface{ Relname() string }); ok {
		return varName(rn.Relname())
	}
	return varName(fmt.Sprint(v))
}

func (g *Generator) nativeZ3Name(arg goivy.Expr) (string, error) {
	if v, ok := arg.(*goivy.LogicVariable); ok {
		return sortName(v.VSort), nil
	}
	if rn, ok := arg.(interface{ Relname() string }); ok {
		return rn.Relname(), nil
	}
	if c, ok := arg.(*goivy.Const); ok {
		return c.Name, nil
	}
	return "", fmt.Errorf("cannot emit native z3 name for %T", arg)
}

func (g *Generator) sortByName(name string) (goivy.Sort, bool) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil, false
	}
	return g.Mod.Sig.Sorts.Get2(name)
}

func emitNativeLines(w *cppWriter, code string) {
	code = strings.TrimRight(code, " \t\r\n")
	lines := strings.Split(code, "\n")
	base := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := nativeIndent(line)
		if base == -1 || indent < base {
			base = indent
		}
	}
	if base < 0 {
		return
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			w.line("")
			continue
		}
		indent := nativeIndent(line) - base
		if indent < 0 {
			indent = 0
		}
		w.line(strings.Repeat(" ", indent) + strings.TrimSpace(line))
	}
}

func nativeIndent(line string) int {
	indent := 0
	for _, r := range line {
		switch r {
		case ' ':
			indent++
		case '\t':
			indent = ((indent + 8) / 8) * 8
		default:
			return indent
		}
	}
	return indent
}
