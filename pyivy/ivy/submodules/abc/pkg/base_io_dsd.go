package abc

// DSD formula reader from src/base/io/ioReadDsd.c.

import "strings"

// ABCDSDReadOptions bounds source-formula preprocessing, parsing and the
// returned SOP network. Numeric zeroes select defaults.
type ABCDSDReadOptions struct {
	MaxFormulaBytes int
	MaxObjects      int
	MaxEdges        int
	MaxFanins       int
	MaxDepth        int
	MaxCubes        int
	MaxWork         int
	MaxBytes        int
	Stop            func() bool
}

func normalizeABCDSDReadOptions(options ABCDSDReadOptions) (ABCDSDReadOptions, error) {
	values := []*int{&options.MaxFormulaBytes, &options.MaxObjects, &options.MaxEdges, &options.MaxFanins, &options.MaxDepth, &options.MaxCubes, &options.MaxWork, &options.MaxBytes}
	defaults := [...]int{1 << 20, 1 << 20, 1 << 22, 32, 256, 1 << 20, 1 << 24, 256 << 20}
	for index, value := range values {
		if *value < 0 {
			return options, ErrDimension
		}
		if *value == 0 {
			*value = defaults[index]
		}
	}
	return options, nil
}

type abcDSDReadBudget struct {
	options ABCDSDReadOptions
	work    int
	bytes   int
}

func (budget *abcDSDReadBudget) take(amount int) error {
	if amount < 0 || budget.work > budget.options.MaxWork-amount {
		return ErrLimit
	}
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	budget.work += amount
	return nil
}

func (budget *abcDSDReadBudget) retain(amount int) error {
	if amount < 0 || budget.bytes > budget.options.MaxBytes-amount {
		return ErrLimit
	}
	budget.bytes += amount
	return nil
}

func abcDSDFindEnd(text string, start int, budget *abcDSDReadBudget) (int, error) {
	if start < 0 || start >= len(text) || text[start] != '(' {
		return -1, ErrFormat
	}
	parts := 0
	for index := start; index < len(text); index++ {
		if err := budget.take(1); err != nil {
			return -1, err
		}
		switch text[index] {
		case '(':
			parts++
		case ')':
			parts--
			if parts == 0 {
				return index, nil
			}
		}
	}
	return -1, ErrFormat
}

// ABCDSDFindEnd implements Io_ReadDsdFindEnd over immutable text.
func ABCDSDFindEnd(text string, start int, options ABCDSDReadOptions) (int, error) {
	options, err := normalizeABCDSDReadOptions(options)
	if err != nil {
		return -1, err
	}
	if len(text) > options.MaxFormulaBytes {
		return -1, ErrLimit
	}
	budget := abcDSDReadBudget{options: options}
	return abcDSDFindEnd(text, start, &budget)
}

type abcDSDSplit struct {
	parts []string
	xor   bool
}

func isABCDSDHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'A' && value <= 'F'
}

func abcDSDSplitParts(text string, budget *abcDSDReadBudget) (abcDSDSplit, error) {
	if text == "" {
		return abcDSDSplit{}, ErrFormat
	}
	result := abcDSDSplit{parts: make([]string, 0, 4)}
	separator := byte(0)
	for cursor := 0; cursor < len(text); {
		if len(result.parts) >= budget.options.MaxFanins {
			return abcDSDSplit{}, ErrLimit
		}
		start := cursor
		if text[cursor] == '!' {
			cursor++
			if cursor == len(text) {
				return abcDSDSplit{}, ErrFormat
			}
		}
		if text[cursor] >= 'a' && text[cursor] <= 'z' {
			cursor++
		} else {
			for cursor < len(text) && isABCDSDHex(text[cursor]) {
				cursor++
			}
			if cursor >= len(text) || text[cursor] != '(' {
				return abcDSDSplit{}, ErrFormat
			}
			end, err := abcDSDFindEnd(text, cursor, budget)
			if err != nil {
				return abcDSDSplit{}, err
			}
			cursor = end + 1
		}
		result.parts = append(result.parts, text[start:cursor])
		if cursor == len(text) {
			break
		}
		current := text[cursor]
		if current != '*' && current != '+' && current != ',' {
			return abcDSDSplit{}, ErrFormat
		}
		if separator == 0 {
			separator = current
		} else if separator != current {
			return abcDSDSplit{}, ErrFormat
		}
		cursor++
		if cursor == len(text) {
			return abcDSDSplit{}, ErrFormat
		}
	}
	result.xor = separator == '+'
	return result, nil
}

