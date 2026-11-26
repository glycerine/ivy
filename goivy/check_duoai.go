package goivy

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// DuoAIExport writes a first-order Ivy program with the isolate's calls inlined.
func DuoAIExport(mod *Module) error {
	var text bytes.Buffer
	for name, typ := range mod.Sig.Sorts.All() {
		if name == "bool" {
			continue
		}
		fmt.Fprintf(&text, "type %s\n", name)
		if native, ok := mod.Sig.Interp.Get(name).(string); ok {
			fmt.Fprintf(&text, "interpret %s -> %s\n", name, native)
		}
		if enum, ok := typ.(*LogicEnumeratedSort); ok {
			writeDuoAIEnumAxioms(&text, name, enum.Extension)
		}
	}
	for name, entry := range mod.Sig.Symbols.All() {
		if entry.Union != nil || mod.Sig.Constructors[name] || IsNumeralName(name) {
			continue
		}
		args := []string{}
		for i, typ := range SortDomain(entry.Sort) {
			args = append(args, fmt.Sprintf("V%d:%s", i, IvySortName(typ)))
		}
		signature := name
		if len(args) > 0 {
			signature += "(" + strings.Join(args, ", ") + ")"
		}
		result := SortRange(entry.Sort)
		switch {
		case SortEqual(result, Boolean):
			fmt.Fprintf(&text, "relation %s\n", signature)
		case len(args) == 0:
			fmt.Fprintf(&text, "individual %s : %s\n", signature, IvySortName(result))
		default:
			fmt.Fprintf(&text, "function %s : %s\n", signature, IvySortName(result))
		}
	}
	for _, axiom := range mod.LabeledAxioms {
		fmt.Fprintf(&text, "axiom %s\n", axiom.Formula)
	}
	for _, assumption := range mod.AssumedInvs {
		fmt.Fprintf(&text, "axiom %s\n", assumption.Formula)
	}
	for _, def := range mod.Definitions {
		fmt.Fprintf(&text, "axiom %s\n", def.Formula)
	}
	for _, initial := range mod.LabeledInits {
		fmt.Fprintf(&text, "init %s\n", initial.Formula)
	}
	for _, initializer := range mod.Initializers {
		action := &duoaiAction{mod: mod, used: make(map[string]bool)}
		body, err := action.translate(initializer.Action)
		if err != nil {
			return err
		}
		if len(action.params) != 0 {
			return fmt.Errorf("DuoAI initializer %s contains action-local variables", initializer.Name)
		}
		fmt.Fprintf(&text, "after init {%s}\n", body)
	}
	var conjs []Expr
	for _, invariant := range mod.LabeledConjs {
		conjs = append(conjs, invariant.Formula.(Expr))
	}
	fmt.Fprintf(&text, "conjecture [1000000] %s\n", &LogicAnd{Terms: conjs})
	var exports []string
	for name := range mod.PublicActions.All() {
		exports = append(exports, name)
	}
	sort.Strings(exports)
	for _, name := range exports {
		action := mod.Actions.Get(name)
		if action == nil {
			continue
		}
		if _, environment := action.(*LogicEnvAction); environment {
			continue
		}
		inliner := duoaiAction{mod: mod, used: make(map[string]bool)}
		for _, symbol := range slices.Concat(action.GetFormalParams(), action.GetFormalReturns()) {
			inliner.used[symbol.Name] = true
		}
		for _, symbol := range collectAllSymbols([]Expr{action}) {
			inliner.used[symbol.Name] = true
		}
		body, err := inliner.translate(action)
		if err != nil {
			return err
		}
		var args []string
		for _, param := range slices.Concat(action.GetFormalParams(), inliner.params) {
			args = append(args, param.Name+":"+IvySortName(param.CSort))
		}
		fmt.Fprintf(&text, "action %s(%s)", name, strings.Join(args, ", "))
		if returns := action.GetFormalReturns(); len(returns) > 0 {
			var names []string
			for _, result := range returns {
				names = append(names, result.Name+":"+IvySortName(result.CSort))
			}
			fmt.Fprintf(&text, " returns (%s)", strings.Join(names, ", "))
		}
		fmt.Fprintf(&text, " = {%s}\nexport %s\n", body, name)
	}
	path := mod.Name + "_duoai.ivy"
	if mod.Name == "" {
		path = "ivy_duoai.ivy"
	}
	output := duoaiQualifiedName.ReplaceAllString(text.String(), "_$1")
	output = strings.NewReplacer("ext:", "ext_", "int$", "int_", "fml:", "fml_", "loc:", "loc_", "prm:", "prm_").Replace(output)
	return os.WriteFile(path, []byte(output), 0o644)
}

var duoaiQualifiedName = regexp.MustCompile(`\.([a-zA-Z0-9_])`)

func writeDuoAIEnumAxioms(out *bytes.Buffer, name string, constructors []string) {
	var members []string
	for i, constructor := range constructors {
		fmt.Fprintf(out, "individual %s : %s\n", constructor, name)
		members = append(members, "X = "+constructor)
		for _, previous := range constructors[:i] {
			fmt.Fprintf(out, "axiom ~(%s = %s)\n", constructor, previous)
		}
	}
	if len(members) > 0 {
		fmt.Fprintf(out, "axiom forall X:%s. %s\n", name, strings.Join(members, " | "))
	}
}

type duoaiAction struct {
	mod    *Module
	used   map[string]bool
	active []string
	params []*Const
	serial int
}

func (d *duoaiAction) takeLocals(body ActionsAction, locals []*Const) (Expr, error) {
	replacements := make(map[NodeKey]Expr)
	for _, local := range locals {
		var name string
		for {
			name = fmt.Sprintf("duoai_local%d", d.serial)
			d.serial++
			if !d.used[name] && d.mod.Sig.Symbols.Get(name) == nil {
				break
			}
		}
		d.used[name] = true
		symbol := NewConst(name, local.CSort)
		d.params = append(d.params, symbol)
		replacements[Key(local)] = symbol
		replacements[Key(OldConst(local))] = OldConst(symbol)
	}
	return d.translate(SubstituteConstantsAction(body, replacements))
}

func (d *duoaiAction) translate(expr Expr) (Expr, error) {
	switch action := expr.(type) {
	case *LogicCallAction:
		name := action.CalleeName()
		for _, active := range d.active {
			if name == active {
				return nil, NewIvyError(action, "cannot inline recursive DuoAI action "+name)
			}
		}
		callee, ok := d.mod.FindAction(name)
		if !ok {
			return nil, NewIvyError(action, "undefined DuoAI action "+name)
		}
		locals, body := action.bindActuals(d.mod, callee)
		d.active = append(d.active, name)
		inner, err := d.takeLocals(body, locals)
		d.active = d.active[:len(d.active)-1]
		return inner, err
	case *LogicLocalAction:
		var locals []*Const
		for _, local := range action.Locals {
			symbol, ok := local.(*Const)
			if !ok {
				return nil, NewIvyError(action, "DuoAI local declaration must name a symbol")
			}
			locals = append(locals, symbol)
		}
		return d.takeLocals(action.Body.(ActionsAction), locals)
	case *LogicBindOldsAction:
		for _, symbol := range collectAllSymbols([]Expr{action.Inner}) {
			if IsOld(symbol.Name) {
				return nil, NewIvyError(action, "cannot translate old operator to DuoAI")
			}
		}
		return d.translate(action.Inner)
	case ActionsAction:
		var args []Expr
		for _, arg := range action.ActionArgs() {
			inner, err := d.translate(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, inner)
		}
		return action.ActionClone(args), nil
	default:
		return expr, nil
	}
}
