package abc

// Port of src/aig/aig/aigJust.c: justification and ternary simulation.

import "slices"

// AigTernary preserves aigJust.c's two-bit value encoding.
type AigTernary uint8

const (
	AigTernaryZero    AigTernary = 1
	AigTernaryOne     AigTernary = 2
	AigTernaryUnknown AigTernary = 3
)

func (v AigTernary) Not() AigTernary {
	switch v {
	case AigTernaryZero:
		return AigTernaryOne
	case AigTernaryOne:
		return AigTernaryZero
	default:
		return AigTernaryUnknown
	}
}

func aigTernaryAnd(a, b AigTernary) AigTernary {
	if a == AigTernaryZero || b == AigTernaryZero {
		return AigTernaryZero
	}
	if a == AigTernaryOne && b == AigTernaryOne {
		return AigTernaryOne
	}
	return AigTernaryUnknown
}

type AigJustificationOptions struct {
	MaxVisits, MaxLiterals int // zero: 1M each
	Stop                   func() bool
}

type AigJustificationExperimentOptions struct {
	Justification AigJustificationOptions
	PatternPack   AigPatternPackOptions
	MaxCandidates int // zero: 1M
}

type AigJustificationExperiment struct {
	Candidates            []AigID
	CountZero, FailedZero int
	CountOne, FailedOne   int
	TotalLiterals         int
	Patterns, Care, Signs []uint64
	PatternStats          AigPatternPackStats
}

func checkAigJustification(g *AigManager, root AigID, o AigJustificationOptions) (AigJustificationOptions, error) {
	if g == nil || o.MaxVisits < 0 || o.MaxLiterals < 0 {
		return o, ErrDimension
	}
	if err := g.Check(); err != nil {
		return o, err
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return o, ErrInvalidLiteral
	}
	if o.MaxVisits == 0 {
		o.MaxVisits = 1 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 1 << 20
	}
	return o, nil
}

type aigJustFrame struct {
	id      AigID
	desired bool
	heur    int
	stage   uint8
}

