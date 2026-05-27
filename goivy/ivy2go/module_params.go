package ivy2go

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) moduleParams() []*goivy.Const {
	if g == nil || g.Mod == nil || len(g.Mod.Params) == 0 {
		return nil
	}
	out := make([]*goivy.Const, 0, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		if p != nil {
			out = append(out, p)
		}
	}
	return out
}

func moduleParamLocalName(p *goivy.Const) string {
	if p == nil {
		return ""
	}
	return "p__" + goIdent(p.Name)
}

func (g *Generator) stateConstructorParamList() string {
	params := g.moduleParams()
	if len(params) == 0 {
		return ""
	}
	decls := make([]string, 0, len(params))
	for _, p := range params {
		decls = append(decls, fmt.Sprintf("%s %s", moduleParamLocalName(p), g.goType(p.CSort)))
	}
	return strings.Join(decls, ", ")
}

func (g *Generator) emitModuleParamSetup() []string {
	params := g.moduleParams()
	if len(params) == 0 {
		return nil
	}
	g.Ctx.AddImport("main", "fmt", "")
	g.Ctx.AddImport("main", "os", "")
	g.emitModuleParamHelpers()

	args := make([]string, 0, len(params))
	g.main.line("__ivyModuleArgs, err := ivyCollectModuleParamArgs(os.Args[1:], []ivyModuleParamSpec{")
	for i, p := range params {
		hasDefault := i < len(g.Mod.ParamDefaults) && g.Mod.ParamDefaults[i] != nil
		defaultText := ""
		if hasDefault {
			defaultText = paramDefaultText(g.Mod.ParamDefaults[i])
		}
		g.main.linef("\t{Name: %q, HasDefault: %t, Default: %q},", p.Name, hasDefault, defaultText)
	}
	g.main.line("})")
	g.main.line("if err != nil {")
	g.main.line("\tfmt.Fprintln(os.Stderr, err)")
	g.main.line("\tos.Exit(2)")
	g.main.line("}")
	for i, p := range params {
		parser := g.moduleParamParserName(p.CSort)
		g.emitOneModuleParamParser(parser, p.CSort)
		local := moduleParamLocalName(p)
		g.main.linef("%s, err := %s(__ivyModuleArgs[%d])", local, parser, i)
		g.main.line("if err != nil {")
		g.main.linef("\tfmt.Fprintf(os.Stderr, \"bad parameter %s: %%v\\n\", err)", p.Name)
		g.main.line("\tos.Exit(2)")
		g.main.line("}")
		args = append(args, local)
	}
	return args
}

