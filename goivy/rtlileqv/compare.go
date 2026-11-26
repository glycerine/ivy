package rtlileqv

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/glycerine/ivy/goivy/smt"
)

type pair struct {
	name       bitName
	gold, gate int
}

type inputBit struct {
	bitName
	role string
}

func (b inputBit) label() string {
	if b.role == "state" {
		return b.bitName.String() + " (state)"
	}
	return b.bitName.String()
}

type namedPair struct {
	inputBit
	gold, gate int
}

// Result describes the number of compared cones and any counterexample.
type Result struct {
	Outputs, Registers int
	Failures           []Failure
}

type Failure struct {
	Cone   string
	Inputs map[string]bool
}

func compare(gold, gate *aiger, goldCfg, gateCfg designConfig) (Result, error) {
	ins, err := pairInputs(gold, gate, goldCfg, gateCfg)
	if err != nil {
		return Result{}, err
	}
	outs, err := pairOutputs(gold, gate)
	if err != nil {
		return Result{}, err
	}
	regs, err := pairRegisters(gold, gate)
	if err != nil {
		return Result{}, err
	}
	result := Result{Outputs: len(outs), Registers: len(regs)}
	if len(outs)+len(regs) == 0 {
		return result, nil
	}

	ctx := smt.NewZ3Context()
	defer ctx.Close()
	g, t := compiledAiger{graph: gold, ctx: ctx}, compiledAiger{graph: gate, ctx: ctx}
	g.leaves, t.leaves = make(map[int]int), make(map[int]int)
	var cutNames []inputBit
	var cuts []smt.Z3Expr
	bind := func(name inputBit, goldLiteral, gateLiteral int) {
		i := len(cuts)
		g.leaves[goldLiteral], t.leaves[gateLiteral] = i, i
		cutNames = append(cutNames, name)
		cuts = append(cuts, ctx.BoolConst(fmt.Sprintf("cut_%d", i)))
	}
	for _, p := range ins {
		bind(p.inputBit, gold.inputs[p.gold], gate.inputs[p.gate])
	}
	for _, p := range regs {
		bind(inputBit{bitName: p.name, role: "state"}, gold.latches[p.gold][0], gate.latches[p.gate][0])
	}
	g.compile(cuts)
	t.compile(cuts)

	type cone struct {
		name       string
		gold, gate int
	}
	var cones []cone
	for _, p := range outs {
		cones = append(cones, cone{"output " + p.name.String(), gold.outputs[p.gold], gate.outputs[p.gate]})
	}
	for _, p := range regs {
		cones = append(cones, cone{"reg " + p.name.String(), gold.latches[p.gold][1], gate.latches[p.gate][1]})
	}
	differences := make([]smt.Z3Expr, len(cones))
	for i, c := range cones {
		differences[i] = ctx.Not(ctx.Eq(g.literal(c.gold), t.literal(c.gate)))
	}
	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.Or(differences...))
	switch solver.Check() {
	case smt.Unsat:
		return result, nil
	case smt.Sat:
	default:
		return Result{}, fmt.Errorf("solver could not decide the equivalence query")
	}
	model := solver.Model()
	if model == nil {
		return Result{}, fmt.Errorf("solver did not return a counterexample")
	}
	for i, c := range cones {
		value, ok := model.Eval(differences[i], true)
		if !ok {
			return Result{}, fmt.Errorf("could not evaluate %s in the counterexample", c.name)
		}
		if value.String() != "true" {
			continue
		}
		support := make(map[int]bool)
		g.collectSupport(c.gold, support)
		t.collectSupport(c.gate, support)
		inputs := make(map[string]bool, len(support))
		for cut := range support {
			value, ok := model.Eval(cuts[cut], true)
			if !ok {
				return Result{}, fmt.Errorf("could not evaluate input %s", cutNames[cut].label())
			}
			inputs[cutNames[cut].label()] = value.String() == "true"
		}
		result.Failures = append(result.Failures, Failure{c.name, inputs})
	}
	if len(result.Failures) == 0 {
		return Result{}, fmt.Errorf("solver counterexample has no differing cones")
	}
	return result, nil
}

