package gia

import (
	"bytes"
	"encoding/binary"
	"fmt"
	timpkg "github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"sort"
	"strings"
)

/**CFile****************************************************************

  FileName    [giaAigerExt.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Custom AIGER extensions.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAigerExt.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

func aigerReadUvar(pos *[]byte) (uint32, error) {
	v, n := binary.Uvarint(*pos)
	if n <= 0 || n > 5 || v > uint64(^uint32(0)) {
		return 0, fmt.Errorf("gia: malformed AIGER unsigned integer")
	}
	*pos = (*pos)[n:]
	return uint32(v), nil
}
func aigerAppendUvar(out []byte, v uint32) []byte { return binary.AppendUvarint(out, uint64(v)) }
func aigerSection(pos []byte) ([]byte, int, error) {
	if len(pos) < 4 {
		return nil, 0, fmt.Errorf("gia: truncated extension length")
	}
	size := int(binary.BigEndian.Uint32(pos))
	if size < 4 || size > len(pos) {
		return nil, 0, fmt.Errorf("gia: invalid internal extension length")
	}
	return pos[4:size], size, nil
}
func aigerFinishSection(out []byte) []byte {
	binary.BigEndian.PutUint32(out, uint32(len(out)))
	return out
}
func aigerVecBytes(data []byte) *Vec_Str_t {
	return vec.Vec_StrAllocArray(append([]byte(nil), data...), len(data))
}
func aigerVecInts(data []int32) *Vec_Int_t {
	return vec.Vec_IntAllocArray(append([]int32(nil), data...), len(data))
}

func Gia_AigerReadEquivClasses(pos *[]byte, nSize int32) ([]Gia_Rpr_t, error) {
	if nSize < 1 {
		return nil, fmt.Errorf("gia: invalid equivalence object count")
	}
	data, size, err := aigerSection(*pos)
	if err != nil {
		return nil, err
	}
	result := make([]Gia_Rpr_t, nSize)
	for i := range result {
		result[i].iRepr = GIA_VOID
	}
	repr, node := uint64(0), uint64(0)
	for len(data) > 0 {
		item, err := aigerReadUvar(&data)
		if err != nil {
			return nil, err
		}
		if item&1 != 0 {
			repr += uint64(item >> 1)
			if repr >= uint64(nSize) {
				return nil, fmt.Errorf("gia: equivalence representative out of range")
			}
			node = repr
			continue
		}
		delta := item >> 2
		node += uint64(delta)
		if delta == 0 || node >= uint64(nSize) || repr >= node || result[node].iRepr != GIA_VOID {
			return nil, fmt.Errorf("gia: invalid equivalence member")
		}
		result[node].iRepr = uint32(repr)
		result[node].fProved = uint8((item >> 1) & 1)
	}
	*pos = (*pos)[size:]
	return result, nil
}
func Gia_WriteEquivClassesInt(p *Gia_Man_t, size *int32) ([]byte, error) {
	if len(p.PReprs) != int(p.nObjs) {
		return nil, fmt.Errorf("gia: equivalence array length mismatch")
	}
	groups := map[int][]int{}
	for node := 1; node < int(p.nObjs); node++ {
		repr := p.PReprs[node].iRepr
		if repr == GIA_VOID {
			continue
		}
		if repr >= uint32(node) || p.PReprs[repr].iRepr != GIA_VOID {
			return nil, fmt.Errorf("gia: invalid equivalence representative")
		}
		groups[int(repr)] = append(groups[int(repr)], node)
	}
	out := make([]byte, 4)
	out = aigerAppendUvar(out, 1)
	members := groups[0]
	prev := 0
	for _, node := range members {
		item := uint32(node-prev)<<2 | uint32(p.PReprs[node].fProved&1)<<1
		out = aigerAppendUvar(out, item)
		prev = node
	}
	var roots []int
	for repr := range groups {
		if repr != 0 {
			roots = append(roots, repr)
		}
	}
	sort.Ints(roots)
	previousRoot := 0
	for _, repr := range roots {
		out = aigerAppendUvar(out, uint32(repr-previousRoot)<<1|1)
		previousRoot = repr
		prev = repr
		for _, node := range groups[repr] {
			out = aigerAppendUvar(out, uint32(node-prev)<<2|uint32(p.PReprs[node].fProved&1)<<1)
			prev = node
		}
	}
	out = aigerFinishSection(out)
	if size != nil {
		*size = int32(len(out))
	}
	return out, nil
}
func Gia_WriteEquivClasses(p *Gia_Man_t) (*Vec_Str_t, error) {
	data, err := Gia_WriteEquivClassesInt(p, nil)
	if err != nil {
		return nil, err
	}
	return aigerVecBytes(data), nil
}

func Gia_AigerReadDiffValue(pos *[]byte, previous int32) (int32, error) {
	item, err := aigerReadUvar(pos)
	if err != nil {
		return 0, err
	}
	value := int64(previous)
	if item&1 != 0 {
		value += int64(item >> 1)
	} else {
		value -= int64(item >> 1)
	}
	if value < 0 || value > 0x7fffffff {
		return 0, fmt.Errorf("gia: differential object ID overflow")
	}
	return int32(value), nil
}
func aigerAppendDiff(out []byte, previous, value int32) ([]byte, error) {
	if previous < 0 || value < 0 {
		return nil, fmt.Errorf("gia: negative differential object ID")
	}
	delta := int64(value) - int64(previous)
	item := uint32(0)
	if delta > 0 {
		item = uint32(delta)<<1 | 1
	} else {
		item = uint32(-delta) << 1
	}
	return aigerAppendUvar(out, item), nil
}
func Gia_AigerWriteDiffValue(buffer []byte, position, previous, value int32) (int32, error) {
	data, err := aigerAppendDiff(nil, previous, value)
	if err != nil {
		return 0, err
	}
	if position < 0 || int(position) > len(buffer) || len(buffer)-int(position) < len(data) {
		return 0, fmt.Errorf("gia: short differential output buffer")
	}
	copy(buffer[position:], data)
	return position + int32(len(data)), nil
}

type aigerLUT struct {
	root int32
	fans []int32
}

func aigerLUTs(mapping []int32, nObjs int32) ([]aigerLUT, error) {
	if nObjs < 1 || len(mapping) < int(nObjs) {
		return nil, fmt.Errorf("gia: short LUT mapping")
	}
	var luts []aigerLUT
	for root := int32(0); root < nObjs; root++ {
		offset := mapping[root]
		if offset == 0 {
			continue
		}
		if offset < nObjs || int64(offset) >= int64(len(mapping)) {
			return nil, fmt.Errorf("gia: invalid LUT mapping offset")
		}
		n := mapping[offset]
		if n < 0 || int64(offset)+int64(n)+1 >= int64(len(mapping)) || mapping[offset+n+1] != root {
			return nil, fmt.Errorf("gia: invalid LUT mapping record")
		}
		fans := mapping[offset+1 : offset+n+1]
		for _, fan := range fans {
			if fan < 0 || fan >= nObjs || fan == root {
				return nil, fmt.Errorf("gia: invalid LUT fanin")
			}
		}
		luts = append(luts, aigerLUT{root, append([]int32(nil), fans...)})
	}
	return luts, nil
}
func aigerMakeMapping(nObjs int32, luts []aigerLUT) ([]int32, error) {
	if nObjs < 1 {
		return nil, fmt.Errorf("gia: invalid mapping object count")
	}
	result := make([]int32, nObjs)
	for _, lut := range luts {
		if lut.root < 0 || lut.root >= nObjs || result[lut.root] != 0 {
			return nil, fmt.Errorf("gia: duplicate or invalid LUT root")
		}
		for _, f := range lut.fans {
			if f < 0 || f >= nObjs || f == lut.root {
				return nil, fmt.Errorf("gia: invalid LUT fanin")
			}
		}
		if len(result) > 0x7fffffff-len(lut.fans)-2 {
			return nil, fmt.Errorf("gia: mapping exceeds integer range")
		}
		result[lut.root] = int32(len(result))
		result = append(result, int32(len(lut.fans)))
		result = append(result, lut.fans...)
		result = append(result, lut.root)
	}
	return result, nil
}
func Gia_AigerReadMapping(pos *[]byte, nObjs int32) ([]int32, error) {
	data, size, err := aigerSection(*pos)
	if err != nil {
		return nil, err
	}
	var luts []aigerLUT
	previous := int32(0)
	for len(data) > 0 {
		n, err := aigerReadUvar(&data)
		if err != nil {
			return nil, err
		}
		if uint64(n)+1 > uint64(len(data)) {
			return nil, fmt.Errorf("gia: truncated LUT fanins")
		}
		fans := make([]int32, int(n))
		for i := range fans {
			fans[i], err = Gia_AigerReadDiffValue(&data, previous)
			if err != nil {
				return nil, err
			}
			previous = fans[i]
		}
		root, err := Gia_AigerReadDiffValue(&data, previous)
		if err != nil {
			return nil, err
		}
		previous = root
		luts = append(luts, aigerLUT{root, fans})
	}
	result, err := aigerMakeMapping(nObjs, luts)
	if err != nil {
		return nil, err
	}
	*pos = (*pos)[size:]
	return result, nil
}
func Gia_AigerWriteMappingInt(p *Gia_Man_t, size *int32) ([]byte, error) {
	if p.VMapping == nil {
		return nil, fmt.Errorf("gia: mapping unavailable")
	}
	luts, err := aigerLUTs(vec.Vec_IntArray(p.VMapping), p.nObjs)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 4)
	previous := int32(0)
	for _, lut := range luts {
		out = aigerAppendUvar(out, uint32(len(lut.fans)))
		for _, fan := range lut.fans {
			out, err = aigerAppendDiff(out, previous, fan)
			if err != nil {
				return nil, err
			}
			previous = fan
		}
		out, err = aigerAppendDiff(out, previous, lut.root)
		if err != nil {
			return nil, err
		}
		previous = lut.root
	}
	out = aigerFinishSection(out)
	if size != nil {
		*size = int32(len(out))
	}
	return out, nil
}
func Gia_AigerWriteMapping(p *Gia_Man_t) (*Vec_Str_t, error) {
	data, err := Gia_AigerWriteMappingInt(p, nil)
	if err != nil {
		return nil, err
	}
	return aigerVecBytes(data), nil
}
func Gia_AigerReadMappingSimple(pos *[]byte, nSize int32) ([]int32, error) {
	if nSize < 0 || nSize%4 != 0 || int(nSize) > len(*pos) {
		return nil, fmt.Errorf("gia: invalid simple mapping size")
	}
	result := make([]int32, nSize/4)
	for i := range result {
		result[i] = int32(binary.LittleEndian.Uint32((*pos)[4*i:]))
	}
	*pos = (*pos)[nSize:]
	return result, nil
}
func Gia_AigerWriteMappingSimple(p *Gia_Man_t) (*Vec_Str_t, error) {
	if p.VMapping == nil {
		return nil, fmt.Errorf("gia: mapping unavailable")
	}
	mapping := vec.Vec_IntArray(p.VMapping)
	if _, err := aigerLUTs(mapping, p.nObjs); err != nil {
		return nil, err
	}
	out := make([]byte, 4*len(mapping))
	for i, value := range mapping {
		binary.LittleEndian.PutUint32(out[4*i:], uint32(value))
	}
	return aigerVecBytes(out), nil
}
func Gia_AigerReadMappingDoc(pos *[]byte, nObjs int32) (*Vec_Int_t, error) {
	data := *pos
	if len(data) < 8 {
		return nil, fmt.Errorf("gia: truncated mapping header")
	}
	count, maxSize := binary.BigEndian.Uint32(data), binary.BigEndian.Uint32(data[4:])
	data = data[8:]
	if uint64(count) > uint64(len(data)/8) {
		return nil, fmt.Errorf("gia: impossible LUT count")
	}
	luts := make([]aigerLUT, 0, int(count))
	for i := uint32(0); i < count; i++ {
		if len(data) < 8 {
			return nil, fmt.Errorf("gia: truncated LUT record")
		}
		root, n := int32(binary.BigEndian.Uint32(data)), binary.BigEndian.Uint32(data[4:])
		data = data[8:]
		if n > maxSize || uint64(n) > uint64(len(data)/4) {
			return nil, fmt.Errorf("gia: invalid LUT size")
		}
		fans := make([]int32, int(n))
		for j := range fans {
			fans[j] = int32(binary.BigEndian.Uint32(data))
			data = data[4:]
		}
		luts = append(luts, aigerLUT{root, fans})
	}
	mapping, err := aigerMakeMapping(nObjs, luts)
	if err != nil {
		return nil, err
	}
	*pos = data
	return aigerVecInts(mapping), nil
}
func Gia_AigerWriteMappingDoc(p *Gia_Man_t) (*Vec_Str_t, error) {
	if p.VMapping == nil {
		return nil, fmt.Errorf("gia: mapping unavailable")
	}
	luts, err := aigerLUTs(vec.Vec_IntArray(p.VMapping), p.nObjs)
	if err != nil {
		return nil, err
	}
	var max int
	for _, lut := range luts {
		if len(lut.fans) > max {
			max = len(lut.fans)
		}
	}
	out := binary.BigEndian.AppendUint32(nil, uint32(len(luts)))
	out = binary.BigEndian.AppendUint32(out, uint32(max))
	for _, lut := range luts {
		out = binary.BigEndian.AppendUint32(out, uint32(lut.root))
		out = binary.BigEndian.AppendUint32(out, uint32(len(lut.fans)))
		for _, fan := range lut.fans {
			out = binary.BigEndian.AppendUint32(out, uint32(fan))
		}
	}
	return aigerVecBytes(out), nil
}
func Gia_AigerReadPacking(pos *[]byte, nSize int32) (*Vec_Int_t, error) {
	if nSize < 0 || nSize%4 != 0 || int(nSize) > len(*pos) {
		return nil, fmt.Errorf("gia: invalid packing size")
	}
	result := make([]int32, nSize/4)
	for i := range result {
		result[i] = int32(binary.BigEndian.Uint32((*pos)[4*i:]))
	}
	*pos = (*pos)[nSize:]
	return aigerVecInts(result), nil
}
func Gia_WritePacking(packing *Vec_Int_t) *Vec_Str_t {
	var out []byte
	if packing != nil {
		for _, value := range vec.Vec_IntArray(packing) {
			out = binary.BigEndian.AppendUint32(out, uint32(value))
		}
	}
	return aigerVecBytes(out)
}

func aigerBEInts(data []byte) ([]int32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("misaligned integer payload")
	}
	values := make([]int32, len(data)/4)
	for i := range values {
		values[i] = int32(binary.BigEndian.Uint32(data[4*i:]))
	}
	return values, nil
}
func aigerLEInts(data []byte) ([]int32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("misaligned native integer payload")
	}
	values := make([]int32, len(data)/4)
	for i := range values {
		values[i] = int32(binary.LittleEndian.Uint32(data[4*i:]))
	}
	return values, nil
}
func aigerBEArray(values []int32) []byte {
	var out []byte
	for _, v := range values {
		out = binary.BigEndian.AppendUint32(out, uint32(v))
	}
	return out
}
func aigerLEArray(values []int32) []byte {
	out := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(out[i*4:], uint32(v))
	}
	return out
}
func aigerDecodeExtension(p *Gia_Man_t, tag byte, payload []byte, opts AigerReadOptions, depth int) error {
	data := append([]byte(nil), payload...)
	switch tag {
	case 'a':
		extra, err := parseAiger(data, opts, depth+1)
		if err != nil {
			return err
		}
		p.PAigExtra = extra
	case 'c', 'd':
		if len(data) != 4 {
			return fmt.Errorf("invalid scalar extension size")
		}
		value := int32(binary.BigEndian.Uint32(data))
		if value < 0 {
			return fmt.Errorf("negative scalar extension")
		}
		if tag == 'c' {
			if int64(value)+int64(p.aigerBadOutputs) > int64(Gia_ManPoNum(p)) {
				return fmt.Errorf("constraint count exceeds outputs")
			}
			p.nConstrs = value
		} else {
			p.nAnd2Delay = value
		}
	case 'e':
		reprs, err := Gia_AigerReadEquivClasses(&data, p.nObjs)
		if err != nil {
			return err
		}
		if len(data) != 0 {
			return fmt.Errorf("trailing equivalence bytes")
		}
		p.PReprs = reprs
		p.PNexts = make([]int32, p.nObjs)
		last := map[uint32]int32{}
		for node, r := range reprs {
			if r.iRepr == GIA_VOID {
				continue
			}
			previous, ok := last[r.iRepr]
			if !ok {
				previous = int32(r.iRepr)
			}
			p.PNexts[previous] = int32(node)
			last[r.iRepr] = int32(node)
		}
	case 'm':
		mapping, err := Gia_AigerReadMappingDoc(&data, p.nObjs)
		if err != nil {
			return err
		}
		if len(data) != 0 {
			return fmt.Errorf("trailing mapping bytes")
		}
		p.VMapping = mapping
	case 'k':
		packing, err := Gia_AigerReadPacking(&data, int32(len(data)))
		if err != nil {
			return err
		}
		if err := aigerValidatePacking(vec.Vec_IntArray(packing), p.nObjs); err != nil {
			return err
		}
		p.VPacking = packing
	case 'n':
		end := bytes.IndexByte(data, 0)
		if end < 0 || end != len(data)-1 {
			return fmt.Errorf("invalid model name")
		}
		p.pName = string(data[:end])
	case 'r', 's':
		values, err := aigerBEInts(data)
		if err != nil || len(values) == 0 || values[0] < 0 || int(values[0]) != len(values)-1 {
			return fmt.Errorf("invalid counted register array")
		}
		if tag == 'r' {
			p.vRegClasses = aigerVecInts(values[1:])
		} else {
			p.vRegInits = aigerVecInts(values[1:])
		}
	case 'f', 'g', 'v':
		values, err := aigerLEInts(data)
		if err != nil {
			return err
		}
		if tag == 'f' {
			if len(values) != int(p.nRegs) {
				return fmt.Errorf("flop-class count mismatch")
			}
			p.VFlopClasses = aigerVecInts(values)
		} else {
			if len(values) != int(p.nObjs) {
				return fmt.Errorf("object-class count mismatch")
			}
			if tag == 'g' {
				p.VGateClasses = aigerVecInts(values)
			} else {
				p.VObjClasses = aigerVecInts(values)
			}
		}
	case 'i', 'o':
		values, err := aigerLEInts(data)
		if err != nil {
			return err
		}
		times := make([]float32, len(values))
		for i, value := range values {
			times[i] = math.Float32frombits(uint32(value))
		}
		if tag == 'i' {
			p.VInArrs = vec.Vec_FltAllocArray(times, len(times))
		} else {
			p.VOutReqs = vec.Vec_FltAllocArray(times, len(times))
		}
	case 'u':
		if len(data) != int(p.nObjs) {
			return fmt.Errorf("switching array size mismatch")
		}
		p.pSwitching = append([]byte(nil), data...)
	case 'p':
		if len(data) != int(p.nObjs)*4 {
			return fmt.Errorf("placement array size mismatch")
		}
		placement := make([]Gia_Plc_t, p.nObjs)
		for i := range placement {
			word := binary.LittleEndian.Uint32(data[i*4:])
			placement[i] = Gia_Plc_t{fFixed: word & 1, xCoord: word >> 1 & 0x7fff, fUndef: word >> 16 & 1, yCoord: word >> 17 & 0x7fff}
		}
		p.pPlacement = placement
	case 'b':
		end := bytes.IndexByte(data, 0)
		if end < 0 {
			return fmt.Errorf("missing configuration cell string")
		}
		values, err := aigerBEInts(data[end+1:])
		if err != nil {
			return err
		}
		p.PCellStr = string(data[:end])
		p.VConfigs = aigerVecInts(values)
	case 'q':
		values, err := aigerBEInts(data)
		if err != nil || len(values) == 0 || values[0] < 0 || int64(values[0])*2 != int64(len(values)-1) {
			return fmt.Errorf("invalid choice pair count")
		}
		siblings := make([]int32, p.nObjs)
		for i := 1; i < len(values); i += 2 {
			head, node := values[i], values[i+1]
			if head <= node || head >= p.nObjs || node <= 0 || siblings[head] != 0 {
				return fmt.Errorf("invalid choice pair")
			}
			siblings[head] = node
		}
		p.pSibls = siblings
	case 'w':
		values, err := aigerBEInts(data)
		if err != nil || len(values) == 0 || values[0] < 0 || int64(values[0])*2 != int64(len(values)-1) {
			return fmt.Errorf("invalid edge pair count")
		}
		first, second := make([]int32, p.nObjs), make([]int32, p.nObjs)
		for i := 1; i < len(values); i += 2 {
			a, b := values[i], values[i+1]
			if a <= 0 || a >= b || b >= p.nObjs {
				return fmt.Errorf("invalid edge pair")
			}
			for _, pair := range [][2]int32{{a, b}, {b, a}} {
				from, to := pair[0], pair[1]
				if first[from] == to || second[from] == to {
					return fmt.Errorf("duplicate edge pair")
				}
				if first[from] == 0 {
					first[from] = to
				} else if second[from] == 0 {
					second[from] = to
				} else {
					return fmt.Errorf("edge degree exceeds two")
				}
			}
		}
		p.VEdge1 = aigerVecInts(first)
		p.VEdge2 = aigerVecInts(second)
	case 'h', 't':
		if tag == 'h' {
			if manager, err := timpkg.UnmarshalAigerTiming(data); err == nil {
				p.PManTime = manager
			} else {
				timing, _ := p.PManTime.(*AigerTimingData)
				if timing == nil {
					timing = new(AigerTimingData)
				}
				timing.Hierarchy = append([]byte(nil), data...)
				p.PManTime = timing
			}
		} else {
			if _, isTimManager := p.PManTime.(*timpkg.Tim_Man_t); !isTimManager {
				timing, _ := p.PManTime.(*AigerTimingData)
				if timing == nil {
					timing = new(AigerTimingData)
				}
				timing.Timing = append([]byte(nil), data...)
				p.PManTime = timing
			}
		}
	}
	if p.aigerExtensions == nil {
		p.aigerExtensions = map[byte][]byte{}
	}
	p.aigerExtensions[tag] = append([]byte(nil), payload...)
	return nil
}
func aigerValidatePacking(values []int32, nObjs int32) error {
	if len(values) == 0 {
		return fmt.Errorf("empty packing data")
	}
	if values[0] < 0 || int64(values[0]) > int64(len(values)-1)/2 {
		return fmt.Errorf("invalid packing block count")
	}
	cursor := 1
	for block := int32(0); block < values[0]; block++ {
		if cursor >= len(values) {
			return fmt.Errorf("truncated packing block")
		}
		n := values[cursor]
		cursor++
		if n < 1 || n > 3 || int(n) > len(values)-cursor {
			return fmt.Errorf("invalid packing block size")
		}
		for _, id := range values[cursor : cursor+int(n)] {
			if id <= 0 || id >= nObjs {
				return fmt.Errorf("invalid packing member")
			}
		}
		cursor += int(n)
	}
	if cursor != len(values) {
		return fmt.Errorf("trailing packing data")
	}
	return nil
}

func aigerKnownTag(tag byte) bool {
	switch tag {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'k', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w':
		return true
	}
	return false
}
func aigerWireMetadata(p *Gia_Man_t, wire *aigerWireModel) (*Gia_Man_t, error) {
	identity := len(wire.ids) == wire.objects
	for old, id := range wire.ids {
		if id < 0 {
			return nil, fmt.Errorf("gia: unlisted interface/object in manager")
		}
		if int(id) != old {
			identity = false
		}
	}
	if identity {
		return p, nil
	}
	for tag := range p.aigerExtensions {
		if !aigerKnownTag(tag) {
			return nil, fmt.Errorf("gia: cannot renumber opaque extension %q", tag)
		}
	}
	// This temporary view is used only by metadata codecs, never as a graph.
	meta := *p
	meta.nObjs = int32(wire.objects)
	remap := func(id int32) (int32, error) {
		if id < 0 || int(id) >= len(wire.ids) || wire.ids[id] < 0 {
			return 0, fmt.Errorf("gia: metadata object ID out of range")
		}
		return wire.ids[id], nil
	}
	remapVector := func(v *Vec_Int_t) (*Vec_Int_t, error) {
		if v == nil {
			return nil, nil
		}
		old := vec.Vec_IntArray(v)
		if len(old) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: object annotation length mismatch")
		}
		values := make([]int32, wire.objects)
		for id, value := range old {
			values[wire.ids[id]] = value
		}
		return aigerVecInts(values), nil
	}
	var err error
	meta.VGateClasses, err = remapVector(p.VGateClasses)
	if err != nil {
		return nil, err
	}
	meta.VObjClasses, err = remapVector(p.VObjClasses)
	if err != nil {
		return nil, err
	}
	if p.pSwitching != nil {
		if len(p.pSwitching) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: switching length mismatch")
		}
		meta.pSwitching = make([]byte, wire.objects)
		for id, value := range p.pSwitching {
			meta.pSwitching[wire.ids[id]] = value
		}
	}
	if p.pPlacement != nil {
		if len(p.pPlacement) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: placement length mismatch")
		}
		meta.pPlacement = make([]Gia_Plc_t, wire.objects)
		for id, value := range p.pPlacement {
			meta.pPlacement[wire.ids[id]] = value
		}
	}
	if p.VMapping != nil {
		luts, err := aigerLUTs(vec.Vec_IntArray(p.VMapping), p.nObjs)
		if err != nil {
			return nil, err
		}
		for i := range luts {
			luts[i].root, err = remap(luts[i].root)
			if err != nil {
				return nil, err
			}
			for j := range luts[i].fans {
				luts[i].fans[j], err = remap(luts[i].fans[j])
				if err != nil {
					return nil, err
				}
			}
		}
		values, err := aigerMakeMapping(meta.nObjs, luts)
		if err != nil {
			return nil, err
		}
		meta.VMapping = aigerVecInts(values)
	}
	if p.VPacking != nil {
		values := append([]int32(nil), vec.Vec_IntArray(p.VPacking)...)
		if err := aigerValidatePacking(values, p.nObjs); err != nil {
			return nil, err
		}
		for i := 1; i < len(values); {
			n := int(values[i])
			i++
			for j := 0; j < n; j++ {
				values[i+j], err = remap(values[i+j])
				if err != nil {
					return nil, err
				}
			}
			i += n
		}
		meta.VPacking = aigerVecInts(values)
	}
	if p.PReprs != nil {
		if len(p.PReprs) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: representative array length mismatch")
		}
		groups := map[uint32][]int32{}
		for node, r := range p.PReprs {
			if r.iRepr == GIA_VOID {
				continue
			}
			if r.iRepr >= uint32(node) || p.PReprs[r.iRepr].iRepr != GIA_VOID {
				return nil, fmt.Errorf("gia: invalid representative class")
			}
			groups[r.iRepr] = append(groups[r.iRepr], int32(node))
		}
		meta.PReprs = make([]Gia_Rpr_t, wire.objects)
		meta.PNexts = make([]int32, wire.objects)
		for i := range meta.PReprs {
			meta.PReprs[i].iRepr = GIA_VOID
		}
		for head, nodes := range groups {
			nodes = append(nodes, int32(head))
			mapped := make([]int32, len(nodes))
			oldAt := map[int32]int32{}
			for i, node := range nodes {
				mapped[i] = wire.ids[node]
				oldAt[mapped[i]] = node
			}
			sort.Slice(mapped, func(i, j int) bool { return mapped[i] < mapped[j] })
			for i, node := range mapped[1:] {
				meta.PReprs[node].iRepr = uint32(mapped[0])
				// Reparenting to an unproved member cannot certify relations to
				// the new head. The old head itself is trivially proved.
				headOld := oldAt[mapped[0]]
				headProved := headOld == int32(head) || p.PReprs[headOld].fProved != 0
				nodeOld := oldAt[node]
				if headProved && (nodeOld == int32(head) || p.PReprs[nodeOld].fProved != 0) {
					meta.PReprs[node].fProved = 1
				}
				meta.PNexts[mapped[i]] = node
			}
		}
	}
	if p.pSibls != nil {
		if len(p.pSibls) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: sibling array length mismatch")
		}
		parent := make([]int32, p.nObjs)
		for i := range parent {
			parent[i] = int32(i)
		}
		var find func(int32) int32
		find = func(id int32) int32 {
			for parent[id] != id {
				parent[id] = parent[parent[id]]
				id = parent[id]
			}
			return id
		}
		used := map[int32]bool{}
		for head, node := range p.pSibls {
			if node == 0 {
				continue
			}
			if node < 0 || node >= int32(head) {
				return nil, fmt.Errorf("gia: invalid sibling link")
			}
			a, b := find(int32(head)), find(node)
			parent[a] = b
			used[int32(head)] = true
			used[node] = true
		}
		groups := map[int32][]int32{}
		for node := range used {
			root := find(node)
			groups[root] = append(groups[root], wire.ids[node])
		}
		meta.pSibls = make([]int32, wire.objects)
		for _, group := range groups {
			sort.Slice(group, func(i, j int) bool { return group[i] > group[j] })
			for i := 0; i+1 < len(group); i++ {
				meta.pSibls[group[i]] = group[i+1]
			}
		}
	}
	if p.VEdge1 != nil || p.VEdge2 != nil {
		if p.VEdge1 == nil || p.VEdge2 == nil {
			return nil, fmt.Errorf("gia: incomplete edge arrays")
		}
		a, b := vec.Vec_IntArray(p.VEdge1), vec.Vec_IntArray(p.VEdge2)
		if len(a) != int(p.nObjs) || len(b) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: edge array length mismatch")
		}
		x, y := make([]int32, wire.objects), make([]int32, wire.objects)
		for id := range a {
			if a[id] != 0 {
				x[wire.ids[id]], err = remap(a[id])
				if err != nil {
					return nil, err
				}
			}
			if b[id] != 0 {
				y[wire.ids[id]], err = remap(b[id])
				if err != nil {
					return nil, err
				}
			}
		}
		meta.VEdge1 = aigerVecInts(x)
		meta.VEdge2 = aigerVecInts(y)
	}
	return &meta, nil
}

type aigerExtension struct {
	tag  byte
	data []byte
}

func aigerEncodeExtensions(p *Gia_Man_t, wire *aigerWireModel, opts AigerWriteOptions, depth int) ([]aigerExtension, error) {
	data := map[byte][]byte{}
	for tag, payload := range p.aigerExtensions {
		if !aigerKnownTag(tag) {
			data[tag] = append([]byte(nil), payload...)
		}
	}
	if p.PAigExtra != nil {
		encoded, err := aigerMarshalManager(p.PAigExtra, AigerWriteOptions{}, depth+1)
		if err != nil {
			return nil, err
		}
		data['a'] = encoded
	}
	if p.nConstrs != 0 {
		data['c'] = binary.BigEndian.AppendUint32(nil, uint32(p.nConstrs))
	}
	if p.nAnd2Delay != 0 {
		if p.nAnd2Delay < 0 {
			return nil, fmt.Errorf("gia: negative AND delay")
		}
		data['d'] = binary.BigEndian.AppendUint32(nil, uint32(p.nAnd2Delay))
	}
	for _, part := range []struct {
		tag byte
		v   *Vec_Flt_t
	}{{'i', p.VInArrs}, {'o', p.VOutReqs}} {
		if part.v == nil {
			continue
		}
		var encoded []byte
		for _, value := range vec.Vec_FltArray(part.v) {
			encoded = binary.LittleEndian.AppendUint32(encoded, math.Float32bits(value))
		}
		data[part.tag] = encoded
	}
	if p.PReprs != nil {
		encoded, err := Gia_WriteEquivClassesInt(p, nil)
		if err != nil {
			return nil, err
		}
		data['e'] = encoded
	}
	for _, part := range []struct {
		tag   byte
		v     *Vec_Int_t
		count int
	}{{'f', p.VFlopClasses, int(p.nRegs)}, {'g', p.VGateClasses, int(p.nObjs)}, {'v', p.VObjClasses, int(p.nObjs)}} {
		if part.v == nil {
			continue
		}
		values := vec.Vec_IntArray(part.v)
		if len(values) != part.count {
			return nil, fmt.Errorf("gia: class-array count mismatch")
		}
		data[part.tag] = aigerLEArray(values)
	}
	if p.PManTime != nil {
		switch timing := p.PManTime.(type) {
		case *AigerTimingData:
			if timing.Hierarchy != nil {
				data['h'] = append([]byte(nil), timing.Hierarchy...)
			}
			if timing.Timing != nil {
				data['t'] = append([]byte(nil), timing.Timing...)
			}
		case interface{ MarshalAigerTiming(bool) ([]byte, error) }:
			encoded, err := timing.MarshalAigerTiming(true)
			if err != nil {
				return nil, err
			}
			data['h'] = encoded
			if payload, ok := p.aigerExtensions['t']; ok {
				data['t'] = append([]byte(nil), payload...)
			}
		default:
			return nil, fmt.Errorf("gia: timing manager has no AIGER serializer")
		}
	}
	if p.VPacking != nil {
		if err := aigerValidatePacking(vec.Vec_IntArray(p.VPacking), p.nObjs); err != nil {
			return nil, err
		}
		data['k'] = append([]byte(nil), vec.Vec_StrArray(Gia_WritePacking(p.VPacking))...)
	}
	if p.VMapping != nil {
		encoded, err := Gia_AigerWriteMappingDoc(p)
		if err != nil {
			return nil, err
		}
		data['m'] = append([]byte(nil), vec.Vec_StrArray(encoded)...)
	}
	if p.pPlacement != nil {
		if len(p.pPlacement) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: placement length mismatch")
		}
		var encoded []byte
		for _, value := range p.pPlacement {
			if value.fFixed > 1 || value.fUndef > 1 || value.xCoord > 0x7fff || value.yCoord > 0x7fff {
				return nil, fmt.Errorf("gia: placement field overflow")
			}
			word := value.fFixed | value.xCoord<<1 | value.fUndef<<16 | value.yCoord<<17
			encoded = binary.LittleEndian.AppendUint32(encoded, word)
		}
		data['p'] = encoded
	}
	for _, part := range []struct {
		tag byte
		v   *Vec_Int_t
	}{{'r', p.vRegClasses}, {'s', p.vRegInits}} {
		if part.v == nil {
			continue
		}
		values := vec.Vec_IntArray(part.v)
		data[part.tag] = aigerBEArray(append([]int32{int32(len(values))}, values...))
	}
	if p.VConfigs != nil {
		if strings.IndexByte(p.PCellStr, 0) >= 0 {
			return nil, fmt.Errorf("gia: invalid configuration cell string")
		}
		encoded := append([]byte(p.PCellStr), 0)
		encoded = append(encoded, aigerBEArray(vec.Vec_IntArray(p.VConfigs))...)
		data['b'] = encoded
	}
	if p.pSibls != nil {
		if len(p.pSibls) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: sibling array length mismatch")
		}
		values := []int32{0}
		for head, node := range p.pSibls {
			if node == 0 {
				continue
			}
			if node < 0 || node >= int32(head) {
				return nil, fmt.Errorf("gia: invalid sibling link")
			}
			values = append(values, int32(head), node)
			values[0]++
		}
		data['q'] = aigerBEArray(values)
	}
	if p.pSwitching != nil {
		if len(p.pSwitching) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: switching length mismatch")
		}
		data['u'] = append([]byte(nil), p.pSwitching...)
	}
	if p.VEdge1 != nil || p.VEdge2 != nil {
		if p.VEdge1 == nil || p.VEdge2 == nil {
			return nil, fmt.Errorf("gia: incomplete edge arrays")
		}
		a, b := vec.Vec_IntArray(p.VEdge1), vec.Vec_IntArray(p.VEdge2)
		if len(a) != int(p.nObjs) || len(b) != int(p.nObjs) {
			return nil, fmt.Errorf("gia: edge array length mismatch")
		}
		values := []int32{0}
		for id := range a {
			if a[id] != 0 && a[id] == b[id] {
				return nil, fmt.Errorf("gia: duplicate edge link")
			}
			for _, neighbor := range []int32{a[id], b[id]} {
				if neighbor == 0 {
					continue
				}
				if neighbor < 0 || neighbor >= p.nObjs || neighbor == int32(id) || a[neighbor] != int32(id) && b[neighbor] != int32(id) {
					return nil, fmt.Errorf("gia: invalid undirected edge")
				}
				if neighbor < int32(id) {
					values = append(values, neighbor, int32(id))
					values[0]++
				}
			}
		}
		data['w'] = aigerBEArray(values)
	}
	if p.pName != "" {
		if strings.IndexByte(p.pName, 0) >= 0 {
			return nil, fmt.Errorf("gia: model name contains NUL")
		}
		data['n'] = append([]byte(p.pName), 0)
	}
	// Preserve the original writer's tag order for all known extensions.
	order := []byte("acdioefghkwmprsbqutvn")
	seen := map[byte]bool{}
	var result []aigerExtension
	for _, tag := range order {
		if payload, ok := data[tag]; ok {
			if len(payload) > math.MaxInt32 {
				return nil, fmt.Errorf("gia: extension too large")
			}
			result = append(result, aigerExtension{tag, payload})
			seen[tag] = true
		}
	}
	var extra []int
	for tag := range data {
		if !seen[tag] {
			extra = append(extra, int(tag))
		}
	}
	sort.Ints(extra)
	for _, tag := range extra {
		result = append(result, aigerExtension{byte(tag), data[byte(tag)]})
	}
	return result, nil
}
