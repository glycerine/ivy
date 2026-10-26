package abc

// Ported from src/opt/fxu/fxu.h, fxu.c, fxuCreate.c, fxuUpdate.c,
// fxuReduce.c, fxuPrint.c, and module.make.

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// FXUFlowOptions controls deterministic unate-cover extraction. Zero bounds
// select conservative defaults. Complement lookahead and the source's
// network adapter are intentionally outside this concrete cover API.
type FXUFlowOptions struct {
	Kernel         FXUOptions
	OnlySingles    bool
	OnlyDoubles    bool
	UseZero        bool
	UseComplement  bool
	CanonicalOnly  bool
	MaxExtractions int
	MaxSingles     int
	MaxPairs       int
	MinWeight      int
	MaxLiterals    int
}

// FXUCover is one unate sum-of-products cover. Literals are opaque positive
// variable IDs; callers represent opposite polarities using distinct IDs as
// the source matrix does.
type FXUCover struct {
	Output int
	Cubes  [][]int
}

// FXUDefinition is an extracted node. Definitions are topological: Variable
// equals InputVariables plus its index, and its cubes reference earlier IDs.
type FXUDefinition struct {
	Variable int
	Cubes    [][]int
}

type FXUStepKind uint8

const (
	FXUStepSingle FXUStepKind = iota + 1
	FXUStepDouble
)

type FXUFlowStep struct {
	Kind     FXUStepKind
	Variable int
	Weight   int
	Var1     int
	Var2     int
	Left     []int
	Right    []int
	Pairs    int
}

type FXUFlowResult struct {
	InputVariables  int
	Covers          []FXUCover
	Definitions     []FXUDefinition
	Steps           []FXUFlowStep
	InitialLiterals int
	FinalLiterals   int
}

type fxuCubeLocation struct {
	cover int
	cube  int
}

type fxuPairCandidate struct {
	cover  int
	first  int
	second int
	cube1  int
	cube2  int
	diff   int
	left   []int
	right  []int
}

func normalizeFXUFlowOptions(options FXUFlowOptions) FXUFlowOptions {
	if options.MaxExtractions <= 0 {
		options.MaxExtractions = 32
	}
	if options.MaxSingles <= 0 {
		options.MaxSingles = 1 << 16
	}
	if options.MaxPairs <= 0 {
		options.MaxPairs = 1 << 18
	}
	if options.MaxLiterals <= 0 {
		options.MaxLiterals = 1 << 16
	}
	return options
}

func fxuFlowBounds(options FXUFlowOptions) (variables, cubes, entries int) {
	variables, cubes, entries = options.Kernel.MaxVariables, options.Kernel.MaxCubes, options.Kernel.MaxEntries
	if variables <= 0 {
		variables = 1 << 16
	}
	if cubes <= 0 {
		cubes = 1 << 18
	}
	if entries <= 0 {
		entries = 1 << 22
	}
	return
}

func checkFXUInputBounds(inputVariables int, covers []FXUCover, options FXUFlowOptions) error {
	maxVariables, maxCubes, maxEntries := fxuFlowBounds(options)
	if inputVariables <= 0 {
		return ErrDimension
	}
	if len(covers) > maxVariables || inputVariables > maxVariables-len(covers) {
		return ErrLimit
	}
	cubeCount, entryCount := 0, 0
	for _, cover := range covers {
		if len(cover.Cubes) > maxCubes-cubeCount {
			return ErrLimit
		}
		cubeCount += len(cover.Cubes)
		for _, cube := range cover.Cubes {
			if len(cube) > maxEntries-entryCount {
				return ErrLimit
			}
			entryCount += len(cube)
		}
	}
	return nil
}

func cloneFXUCubes(cubes [][]int) [][]int {
	out := make([][]int, len(cubes))
	for i := range cubes {
		out[i] = append([]int(nil), cubes[i]...)
	}
	return out
}

func cloneFXUCovers(covers []FXUCover) []FXUCover {
	out := make([]FXUCover, len(covers))
	for i := range covers {
		out[i] = FXUCover{Output: covers[i].Output, Cubes: cloneFXUCubes(covers[i].Cubes)}
	}
	return out
}

func cloneFXUDefinitions(definitions []FXUDefinition) []FXUDefinition {
	out := make([]FXUDefinition, len(definitions))
	for i := range definitions {
		out[i] = FXUDefinition{Variable: definitions[i].Variable, Cubes: cloneFXUCubes(definitions[i].Cubes)}
	}
	return out
}

