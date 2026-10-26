package abc

// Derived from ABC src/base/abc/{abcNames.c,abcLatch.c,abcFanOrder.c,abcSop.c}.

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// ABCMetadataOptions bounds metadata and SOP work.
type ABCMetadataOptions struct {
	MaxVariables  int
	MaxCubes      int
	MaxBytes      int
	MaxOperations int
	Stop          func() bool
}

func (o ABCMetadataOptions) normalized() ABCMetadataOptions {
	if o.MaxVariables <= 0 {
		o.MaxVariables = 16
	}
	if o.MaxCubes <= 0 {
		o.MaxCubes = 1 << 20
	}
	if o.MaxBytes <= 0 {
		o.MaxBytes = 64 << 20
	}
	if o.MaxOperations <= 0 {
		o.MaxOperations = 1 << 26
	}
	return o
}

func abcMetadataTick(o ABCMetadataOptions, operations *int, count int) error {
	if count < 0 || *operations > o.MaxOperations-count {
		return ErrLimit
	}
	*operations += count
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	return nil
}

// ABCSOP is an owned, uniform-phase SOP cover. Phase is 0/1 for an off/on
// cover and x/n for the source's special XOR markers.
type ABCSOP struct {
	Variables int
	Cubes     []string
	Phase     byte
}

func (s ABCSOP) Clone() ABCSOP {
	s.Cubes = slices.Clone(s.Cubes)
	return s
}

func (s ABCSOP) Check(o ABCMetadataOptions) error {
	o = o.normalized()
	if s.Variables < 0 || s.Variables > o.MaxVariables || len(s.Cubes) == 0 || len(s.Cubes) > o.MaxCubes || !strings.ContainsRune("01xn", rune(s.Phase)) {
		return ErrDimension
	}
	if s.Variables > o.MaxBytes || o.MaxBytes-s.Variables < 3 {
		return ErrLimit
	}
	cubeBytes := s.Variables + 3
	if len(s.Cubes) > o.MaxBytes/cubeBytes {
		return ErrLimit
	}
	operations := 0
	for _, cube := range s.Cubes {
		if len(cube) != s.Variables {
			return ErrFormat
		}
		for i := 0; i < len(cube); i++ {
			if cube[i] != '0' && cube[i] != '1' && cube[i] != '-' {
				return ErrFormat
			}
		}
		if err := abcMetadataTick(o, &operations, len(cube)+1); err != nil {
			return err
		}
	}
	return nil
}

// ParseABCSOP parses the source's fixed-width "cube phase\n" representation.
func ParseABCSOP(text string, o ABCMetadataOptions) (ABCSOP, error) {
	o = o.normalized()
	if len(text) > o.MaxBytes {
		return ABCSOP{}, ErrLimit
	}
	if text == "" || strings.IndexByte(text, 0) >= 0 || text[len(text)-1] != '\n' {
		return ABCSOP{}, ErrFormat
	}
	lines := strings.Split(text[:len(text)-1], "\n")
	if len(lines) == 0 || len(lines) > o.MaxCubes {
		return ABCSOP{}, ErrLimit
	}
	variables := -1
	phase := byte(0)
	cubes := make([]string, 0, len(lines))
	operations := 0
	for _, line := range lines {
		if len(line) < 2 || line[len(line)-2] != ' ' || !strings.ContainsRune("01xn", rune(line[len(line)-1])) {
			return ABCSOP{}, ErrFormat
		}
		cube := line[:len(line)-2]
		if variables < 0 {
			variables = len(cube)
			phase = line[len(line)-1]
		}
		if len(cube) != variables || line[len(line)-1] != phase {
			return ABCSOP{}, ErrFormat
		}
		for i := range cube {
			if cube[i] != '0' && cube[i] != '1' && cube[i] != '-' {
				return ABCSOP{}, ErrFormat
			}
		}
		cubes = append(cubes, cube)
		if err := abcMetadataTick(o, &operations, len(line)+1); err != nil {
			return ABCSOP{}, err
		}
	}
	result := ABCSOP{Variables: variables, Cubes: cubes, Phase: phase}
	if err := result.Check(o); err != nil {
		return ABCSOP{}, err
	}
	return result, nil
}