type compiledAiger struct {
	graph  *aiger
	ctx    *smt.Z3Context
	values map[int]smt.Z3Expr
	leaves map[int]int
	gates  map[int][2]int
}

func (a *compiledAiger) compile(cuts []smt.Z3Expr) {
	a.values = map[int]smt.Z3Expr{0: a.ctx.BoolVal(false)}
	a.gates = make(map[int][2]int, len(a.graph.gates))
	for lit, cut := range a.leaves {
		a.values[lit] = cuts[cut]
	}
	for _, gate := range a.graph.gates {
		a.values[gate[0]] = a.ctx.And(a.literal(gate[1]), a.literal(gate[2]))
		a.gates[gate[0]] = [2]int{gate[1], gate[2]}
	}
}

func (a *compiledAiger) literal(lit int) smt.Z3Expr {
	value := a.values[lit&^1]
	if lit%2 != 0 {
		return a.ctx.Not(value)
	}
	return value
}

func (a *compiledAiger) collectSupport(lit int, support map[int]bool) {
	seen := make(map[int]bool)
	stack := []int{lit}
	for len(stack) > 0 {
		pos := len(stack) - 1
		v := stack[pos] &^ 1
		stack = stack[:pos]
		if seen[v] {
			continue
		}
		seen[v] = true
		if cut, ok := a.leaves[v]; ok {
			support[cut] = true
		} else if gate, ok := a.gates[v]; ok {
			stack = append(stack, gate[:]...)
		}
	}
}

func pairInputs(gold, gate *aiger, goldCfg, gateCfg designConfig) ([]namedPair, error) {
	index := func(names [][]bitName, cfg designConfig) (map[inputBit]int, error) {
		result := make(map[inputBit]int)
		var hasClock, hasReset bool
		for i, aliases := range names {
			if len(aliases) == 0 {
				return nil, fmt.Errorf("[%s] unnamed primary input bit at position %d", cfg.section, i)
			}
			for _, alias := range aliases {
				key := inputBit{bitName: alias}
				switch alias.name {
				case cfg.clock:
					key.name, key.role, hasClock = "clk", "clock", true
				case cfg.reset:
					key.name, key.role, hasReset = "rst", "reset", true
				}
				if prior, ok := result[key]; ok && prior != i {
					return nil, fmt.Errorf("[%s] ambiguous input name %s", cfg.section, alias)
				}
				result[key] = i
			}
		}
		if !hasClock || !hasReset {
			return nil, fmt.Errorf("[%s] clock %q and reset %q must name inputs", cfg.section, cfg.clock, cfg.reset)
		}
		return result, nil
	}
	g, err := index(gold.inputNames, goldCfg)
	if err != nil {
		return nil, err
	}
	t, err := index(gate.inputNames, gateCfg)
	if err != nil {
		return nil, err
	}
	var pairs []namedPair
	gPartner, tPartner := make(map[int]int), make(map[int]int)
	for _, key := range slices.SortedFunc(maps.Keys(g), func(a, b inputBit) int {
		return cmp.Or(cmp.Compare(a.role, b.role), compareBitName(a.bitName, b.bitName))
	}) {
		goldIndex := g[key]
		gateIndex, ok := t[key]
		if !ok {
			return nil, fmt.Errorf("input %s is only in gold", key.label())
		}
		if partner, ok := gPartner[goldIndex]; ok {
			if partner != gateIndex {
				return nil, fmt.Errorf("input %s aliases different positions across designs", key.label())
			}
			continue
		}
		if _, exists := tPartner[gateIndex]; exists {
			return nil, fmt.Errorf("input %s aliases different positions across designs", key.label())
		}
		gPartner[goldIndex], tPartner[gateIndex] = gateIndex, goldIndex
		pairs = append(pairs, namedPair{key, goldIndex, gateIndex})
	}
	for key := range t {
		if _, ok := g[key]; !ok {
			return nil, fmt.Errorf("input %s is only in gate", key.label())
		}
	}
	return pairs, nil
}

