package goivy

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// Python add_prem_match turns a match problem's pattern and instance into
// ia.Tuple(*(pats + [prob.pat])) and ia.Tuple(*(insts + [prob.inst])):
// the premise patterns and instances are the LabeledFormulas themselves
// (sprem from the schema, gprem from lookup_schema), and the last element
// is the original pattern/instance expression. The functions here apply
// matches to such a tuple and match its elements, mirroring how Python's
// generic apply_match/fo_match/match recurse through a LabeledFormula's
// args (label, formula).

// tupleCanon is Python's '(tuple elems:[...])' canon.
func tupleCanon(elems []Node) string {
	parts := make([]string, len(elems))
	for i, e := range elems {
		parts[i] = string(e.Canon())
	}
	return "(tuple elems:[" + strings.Join(parts, " ") + "])"
}

// tupleElemFormula returns the formula part of a tuple element: a
// LabeledFormula's formula, or the element itself.
func tupleElemFormula(n Node) (Expr, bool) {
	if lf, ok := n.(*LabeledFormula); ok {
		e, ok := lf.Formula.(Expr)
		return e, ok
	}
	e, ok := n.(Expr)
	return e, ok
}

// tupleElemWith rebuilds a tuple element with a new formula; a
// LabeledFormula is cloned (Python LabeledFormula.clone, traced).
func tupleElemWith(n Node, f Expr) Node {
	if lf, ok := n.(*LabeledFormula); ok {
		return lf.Clone([]Node{lf.Label, f})
	}
	return f
}

// applyMatchTuple is Python apply_match / apply_match_alt on an ia.Tuple:
//
//	freevars = match_rhs_vars(match)
//	fmla = il.alpha_avoid(fmla, freevars)   # one uniqifier over the tuple
//	result = apply_match_rec(match, fmla, env)
//
// Both passes rebuild every element, so a LabeledFormula is cloned twice.
func applyMatchTuple(match map[NodeKey]Expr, elems []Node, alt bool) []Node {
	name := "proof.ApplyMatch"
	if alt {
		name = "proof.ApplyMatchAlt"
	}
	xtracer.Trace("%s ENTER nmatch=%d fmlaType=Tuple", name, len(match))

	// alpha_avoid(Tuple, vs): reserve set(vs) | free_variables(Tuple), each
	// once, then rename bound variables with one uniqifier.
	vs := MatchRhsVars(match)
	vu := NewVariableUniqifier(nil)
	for _, v := range vs {
		switch x := v.(type) {
		case *LogicVariable:
			vu.rn.Rename(x.Name)
		case *Const:
			vu.rn.Rename(x.Name)
		case *UninterpretedSort:
			vu.rn.Rename(x.Name)
		}
	}
	vmap := make(map[NodeKey]*LogicVariable)
	reserved := make(map[NodeKey]bool)
	for _, e := range elems {
		f, ok := tupleElemFormula(e)
		if !ok || f == nil {
			continue
		}
		for _, v := range FreeVariablesList(f) {
			if _, inVs := vs[Key(v)]; !inVs && !reserved[Key(v)] {
				reserved[Key(v)] = true
				vu.rn.Rename(v.Name)
			}
			vmap[Key(v)] = v
		}
	}
	for _, v := range vs {
		if lv, ok := v.(*LogicVariable); ok {
			vmap[Key(lv)] = lv
		}
	}
	renamed := make([]Node, len(elems))
	for i, e := range elems {
		f, ok := tupleElemFormula(e)
		if !ok || f == nil {
			renamed[i] = e
			continue
		}
		renamed[i] = tupleElemWith(e, vu.rec(f, vmap))
	}

	// apply_match_rec / apply_match_alt_rec over the tuple's args.
	env := make(map[NodeKey]bool)
	result := make([]Node, len(renamed))
	for i, e := range renamed {
		f, ok := tupleElemFormula(e)
		if !ok || f == nil {
			result[i] = e
			continue
		}
		var g Expr
		if alt {
			g = applyMatchAltRec(match, f, env)
		} else {
			g = applyMatchRec(match, f)
		}
		result[i] = tupleElemWith(e, g)
	}
	xtracer.Trace("%s EXIT HASH canon=%s", name, tupleCanon(result))
	return result
}

