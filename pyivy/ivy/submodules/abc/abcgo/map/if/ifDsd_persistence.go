package if_

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// dsd1 is the original little-endian GCC/Clang layout: an eight-byte
// packed node header followed by fanin literals, rounded to whole words.
// Encode fields explicitly; Go's record layout is deliberately different.
type ifDsdFileWriter struct {
	out io.Writer
	err error
}

func (w *ifDsdFileWriter) data(data []byte) {
	if w.err != nil {
		return
	}
	var n int
	n, w.err = w.out.Write(data)
	if w.err == nil && n != len(data) {
		w.err = io.ErrShortWrite
	}
}
func (w *ifDsdFileWriter) integer(value int32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(value))
	w.data(b[:])
}
func (w *ifDsdFileWriter) word(value word) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], value)
	w.data(b[:])
}
func ifDsdWrite(p *If_DsdMan_t, out io.Writer) error {
	w := ifDsdFileWriter{out: out}
	w.data([]byte(DSD_VERSION))
	w.integer(p.nVars)
	w.integer(p.LutSize)
	w.integer(If_DsdManObjNum(p))
	for id := int32(2); id < If_DsdManObjNum(p); id++ {
		o := If_DsdVecObj(&p.vObjs, id)
		words := 1 + (o.nFans+1)/2
		record := make([]byte, 8*words)
		binary.LittleEndian.PutUint32(record, uint32(o.Id))
		flags := uint32(o.Type) | uint32(o.nSupp)<<3 | uint32(o.fMark)<<8 | uint32(o.Count)<<9 | uint32(o.nFans)<<27
		binary.LittleEndian.PutUint32(record[4:], flags)
		for i, fan := range o.pFans {
			binary.LittleEndian.PutUint32(record[8+4*i:], uint32(fan))
		}
		w.integer(words)
		w.data(record)
		if o.Type == IF_DSD_PRIME {
			w.integer(If_DsdObjTruthId(p, o))
		}
	}
	for v := int32(3); v <= p.nVars; v++ {
		memory := p.vTtMem[v]
		w.integer(int32(vec.Vec_MemEntryNum(memory)))
		for i := 0; i < vec.Vec_MemEntryNum(memory); i++ {
			for _, value := range vec.Vec_MemReadEntry(memory, i) {
				w.word(value)
			}
		}
		w.integer(int32(Vec_PtrSize(p.vTtDecs[v])))
		for _, entry := range Vec_PtrArray(p.vTtDecs[v]) {
			sets := entry.(*Vec_Int_t)
			w.integer(int32(Vec_IntSize(sets)))
			for _, value := range Vec_IntArray(sets) {
				w.integer(value)
			}
		}
	}
	w.integer(p.nConfigWords)
	w.integer(p.nTtBits)
	if p.vConfigs == nil {
		w.integer(0)
	} else {
		w.integer(int32(vec.Vec_WrdSize(p.vConfigs)))
		for _, value := range vec.Vec_WrdArray(p.vConfigs) {
			w.word(value)
		}
	}
	w.integer(int32(len(p.pCellStr)))
	w.data([]byte(p.pCellStr))
	return w.err
}
func If_DsdManSave(p *If_DsdMan_t, fileName string) {
	if fileName == "" {
		fileName = p.pStore
	}
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Writing DSD manager file %q has failed.\n", fileName)
		return
	}
	buffer := bufio.NewWriter(file)
	err = ifDsdWrite(p, buffer)
	if err == nil {
		err = buffer.Flush()
	}
	if closed := file.Close(); err == nil {
		err = closed
	}
	if err != nil {
		fmt.Printf("Writing DSD manager file %q has failed: %v.\n", fileName, err)
	}
}

var ifDsdBadFormat = errors.New("unrecognized DSD format")

type ifDsdFileReader struct {
	input  []byte
	offset int
	err    error
}

func (r *ifDsdFileReader) remaining() int { return len(r.input) - r.offset }

// Borrow raw bytes only while decoding; retained node/cache/configuration
// values are copied into their typed owners.
func (r *ifDsdFileReader) data(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || n > r.remaining() {
		r.err = io.ErrUnexpectedEOF
		return nil
	}
	data := r.input[r.offset : r.offset+n]
	r.offset += n
	return data
}

