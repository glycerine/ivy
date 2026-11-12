package startingpoint

import (
	"regexp"
	"strings"
)

// MacroTypes optionally supplies reviewed Go types. Unspecified parameters use
// int placeholders. An unspecified result also uses int; expression macros
// return their value after redundant enclosing parentheses are removed.
type MacroTypes struct {
	Parameters []string
	Result     string
}
type Options struct {
	InlineMacros   bool     // Expand function-like macros at call sites instead of generating functions.
	IncludePaths   []string // Additional local header search directories for macro inlining.
	inlinePrepared bool
	TypeNames      map[string]bool  // Known typedef spellings from an input inventory.
	Defines        map[string]int64 // Explicit platform choices; zero means disabled.
	opaque         []opaqueRegion
	Package        string
	Macros         map[string]MacroTypes
}

var functionMacro = regexp.MustCompile(`^\s*#\s*define\s+([A-Za-z_]\w*)\(([^)]*)\)[ \t]*([\s\S]*)$`)
var identifier = regexp.MustCompile(`^[A-Za-z_]\w*$`)

func macroDraft(t Token, opts Options) (name, draft, message string, ok bool) {
	// Remove directive line continuations; otherwise retain replacement text.
	text := strings.ReplaceAll(strings.ReplaceAll(t.Text, "\\\r\n", ""), "\\\n", "")
	m := functionMacro.FindStringSubmatch(text)
	if m == nil {
		return
	}
	name, ok = m[1], true
	original := "// Original macro:\n// " + strings.ReplaceAll(t.Text, "\n", "\n// ") + "\n"
	if opts.InlineMacros {
		return name, original, "Function-like macro retained as a comment; inline mode expands supported visible invocations and reports retained calls separately.", true
	}
	var args []string
	if strings.TrimSpace(m[2]) != "" {
		args = strings.Split(m[2], ",")
	}
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
		if !identifier.MatchString(args[i]) {
			return name, original, "Variadic or unsupported macro parameters require explicit function signature repair.", true
		}
	}
	types, typed := opts.Macros[name]
	if typed && len(types.Parameters) != len(args) {
		return name, original, "Configured macro parameter count does not match definition.", true
	}
	params := make([]string, len(args))
	for i, arg := range args {
		typ := "int"
		if typed {
			typ = types.Parameters[i]
		}
		params[i] = arg + " " + typ
	}
	result := " int"
	if typed && types.Result != "" {
		result = " " + types.Result
	}
	note := ""
	if !typed && len(args) > 0 {
		note = " //TODO: argument types were guessed, go compiler should tell us the correction."
	}
	replacement := unwrapMacroExpression(m[3])
	valueMacro := macroHasValue(replacement)
	if !valueMacro {
		if converted, err := emptyForBodies(replacement, true); err == nil {
			replacement = converted
		}
		if converted, err := braceStatements(replacement); err == nil {
			replacement = converted
		}
		if converted, err := unwrapControlHeaders(replacement); err == nil {
			replacement = converted
		}
		replacement = strings.TrimSpace(replacement)
		if replacement != "" {
			replacement += "\n"
		}
		replacement += "return 0 //TODO: statement-style macro; review placeholder result."
	} else {
		// A C comma expression evaluates left-to-right and yields its final operand.
		if parts, multiple := splitExpressions(replacement); multiple {
			for i := range parts {
				parts[i] = unwrapMacroExpression(parts[i])
			}
			replacement = strings.Join(parts[:len(parts)-1], ";\n") + ";\n" + macroReturn(parts[len(parts)-1])
		} else {
			replacement = macroReturn(replacement)
		}
	}

	draft = original + "func " + name + "(" + strings.Join(params, ", ") + ")" + result + " {" + note + "\n" + replacement + "\n}\n"
	message = "Mechanical macro function draft: enclosing expression parentheses removed and values returned. Unspecified parameter and result types use int placeholders. Repair retained C casts, operators and call-site semantics in the second pass."
	if !valueMacro {
		message += " Statement-style macro has a placeholder return 0; recover its intended result and control flow."
	}
	return
}

// Remove only pairs enclosing the complete replacement, not C cast parentheses
// or grouping of one operand. Token edits preserve leading/trailing comments.
func unwrapMacroExpression(source string) string {
	source = strings.TrimSpace(source)
	for {
		ts, pairs, err := significantTokens(source)
		if err != nil || len(ts) == 0 {
			return source
		}
		if ts[len(ts)-1].Text == ";" {
			// A single trailing terminator is not part of an expression's value.
			if len(ts) > 1 && ts[0].Text == "(" && pairs[0] == len(ts)-2 {
				source = source[:ts[len(ts)-1].Start] + source[ts[len(ts)-1].End:]
				continue
			}
		}
		if ts[0].Text != "(" || pairs[0] != len(ts)-1 {
			return source
		}
		result, err := ApplyEdits(source, []Edit{{ts[0].Start, ts[0].End, ""}, {ts[len(ts)-1].Start, ts[len(ts)-1].End, ""}})
		if err != nil {
			return source
		}
		source = strings.TrimSpace(result)
	}
}

func macroHasValue(source string) bool {
	ts, pairs, err := significantTokens(source)
	if err != nil || len(ts) == 0 {
		return false
	}
	switch ts[0].Text {
	case "{", "for", "while", "do", "if", "switch", "return", "break", "continue", "goto", "typedef", "static", "var", "struct", "union", "enum", "const":
		return false
	case "int", "char", "short", "long", "unsigned", "signed", "float", "double", "bool":
		if len(ts) < 2 || ts[1].Text != "(" {
			return false
		}
	}
	for i := 0; i < len(ts); i++ {
		if end, ok := pairs[i]; ok && end > i {
			i = end
			continue
		}
		if ts[i].Text == ";" && i != len(ts)-1 {
			return false
		}
	}
	return true
}

func macroReturn(source string) string {
	ts, _, err := significantTokens(source)
	if err != nil || len(ts) == 0 {
		return "return " + source
	}
	// Put return after leading comments so a multiline comment cannot trigger
	// automatic semicolon insertion between return and its expression.
	return source[:ts[0].Start] + "return " + source[ts[0].Start:]
}
