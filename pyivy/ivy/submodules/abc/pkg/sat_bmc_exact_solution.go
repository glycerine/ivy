package abc

// Owned solution text and verifier from src/sat/bmc/bmcMaj3.c (Zyx test utilities).

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type ExactSolutionMetadata struct {
	Inputs, LUTSize, Nodes int
	Target                 []bool
}

func exactHexDigit(value byte) (int, bool) {
	switch {
	case value >= '0' && value <= '9':
		return int(value - '0'), true
	case value >= 'a' && value <= 'f':
		return int(value-'a') + 10, true
	case value >= 'A' && value <= 'F':
		return int(value-'A') + 10, true
	default:
		return 0, false
	}
}

// ParseExactSolutionName replaces Zyx_TestGetTruthTablePars. The basename is
// explicit because Go does not derive behavior from an opened process path.
func ParseExactSolutionName(name string, limits ExactSynthesisLimits) (ExactSolutionMetadata, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil {
		return ExactSolutionMetadata{}, err
	}
	if name == "" {
		return ExactSolutionMetadata{}, ErrFormat
	}
	if len(name) > limits.MaxBytes {
		return ExactSolutionMetadata{}, ErrLimit
	}
	hexLength := 0
	for hexLength < len(name) {
		if _, ok := exactHexDigit(name[hexLength]); !ok {
			break
		}
		hexLength++
	}
	inputByLength := map[int]int{1: 2, 2: 3, 4: 4, 8: 5, 16: 6, 32: 7, 64: 8}
	inputs, ok := inputByLength[hexLength]
	if !ok || hexLength >= len(name) || name[hexLength] != '-' {
		return ExactSolutionMetadata{}, ErrFormat
	}
	rest := name[hexLength+1:]
	separator := strings.IndexByte(rest, '-')
	if separator <= 0 {
		return ExactSolutionMetadata{}, ErrFormat
	}
	lutSize, err := strconv.Atoi(rest[:separator])
	if err != nil {
		return ExactSolutionMetadata{}, ErrFormat
	}
	nodeText := rest[separator+1:]
	if dot := strings.IndexByte(nodeText, '.'); dot >= 0 {
		if nodeText[dot:] != ".bool" {
			return ExactSolutionMetadata{}, ErrFormat
		}
		nodeText = nodeText[:dot]
	}
	nodes, err := strconv.Atoi(nodeText)
	if err != nil || lutSize < 2 || lutSize > 6 || nodes < 1 || inputs+nodes > 26 {
		return ExactSolutionMetadata{}, ErrFormat
	}
	target := make([]bool, 1<<inputs)
	if len(target) > limits.MaxAssignments {
		return ExactSolutionMetadata{}, ErrLimit
	}
	for fromRight := 0; fromRight < hexLength; fromRight++ {
		digit, _ := exactHexDigit(name[hexLength-1-fromRight])
		for bit := 0; bit < 4; bit++ {
			target[4*fromRight+bit] = digit&(1<<bit) != 0
		}
	}
	return ExactSolutionMetadata{Inputs: inputs, LUTSize: lutSize, Nodes: nodes, Target: target}, nil
}

func exactSolutionObjectCharacter(object, inputs int) (byte, error) {
	if object < 0 || object >= 26 {
		return 0, ErrDimension
	}
	if object < inputs {
		return byte('a' + object), nil
	}
	return byte('A' + object), nil
}

// MarshalExactSolutionNetwork preserves Zyx_ManPrintSolutionFile's compact
// LUT record and returns owned bytes instead of choosing or appending a file.
func MarshalExactSolutionNetwork(network ExactNetwork, limits ExactSynthesisLimits) ([]byte, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil {
		return nil, err
	}
	network = network.Clone()
	if network.ComplementOutput {
		if len(network.Gates) == 0 || network.Gates[len(network.Gates)-1].Majority {
			return nil, ErrDimension
		}
		patterns := 1 << len(network.Gates[len(network.Gates)-1].Fanins)
		mask := ^uint64(0)
		if patterns < 64 {
			mask = uint64(1)<<patterns - 1
		}
		network.Gates[len(network.Gates)-1].Truth ^= mask
		network.ComplementOutput = false
	}
	if err = network.Check(limits); err != nil {
		return nil, err
	}
	if network.Inputs+len(network.Gates) > 26 {
		return nil, ErrDimension
	}
	var result bytes.Buffer
	for row, gate := range network.Gates {
		if gate.Majority || len(gate.Fanins) < 2 || len(gate.Fanins) > 6 || !slicesSortedUnique(gate.Fanins) {
			return nil, ErrDimension
		}
		object := network.Inputs + row
		output, _ := exactSolutionObjectCharacter(object, network.Inputs)
		result.WriteByte(output)
		for pattern := (1 << len(gate.Fanins)) - 1; pattern >= 0; pattern-- {
			if gate.Truth&(uint64(1)<<pattern) != 0 {
				result.WriteByte('1')
			} else {
				result.WriteByte('0')
			}
		}
		for _, source := range gate.Fanins {
			character, characterErr := exactSolutionObjectCharacter(source, network.Inputs)
			if characterErr != nil {
				return nil, characterErr
			}
			result.WriteByte(character)
		}
		result.WriteByte('\n')
		if result.Len()+1 > limits.MaxBytes {
			return nil, ErrLimit
		}
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
	}
	result.WriteByte('\n')
	return result.Bytes(), nil
}