func (r *ifDsdFileReader) integer() int32 {
	data := r.data(4)
	if r.err != nil {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(data))
}
func (r *ifDsdFileReader) count(stride int) int {
	n := r.integer()
	if n < 0 {
		r.err = errors.New("negative DSD count")
		return 0
	}
	if r.err != nil {
		return 0
	}
	if int(n) > r.remaining()/stride {
		r.err = io.ErrUnexpectedEOF
		return 0
	}
	return int(n)
}
func (r *ifDsdFileReader) word() word {
	data := r.data(8)
	if r.err != nil {
		return 0
	}
	return binary.LittleEndian.Uint64(data)
}
func ifDsdRead(data []byte, fileName string) (*If_DsdMan_t, error) {
	r := ifDsdFileReader{input: data}
	if string(r.data(4)) != DSD_VERSION {
		return nil, ifDsdBadFormat
	}
	variables, lut, nodes := r.integer(), r.integer(), r.integer()
	if r.err != nil {
		return nil, r.err
	}
	if variables < 0 || variables > 16 || lut != 0 && (lut < 2 || lut > 6) || nodes < 2 || nodes > math.MaxInt32/2 {
		return nil, errors.New("invalid DSD manager dimensions")
	}
	minimum := int64(nodes-2)*20 + int64(max(variables-2, 0))*8 + 16
	if minimum > int64(r.remaining()) {
		return nil, io.ErrUnexpectedEOF
	}
	p := If_DsdManAlloc(variables, 0)
	success := false
	defer func() {
		if !success {
			If_DsdManFree(p, 0)
		}
	}()
	p.pStore, p.LutSize = fileName, lut
	// Zero is the source's "no decomposition LUT" mode. Do not reproduce
	// Load's undefined BuildXY(0) negative shift; it needs no SAT instance.
	if lut != 0 {
		p.pSat = If_ManSatBuildXY(lut)
	}
	p.nBins = abcutil.Abc_PrimeCudd(uint32(2 * nodes))
	p.pBins = make([]int32, p.nBins)
	for id := int32(2); id < nodes; id++ {
		words := r.integer()
		if r.err != nil {
			return nil, r.err
		}
		if words < 2 || words > 9 {
			return nil, errors.New("invalid DSD node record length")
		}
		record := r.data(int(words) * 8)
		if r.err != nil {
			return nil, r.err
		}
		savedID := int32(binary.LittleEndian.Uint32(record))
		flags := binary.LittleEndian.Uint32(record[4:])
		kind, support, fans := int32(flags&7), int32(flags>>3&31), int32(flags>>27)
		if savedID != id || kind < int32(IF_DSD_AND) || kind > int32(IF_DSD_PRIME) || fans < 2 || fans > 16 || words != 1+(fans+1)/2 || support > variables || kind == int32(IF_DSD_MUX) && fans != 3 || kind == int32(IF_DSD_PRIME) && fans < 3 {
			return nil, errors.New("invalid DSD node header")
		}
		o := If_DsdObjAlloc(p, kind, fans)
		o.nSupp, o.fMark, o.Count = support, int32(flags>>8&1), int32(flags>>9&0x3ffff)
		var actualSupport int32
		for i := range o.pFans {
			fan := int32(binary.LittleEndian.Uint32(record[8+4*i:]))
			if fan < 2 || fan>>1 >= id {
				return nil, errors.New("non-topological DSD fanin")
			}
			o.pFans[i] = fan
			actualSupport += If_DsdVecLitSuppSize(&p.vObjs, fan)
		}
		if actualSupport != support {
			return nil, errors.New("inconsistent DSD node support")
		}
		truthID := int32(-1)
		if o.Type == IF_DSD_PRIME {
			truthID = r.integer()
			if r.err != nil {
				return nil, r.err
			}
			if truthID < 0 {
				return nil, errors.New("invalid DSD truth identifier")
			}
			vec.Vec_IntWriteEntry(&p.vTruths, int(id), truthID)
		}
		spot := If_DsdObjHashLookup(p, kind, o.pFans, fans, truthID)
		if *spot != 0 {
			return nil, errors.New("duplicate canonical DSD node")
		}
		*spot = id
	}
	p.nUniqueMisses = 0
	for v := int32(3); v <= variables; v++ {
		words := dsdWordCount(v)
		records := r.count(8 * words)
		truth := make([]word, words)
		for i := 0; i < records && r.err == nil; i++ {
			for j := range truth {
				truth[j] = r.word()
			}
			if r.err == nil && vec.Vec_MemHashInsert(p.vTtMem[v], truth) != i {
				return nil, errors.New("duplicate DSD truth record")
			}
		}
		lists := r.count(4)
		if lists > records {
			return nil, errors.New("excess DSD decomposition records")
		}
		for i := 0; i < lists && r.err == nil; i++ {
			size := r.count(4)
			sets := vec.Vec_IntAlloc(size)
			for j := 0; j < size; j++ {
				Vec_IntPush(sets, r.integer())
			}
			Vec_PtrPush(p.vTtDecs[v], sets)
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	for id := int32(2); id < nodes; id++ {
		o := If_DsdVecObj(&p.vObjs, id)
		if o.Type == IF_DSD_PRIME && int(If_DsdObjTruthId(p, o)) >= vec.Vec_MemEntryNum(p.vTtMem[o.nFans]) {
			return nil, errors.New("missing DSD prime truth")
		}
	}
	p.nConfigWords, p.nTtBits = r.integer(), r.integer()
	if p.nConfigWords <= 0 || p.nTtBits < 0 || int64(p.nConfigWords)*int64(nodes) > math.MaxInt32 {
		return nil, errors.New("invalid DSD configuration dimensions")
	}
	configurations := r.count(8)
	if configurations > 0 {
		p.vConfigs = vec.Vec_WrdAlloc(configurations)
		for i := 0; i < configurations; i++ {
			vec.Vec_WrdPush(p.vConfigs, r.word())
		}
	}
	size := r.count(1)
	p.pCellStr = string(r.data(size))
	if r.err != nil {
		return nil, r.err
	}
	success = true
	return p, nil
}
func If_DsdManLoad(fileName string) *If_DsdMan_t {
	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Reading DSD manager file %q has failed.\n", fileName)
		return nil
	}
	p, err := ifDsdRead(data, fileName)
	if err == ifDsdBadFormat {
		fmt.Printf("Unrecognized format of file %q.\n", fileName)
		return nil
	}
	if err != nil {
		fmt.Printf("Reading DSD manager file %q has failed: %v.\n", fileName, err)
		return nil
	}
	return p
}