// SplitABCDSDParts implements Io_ReadDsdStrSplit without mutating the input.
func SplitABCDSDParts(text string, options ABCDSDReadOptions) ([]string, bool, error) {
	options, err := normalizeABCDSDReadOptions(options)
	if err != nil {
		return nil, false, err
	}
	if len(text) > options.MaxFormulaBytes {
		return nil, false, ErrLimit
	}
	budget := abcDSDReadBudget{options: options}
	result, err := abcDSDSplitParts(text, &budget)
	if err != nil {
		return nil, false, err
	}
	return append([]string(nil), result.parts...), result.xor, nil
}

func abcDSDTransform(formula string, budget *abcDSDReadBudget) (string, error) {
	capacity, err := checkedSize(2, len(formula), 3, budget.options.MaxBytes)
	if err != nil {
		return "", ErrLimit
	}
	var builder strings.Builder
	builder.Grow(capacity)
	builder.WriteByte('(')
	for index := range formula {
		if err = budget.take(1); err != nil {
			return "", err
		}
		switch formula[index] {
		case '(':
			builder.WriteString("((")
		case ')':
			builder.WriteString("))")
		case ',':
			builder.WriteString("),(")
		default:
			builder.WriteByte(formula[index])
		}
	}
	builder.WriteByte(')')
	result := builder.String()
	if err = budget.retain(capacity); err != nil {
		return "", err
	}
	return result, nil
}

func abcDSDHexSOP(digits string, fanins int, budget *abcDSDReadBudget) (ABCSOP, error) {
	if len(digits) == 0 || len(digits)&(len(digits)-1) != 0 {
		return ABCSOP{}, ErrFormat
	}
	variables := 2
	for size := len(digits); size > 1; size >>= 1 {
		variables++
	}
	if variables != fanins || variables > budget.options.MaxFanins {
		return ABCSOP{}, ErrDimension
	}
	cubes := make([]string, 0)
	for index := range digits {
		value := digits[index]
		digit := int(value - '0')
		if value >= 'A' {
			digit = int(value-'A') + 10
		}
		for bit := 0; bit < 4; bit++ {
			if digit&(1<<bit) == 0 {
				continue
			}
			if err := budget.take(1); err != nil {
				return ABCSOP{}, err
			}
			if len(cubes) >= budget.options.MaxCubes {
				return ABCSOP{}, ErrLimit
			}
			minterm := 4*(len(digits)-1-index) + bit
			cube := make([]byte, variables)
			for variable := range cube {
				cube[variable] = byte('0' + ((minterm >> variable) & 1))
			}
			if err := budget.retain(16 + len(cube)); err != nil {
				return ABCSOP{}, err
			}
			cubes = append(cubes, string(cube))
		}
	}
	if len(cubes) == 0 {
		cubes = []string{strings.Repeat("-", variables)}
		return ABCSOP{Variables: variables, Cubes: cubes, Phase: '0'}, nil
	}
	return ABCSOP{Variables: variables, Cubes: cubes, Phase: '1'}, nil
}

type abcDSDReader struct {
	network *ABCNetwork
	budget  *abcDSDReadBudget
}

func (reader *abcDSDReader) parse(text string, depth int, primeDigits string) (int, error) {
	if depth > reader.budget.options.MaxDepth {
		return -1, ErrLimit
	}
	if err := reader.budget.take(1); err != nil {
		return -1, err
	}
	if text == "" {
		return -1, ErrFormat
	}
	if text[0] == '!' {
		fanin, err := reader.parse(text[1:], depth+1, "")
		if err != nil {
			return -1, err
		}
		return reader.createNode(ABCSOP{Variables: 1, Cubes: []string{"0"}, Phase: '1'}, []int{fanin})
	}
	if text[0] == '(' {
		end, err := abcDSDFindEnd(text, 0, reader.budget)
		if err != nil || end != len(text)-1 {
			return -1, ErrFormat
		}
		split, err := abcDSDSplitParts(text[1:len(text)-1], reader.budget)
		if err != nil {
			return -1, err
		}
		parent, err := reader.network.CreateNode()
		if err != nil {
			return -1, err
		}
		if err = reader.budget.retain(64); err != nil {
			return -1, err
		}
		if len(split.parts) > (reader.budget.options.MaxBytes-reader.budget.bytes)/24 {
			return -1, ErrLimit
		}
		if err = reader.budget.retain(24 * len(split.parts)); err != nil {
			return -1, err
		}
		fanins := make([]int, len(split.parts))
		for index, part := range split.parts {
			fanins[index], err = reader.parse(part, depth+1, "")
			if err != nil {
				return -1, err
			}
			if err = reader.network.AddFanin(parent, ABCObjectEdge{ID: fanins[index]}); err != nil {
				return -1, err
			}
		}
		var sop ABCSOP
		if primeDigits != "" {
			sop, err = abcDSDHexSOP(primeDigits, len(fanins), reader.budget)
		} else if split.xor {
			sop = ABCSOP{Variables: len(fanins), Cubes: []string{strings.Repeat("1", len(fanins))}, Phase: 'x'}
		} else {
			sop, err = ABCSOPAnd(len(fanins), nil, false)
		}
		if err != nil {
			return -1, err
		}
		function, err := sop.Format(ABCMetadataOptions{MaxVariables: reader.budget.options.MaxFanins, MaxCubes: reader.budget.options.MaxCubes, MaxBytes: reader.budget.options.MaxBytes, MaxOperations: reader.budget.options.MaxWork, Stop: reader.budget.options.Stop})
		if err != nil {
			return -1, err
		}
		if err = reader.budget.retain(len(function)); err != nil {
			return -1, err
		}
		if err = reader.network.SetNodeFunction(parent, function); err != nil {
			return -1, err
		}
		return parent, nil
	}
	if text[0] >= 'a' && text[0] <= 'z' {
		if len(text) != 1 {
			return -1, ErrFormat
		}
		index := int(text[0] - 'a')
		if index >= len(reader.network.pis) {
			return -1, ErrInvariant
		}
		return reader.network.pis[index], nil
	}
	cursor := 0
	for cursor < len(text) && isABCDSDHex(text[cursor]) {
		cursor++
	}
	if cursor == 0 || cursor >= len(text) || text[cursor] != '(' {
		return -1, ErrFormat
	}
	end, err := abcDSDFindEnd(text, cursor, reader.budget)
	if err != nil || end != len(text)-1 {
		return -1, ErrFormat
	}
	return reader.parse(text[cursor:], depth+1, text[:cursor])
}

