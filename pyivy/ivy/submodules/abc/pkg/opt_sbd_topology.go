package abc

// Deterministic topology counting from src/opt/sbd/sbd.c.

import (
	"math"
	"slices"
	"strconv"

	"github.com/glycerine/insdict"
)

type SBDTopologyLimits struct {
	MaxLevels, MaxConfigurations, MaxLevelEntries, MaxWork int
	Stop                                                   func() bool
}

type SBDTopologyConfiguration struct {
	Levels                 []int
	ConfigurationVariables int
}

type SBDTopologyStage struct {
	Nodes          int
	New, Total     int
	Configurations []SBDTopologyConfiguration
}

type SBDTopologyResult struct {
	Inputs, Nodes, Degree int
	Stages                []SBDTopologyStage
}

type sbdTopologyBudget struct {
	limits SBDTopologyLimits
	work   int
}

func normalizeSBDTopologyLimits(limits SBDTopologyLimits) (SBDTopologyLimits, error) {
	if limits.MaxLevels < 0 || limits.MaxConfigurations < 0 || limits.MaxLevelEntries < 0 || limits.MaxWork < 0 {
		return limits, ErrDimension
	}
	if limits.MaxLevels == 0 {
		limits.MaxLevels = 1 << 16
	}
	if limits.MaxConfigurations == 0 {
		limits.MaxConfigurations = 1 << 20
	}
	if limits.MaxLevelEntries == 0 {
		limits.MaxLevelEntries = 1 << 24
	}
	if limits.MaxWork == 0 {
		limits.MaxWork = 1 << 26
	}
	return limits, nil
}

func newSBDTopologyBudget(limits SBDTopologyLimits) (*sbdTopologyBudget, error) {
	normalized, err := normalizeSBDTopologyLimits(limits)
	if err != nil {
		return nil, err
	}
	return &sbdTopologyBudget{limits: normalized}, nil
}

