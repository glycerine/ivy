package goivy

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type traceTracking struct {
	Version Version              `json:"version"`
	Sorts   map[string]traceSort `json:"sorts"`
	Steps   []traceRecordedStep  `json:"steps"`
	Tail    string               `json:"tail"`
}

type traceSort struct {
	Interpretation string   `json:"interpretation,omitempty"`
	Constructors   []string `json:"constructors,omitempty"`
	Bounds         []string `json:"bounds,omitempty"`
}

type traceRecordedStep struct {
	Prefix string          `json:"prefix"`
	Indent int             `json:"indent"`
	Facts  [][2]traceValue `json:"facts"`
}

// traceValue describes one ground model term. Signature lists its argument
// sorts followed by its result sort. Boolean literals have no signature.
type traceValue struct {
	Name      string       `json:"name"`
	Signature []string     `json:"signature,omitempty"`
	Arguments []traceValue `json:"arguments,omitempty"`
}

func recordTraceTracking(trace *TraceBase) *traceTracking {
	tracking := &traceTracking{Version: DefaultVersion, Sorts: make(map[string]traceSort)}
	version := trace.Domain.Cfg.IuCfg.GetStringVersion()
	fmt.Sscanf(version, "%d.%d", &tracking.Version[0], &tracking.Version[1])
	lines := append([]string(nil), trace.Notes...)
	var offset int
	trace.toLines(&lines, nil, 0, trace.HiddenSymbols, false, nil, nil, true,
		func(state *State, indent int, prefix []string) {
			tracking.addState(state, indent, strings.Join(prefix[offset:], ""))
			offset = len(prefix)
		})
	tracking.Tail = strings.Join(lines[offset:], "")
	return tracking
}

func recordTraceSort(s Sort, sig *Sig, sorts map[string]traceSort) string {
	name := IvySortName(s)
	if _, known := sorts[name]; known {
		return name
	}
	info := traceSort{}
	interpretation := sig.Interp.Get(name)
	switch typ := interpretation.(type) {
	case string:
		info.Interpretation = typ
	case *RangeSort:
		s = typ
	}
	switch typ := s.(type) {
	case *RangeSort:
		if typ.Lb != nil && typ.Ub != nil && typ.Lb.IsNumeral() && typ.Ub.IsNumeral() {
			info.Bounds = []string{typ.LbString(), typ.UbString()}
		}
	case *LogicEnumeratedSort:
		info.Constructors = typ.Extension
	}
	sorts[name] = info
	return name
}

func recordTraceValue(expr Expr, sig *Sig, sorts map[string]traceSort) (value traceValue, ok bool) {
	if IsTrue(expr) || IsFalse(expr) {
		return traceValue{Name: expr.String()}, true
	}
	var sym *Const
	switch node := expr.(type) {
	case *Const:
		sym = node
	case *Apply:
		sym, _ = node.Func.(*Const)
		for _, arg := range node.Terms {
			record, supported := recordTraceValue(arg, sig, sorts)
			if !supported {
				return traceValue{}, false
			}
			value.Arguments = append(value.Arguments, record)
		}
	}
	if sym == nil || sym.CSort == nil {
		return traceValue{}, false
	}
	if SortEqual(sym.CSort, Boolean) && (sym.Name == "true" || sym.Name == "false") {
		return traceValue{Name: sym.Name}, true
	}
	value.Name = sym.Name
	for _, s := range SortDomain(sym.CSort) {
		value.Signature = append(value.Signature, recordTraceSort(s, sig, sorts))
	}
	value.Signature = append(value.Signature, recordTraceSort(SortRange(sym.CSort), sig, sorts))
	return value, true
}

func (tracking *traceTracking) addState(state *State, indent int, prefix string) {
	step := traceRecordedStep{Prefix: prefix, Indent: indent}
	if state.Clauses != nil {
		for _, formula := range state.Clauses.Fmlas {
			eq, ok := formula.(*Eq)
			if !ok {
				continue
			}
			lhs, leftOK := recordTraceValue(eq.T1, state.Domain.Sig, tracking.Sorts)
			rhs, rightOK := recordTraceValue(eq.T2, state.Domain.Sig, tracking.Sorts)
			if leftOK && rightOK {
				step.Facts = append(step.Facts, [2]traceValue{lhs, rhs})
			}
		}
	}
	tracking.Steps = append(tracking.Steps, step)
}

