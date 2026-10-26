package abc

// Owned functional port of src/bdd/cas/cas.h, casCore.c, and casDec.c.

import (
	"slices"
	"strconv"
	"strings"
)

const (
	CascadeMaxInputs  = 1024
	CascadeMaxOutputs = 1024
)

// CascadeOptions bounds the truth-based counterpart of the source BDD
// cascade. Zero limits select conservative defaults.
type CascadeOptions struct {
	MaxVariables int
	MaxStages    int
	MaxStates    int
	MaxEntries   int
	MaxBytes     int
	Stop         func() bool
}

func normalizeCascadeOptions(options CascadeOptions) (CascadeOptions, error) {
	if options.MaxVariables < 0 || options.MaxStages < 0 || options.MaxStates < 0 || options.MaxEntries < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 20
	}
	if options.MaxStages == 0 {
		options.MaxStages = CascadeMaxInputs
	}
	if options.MaxStates == 0 {
		options.MaxStates = 1 << 16
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 22
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 27
	}
	return options, nil
}

// CascadeGrayCode is the active GrayCode helper in casCore.c.
func CascadeGrayCode(code uint32) uint32 { return code ^ code>>1 }

func cascadeLog2(count int) int {
	bits := 0
	for value := 1; value < count; value <<= 1 {
		bits++
	}
	return bits
}

// CascadeSupportOrder preserves GetSingleOutputFunction's descending support
// order. Equal supports retain output order, making the source qsort tie
// deterministic.
func CascadeSupportOrder(outputs []*TruthTable) ([]int, error) {
	if len(outputs) == 0 || len(outputs) > CascadeMaxOutputs {
		return nil, ErrDimension
	}
	variables := -1
	sizes := make([]int, len(outputs))
	for index, output := range outputs {
		if output == nil || output.Check() != nil {
			return nil, ErrDimension
		}
		if variables < 0 {
			variables = output.Variables()
		} else if output.Variables() != variables {
			return nil, ErrDimension
		}
		support, err := output.Support()
		if err != nil {
			return nil, err
		}
		sizes[index] = len(support)
	}
	order := make([]int, len(outputs))
	for index := range order {
		order[index] = index
	}
	slices.SortStableFunc(order, func(first, second int) int {
		return sizes[second] - sizes[first]
	})
	return order, nil
}

// CascadeMintermOrder is CompareMinterms' deterministic counterpart. The C
// experiment keeps this alternate ordering compiled but does not select it.
func CascadeMintermOrder(oneCounts []int) ([]int, error) {
	if len(oneCounts) == 0 || len(oneCounts) > CascadeMaxOutputs {
		return nil, ErrDimension
	}
	order := make([]int, len(oneCounts))
	for index, count := range oneCounts {
		if count < 0 {
			return nil, ErrDimension
		}
		order[index] = index
	}
	slices.SortStableFunc(order, func(first, second int) int {
		return oneCounts[second] - oneCounts[first]
	})
	return order, nil
}

// CascadeEncodedFunction is the owned 0/1 relation produced by the source's
// multiple-output encoding. Original input variables occupy the low indexes;
// encoding bits follow them. Code i selects OutputOrder[i], with the first
// encoding variable as the most-significant bit, matching Extra_bddBitsToCube.
// Unused codes are 0.
type CascadeEncodedFunction struct {
	Relation     *TruthTable
	InputCount   int
	EncodingBits int
	OutputOrder  []int
}

func (encoded CascadeEncodedFunction) Clone() CascadeEncodedFunction {
	encoded.Relation = encoded.Relation.Clone()
	encoded.OutputOrder = slices.Clone(encoded.OutputOrder)
	return encoded
}

func cascadeEncodingAssignment(code, width int) uint64 {
	var assignment uint64
	for variable := 0; variable < width; variable++ {
		if code&(1<<uint(width-1-variable)) != 0 {
			assignment |= 1 << uint(variable)
		}
	}
	return assignment
}

