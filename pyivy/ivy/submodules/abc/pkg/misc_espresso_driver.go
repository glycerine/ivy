package abc

// Bounded driver argument parsing from src/misc/espresso/main.c.

import (
	"slices"
	"strconv"
	"strings"
)

type EspressoDriverParseOptions struct {
	MaxArguments, MaxBytes int
	Stop                   func() bool
}

type EspressoDriverConfig struct {
	Command               EspressoCommandDescriptor
	OutputType, InputType EspressoPLAOutputType
	Runtime               EspressoRuntimeOptions
	Strategy              int
	First, Last           int
	Inputs                []string
}

func normalizeEspressoDriverParseOptions(options EspressoDriverParseOptions) (EspressoDriverParseOptions, error) {
	if options.MaxArguments < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxArguments == 0 {
		options.MaxArguments = 1 << 16
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 16 << 20
	}
	return options, nil
}

func espressoDriverTakeValue(args []string, index int, short byte) (string, int, bool) {
	argument := args[index]
	prefix := "-" + string(short)
	if argument == prefix {
		if index+1 >= len(args) {
			return "", index, false
		}
		return args[index+1], index + 1, true
	}
	if strings.HasPrefix(argument, prefix) && len(argument) > len(prefix) {
		return argument[len(prefix):], index, true
	}
	return "", index, false
}

func espressoDriverRemovePair(args []string, option string, apply func(string) error) ([]string, error) {
	for index := 0; index+1 < len(args); index++ {
		if args[index] != option {
			continue
		}
		if err := apply(args[index+1]); err != nil {
			return nil, err
		}
		return append(slices.Clone(args[:index]), args[index+2:]...), nil
	}
	return args, nil
}

