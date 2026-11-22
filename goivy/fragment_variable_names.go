package goivy

import (
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// A fragment check alpha-converts its bound and free variables
// together. Only the numeric counter is needed to issue their private names.
// The original formula slices are borrowed until the check finishes.
type fragmentVariableNames struct {
	next                     uint64
	assumes, asserts, macros []fmlaPair
	sourceSpelling           bool
	diagnostic               *VariableUniqifier
	spellings                []*LogicVariable
}

const fragmentVariablePrefix = "V!feu!"

func (n *fragmentVariableNames) uniquify(fmla Expr) Expr {
	return rewriteVariables(fmla, make(map[NodeKey]*LogicVariable), n.fresh, n.canRename)
}

// A macro can use a constant as a formal or actual parameter. Its name and
// sort then share a key with variables. Conventional variable names always
// begin with uppercase, as do the private names. Lowercase constants are
// disjoint from both. Check this during the alpha conversion already in flight.
func (n *fragmentVariableNames) canRename(fmla Expr) bool {
	if n.sourceSpelling {
		return false
	}
	switch x := fmla.(type) {
	case *Const:
		if len(n.macros) == 0 || len(x.Name) == 0 || !isUpper(x.Name[0]) {
			return true
		}
	case *Apply:
		// CloneApplyTerms preserves the head, which can itself contain variables.
		if _, ok := x.Func.(*Const); ok {
			return true
		}
	case *LogicVariable, *Eq, *LogicNot, *LogicAnd, *LogicOr, *LogicImplies, *LogicIff, *LogicIte,
		*ForAll, *LogicExists, *Lambda, *Cond, *LogicGlobally, *LogicEventually, *LogicWhenOperator, *IvyDefinition:
		return true
	}
	// Some.clone_binder keeps its original body. Other syntax, including raw
	// binders, can also preserve source variables outside the replaced children.
	n.sourceSpelling = true
	return false
}

func (n *fragmentVariableNames) fresh(input *LogicVariable) *LogicVariable {
	name := fragmentVariablePrefix + strconv.FormatUint(n.next, 10)
	n.next++
	if n.next == 0 {
		panic("fragment variable name counter overflow")
	}
	return &LogicVariable{Name: name, VSort: input.VSort}
}

// Only diagnostics need the spellings used by Python. Replay the unchanged
// input in the original assumptions, assertions and macros order so suffixes, shadowed binders
// and the standard inverse substitution retain their public meaning.
func (n *fragmentVariableNames) prepareDiagnostics() {
	if n.diagnostic != nil {
		return
	}
	n.diagnostic = NewVariableUniqifier(nil)
	remember := func(input *LogicVariable) *LogicVariable {
		visible := n.diagnostic.fresh(input)
		n.spellings = append(n.spellings, visible)
		return visible
	}
	for _, group := range [][]fmlaPair{n.assumes, n.asserts, n.macros} {
		for _, p := range group {
			rewriteVariables(p.fmla, make(map[NodeKey]*LogicVariable), remember, nil)
		}
	}
}

func (c *checker) diagnosticFormula(fmla Expr) Expr {
	n := c.variableNames
	if n == nil {
		return fmla
	}
	n.prepareDiagnostics()
	return rewriteVariables(fmla, make(map[NodeKey]*LogicVariable), func(v *LogicVariable) *LogicVariable {
		suffix, ok := strings.CutPrefix(v.Name, fragmentVariablePrefix)
		if !ok {
			return v
		}
		id, err := strconv.ParseUint(suffix, 10, 64)
		if err != nil || id >= uint64(len(n.spellings)) {
			return v
		}
		return n.spellings[id]
	}, nil)
}

func (c *checker) originalFormula(fmla Expr) Expr {
	if c.variableNames == nil {
		return c.varUniq.Undo(fmla)
	}
	visible := c.diagnosticFormula(fmla)
	return c.variableNames.diagnostic.Undo(visible)
}

// renameFragmentInputs finishes choosing names before any stratification nodes
// are created. The source pass can reuse the output slices if it is needed.
func (c *checker) renameFragmentInputs(assumes, asserts, macros []fmlaPair) ([]fmlaPair, []fmlaPair, []fmlaPair) {
	inputs := [3][]fmlaPair{assumes, asserts, macros}
	var output [3][]fmlaPair
	for i, xs := range inputs {
		output[i] = make([]fmlaPair, len(xs))
	}
	rename := func(uniq func(Expr) Expr) {
		for i, xs := range inputs {
			for j, p := range xs {
				p.fmla = uniq(p.fmla)
				output[i][j] = p
			}
		}
	}
	if !xtracer.Enabled || xtracer.Suppressed {
		n := &fragmentVariableNames{assumes: assumes, asserts: asserts, macros: macros}
		rename(n.uniquify)
		if !n.sourceSpelling {
			c.variableNames = n
			return output[0], output[1], output[2]
		}
	}
	c.varUniq = NewVariableUniqifier(nil)
	rename(c.varUniq.Uniquify)
	return output[0], output[1], output[2]
}