// EncodeCascadeOutputs implements GetSingleOutputFunction without exposing a
// BDD manager or process-global comparator arrays.
func EncodeCascadeOutputs(outputs []*TruthTable, options CascadeOptions) (CascadeEncodedFunction, error) {
	options, err := normalizeCascadeOptions(options)
	if err != nil {
		return CascadeEncodedFunction{}, err
	}
	if options.Stop != nil && options.Stop() {
		return CascadeEncodedFunction{}, ErrCancelled
	}
	order, err := CascadeSupportOrder(outputs)
	if err != nil {
		return CascadeEncodedFunction{}, err
	}
	return encodeCascadeOutputsInOrder(outputs, order, options)
}

func encodeCascadeOutputsInOrder(outputs []*TruthTable, order []int, options CascadeOptions) (CascadeEncodedFunction, error) {
	inputs := outputs[0].Variables()
	encoding := cascadeLog2(len(outputs))
	variables := inputs + encoding
	if inputs > CascadeMaxInputs || variables > CascadeMaxInputs || variables > options.MaxVariables || variables >= 63 {
		return CascadeEncodedFunction{}, ErrLimit
	}
	entries := 1 << uint(variables)
	words, err := TruthTableWordCount(variables)
	if err != nil {
		return CascadeEncodedFunction{}, err
	}
	if entries > options.MaxEntries || words > options.MaxBytes/8 {
		return CascadeEncodedFunction{}, ErrLimit
	}
	relation, err := NewConstantTruthTable(variables, false, words)
	if err != nil {
		return CascadeEncodedFunction{}, err
	}
	inputEntries := uint64(1) << uint(inputs)
	for code, outputIndex := range order {
		selector := cascadeEncodingAssignment(code, encoding)
		for assignment := uint64(0); assignment < inputEntries; assignment++ {
			if assignment&1023 == 0 && options.Stop != nil && options.Stop() {
				return CascadeEncodedFunction{}, ErrCancelled
			}
			value, getErr := outputs[outputIndex].Get(assignment)
			if getErr != nil {
				return CascadeEncodedFunction{}, getErr
			}
			if value {
				if setErr := relation.Set(assignment|selector<<uint(inputs), true); setErr != nil {
					return CascadeEncodedFunction{}, setErr
				}
			}
		}
	}
	return CascadeEncodedFunction{relation, inputs, encoding, order}, nil
}

// CascadeRemappedFunction is GetSingleOutputFunctionRemapped's functional
// result. Each output is compacted independently onto low variables before
// encoding; SupportMaps records the original variable represented by each low
// position.
type CascadeRemappedFunction struct {
	Encoded     CascadeEncodedFunction
	SupportMaps [][]int
}

func EncodeRemappedCascadeOutputs(outputs []*TruthTable, options CascadeOptions) (CascadeRemappedFunction, error) {
	if len(outputs) == 0 || len(outputs) > CascadeMaxOutputs {
		return CascadeRemappedFunction{}, ErrDimension
	}
	options, err := normalizeCascadeOptions(options)
	if err != nil {
		return CascadeRemappedFunction{}, err
	}
	if options.Stop != nil && options.Stop() {
		return CascadeRemappedFunction{}, ErrCancelled
	}
	compacted := make([]*TruthTable, len(outputs))
	maps := make([][]int, len(outputs))
	maximum := 0
	variables := -1
	for index, output := range outputs {
		if output == nil || output.Check() != nil {
			return CascadeRemappedFunction{}, ErrDimension
		}
		if variables < 0 {
			variables = output.Variables()
		} else if output.Variables() != variables {
			return CascadeRemappedFunction{}, ErrDimension
		}
		reduction, err := output.CompactSupport()
		if err != nil {
			return CascadeRemappedFunction{}, err
		}
		compacted[index] = reduction.Table
		maps[index] = slices.Clone(reduction.SourceVariables)
		maximum = max(maximum, reduction.Table.Variables())
	}
	for index := range compacted {
		stretched, err := compacted[index].Stretch(maximum)
		if err != nil {
			return CascadeRemappedFunction{}, err
		}
		compacted[index] = stretched
	}
	order := make([]int, len(compacted))
	for index := range order {
		order[index] = index
	}
	encoded, err := encodeCascadeOutputsInOrder(compacted, order, options)
	if err != nil {
		return CascadeRemappedFunction{}, err
	}
	return CascadeRemappedFunction{encoded, maps}, nil
}

