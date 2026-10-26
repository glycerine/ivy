package abc

// Owned replacement for the fixed diagnostic in src/aig/aig/aigTest.c.

// BuildAigProcedureTest constructs the native diagnostic's three-input,
// two-output, one-register graph. The caller owns the checked result; no fixed
// filename is written as a side effect.
func BuildAigProcedureTest() (*AigManager, error) {
	graph := NewAigManager()
	inputs := make([]AigLit, 3)
	for index := range inputs {
		var err error
		inputs[index], err = graph.IthVar(index)
		if err != nil {
			return nil, err
		}
	}
	mux, err := graph.Mux(inputs[0], inputs[1], inputs[2])
	if err != nil {
		return nil, err
	}
	and, err := graph.And(inputs[0], inputs[1])
	if err != nil {
		return nil, err
	}
	if _, err = graph.CreateCO(mux); err != nil {
		return nil, err
	}
	if _, err = graph.CreateCO(and); err != nil {
		return nil, err
	}
	if err = graph.SetRegisterCount(1); err != nil {
		return nil, err
	}
	graph.Cleanup()
	if err = graph.Check(); err != nil {
		return nil, err
	}
	return graph, nil
}
