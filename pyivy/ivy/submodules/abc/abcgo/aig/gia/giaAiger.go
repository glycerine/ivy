package gia

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [giaAiger.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Procedures to read/write binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAiger.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// AigerReadOptions retains source read modes. SkipStrash preserves object IDs;
// structural hashing refuses to invalidate node-indexed extension data.
type AigerReadOptions struct {
	Simple, SkipStrash, Check bool
	MaxObjects                int
}
type AigerWriteOptions struct{ Compact, OmitSymbols, OmitExtensions, CommentNewline bool }

// AigerTimingData preserves serialized hierarchy/timing data without claiming
// to provide a timing-manager implementation. Native timing managers may expose
// MarshalAigerTiming(bool) ([]byte,error) to provide a fresh hierarchy payload.
type AigerTimingData struct{ Hierarchy, Timing []byte }

type aigerCursor struct {
	data []byte
	pos  int
}

func (c *aigerCursor) line() ([]byte, error) {
	if c.pos >= len(c.data) {
		return nil, io.ErrUnexpectedEOF
	}
	end := bytes.IndexByte(c.data[c.pos:], '\n')
	if end < 0 {
		return nil, fmt.Errorf("gia: unterminated AIGER line")
	}
	line := c.data[c.pos : c.pos+end]
	c.pos += end + 1
	return line, nil
}
func (c *aigerCursor) unsigned() (uint32, error) {
	data := c.data[c.pos:]
	if len(data) > 0 && data[0] < 0x80 {
		c.pos++
		return uint32(data[0]), nil
	}
	if len(data) > 1 && data[1] < 0x80 {
		c.pos += 2
		return uint32(data[0]&0x7f) | uint32(data[1])<<7, nil
	}
	v, err := aigerReadUvar(&data)
	if err != nil {
		return 0, err
	}
	c.pos = len(c.data) - len(data)
	return v, nil
}

// A record contains at most three integers. Reuse the caller's buffer while
// accepting the same whitespace and unsigned decimal syntax as the header.
func aigerNumbers(line []byte, dst *[3]uint32) ([]uint32, error) {
	n := 0
	for field := range bytes.FieldsSeq(line) {
		if n == len(dst) {
			return nil, fmt.Errorf("gia: too many AIGER record fields")
		}
		v, err := strconv.ParseUint(string(field), 10, 31)
		if err != nil {
			return nil, fmt.Errorf("gia: invalid AIGER integer %q", field)
		}
		dst[n] = uint32(v)
		n++
	}
	return dst[:n], nil
}

func Gia_FileFixName(name string) string     { return strings.ReplaceAll(name, ">", "\\") }
func Gia_FileNameGeneric(name string) string { return strings.TrimSuffix(name, filepath.Ext(name)) }
func Gia_FileSize(name string) (int64, error) {
	info, err := os.Stat(name)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
func Gia_FileWriteBufferSize(w io.Writer, size int32) error {
	if size < 0 {
		return fmt.Errorf("gia: negative extension size")
	}
	var data [4]byte
	Gia_AigerWriteInt(data[:], size)
	n, err := w.Write(data[:])
	if err == nil && n != 4 {
		return io.ErrShortWrite
	}
	return err
}
func Gia_AigerCollectLiterals(p *Gia_Man_t) *Vec_Int_t {
	result := vec.Vec_IntAlloc(int(Gia_ManCoNum(p)))
	for i := Gia_ManPoNum(p); i < Gia_ManCoNum(p); i++ {
		o := Gia_ManCo(p, i)
		vec.Vec_IntPush(result, Gia_ObjFaninLit0(o, o.id))
	}
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		o := Gia_ManCo(p, i)
		vec.Vec_IntPush(result, Gia_ObjFaninLit0(o, o.id))
	}
	return result
}
func Gia_AigerReadLiterals(pos *[]byte, count int32) (*Vec_Int_t, error) {
	if count < 0 {
		return nil, fmt.Errorf("gia: invalid compact literal count")
	}
	if count == 0 {
		return vec.Vec_IntAlloc(0), nil
	}
	data := *pos
	first, err := aigerReadUvar(&data)
	if err != nil || first > math.MaxInt32 {
		return nil, fmt.Errorf("gia: invalid first compact literal")
	}
	lits := make([]int32, count)
	lits[0] = int32(first)
	for i := 1; i < len(lits); i++ {
		item, err := aigerReadUvar(&data)
		if err != nil {
			return nil, err
		}
		delta := int64(item >> 1)
		if item&1 != 0 {
			delta = -delta
		}
		next := int64(lits[i-1]) + delta
		if next < 0 || next > math.MaxInt32 {
			return nil, fmt.Errorf("gia: compact literal overflow")
		}
		lits[i] = int32(next)
	}
	*pos = data
	return aigerVecInts(lits), nil
}
func Gia_AigerWriteLiterals(lits *Vec_Int_t) (*Vec_Str_t, error) {
	var out []byte
	values := vec.Vec_IntArray(lits)
	for i, value := range values {
		if value < 0 {
			return nil, fmt.Errorf("gia: negative compact literal")
		}
		if i == 0 {
			out = aigerAppendUvar(out, uint32(value))
			continue
		}
		delta := int64(value) - int64(values[i-1])
		var item uint32
		if delta < 0 {
			item = uint32(-delta)<<1 | 1
		} else {
			item = uint32(delta) << 1
		}
		out = aigerAppendUvar(out, item)
	}
	return aigerVecBytes(out), nil
}

func ParseAiger(contents []byte, opts AigerReadOptions) (*Gia_Man_t, error) {
	return parseAiger(contents, opts, 0)
}
func parseAiger(contents []byte, opts AigerReadOptions, depth int) (*Gia_Man_t, error) {
	if depth > 64 {
		return nil, fmt.Errorf("gia: nested AIGER extension limit exceeded")
	}
	cursor := &aigerCursor{data: contents}
	header, err := cursor.line()
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(header))
	if len(fields) < 6 || len(fields) > 10 || fields[0] != "aig" && fields[0] != "aig2" && fields[0] != "aag" {
		return nil, fmt.Errorf("gia: invalid AIGER header")
	}
	var counts [9]uint32
	for i, field := range fields[1:] {
		n, err := strconv.ParseUint(field, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("gia: invalid AIGER integer %q", field)
		}
		counts[i] = uint32(n)
	}
	var record [3]uint32
	m, inputs, latches, ordinary, ands, bad, constraints := counts[0], counts[1], counts[2], counts[3], counts[4], counts[5], counts[6]
	if counts[7] != 0 || counts[8] != 0 {
		return nil, fmt.Errorf("gia: AIGER liveness properties are unsupported")
	}
	ascii, compact := fields[0] == "aag", fields[0] == "aig2"
	sum := uint64(inputs) + uint64(latches) + uint64(ands)
	outputs := uint64(ordinary) + uint64(bad) + uint64(constraints)
	if uint64(m) < sum || !ascii && uint64(m) != sum {
		return nil, fmt.Errorf("gia: AIGER variable count mismatch")
	}
	objects := sum + outputs + uint64(latches) + 1
	limit := opts.MaxObjects
	if limit <= 0 {
		limit = 4_000_000
	}
	if objects > uint64(limit) || objects >= uint64(objectNoFanin) || m > math.MaxInt32/2 {
		return nil, fmt.Errorf("gia: AIGER object limit exceeded")
	}
	// Validate minimum explicitly encoded data before allocating declared records.
	minimum := uint64(ands) * 2
	if !compact {
		minimum += (uint64(latches) + outputs) * 2
	} else {
		minimum += uint64(latches) + outputs
	}
	if minimum > uint64(len(contents)-cursor.pos) {
		return nil, io.ErrUnexpectedEOF
	}
	p := Gia_ManStart(int32(objects))
	hash := !opts.Simple && !opts.SkipStrash
	p.fGiaSimple = giaBool(!hash)
	p.nConstrs = int32(constraints)
	p.aigerBadOutputs = int32(bad)
	p.aigerLatchInits = make([]int32, latches)
	// ASCII records may use sparse, explicitly numbered variables.
	var variables map[uint32]int32
	if ascii {
		variables = map[uint32]int32{0: 0}
	}
	for i := uint32(0); i < inputs+latches; i++ {
		id := i + 1
		if ascii && i < inputs {
			line, err := cursor.line()
			if err != nil {
				return nil, err
			}
			nums, err := aigerNumbers(line, &record)
			if err != nil || len(nums) != 1 || nums[0]&1 != 0 || nums[0] == 0 || nums[0]>>1 > m {
				return nil, fmt.Errorf("gia: invalid AAG input literal")
			}
			id = nums[0] >> 1
		}
		if ascii && i >= inputs {
			continue
		}
		if _, exists := variables[id]; exists {
			return nil, fmt.Errorf("gia: duplicate input variable")
		}
		lit := Gia_ManAppendCi(p)
		if ascii {
			variables[id] = lit
		}
	}
	drivers := make([]uint32, int(uint64(latches)+outputs))
	if compact {
		data := contents[cursor.pos:]
		lits, err := Gia_AigerReadLiterals(&data, int32(len(drivers)))
		if err != nil {
			return nil, err
		}
		for i, value := range vec.Vec_IntArray(lits) {
			drivers[i] = uint32(value)
		}
		cursor.pos = len(contents) - len(data)
	} else {
		for i := uint32(0); i < latches; i++ {
			line, err := cursor.line()
			if err != nil {
				return nil, err
			}
			nums, err := aigerNumbers(line, &record)
			min, max := 1, 2
			if ascii {
				min, max = 2, 3
			}
			if err != nil || len(nums) < min || len(nums) > max {
				return nil, fmt.Errorf("gia: invalid latch line")
			}
			current := (inputs + i + 1) << 1
			driver := nums[0]
			if ascii {
				current, driver = nums[0], nums[1]
				if _, exists := variables[current>>1]; current == 0 || current&1 != 0 || current>>1 > m || exists {
					return nil, fmt.Errorf("gia: invalid AAG latch variable")
				}
				variables[current>>1] = Gia_ManAppendCi(p)
			}
			drivers[i] = driver
			init := uint32(0)
			if len(nums) == max {
				init = nums[max-1]
			}
			switch init {
			case 0, 1:
				p.aigerLatchInits[i] = int32(init)
			default:
				if init != current {
					return nil, fmt.Errorf("gia: invalid AIGER latch initialization")
				}
				p.aigerLatchInits[i] = -1
			}
		}
		for i := 0; i < int(outputs); i++ {
			line, err := cursor.line()
			if err != nil {
				return nil, err
			}
			nums, err := aigerNumbers(line, &record)
			if err != nil || len(nums) != 1 {
				return nil, fmt.Errorf("gia: invalid AIGER output line")
			}
			drivers[int(latches)+i] = nums[0]
		}
	}
	if hash {
		Gia_ManHashAlloc(p)
	}
	// Binary literals are their own mapping up to the first simplified gate.
	firstRemapped := m + 1
	var binaryMap []int32
	remap := func(lit uint32) (int32, error) {
		if !ascii && lit>>1 <= m {
			if lit>>1 >= firstRemapped {
				return binaryMap[(lit>>1)-firstRemapped] ^ int32(lit&1), nil
			}
			return int32(lit), nil
		}
		value, ok := variables[lit>>1]
		if !ok || lit>>1 > m {
			return 0, fmt.Errorf("gia: undefined AIGER literal %d", lit)
		}
		return value ^ int32(lit&1), nil
	}
	for i := uint32(0); i < ands; i++ {
		lhs := (inputs + latches + i + 1) << 1
		var a, b uint32
		if ascii {
			line, err := cursor.line()
			if err != nil {
				return nil, err
			}
			nums, err := aigerNumbers(line, &record)
			if err != nil || len(nums) != 3 || nums[0] == 0 || nums[0]&1 != 0 || nums[0]>>1 > m {
				return nil, fmt.Errorf("gia: invalid AAG AND line")
			}
			lhs, a, b = nums[0], nums[1], nums[2]
		} else {
			d1, err := cursor.unsigned()
			if err != nil || d1 > lhs {
				return nil, fmt.Errorf("gia: invalid AIGER AND delta")
			}
			b = lhs - d1
			d0, err := cursor.unsigned()
			if err != nil || d0 > b {
				return nil, fmt.Errorf("gia: invalid AIGER AND delta")
			}
			a = b - d0
		}
		if _, exists := variables[lhs>>1]; exists || a >= lhs || b >= lhs {
			return nil, fmt.Errorf("gia: non-topological AIGER AND")
		}
		fa, err := remap(a)
		if err != nil {
			return nil, err
		}
		fb, err := remap(b)
		if err != nil {
			return nil, err
		}
		var lit int32
		if hash || opts.Simple {
			lit = Gia_ManHashAnd(p, fa, fb)
		} else {
			lit = Gia_ManAppendAnd(p, fa, fb)
		}
		if ascii {
			variables[lhs>>1] = lit
		} else if uint32(lit) != lhs {
			if binaryMap == nil {
				firstRemapped = lhs >> 1
				binaryMap = make([]int32, m-firstRemapped+1)
			}
			binaryMap[lhs>>1-firstRemapped] = lit
		}
	}
	if hash {
		Gia_ManHashStop(p)
	}
	for i := 0; i < int(outputs); i++ {
		lit, err := remap(drivers[int(latches)+i])
		if err != nil {
			return nil, err
		}
		if i >= int(outputs)-int(constraints) {
			lit ^= 1
		}
		Gia_ManAppendCo(p, lit)
	}
	for i := uint32(0); i < latches; i++ {
		lit, err := remap(drivers[i])
		if err != nil {
			return nil, err
		}
		Gia_ManAppendCo(p, lit)
	}
	p.nRegs = int32(latches)
	if hash && Gia_ManHasDangling(p) != 0 {
		p = Gia_ManCleanup(p)
	}
	if !hash {
		p.fGiaSimple = giaBool(opts.Simple)
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 && obj.iDiff0 == obj.iDiff1 {
				// Raw opposite-phase pins also need simple mode to copy both edges.
				p.fGiaSimple = 1
				if Gia_ObjIsBuf(obj) != 0 {
					p.nBufs++
				}
			}
		}
	}
	p.aigerSymbols = map[byte][]string{}
	p.aigerExtensions = map[byte][]byte{}
	renumbered := uint64(p.nObjs) != objects
	if err := aigerReadFooter(p, contents[cursor.pos:], int(ordinary), int(bad), int(constraints), opts, depth, renumbered); err != nil {
		return nil, err
	}
	return p, nil
}