func (budget *sbdTopologyBudget) step(amount int) error {
	if amount < 0 || budget.work > budget.limits.MaxWork-amount {
		return ErrLimit
	}
	budget.work += amount
	if budget.limits.Stop != nil && budget.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func sbdCheckedAdd(left, right int) (int, error) {
	if left < 0 || right < 0 || left > math.MaxInt-right {
		return 0, ErrLimit
	}
	return left + right, nil
}

func sbdCheckedMultiply(left, right int) (int, error) {
	if left < 0 || right < 0 || right != 0 && left > math.MaxInt/right {
		return 0, ErrLimit
	}
	return left * right, nil
}

func sbdCountConfigurationVariables(levels []int, inputs, degree int, budget *sbdTopologyBudget) (int, error) {
	if len(levels) == 0 || inputs < 1 || degree < 1 {
		return 0, ErrDimension
	}
	if len(levels) > budget.limits.MaxLevels {
		return 0, ErrLimit
	}
	total := 0
	for _, entry := range levels {
		if entry < 0 {
			return 0, ErrDimension
		}
		var err error
		total, err = sbdCheckedAdd(total, entry)
		if err != nil {
			return 0, err
		}
		if err = budget.step(1); err != nil {
			return 0, err
		}
	}

	previous := levels[0]
	suffix := total - previous
	count := 0
	for index := 1; index < len(levels); index++ {
		entry := levels[index]
		maximum, err := sbdCheckedMultiply(degree, previous)
		if err != nil {
			return 0, err
		}
		if maximum < entry {
			return 0, ErrDimension
		}
		if maximum != entry {
			below, addErr := sbdCheckedAdd(inputs, suffix)
			if addErr != nil {
				return 0, addErr
			}
			term, multiplyErr := sbdCheckedMultiply(below, maximum-1)
			if multiplyErr != nil {
				return 0, multiplyErr
			}
			count, addErr = sbdCheckedAdd(count, term)
			if addErr != nil {
				return 0, addErr
			}
		}
		previous = entry
		suffix -= entry
		if err = budget.step(1); err != nil {
			return 0, err
		}
	}
	last, err := sbdCheckedMultiply(degree, previous)
	if err != nil {
		return 0, err
	}
	last, err = sbdCheckedMultiply(inputs, last)
	if err != nil {
		return 0, err
	}
	count, err = sbdCheckedAdd(count, last)
	if err != nil {
		return 0, err
	}
	if total < inputs-1 {
		return 0, nil
	}
	return count, nil
}

// SBDCountConfigurationVariables is the checked, non-mutating form of
// Sbd_CountConfigVars. Levels retain the source's root-to-leaf level order.
func SBDCountConfigurationVariables(levels []int, inputs, degree int, limits SBDTopologyLimits) (int, error) {
	budget, err := newSBDTopologyBudget(limits)
	if err != nil {
		return 0, err
	}
	return sbdCountConfigurationVariables(levels, inputs, degree, budget)
}

func sbdTopologyKey(levels []int) string {
	key := make([]byte, 0, len(levels)*3)
	for _, entry := range levels {
		key = strconv.AppendInt(key, int64(entry), 10)
		key = append(key, ',')
	}
	return string(key)
}

// SBDCountTopologies generalizes the source's fixed 9-input, 10-node,
// degree-3 diagnostic and returns its insertion-ordered output as owned data.
func SBDCountTopologies(inputs, nodes, degree int, limits SBDTopologyLimits) (SBDTopologyResult, error) {
	var result SBDTopologyResult
	budget, err := newSBDTopologyBudget(limits)
	if err != nil {
		return result, err
	}
	if inputs < 1 || nodes < 0 || degree < 1 {
		return result, ErrDimension
	}
	if nodes > budget.limits.MaxLevels {
		return result, ErrLimit
	}
	if err = budget.step(1); err != nil {
		return result, err
	}
	result = SBDTopologyResult{Inputs: inputs, Nodes: nodes, Degree: degree, Stages: make([]SBDTopologyStage, 0, nodes)}
	index := insdict.NewDictSize[string, int](min(budget.limits.MaxConfigurations, 1024))
	index.Put("", 0)
	configurations := [][]int{nil}
	retainedEntries := 0
	start := 0

	add := func(levels []int) error {
		key := sbdTopologyKey(levels)
		if _, found := index.Get2(key); found {
			return nil
		}
		if len(configurations) >= budget.limits.MaxConfigurations || len(levels) > budget.limits.MaxLevelEntries-retainedEntries {
			return ErrLimit
		}
		owned := slices.Clone(levels)
		index.Put(key, len(configurations))
		configurations = append(configurations, owned)
		retainedEntries += len(owned)
		return nil
	}

	for node := 1; node <= nodes; node++ {
		stop := len(configurations)
		for entryIndex := start; entryIndex < stop; entryIndex++ {
			working := slices.Clone(configurations[entryIndex])
			for level := range working {
				if err = budget.step(1); err != nil {
					return SBDTopologyResult{}, err
				}
				if level == 0 {
					if working[0] > 0 {
						continue
					}
				} else {
					maximum, multiplyErr := sbdCheckedMultiply(degree, working[level-1])
					if multiplyErr != nil {
						return SBDTopologyResult{}, multiplyErr
					}
					if working[level] == maximum {
						continue
					}
				}
				if working[level] == math.MaxInt {
					return SBDTopologyResult{}, ErrLimit
				}
				working[level]++
				if err = add(working); err != nil {
					return SBDTopologyResult{}, err
				}
				working[level]--
			}
			if len(working) >= budget.limits.MaxLevels {
				return SBDTopologyResult{}, ErrLimit
			}
			working = append(working, 1)
			if err = add(working); err != nil {
				return SBDTopologyResult{}, err
			}
		}

		stage := SBDTopologyStage{Nodes: node, New: len(configurations) - stop, Total: len(configurations)}
		stage.Configurations = make([]SBDTopologyConfiguration, 0, stage.New)
		for configuration := stop; configuration < len(configurations); configuration++ {
			count, countErr := sbdCountConfigurationVariables(configurations[configuration], inputs, degree, budget)
			if countErr != nil {
				return SBDTopologyResult{}, countErr
			}
			stage.Configurations = append(stage.Configurations, SBDTopologyConfiguration{
				Levels: slices.Clone(configurations[configuration]), ConfigurationVariables: count,
			})
		}
		result.Stages = append(result.Stages, stage)
		start = stop
	}
	if err = budget.step(1); err != nil {
		return SBDTopologyResult{}, err
	}
	return result, nil
}