// FindJustification implements Aig_ObjFindSatAssign. Returned SAT literals use
// CI ordinals as variables; a negative literal assigns that CI false. The
// eight native deterministic heuristics are tried in order, and the first
// consistent partial assignment is returned.
func (g *AigManager) FindJustification(root AigID, desired bool, options AigJustificationOptions) ([]SatLit, bool, error) {
	options, err := checkAigJustification(g, root, options)
	if err != nil {
		return nil, false, err
	}
	object := g.objects[root]
	if object.Type == AigCO {
		desired = desired != object.Fanin0.Complemented()
		root = object.Fanin0.ID()
	}
	ciIndex := make([]int, len(g.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for i, id := range g.cis {
		ciIndex[id] = i
	}
	visits := 0
	step := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if visits == options.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
	var lastLiterals []SatLit
	for initialHeuristic := 0; initialHeuristic < 8; initialHeuristic++ {
		assigned := make([]bool, len(g.objects))
		values := make([]bool, len(g.objects))
		literals := make([]SatLit, 0)
		stack := []aigJustFrame{{id: root, desired: desired, heur: initialHeuristic}}
		result := false
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			if frame.stage != 0 {
				switch frame.stage {
				case 1:
					if !result {
						stack = stack[:len(stack)-1]
						continue
					}
					object := g.objects[frame.id]
					frame.stage = 2
					stack = append(stack, aigJustFrame{id: object.Fanin1.ID(), desired: !object.Fanin1.Complemented(), heur: frame.heur})
					continue
				case 2, 3:
					stack = stack[:len(stack)-1]
					continue
				}
			}
			if err := step(); err != nil {
				return nil, false, err
			}
			object := g.objects[frame.id]
			if object.Type == AigConstant1 {
				result = true
				stack = stack[:len(stack)-1]
				continue
			}
			if assigned[frame.id] {
				result = values[frame.id] == frame.desired
				stack = stack[:len(stack)-1]
				continue
			}
			assigned[frame.id], values[frame.id] = true, frame.desired
			if object.Type == AigCI {
				if len(literals) == options.MaxLiterals {
					return nil, false, ErrLimit
				}
				literal, _ := MakeSatLit(SatVar(ciIndex[frame.id]), !frame.desired)
				literals = append(literals, literal)
				result = true
				stack = stack[:len(stack)-1]
				continue
			}
			if object.Type != AigAND {
				return nil, false, ErrUnsupported
			}
			if frame.desired {
				frame.stage = 1
				stack = append(stack, aigJustFrame{id: object.Fanin0.ID(), desired: !object.Fanin0.Complemented(), heur: frame.heur})
				continue
			}
			read := func(literal AigLit) AigTernary {
				if !assigned[literal.ID()] {
					return AigTernaryUnknown
				}
				value := values[literal.ID()] != literal.Complemented()
				if value {
					return AigTernaryOne
				}
				return AigTernaryZero
			}
			left, right := read(object.Fanin0), read(object.Fanin1)
			switch {
			case left == AigTernaryZero || right == AigTernaryZero:
				result = true
				stack = stack[:len(stack)-1]
			case left == AigTernaryOne && right == AigTernaryOne:
				result = false
				stack = stack[:len(stack)-1]
			default:
				selected := object.Fanin0
				if left == AigTernaryOne {
					selected = object.Fanin1
				} else if right != AigTernaryOne {
					frame.heur++
					if frame.heur%8 == 0 {
						selected = object.Fanin1
					}
				}
				frame.stage = 3
				stack = append(stack, aigJustFrame{id: selected.ID(), desired: selected.Complemented(), heur: frame.heur})
			}
		}
		if result {
			return slices.Clone(literals), true, nil
		}
		lastLiterals = literals
	}
	return slices.Clone(lastLiterals), false, nil
}

// RunJustificationExperiment implements Aig_ManJustExperiment without its
// process-output side effects. It returns the candidate order, counters and
// final packed simulation state as owned values.
func (g *AigManager) RunJustificationExperiment(options AigJustificationExperimentOptions) (AigJustificationExperiment, error) {
	var result AigJustificationExperiment
	if options.MaxCandidates < 0 {
		return result, ErrDimension
	}
	if options.MaxCandidates == 0 {
		options.MaxCandidates = 1 << 20
	}
	pack, err := NewAigPatternPack(g, options.PatternPack)
	if err != nil {
		return result, err
	}
	result.Candidates, err = pack.RareNodes()
	if err != nil {
		return AigJustificationExperiment{}, err
	}
	if len(result.Candidates) > options.MaxCandidates {
		return AigJustificationExperiment{}, ErrLimit
	}
	for _, id := range result.Candidates {
		object := g.objects[id]
		desired := !object.Phase
		literals, found, err := g.FindJustification(id, desired, options.Justification)
		if err != nil {
			return AigJustificationExperiment{}, err
		}
		if !found {
			if desired {
				result.FailedOne++
			} else {
				result.FailedZero++
			}
			continue
		}
		if desired {
			result.CountOne++
		} else {
			result.CountZero++
		}
		result.TotalLiterals += len(literals)
		if _, err = pack.AddPattern(literals); err != nil {
			return AigJustificationExperiment{}, err
		}
	}
	if err = pack.Simulate(); err != nil {
		return AigJustificationExperiment{}, err
	}
	result.PatternStats, err = pack.Stats()
	if err != nil {
		return AigJustificationExperiment{}, err
	}
	result.Patterns = pack.InputPatterns()
	result.Care = pack.InputCare()
	result.Signs = pack.Signs()
	return result, nil
}

// TernarySimulate implements Aig_ObjTerSimulate with owned values. Assignment
// literals use CI ordinals; later duplicates replace earlier values, as in C.
func (g *AigManager) TernarySimulate(root AigID, assignment []SatLit, options AigJustificationOptions) (AigTernary, error) {
	options, err := checkAigJustification(g, root, options)
	if err != nil {
		return 0, err
	}
	values := make([]AigTernary, len(g.objects))
	seen := make([]bool, len(g.objects))
	values[0], seen[0] = AigTernaryOne, true
	for _, literal := range assignment {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(len(g.cis)) {
			return 0, ErrInvalidLiteral
		}
		value := AigTernaryOne
		if literal.Negative() {
			value = AigTernaryZero
		}
		id := g.cis[literal.Var()]
		values[id], seen[id] = value, true
	}
	visits := 0
	step := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if visits == options.MaxVisits {
			return ErrLimit
		}
		visits++
		return nil
	}
	read := func(literal AigLit) AigTernary {
		value := values[literal.ID()]
		if literal.Complemented() {
			return value.Not()
		}
		return value
	}
	type frame struct {
		id    AigID
		stage uint8
	}
	stack := []frame{{id: root}}
	for len(stack) != 0 {
		current := &stack[len(stack)-1]
		if seen[current.id] {
			stack = stack[:len(stack)-1]
			continue
		}
		object := g.objects[current.id]
		switch current.stage {
		case 0:
			if err := step(); err != nil {
				return 0, err
			}
			switch object.Type {
			case AigCI:
				values[current.id], seen[current.id] = AigTernaryUnknown, true
				stack = stack[:len(stack)-1]
			case AigAND, AigCO:
				current.stage = 1
				stack = append(stack, frame{id: object.Fanin0.ID()})
			default:
				return 0, ErrUnsupported
			}
		case 1:
			left := read(object.Fanin0)
			if object.Type == AigCO || left == AigTernaryZero {
				values[current.id], seen[current.id] = left, true
				stack = stack[:len(stack)-1]
				continue
			}
			current.stage = 2
			stack = append(stack, frame{id: object.Fanin1.ID()})
		case 2:
			values[current.id] = aigTernaryAnd(read(object.Fanin0), read(object.Fanin1))
			seen[current.id] = true
			stack = stack[:len(stack)-1]
		}
	}
	return values[root], nil
}