// aigerNamedProperties retains ABC's legacy name-based property convention.
func aigerNamedProperties(p *Gia_Man_t, cleanup bool) *Gia_Man_t {
	names := p.aigerSymbols['o']
	var types *Vec_Int_t
	for i, name := range names {
		var kind int32
		switch {
		case strings.HasPrefix(name, "constraint:"):
			kind = 1
		case strings.HasPrefix(name, "invariant:"):
			kind = 2
		default:
			continue
		}
		if types == nil {
			types = vec.Vec_IntStart(int(Gia_ManPoNum(p)))
		}
		vec.Vec_IntWriteEntry(types, i, kind)
	}
	if types == nil {
		return p
	}
	for i := Gia_ManPoNum(p) - p.nConstrs; i < Gia_ManPoNum(p); i++ {
		vec.Vec_IntWriteEntry(types, int(i), 1)
	}
	q := Gia_ManDupWithConstraints(p, types)
	// Header constraints were inverted during parsing. The duplication routine
	// inverts every marked constraint, so restore these outputs' existing phase.
	for i := Gia_ManPoNum(q) - p.nConstrs; i < Gia_ManPoNum(q); i++ {
		Gia_ObjFlipFaninC0(Gia_ManPo(q, i))
	}
	q.aigerBadOutputs = p.aigerBadOutputs
	q.aigerLatchInits = p.aigerLatchInits
	if cleanup && Gia_ManHasDangling(q) != 0 {
		q = Gia_ManCleanup(q)
	}
	return q
}