// labelTypeName is the Python class name of a LabeledFormula label.
func labelTypeName(n Node) string {
	if n == nil {
		return "NoneType"
	}
	return TypeName(n)
}

// labelsHeadsMatch is heads_match on two labels: both non-App AST nodes,
// so they match when they have the same type and number of args.
func labelsHeadsMatch(p, i Node) bool {
	if p == nil || i == nil {
		return false
	}
	return TypeName(p) == TypeName(i) && len(p.Args()) == len(i.Args())
}

// foMatchTupleElem is Python fo_match on one pair of tuple elements.
func foMatchTupleElem(pat, inst Node, freesyms *InsMap[NodeKey, Expr], constants map[NodeKey]Expr) map[NodeKey]Expr {
	plf, pIsLF := pat.(*LabeledFormula)
	if !pIsLF {
		p, _ := pat.(Expr)
		i, _ := inst.(Expr)
		return FOMatch(p, i, freesyms, constants)
	}
	xtracer.Trace("proof.FOMatch ENTER patType=LabeledFormula instType=%s", TypeName(inst))
	ilf, iIsLF := inst.(*LabeledFormula)
	// heads_match: same type and number of args (label, formula).
	if !iIsLF {
		return map[NodeKey]Expr{}
	}
	// fo_match on the labels: neither is a variable or quantifier; heads
	// match when they have the same type and arity, and their (empty)
	// args merge to {}.
	xtracer.Trace("proof.FOMatch ENTER patType=%s instType=%s", labelTypeName(plf.Label), labelTypeName(ilf.Label))
	labelMatch := map[NodeKey]Expr{}
	pf, _ := plf.Formula.(Expr)
	inf, _ := ilf.Formula.(Expr)
	formulaMatch := FOMatch(pf, inf, freesyms, constants)
	return MergeMatches(labelMatch, formulaMatch)
}

// matchTupleElem is Python match on one pair of tuple elements.
func matchTupleElem(pat, inst Node, freesyms *InsMap[NodeKey, Expr], constants map[NodeKey]Expr) map[NodeKey]Expr {
	plf, pIsLF := pat.(*LabeledFormula)
	if !pIsLF {
		p, _ := pat.(Expr)
		i, _ := inst.(Expr)
		return Match(p, i, freesyms, constants)
	}
	nfree := 0
	if freesyms != nil {
		nfree = freesyms.Len()
	}
	xtracer.Trace("proof.Match ENTER patType=LabeledFormula instType=%s nfreesyms=%d", TypeName(inst), nfree)
	ilf, iIsLF := inst.(*LabeledFormula)
	if !iIsLF {
		xtracer.Trace("proof.Match EXIT noMatch")
		return nil
	}
	// match on the labels.
	xtracer.Trace("proof.Match ENTER patType=%s instType=%s nfreesyms=%d", labelTypeName(plf.Label), labelTypeName(ilf.Label), nfree)
	var labelMatch map[NodeKey]Expr
	if labelsHeadsMatch(plf.Label, ilf.Label) {
		labelMatch = map[NodeKey]Expr{}
		xtracer.Trace("proof.Match EXIT viaHeads success=True")
	} else {
		xtracer.Trace("proof.Match EXIT noMatch")
	}
	pf, _ := plf.Formula.(Expr)
	inf, _ := ilf.Formula.(Expr)
	formulaMatch := Match(pf, inf, freesyms, constants)
	res := MergeMatches(labelMatch, formulaMatch)
	xtracer.Trace("proof.Match EXIT viaHeads success=%s", BoolPythonStr(res != nil))
	return res
}

// tupleElemsString renders tuple elements for debugging.
func tupleElemsString(elems []Node) string {
	parts := make([]string, len(elems))
	for i, e := range elems {
		parts[i] = fmt.Sprint(e)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