func (tracking *traceTracking) newModule() *Module {
	mod := New()
	SetStringVersionOn(mod.Cfg.IuCfg, fmt.Sprintf("%d.%d", tracking.Version[0], tracking.Version[1]))
	for name, record := range tracking.Sorts {
		if name == "bool" {
			continue
		}
		var s Sort = &UninterpretedSort{Name: name}
		switch {
		case len(record.Constructors) != 0:
			s = &LogicEnumeratedSort{Name: name, Extension: record.Constructors}
			for _, constructor := range record.Constructors {
				mod.Sig.Constructors[constructor] = true
			}
		case len(record.Bounds) == 2:
			s = &RangeSort{Name: name, Lb: NumeralBound{Value: record.Bounds[0]}, Ub: NumeralBound{Value: record.Bounds[1]}}
			mod.Sig.Interp.Set(name, s)
		}
		mod.Sig.Sorts.Set(name, s)
		if record.Interpretation != "" {
			mod.Sig.Interp.Set(name, record.Interpretation)
		}
	}
	return mod
}

func (value traceValue) term(sig *Sig) (Expr, error) {
	if len(value.Signature) == 0 {
		switch value.Name {
		case "true":
			return &LogicAnd{}, nil
		case "false":
			return &LogicOr{}, nil
		default:
			return nil, fmt.Errorf("invalid Boolean literal %q in trace", value.Name)
		}
	}
	var sorts []Sort
	for _, name := range value.Signature {
		s := sig.Sorts.Get(name)
		if s == nil {
			return nil, fmt.Errorf("unknown sort %q in trace", name)
		}
		sorts = append(sorts, s)
	}
	s := sorts[0]
	if len(sorts) > 1 {
		var err error
		s, err = NewFunctionSort(sorts...)
		if err != nil {
			return nil, err
		}
	}
	sym := NewConst(value.Name, s)
	if !IsNumeral(sym) {
		var err error
		sym, err = sig.AddSymbol(value.Name, s)
		if err != nil {
			return nil, err
		}
	}
	if len(value.Arguments) == 0 {
		return sym, nil
	}
	var args []Expr
	for _, record := range value.Arguments {
		argument, err := record.term(sig)
		if err != nil {
			return nil, err
		}
		args = append(args, argument)
	}
	return NewApply(sym, args...)
}

func (step traceRecordedStep) track(tracking *traceTracking, terms []string, missing map[string]error) (map[string]string, error) {
	mod := tracking.newModule()
	var facts []Expr
	for _, record := range step.Facts {
		lhs, err := record[0].term(mod.Sig)
		if err != nil {
			return nil, err
		}
		rhs, err := record[1].term(mod.Sig)
		if err != nil {
			return nil, err
		}
		eq, err := NewEq(lhs, rhs)
		if err != nil {
			return nil, err
		}
		facts = append(facts, eq)
	}
	values := make(map[string]string)
	if len(facts) == 0 {
		return values, nil
	}
	clauses := NewClauses(facts, nil, EmptyAnnotation{})
	solver := NewSolverFromSig(mod.Sig, nil)
	defer solver.Close()
	model, err := solver.GetModelClauses(clauses)
	if model == nil || err != nil {
		return values, err
	}
	compiler := NewCompiler(mod.Sig, mod)
	herbrand := NewHerbrandModel(solver, model.Solver, model.Model, model.Vocab)
	numerals := NumeralAssignWithClauses(herbrand, clauses)
	for _, term := range terms {
		formula, err := compiler.ToFormulaV(term, tracking.Version)
		if err != nil {
			if _, seen := missing[term]; !seen {
				missing[term] = err
			}
			continue
		}
		missing[term] = nil
		value := RenameASTByName(herbrand.EvalToConstant(formula), numerals)
		values[formula.String()] = value.String()
	}
	return values, nil
}

func (tracking *traceTracking) replay(out io.Writer, track string) error {
	if strings.HasSuffix(track, ".trk") {
		data, err := os.ReadFile(track)
		if err != nil {
			return err
		}
		track = string(data)
	}
	var terms []string
	for _, term := range strings.Split(track, ";") {
		if term = strings.TrimSpace(term); term != "" {
			terms = append(terms, term)
		}
	}
	seen := make(map[string]string)
	missing := make(map[string]error)
	var text strings.Builder
	for _, step := range tracking.Steps {
		if step.Indent < 0 {
			return fmt.Errorf("invalid trace indentation %d", step.Indent)
		}
		text.WriteString(step.Prefix)
		values, err := step.track(tracking, terms, missing)
		if err != nil {
			return err
		}
		var changes []string
		for term, value := range values {
			if previous, ok := seen[term]; !ok || previous != value {
				seen[term] = value
				changes = append(changes, term+" = "+value)
			}
		}
		if len(changes) > 0 {
			sort.Strings(changes)
			indent := strings.Repeat("    ", step.Indent)
			fmt.Fprintln(&text, indent+"[")
			for _, line := range changes {
				fmt.Fprintln(&text, indent+"    "+line)
			}
			fmt.Fprintln(&text, indent+"]")
		}
	}
	for _, term := range terms {
		if err, known := missing[term]; !known || err != nil {
			return fmt.Errorf("cannot track %q: expression is unavailable in this trace", term)
		}
	}
	text.WriteString(tracking.Tail)
	_, err := io.WriteString(out, text.String())
	return err
}
