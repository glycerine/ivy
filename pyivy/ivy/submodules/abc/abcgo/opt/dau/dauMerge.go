package dau

/**CFile****************************************************************

  FileName    [dauMerge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Enumeration of decompositions.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauMerge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"sort"
	"strings"
)

type Dau_Sto_t struct {
	Definitions []*dssExpr
	Output      string
	Shared      int32
}

func Dau_DsdMergeMatches(text string, matches []int32) {
	stack := make([]int, 0)
	for i, c := range []byte(text) {
		if i < len(matches) {
			matches[i] = 0
		}
		switch c {
		case '(', '[', '<', '{':
			stack = append(stack, i)
		case ')', ']', '>', '}':
			if len(stack) == 0 {
				panic("unmatched DSD close")
			}
			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			expected := map[byte]byte{'(': ')', '[': ']', '<': '>', '{': '}'}[text[open]]
			if c != expected {
				panic("mismatched DSD delimiter")
			}
			if open >= len(matches) {
				panic("match array is too short")
			}
			matches[open] = int32(i)
		}
	}
	if len(stack) != 0 {
		panic("unclosed DSD delimiter")
	}
}
func dssExprVars(e *dssExpr, present map[int32]bool) {
	if e.kind == DAU_DSD_VAR {
		present[e.variable] = true
	}
	for _, c := range e.children {
		dssExprVars(c, present)
	}
}
func dssExprMap(e *dssExpr, mapping []int32) {
	if e.kind == DAU_DSD_VAR {
		if e.variable < 0 || int(e.variable) >= len(mapping) || mapping[e.variable] < 0 {
			panic("DSD variable mapping is incomplete")
		}
		e.variable = mapping[e.variable]
	}
	for _, c := range e.children {
		dssExprMap(c, mapping)
	}
}
func Dau_DsdMergeReplace(text string, matches, mapping []int32) string {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	dssExprMap(e, mapping)
	return dssExprString(e)
}
func Dau_DsdMergeCopy(text string, complement int32) string {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	e.phase = e.phase != (complement != 0)
	return dssExprString(e)
}
func Dau_DsdMergeVarPres(text string, matches, presence []int32, mask int32) {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	present := make(map[int32]bool)
	dssExprVars(e, present)
	for v := range present {
		presence[v] |= mask
	}
}
func Dau_DsdMergeCountShared(presence []int32, mask int32) int32 {
	count := int32(0)
	for _, v := range presence {
		if v == mask {
			count++
		}
	}
	return count
}
func Dau_DsdMergeFindShared(a, b string, matchesA, matchesB, presence []int32) int32 {
	clear(presence)
	Dau_DsdMergeVarPres(a, matchesA, presence, 1)
	Dau_DsdMergeVarPres(b, matchesB, presence, 2)
	return Dau_DsdMergeCountShared(presence, 3)
}
func Dau_DsdMergeCreateMaps(presence []int32, shared int32, oldToNew, newToOld []int32) int32 {
	counter, unique := int32(0), shared
	for i, v := range presence {
		if v == 0 {
			continue
		}
		next := unique
		if v == 3 {
			next = counter
			counter++
		} else {
			unique++
		}
		oldToNew[i] = next
		newToOld[next] = int32(i)
	}
	return unique
}
func dssExprStatus(e *dssExpr, shared int32, status []int32) int32 {
	value := int32(3)
	if e.kind == DAU_DSD_VAR {
		if e.variable < shared {
			value = 0
		}
	} else if e.kind != DAU_DSD_CONST0 {
		pure := 0
		for _, c := range e.children {
			if dssExprStatus(c, shared, status) == 3 {
				pure++
			}
		}
		switch {
		case pure == 0:
			value = 0
		case pure == len(e.children):
			value = 3
		case pure == 1:
			value = 1
		default:
			value = 2
		}
	}
	if e.open < len(status) {
		status[e.open] = value
	}
	for i := e.start; i < e.open && i < len(status); i++ {
		status[i] = -1
	}
	return value
}
func Dau_DsdMergeStatus(text string, matches []int32, shared int32, status []int32) int32 {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	return dssExprStatus(e, shared, status)
}
func dssPure(e *dssExpr, shared int32) bool {
	if e.kind == DAU_DSD_VAR {
		return e.variable >= shared
	}
	for _, c := range e.children {
		if !dssPure(c, shared) {
			return false
		}
	}
	return true
}
func dssStoreDefinition(store *Dau_Sto_t, e *dssExpr) *dssExpr {
	index := int32(len(store.Definitions))
	store.Definitions = append(store.Definitions, dssClone(e))
	return &dssExpr{kind: DAU_DSD_VAR, variable: index}
}
func dssSubstitutePure(store *Dau_Sto_t, e *dssExpr) *dssExpr {
	if e.kind == DAU_DSD_VAR || e.kind == DAU_DSD_CONST0 {
		return dssClone(e)
	}
	result := *e
	result.children = nil
	pure := 0
	for _, child := range e.children {
		if dssPure(child, store.Shared) {
			pure++
		}
	}
	if pure >= 2 && pure < len(e.children) && (e.kind == DAU_DSD_AND || e.kind == DAU_DSD_XOR) {
		group := &dssExpr{kind: e.kind}
		for _, child := range e.children {
			if dssPure(child, store.Shared) {
				group.children = append(group.children, dssClone(child))
			} else {
				result.children = append(result.children, dssSubstitutePure(store, child))
			}
		}
		result.children = append(result.children, dssStoreDefinition(store, group))
		return &result
	}
	for _, child := range e.children {
		if dssPure(child, store.Shared) {
			definition := dssClone(child)
			phase := definition.phase
			definition.phase = false
			variable := dssStoreDefinition(store, definition)
			variable.phase = phase
			result.children = append(result.children, variable)
		} else {
			result.children = append(result.children, dssSubstitutePure(store, child))
		}
	}
	return &result
}
func Dau_DsdMergeSubstitute(store *Dau_Sto_t, text string, matches, status []int32) string {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	store.Output = dssExprString(dssSubstitutePure(store, e))
	return store.Output
}
func dssInlineDefinitions(e *dssExpr, store *Dau_Sto_t) *dssExpr {
	if e.kind == DAU_DSD_VAR && e.variable >= store.Shared {
		if int(e.variable) >= len(store.Definitions) {
			panic("DSD definition is missing")
		}
		definition := dssClone(store.Definitions[e.variable])
		definition.phase = definition.phase != e.phase
		return definition
	}
	copy := *e
	copy.children = make([]*dssExpr, len(e.children))
	for i, c := range e.children {
		copy.children[i] = dssInlineDefinitions(c, store)
	}
	return &copy
}
func Dau_DsdMergeInlineDefinitions(text string, matches []int32, store *Dau_Sto_t, shared int32) string {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	return dssExprString(dssInlineDefinitions(e, store))
}
func dssRemoveBraces(e *dssExpr) *dssExpr {
	copy := *e
	copy.children = nil
	for _, child := range e.children {
		child = dssRemoveBraces(child)
		flatten := child.kind == e.kind && (e.kind == DAU_DSD_XOR || e.kind == DAU_DSD_AND && !child.phase)
		if flatten {
			fans := child.children
			if child.phase && len(fans) > 0 {
				first := dssClone(fans[0])
				first.phase = !first.phase
				fans = append([]*dssExpr{first}, fans[1:]...)
			}
			copy.children = append(copy.children, fans...)
		} else {
			copy.children = append(copy.children, child)
		}
	}
	return &copy
}
func Dau_DsdRemoveBraces(text string, matches []int32) string {
	e, err := dssParse(text, nil)
	if err != nil {
		panic(err)
	}
	return dssExprString(dssRemoveBraces(e))
}
func RemoveBracesDSD(text string) string { return Dau_DsdRemoveBraces(text, nil) }
func dssEvalExpr(e *dssExpr, assignment int) bool {
	value := false
	switch e.kind {
	case DAU_DSD_CONST0:
	case DAU_DSD_VAR:
		value = assignment&(1<<uint(e.variable)) != 0
	case DAU_DSD_AND:
		value = true
		for _, c := range e.children {
			child := dssEvalExpr(c, assignment)
			value = value && child
		}
	case DAU_DSD_XOR:
		for _, c := range e.children {
			value = value != dssEvalExpr(c, assignment)
		}
	case DAU_DSD_MUX:
		a, b, c := dssEvalExpr(e.children[0], assignment), dssEvalExpr(e.children[1], assignment), dssEvalExpr(e.children[2], assignment)
		value = a && b || !a && c
	case DAU_DSD_PRIME:
		index := 0
		for i, c := range e.children {
			if dssEvalExpr(c, assignment) {
				index |= 1 << uint(i)
			}
		}
		value = dssBit(e.truth, index)
	default:
		panic("invalid DSD operation")
	}
	return value != e.phase
}
func dssExprTruth(e *dssExpr, n int32) []uint64 {
	table := make([]uint64, dssWords(n))
	for i := 0; i < len(table)*64; i++ {
		if dssEvalExpr(e, i) {
			dssSet(table, i)
		}
	}
	return table
}

// MergeDSD follows the source shared-variable/pure-subtree substitution before
// truth decomposition, then inlines definitions and restores original names.
func MergeDSD(a string, permA []int32, b string, permB []int32, complA, complB, nVars int32) (string, error) {
	x, err := dssParse(a, nil)
	if err != nil {
		return "", err
	}
	y, err := dssParse(b, nil)
	if err != nil {
		return "", err
	}
	if permA != nil {
		dssExprMap(x, permA)
	}
	if permB != nil {
		dssExprMap(y, permB)
	}
	x.phase = x.phase != (complA != 0)
	y.phase = y.phase != (complB != 0)
	if x.kind == DAU_DSD_CONST0 {
		if x.phase {
			return dssExprString(y), nil
		}
		return "0", nil
	}
	if y.kind == DAU_DSD_CONST0 {
		if y.phase {
			return dssExprString(x), nil
		}
		return "0", nil
	}
	left, right := make(map[int32]bool), make(map[int32]bool)
	dssExprVars(x, left)
	dssExprVars(y, right)
	var sharedVars, uniqueVars []int32
	all := make(map[int32]bool)
	for v := range left {
		all[v] = true
	}
	for v := range right {
		all[v] = true
	}
	for v := range all {
		if v < 0 || v >= nVars {
			return "", fmt.Errorf("DSD merge variable %d outside %d-variable domain", v, nVars)
		}
		if left[v] && right[v] {
			sharedVars = append(sharedVars, v)
		} else {
			uniqueVars = append(uniqueVars, v)
		}
	}
	sort.Slice(sharedVars, func(i, j int) bool { return sharedVars[i] < sharedVars[j] })
	sort.Slice(uniqueVars, func(i, j int) bool { return uniqueVars[i] < uniqueVars[j] })
	if len(sharedVars) == 0 {
		combined := &dssExpr{kind: DAU_DSD_AND, children: []*dssExpr{x, y}}
		return Dau_DsdNormalize(dssExprString(dssRemoveBraces(combined))), nil
	}
	newToOld := append(append([]int32(nil), sharedVars...), uniqueVars...)
	oldToNew := make([]int32, nVars)
	for i, v := range newToOld {
		oldToNew[v] = int32(i)
	}
	dssExprMap(x, oldToNew)
	dssExprMap(y, oldToNew)
	store := &Dau_Sto_t{Shared: int32(len(sharedVars)), Definitions: make([]*dssExpr, len(sharedVars))}
	reducedX, reducedY := dssSubstitutePure(store, x), dssSubstitutePure(store, y)
	variables := int32(len(newToOld))
	if int32(len(store.Definitions)) > variables {
		variables = int32(len(store.Definitions))
	}
	if variables > DAU_MAX_VAR {
		return "", fmt.Errorf("DSD merge reduced support exceeds %d", DAU_MAX_VAR)
	}
	combined := &dssExpr{kind: DAU_DSD_AND, children: []*dssExpr{reducedX, reducedY}}
	formula, nondec := DecomposeDSD(dssExprTruth(combined, variables), variables)
	if nondec < 0 {
		return "", fmt.Errorf("DSD decomposition failed")
	}
	result, err := dssParse(formula, nil)
	if err != nil {
		return "", err
	}
	result = dssInlineDefinitions(result, store)
	dssExprMap(result, newToOld)
	return Dau_DsdNormalize(dssExprString(dssRemoveBraces(result))), nil
}
func Dau_DsdMerge(a string, permA []int32, b string, permB []int32, complA, complB, nVars int32) string {
	result, err := MergeDSD(a, permA, b, permB, complA, complB, nVars)
	if err != nil {
		return ""
	}
	return result
}
func Dau_DsdMergeStoreClean(p *Dau_Sto_t, shared int32) {
	*p = Dau_Sto_t{Shared: shared, Definitions: make([]*dssExpr, shared)}
}
func Dau_DsdMergeStoreCleanOutput(p *Dau_Sto_t) { p.Output = "" }
func Dau_DsdMergeStorePrintDefs(p *Dau_Sto_t) string {
	var b strings.Builder
	for i, e := range p.Definitions {
		if e != nil {
			fmt.Fprintf(&b, "%c = %s\n", 'a'+i, dssExprString(e))
		}
	}
	return b.String()
}

// The original diagnostic computes this truth and returns before its merge.
func Dau_DsdTest66() { Dau_Dsd6ToTruth("[!(a[be])!(c!df)]") }