// ParseAiger preserves explicit reset values. The legacy reader changes
// register coordinates and models unknown resets with first-frame inputs.
func aigerZeroInitialize(p *Gia_Man_t) *Gia_Man_t {
	var init []byte
	for i, reset := range p.aigerLatchInits {
		if reset == 0 {
			continue
		}
		if init == nil {
			init = bytes.Repeat([]byte{'0'}, len(p.aigerLatchInits))
		}
		if reset == 1 {
			init[i] = '1'
		} else {
			init[i] = 'X'
		}
	}
	if init == nil {
		return p
	}
	q := Gia_ManDupZeroUndc(p, string(init), 0, p.fGiaSimple, 0)
	q.nConstrs, q.aigerBadOutputs = p.nConstrs, p.aigerBadOutputs
	q.aigerLatchInits = make([]int32, q.nRegs)
	return q
}

func Gia_AigerReadFromMemory(contents []byte, size, fSimple, fSkip, fCheck int32) (*Gia_Man_t, error) {
	if size < 0 || int(size) > len(contents) {
		return nil, fmt.Errorf("gia: invalid AIGER memory size")
	}
	p, err := ParseAiger(contents[:size], AigerReadOptions{Simple: fSimple != 0, SkipStrash: fSkip != 0, Check: fCheck != 0})
	if err != nil {
		return nil, err
	}
	return aigerZeroInitialize(aigerNamedProperties(p, fSimple == 0 && fSkip == 0)), nil
}
func Gia_AigerRead(name string, fSimple, fSkip, fCheck int32) (*Gia_Man_t, error) {
	name = Gia_FileFixName(name)
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(strings.ToLower(name), ".gz") {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		data, err = io.ReadAll(reader)
		closeErr := reader.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
	}
	p, err := Gia_AigerReadFromMemory(data, int32(len(data)), fSimple, fSkip, fCheck)
	if err != nil {
		return nil, err
	}
	p.PSpec = name
	if p.pName == "" {
		p.pName = Gia_FileNameGeneric(name)
	}
	return p, nil
}

