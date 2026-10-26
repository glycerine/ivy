package abc

// Truth-table/CLOUD conversion from src/bool/kit/kitCloud.c.

var kitCompiledSourceInventory = [...]string{
	"kitAig.c",
	"kitBdd.c",
	"kitCloud.c",
	"cloud.c",
	"kitDsd.c",
	"kitFactor.c",
	"kitGraph.c",
	"kitHop.c",
	"kitIsop.c",
	"kitPla.c",
	"kitSop.c",
	"kitTruth.c",
}

func KitCompiledSourceInventory() []string {
	result := make([]string, len(kitCompiledSourceInventory))
	copy(result, kitCompiledSourceInventory[:])
	return result
}

// KitTruthToCloud derives a canonical complemented-edge Cloud BDD. The
// manager is rolled back exactly if validation, work limits, node limits or
// cancellation fail.
func KitTruthToCloud(manager *CloudManager, truth *TruthTable, options CloudOptions) (CloudEdge, error) {
	if manager == nil || truth == nil {
		return 0, ErrDimension
	}
	if err := manager.check(); err != nil {
		return 0, err
	}
	if err := truth.Check(); err != nil {
		return 0, err
	}
	if truth.variables != manager.variables || truth.variables > 20 {
		return 0, ErrDimension
	}
	normalized, err := normalizeCloudOptions(options)
	if err != nil {
		return 0, err
	}
	work := cloudWork{options: normalized}
	transaction := manager.beginTransaction()
	var build func(int, uint64) (CloudEdge, error)
	build = func(variable int, prefix uint64) (CloudEdge, error) {
		if err := work.charge(); err != nil {
			return 0, err
		}
		if variable == truth.variables {
			if truth.words[prefix/64]&(uint64(1)<<uint(prefix%64)) != 0 {
				return manager.One(), nil
			}
			return manager.Zero(), nil
		}
		zero, err := build(variable+1, prefix)
		if err != nil {
			return 0, err
		}
		one, err := build(variable+1, prefix|(uint64(1)<<uint(variable)))
		if err != nil {
			return 0, err
		}
		if zero == one {
			return zero, nil
		}
		return manager.MakeNode(variable, one, zero)
	}
	root, err := build(0, 0)
	if err != nil {
		manager.rollback(transaction)
		return 0, err
	}
	return root, nil
}

func kitCloudEvaluate(manager *CloudManager, root CloudEdge, assignment uint64, work *cloudWork) (bool, error) {
	current := root
	for current.nodeID() != 0 {
		if err := work.charge(); err != nil {
			return false, err
		}
		node := manager.nodes[current.nodeID()]
		then, otherwise := manager.cofactors(current)
		if assignment&(uint64(1)<<uint(node.variable)) != 0 {
			current = then
		} else {
			current = otherwise
		}
	}
	if err := work.charge(); err != nil {
		return false, err
	}
	return !current.Complemented(), nil
}

// KitCloudToTruth reconstructs the represented function without retaining
// manager storage. It is the semantic counterpart of Kit_CloudToTruth's packed
// mux replay.
func KitCloudToTruth(manager *CloudManager, root CloudEdge, maxWords int, options CloudOptions) (*TruthTable, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	if err := manager.checkEdge(root); err != nil {
		return nil, err
	}
	if manager.variables > 20 {
		return nil, ErrDimension
	}
	normalized, err := normalizeCloudOptions(options)
	if err != nil {
		return nil, err
	}
	result, err := NewConstantTruthTable(manager.variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	work := cloudWork{options: normalized}
	assignments := uint64(1) << uint(manager.variables)
	for assignment := uint64(0); assignment < assignments; assignment++ {
		value, evalErr := kitCloudEvaluate(manager, root, assignment, &work)
		if evalErr != nil {
			return nil, evalErr
		}
		if value {
			result.words[assignment/64] |= uint64(1) << uint(assignment%64)
		}
	}
	if manager.variables < 6 {
		period := uint64(1) << uint(manager.variables)
		for bit := period; bit < 64; bit++ {
			if result.words[0]&(uint64(1)<<uint(bit%period)) != 0 {
				result.words[0] |= uint64(1) << uint(bit)
			}
		}
	}
	if err = result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}