func normalizeFXUCube(cube []int, variableLimit int) ([]int, error) {
	out := append([]int(nil), cube...)
	sort.Ints(out)
	for i, literal := range out {
		if literal < 0 || literal >= variableLimit {
			return nil, ErrDimension
		}
		if i != 0 && out[i-1] == literal {
			return nil, ErrInvariant
		}
	}
	return out, nil
}

func fxuCubeSubset(a, b []int) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
			j++
		} else if a[i] > b[j] {
			j++
		} else {
			return false
		}
	}
	return i == len(a)
}

func canonicalizeFXUCubes(cubes [][]int, variableLimit int) ([][]int, error) {
	normalized := make([][]int, 0, len(cubes))
	for _, cube := range cubes {
		copyCube, err := normalizeFXUCube(cube, variableLimit)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, copyCube)
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if len(normalized[i]) != len(normalized[j]) {
			return len(normalized[i]) < len(normalized[j])
		}
		return compareInts(normalized[i], normalized[j]) < 0
	})
	kept := normalized[:0]
	for _, cube := range normalized {
		redundant := false
		for _, prior := range kept {
			if fxuCubeSubset(prior, cube) {
				redundant = true
				break
			}
		}
		if !redundant {
			kept = append(kept, cube)
		}
	}
	return append([][]int(nil), kept...), nil
}

func validateFXUProgram(inputVariables int, covers []FXUCover, definitions []FXUDefinition) error {
	if inputVariables <= 0 || len(definitions) > math.MaxInt-inputVariables {
		return ErrDimension
	}
	seenOutputs := make(map[int]bool, len(covers))
	limit := inputVariables + len(definitions)
	for _, cover := range covers {
		if cover.Output < 0 || seenOutputs[cover.Output] {
			return ErrInvariant
		}
		seenOutputs[cover.Output] = true
		for _, cube := range cover.Cubes {
			if _, err := normalizeFXUCube(cube, limit); err != nil {
				return err
			}
		}
	}
	for i, definition := range definitions {
		if definition.Variable != inputVariables+i {
			return ErrInvariant
		}
		for _, cube := range definition.Cubes {
			if _, err := normalizeFXUCube(cube, definition.Variable); err != nil {
				return err
			}
		}
	}
	return nil
}

func fxuLiteralCount(covers []FXUCover, definitions []FXUDefinition) int {
	total := 0
	for _, cover := range covers {
		for _, cube := range cover.Cubes {
			total += len(cube)
		}
	}
	for _, definition := range definitions {
		for _, cube := range definition.Cubes {
			total += len(cube)
		}
	}
	return total
}

func fxuFlowTick(options FXUFlowOptions, operations *int, amount int) error {
	maximum := options.Kernel.MaxOperations
	if maximum <= 0 {
		maximum = 1 << 27
	}
	if amount < 0 || *operations > maximum-amount {
		return ErrLimit
	}
	if options.Kernel.Stop != nil && options.Kernel.Stop() {
		return ErrCancelled
	}
	*operations += amount
	return nil
}