func aigerReadFooter(p *Gia_Man_t, data []byte, ordinary, bad, constraints int, opts AigerReadOptions, depth int, renumbered bool) error {
	lengths := map[byte]int{'i': int(Gia_ManPiNum(p)), 'l': int(p.nRegs), 'o': ordinary, 'b': bad, 'c': constraints}
	// Keep the symbol lines in one owned string. The caller may reuse the
	// input bytes, and binary extension payloads can be much larger than names.
	end := 0
	for end < len(data) {
		rest := data[end:]
		if rest[0] == 'c' && (len(rest) == 1 || rest[1] < '0' || rest[1] > '9') {
			break
		}
		n := bytes.IndexByte(rest, '\n')
		if n < 0 {
			return fmt.Errorf("gia: unterminated symbol line")
		}
		end += n + 1
	}
	symbols := string(data[:end])
	data = data[end:]
	if len(data) > 0 {
		data = data[1:]
	} // comment marker
	type symbolIndex struct {
		kind  byte
		index int
	}
	var empty map[symbolIndex]bool
	for record := range strings.Lines(symbols) {
		line := strings.TrimSuffix(strings.TrimSuffix(record, "\n"), "\r")
		space := strings.IndexByte(line, ' ')
		if space < 2 {
			return fmt.Errorf("gia: invalid AIGER symbol")
		}
		kind := line[0]
		count, ok := lengths[kind]
		index, err := strconv.Atoi(line[1:space])
		if !ok || err != nil || index < 0 || index >= count {
			return fmt.Errorf("gia: AIGER symbol index out of range")
		}
		if p.aigerSymbols[kind] == nil {
			p.aigerSymbols[kind] = make([]string, count)
		}
		key := symbolIndex{kind, index}
		if p.aigerSymbols[kind][index] != "" || empty[key] {
			return fmt.Errorf("gia: duplicate AIGER symbol")
		}
		name := line[space+1:]
		p.aigerSymbols[kind][index] = name
		if name == "" {
			if empty == nil {
				empty = make(map[symbolIndex]bool)
			}
			empty[key] = true
		}
	}
	if len(data) > 0 && data[0] == '\r' {
		data = data[1:]
	}
	if len(data) > 0 && data[0] == '\n' {
		data = data[1:]
	}
	for len(data) > 0 {
		tag := data[0]
		if tag == 'n' && len(data) > 1 && ((data[1] >= 'A' && data[1] <= 'Z') || (data[1] >= 'a' && data[1] <= 'z') || (data[1] >= '0' && data[1] <= '9')) {
			end := bytes.IndexByte(data[1:], 0)
			if end < 0 {
				break
			}
			p.pName = string(data[1 : 1+end])
			data = data[end+2:]
			continue
		}
		if tag < 'a' || tag > 'z' {
			break
		}
		binaryRecord := len(data) > 1 && data[1] < 32 && data[1] != '\n' && data[1] != '\r'
		if len(data) < 5 {
			if binaryRecord {
				return fmt.Errorf("gia: truncated extension header")
			}
			break
		}
		size := uint64(binary.BigEndian.Uint32(data[1:]))
		if size > uint64(len(data)-5) {
			if binaryRecord {
				return fmt.Errorf("gia: truncated extension payload")
			}
			break
		}
		if renumbered && aigerTagUsesObjectIDs(tag) {
			return fmt.Errorf("gia: structural hashing would invalidate object-indexed AIGER metadata; use SkipStrash")
		}
		payload := data[5 : 5+int(size)]
		if err := aigerDecodeExtension(p, tag, payload, opts, depth); err != nil {
			return fmt.Errorf("gia: extension %q: %w", tag, err)
		}
		data = data[5+int(size):]
	}
	p.aigerComment = append([]byte(nil), data...)
	aigerInstallNames(p, ordinary, bad, constraints)
	return nil
}
func aigerInstallNames(p *Gia_Man_t, ordinary, bad, constraints int) {
	if len(p.aigerSymbols) == 0 {
		return
	}
	installID := func(ids **Vec_Int_t, count, index int, name string, kind byte) {
		if len(name) < 3 || name[0] != '@' || name[1] != kind {
			return
		}
		id, err := strconv.ParseInt(name[2:], 10, 32)
		if err != nil || id < 0 {
			return
		}
		if *ids == nil {
			*ids = vec.Vec_IntStartFull(count)
		}
		vec.Vec_IntWriteEntry(*ids, index, int32(id))
	}
	inputs := make([]any, Gia_ManCiNum(p))
	anyInput := false
	for i, name := range p.aigerSymbols['i'] {
		inputs[i] = name
		anyInput = anyInput || name != ""
		installID(&p.vUserPiIds, int(Gia_ManPiNum(p)), i, name, 'i')
	}
	for i, name := range p.aigerSymbols['l'] {
		inputs[int(Gia_ManPiNum(p))+i] = name
		anyInput = anyInput || name != ""
		installID(&p.vUserFfIds, int(p.nRegs), i, name, 'l')
	}
	if anyInput {
		p.VNamesIn = vec.Vec_PtrAllocArray(inputs, len(inputs))
	}
	outputs := make([]any, Gia_ManCoNum(p))
	anyOutput := false
	for _, part := range []struct {
		kind  byte
		start int
	}{{'o', 0}, {'b', ordinary}, {'c', ordinary + bad}} {
		for i, name := range p.aigerSymbols[part.kind] {
			index := part.start + i
			outputs[index] = name
			anyOutput = anyOutput || name != ""
			installID(&p.vUserPoIds, ordinary+bad+constraints, index, name, 'o')
		}
	}
	if anyOutput {
		p.VNamesOut = vec.Vec_PtrAllocArray(outputs, len(outputs))
	}
}

