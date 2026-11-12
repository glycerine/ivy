package io

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
)

// ParseBAF reads ABC's NUL-delimited header and little-endian 32-bit AIG
// records. All lengths and references are checked before constructing a graph.
func ParseBAF(filename string, data []byte) (*abc.Network, error) {
	offset := 0
	for offset < len(data) && data[offset] == '#' {
		end := bytes.IndexByte(data[offset:], '\n')
		if end < 0 {
			return nil, fmt.Errorf("%s: unterminated BAF comment", filename)
		}
		offset += end + 1
	}
	field := func() (string, error) {
		if offset >= len(data) {
			return "", fmt.Errorf("%s: truncated BAF header", filename)
		}
		end := bytes.IndexByte(data[offset:], 0)
		if end < 0 {
			return "", fmt.Errorf("%s: missing BAF field terminator", filename)
		}
		value := string(data[offset : offset+end])
		offset += end + 1
		return value, nil
	}
	name, err := field()
	if err != nil {
		return nil, err
	}
	counts := make([]int, 4)
	for i := range counts {
		s, err := field()
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseUint(s, 10, 31)
		if err != nil || value > uint64(len(data)) {
			return nil, fmt.Errorf("%s: invalid BAF count", filename)
		}
		counts[i] = int(value)
	}
	ni, no, nl, na := counts[0], counts[1], counts[2], counts[3]
	if uint64(2)*uint64(na)+uint64(no)+uint64(nl) > uint64(len(data))/4 {
		return nil, fmt.Errorf("%s: BAF counts exceed file size", filename)
	}
	model := &ioa.Model{Name: name, MaxVar: uint32(ni + nl + na), Symbols: map[ioa.Symbol]string{}}
	for i := 0; i < ni; i++ {
		name, err := field()
		if err != nil {
			return nil, err
		}
		model.Inputs = append(model.Inputs, uint32(2*(i+1)))
		model.Symbols[ioa.Symbol{Kind: 'i', Index: uint32(i)}] = name
	}
	for i := 0; i < no; i++ {
		name, err := field()
		if err != nil {
			return nil, err
		}
		model.Symbols[ioa.Symbol{Kind: 'o', Index: uint32(i)}] = name
	}
	latchNames := make([][3]string, nl)
	for i := range latchNames {
		for j := 0; j < 3; j++ {
			latchNames[i][j], err = field()
			if err != nil {
				return nil, err
			}
		}
		model.Symbols[ioa.Symbol{Kind: 'l', Index: uint32(i)}] = latchNames[i][0]
	}
	words := 2*na + no + nl
	if len(data)-offset != words*4 {
		return nil, fmt.Errorf("%s: BAF payload size mismatch", filename)
	}
	position := offset
	read := func() uint32 { value := binary.LittleEndian.Uint32(data[position:]); position += 4; return value }
	decode := func(lit uint32) uint32 {
		if lit < 2 {
			return lit ^ 1
		}
		return lit
	}
	for i := 0; i < na; i++ {
		left, right := read(), read()
		limit := uint32(ni + nl + i + 1)
		if left>>1 >= limit || right>>1 >= limit {
			return nil, fmt.Errorf("%s: forward or invalid BAF gate reference", filename)
		}
		model.ANDs = append(model.ANDs, ioa.And{Literal: 2 * limit, Left: decode(left), Right: decode(right)})
	}
	for i := 0; i < no; i++ {
		model.Outputs = append(model.Outputs, decode(read()))
	}
	initial := make([]abc.Abc_InitType_t, nl)
	for i := 0; i < nl; i++ {
		packed := read()
		lit := uint32(2 * (ni + i + 1))
		initial[i] = abc.Abc_InitType_t(packed & 3)
		init := lit
		if initial[i] == abc.ABC_INIT_ZERO {
			init = 0
		} else if initial[i] == abc.ABC_INIT_ONE {
			init = 1
		}
		model.Latches = append(model.Latches, ioa.Latch{Literal: lit, Next: decode(packed >> 2), Init: init})
	}
	network, err := NetworkFromAigerModel(model)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	network.Spec = filename
	for i, latch := range network.ObjectsOfKind(abc.ObjectLatch) {
		latch.Data = initial[i]
		latch.SetName(latchNames[i][0])
		latch.Fanins()[0].Object.SetName(latchNames[i][1])
		latch.Fanouts()[0].SetName(latchNames[i][2])
	}
	return network, nil
}
func ReadBAF(filename string, check bool) (*abc.Network, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseBAF(filename, data)
}
func Io_ReadBaf(filename string, check int32) *abc.Network {
	n, err := ReadBAF(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func init() { RegisterFormat(IO_FILE_BAF, ReadBAF, nil) }
