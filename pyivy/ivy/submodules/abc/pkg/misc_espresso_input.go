package abc

// Bounded PLA parsing from src/misc/espresso/cvrin.c.

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

var errEspressoProductIncomplete = errors.New("incomplete PLA product")

type EspressoSymbolicReference struct {
	Variable int
	Part     int
}

type EspressoSymbolicDirective struct {
	References []EspressoSymbolicReference
	Labels     []string
}

type EspressoPLAInputOptions struct {
	Type                              EspressoPLAOutputType
	NeedsDontCare, NeedsOff, Positive bool
	DeferSymbolicMappings             bool
	Filename                          string
	MaxBytes, MaxLines, MaxVariables  int
	MaxParts, MaxCubes, MaxTokens     int
	MaxOperations, MaxDepth           int
	Stop                              func() bool
}

type EspressoPLAParseResult struct {
	PLA                         *EspressoPLA
	Phase                       []bool
	Pairing                     EspressoPairing
	SymbolicInputs              []EspressoSymbolicDirective
	SymbolicOutputs             []EspressoSymbolicDirective
	Comments, Unknown, Warnings []string
	Lines, Products             int
	KISS                        bool
}

type espressoPLAParser struct {
	options                   EspressoPLAInputOptions
	result                    EspressoPLAParseResult
	layout                    *EspressoCubeLayout
	on, dc, off               *EspressoSetFamily
	labels                    []string
	typeBits                  EspressoPLAOutputType
	pendingBinary             int
	operations, tokens, cubes int
}

