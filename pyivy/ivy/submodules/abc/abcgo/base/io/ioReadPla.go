package io

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

type plaRow struct {
	inputs  string
	outputs string
	line    blifLine
}

type plaDocument struct {
	name        string
	nInputs     int
	nOutputs    int
	nProducts   int
	inputNames  []string
	outputNames []string
	rows        []plaRow
}

func init() {
	RegisterFormat(IO_FILE_PLA, readPLARegistered, writePLARegistered)
	RegisterFormat(IO_FILE_MOPLA, readPLARegistered, writeMOPLARegistered)
}

func readPLARegistered(filename string, check bool) (*abc.Network, error) {
	return readPLAFile(filename, 0, 0, 0, false, check)
}

func parsePLADocument(filename, text string) (*plaDocument, error) {
	lines, err := blifLines(filename, text)
	if err != nil {
		return nil, err
	}
	doc := &plaDocument{name: filename, nInputs: -1, nOutputs: -1, nProducts: -1}
	ended := false
	for _, line := range lines {
		if ended {
			break
		}
		fields := strings.Fields(strings.ReplaceAll(line.text, "|", " "))
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case ".e", ".end":
			ended = true
			continue
		case ".type":
			continue
		case ".model":
			if len(fields) != 2 {
				return nil, blifError(filename, line, ".model expects one name")
			}
			doc.name = fields[1]
			continue
		case ".i", ".o", ".p":
			if len(fields) != 2 {
				return nil, blifError(filename, line, "%s expects one integer", fields[0])
			}
			value, parseErr := strconv.Atoi(fields[1])
			if parseErr != nil || value < 0 {
				return nil, blifError(filename, line, "invalid %s count %q", fields[0], fields[1])
			}
			switch fields[0] {
			case ".i":
				doc.nInputs = value
			case ".o":
				doc.nOutputs = value
			case ".p":
				doc.nProducts = value
			}
			continue
		case ".ilb":
			doc.inputNames = append([]string(nil), fields[1:]...)
			continue
		case ".ob":
			doc.outputNames = append([]string(nil), fields[1:]...)
			continue
		}
		if len(fields) != 2 {
			return nil, blifError(filename, line, "expected input and output cubes")
		}
		doc.rows = append(doc.rows, plaRow{inputs: fields[0], outputs: fields[1], line: line})
	}
	if doc.nInputs < 0 || doc.nOutputs < 0 {
		return nil, fmt.Errorf("%s: missing .i or .o directive", filename)
	}
	if len(doc.inputNames) != 0 && len(doc.inputNames) != doc.nInputs {
		return nil, fmt.Errorf("%s: .ilb has %d names, .i declares %d inputs", filename, len(doc.inputNames), doc.nInputs)
	}
	if len(doc.outputNames) != 0 && len(doc.outputNames) != doc.nOutputs {
		return nil, fmt.Errorf("%s: .ob has %d names, .o declares %d outputs", filename, len(doc.outputNames), doc.nOutputs)
	}
	for _, row := range doc.rows {
		if len(row.inputs) != doc.nInputs {
			return nil, blifError(filename, row.line, "input cube length %d differs from .i count %d", len(row.inputs), doc.nInputs)
		}
		if len(row.outputs) != doc.nOutputs {
			return nil, blifError(filename, row.line, "output cube length %d differs from .o count %d", len(row.outputs), doc.nOutputs)
		}
		for i, c := range row.inputs {
			if c != '0' && c != '1' && c != '-' {
				return nil, blifError(filename, row.line, "invalid input cube symbol %q at column %d", c, i+1)
			}
		}
		for i, c := range row.outputs {
			if c != '0' && c != '1' && c != '-' {
				return nil, blifError(filename, row.line, "invalid output cube symbol %q at column %d", c, i+1)
			}
		}
	}
	return doc, nil
}

func readPLAFile(filename string, fZeros, fBoth, fOnDc int32, skipPreprocess, check bool) (*abc.Network, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	doc, err := parsePLADocument(filename, string(contents))
	if err != nil {
		return nil, err
	}
	return plaDocumentNetwork(doc, fZeros, fBoth, fOnDc, skipPreprocess, check)
}

