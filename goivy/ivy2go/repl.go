package ivy2go

import (
	"fmt"
	"sort"

	"github.com/glycerine/ivy/goivy"
)

// repl.go: emits a small REPL into the generated package's repl.go
// file (when target=repl or test). Mirrors the role of ivy2cpp/repl.go
// but the C++ stream-parsing layer collapses to bufio.Scanner +
// strings.Fields in Go (per ARCHITECTURE_TODO.md §3.5.8).
//
// M7 ships the minimum REPL: read lines, split into tokens, dispatch
// the first token to an action method, parse the rest as arguments
// per the action's formal parameter sorts.

// emitReplLoop writes the runRepl() function plus per-action wrappers.
func (g *Generator) emitReplLoop(w *goWriter) {
	if g == nil || g.Mod == nil {
		return
	}
	g.Ctx.AddImport("repl", "bufio", "")
	g.Ctx.AddImport("repl", "fmt", "")
	g.Ctx.AddImport("repl", "io", "")
	g.Ctx.AddImport("repl", "strings", "")
	g.Ctx.AddImport("repl", "strconv", "")

	// runRepl reads commands from r, dispatches to State methods,
	// and writes responses to w. EOF returns nil; a parse error
	// prints a diagnostic and continues the loop.
	w.linef("// runRepl is the REPL loop. Reads `action_name arg1 arg2 …`")
	w.linef("// per line; dispatches to %s methods.", g.StateTypeName)
	w.open(fmt.Sprintf("func runRepl(state *%s, in io.Reader, out io.Writer) error {", g.StateTypeName))
	w.line("scanner := bufio.NewScanner(in)")
	w.open("for scanner.Scan() {")
	w.line("line := strings.TrimSpace(scanner.Text())")
	w.line(`if line == "" { continue }`)
	w.open(`if line == "exit" || line == "quit" {`)
	w.line("return nil")
	w.close("")
	w.line("tokens := strings.Fields(line)")
	w.line(`if err := dispatchReplCommand(state, tokens, out); err != nil {`)
	w.linef("\tfmt.Fprintf(out, \"error: %%s\\n\", err)")
	w.line("}")
	w.close("")
	w.line("return scanner.Err()")
	w.close("")
	w.blank()

	// Per-action dispatch chain.
	g.emitReplDispatch(w)

	// Per-sort arg parsers.
	g.emitReplArgParsers(w)

	// Suppress strconv unused-import in degenerate cases.
	w.line("var _ = strconv.Atoi")
}

// emitReplDispatch writes a single dispatcher that switches on the
// action name and parses its args. Mirrors ivy2cpp/repl.go
// emitCmdReaderDispatchChain.
func (g *Generator) emitReplDispatch(w *goWriter) {
	w.open(fmt.Sprintf("func dispatchReplCommand(state *%s, tokens []string, out io.Writer) error {", g.StateTypeName))
	w.open("if len(tokens) == 0 {")
	w.line("return nil")
	w.close("")
	w.line("name := tokens[0]")
	w.line("args := tokens[1:]")
	w.line("switch name {")

	for _, name := range g.replActionNames() {
		act, _ := g.Mod.Actions.Get2(name)
		if act == nil {
			continue
		}
		params := act.GetFormalParams()
		w.linef("case %q:", name)
		w.linef("\tif len(args) != %d {", len(params))
		w.linef("\t\treturn fmt.Errorf(%q+\" expects %d args, got %%d\", len(args))", name, len(params))
		w.line("\t}")
		// Parse each arg.
		callArgs := make([]string, len(params))
		for i, p := range params {
			parser := g.replParserName(p.CSort)
			w.linef("\tv%d, err := %s(args[%d])", i, parser, i)
			w.line("\tif err != nil { return err }")
			callArgs[i] = fmt.Sprintf("v%d", i)
		}
		methodName := goExportedName(name)
		returns := act.GetFormalReturns()
		switch len(returns) {
		case 0:
			w.linef("\tstate.%s(%s)", methodName, joinComma(callArgs))
		case 1:
			w.linef("\tret := state.%s(%s)", methodName, joinComma(callArgs))
			w.linef(`	fmt.Fprintln(out, ret)`)
		default:
			// Multi-return: just print them.
			retNames := make([]string, len(returns))
			for i := range returns {
				retNames[i] = fmt.Sprintf("r%d", i)
			}
			w.linef("\t%s := state.%s(%s)", joinComma(retNames), methodName, joinComma(callArgs))
			w.linef(`	fmt.Fprintln(out, %s)`, joinComma(retNames))
		}
	}
	w.line("default:")
	w.line(`	return fmt.Errorf("unknown action %q", name)`)
	w.line("}")
	w.line("return nil")
	w.close("")
	w.blank()
}

