package goivy

import (
	"fmt"
	"regexp"
	"strings"
)

type InvariantPatternDifference struct {
	Base
	Left, Right Node
}

func (p *InvariantPatternDifference) Args() []Node { return []Node{p.Left, p.Right} }

func (p *InvariantPatternDifference) Clone(args []Node) Node {
	return &InvariantPatternDifference{Base: p.Base, Left: args[0], Right: args[1]}
}

func (p *InvariantPatternDifference) String() string {
	return fmt.Sprintf("(%s - %s)", p.Left, p.Right)
}

func (p *InvariantPatternDifference) Canon() Canonical {
	return Canonical(fmt.Sprintf("(pattern-diff%v args:%v)", p.Base.canonFields(), SliceCanon(p.Args())))
}

type InvariantUsingDecl struct{ DeclBase }

func (d *InvariantUsingDecl) Clone(args []Node) Node {
	return &InvariantUsingDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}

func (d *InvariantUsingDecl) Defines() []string { return nil }
func (d *InvariantUsingDecl) String() string    { return "usingpat" }

type InvariantPatternDecl struct{ DeclBase }

func (d *InvariantPatternDecl) Clone(args []Node) Node {
	return &InvariantPatternDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}

func (d *InvariantPatternDecl) Defines() []string { return []string{labelName(d.DeclArgs[0])} }
func (d *InvariantPatternDecl) String() string    { return "patdef" }

func expandInvariantPattern(pattern Node, definitions map[string]Node) Node {
	if atom, ok := pattern.(*Atom); ok {
		if replacement, ok := definitions[atom.Rep]; ok {
			return replacement
		}
		return atom
	}
	args := pattern.Args()
	expanded := make([]Node, len(args))
	for i, arg := range args {
		expanded[i] = expandInvariantPattern(arg, definitions)
	}
	return pattern.Clone(expanded)
}

func compileInvariantPattern(pattern Node, owner string) func(string) bool {
	if pattern == nil {
		return func(string) bool { return true }
	}
	switch p := pattern.(type) {
	case *Or:
		var alternatives []func(string) bool
		for _, term := range p.Args() {
			alternatives = append(alternatives, compileInvariantPattern(term, owner))
		}
		return func(name string) bool {
			for _, matches := range alternatives {
				if matches(name) {
					return true
				}
			}
			return false
		}
	case *InvariantPatternDifference:
		left, right := compileInvariantPattern(p.Left, owner), compileInvariantPattern(p.Right, owner)
		return func(name string) bool { return left(name) && !right(name) }
	case *Atom:
		leaf, root := strings.CutPrefix(p.Rep, "$")
		prefixes := []string{""}
		if !root {
			for i, c := range owner {
				if c == '.' {
					prefixes = append(prefixes, regexp.QuoteMeta(owner[:i+1]))
				}
			}
		}
		body := strings.ReplaceAll(regexp.QuoteMeta(leaf), `\*`, `.*`)
		re := regexp.MustCompile(`^(?:` + strings.Join(prefixes, "|") + `)` + body + `\z`)
		return re.MatchString
	default:
		return func(string) bool { return false }
	}
}

func gatedInvariantClauses(mod *Module) (*Clauses, *InsMap[string, Expr]) {
	var formulas []Expr
	literals := NewInsMap[string, Expr]()
	for _, invariants := range [][]*LabeledFormula{mod.LabeledConjs, mod.AssumedInvs} {
		for _, lf := range invariants {
			if lf.Explicit || lf.Unprovable {
				continue
			}
			formula := lf.Formula.(Expr)
			if name := lfLabelName(lf); name != "" {
				lit := NewConst("__using$"+name, Boolean)
				literals.Set(name, lit)
				formula = &LogicImplies{T1: lit, T2: formula}
			}
			formulas = append(formulas, formula)
		}
	}
	return NewClauses(formulas, nil, EmptyAnnotation{}), literals
}

func applyInvariantUsing(lf *LabeledFormula, pattern Node, literals *InsMap[string, Expr]) *LabeledFormula {
	if literals.Len() == 0 {
		return lf
	}
	matches := compileInvariantPattern(pattern, lfLabelName(lf))
	var selected []Expr
	for name, lit := range literals.All() {
		if matches(name) {
			selected = append(selected, lit)
		}
	}
	if len(selected) == 0 {
		return lf
	}
	formula := &LogicImplies{T1: &LogicAnd{Terms: selected}, T2: lf.Formula.(Expr)}
	return lf.Clone([]Node{lf.Label, formula}).(*LabeledFormula)
}