func normalizeEspressoPLAInputOptions(options EspressoPLAInputOptions) (EspressoPLAInputOptions, error) {
	if options.MaxBytes < 0 || options.MaxLines < 0 || options.MaxVariables < 0 || options.MaxParts < 0 || options.MaxCubes < 0 || options.MaxTokens < 0 || options.MaxOperations < 0 || options.MaxDepth < 0 {
		return options, ErrDimension
	}
	if options.Type == 0 {
		options.Type = EspressoPLAOutputOn
	}
	if options.Type&^(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff) != 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	if options.MaxLines == 0 {
		options.MaxLines = 1 << 20
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxParts == 0 {
		options.MaxParts = 16 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxTokens == 0 {
		options.MaxTokens = 16 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 64 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	return options, nil
}

func (parser *espressoPLAParser) step(amount int) error {
	if amount < 0 || parser.operations > parser.options.MaxOperations-amount {
		return ErrLimit
	}
	parser.operations += amount
	if parser.options.Stop != nil && parser.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (parser *espressoPLAParser) token() error {
	parser.tokens++
	if parser.tokens > parser.options.MaxTokens {
		return ErrLimit
	}
	return parser.step(1)
}

func (parser *espressoPLAParser) setLayout(partSizes []int, binary int, symbolicSource bool) error {
	if parser.layout != nil || binary < 0 || binary > len(partSizes) {
		return ErrDimension
	}
	if len(partSizes) > parser.options.MaxVariables {
		return ErrLimit
	}
	total := 0
	for _, size := range partSizes {
		absolute := size
		if absolute < 0 {
			absolute = -absolute
		}
		if absolute <= 0 || absolute > parser.options.MaxParts-total {
			return ErrLimit
		}
		total += absolute
	}
	var layout *EspressoCubeLayout
	var err error
	if symbolicSource {
		layout, err = NewEspressoCubeLayoutFromSource(partSizes, binary)
	} else {
		layout, err = NewEspressoCubeLayout(partSizes, binary)
	}
	if err != nil {
		return err
	}
	parser.layout = layout
	parser.labels = make([]string, layout.bitCount)
	parser.on, _ = NewEspressoSetFamily(0, layout.bitCount)
	parser.dc, _ = NewEspressoSetFamily(0, layout.bitCount)
	parser.off, _ = NewEspressoSetFamily(0, layout.bitCount)
	return nil
}

func espressoPLAType(text string) (EspressoPLAOutputType, bool) {
	var result EspressoPLAOutputType
	for _, letter := range text {
		switch letter {
		case 'f':
			result |= EspressoPLAOutputOn
		case 'd':
			result |= EspressoPLAOutputDontCare
		case 'r':
			result |= EspressoPLAOutputOff
		default:
			return 0, false
		}
	}
	return result, result != 0
}

func espressoFields(text string) []string {
	return strings.Fields(strings.ReplaceAll(strings.ReplaceAll(text, "(", " ("), ")", ") "))
}

func parseEspressoInt(text string) (int, error) {
	value, err := strconv.Atoi(text)
	if err != nil || value < 0 {
		return 0, ErrDimension
	}
	return value, nil
}

func (parser *espressoPLAParser) requireLayout() error {
	if parser.layout == nil {
		return ErrDimension
	}
	return nil
}

func (parser *espressoPLAParser) labelIndex(word string) (EspressoSymbolicReference, bool) {
	if parser.layout == nil {
		return EspressoSymbolicReference{}, false
	}
	if len(parser.labels) == 0 || parser.labels[0] == "" {
		value, err := strconv.Atoi(word)
		if err == nil && value >= 0 {
			return EspressoSymbolicReference{Variable: value, Part: value}, true
		}
		return EspressoSymbolicReference{}, false
	}
	for variable, size := range parser.layout.partSizes {
		for part := 0; part < size; part++ {
			if parser.labels[parser.layout.firstParts[variable]+part] == word {
				return EspressoSymbolicReference{Variable: variable, Part: part}, true
			}
		}
	}
	return EspressoSymbolicReference{}, false
}

func (parser *espressoPLAParser) parseSymbolic(fields []string) (EspressoSymbolicDirective, error) {
	directive := EspressoSymbolicDirective{}
	separator := slices.Index(fields, ";")
	if separator < 0 {
		return directive, ErrDimension
	}
	last := slices.Index(fields[separator+1:], ";")
	if last < 0 {
		return directive, ErrDimension
	}
	last += separator + 1
	if separator == 0 {
		return directive, ErrDimension
	}
	for _, field := range fields[:separator] {
		if err := parser.token(); err != nil {
			return EspressoSymbolicDirective{}, err
		}
		reference, ok := parser.labelIndex(field)
		if !ok {
			return EspressoSymbolicDirective{}, ErrDimension
		}
		directive.References = append(directive.References, reference)
	}
	for _, field := range fields[separator+1 : last] {
		if err := parser.token(); err != nil {
			return EspressoSymbolicDirective{}, err
		}
		directive.Labels = append(directive.Labels, field)
	}
	return directive, nil
}

func (parser *espressoPLAParser) parseDirective(line string) (done bool, err error) {
	fields := espressoFields(strings.TrimSpace(line[1:]))
	if len(fields) == 0 {
		return false, ErrDimension
	}
	if err = parser.token(); err != nil {
		return false, err
	}
	command := fields[0]
	args := fields[1:]
	switch command {
	case "i":
		if len(args) != 1 || parser.layout != nil || parser.pendingBinary >= 0 {
			return false, ErrDimension
		}
		parser.pendingBinary, err = parseEspressoInt(args[0])
		return false, err
	case "o":
		if len(args) != 1 || parser.pendingBinary < 0 || parser.layout != nil {
			return false, ErrDimension
		}
		outputs, parseErr := parseEspressoInt(args[0])
		if parseErr != nil || outputs == 0 {
			return false, ErrDimension
		}
		parts := make([]int, parser.pendingBinary+1)
		for index := 0; index < parser.pendingBinary; index++ {
			parts[index] = 2
		}
		parts[len(parts)-1] = outputs
		return false, parser.setLayout(parts, parser.pendingBinary, false)
	case "mv":
		if len(args) < 2 || parser.pendingBinary >= 0 || parser.layout != nil {
			return false, ErrDimension
		}
		variables, parseErr := parseEspressoInt(args[0])
		if parseErr != nil {
			return false, parseErr
		}
		binary, parseErr := parseEspressoInt(args[1])
		if parseErr != nil || binary > variables || len(args) != 2+variables-binary {
			return false, ErrDimension
		}
		parts := make([]int, variables)
		for index := 0; index < binary; index++ {
			parts[index] = 2
		}
		for index := binary; index < variables; index++ {
			value, conversionErr := strconv.Atoi(args[2+index-binary])
			if conversionErr != nil || value == 0 {
				return false, ErrDimension
			}
			parts[index] = value
		}
		return false, parser.setLayout(parts, binary, true)
	case "p":
		if len(args) != 1 {
			return false, ErrDimension
		}
		_, err = parseEspressoInt(args[0])
		return false, err
	case "e", "end":
		return true, nil
	case "kiss":
		parser.result.KISS = true
		return false, nil
	case "type":
		if len(args) != 1 {
			return false, ErrDimension
		}
		value, ok := espressoPLAType(args[0])
		if !ok {
			return false, ErrDimension
		}
		parser.typeBits = value
		return false, nil
	case "ilb":
		if err = parser.requireLayout(); err != nil || len(args) != parser.layout.binaryVariables {
			return false, ErrDimension
		}
		for variable, label := range args {
			if err = parser.token(); err != nil || label == "" {
				return false, ErrDimension
			}
			start := parser.layout.firstParts[variable]
			parser.labels[start], parser.labels[start+1] = label+".bar", label
		}
		return false, nil
	case "ob":
		if err = parser.requireLayout(); err != nil {
			return false, err
		}
		output := parser.layout.OutputVariable()
		if output < 0 || len(args) != parser.layout.partSizes[output] {
			return false, ErrDimension
		}
		copy(parser.labels[parser.layout.firstParts[output]:], args)
		return false, nil
	case "label":
		if err = parser.requireLayout(); err != nil || len(args) < 1 || !strings.HasPrefix(args[0], "var=") {
			return false, ErrDimension
		}
		variable, parseErr := parseEspressoInt(strings.TrimPrefix(args[0], "var="))
		if parseErr != nil || variable < parser.layout.binaryVariables || variable >= len(parser.layout.partSizes) || len(args)-1 != parser.layout.partSizes[variable] {
			return false, ErrDimension
		}
		copy(parser.labels[parser.layout.firstParts[variable]:], args[1:])
		return false, nil
	case "phase":
		if err = parser.requireLayout(); err != nil || len(args) != 1 || len(parser.result.Phase) != 0 {
			return false, ErrDimension
		}
		output := parser.layout.OutputVariable()
		if output < 0 || len(args[0]) != parser.layout.partSizes[output] {
			return false, ErrDimension
		}
		parser.result.Phase = make([]bool, len(args[0]))
		for index, value := range []byte(args[0]) {
			if value != '0' && value != '1' {
				return false, ErrDimension
			}
			parser.result.Phase[index] = value == '1'
		}
		return false, nil
	case "pair":
		if err = parser.requireLayout(); err != nil || len(args) < 1 || len(parser.result.Pairing.Pairs) != 0 {
			return false, ErrDimension
		}
		count, parseErr := parseEspressoInt(args[0])
		if parseErr != nil || len(args) != 1+2*count {
			return false, ErrDimension
		}
		for index := 0; index < count; index++ {
			firstWord := strings.Trim(args[1+2*index], "()")
			secondWord := strings.Trim(args[2+2*index], "()")
			first, firstOK := parser.labelIndex(firstWord)
			second, secondOK := parser.labelIndex(secondWord)
			if !firstOK || !secondOK || first.Variable >= parser.layout.binaryVariables || second.Variable >= parser.layout.binaryVariables {
				return false, ErrDimension
			}
			parser.result.Pairing.Pairs = append(parser.result.Pairing.Pairs, EspressoVariablePair{First: first.Variable, Second: second.Variable})
		}
		_, err = parser.layout.validatePairing(parser.result.Pairing)
		return false, err
	case "symbolic", "symbolic-output":
		if err = parser.requireLayout(); err != nil {
			return false, err
		}
		directive, parseErr := parser.parseSymbolic(args)
		if parseErr != nil {
			return false, parseErr
		}
		if command == "symbolic" {
			parser.result.SymbolicInputs = append(parser.result.SymbolicInputs, directive)
		} else {
			parser.result.SymbolicOutputs = append(parser.result.SymbolicOutputs, directive)
		}
		return false, nil
	default:
		parser.result.Unknown = append(parser.result.Unknown, line)
		return false, nil
	}
}

type espressoProductReader struct {
	parser *espressoPLAParser
	text   string
	pos    int
}

func (reader *espressoProductReader) nextCharacter() (byte, error) {
	for reader.pos < len(reader.text) {
		value := reader.text[reader.pos]
		reader.pos++
		if value == ' ' || value == '\t' || value == '|' || value == '\r' || value == '\n' {
			continue
		}
		if err := reader.parser.step(1); err != nil {
			return 0, err
		}
		return value, nil
	}
	return 0, errEspressoProductIncomplete
}

func (reader *espressoProductReader) nextWord() (string, error) {
	for reader.pos < len(reader.text) && (reader.text[reader.pos] == ' ' || reader.text[reader.pos] == '\t' || reader.text[reader.pos] == '|' || reader.text[reader.pos] == '\n') {
		reader.pos++
	}
	start := reader.pos
	for reader.pos < len(reader.text) && reader.text[reader.pos] != ' ' && reader.text[reader.pos] != '\t' && reader.text[reader.pos] != '|' && reader.text[reader.pos] != '\r' && reader.text[reader.pos] != '\n' {
		reader.pos++
	}
	if start == reader.pos {
		return "", errEspressoProductIncomplete
	}
	if err := reader.parser.token(); err != nil {
		return "", err
	}
	return reader.text[start:reader.pos], nil
}

func (parser *espressoPLAParser) addCube(family *EspressoSetFamily, cube *EspressoSet) error {
	parser.cubes++
	if parser.cubes > parser.options.MaxCubes {
		return ErrLimit
	}
	return family.Add(cube)
}

func (parser *espressoPLAParser) parseProduct(line string) error {
	if err := parser.requireLayout(); err != nil {
		return err
	}
	reader := espressoProductReader{parser: parser, text: line}
	prefix, _ := NewEspressoSet(parser.layout.bitCount)
	for variable := 0; variable < parser.layout.binaryVariables; variable++ {
		value, err := reader.nextCharacter()
		if err != nil {
			return err
		}
		start := parser.layout.firstParts[variable]
		switch value {
		case '2', '-':
			_ = prefix.Set(start+1, true)
			fallthrough
		case '0':
			_ = prefix.Set(start, true)
		case '1':
			_ = prefix.Set(start+1, true)
		case '?':
		default:
			return ErrDimension
		}
	}
	output := parser.layout.OutputVariable()
	for variable := parser.layout.binaryVariables; variable < output; variable++ {
		if parser.layout.symbolic[variable] {
			word, err := reader.nextWord()
			if err != nil {
				return err
			}
			if word == "-" || word == "ANY" {
				if !(parser.result.KISS && variable == output-1) {
					for part := 0; part < parser.layout.partSizes[variable]; part++ {
						_ = prefix.Set(parser.layout.firstParts[variable]+part, true)
					}
				}
			} else if word != "~" {
				labelVariable, offset := variable, 0
				if parser.result.KISS && variable == output-1 {
					labelVariable, offset = variable-1, parser.layout.partSizes[variable-1]
				}
				if labelVariable < 0 || parser.layout.partSizes[labelVariable] != parser.layout.partSizes[variable] {
					return ErrDimension
				}
				found := -1
				start := parser.layout.firstParts[labelVariable]
				for part := 0; part < parser.layout.partSizes[labelVariable]; part++ {
					if parser.labels[start+part] == word || parser.labels[start+part] == "" && found < 0 {
						if parser.labels[start+part] == word {
							found = part
							break
						}
						found = part
					}
				}
				if found < 0 {
					return ErrLimit
				}
				if parser.labels[start+found] == "" {
					parser.labels[start+found] = word
				}
				_ = prefix.Set(start+found+offset, true)
			}
		} else {
			for part := 0; part < parser.layout.partSizes[variable]; part++ {
				value, err := reader.nextCharacter()
				if err != nil {
					return err
				}
				if value == '1' {
					_ = prefix.Set(parser.layout.firstParts[variable]+part, true)
				} else if value != '0' {
					return ErrDimension
				}
			}
		}
	}
	onCube, _ := prefix.Clone()
	dcCube, _ := prefix.Clone()
	offCube, _ := prefix.Clone()
	saveOn, saveDC, saveOff := false, false, false
	if parser.result.KISS {
		if output < 1 {
			return ErrDimension
		}
		saveOn, saveOff = true, true
		state := output - 1
		for part := 0; part < parser.layout.partSizes[state]; part++ {
			index := parser.layout.firstParts[state] + part
			present, _ := offCube.Has(index)
			_ = offCube.Set(index, !present)
		}
	}
	if output < 0 {
		return ErrDimension
	}
	for part := 0; part < parser.layout.partSizes[output]; part++ {
		value, err := reader.nextCharacter()
		if err != nil {
			return err
		}
		index := parser.layout.firstParts[output] + part
		switch value {
		case '4', '1':
			if parser.typeBits&EspressoPLAOutputOn != 0 {
				_ = onCube.Set(index, true)
				saveOn = true
			}
		case '3', '0':
			if parser.typeBits&EspressoPLAOutputOff != 0 {
				_ = offCube.Set(index, true)
				saveOff = true
			}
		case '2', '-':
			if parser.typeBits&EspressoPLAOutputDontCare != 0 {
				_ = dcCube.Set(index, true)
				saveDC = true
			}
		case '~':
		default:
			return ErrDimension
		}
	}
	if strings.TrimSpace(reader.text[reader.pos:]) != "" {
		return ErrDimension
	}
	if saveOn {
		if err := parser.addCube(parser.on, onCube); err != nil {
			return err
		}
	}
	if saveDC {
		if err := parser.addCube(parser.dc, dcCube); err != nil {
			return err
		}
	}
	if saveOff {
		if err := parser.addCube(parser.off, offCube); err != nil {
			return err
		}
	}
	parser.result.Products++
	return nil
}

func (parser *espressoPLAParser) mergeKISSLayout() error {
	if !parser.result.KISS {
		return nil
	}
	if parser.layout.MultipleValuedVariableCount() < 3 {
		return ErrDimension
	}
	last := len(parser.layout.partSizes) - 1
	next := last - 1
	present := last - 2
	if parser.layout.partSizes[present] != parser.layout.partSizes[next] {
		return ErrDimension
	}
	for part := 0; part < parser.layout.partSizes[next]; part++ {
		parser.labels[parser.layout.firstParts[next]+part] = parser.labels[parser.layout.firstParts[present]+part]
	}
	parts := slices.Clone(parser.layout.partSizes[:last])
	parts[next] += parser.layout.partSizes[last]
	layout, err := NewEspressoCubeLayout(parts, parser.layout.binaryVariables)
	if err != nil || layout.bitCount != parser.layout.bitCount {
		return ErrInvariant
	}
	parser.layout = layout
	if len(parser.result.Phase) != 0 {
		phase := make([]bool, parser.layout.partSizes[next])
		for index := 0; index < parser.layout.partSizes[present]; index++ {
			phase[index] = true
		}
		copy(phase[parser.layout.partSizes[present]:], parser.result.Phase)
		parser.result.Phase = phase
	}
	return nil
}

func (parser *espressoPLAParser) complete() (EspressoPLAParseResult, error) {
	if parser.layout == nil {
		return EspressoPLAParseResult{}, ErrDimension
	}
	if err := parser.mergeKISSLayout(); err != nil {
		return EspressoPLAParseResult{}, err
	}
	labels := parser.labels
	allEmpty := true
	for _, label := range labels {
		allEmpty = allEmpty && label == ""
	}
	if allEmpty {
		labels = nil
	}
	pla, err := NewEspressoPLA(parser.layout, parser.on, parser.dc, parser.off, labels)
	if err != nil {
		return EspressoPLAParseResult{}, err
	}
	pla.Filename, pla.Type = parser.options.Filename, int(parser.typeBits)
	sharp := EspressoSharpOptions{MaxCubes: parser.options.MaxCubes, MaxOperations: parser.options.MaxOperations, Stop: parser.options.Stop}
	if parser.options.NeedsOff && (parser.typeBits == EspressoPLAOutputOn || parser.typeBits == EspressoPLAOutputOn|EspressoPLAOutputDontCare) {
		union, joinErr := JoinEspressoSetFamilies(pla.On, pla.DontCare)
		if joinErr != nil {
			return EspressoPLAParseResult{}, joinErr
		}
		pla.Off, err = pla.Layout.ComplementCover(union, sharp)
	} else if parser.options.NeedsDontCare && parser.typeBits == EspressoPLAOutputOn|EspressoPLAOutputOff {
		union, joinErr := JoinEspressoSetFamilies(pla.On, pla.Off)
		if joinErr != nil {
			return EspressoPLAParseResult{}, joinErr
		}
		pla.DontCare, err = pla.Layout.ComplementCover(union, sharp)
	} else if parser.typeBits == EspressoPLAOutputOff || parser.typeBits == EspressoPLAOutputDontCare|EspressoPLAOutputOff {
		union, joinErr := JoinEspressoSetFamilies(pla.DontCare, pla.Off)
		if joinErr != nil {
			return EspressoPLAParseResult{}, joinErr
		}
		pla.On, err = pla.Layout.ComplementCover(union, sharp)
	}
	if err != nil {
		return EspressoPLAParseResult{}, err
	}
	if parser.options.Positive {
		pla.On, pla.Off = pla.Off, pla.On
		parser.result.Phase = make([]bool, pla.Layout.partSizes[pla.Layout.OutputVariable()])
	} else if len(parser.result.Phase) != 0 {
		pla, err = pla.ApplyOutputPhase(parser.result.Phase, EspressoPhaseOptions{MaxCubes: parser.options.MaxCubes, MaxOperations: parser.options.MaxOperations, Stop: parser.options.Stop})
		if err != nil {
			return EspressoPLAParseResult{}, err
		}
	}
	if len(parser.result.Pairing.Pairs) != 0 {
		paired, pairErr := pla.PairVariables(parser.result.Pairing, true, EspressoPairOptions{MaxCubes: parser.options.MaxCubes, MaxOperations: parser.options.MaxOperations, MaxDepth: parser.options.MaxDepth, Stop: parser.options.Stop})
		if pairErr != nil {
			return EspressoPLAParseResult{}, pairErr
		}
		pla = paired.PLA
	}
	if !parser.options.DeferSymbolicMappings {
		symbolicOptions := EspressoSymbolicOptions{MaxCubes: parser.options.MaxCubes, MaxOperations: parser.options.MaxOperations, MaxDepth: parser.options.MaxDepth, Stop: parser.options.Stop}
		if len(parser.result.SymbolicInputs) != 0 {
			pla, err = pla.ApplySymbolicInputs(parser.result.SymbolicInputs, symbolicOptions)
			if err != nil {
				return EspressoPLAParseResult{}, err
			}
		}
		if len(parser.result.SymbolicOutputs) != 0 {
			pla, err = pla.ApplySymbolicOutputs(parser.result.SymbolicOutputs, symbolicOptions)
			if err != nil {
				return EspressoPLAParseResult{}, err
			}
			parser.result.Phase = nil
		}
	}
	parser.result.PLA = pla
	parser.result.Pairing = parser.result.Pairing.Clone()
	parser.result.Phase = slices.Clone(parser.result.Phase)
	return parser.result, nil
}

func ParseEspressoPLA(data []byte, options EspressoPLAInputOptions) (EspressoPLAParseResult, error) {
	options, err := normalizeEspressoPLAInputOptions(options)
	if err != nil {
		return EspressoPLAParseResult{}, err
	}
	if len(data) > options.MaxBytes {
		return EspressoPLAParseResult{}, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return EspressoPLAParseResult{}, ErrCancelled
	}
	parser := espressoPLAParser{options: options, typeBits: options.Type, pendingBinary: -1}
	lines := strings.Split(string(data), "\n")
	if len(lines) > options.MaxLines {
		return EspressoPLAParseResult{}, ErrLimit
	}
	for index := 0; index < len(lines); index++ {
		raw := lines[index]
		parser.result.Lines = index + 1
		if err = parser.step(len(raw) + 1); err != nil {
			return EspressoPLAParseResult{}, err
		}
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}
		if line[0] == '#' {
			parser.result.Comments = append(parser.result.Comments, line)
			continue
		}
		if line[0] == '.' {
			done, directiveErr := parser.parseDirective(line)
			if directiveErr != nil {
				return EspressoPLAParseResult{}, directiveErr
			}
			if done {
				break
			}
			continue
		}
		productLine := index + 1
		product := line
		spanned := false
		err = parser.parseProduct(product)
		for errors.Is(err, errEspressoProductIncomplete) && index+1 < len(lines) {
			nextRaw := lines[index+1]
			next := strings.TrimSpace(strings.TrimSuffix(nextRaw, "\r"))
			if next != "" && (next[0] == '.' || next[0] == '#') {
				break
			}
			index++
			parser.result.Lines = index + 1
			if stepErr := parser.step(len(nextRaw) + 1); stepErr != nil {
				return EspressoPLAParseResult{}, stepErr
			}
			product += "\n" + nextRaw
			spanned = true
			err = parser.parseProduct(product)
		}
		if err != nil {
			parser.result.Warnings = append(parser.result.Warnings, "ignored malformed product at line "+strconv.Itoa(productLine))
			if errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled) {
				return EspressoPLAParseResult{}, err
			}
		} else if spanned {
			parser.result.Warnings = append(parser.result.Warnings, "product term spans more than one line at line "+strconv.Itoa(productLine))
		}
	}
	return parser.complete()
}

func (result EspressoPLAParseResult) Summary(options EspressoCoverCostOptions) (string, error) {
	if result.PLA == nil {
		return "", ErrDimension
	}
	if err := result.PLA.check(); err != nil {
		return "", err
	}
	name := result.PLA.Filename
	var text strings.Builder
	text.WriteString("# PLA is " + name)
	layout := result.PLA.Layout
	if layout.binaryVariables == len(layout.partSizes)-1 {
		text.WriteString(" with " + strconv.Itoa(layout.binaryVariables) + " inputs and " + strconv.Itoa(layout.partSizes[len(layout.partSizes)-1]) + " outputs\n")
	} else {
		text.WriteString(" with " + strconv.Itoa(len(layout.partSizes)) + " variables (" + strconv.Itoa(layout.binaryVariables) + " binary, mv sizes")
		for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
			text.WriteString(" " + strconv.Itoa(layout.partSizes[variable]))
		}
		text.WriteString(")\n")
	}
	for _, item := range []struct {
		name, spacing string
		family        *EspressoSetFamily
	}{{"ON", "  ", result.PLA.On}, {"OFF", " ", result.PLA.Off}, {"DC", "  ", result.PLA.DontCare}} {
		cost, err := layout.CoverCost(item.family, options)
		if err != nil {
			return "", err
		}
		formatted, err := layout.FormatCoverCost(cost)
		if err != nil {
			return "", err
		}
		text.WriteString("# " + item.name + "-set cost is" + item.spacing + formatted + "\n")
	}
	if len(result.Phase) != 0 {
		text.WriteString("# phase is ")
		for _, positive := range result.Phase {
			text.WriteByte("01"[btoi(positive)])
		}
		text.WriteByte('\n')
	}
	if len(result.Pairing.Pairs) != 0 {
		text.WriteString("# two-bit decoders:")
		for _, pair := range result.Pairing.Pairs {
			text.WriteString(" (" + strconv.Itoa(pair.First+1) + " " + strconv.Itoa(pair.Second+1) + ")")
		}
		text.WriteByte('\n')
	}
	return text.String(), nil
}