func slicesSortedUnique(values []int) bool {
	for i, value := range values {
		if value < 0 || i > 0 && values[i-1] >= value {
			return false
		}
	}
	return true
}

func MarshalExactSolutionDataset(networks []ExactNetwork, limits ExactSynthesisLimits) ([]byte, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil || len(networks) == 0 {
		return nil, ErrDimension
	}
	var result []byte
	for _, network := range networks {
		encoded, encodeErr := MarshalExactSolutionNetwork(network, limits)
		if encodeErr != nil {
			return nil, encodeErr
		}
		if len(encoded) > limits.MaxBytes-len(result) {
			return nil, ErrLimit
		}
		result = append(result, encoded...)
	}
	return result, nil
}

func parseExactSolutionNode(line string, metadata ExactSolutionMetadata, object int) (ExactGate, error) {
	wantLength := 1 + (1 << metadata.LUTSize) + metadata.LUTSize
	if len(line) != wantLength || line[0] != byte('A'+object) {
		return ExactGate{}, ErrFormat
	}
	gate := ExactGate{Fanins: make([]int, metadata.LUTSize)}
	truthText := line[1 : 1+(1<<metadata.LUTSize)]
	for index, character := range []byte(truthText) {
		if character != '0' && character != '1' {
			return ExactGate{}, ErrFormat
		}
		if character == '1' {
			pattern := len(truthText) - 1 - index
			gate.Truth |= uint64(1) << pattern
		}
	}
	faninText := line[1+(1<<metadata.LUTSize):]
	for index, character := range []byte(faninText) {
		var source int
		switch {
		case character >= 'a' && character <= 'z':
			source = int(character - 'a')
			if source >= metadata.Inputs {
				return ExactGate{}, ErrFormat
			}
		case character >= 'A' && character <= 'Z':
			source = int(character - 'A')
			if source < metadata.Inputs {
				return ExactGate{}, ErrFormat
			}
		default:
			return ExactGate{}, ErrFormat
		}
		if source >= object {
			return ExactGate{}, ErrFormat
		}
		gate.Fanins[index] = source
	}
	if !slicesSortedUnique(gate.Fanins) {
		return ExactGate{}, ErrFormat
	}
	return gate, nil
}

// VerifyExactSolutionDataset replaces Zyx_TestCreateTruthTables,
// Zyx_TestReadNode and Zyx_TestExact with a strict, owned-data verifier.
func VerifyExactSolutionDataset(name string, data []byte, limits ExactSynthesisLimits) ([]ExactNetwork, error) {
	limits, err := exactSynthesisLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	metadata, err := ParseExactSolutionName(name, limits)
	if err != nil {
		return nil, err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.ContainsRune(text, '\r') {
		return nil, ErrFormat
	}
	if !strings.HasSuffix(text, "\n\n") {
		return nil, ErrFormat
	}
	text = strings.TrimSuffix(text, "\n\n")
	blocks := strings.Split(text, "\n\n")
	if len(blocks) == 0 || len(blocks) > limits.MaxObjects {
		return nil, ErrLimit
	}
	result := make([]ExactNetwork, 0, len(blocks))
	for _, block := range blocks {
		if block == "" {
			return nil, ErrFormat
		}
		lines := strings.Split(block, "\n")
		if len(lines) != metadata.Nodes {
			return nil, ErrFormat
		}
		network := ExactNetwork{Inputs: metadata.Inputs, Gates: make([]ExactGate, metadata.Nodes)}
		for row, line := range lines {
			gate, parseErr := parseExactSolutionNode(line, metadata, metadata.Inputs+row)
			if parseErr != nil {
				return nil, fmt.Errorf("%w: structure %d node %d", parseErr, len(result), row)
			}
			network.Gates[row] = gate
		}
		truth, truthErr := network.Truth(limits)
		if truthErr != nil {
			return nil, truthErr
		}
		if len(truth) != len(metadata.Target) {
			return nil, ErrInvariant
		}
		for assignment := range truth {
			if truth[assignment] != metadata.Target[assignment] {
				return nil, fmt.Errorf("%w: structure %d assignment %d", ErrInvariant, len(result), assignment)
			}
		}
		result = append(result, network)
	}
	return result, nil
}