func buildFXUFlowMatrix(covers []FXUCover, variableCount int, options FXUFlowOptions, operations *int) (*FXUMatrix, []fxuCubeLocation, error) {
	kernel := options.Kernel
	kernel.Stop = options.Kernel.Stop
	matrix := NewFXUMatrix(kernel)
	// The flow and its matrix are one logical operation. Continue the caller's
	// counter instead of giving each rebuilt matrix a fresh full work budget.
	matrix.operations = *operations
	defer func() { *operations = matrix.operations }()
	for i := 0; i < variableCount+len(covers); i++ {
		if _, err := matrix.AddVariable(); err != nil {
			return nil, nil, err
		}
	}
	locations := make([]fxuCubeLocation, 0)
	for coverIndex, cover := range covers {
		for cubeIndex, cube := range cover.Cubes {
			if err := fxuFlowTick(options, &matrix.operations, 1+len(cube)); err != nil {
				return nil, nil, err
			}
			id, err := matrix.AddCube(variableCount+coverIndex, cubeIndex)
			if err != nil {
				return nil, nil, err
			}
			for _, literal := range cube {
				if err = matrix.AddLiteral(id, literal); err != nil {
					return nil, nil, err
				}
			}
			locations = append(locations, fxuCubeLocation{cover: coverIndex, cube: cubeIndex})
		}
	}
	if !options.OnlyDoubles {
		if err := matrix.ComputeSingles(options.UseZero, options.MaxSingles); err != nil {
			return nil, nil, err
		}
	}
	if options.OnlySingles {
		return matrix, locations, nil
	}
	candidates := make([]fxuPairCandidate, 0)
	baseID := 0
	for coverIndex, cover := range covers {
		for first := 0; first < len(cover.Cubes); first++ {
			for second := first + 1; second < len(cover.Cubes); second++ {
				_, left, right := residues(cover.Cubes[first], cover.Cubes[second])
				if len(left) == 0 || len(right) == 0 || len(left)+len(right) > options.MaxLiterals {
					continue
				}
				if options.CanonicalOnly && (len(left) != 1 || len(right) != 1) {
					continue
				}
				candidates = append(candidates, fxuPairCandidate{cover: coverIndex, first: first, second: second, cube1: baseID + first, cube2: baseID + second, diff: len(left) + len(right), left: left, right: right})
			}
		}
		baseID += len(cover.Cubes)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].diff != candidates[j].diff {
			return candidates[i].diff < candidates[j].diff
		}
		if candidates[i].cover != candidates[j].cover {
			return candidates[i].cover < candidates[j].cover
		}
		if candidates[i].first != candidates[j].first {
			return candidates[i].first < candidates[j].first
		}
		return candidates[i].second < candidates[j].second
	})
	if len(candidates) > options.MaxPairs {
		candidates = candidates[:options.MaxPairs]
	}
	usedByKey := make(map[string]map[int]bool)
	for _, candidate := range candidates {
		if err := fxuFlowTick(options, &matrix.operations, 1); err != nil {
			return nil, nil, err
		}
		left, right := candidate.left, candidate.right
		if compareInts(left, right) > 0 {
			left, right = right, left
		}
		key := fxuDivisorKey(left, right)
		used := usedByKey[key]
		if used == nil {
			used = make(map[int]bool)
			usedByKey[key] = used
		}
		if used[candidate.cube1] || used[candidate.cube2] {
			continue
		}
		if _, err := matrix.AddDouble(candidate.cube1, candidate.cube2); err != nil {
			return nil, nil, err
		}
		used[candidate.cube1], used[candidate.cube2] = true, true
	}
	if err := matrix.Validate(); err != nil {
		return nil, nil, err
	}
	return matrix, locations, nil
}

func bestFXUFlowSelection(matrix *FXUMatrix, options FXUFlowOptions) FXUSelection {
	minimum := -int(^uint(0)>>1) - 1
	result := FXUSelection{Weight: minimum, SingleID: -1, DoubleID: -1, Var1: -1, Var2: -1}
	found := false
	if !options.OnlyDoubles {
		for _, single := range matrix.singles {
			if !found || better(single.Weight, single.ID, result.Weight, result.SingleID) {
				result = FXUSelection{Weight: single.Weight, SingleID: single.ID, DoubleID: -1, Var1: single.Var1, Var2: single.Var2}
				found = true
			}
		}
	}
	if !options.OnlySingles {
		for _, double := range matrix.doubles {
			if len(double.Pairs) == 0 {
				continue
			}
			if !found || double.Weight > result.Weight || (result.SingleID < 0 && double.Weight == result.Weight && (result.DoubleID < 0 || double.ID < result.DoubleID)) {
				result = FXUSelection{Weight: double.Weight, SingleID: -1, DoubleID: double.ID, Var1: -1, Var2: -1}
				found = true
			}
		}
	}
	if !found {
		result.Weight = -1
	}
	return result
}

func acceptFXUWeight(weight int, options FXUFlowOptions) bool {
	return weight > options.MinWeight || (weight == 0 && options.UseZero)
}

func containsFXULiteral(cube []int, literal int) bool {
	i := sort.SearchInts(cube, literal)
	return i < len(cube) && cube[i] == literal
}

func replaceFXUSingle(covers []FXUCover, var1, var2, replacement int) (int, error) {
	count := 0
	for coverIndex := range covers {
		for cubeIndex, cube := range covers[coverIndex].Cubes {
			if !containsFXULiteral(cube, var1) || !containsFXULiteral(cube, var2) {
				continue
			}
			next := make([]int, 0, len(cube)-1)
			for _, literal := range cube {
				if literal != var1 && literal != var2 {
					next = append(next, literal)
				}
			}
			next, _ = sortedInsertUnique(next, replacement)
			covers[coverIndex].Cubes[cubeIndex] = next
			count++
		}
	}
	if count == 0 {
		return 0, ErrInvariant
	}
	return count, nil
}

