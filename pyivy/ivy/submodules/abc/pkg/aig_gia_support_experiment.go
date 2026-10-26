package abc

// Support/cone pair experiments from src/aig/gia/giaSupMin.c.

type GiaSupportExperimentOptions struct {
	MaxObjects int
	MaxPairs   int
	Stop       func() bool
}

type GiaSupportPairAnalysis struct {
	DataOutput, CareOutput int
	CareIsAnd              bool
	Data, Care             GiaConeStatistics
}

func checkGiaSupportExperimentOptions(options GiaSupportExperimentOptions) (GiaSupportExperimentOptions, error) {
	if options.MaxObjects < 0 || options.MaxPairs < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxPairs == 0 {
		options.MaxPairs = 1 << 20
	}
	return options, nil
}

func (manager *GiaManager) supportExperimentStatistics(root GiaID, options GiaSupportExperimentOptions) (GiaConeStatistics, error) {
	result := GiaConeStatistics{}
	seen := make([]bool, len(manager.objects))
	seen[0] = true
	stack := []GiaID{root}
	for len(stack) != 0 {
		if options.Stop != nil && options.Stop() {
			return GiaConeStatistics{}, ErrCancelled
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		object := manager.objects[id]
		if object.IsCI() {
			result.Inputs++
			continue
		}
		if object.IsAnd() {
			result.Gates++
		}
		fanins, count, err := manager.crossCutFanins(id)
		if err != nil {
			return GiaConeStatistics{}, err
		}
		for index := count - 1; index >= 0; index-- {
			stack = append(stack, fanins[index])
		}
	}
	return result, nil
}

// AnalyzeSupportPairs replaces Gia_ManSupStart/Stop and the two printing
// experiments with one owned operation. Pairs contain output indices. An
// invalid care shape is reported per pair, as the source prints a message and
// continues with later pairs.
func (manager *GiaManager) AnalyzeSupportPairs(pairs [][2]int, options GiaSupportExperimentOptions) ([]GiaSupportPairAnalysis, error) {
	options, err := checkGiaSupportExperimentOptions(options)
	if err != nil {
		return nil, err
	}
	if manager == nil {
		return nil, ErrDimension
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	if len(manager.objects) > options.MaxObjects || len(pairs) > options.MaxPairs {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	result := make([]GiaSupportPairAnalysis, len(pairs))
	for index, pair := range pairs {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if pair[0] < 0 || pair[0] >= len(manager.cos) || pair[1] < 0 || pair[1] >= len(manager.cos) {
			return nil, ErrDimension
		}
		dataObject := manager.objects[manager.cos[pair[0]]]
		careObject := manager.objects[manager.cos[pair[1]]]
		dataDriver, dataErr := dataObject.Fanin0(manager.cos[pair[0]])
		if dataErr != nil {
			return nil, dataErr
		}
		careDriver, careErr := careObject.Fanin0(manager.cos[pair[1]])
		if careErr != nil {
			return nil, careErr
		}
		record := GiaSupportPairAnalysis{DataOutput: pair[0], CareOutput: pair[1]}
		record.CareIsAnd = manager.objects[careDriver.ID()].IsAnd()
		if !record.CareIsAnd {
			result[index] = record
			continue
		}
		record.Data, err = manager.supportExperimentStatistics(dataDriver.ID(), options)
		if err != nil {
			return nil, err
		}
		record.Care, err = manager.supportExperimentStatistics(careDriver.ID(), options)
		if err != nil {
			return nil, err
		}
		result[index] = record
	}
	return result, nil
}