// ParseEspressoDriverArguments parses arguments after argv[0]. It supports the
// source's legacy -do/-out/bare behavior spellings and its getopt spellings.
func ParseEspressoDriverArguments(arguments []string, options EspressoDriverParseOptions) (EspressoDriverConfig, error) {
	options, err := normalizeEspressoDriverParseOptions(options)
	if err != nil {
		return EspressoDriverConfig{}, err
	}
	if len(arguments) > options.MaxArguments {
		return EspressoDriverConfig{}, ErrLimit
	}
	bytes := 0
	for _, argument := range arguments {
		if len(argument) > options.MaxBytes-bytes {
			return EspressoDriverConfig{}, ErrLimit
		}
		bytes += len(argument)
	}
	if options.Stop != nil && options.Stop() {
		return EspressoDriverConfig{}, ErrCancelled
	}
	command, _ := LookupEspressoCommand("ESPRESSO")
	config := EspressoDriverConfig{
		Command: command, OutputType: EspressoPLAOutputOn,
		InputType: EspressoPLAOutputOn | EspressoPLAOutputDontCare,
		Runtime:   DefaultEspressoRuntimeOptions(), First: -1, Last: -1,
	}
	args := slices.Clone(arguments)
	args, err = espressoDriverRemovePair(args, "-do", func(name string) error {
		descriptor, ok := LookupEspressoCommand(name)
		if !ok {
			return ErrUnsupported
		}
		config.Command = descriptor
		return nil
	})
	if err != nil {
		return EspressoDriverConfig{}, err
	}
	args, err = espressoDriverRemovePair(args, "-out", func(name string) error {
		value, ok := LookupEspressoPLAType(name)
		if !ok {
			return ErrUnsupported
		}
		config.OutputType = value
		return nil
	})
	if err != nil {
		return EspressoDriverConfig{}, err
	}

	filtered := make([]string, 0, len(args))
	inputFDR, inputFR, inputF := false, false, false
	for _, argument := range args {
		switch argument {
		case "-fdr":
			inputFDR = true
			continue
		case "-fr":
			inputFR = true
			continue
		case "-f":
			inputF = true
			continue
		}
		if strings.HasPrefix(argument, "-") && len(argument) > 1 {
			updated, applyErr := ApplyEspressoBehaviorOption(config.Runtime, argument[1:])
			if applyErr == nil {
				config.Runtime = updated
				continue
			}
		}
		filtered = append(filtered, argument)
	}
	if inputFDR {
		config.InputType = EspressoPLAOutputOn | EspressoPLAOutputDontCare | EspressoPLAOutputOff
	}
	if inputFR {
		config.InputType = EspressoPLAOutputOn | EspressoPLAOutputOff
	}
	if inputF {
		config.InputType = EspressoPLAOutputOn
	}
	args = filtered

	for index := 0; index < len(args); index++ {
		if options.Stop != nil && options.Stop() {
			return EspressoDriverConfig{}, ErrCancelled
		}
		argument := args[index]
		if argument == "-d" {
			config.Runtime.Debug, _ = LookupEspressoDebug("")
			config.Runtime.Trace, config.Runtime.Summary = true, true
			continue
		}
		if argument == "-t" {
			config.Runtime.Trace = true
			continue
		}
		if argument == "-s" {
			config.Runtime.Summary = true
			continue
		}
		if argument == "-x" {
			config.Runtime.PrintSolution = false
			continue
		}
		if value, consumed, ok := espressoDriverTakeValue(args, index, 'D'); ok {
			descriptor, found := LookupEspressoCommand(value)
			if !found {
				return EspressoDriverConfig{}, ErrUnsupported
			}
			config.Command, index = descriptor, consumed
			continue
		}
		if value, consumed, ok := espressoDriverTakeValue(args, index, 'o'); ok {
			kind, found := LookupEspressoPLAType(value)
			if !found {
				return EspressoDriverConfig{}, ErrUnsupported
			}
			config.OutputType, index = kind, consumed
			continue
		}
		if value, consumed, ok := espressoDriverTakeValue(args, index, 'e'); ok {
			updated, applyErr := ApplyEspressoBehaviorOption(config.Runtime, value)
			if applyErr != nil {
				return EspressoDriverConfig{}, applyErr
			}
			config.Runtime, index = updated, consumed
			continue
		}
		if value, consumed, ok := espressoDriverTakeValue(args, index, 'v'); ok {
			mask, found := LookupEspressoDebug(value)
			if !found {
				return EspressoDriverConfig{}, ErrUnsupported
			}
			config.Runtime.VerboseDebug = true
			config.Runtime.Debug |= mask
			index = consumed
			continue
		}
		if value, consumed, ok := espressoDriverTakeValue(args, index, 'S'); ok {
			strategy, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return EspressoDriverConfig{}, ErrDimension
			}
			config.Strategy, index = strategy, consumed
			continue
		}
		if value, consumed, ok := espressoDriverTakeValue(args, index, 'r'); ok {
			separator := strings.IndexByte(value, '-')
			if separator <= 0 || separator == len(value)-1 {
				return EspressoDriverConfig{}, ErrDimension
			}
			first, firstErr := strconv.Atoi(value[:separator])
			last, lastErr := strconv.Atoi(value[separator+1:])
			if firstErr != nil || lastErr != nil {
				return EspressoDriverConfig{}, ErrDimension
			}
			config.First, config.Last, index = first, last, consumed
			continue
		}
		if strings.HasPrefix(argument, "-") && argument != "-" {
			return EspressoDriverConfig{}, ErrUnsupported
		}
		config.Inputs = append(config.Inputs, argument)
	}
	if len(config.Inputs) > config.Command.InputPLAs {
		return EspressoDriverConfig{}, ErrDimension
	}
	config.Inputs = slices.Clone(config.Inputs)
	return config, nil
}

func EspressoDriverUsage() string {
	var builder strings.Builder
	builder.WriteString("Espresso 2.3\n\nSYNOPSIS: espresso [options] [file]\n\n")
	builder.WriteString("  -e[opt]   fast, ness, nirr, nunwrap, onset, pos, strong,\n")
	builder.WriteString("             eat, eatdots, kiss, random\n")
	builder.WriteString("  -o[type]  f, fd, fr, fdr, pleasure, eqntott, kiss, cons\n")
	builder.WriteString("  -D[cmd]   ")
	column := 13
	for index, descriptor := range espressoCommandDescriptors {
		separator := ""
		if index != 0 {
			separator = ", "
		}
		if column+len(separator)+len(descriptor.Name) > 75 {
			builder.WriteString(",\n             ")
			column, separator = 13, ""
		}
		builder.WriteString(separator)
		builder.WriteString(descriptor.Name)
		column += len(separator) + len(descriptor.Name)
	}
	builder.WriteByte('\n')
	return builder.String()
}
