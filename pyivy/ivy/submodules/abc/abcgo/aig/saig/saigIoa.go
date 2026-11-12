package saig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ObjName returns the source-compatible BLIF signal name for an AIG
// object. Unlike the C implementation, the returned string is independently
// owned and safe to retain.
func Saig_ObjName(p *aig.Graph, object aig.Edge) string {
	object = aig.Aig_Regular(object)
	nPis := int32(p.InputCount()) - aig.Aig_ManRegNum(p)
	nPos := int32(p.OutputCount()) - aig.Aig_ManRegNum(p)
	width := func(n int32) int {
		if n < 1 {
			n = 1
		}
		return len(strconv.FormatInt(int64(n-1), 10))
	}
	var prefix string
	var index int32
	switch {
	case aig.Aig_ObjIsConst1(p, object), aig.Aig_ObjIsNode(p, object):
		prefix, index = "n", aig.Aig_ObjId(object)
		return fmt.Sprintf("%s%0*d", prefix, width(aig.Aig_ManObjNumMax(p)), index)
	case aig.Aig_ObjIsCi(p, object):
		index = aig.Aig_ObjCioId(p, object)
		if index < nPis {
			return fmt.Sprintf("pi%0*d", width(nPis), index)
		}
		return fmt.Sprintf("lo%0*d", width(aig.Aig_ManRegNum(p)), index-nPis)
	case aig.Aig_ObjIsCo(p, object):
		index = aig.Aig_ObjCioId(p, object)
		if index < nPos {
			return fmt.Sprintf("po%0*d", width(nPos), index)
		}
		return fmt.Sprintf("li%0*d", width(aig.Aig_ManRegNum(p)), index-nPos)
	default:
		panic("Saig_ObjName: object is not part of the sequential AIG")
	}
}

// Saig_ManDumpBlif preserves the legacy C entry point. Errors are reported to
// stderr; new callers should use WriteBLIF to handle errors directly.
func Saig_ManDumpBlif(p *aig.Graph, fileName string) {
	if err := WriteBLIF(fileName, p); err != nil {
		fmt.Fprintf(os.Stderr, "Saig_ManDumpBlif(): %v\n", err)
	}
}