func (s ABCSOP) Format(o ABCMetadataOptions) (string, error) {
	if err := s.Check(o); err != nil {
		return "", err
	}
	var b strings.Builder
	b.Grow(len(s.Cubes) * (s.Variables + 3))
	for _, cube := range s.Cubes {
		b.WriteString(cube)
		b.WriteByte(' ')
		b.WriteByte(s.Phase)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func (s ABCSOP) Evaluate(assignment uint64) (bool, error) {
	if err := s.Check(ABCMetadataOptions{}); err != nil {
		return false, err
	}
	if s.Variables < 64 && assignment >= uint64(1)<<uint(s.Variables) {
		return false, ErrDimension
	}
	covered := false
	for _, cube := range s.Cubes {
		match := true
		for variable := 0; variable < s.Variables; variable++ {
			value := byte('0' + ((assignment >> uint(variable)) & 1))
			if cube[variable] != '-' && cube[variable] != value {
				match = false
				break
			}
		}
		if match {
			covered = true
			break
		}
	}
	if s.Phase == '0' || s.Phase == 'n' {
		covered = !covered
	}
	return covered, nil
}

func (s ABCSOP) TruthWords(o ABCMetadataOptions) ([]uint64, error) {
	if err := s.Check(o); err != nil {
		return nil, err
	}
	if s.Variables > 20 {
		return nil, ErrLimit
	}
	assignments := 1 << uint(s.Variables)
	words := make([]uint64, (assignments+63)/64)
	for assignment := 0; assignment < assignments; assignment++ {
		value, err := s.Evaluate(uint64(assignment))
		if err != nil {
			return nil, err
		}
		if value {
			words[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	return words, nil
}

func ABCSOPConst(value bool) ABCSOP {
	phase := byte('0')
	if value {
		phase = '1'
	}
	return ABCSOP{Cubes: []string{""}, Phase: phase}
}

func ABCSOPAnd(variables int, complemented []bool, inverse bool) (ABCSOP, error) {
	if variables < 0 || complemented != nil && len(complemented) != variables {
		return ABCSOP{}, ErrDimension
	}
	cube := make([]byte, variables)
	for i := range cube {
		cube[i] = '1'
		if complemented != nil && complemented[i] {
			cube[i] = '0'
		}
	}
	phase := byte('1')
	if inverse {
		phase = '0'
	}
	return ABCSOP{Variables: variables, Cubes: []string{string(cube)}, Phase: phase}, nil
}

func ABCSOPOr(variables int, complemented []bool, inverse bool) (ABCSOP, error) {
	if variables < 0 || complemented != nil && len(complemented) != variables {
		return ABCSOP{}, ErrDimension
	}
	cube := make([]byte, variables)
	for i := range cube {
		cube[i] = '0'
		if complemented != nil && complemented[i] {
			cube[i] = '1'
		}
	}
	phase := byte('0')
	if inverse {
		phase = '1'
	}
	return ABCSOP{Variables: variables, Cubes: []string{string(cube)}, Phase: phase}, nil
}

func ABCSOPXOR() ABCSOP {
	return ABCSOP{Variables: 2, Cubes: []string{"01", "10"}, Phase: '1'}
}

func ABCSOPXNOR() ABCSOP {
	return ABCSOP{Variables: 2, Cubes: []string{"00", "11"}, Phase: '1'}
}

func ABCSOPMux() ABCSOP {
	return ABCSOP{Variables: 3, Cubes: []string{"11-", "0-1"}, Phase: '1'}
}

func ABCSOPFromTruth(variables int, words []uint64, o ABCMetadataOptions) (ABCSOP, error) {
	o = o.normalized()
	if variables < 0 || variables > o.MaxVariables || variables > 20 {
		return ABCSOP{}, ErrDimension
	}
	assignments := 1 << uint(variables)
	if len(words) != (assignments+63)/64 {
		return ABCSOP{}, ErrDimension
	}
	cubes := make([]string, 0)
	operations := 0
	for assignment := 0; assignment < assignments; assignment++ {
		if words[assignment>>6]&(uint64(1)<<uint(assignment&63)) == 0 {
			continue
		}
		if len(cubes) >= o.MaxCubes {
			return ABCSOP{}, ErrLimit
		}
		cube := make([]byte, variables)
		for variable := range cube {
			cube[variable] = byte('0' + ((assignment >> uint(variable)) & 1))
		}
		cubes = append(cubes, string(cube))
		if err := abcMetadataTick(o, &operations, variables+1); err != nil {
			return ABCSOP{}, err
		}
	}
	if len(cubes) == 0 {
		return ABCSOPConst(false), nil
	}
	return ABCSOP{Variables: variables, Cubes: cubes, Phase: '1'}, nil
}

func (s ABCSOP) ISOP() ([]uint32, error) {
	if err := s.Check(ABCMetadataOptions{MaxVariables: 16}); err != nil {
		return nil, err
	}
	result := make([]uint32, len(s.Cubes))
	for i, cube := range s.Cubes {
		for variable := s.Variables - 1; variable >= 0; variable-- {
			result[i] <<= 2
			switch cube[variable] {
			case '0':
				result[i] |= 1
			case '1':
				result[i] |= 2
			}
		}
	}
	return result, nil
}

func ABCSOPFromISOP(variables int, cubes []uint32, complemented bool) (ABCSOP, error) {
	if variables <= 0 || variables > 16 || len(cubes) == 0 {
		return ABCSOP{}, ErrDimension
	}
	result := ABCSOP{Variables: variables, Cubes: make([]string, len(cubes)), Phase: '1'}
	if complemented {
		result.Phase = '0'
	}
	for i, encoded := range cubes {
		cube := make([]byte, variables)
		for variable := range cube {
			switch (encoded >> uint(2*variable)) & 3 {
			case 0:
				cube[variable] = '-'
			case 1:
				cube[variable] = '0'
			case 2:
				cube[variable] = '1'
			default:
				return ABCSOP{}, ErrFormat
			}
		}
		result.Cubes[i] = string(cube)
	}
	return result, nil
}

func (s ABCSOP) Complement() ABCSOP {
	result := s.Clone()
	switch result.Phase {
	case '0':
		result.Phase = '1'
	case '1':
		result.Phase = '0'
	case 'x':
		result.Phase = 'n'
	case 'n':
		result.Phase = 'x'
	}
	return result
}

func (s ABCSOP) ComplementVariable(variable int) (ABCSOP, error) {
	if err := s.Check(ABCMetadataOptions{}); err != nil || variable < 0 || variable >= s.Variables {
		return ABCSOP{}, ErrDimension
	}
	result := s.Clone()
	for i, cube := range result.Cubes {
		copy := []byte(cube)
		if copy[variable] == '0' {
			copy[variable] = '1'
		} else if copy[variable] == '1' {
			copy[variable] = '0'
		}
		result.Cubes[i] = string(copy)
	}
	return result, nil
}

func (s ABCSOP) CareLiteral(variable int) (int, error) {
	if err := s.Check(ABCMetadataOptions{}); err != nil || variable < 0 || variable >= s.Variables {
		return -1, ErrDimension
	}
	for _, cube := range s.Cubes {
		if cube[variable] != '-' {
			return int(cube[variable] - '0'), nil
		}
	}
	return -1, nil
}

func (s ABCSOP) LiteralCount() int {
	count := 0
	for _, cube := range s.Cubes {
		for i := range cube {
			if cube[i] != '-' {
				count++
			}
		}
	}
	return count
}

func (s ABCSOP) IsConst0() bool  { return s.Variables == 0 && s.Phase == '0' }
func (s ABCSOP) IsConst1() bool  { return s.Variables == 0 && s.Phase == '1' }
func (s ABCSOP) IsXORType() bool { return s.Phase == 'x' || s.Phase == 'n' }
func (s ABCSOP) IsANDType() bool {
	return len(s.Cubes) == 1 && !strings.ContainsRune(s.Cubes[0], '-') && s.Phase == '1'
}
func (s ABCSOP) IsORType() bool {
	if len(s.Cubes) != s.Variables {
		return false
	}
	for _, cube := range s.Cubes {
		if s.Variables-strings.Count(cube, "-") != 1 {
			return false
		}
	}
	return true
}
func (s ABCSOP) IsBuffer() bool {
	if s.Variables != 1 || len(s.Cubes) != 1 {
		return false
	}
	return s.Cubes[0] == "1" && s.Phase == '1' || s.Cubes[0] == "0" && s.Phase == '0'
}
func (s ABCSOP) IsInverter() bool {
	if s.Variables != 1 || len(s.Cubes) != 1 {
		return false
	}
	return s.Cubes[0] == "0" && s.Phase == '1' || s.Cubes[0] == "1" && s.Phase == '0'
}

func (s ABCSOP) Permute(order []int) (ABCSOP, error) {
	if err := s.Check(ABCMetadataOptions{}); err != nil || len(order) != s.Variables {
		return ABCSOP{}, ErrDimension
	}
	seen := make([]bool, s.Variables)
	for _, old := range order {
		if old < 0 || old >= s.Variables || seen[old] {
			return ABCSOP{}, ErrDimension
		}
		seen[old] = true
	}
	result := s.Clone()
	for i, cube := range s.Cubes {
		permuted := make([]byte, s.Variables)
		for variable, old := range order {
			permuted[variable] = cube[old]
		}
		result.Cubes[i] = string(permuted)
	}
	return result, nil
}

func (s ABCSOP) SortCubes(weighted bool) ABCSOP {
	result := s.Clone()
	sort.SliceStable(result.Cubes, func(i, j int) bool {
		if weighted {
			left := result.Variables - strings.Count(result.Cubes[i], "-")
			right := result.Variables - strings.Count(result.Cubes[j], "-")
			if left != right {
				return left > right
			}
		}
		return result.Cubes[i] < result.Cubes[j]
	})
	return result
}

func abcCubeContains(general, specific string) bool {
	for i := range general {
		if general[i] != '-' && general[i] != specific[i] {
			return false
		}
	}
	return true
}

// Legit removes contained cubes and repeatedly merges distance-one opposite
// literals, preserving the source fast-extract normalization semantics.
func (s ABCSOP) Legit(o ABCMetadataOptions) (ABCSOP, bool, error) {
	if err := s.Check(o); err != nil {
		return ABCSOP{}, false, err
	}
	result := s.Clone()
	changed := false
	operations := 0
	for {
		pass := false
		for i := 0; i < len(result.Cubes); i++ {
			for j := i + 1; j < len(result.Cubes); j++ {
				if err := abcMetadataTick(o.normalized(), &operations, result.Variables+1); err != nil {
					return ABCSOP{}, false, err
				}
				if abcCubeContains(result.Cubes[i], result.Cubes[j]) {
					result.Cubes = append(result.Cubes[:j], result.Cubes[j+1:]...)
					j--
					pass, changed = true, true
					continue
				}
				if abcCubeContains(result.Cubes[j], result.Cubes[i]) {
					result.Cubes = append(result.Cubes[:i], result.Cubes[i+1:]...)
					i--
					pass, changed = true, true
					break
				}
				difference := -1
				valid := true
				for variable := 0; variable < result.Variables; variable++ {
					if result.Cubes[i][variable] == result.Cubes[j][variable] {
						continue
					}
					if difference >= 0 || result.Cubes[i][variable] == '-' || result.Cubes[j][variable] == '-' {
						valid = false
						break
					}
					difference = variable
				}
				if valid && difference >= 0 {
					merged := []byte(result.Cubes[i])
					merged[difference] = '-'
					result.Cubes[i], result.Cubes[j] = string(merged), string(merged)
					pass, changed = true, true
				}
			}
		}
		if !pass {
			break
		}
		unique := make([]string, 0, len(result.Cubes))
		for _, cube := range result.Cubes {
			if !slices.Contains(unique, cube) {
				unique = append(unique, cube)
			}
		}
		result.Cubes = unique
	}
	return result, changed, nil
}

func ABCDummyName(prefix string, number, digits int) (string, error) {
	if prefix == "" || strings.IndexByte(prefix, 0) >= 0 || number < 0 || digits < 0 || digits > 18 {
		return "", ErrDimension
	}
	return prefix + fmt.Sprintf("%0*d", digits, number), nil
}

func ABCObjectDisplayName(network *ABCNetwork, id int) (string, error) {
	object, err := network.Object(id)
	if err != nil {
		return "", err
	}
	if object.Name != "" {
		return object.Name, nil
	}
	prefix := "obj"
	switch object.Type {
	case ABCObjectPI:
		prefix = "pi"
	case ABCObjectPO:
		prefix = "po"
	case ABCObjectBI:
		prefix = "li"
	case ABCObjectBO:
		prefix = "lo"
	case ABCObjectLatch:
		prefix = "l"
	case ABCObjectNode:
		prefix = "n"
	case ABCObjectNet:
		prefix = "net"
	}
	return prefix + strconv.Itoa(id), nil
}

func ABCFaninNames(network *ABCNetwork, node int) ([]string, error) {
	object, err := network.Object(node)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(object.Fanins))
	for i, fanin := range object.Fanins {
		result[i], err = ABCObjectDisplayName(network, fanin.ID)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ABCOrderedObjectsByName(network *ABCNetwork, ids []int) ([]int, error) {
	if err := network.checkOpen(); err != nil {
		return nil, err
	}
	result := slices.Clone(ids)
	names := make(map[int]string, len(ids))
	for _, id := range ids {
		if _, found := names[id]; found {
			return nil, ErrDimension
		}
		name, err := ABCObjectDisplayName(network, id)
		if err != nil {
			return nil, err
		}
		names[id] = name
	}
	sort.SliceStable(result, func(i, j int) bool {
		if names[result[i]] != names[result[j]] {
			return names[result[i]] < names[result[j]]
		}
		return result[i] < result[j]
	})
	return result, nil
}

func ABCCompareTerminalNames(first, second *ABCNetwork) (bool, error) {
	if err := first.checkOpen(); err != nil {
		return false, err
	}
	if err := second.checkOpen(); err != nil {
		return false, err
	}
	if len(first.pis) != len(second.pis) || len(first.pos) != len(second.pos) ||
		len(abcLatchIDs(first)) != len(abcLatchIDs(second)) ||
		len(first.cis) != len(second.cis) || len(first.cos) != len(second.cos) {
		return false, nil
	}
	for i := range first.cis {
		left, err := ABCObjectDisplayName(first, first.cis[i])
		if err != nil {
			return false, err
		}
		right, err := ABCObjectDisplayName(second, second.cis[i])
		if err != nil || left != right {
			return false, err
		}
	}
	for i := range first.cos {
		left, err := ABCObjectDisplayName(first, first.cos[i])
		if err != nil {
			return false, err
		}
		right, err := ABCObjectDisplayName(second, second.cos[i])
		if err != nil || left != right {
			return false, err
		}
	}
	return true, nil
}

func abcClearNames(network *ABCNetwork) {
	for id := range network.objects {
		network.objects[id].Name = ""
	}
	network.nameToID = make(map[string]int)
	network.nameBytes = 0
}

// ABCShortNames replaces terminal/latch names with deterministic source-style
// names without exposing the native mutable name manager.
func ABCShortNames(network *ABCNetwork) error {
	if err := network.checkOpen(); err != nil {
		return err
	}
	backup := network.Clone()
	abcClearNames(network)
	piDigits := len(strconv.Itoa(max(0, len(network.pis)-1)))
	poDigits := len(strconv.Itoa(max(0, len(network.pos)-1)))
	for i, id := range network.pis {
		if err := network.SetObjectName(id, fmt.Sprintf("pi%0*d", piDigits, i)); err != nil {
			network.restore(backup)
			return err
		}
	}
	for i, id := range network.pos {
		if err := network.SetObjectName(id, fmt.Sprintf("po%0*d", poDigits, i)); err != nil {
			network.restore(backup)
			return err
		}
	}
	latches := abcLatchIDs(network)
	digits := len(strconv.Itoa(max(0, len(latches)-1)))
	for i, id := range latches {
		triple, err := ABCLatchTripleFor(network, id)
		if err != nil {
			network.restore(backup)
			return err
		}
		objects := []int{triple.Latch, triple.Input, triple.Output}
		names := []string{fmt.Sprintf("l%0*d", digits, i), fmt.Sprintf("li%0*d", digits, i), fmt.Sprintf("lo%0*d", digits, i)}
		for j, object := range objects {
			if err := network.SetObjectName(object, names[j]); err != nil {
				network.restore(backup)
				return err
			}
		}
	}
	if err := network.Validate(); err != nil {
		network.restore(backup)
		return err
	}
	return nil
}

func ABCTransferNames(source, destination *ABCNetwork, sourceIDs, destinationIDs []int, suffix string) error {
	if len(sourceIDs) != len(destinationIDs) || strings.IndexByte(suffix, 0) >= 0 {
		return ErrDimension
	}
	if err := source.checkOpen(); err != nil {
		return err
	}
	if err := destination.checkOpen(); err != nil {
		return err
	}
	backup := destination.Clone()
	for i := range sourceIDs {
		name, err := ABCObjectDisplayName(source, sourceIDs[i])
		if err == nil {
			err = destination.SetObjectName(destinationIDs[i], name+suffix)
		}
		if err != nil {
			destination.restore(backup)
			return err
		}
	}
	return nil
}

type ABCLatchTriple struct {
	Input  int
	Latch  int
	Output int
	Init   ABCLatchInit
}

func abcLatchIDs(network *ABCNetwork) []int {
	result := make([]int, 0)
	for _, id := range network.boxes {
		if network.objects[id].Alive && network.objects[id].Type == ABCObjectLatch {
			result = append(result, id)
		}
	}
	return result
}

func ABCLatchTripleFor(network *ABCNetwork, latch int) (ABCLatchTriple, error) {
	object, err := network.Object(latch)
	if err != nil || object.Type != ABCObjectLatch || len(object.Fanins) != 1 || len(object.Fanouts) != 1 {
		return ABCLatchTriple{}, ErrDimension
	}
	input := object.Fanins[0].ID
	inputObject, err := network.Object(input)
	if err != nil || inputObject.Type != ABCObjectBI || len(inputObject.Fanouts) != 1 || inputObject.Fanouts[0] != latch {
		return ABCLatchTriple{}, ErrInvariant
	}
	output := object.Fanouts[0]
	out, getErr := network.Object(output)
	if getErr != nil || out.Type != ABCObjectBO || len(out.Fanins) != 1 || out.Fanins[0].ID != latch {
		return ABCLatchTriple{}, ErrInvariant
	}
	return ABCLatchTriple{Input: input, Latch: latch, Output: output, Init: object.Init}, nil
}

func ABCLatchTriples(network *ABCNetwork) ([]ABCLatchTriple, error) {
	if err := network.checkOpen(); err != nil {
		return nil, err
	}
	latches := abcLatchIDs(network)
	result := make([]ABCLatchTriple, len(latches))
	for i, latch := range latches {
		var err error
		result[i], err = ABCLatchTripleFor(network, latch)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ABCLatchValues(network *ABCNetwork) ([]ABCLatchInit, error) {
	triples, err := ABCLatchTriples(network)
	if err != nil {
		return nil, err
	}
	result := make([]ABCLatchInit, len(triples))
	for i, triple := range triples {
		result[i] = triple.Init
	}
	return result, nil
}

func ABCLatchValueString(network *ABCNetwork) (string, error) {
	values, err := ABCLatchValues(network)
	if err != nil {
		return "", err
	}
	result := make([]byte, len(values))
	for i, value := range values {
		switch value {
		case ABCInitZero:
			result[i] = '0'
		case ABCInitOne:
			result[i] = '1'
		case ABCInitDC:
			result[i] = 'x'
		default:
			return "", ErrInvariant
		}
	}
	return string(result), nil
}

func ABCSetLatchValues(network *ABCNetwork, values []ABCLatchInit) error {
	if err := network.checkOpen(); err != nil {
		return err
	}
	latches := abcLatchIDs(network)
	if values != nil && len(values) != len(latches) {
		return ErrDimension
	}
	backup := network.Clone()
	for i, latch := range latches {
		if err := network.checkOpen(); err != nil {
			network.restore(backup)
			return err
		}
		value := ABCInitDC
		if values != nil {
			value = values[i]
		}
		if value < ABCInitZero || value > ABCInitDC {
			network.restore(backup)
			return ErrDimension
		}
		network.objects[latch].Init = value
	}
	return nil
}

func ABCAddNamedLatch(network *ABCNetwork, driver ABCObjectEdge, hasDriver bool, init ABCLatchInit, name string) (ABCLatchTriple, error) {
	if name == "" || strings.IndexByte(name, 0) >= 0 || init < ABCInitZero || init > ABCInitDC {
		return ABCLatchTriple{}, ErrDimension
	}
	backup := network.Clone()
	bi, latch, bo, err := network.CreateLatchBox(init)
	if err == nil && hasDriver {
		err = network.AddFanin(bi, driver)
	}
	if err == nil {
		objects := []int{latch, bi, bo}
		names := []string{name, name + "_li", name + "_lo"}
		for i, id := range objects {
			err = network.SetObjectName(id, names[i])
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		network.restore(backup)
		return ABCLatchTriple{}, err
	}
	return ABCLatchTriple{Input: bi, Latch: latch, Output: bo, Init: init}, nil
}

func abcLatchPredecessor(network *ABCNetwork, latch int) (int, bool) {
	triple, err := ABCLatchTripleFor(network, latch)
	if err != nil || len(network.objects[triple.Input].Fanins) != 1 {
		return -1, false
	}
	driver := network.objects[triple.Input].Fanins[0].ID
	if driver < 0 || driver >= len(network.objects) || network.objects[driver].Type != ABCObjectBO || len(network.objects[driver].Fanins) != 1 {
		return -1, false
	}
	predecessor := network.objects[driver].Fanins[0].ID
	return predecessor, predecessor >= 0 && predecessor < len(network.objects) && network.objects[predecessor].Type == ABCObjectLatch
}

func abcLatchIsSelfFeed(network *ABCNetwork, root int) (bool, error) {
	seen := make(map[int]bool)
	current := root
	for {
		if err := network.checkOpen(); err != nil {
			return false, err
		}
		if seen[current] {
			return false, nil
		}
		seen[current] = true
		next, ok := abcLatchPredecessor(network, current)
		if !ok {
			return false, nil
		}
		if next == root {
			return true, nil
		}
		current = next
	}
}

func ABCSelfFeedLatches(network *ABCNetwork) ([]int, error) {
	if err := network.checkOpen(); err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for _, root := range abcLatchIDs(network) {
		selfFeed, err := abcLatchIsSelfFeed(network, root)
		if err != nil {
			return nil, err
		}
		if selfFeed {
			result = append(result, root)
		}
	}
	return result, nil
}

func ABCRemoveSelfFeedLatches(network *ABCNetwork) (int, error) {
	if err := network.checkOpen(); err != nil {
		return 0, err
	}
	backup := network.Clone()
	removed := 0
	for _, latch := range abcLatchIDs(network) {
		selfFeed, err := abcLatchIsSelfFeed(network, latch)
		if err != nil {
			network.restore(backup)
			return 0, err
		}
		if !selfFeed {
			continue
		}
		triple, getErr := ABCLatchTripleFor(network, latch)
		if getErr != nil || len(network.objects[triple.Input].Fanins) != 1 {
			network.restore(backup)
			return 0, ErrInvariant
		}
		constant := 0
		if network.Type != ABCNetworkStrash {
			constant, err = network.CreateConst1()
			if err != nil {
				network.restore(backup)
				return 0, err
			}
		}
		driver := network.objects[triple.Input].Fanins[0].ID
		if err = network.PatchFanin(triple.Input, driver, ABCObjectEdge{ID: constant}); err != nil {
			network.restore(backup)
			return 0, err
		}
		removed++
	}
	return removed, nil
}

type ABCFaninOrder uint8

const (
	ABCFaninByID ABCFaninOrder = iota
	ABCFaninByColumn
	ABCFaninByLiteralCount
)

func ABCOrderNodeFanins(network *ABCNetwork, node int, mode ABCFaninOrder) error {
	if err := network.checkOpen(); err != nil {
		return err
	}
	object, err := network.Object(node)
	if err != nil || object.Type != ABCObjectNode || network.Function != ABCFunctionSOP {
		return ErrDimension
	}
	sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{})
	if err != nil || sop.Variables != len(object.Fanins) {
		return ErrFormat
	}
	order := make([]int, len(object.Fanins))
	for i := range order {
		order[i] = i
	}
	columns := make([]string, len(order))
	counts := make([]int, len(order))
	for variable := range order {
		var b strings.Builder
		for _, cube := range sop.Cubes {
			b.WriteByte(cube[variable])
			if cube[variable] != '-' {
				counts[variable]++
			}
		}
		columns[variable] = b.String()
	}
	sort.SliceStable(order, func(i, j int) bool {
		left, right := order[i], order[j]
		switch mode {
		case ABCFaninByID:
			if object.Fanins[left].ID != object.Fanins[right].ID {
				return object.Fanins[left].ID < object.Fanins[right].ID
			}
		case ABCFaninByColumn:
			if columns[left] != columns[right] {
				return columns[left] < columns[right]
			}
		case ABCFaninByLiteralCount:
			if counts[left] != counts[right] {
				return counts[left] < counts[right]
			}
			if columns[left] != columns[right] {
				return columns[left] < columns[right]
			}
		default:
			return false
		}
		return left < right
	})
	if mode > ABCFaninByLiteralCount {
		return ErrDimension
	}
	permuted, err := sop.Permute(order)
	if err != nil {
		return err
	}
	text, err := permuted.Format(ABCMetadataOptions{})
	if err != nil {
		return err
	}
	newFanins := make([]ABCObjectEdge, len(order))
	for i, old := range order {
		newFanins[i] = object.Fanins[old]
	}
	backup := network.Clone()
	network.objects[node].Fanins = newFanins
	network.objects[node].Function = text
	if err = network.Validate(); err != nil {
		network.restore(backup)
		return err
	}
	return nil
}

func ABCNormalizeNodeSOP(network *ABCNetwork, node int, weighted bool, o ABCMetadataOptions) (bool, error) {
	object, err := network.Object(node)
	if err != nil || object.Type != ABCObjectNode {
		return false, ErrDimension
	}
	sop, err := ParseABCSOP(object.Function, o)
	if err != nil {
		return false, err
	}
	legit, changed, err := sop.Legit(o)
	if err != nil {
		return false, err
	}
	legit = legit.SortCubes(weighted)
	text, err := legit.Format(o)
	if err != nil {
		return false, err
	}
	if text != object.Function {
		changed = true
	}
	if changed {
		if err = network.SetNodeFunction(node, text); err != nil {
			return false, err
		}
	}
	return changed, nil
}

// ABCOrderNetworkFanins applies a checked fanin/SOP permutation atomically to
// every live logic node.
func ABCOrderNetworkFanins(network *ABCNetwork, mode ABCFaninOrder) error {
	if err := network.checkOpen(); err != nil {
		return err
	}
	backup := network.Clone()
	for id, object := range network.objects {
		if !object.Alive || object.Type != ABCObjectNode {
			continue
		}
		if err := ABCOrderNodeFanins(network, id, mode); err != nil {
			network.restore(backup)
			return err
		}
	}
	return nil
}

// ABCNormalizeNetworkSOPs sorts and legitimizes every live SOP atomically.
func ABCNormalizeNetworkSOPs(network *ABCNetwork, weighted bool, o ABCMetadataOptions) (int, error) {
	if err := network.checkOpen(); err != nil {
		return 0, err
	}
	backup := network.Clone()
	changed := 0
	for id, object := range network.objects {
		if !object.Alive || object.Type != ABCObjectNode {
			continue
		}
		wasChanged, err := ABCNormalizeNodeSOP(network, id, weighted, o)
		if err != nil {
			network.restore(backup)
			return 0, err
		}
		if wasChanged {
			changed++
		}
	}
	return changed, nil
}