// emitReplArgParsers emits one parser function per *distinct* sort
// referenced by an action parameter. Per-sort specialisation keeps
// the parsers monomorphic (D6: no generics) and small.
func (g *Generator) emitReplArgParsers(w *goWriter) {
	seen := map[string]bool{}
	for _, name := range g.replActionNames() {
		act, _ := g.Mod.Actions.Get2(name)
		if act == nil {
			continue
		}
		for _, p := range act.GetFormalParams() {
			if p == nil {
				continue
			}
			parser := g.replParserName(p.CSort)
			if seen[parser] {
				continue
			}
			seen[parser] = true
			g.emitOneReplArgParser(w, parser, p.CSort)
		}
	}
}

// replParserName returns the name of the per-sort arg parser. Stable
// across emissions so the dispatch and parsers stay in sync.
func (g *Generator) replParserName(s goivy.Sort) string {
	typ := g.goType(s)
	return "parseArg_" + iteHelperSuffix(typ)
}

// emitOneReplArgParser emits one parser. Bool gets "true"/"false";
// enums match by member name; integer-backed types use strconv.Atoi.
func (g *Generator) emitOneReplArgParser(w *goWriter, parser string, s goivy.Sort) {
	typeName := g.goType(s)
	w.linef("func %s(token string) (%s, error) {", parser, typeName)
	switch st := s.(type) {
	case *goivy.BooleanSort:
		w.line(`	switch token {`)
		w.line(`	case "true": return true, nil`)
		w.line(`	case "false": return false, nil`)
		w.line(`	}`)
		w.line(`	return false, fmt.Errorf("expected true/false, got %q", token)`)
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			w.line(`	n, err := strconv.Atoi(token); return n, err`)
		} else {
			w.line(`	switch token {`)
			for _, v := range st.Extension {
				w.linef(`	case %q: return %s, nil`, v, goExportedName(v))
			}
			w.line(`	}`)
			w.linef(`	return 0, fmt.Errorf("expected %s value, got %%q", token)`, goExportedName(st.Name))
		}
	default:
		if goIsAnyIntegerType(g, s) {
			// Integer-backed (range, uninterp, BV ≤ 64): parse
			// via Atoi and cast.
			w.linef(`	n, err := strconv.Atoi(token)`)
			w.linef(`	if err != nil { return %s, err }`, g.goZeroValue(s))
			w.linef(`	return %s(n), nil`, typeName)
		} else {
			// Struct types (destructor records, variants) can't be
			// parsed from a single token; return zero value with a
			// note. Real struct parsing is a future enhancement.
			w.linef(`	var z %s`, typeName)
			w.line(`	_ = token`)
			w.line(`	return z, nil`)
		}
	}
	w.line("}")
	w.blank()
}

// replActionNames returns the action names eligible for REPL
// dispatch, in deterministic order. M7 keeps all named actions; M9
// can prune internal/ext: ones.
func (g *Generator) replActionNames() []string {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil {
		return nil
	}
	names := make([]string, 0)
	for name := range g.Mod.Actions.All() {
		// Skip internal actions starting with "ext:" or "imp:" so
		// the REPL surfaces only user-callable actions (mirrors
		// ivy2cpp/repl.go's filtering).
		if hasPrefixAny(name, "ext:", "imp:", "ivy:", "__") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func hasPrefixAny(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if len(s) >= len(p) && s[:len(p)] == p {
			return true
		}
	}
	return false
}

// joinComma joins parts with ", ". A small helper to keep emitted
// lines compact.
func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