func replaceFXUDouble(covers []FXUCover, locations []fxuCubeLocation, divisor FXUDouble, replacement int) (int, error) {
	replacements := make(map[[2]int][]int, len(divisor.Pairs))
	deletions := make(map[[2]int]bool, len(divisor.Pairs))
	for _, pair := range divisor.Pairs {
		if pair.Cube1 < 0 || pair.Cube2 < 0 || pair.Cube1 >= len(locations) || pair.Cube2 >= len(locations) {
			return 0, ErrInvariant
		}
		first, second := locations[pair.Cube1], locations[pair.Cube2]
		if first.cover != second.cover {
			return 0, ErrInvariant
		}
		use, remove := first, second
		if use.cube > remove.cube {
			use, remove = remove, use
		}
		if _, exists := replacements[[2]int{use.cover, use.cube}]; exists || deletions[[2]int{use.cover, use.cube}] || deletions[[2]int{remove.cover, remove.cube}] {
			return 0, ErrInvariant
		}
		base, _, _ := residues(covers[use.cover].Cubes[use.cube], covers[remove.cover].Cubes[remove.cube])
		base, _ = sortedInsertUnique(base, replacement)
		replacements[[2]int{use.cover, use.cube}] = base
		deletions[[2]int{remove.cover, remove.cube}] = true
	}
	for coverIndex := range covers {
		next := make([][]int, 0, len(covers[coverIndex].Cubes))
		for cubeIndex, cube := range covers[coverIndex].Cubes {
			key := [2]int{coverIndex, cubeIndex}
			if deletions[key] {
				continue
			}
			if replacementCube, ok := replacements[key]; ok {
				next = append(next, replacementCube)
			} else {
				next = append(next, cube)
			}
		}
		covers[coverIndex].Cubes = next
	}
	return len(divisor.Pairs), nil
}

// ExtractFXU performs bounded deterministic single- and double-cube
// extraction. Caller-owned covers are never mutated, including on failure.
func ExtractFXU(inputVariables int, input []FXUCover, rawOptions FXUFlowOptions) (FXUFlowResult, error) {
	options := normalizeFXUFlowOptions(rawOptions)
	if options.OnlySingles && options.OnlyDoubles {
		return FXUFlowResult{}, ErrInvariant
	}
	if options.UseComplement {
		return FXUFlowResult{}, ErrUnsupported
	}
	if err := checkFXUInputBounds(inputVariables, input, options); err != nil {
		return FXUFlowResult{}, err
	}
	operations := 0
	covers := cloneFXUCovers(input)
	for i := range covers {
		for _, cube := range covers[i].Cubes {
			if err := fxuFlowTick(options, &operations, 1+len(cube)); err != nil {
				return FXUFlowResult{}, err
			}
		}
		canonical, err := canonicalizeFXUCubes(covers[i].Cubes, inputVariables)
		if err != nil {
			return FXUFlowResult{}, err
		}
		covers[i].Cubes = canonical
	}
	if err := validateFXUProgram(inputVariables, covers, nil); err != nil {
		return FXUFlowResult{}, err
	}
	result := FXUFlowResult{InputVariables: inputVariables, Covers: covers, InitialLiterals: fxuLiteralCount(covers, nil)}
	for extraction := 0; extraction < options.MaxExtractions; extraction++ {
		if err := fxuFlowTick(options, &operations, 1); err != nil {
			return FXUFlowResult{}, err
		}
		maxVariables, _, maxEntries := fxuFlowBounds(options)
		if len(result.Definitions) > maxVariables-inputVariables-len(result.Covers) {
			return FXUFlowResult{}, ErrLimit
		}
		variableCount := inputVariables + len(result.Definitions)
		matrix, locations, err := buildFXUFlowMatrix(result.Covers, variableCount, options, &operations)
		if err != nil {
			return FXUFlowResult{}, err
		}
		selection := bestFXUFlowSelection(matrix, options)
		if !acceptFXUWeight(selection.Weight, options) {
			break
		}
		// The current matrix may exactly fill the variable budget. An accepted
		// divisor needs one additional definition before another rebuild occurs.
		if len(result.Definitions) >= maxVariables-inputVariables-len(result.Covers) {
			return FXUFlowResult{}, ErrLimit
		}
		nextCovers := cloneFXUCovers(result.Covers)
		nextDefinitions := cloneFXUDefinitions(result.Definitions)
		newVariable := inputVariables + len(nextDefinitions)
		step := FXUFlowStep{Variable: newVariable, Weight: selection.Weight, Var1: -1, Var2: -1}
		if selection.SingleID >= 0 {
			single := matrix.singles[selection.SingleID]
			if _, err = replaceFXUSingle(nextCovers, single.Var1, single.Var2, newVariable); err != nil {
				return FXUFlowResult{}, err
			}
			nextDefinitions = append(nextDefinitions, FXUDefinition{Variable: newVariable, Cubes: [][]int{{single.Var1, single.Var2}}})
			step.Kind, step.Var1, step.Var2 = FXUStepSingle, single.Var1, single.Var2
		} else if selection.DoubleID >= 0 {
			divisor := matrix.doubles[selection.DoubleID]
			pairs, replaceErr := replaceFXUDouble(nextCovers, locations, divisor, newVariable)
			if replaceErr != nil {
				return FXUFlowResult{}, replaceErr
			}
			nextDefinitions = append(nextDefinitions, FXUDefinition{Variable: newVariable, Cubes: [][]int{append([]int(nil), divisor.Left...), append([]int(nil), divisor.Right...)}})
			step.Kind, step.Left, step.Right, step.Pairs = FXUStepDouble, append([]int(nil), divisor.Left...), append([]int(nil), divisor.Right...), pairs
		} else {
			break
		}
		for i := range nextCovers {
			for _, cube := range nextCovers[i].Cubes {
				if err = fxuFlowTick(options, &operations, 1+len(cube)); err != nil {
					return FXUFlowResult{}, err
				}
			}
			nextCovers[i].Cubes, err = canonicalizeFXUCubes(nextCovers[i].Cubes, newVariable+1)
			if err != nil {
				return FXUFlowResult{}, err
			}
		}
		if err = validateFXUProgram(inputVariables, nextCovers, nextDefinitions); err != nil {
			return FXUFlowResult{}, err
		}
		before := fxuLiteralCount(result.Covers, result.Definitions)
		after := fxuLiteralCount(nextCovers, nextDefinitions)
		if after > maxEntries || after > before-step.Weight {
			return FXUFlowResult{}, ErrInvariant
		}
		result.Covers, result.Definitions = nextCovers, nextDefinitions
		result.Steps = append(result.Steps, step)
	}
	result.FinalLiterals = fxuLiteralCount(result.Covers, result.Definitions)
	return result, nil
}