func Io_ReadPla(filename string, fZeros, fBoth, fOnDc, fSkipPrepro, fCheck int32) *Abc_Ntk_t {
	network, err := readPLAFile(filename, fZeros, fBoth, fOnDc, fSkipPrepro != 0, fCheck != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return network
}

// Io_ReadPlaNetwork retains the legacy pointer-returning API over the native
// file parser. Errors are printed and represented by a nil network.
func Io_ReadPlaNetwork(filename string, fZeros, fBoth, fOnDc, fSkipPrepro int32) *Abc_Ntk_t {
	return Io_ReadPla(filename, fZeros, fBoth, fOnDc, fSkipPrepro, 0)
}

func plaDocumentNetwork(doc *plaDocument, fZeros, fBoth, fOnDc int32, skipPreprocess, check bool) (*abc.Network, error) {
	inputNames := append([]string(nil), doc.inputNames...)
	if len(inputNames) == 0 {
		width := len(strconv.Itoa(max(0, doc.nInputs-1)))
		if width < 1 {
			width = 1
		}
		for i := 0; i < doc.nInputs; i++ {
			inputNames = append(inputNames, fmt.Sprintf("x%0*d", width, i))
		}
	}
	outputNames := append([]string(nil), doc.outputNames...)
	if len(outputNames) == 0 {
		width := len(strconv.Itoa(max(0, doc.nOutputs-1)))
		if width < 1 {
			width = 1
		}
		for i := 0; i < doc.nOutputs; i++ {
			outputNames = append(outputNames, fmt.Sprintf("z%0*d", width, i))
		}
	}
	network := abc.Abc_NtkStartRead(doc.name)
	network.Name = doc.name
	builder := ioBuilder(network)
	for _, name := range inputNames {
		if err := builder.input(name); err != nil {
			abc.Abc_NtkDelete(network)
			return nil, fmt.Errorf("%s: %w", doc.name, err)
		}
	}
	for _, name := range outputNames {
		if err := builder.output(name); err != nil {
			abc.Abc_NtkDelete(network)
			return nil, fmt.Errorf("%s: %w", doc.name, err)
		}
	}
	covers := make([][]string, doc.nOutputs)
	for _, row := range doc.rows {
		for output, value := range []byte(row.outputs) {
			selected := value == '1'
			switch {
			case fZeros != 0:
				selected = value == '0'
			case fBoth != 0:
				selected = value == '0' || value == '1'
			case fOnDc != 0:
				selected = value == '-' || value == '1'
			}
			if selected {
				covers[output] = append(covers[output], row.inputs)
			}
		}
	}
	for output, name := range outputNames {
		if !skipPreprocess {
			covers[output] = preprocessPLACubes(covers[output])
		}
		cover := plaSOP(covers[output])
		inputs := inputNames
		if len(covers[output]) == 0 {
			inputs = nil
		}
		if _, err := builder.node(append(append([]string(nil), inputs...), name), cover); err != nil {
			abc.Abc_NtkDelete(network)
			return nil, fmt.Errorf("%s: output %q: %w", doc.name, name, err)
		}
	}
	abc.Abc_NtkFinalizeRead(network)
	if check {
		if abc.Abc_NtkCheckRead(network) == 0 {
			abc.Abc_NtkDelete(network)
			return nil, fmt.Errorf("%s: network check failed", doc.name)
		}
	}
	return network, nil
}

func plaSOP(cubes []string) string {
	if len(cubes) == 0 {
		return " 0\n"
	}
	var result strings.Builder
	for _, cube := range cubes {
		result.WriteString(cube)
		result.WriteString(" 1\n")
	}
	return result.String()
}

func preprocessPLACubes(cubes []string) []string {
	current := append([]string(nil), cubes...)
	for {
		current = reducePLASubsumed(current)
		merged := false
		used := make([]bool, len(current))
		next := make([]string, 0, len(current))
		for i := range current {
			if used[i] {
				continue
			}
			for j := i + 1; j < len(current); j++ {
				if used[j] {
					continue
				}
				if combined, ok := mergePLACubes(current[i], current[j]); ok {
					used[i], used[j] = true, true
					next = append(next, combined)
					merged = true
					break
				}
			}
			if !used[i] {
				used[i] = true
				next = append(next, current[i])
			}
		}
		current = next
		if !merged {
			return reducePLASubsumed(current)
		}
	}
}

func reducePLASubsumed(cubes []string) []string {
	unique := make([]string, 0, len(cubes))
	seen := make(map[string]bool, len(cubes))
	for _, cube := range cubes {
		if !seen[cube] {
			seen[cube] = true
			unique = append(unique, cube)
		}
	}
	keep := make([]bool, len(unique))
	for i := range keep {
		keep[i] = true
	}
	for i := range unique {
		for j := range unique {
			if i != j && keep[i] && plaCubeContains(unique[j], unique[i]) {
				keep[i] = false
			}
		}
	}
	result := make([]string, 0, len(unique))
	for i, cube := range unique {
		if keep[i] {
			result = append(result, cube)
		}
	}
	return result
}

func plaCubeContains(big, small string) bool {
	if len(big) != len(small) {
		return false
	}
	for i := range big {
		if big[i] != '-' && big[i] != small[i] {
			return false
		}
	}
	return true
}

func mergePLACubes(a, b string) (string, bool) {
	if len(a) != len(b) {
		return "", false
	}
	difference := -1
	for i := range a {
		if a[i] == b[i] {
			continue
		}
		if a[i] == '-' || b[i] == '-' || difference >= 0 {
			return "", false
		}
		difference = i
	}
	if difference < 0 {
		return "", false
	}
	merged := []byte(a)
	merged[difference] = '-'
	return string(merged), true
}

// Io_ReadPlaCubePreprocess reduces a single positive SOP cover in place.
func Io_ReadPlaCubePreprocess(cover *vec.Vec_Str_t, _ int32, _ int32) {
	if cover == nil {
		return
	}
	text := string(vec.Vec_StrArray(cover))
	text = strings.TrimSuffix(text, "\x00")
	var cubes []string
	for _, line := range strings.Split(text, "\n") {
		if at := strings.IndexByte(line, ' '); at >= 0 && len(line) > at+1 && line[at+1] == '1' {
			cubes = append(cubes, line[:at])
		}
	}
	vec.Vec_StrClear(cover)
	for _, cube := range preprocessPLACubes(cubes) {
		for _, c := range []byte(cube + " 1\n") {
			vec.Vec_StrPush(cover, c)
		}
	}
	vec.Vec_StrPush(cover, 0)
}