// TransferCascadeTruth is Cudd_bddTransferPermute at the owned truth boundary.
// permutation[source] gives the destination variable index.
func TransferCascadeTruth(source *TruthTable, permutation []int) (*TruthTable, error) {
	if source == nil || len(permutation) != source.Variables() {
		return nil, ErrDimension
	}
	return source.PermuteVariables(permutation)
}

// CascadeStage is one source LUT group. PreviousBits occupy the low table
// variables; InputVariables follow them. Nonfinal stages emit binary state;
// the final stage emits the encoded Boolean relation.
type CascadeStage struct {
	InputVariables []int
	PreviousBits   int
	OutputBits     int
	StateCount     int
	Final          bool
	Tables         []*TruthTable
}

func cloneCascadeStage(stage CascadeStage) CascadeStage {
	stage.InputVariables = slices.Clone(stage.InputVariables)
	stage.Tables = make([]*TruthTable, len(stage.Tables))
	for index, table := range stage.Tables {
		stage.Tables[index] = table.Clone()
	}
	return stage
}

// CascadeNetwork owns a deterministic residual-function encoding of the BDD
// width cascade in casDec.c.
type CascadeNetwork struct {
	inputCount, outputCount, encodingBits, lutSize int
	outputOrder                                    []int
	variableOrder                                  []int
	stages                                         []CascadeStage
	relation                                       *TruthTable
	closed                                         bool
}

func (network *CascadeNetwork) check() error {
	if network == nil || network.closed || network.relation == nil || len(network.stages) == 0 {
		return ErrInvariant
	}
	return network.relation.Check()
}

func (network *CascadeNetwork) Stages() ([]CascadeStage, error) {
	if err := network.check(); err != nil {
		return nil, err
	}
	result := make([]CascadeStage, len(network.stages))
	for index, stage := range network.stages {
		result[index] = cloneCascadeStage(stage)
	}
	return result, nil
}

func (network *CascadeNetwork) OutputOrder() ([]int, error) {
	if err := network.check(); err != nil {
		return nil, err
	}
	return slices.Clone(network.outputOrder), nil
}

func (network *CascadeNetwork) Relation() (*TruthTable, error) {
	if err := network.check(); err != nil {
		return nil, err
	}
	return network.relation.Clone(), nil
}

func (network *CascadeNetwork) Close() {
	if network == nil || network.closed {
		return
	}
	network.closed = true
	network.outputOrder = nil
	network.variableOrder = nil
	network.stages = nil
	network.relation = nil
}

func residualKey(values []bool) string {
	data := make([]byte, (len(values)+7)/8)
	for index, value := range values {
		if value {
			data[index/8] |= 1 << uint(index%8)
		}
	}
	return string(data)
}

func relationInOrder(relation *TruthTable, order []int, options CascadeOptions) ([]bool, error) {
	entries := 1 << uint(len(order))
	values := make([]bool, entries)
	for ordered := 0; ordered < entries; ordered++ {
		if ordered&1023 == 0 && options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		var source uint64
		for bit, variable := range order {
			if ordered&(1<<uint(bit)) != 0 {
				source |= 1 << uint(variable)
			}
		}
		value, err := relation.Get(source)
		if err != nil {
			return nil, err
		}
		values[ordered] = value
	}
	return values, nil
}