// WriteBLIF writes an AIG in the restricted BLIF dialect understood by
// ReadBLIF. It follows graph topology and keeps primary and latch terminals in
// ABC's original order.
func WriteBLIF(fileName string, p *aig.Graph) error {
	if p == nil {
		return errors.New("saig: cannot write a nil AIG")
	}
	nRegs := aig.Aig_ManRegNum(p)
	nPis := int32(p.InputCount()) - nRegs
	nPos := int32(p.OutputCount()) - nRegs
	if nPos < 0 || nPis < 0 || nRegs > int32(p.OutputCount()) {
		return errors.New("saig: inconsistent sequential terminal counts")
	}
	if p.OutputCount() == 0 {
		return errors.New("saig: AIG has no outputs")
	}
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	writeErr := writeBLIF(f, p, fileName, nPis, nPos, nRegs)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func writeBLIF(w io.Writer, p *aig.Graph, sourceName string, nPis, nPos, nRegs int32) error {
	write := func(format string, args ...any) error {
		_, err := fmt.Fprintf(w, format, args...)
		return err
	}
	model := strings.TrimSuffix(filepath.Base(sourceName), filepath.Ext(sourceName))
	if model == "" {
		model = "aig"
	}
	if err := write("# BLIF file written by procedure Saig_ManDumpBlif()\n# If unedited, this file can be read by Saig_ManReadBlif()\n# AIG stats: pi=%d po=%d reg=%d and=%d obj=%d maxid=%d\n.model %s\n", nPis, nPos, nRegs, aig.Aig_ManNodeNum(p), aig.Aig_ManObjNum(p), aig.Aig_ManObjNumMax(p), model); err != nil {
		return err
	}
	if err := write(".inputs"); err != nil {
		return err
	}
	inputs := p.Inputs()
	for i := int32(0); i < nPis; i++ {
		if err := write(" %s", Saig_ObjName(p, inputs[i])); err != nil {
			return err
		}
	}
	if err := write("\n.outputs"); err != nil {
		return err
	}
	outputs := p.Outputs()
	for i := int32(0); i < nPos; i++ {
		if err := write(" %s", Saig_ObjName(p, outputs[i])); err != nil {
			return err
		}
	}
	if err := write("\n"); err != nil {
		return err
	}
	for i := int32(0); i < nRegs; i++ {
		if err := write(".latch %s %s 0\n", Saig_ObjName(p, aig.Aig_ManLi(p, i)), Saig_ObjName(p, aig.Aig_ManLo(p, i))); err != nil {
			return err
		}
	}
	if aig.Aig_ObjRefs(p, p.Const1()) != 0 {
		if err := write(".names %s\n 1\n", Saig_ObjName(p, p.Const1())); err != nil {
			return err
		}
	}
	for id := int32(0); id < aig.Aig_ManObjNumMax(p); id++ {
		node := aig.Aig_ManObj(p, id)
		if aig.Aig_ObjIsNone(p, node) || !aig.Aig_ObjIsNode(p, node) {
			continue
		}
		row := fmt.Sprintf("%d%d", 1-aig.Aig_ObjFaninC0(p, node), 1-aig.Aig_ObjFaninC1(p, node))
		if err := write(".names %s %s %s\n%s 1\n", Saig_ObjName(p, aig.Aig_ObjFanin0(p, node)), Saig_ObjName(p, aig.Aig_ObjFanin1(p, node)), Saig_ObjName(p, node), row); err != nil {
			return err
		}
	}
	for _, output := range outputs {
		if !aig.Aig_ObjIsCo(p, output) {
			return errors.New("saig: graph output is not represented by an output node")
		}
		row := 1 - aig.Aig_ObjFaninC0(p, output)
		if err := write(".names %s %s\n%d 1\n", Saig_ObjName(p, aig.Aig_ObjFanin0(p, output)), Saig_ObjName(p, output), row); err != nil {
			return err
		}
	}
	return write(".end\n")
}

// Saig_ManReadBlif preserves the legacy C entry point and returns nil when a
// file is malformed. New callers should use ReadBLIF to receive the error.
func Saig_ManReadBlif(fileName string) *aig.Graph {
	p, err := ReadBLIF(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Saig_ManReadBlif(): %v\n", err)
		return nil
	}
	return p
}

type blifLatch struct{ input, output string }
type blifNames struct {
	inputs []string
	output string
	rows   []string
}

// ReadBLIF reads the sequential BLIF subset emitted by WriteBLIF. It rejects
// unsupported multi-row covers instead of silently approximating them.
func ReadBLIF(fileName string) (*aig.Graph, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readBLIF(f)
}

func readBLIF(r io.Reader) (*aig.Graph, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	var piNames, poNames []string
	var latches []blifLatch
	var blocks []blifNames
	var model string
	var current *blifNames
	seenEnd := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if current != nil && !strings.HasPrefix(fields[0], ".") {
			current.rows = append(current.rows, line)
			continue
		}
		current = nil
		switch fields[0] {
		case ".model":
			if len(fields) != 2 || model != "" {
				return nil, errors.New("saig: malformed or repeated .model directive")
			}
			model = fields[1]
		case ".inputs":
			piNames = append(piNames, fields[1:]...)
		case ".outputs":
			poNames = append(poNames, fields[1:]...)
		case ".latch":
			if len(fields) != 4 || fields[3] != "0" {
				return nil, errors.New("saig: only three-field zero-initialized .latch records are supported")
			}
			latches = append(latches, blifLatch{fields[1], fields[2]})
		case ".names":
			if len(fields) < 2 || len(fields) > 4 {
				return nil, fmt.Errorf("saig: unsupported .names arity on line %q", line)
			}
			blocks = append(blocks, blifNames{inputs: append([]string(nil), fields[1:len(fields)-1]...), output: fields[len(fields)-1]})
			current = &blocks[len(blocks)-1]
		case ".end":
			if len(fields) != 1 {
				return nil, errors.New("saig: malformed .end directive")
			}
			seenEnd = true
			current = nil
		default:
			return nil, fmt.Errorf("saig: unsupported directive %q", fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if !seenEnd || model == "" {
		return nil, errors.New("saig: missing .model or .end")
	}
	for i := range blocks {
		if len(blocks[i].rows) == 0 {
			return nil, fmt.Errorf("saig: .names %s has no truth row", blocks[i].output)
		}
	}

	p := aig.NewGraph()
	values := make(map[string]aig.Edge, len(piNames)+len(latches)+len(blocks))
	for _, name := range piNames {
		if name == "" {
			return nil, errors.New("saig: empty primary input name")
		}
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("saig: duplicate signal name %q", name)
		}
		values[name] = p.NewInput()
	}
	for _, latch := range latches {
		if latch.output == "" {
			return nil, errors.New("saig: empty latch output name")
		}
		if _, exists := values[latch.output]; exists {
			return nil, fmt.Errorf("saig: duplicate signal name %q", latch.output)
		}
		values[latch.output] = p.NewInput()
	}
	outputNames := make(map[string]struct{}, len(poNames)+len(latches))
	for _, name := range poNames {
		outputNames[name] = struct{}{}
	}
	for _, latch := range latches {
		outputNames[latch.input] = struct{}{}
	}
	oldNodeIDs := make(map[int32]int32)
	for _, block := range blocks {
		if block.output == "" {
			return nil, errors.New("saig: empty .names output name")
		}
		if _, exists := values[block.output]; exists {
			return nil, fmt.Errorf("saig: duplicate signal definition %q", block.output)
		}
		var value aig.Edge
		switch len(block.inputs) {
		case 0:
			if len(block.rows) != 1 || strings.TrimSpace(block.rows[0]) != "1" {
				return nil, fmt.Errorf("saig: unsupported constant table for %s", block.output)
			}
			value = p.Const1()
		case 1:
			if len(block.rows) != 1 {
				return nil, fmt.Errorf("saig: unsupported multi-row buffer table for %s", block.output)
			}
			parts := strings.Fields(block.rows[0])
			if len(parts) != 2 || parts[1] != "1" || (parts[0] != "0" && parts[0] != "1") {
				return nil, fmt.Errorf("saig: malformed buffer table for %s", block.output)
			}
			input, ok := values[block.inputs[0]]
			if !ok {
				return nil, fmt.Errorf("saig: undefined input %q", block.inputs[0])
			}
			value = input
			if parts[0] == "0" {
				value = value.Not()
			}
		case 2:
			if len(block.rows) != 1 {
				return nil, fmt.Errorf("saig: unsupported multi-row AND table for %s", block.output)
			}
			parts := strings.Fields(block.rows[0])
			if len(parts) != 2 || len(parts[0]) != 2 || parts[1] != "1" || (parts[0][0] != '0' && parts[0][0] != '1') || (parts[0][1] != '0' && parts[0][1] != '1') {
				return nil, fmt.Errorf("saig: malformed AND table for %s", block.output)
			}
			left, ok := values[block.inputs[0]]
			if !ok {
				return nil, fmt.Errorf("saig: undefined input %q", block.inputs[0])
			}
			right, ok := values[block.inputs[1]]
			if !ok {
				return nil, fmt.Errorf("saig: undefined input %q", block.inputs[1])
			}
			if parts[0][0] == '0' {
				left = left.Not()
			}
			if parts[0][1] == '0' {
				right = right.Not()
			}
			value = p.AND(left, right)
		}
		if _, isOutput := outputNames[block.output]; isOutput {
			values[block.output] = p.AddOutput(value)
		} else {
			values[block.output] = value
			if strings.HasPrefix(block.output, "n") {
				if oldID, err := strconv.ParseInt(block.output[1:], 10, 32); err == nil {
					oldNodeIDs[int32(oldID)] = aig.Aig_ObjId(value)
				}
			}
		}
	}
	if int32(p.OutputCount()) != int32(len(poNames)+len(latches)) {
		return nil, fmt.Errorf("saig: parsed %d outputs, expected %d primary and latch outputs", p.OutputCount(), len(poNames)+len(latches))
	}
	aig.Aig_ManSetRegNum(p, int32(len(latches)))
	p.Data = oldNodeIDs
	return p, nil
}