func (g *Generator) emitModuleParamHelpers() {
	const key = "__ivy_module_param_helpers"
	if g.Ctx.OnceGlobals[key] {
		return
	}
	g.Ctx.OnceGlobals[key] = true
	g.Ctx.AddImport("runtime", "fmt", "")
	g.Ctx.AddImport("runtime", "strings", "")
	w := &g.runtime
	w.line("type ivyModuleParamSpec struct {")
	w.line("\tName       string")
	w.line("\tHasDefault bool")
	w.line("\tDefault    string")
	w.line("}")
	w.blank()
	w.line("func ivyCollectModuleParamArgs(args []string, specs []ivyModuleParamSpec) ([]string, error) {")
	w.line("\tvalues := make([]string, len(specs))")
	w.line("\tseen := make([]bool, len(specs))")
	w.line("\trequired := make([]int, 0, len(specs))")
	w.line("\tbyName := make(map[string]int, len(specs))")
	w.line("\tfor i, spec := range specs {")
	w.line("\t\tbyName[spec.Name] = i")
	w.line("\t\tif spec.HasDefault {")
	w.line("\t\t\tvalues[i] = spec.Default")
	w.line("\t\t\tseen[i] = true")
	w.line("\t\t} else {")
	w.line("\t\t\trequired = append(required, i)")
	w.line("\t\t}")
	w.line("\t}")
	w.line("\tpositional := 0")
	w.line("\tfor _, arg := range args {")
	w.line("\t\tif ivyIsSpecialRuntimeArg(arg) {")
	w.line("\t\t\tcontinue")
	w.line("\t\t}")
	w.line("\t\tif eq := strings.Index(arg, \"=\"); eq >= 0 {")
	w.line("\t\t\tname := strings.TrimPrefix(arg[:eq], \"--\")")
	w.line("\t\t\tif idx, ok := byName[name]; ok {")
	w.line("\t\t\t\tvalues[idx] = arg[eq+1:]")
	w.line("\t\t\t\tseen[idx] = true")
	w.line("\t\t\t\tcontinue")
	w.line("\t\t\t}")
	w.line("\t\t\treturn nil, fmt.Errorf(\"unknown option: %s\", name)")
	w.line("\t\t}")
	w.line("\t\tif positional >= len(required) {")
	w.line("\t\t\treturn nil, fmt.Errorf(\"unexpected argument: %s\", arg)")
	w.line("\t\t}")
	w.line("\t\tidx := required[positional]")
	w.line("\t\tvalues[idx] = arg")
	w.line("\t\tseen[idx] = true")
	w.line("\t\tpositional++")
	w.line("\t}")
	w.line("\tfor _, idx := range required {")
	w.line("\t\tif !seen[idx] {")
	w.line("\t\t\treturn nil, fmt.Errorf(\"usage: %s\", ivyModuleParamUsage(specs))")
	w.line("\t\t}")
	w.line("\t}")
	w.line("\treturn values, nil")
	w.line("}")
	w.blank()
	w.line("func ivyModuleParamUsage(specs []ivyModuleParamSpec) string {")
	w.line("\tnames := make([]string, 0, len(specs))")
	w.line("\tfor _, spec := range specs {")
	w.line("\t\tif !spec.HasDefault {")
	w.line("\t\t\tnames = append(names, spec.Name)")
	w.line("\t\t}")
	w.line("\t}")
	w.line("\treturn strings.Join(names, \" \")")
	w.line("}")
	w.blank()
	w.line("func ivyIsSpecialRuntimeArg(arg string) bool {")
	w.line("\tname := strings.TrimPrefix(arg, \"--\")")
	w.line("\tfor _, special := range []string{\"out\", \"iters\", \"runs\", \"seed\", \"delay\", \"wait\", \"modelfile\"} {")
	w.line("\t\tif strings.HasPrefix(name, special+\"=\") {")
	w.line("\t\t\treturn true")
	w.line("\t\t}")
	w.line("\t}")
	w.line("\treturn false")
	w.line("}")
	w.blank()
}

func (g *Generator) moduleParamParserName(s goivy.Sort) string {
	return "parseModuleParam_" + iteHelperSuffix(g.goType(s))
}

func (g *Generator) emitOneModuleParamParser(parser string, s goivy.Sort) {
	key := "__" + parser
	if g.Ctx.OnceGlobals[key] {
		return
	}
	g.Ctx.OnceGlobals[key] = true
	g.Ctx.AddImport("runtime", "fmt", "")
	w := &g.runtime
	typeName := g.goType(s)
	w.linef("func %s(token string) (%s, error) {", parser, typeName)
	switch st := s.(type) {
	case *goivy.BooleanSort:
		w.line("\tswitch token {")
		w.line("\tcase \"true\":")
		w.line("\t\treturn true, nil")
		w.line("\tcase \"false\":")
		w.line("\t\treturn false, nil")
		w.line("\t}")
		w.line("\treturn false, fmt.Errorf(\"expected true/false, got %q\", token)")
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			w.line("\tn, err := strconv.Atoi(token)")
			w.linef("\treturn %s(n), err", typeName)
		} else {
			w.line("\tswitch token {")
			for _, v := range st.Extension {
				w.linef("\tcase %q:", v)
				w.linef("\t\treturn %s, nil", goExportedName(v))
			}
			w.line("\t}")
			w.linef("\treturn 0, fmt.Errorf(\"expected %s value, got %%q\", token)", goExportedName(st.Name))
		}
	default:
		if goIsAnyIntegerType(g, s) {
			w.line("\tn, err := strconv.Atoi(token)")
			w.linef("\tif err != nil { return %s, err }", g.goZeroValue(s))
			w.linef("\treturn %s(n), nil", typeName)
		} else {
			w.linef("\tvar z %s", typeName)
			w.line("\t_ = token")
			w.line("\treturn z, nil")
		}
	}
	w.line("}")
	w.blank()
}