func (reader *abcDSDReader) createNode(sop ABCSOP, fanins []int) (int, error) {
	function, err := sop.Format(ABCMetadataOptions{MaxVariables: reader.budget.options.MaxFanins, MaxCubes: reader.budget.options.MaxCubes, MaxBytes: reader.budget.options.MaxBytes, MaxOperations: reader.budget.options.MaxWork, Stop: reader.budget.options.Stop})
	if err != nil {
		return -1, err
	}
	if len(fanins) > (reader.budget.options.MaxBytes-reader.budget.bytes)/16 {
		return -1, ErrLimit
	}
	if err = reader.budget.retain(64 + len(function) + 16*len(fanins)); err != nil {
		return -1, err
	}
	edges := make([]ABCObjectEdge, len(fanins))
	for index, id := range fanins {
		edges[index] = ABCObjectEdge{ID: id}
	}
	return reader.network.CreateLogicNode(function, edges...)
}

func abcDSDFakeName(index, count int) string {
	if count < 26 {
		return string([]byte{byte('a' + index)})
	}
	return string([]byte{byte('a' + index%26), byte('0' + index/26)})
}

// ReadABCDSD derives the source-shaped owned SOP network for one DSD formula.
func ReadABCDSD(formula string, options ABCDSDReadOptions) (*ABCNetwork, error) {
	options, err := normalizeABCDSDReadOptions(options)
	if err != nil {
		return nil, err
	}
	if len(formula) > options.MaxFormulaBytes || len(formula) > options.MaxBytes {
		return nil, ErrLimit
	}
	budget := &abcDSDReadBudget{options: options}
	variables := 1
	for index := range formula {
		if formula[index] >= 'a' && formula[index] <= 'z' && int(formula[index]-'a')+1 > variables {
			variables = int(formula[index]-'a') + 1
		}
	}
	if variables > 26 || variables+2 > options.MaxObjects {
		return nil, ErrLimit
	}
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: options.MaxObjects, MaxEdges: options.MaxEdges, MaxNameBytes: options.MaxBytes, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	network.Name = "dsd"
	for index := 0; index < variables; index++ {
		id, createErr := network.CreatePI()
		if createErr != nil {
			return nil, createErr
		}
		name := abcDSDFakeName(index, variables)
		if err = network.SetObjectName(id, name); err != nil {
			return nil, err
		}
		if err = budget.retain(64 + len(name)); err != nil {
			return nil, err
		}
	}
	transformed, err := abcDSDTransform(formula, budget)
	if err != nil {
		return nil, err
	}
	reader := abcDSDReader{network: network, budget: budget}
	root, err := reader.parse(transformed, 1, "")
	if err != nil {
		return nil, err
	}
	output, err := network.CreatePO()
	if err != nil {
		return nil, err
	}
	if err = network.SetObjectName(output, "F"); err != nil {
		return nil, err
	}
	if err = network.AddFanin(output, ABCObjectEdge{ID: root}); err != nil {
		return nil, err
	}
	if err = budget.retain(65); err != nil {
		return nil, err
	}
	if err = network.Validate(); err != nil {
		return nil, err
	}
	return network, nil
}