func pairOutputs(gold, gate *aiger) ([]pair, error) {
	index := func(names [][]bitName, side string) (map[bitName]int, error) {
		result := make(map[bitName]int)
		for i, aliases := range names {
			if len(aliases) == 0 {
				return nil, fmt.Errorf("[%s] unnamed primary output bit at position %d", side, i)
			}
			for _, name := range aliases {
				if prior, ok := result[name]; ok && prior != i {
					return nil, fmt.Errorf("[%s] ambiguous output name %s", side, name)
				}
				result[name] = i
			}
		}
		return result, nil
	}
	g, err := index(gold.outputNames, "gold")
	if err != nil {
		return nil, err
	}
	t, err := index(gate.outputNames, "gate")
	if err != nil {
		return nil, err
	}
	var pairs []pair
	for _, name := range slices.SortedFunc(maps.Keys(g), compareBitName) {
		pos, ok := t[name]
		if !ok {
			return nil, fmt.Errorf("output %s is only in gold", name)
		}
		pairs = append(pairs, pair{name, g[name], pos})
	}
	for _, name := range slices.SortedFunc(maps.Keys(t), compareBitName) {
		if _, ok := g[name]; !ok {
			return nil, fmt.Errorf("output %s is only in gate", name)
		}
	}
	return pairs, nil
}

func compareBitName(a, b bitName) int {
	return cmp.Or(strings.Compare(a.name, b.name), cmp.Compare(a.bit, b.bit))
}

func pairRegisters(gold, gate *aiger) ([]pair, error) {
	for _, design := range []struct {
		side    string
		circuit *aiger
	}{{"gold", gold}, {"gate", gate}} {
		if pos, found := unnamedLiveLatch(design.circuit); found {
			return nil, fmt.Errorf("[%s] live register bit at position %d has no public name", design.side, pos)
		}
	}
	byAlias := make(map[bitName][]int)
	for pos, aliases := range gate.latchNames {
		for _, name := range aliases {
			byAlias[name] = append(byAlias[name], pos)
		}
	}
	var pairs []pair
	matched := make(map[int]bool)
	for goldPos, aliases := range gold.latchNames {
		if len(aliases) == 0 {
			continue
		}
		var names []bitName
		gatePos := -1
		for _, alias := range aliases {
			for _, pos := range byAlias[alias] {
				if gatePos >= 0 && gatePos != pos {
					return nil, fmt.Errorf("gold register %s aliases several gate registers", aliases[0])
				}
				gatePos = pos
				names = append(names, alias)
			}
		}
		if gatePos < 0 {
			return nil, fmt.Errorf("register %s is only in gold", aliases[0])
		}
		if matched[gatePos] {
			return nil, fmt.Errorf("several gold registers alias gate register %s", gate.latchNames[gatePos][0])
		}
		slices.SortFunc(names, func(a, b bitName) int {
			return cmp.Or(cmp.Compare(len(a.name), len(b.name)), compareBitName(a, b))
		})
		matched[gatePos] = true
		pairs = append(pairs, pair{names[0], goldPos, gatePos})
	}
	for pos, aliases := range gate.latchNames {
		if len(aliases) > 0 && !matched[pos] {
			return nil, fmt.Errorf("register %s is only in gate", aliases[0])
		}
	}
	slices.SortFunc(pairs, func(a, b pair) int { return compareBitName(a.name, b.name) })
	return pairs, nil
}

func unnamedLiveLatch(a *aiger) (int, bool) {
	used := make(map[int]bool)
	for _, gate := range a.gates {
		used[gate[1]&^1], used[gate[2]&^1] = true, true
	}
	for _, output := range a.outputs {
		used[output&^1] = true
	}
	for _, latch := range a.latches {
		used[latch[1]&^1] = true
	}
	for i, latch := range a.latches {
		if len(a.latchNames[i]) == 0 && used[latch[0]] {
			return i, true
		}
	}
	return 0, false
}
