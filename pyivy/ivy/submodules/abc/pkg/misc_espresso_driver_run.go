package abc

// Side-effect-free end-to-end driver path from src/misc/espresso/main.c.

type EspressoDriverRunOptions struct {
	Parse          EspressoDriverParseOptions
	Input          EspressoPLAInputOptions
	Execution      EspressoDriverExecutionOptions
	Output         EspressoPLAOutputOptions
	MaxOutputBytes int
}

type EspressoDriverRunResult struct {
	Config     EspressoDriverConfig
	Execution  EspressoDriverExecution
	Executions []EspressoDriverExecution
	Output     []byte
}

func espressoDriverPrintsPLA(kind EspressoCommandKind) bool {
	switch kind {
	case EspressoCommandCheck, EspressoCommandStats, EspressoCommandVerify,
		EspressoCommandPLAVerify, EspressoCommandEquivalentOutputs,
		EspressoCommandMap, EspressoCommandFSM, EspressoCommandTautology:
		return false
	default:
		return true
	}
}

func espressoDriverInputOptions(config EspressoDriverConfig, run EspressoDriverRunOptions, index int) EspressoPLAInputOptions {
	options := run.Input
	options.Type = config.InputType
	options.Positive = config.Runtime.Positive
	if config.Command.Kind == EspressoCommandEcho {
		options.NeedsDontCare = config.OutputType&EspressoPLAOutputDontCare != 0
		options.NeedsOff = config.OutputType&EspressoPLAOutputOff != 0
	} else {
		options.NeedsDontCare = config.Command.NeedsDontCare
		options.NeedsOff = config.Command.NeedsOff
	}
	if index < len(config.Inputs) {
		options.Filename = config.Inputs[index]
	} else {
		options.Filename = "(stdin)"
	}
	return options
}

// RunEspressoDriver parses arguments, loads already-owned PLA bytes, executes
// one command and returns owned output bytes. The caller handles files/stdin.
func RunEspressoDriver(arguments []string, inputs [][]byte, options EspressoDriverRunOptions) (EspressoDriverRunResult, error) {
	config, err := ParseEspressoDriverArguments(arguments, options.Parse)
	if err != nil {
		return EspressoDriverRunResult{}, err
	}
	if options.MaxOutputBytes < 0 {
		return EspressoDriverRunResult{}, ErrDimension
	}
	if options.MaxOutputBytes == 0 {
		options.MaxOutputBytes = 64 << 20
	}
	if config.Command.Kind == EspressoCommandMany {
		return runManyEspressoDriver(config, inputs, options)
	}
	if len(inputs) != config.Command.InputPLAs {
		return EspressoDriverRunResult{}, ErrDimension
	}
	plas := make([]*EspressoPLA, len(inputs))
	for index, data := range inputs {
		parsed, parseErr := ParseEspressoPLA(data, espressoDriverInputOptions(config, options, index))
		if parseErr != nil {
			return EspressoDriverRunResult{}, parseErr
		}
		plas[index] = parsed.PLA
	}
	var second *EspressoPLA
	if len(plas) == 2 {
		second = plas[1]
	}
	execution, err := ExecuteEspressoCommand(config, plas[0], second, options.Execution)
	if err != nil {
		return EspressoDriverRunResult{}, err
	}
	result := EspressoDriverRunResult{Config: config, Execution: execution, Executions: []EspressoDriverExecution{execution}}
	if config.Runtime.PrintSolution && espressoDriverPrintsPLA(config.Command.Kind) {
		if execution.PLA == nil {
			return EspressoDriverRunResult{}, ErrInvariant
		}
		result.Output, err = execution.PLA.MarshalPLA(config.OutputType, options.Output)
		if err != nil {
			return EspressoDriverRunResult{}, err
		}
		if len(result.Output) > options.MaxOutputBytes {
			return EspressoDriverRunResult{}, ErrLimit
		}
	}
	return result, nil
}

func runManyEspressoDriver(config EspressoDriverConfig, inputs [][]byte, options EspressoDriverRunOptions) (EspressoDriverRunResult, error) {
	if len(inputs) == 0 {
		return EspressoDriverRunResult{}, ErrDimension
	}
	minimize, _ := LookupEspressoCommand("ESPRESSO")
	iterationConfig := config
	iterationConfig.Command = minimize
	result := EspressoDriverRunResult{Config: config, Executions: make([]EspressoDriverExecution, 0, len(inputs))}
	for _, data := range inputs {
		if options.Execution.Stop != nil && options.Execution.Stop() {
			return EspressoDriverRunResult{}, ErrCancelled
		}
		parsed, err := ParseEspressoPLA(data, espressoDriverInputOptions(iterationConfig, options, 0))
		if err != nil {
			return EspressoDriverRunResult{}, err
		}
		execution, err := ExecuteEspressoCommand(iterationConfig, parsed.PLA, nil, options.Execution)
		if err != nil {
			return EspressoDriverRunResult{}, err
		}
		result.Executions = append(result.Executions, execution)
		result.Execution = execution
		if config.Runtime.PrintSolution {
			raw, marshalErr := execution.PLA.MarshalPLA(config.OutputType, options.Output)
			if marshalErr != nil {
				return EspressoDriverRunResult{}, marshalErr
			}
			if len(raw) > options.MaxOutputBytes-len(result.Output) {
				return EspressoDriverRunResult{}, ErrLimit
			}
			result.Output = append(result.Output, raw...)
		}
	}
	return result, nil
}