func aigerTagUsesObjectIDs(tag byte) bool {
	switch tag {
	case 'e', 'q', 'm', 'k', 'g', 'v', 'u', 'p', 'w':
		return true
	}
	return !aigerKnownTag(tag)
}

type aigerWireModel struct {
	inputs, latches, ordinary, bad, constraints int
	ands                                        [][2]uint32
	drivers                                     []uint32
	ids                                         []int32
	objects                                     int
}

func aigerWire(p *Gia_Man_t, cis, ands, cos []int32, regs int32, properties bool) (*aigerWireModel, error) {
	if p == nil || regs < 0 || int(regs) > len(cis) || int(regs) > len(cos) {
		return nil, fmt.Errorf("gia: invalid AIGER register/interface counts")
	}
	model := &aigerWireModel{inputs: len(cis) - int(regs), latches: int(regs), ordinary: len(cos) - int(regs), ids: make([]int32, p.nObjs), objects: 1 + len(cis) + len(ands) + len(cos)}
	if properties {
		model.constraints = int(p.nConstrs)
		model.bad = int(p.aigerBadOutputs)
		model.ordinary -= model.constraints + model.bad
		if model.constraints < 0 || model.bad < 0 || model.ordinary < 0 {
			return nil, fmt.Errorf("gia: invalid output property counts")
		}
	}
	for i := range model.ids {
		model.ids[i] = -1
	}
	model.ids[0] = 0
	for i, id := range cis {
		if id <= 0 || id >= p.nObjs || Gia_ObjIsCi(p.pObjs[id]) == 0 || model.ids[id] != -1 {
			return nil, fmt.Errorf("gia: invalid or duplicate selected CI")
		}
		model.ids[id] = int32(i + 1)
	}
	remap := func(lit int32) (uint32, error) {
		if lit < 0 || lit>>1 >= p.nObjs || model.ids[lit>>1] < 0 {
			return 0, fmt.Errorf("gia: selected AIGER cone is not closed or topological")
		}
		return uint32(model.ids[lit>>1]<<1 | lit&1), nil
	}
	for i, id := range ands {
		if id <= 0 || id >= p.nObjs || Gia_ObjIsAnd(p.pObjs[id]) == 0 || model.ids[id] != -1 {
			return nil, fmt.Errorf("gia: invalid or duplicate selected AND")
		}
		o := p.pObjs[id]
		if Gia_ObjIsXor(o) != 0 || int(id) < len(p.PMuxes) && p.PMuxes[id] != 0 {
			return nil, fmt.Errorf("gia: normalize specialized XOR/MUX objects before writing AIGER")
		}
		a, err := remap(Gia_ObjFaninLit0(o, id))
		if err != nil {
			return nil, err
		}
		b, err := remap(Gia_ObjFaninLit1(o, id))
		if err != nil {
			return nil, err
		}
		if a > b {
			a, b = b, a
		}
		model.ands = append(model.ands, [2]uint32{a, b})
		model.ids[id] = int32(len(cis) + i + 1)
	}
	coDrivers := make([]uint32, len(cos))
	for i, id := range cos {
		if id <= 0 || id >= p.nObjs || Gia_ObjIsCo(p.pObjs[id]) == 0 || model.ids[id] != -1 {
			return nil, fmt.Errorf("gia: invalid or duplicate selected CO")
		}
		lit, err := remap(Gia_ObjFaninLit0(p.pObjs[id], id))
		if err != nil {
			return nil, err
		}
		coDrivers[i] = lit
		model.ids[id] = int32(len(cis) + len(ands) + i + 1)
	}
	po := len(cos) - int(regs)
	model.drivers = append(model.drivers, coDrivers[po:]...)
	for i, lit := range coDrivers[:po] {
		if properties && i >= po-model.constraints {
			lit ^= 1
		}
		model.drivers = append(model.drivers, lit)
	}
	return model, nil
}
func aigerAllAnds(p *Gia_Man_t) []int32 {
	var ids []int32
	for id, o := range p.pObjs {
		if Gia_ObjIsAnd(o) != 0 {
			ids = append(ids, int32(id))
		}
	}
	return ids
}
func MarshalAiger(p *Gia_Man_t, opts AigerWriteOptions) ([]byte, error) {
	return aigerMarshalManager(p, opts, 0)
}
func aigerMarshalManager(p *Gia_Man_t, opts AigerWriteOptions, depth int) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("gia: nil AIGER manager")
	}
	model, err := aigerWire(p, vec.Vec_IntArray(p.vCis), aigerAllAnds(p), vec.Vec_IntArray(p.vCos), p.nRegs, true)
	if err != nil {
		return nil, err
	}
	meta, err := aigerWireMetadata(p, model)
	if err != nil {
		return nil, err
	}
	return aigerMarshalWire(meta, model, opts, depth)
}
func aigerMarshalWire(p *Gia_Man_t, model *aigerWireModel, opts AigerWriteOptions, depth int) ([]byte, error) {
	if len(p.aigerLatchInits) != 0 && len(p.aigerLatchInits) != model.latches {
		return nil, fmt.Errorf("gia: latch initialization count does not match register count")
	}
	if depth > 64 {
		return nil, fmt.Errorf("gia: nested AIGER write limit")
	}
	if opts.Compact {
		for _, init := range p.aigerLatchInits {
			if init != 0 {
				return nil, fmt.Errorf("gia: compact aig2 cannot encode nonzero/unknown latch reset")
			}
		}
	}
	var out bytes.Buffer
	magic := "aig"
	if opts.Compact {
		magic = "aig2"
	}
	fmt.Fprintf(&out, "%s %d %d %d %d %d", magic, model.inputs+model.latches+len(model.ands), model.inputs, model.latches, model.ordinary, len(model.ands))
	if model.bad != 0 || model.constraints != 0 {
		fmt.Fprintf(&out, " %d %d", model.bad, model.constraints)
	}
	out.WriteByte('\n')
	if opts.Compact {
		lits := make([]int32, len(model.drivers))
		for i, v := range model.drivers {
			lits[i] = int32(v)
		}
		encoded, err := Gia_AigerWriteLiterals(aigerVecInts(lits))
		if err != nil {
			return nil, err
		}
		out.Write(vec.Vec_StrArray(encoded))
	} else {
		for i, lit := range model.drivers {
			fmt.Fprintf(&out, "%d", lit)
			if i < model.latches && i < len(p.aigerLatchInits) && p.aigerLatchInits[i] != 0 {
				init := p.aigerLatchInits[i]
				if init == -1 {
					init = int32((model.inputs + i + 1) << 1)
				} else if init != 1 {
					return nil, fmt.Errorf("gia: invalid latch reset state")
				}
				fmt.Fprintf(&out, " %d", init)
			}
			out.WriteByte('\n')
		}
	}
	for i, fanin := range model.ands {
		lhs := uint32((model.inputs + model.latches + i + 1) << 1)
		if fanin[1] >= lhs {
			return nil, fmt.Errorf("gia: non-topological encoded AND")
		}
		out.Write(aigerAppendUvar(nil, lhs-fanin[1]))
		out.Write(aigerAppendUvar(nil, fanin[1]-fanin[0]))
	}
	if !opts.OmitSymbols {
		if err := aigerWriteSymbols(&out, p, model); err != nil {
			return nil, err
		}
	}
	out.WriteByte('c')
	if opts.CommentNewline {
		out.WriteByte('\n')
	}
	if !opts.OmitExtensions {
		extensions, err := aigerEncodeExtensions(p, model, opts, depth)
		if err != nil {
			return nil, err
		}
		for _, extension := range extensions {
			out.WriteByte(extension.tag)
			out.Write(binary.BigEndian.AppendUint32(nil, uint32(len(extension.data))))
			out.Write(extension.data)
		}
	}
	out.Write(p.aigerComment)
	return out.Bytes(), nil
}
func aigerWriteSymbols(out *bytes.Buffer, p *Gia_Man_t, model *aigerWireModel) error {
	lengths := map[byte]int{'i': model.inputs, 'l': model.latches, 'o': model.ordinary, 'b': model.bad, 'c': model.constraints}
	names := map[byte][]string{}
	for tag, n := range lengths {
		names[tag] = make([]string, n)
		copy(names[tag], p.aigerSymbols[tag])
	}
	if p.VNamesIn != nil {
		raw := vec.Vec_PtrArray(p.VNamesIn)
		if len(raw) != model.inputs+model.latches {
			return fmt.Errorf("gia: input name count mismatch")
		}
		for i, value := range raw {
			if name, ok := value.(string); ok && name != "" {
				if i < model.inputs {
					names['i'][i] = name
				} else {
					names['l'][i-model.inputs] = name
				}
			}
		}
	}
	if p.VNamesOut != nil {
		raw := vec.Vec_PtrArray(p.VNamesOut)
		if len(raw) < model.ordinary+model.bad+model.constraints {
			return fmt.Errorf("gia: output name count mismatch")
		}
		for i, value := range raw[:model.ordinary+model.bad+model.constraints] {
			name, ok := value.(string)
			if !ok || name == "" {
				continue
			}
			switch {
			case i < model.ordinary:
				names['o'][i] = name
			case i < model.ordinary+model.bad:
				names['b'][i-model.ordinary] = name
			default:
				names['c'][i-model.ordinary-model.bad] = name
			}
		}
	}
	for _, part := range []struct {
		kind byte
		ids  *Vec_Int_t
	}{{'i', p.vUserPiIds}, {'l', p.vUserFfIds}} {
		if part.ids == nil {
			continue
		}
		values := vec.Vec_IntArray(part.ids)
		if len(values) != len(names[part.kind]) {
			return fmt.Errorf("gia: custom symbol ID count mismatch")
		}
		for i, id := range values {
			if id >= 0 && names[part.kind][i] == "" {
				names[part.kind][i] = fmt.Sprintf("@%c%d", part.kind, id)
			}
		}
	}
	if p.vUserPoIds != nil {
		values := vec.Vec_IntArray(p.vUserPoIds)
		if len(values) != model.ordinary+model.bad+model.constraints {
			return fmt.Errorf("gia: custom output ID count mismatch")
		}
		for i, id := range values {
			tag, index := byte('o'), i
			if i >= model.ordinary+model.bad {
				tag, index = 'c', i-model.ordinary-model.bad
			} else if i >= model.ordinary {
				tag, index = 'b', i-model.ordinary
			}
			if id >= 0 && names[tag][index] == "" {
				names[tag][index] = fmt.Sprintf("@o%d", id)
			}
		}
	}
	for _, tag := range []byte{'i', 'l', 'o', 'b', 'c'} {
		for i, name := range names[tag] {
			if strings.ContainsAny(name, "\r\n\x00") {
				return fmt.Errorf("gia: invalid symbol name")
			}
			if name != "" {
				fmt.Fprintf(out, "%c%d %s\n", tag, i, name)
			}
		}
	}
	return nil
}
func Gia_AigerWriteIntoMemoryStr(p *Gia_Man_t) (*Vec_Str_t, error) {
	data, err := MarshalAiger(p, AigerWriteOptions{OmitExtensions: true, OmitSymbols: true})
	if err != nil {
		return nil, err
	}
	return aigerVecBytes(data), nil
}
func Gia_AigerWriteIntoMemoryStrPart(p *Gia_Man_t, cis, ands, cos *Vec_Int_t, regs int32) (*Vec_Str_t, error) {
	if p == nil || cis == nil || ands == nil || cos == nil {
		return nil, fmt.Errorf("gia: nil partial AIGER selection")
	}
	model, err := aigerWire(p, vec.Vec_IntArray(cis), vec.Vec_IntArray(ands), vec.Vec_IntArray(cos), regs, false)
	if err != nil {
		return nil, err
	}
	meta := &Gia_Man_t{}
	if len(p.aigerLatchInits) > 0 {
		if len(p.aigerLatchInits) != int(p.nRegs) {
			return nil, fmt.Errorf("gia: invalid source latch initialization count")
		}
		selected := vec.Vec_IntArray(cis)
		original := vec.Vec_IntArray(p.vCis)
		states := make(map[int32]int32)
		for i, id := range original[len(original)-int(p.nRegs):] {
			states[id] = p.aigerLatchInits[i]
		}
		meta.aigerLatchInits = make([]int32, regs)
		for i, id := range selected[len(selected)-int(regs):] {
			meta.aigerLatchInits[i] = states[id]
		}
	}
	data, err := aigerMarshalWire(meta, model, AigerWriteOptions{OmitExtensions: true, OmitSymbols: true}, 0)
	if err != nil {
		return nil, err
	}
	return aigerVecBytes(data), nil
}
func Gia_AigerWrite(p *Gia_Man_t, name string, symbols, compact, newline int32) error {
	if p == nil || Gia_ManCoNum(p) == 0 {
		return fmt.Errorf("gia: cannot write AIG without outputs")
	}
	data, err := MarshalAiger(p, AigerWriteOptions{Compact: compact != 0, CommentNewline: newline != 0})
	if err != nil {
		return err
	}
	return aigerWriteFile(name, data)
}
func Gia_AigerWriteSimple(p *Gia_Man_t, name string) error {
	if p == nil || Gia_ManPoNum(p) == 0 {
		return fmt.Errorf("gia: cannot write simple AIG without primary outputs")
	}
	data, err := Gia_AigerWriteIntoMemoryStr(p)
	if err != nil {
		return err
	}
	return aigerWriteFile(name, vec.Vec_StrArray(data))
}
func Gia_DumpAiger(p *Gia_Man_t, prefix string, number, digits int32) error {
	if digits < 0 || digits > 1000 {
		return fmt.Errorf("gia: invalid dump filename width")
	}
	return Gia_AigerWrite(p, fmt.Sprintf("%s%0*d.aig", prefix, int(digits), number), 0, 0, 0)
}
func aigerWriteFile(name string, data []byte) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	var writer io.Writer = file
	var zipped *gzip.Writer
	if strings.HasSuffix(strings.ToLower(name), ".gz") {
		zipped = gzip.NewWriter(file)
		writer = zipped
	}
	n, writeErr := writer.Write(data)
	if writeErr == nil && n != len(data) {
		writeErr = io.ErrShortWrite
	}
	if zipped != nil {
		if err := zipped.Close(); writeErr == nil {
			writeErr = err
		}
	}
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func (p *Gia_Man_t) AigerLatchInitialStates() []int32 {
	return append([]int32(nil), p.aigerLatchInits...)
}
func (p *Gia_Man_t) SetAigerLatchInitialStates(states []int32) error {
	if len(states) != int(p.nRegs) {
		return fmt.Errorf("gia: latch reset count mismatch")
	}
	for _, state := range states {
		if state != 0 && state != 1 && state != -1 {
			return fmt.Errorf("gia: invalid latch reset state")
		}
	}
	p.aigerLatchInits = append([]int32(nil), states...)
	return nil
}
func (p *Gia_Man_t) AigerExtension(tag byte) []byte {
	return append([]byte(nil), p.aigerExtensions[tag]...)
}
func (p *Gia_Man_t) SetAigerExtension(tag byte, payload []byte) error {
	if tag < 'a' || tag > 'z' {
		return fmt.Errorf("gia: invalid extension tag")
	}
	return aigerDecodeExtension(p, tag, payload, AigerReadOptions{SkipStrash: true}, 0)
}
func (p *Gia_Man_t) AigerComment() []byte           { return append([]byte(nil), p.aigerComment...) }
func (p *Gia_Man_t) SetAigerComment(comment []byte) { p.aigerComment = append([]byte(nil), comment...) }
func (p *Gia_Man_t) SetAigerProperties(bad, constraints int32) error {
	if bad < 0 || constraints < 0 || int64(bad)+int64(constraints) > int64(Gia_ManPoNum(p)) {
		return fmt.Errorf("gia: invalid property counts")
	}
	p.aigerBadOutputs = bad
	p.nConstrs = constraints
	return nil
}
