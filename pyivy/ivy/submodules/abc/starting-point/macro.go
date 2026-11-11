package startingpoint

import (
	"regexp"
	"strings"
)

// MacroTypes optionally supplies reviewed Go types. Unspecified parameters use
// int placeholders. Replacement text is copied, not interpreted or rewritten.
type MacroTypes struct {
	Parameters []string
	Result     string
}
type Options struct {
	opaque  []opaqueRegion
	Package string
	Macros  map[string]MacroTypes
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
	result := ""
	if typed && types.Result != "" {
		result = " " + types.Result
	}
	note := ""
	if !typed && len(args) > 0 {
		note = " //TODO: argument types were guessed, go compiler should tell us the correction."
	}
	replacement := m[3]
	if converted, err := emptyForBodies(replacement, true); err == nil {
		replacement = converted
	}
	if converted, err := braceStatements(replacement); err == nil {
		replacement = converted
	}
	if converted, err := unwrapControlHeaders(replacement); err == nil {
		replacement = converted
	}
	draft = original + "func " + name + "(" + strings.Join(params, ", ") + ")" + result + " {" + note + "\n" + replacement + "\n}\n"
	message = "Mechanical macro function draft: replacement text copied after directive line splicing, with empty for-loop bodies and structural control-flow braces made explicit. Repair parameter/result types, C syntax, returns and call-site semantics in the second pass. Unspecified parameter types are int placeholders, not inferred types."
	return
}