func evaluateFXUCubes(cubes [][]int, values []bool) (bool, error) {
	for _, cube := range cubes {
		product := true
		for _, literal := range cube {
			if literal < 0 || literal >= len(values) {
				return false, ErrDimension
			}
			product = product && values[literal]
		}
		if product {
			return true, nil
		}
	}
	return false, nil
}

// Evaluate evaluates all output covers from concrete input values.
func (result *FXUFlowResult) Evaluate(inputs []bool) ([]bool, error) {
	if result == nil || len(inputs) != result.InputVariables || validateFXUProgram(result.InputVariables, result.Covers, result.Definitions) != nil {
		return nil, ErrDimension
	}
	values := append([]bool(nil), inputs...)
	for _, definition := range result.Definitions {
		value, err := evaluateFXUCubes(definition.Cubes, values)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	outputs := make([]bool, len(result.Covers))
	for i, cover := range result.Covers {
		value, err := evaluateFXUCubes(cover.Cubes, values)
		if err != nil {
			return nil, err
		}
		outputs[i] = value
	}
	return outputs, nil
}

// FormatFXUFlow returns deterministic owned text in place of global printing
// and fixed output files.
func (result *FXUFlowResult) FormatFXUFlow() (string, error) {
	if result == nil || validateFXUProgram(result.InputVariables, result.Covers, result.Definitions) != nil {
		return "", ErrInvariant
	}
	var b strings.Builder
	b.WriteString("fxu inputs=")
	b.WriteString(strconv.Itoa(result.InputVariables))
	b.WriteString(" extracted=")
	b.WriteString(strconv.Itoa(len(result.Definitions)))
	b.WriteString(" literals=")
	b.WriteString(strconv.Itoa(result.InitialLiterals))
	b.WriteByte('-')
	b.WriteString(strconv.Itoa(result.FinalLiterals))
	b.WriteByte('\n')
	for _, step := range result.Steps {
		b.WriteString("v")
		b.WriteString(strconv.Itoa(step.Variable))
		b.WriteString(" w=")
		b.WriteString(strconv.Itoa(step.Weight))
		if step.Kind == FXUStepSingle {
			b.WriteString(" single ")
			b.WriteString(strconv.Itoa(step.Var1))
			b.WriteByte(',')
			b.WriteString(strconv.Itoa(step.Var2))
		} else {
			b.WriteString(" double pairs=")
			b.WriteString(strconv.Itoa(step.Pairs))
			b.WriteString(" left=")
			b.WriteString(fxuDivisorKey(step.Left, step.Right))
		}
		b.WriteByte('\n')
	}
	return b.String(), nil
}