// BuildCascade decomposes all outputs into a LUT cascade. Encoding bits are
// processed first, matching the source's NewVarAtLevel insertion; original
// inputs follow in index order.
func BuildCascade(outputs []*TruthTable, lutSize int, options CascadeOptions) (*CascadeNetwork, error) {
	options, err := normalizeCascadeOptions(options)
	if err != nil {
		return nil, err
	}
	if lutSize < 2 || lutSize > 30 {
		return nil, ErrDimension
	}
	encoded, err := EncodeCascadeOutputs(outputs, options)
	if err != nil {
		return nil, err
	}
	variables := encoded.Relation.Variables()
	words, err := TruthTableWordCount(variables)
	if err != nil {
		return nil, err
	}
	entries := 1 << uint(variables)
	if entries > options.MaxBytes-8*words {
		return nil, ErrLimit
	}
	order := make([]int, 0, variables)
	for bit := 0; bit < encoded.EncodingBits; bit++ {
		order = append(order, encoded.InputCount+bit)
	}
	for input := 0; input < encoded.InputCount; input++ {
		order = append(order, input)
	}
	initial, err := relationInOrder(encoded.Relation, order, options)
	if err != nil {
		return nil, err
	}
	states := [][]bool{initial}
	consumed := 0
	previousBits := 0
	stages := make([]CascadeStage, 0)
	totalStored := len(initial)
	for {
		remaining := variables - consumed
		if len(stages) >= options.MaxStages {
			return nil, ErrLimit
		}
		if previousBits+remaining <= lutSize {
			stage, stageErr := buildFinalCascadeStage(states, order[consumed:], previousBits, options)
			if stageErr != nil {
				return nil, stageErr
			}
			stages = append(stages, stage)
			break
		}
		block := lutSize - previousBits
		if block <= 0 || block > remaining {
			return nil, ErrUnsupported
		}
		stage, next, stageEntries, stageErr := buildCascadeStage(states, order[consumed:consumed+block], previousBits, options)
		if stageErr != nil {
			return nil, stageErr
		}
		totalStored += stageEntries
		if totalStored > options.MaxEntries {
			return nil, ErrLimit
		}
		stages = append(stages, stage)
		states = next
		consumed += block
		previousBits = stage.OutputBits
	}
	return &CascadeNetwork{
		inputCount: encoded.InputCount, outputCount: len(outputs), encodingBits: encoded.EncodingBits,
		lutSize: lutSize, outputOrder: slices.Clone(encoded.OutputOrder), variableOrder: order,
		stages: stages, relation: encoded.Relation,
	}, nil
}

func buildCascadeStage(states [][]bool, inputs []int, previousBits int, options CascadeOptions) (CascadeStage, [][]bool, int, error) {
	blockEntries := 1 << uint(len(inputs))
	next := make([][]bool, 0)
	indexes := make(map[string]int)
	transitions := make([]int, (1<<uint(previousBits))*blockEntries)
	for state, residual := range states {
		if len(residual)%blockEntries != 0 {
			return CascadeStage{}, nil, 0, ErrInvariant
		}
		nextLength := len(residual) / blockEntries
		for assignment := 0; assignment < blockEntries; assignment++ {
			if options.Stop != nil && options.Stop() {
				return CascadeStage{}, nil, 0, ErrCancelled
			}
			child := make([]bool, nextLength)
			for suffix := 0; suffix < nextLength; suffix++ {
				child[suffix] = residual[assignment|(suffix<<uint(len(inputs)))]
			}
			key := residualKey(child)
			id, found := indexes[key]
			if !found {
				if len(next) >= options.MaxStates {
					return CascadeStage{}, nil, 0, ErrLimit
				}
				id = len(next)
				indexes[key] = id
				next = append(next, child)
			}
			transitions[state|(assignment<<uint(previousBits))] = id
		}
	}
	outputBits := cascadeLog2(len(next))
	tables := make([]*TruthTable, outputBits)
	for bit := 0; bit < outputBits; bit++ {
		table, err := NewConstantTruthTable(previousBits+len(inputs), false, options.MaxEntries)
		if err != nil {
			return CascadeStage{}, nil, 0, err
		}
		for assignment, state := range transitions {
			if state&(1<<uint(bit)) != 0 {
				if err = table.Set(uint64(assignment), true); err != nil {
					return CascadeStage{}, nil, 0, err
				}
			}
		}
		tables[bit] = table
	}
	return CascadeStage{slices.Clone(inputs), previousBits, outputBits, len(next), false, tables}, next, len(transitions), nil
}

func buildFinalCascadeStage(states [][]bool, inputs []int, previousBits int, options CascadeOptions) (CascadeStage, error) {
	table, err := NewConstantTruthTable(previousBits+len(inputs), false, options.MaxEntries)
	if err != nil {
		return CascadeStage{}, err
	}
	inputEntries := 1 << uint(len(inputs))
	for state, residual := range states {
		if len(residual) != inputEntries {
			return CascadeStage{}, ErrInvariant
		}
		for assignment, value := range residual {
			if options.Stop != nil && options.Stop() {
				return CascadeStage{}, ErrCancelled
			}
			if value {
				index := state | assignment<<uint(previousBits)
				if err = table.Set(uint64(index), true); err != nil {
					return CascadeStage{}, err
				}
			}
		}
	}
	return CascadeStage{slices.Clone(inputs), previousBits, 1, 2, true, []*TruthTable{table}}, nil
}

