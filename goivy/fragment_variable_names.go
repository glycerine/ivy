package goivy

import (
	"strconv"
	"strings"
)

// A macro-free fragment check alpha-converts its bound and free variables
// together. Only the numeric counter is needed to issue their private names.
// The original formula slices are borrowed until the check finishes.
type fragmentVariableNames struct {
	next             uint64
	assumes, asserts []fmlaPair
	diagnostic       *VariableUniqifier
	spellings        []*LogicVariable
}

const fragmentVariablePrefix = "V!feu!"

func (n *fragmentVariableNames) uniquify(fmla Expr) Expr {
	return rewriteVariables(fmla, make(map[NodeKey]*LogicVariable), n.fresh)
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
// input in the original assumes/asserts order so suffixes, shadowed binders
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
	for _, group := range [][]fmlaPair{n.assumes, n.asserts} {
		for _, p := range group {
			rewriteVariables(p.fmla, make(map[NodeKey]*LogicVariable), remember)
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
	})
}

func (c *checker) originalFormula(fmla Expr) Expr {
	if c.variableNames == nil {
		return c.varUniq.Undo(fmla)
	}
	visible := c.diagnosticFormula(fmla)
	return c.variableNames.diagnostic.Undo(visible)
}