func cascadeInputBits(assignment uint64, variables []int) uint64 {
	var result uint64
	for bit, variable := range variables {
		if assignment&(1<<uint(variable)) != 0 {
			result |= 1 << uint(bit)
		}
	}
	return result
}

func (network *CascadeNetwork) evaluateRelation(assignment uint64) (bool, error) {
	if err := network.check(); err != nil {
		return false, err
	}
	if assignment >= uint64(1)<<uint(network.relation.Variables()) {
		return false, ErrDimension
	}
	state := uint64(0)
	for _, stage := range network.stages {
		inputs := cascadeInputBits(assignment, stage.InputVariables)
		index := state | inputs<<uint(stage.PreviousBits)
		if stage.Final {
			return stage.Tables[0].Get(index)
		}
		state = 0
		for bit, table := range stage.Tables {
			value, err := table.Get(index)
			if err != nil {
				return false, err
			}
			if value {
				state |= 1 << uint(bit)
			}
		}
		if state >= uint64(stage.StateCount) {
			return false, ErrInvariant
		}
	}
	return false, ErrInvariant
}

// Evaluate returns outputs in caller order, not support-sorted order.
func (network *CascadeNetwork) Evaluate(inputs uint64) ([]bool, error) {
	if err := network.check(); err != nil {
		return nil, err
	}
	if inputs >= uint64(1)<<uint(network.inputCount) {
		return nil, ErrDimension
	}
	result := make([]bool, network.outputCount)
	for code, output := range network.outputOrder {
		selector := cascadeEncodingAssignment(code, network.encodingBits)
		assignment := inputs | selector<<uint(network.inputCount)
		value, err := network.evaluateRelation(assignment)
		if err != nil {
			return nil, err
		}
		result[output] = value
	}
	return result, nil
}

type cascadeBDDNode struct{ variable, low, high int }

func buildCascadeBDD(table *TruthTable, maxEntries int) ([]cascadeBDDNode, int, error) {
	variables := table.Variables()
	entries := 1 << uint(variables)
	if entries > maxEntries {
		return nil, 0, ErrLimit
	}
	values := make([]bool, entries)
	for assignment := 0; assignment < entries; assignment++ {
		value, err := table.Get(uint64(assignment))
		if err != nil {
			return nil, 0, err
		}
		values[assignment] = value
	}
	nodes := []cascadeBDDNode{{variable: -1}, {variable: -1}}
	unique := make(map[[3]int]int)
	var build func(int, []bool) int
	build = func(level int, data []bool) int {
		allFalse, allTrue := true, true
		for _, value := range data {
			allFalse = allFalse && !value
			allTrue = allTrue && value
		}
		if allFalse {
			return 0
		}
		if allTrue {
			return 1
		}
		lowData := make([]bool, len(data)/2)
		highData := make([]bool, len(data)/2)
		for suffix := range lowData {
			lowData[suffix] = data[suffix<<1]
			highData[suffix] = data[suffix<<1|1]
		}
		low, high := build(level+1, lowData), build(level+1, highData)
		if low == high {
			return low
		}
		key := [3]int{level, low, high}
		if id, found := unique[key]; found {
			return id
		}
		id := len(nodes)
		unique[key] = id
		nodes = append(nodes, cascadeBDDNode{level, low, high})
		return id
	}
	return nodes, build(0, values), nil
}

func appendBounded(builder *strings.Builder, text string, maxBytes int) error {
	if builder.Len() > maxBytes-len(text) {
		return ErrLimit
	}
	builder.WriteString(text)
	return nil
}

func cascadeNameValid(name string) bool {
	return name != "" && !strings.ContainsAny(name, " \t\r\n")
}

func appendCascadeTruthBLIF(builder *strings.Builder, table *TruthTable, inputNames []string, output, prefix string, maxBytes int) error {
	if table == nil || table.Check() != nil || len(inputNames) != table.Variables() || !cascadeNameValid(output) || !cascadeNameValid(prefix) {
		return ErrDimension
	}
	for _, name := range inputNames {
		if !cascadeNameValid(name) {
			return ErrFormat
		}
	}
	remaining := maxBytes - builder.Len()
	if remaining < 0 {
		return ErrLimit
	}
	nodes, root, err := buildCascadeBDD(table, remaining)
	if err != nil {
		return err
	}
	write := func(text string) error { return appendBounded(builder, text, maxBytes) }
	if err = write(".names " + prefix + "0\n.names " + prefix + "1\n1\n"); err != nil {
		return err
	}
	for id := 2; id < len(nodes); id++ {
		node := nodes[id]
		line := ".names " + inputNames[node.variable] + " " + prefix + strconv.Itoa(node.low) + " " + prefix + strconv.Itoa(node.high) + " " + prefix + strconv.Itoa(id) + "\n01- 1\n1-1 1\n"
		if err = write(line); err != nil {
			return err
		}
	}
	return write(".names " + prefix + strconv.Itoa(root) + " " + output + "\n1 1\n")
}

// MarshalCascadeTruthBLIF replaces the source's pointer-address node names
// with deterministic reduced-BDD IDs while retaining its two-row mux encoding.
func MarshalCascadeTruthBLIF(table *TruthTable, inputNames []string, model, output string, maxBytes int) ([]byte, error) {
	if table == nil || table.Check() != nil || len(inputNames) != table.Variables() || model == "" || output == "" || maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 27
	}
	names := append(slices.Clone(inputNames), model, output)
	for _, name := range names {
		if !cascadeNameValid(name) {
			return nil, ErrFormat
		}
	}
	var builder strings.Builder
	write := func(text string) error { return appendBounded(&builder, text, maxBytes) }
	if err := write(".model " + model + "\n.inputs " + strings.Join(inputNames, " ") + "\n.outputs " + output + "\n"); err != nil {
		return nil, err
	}
	if err := appendCascadeTruthBLIF(&builder, table, inputNames, output, "cas_", maxBytes); err != nil {
		return nil, err
	}
	if err := write(".end\n"); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// MarshalCascadeNetworkBLIF is the owned counterpart of
// WriteLUTSintoBLIFfile. Input names are indexed like the encoded relation:
// original inputs first, followed by output-selector bits.
func MarshalCascadeNetworkBLIF(network *CascadeNetwork, inputNames []string, model string, maxBytes int) ([]byte, error) {
	if err := network.check(); err != nil {
		return nil, err
	}
	if len(inputNames) != network.relation.Variables() || !cascadeNameValid(model) || maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 27
	}
	for _, name := range inputNames {
		if !cascadeNameValid(name) {
			return nil, ErrFormat
		}
	}
	var builder strings.Builder
	write := func(text string) error { return appendBounded(&builder, text, maxBytes) }
	if err := write(".model " + model + "\n.inputs " + strings.Join(inputNames, " ") + "\n.outputs F\n"); err != nil {
		return nil, err
	}
	for stageIndex, stage := range network.stages {
		if err := write("#----------------- LUT #" + strconv.Itoa(stageIndex) + " ----------------------\n"); err != nil {
			return nil, err
		}
		localInputs := make([]string, 0, stage.PreviousBits+len(stage.InputVariables))
		for bit := 0; bit < stage.PreviousBits; bit++ {
			localInputs = append(localInputs, "LUT"+fmtCascadeIndex(stageIndex-1)+"_"+fmtCascadeIndex(bit))
		}
		for _, variable := range stage.InputVariables {
			localInputs = append(localInputs, inputNames[variable])
		}
		for outputIndex, table := range stage.Tables {
			outputName := "F"
			if !stage.Final {
				outputName = "LUT" + fmtCascadeIndex(stageIndex) + "_" + fmtCascadeIndex(outputIndex)
			}
			prefix := "L" + fmtCascadeIndex(stageIndex) + "_" + fmtCascadeIndex(outputIndex) + "_"
			if err := appendCascadeTruthBLIF(&builder, table, localInputs, outputName, prefix, maxBytes); err != nil {
				return nil, err
			}
		}
	}
	if err := write(".end\n"); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

func fmtCascadeIndex(index int) string {
	if index >= 0 && index < 10 {
		return "0" + strconv.Itoa(index)
	}
	return strconv.Itoa(index)
}
