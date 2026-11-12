package bbl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func Bbl_ManStart(name string) *Bbl_Man_t {
	return &Bbl_Man_t{Name: name, byID: map[int]*Bbl_Obj_t{}, handles: map[string]int{}}
}
func Bbl_ManStop(p *Bbl_Man_t) {
	if p != nil {
		*p = Bbl_Man_t{}
	}
}
func Bbl_ManName(p *Bbl_Man_t) string { return p.Name }
func Bbl_ManCreateObject(p *Bbl_Man_t, kind Bbl_Type_t, id, n int, sop string) error {
	if id < 0 || id > mathMaxInt32 || n < 0 || n > 0xfffffff {
		return fmt.Errorf("invalid object id/fanin count")
	}
	if _, exists := p.byID[id]; exists {
		return fmt.Errorf("duplicate object id %d", id)
	}
	if kind != BBL_OBJ_CI && kind != BBL_OBJ_CO && kind != BBL_OBJ_NODE {
		return fmt.Errorf("invalid object kind %d", kind)
	}
	if kind == BBL_OBJ_CI && n != 0 || kind == BBL_OBJ_CO && n != 1 {
		return fmt.Errorf("invalid fanin count for kind %d", kind)
	}
	if kind != BBL_OBJ_NODE && sop != "" {
		return fmt.Errorf("only nodes may have SOPs")
	}
	if sop != "" {
		if err := validateSOP(sop, n); err != nil {
			return err
		}
		sop = Bbl_ManSortSop(sop, n)
	}
	o := &Bbl_Obj_t{ID: id, OriginalID: id, Type: kind, SOP: sop, declared: n, index: len(p.Objects), function: -1}
	if sop != "" {
		o.function = Bbl_ManSaveSop(p, sop, n)
	}
	p.Objects = append(p.Objects, o)
	p.byID[id] = o
	return nil
}
func Bbl_ManAddFanin(p *Bbl_Man_t, id, fanin int) error {
	o, f := p.byID[id], p.byID[fanin]
	if o == nil || f == nil {
		return fmt.Errorf("unknown object/fanin %d/%d", id, fanin)
	}
	if o.Type == BBL_OBJ_CI || f.Type == BBL_OBJ_CO {
		return fmt.Errorf("invalid CI/CO edge")
	}
	if len(o.Fanins) >= o.declared {
		return fmt.Errorf("too many fanins for %d", id)
	}
	o.Fanins = append(o.Fanins, f)
	return nil
}
func (p *Bbl_Man_t) Check() error {
	members := map[*Bbl_Obj_t]bool{}
	for _, o := range p.Objects {
		if o == nil {
			return fmt.Errorf("nil object")
		}
		if members[o] {
			return fmt.Errorf("duplicate object pointer")
		}
		members[o] = true
	}
	for _, o := range p.Objects {
		if o.Type != BBL_OBJ_CI && o.Type != BBL_OBJ_CO && o.Type != BBL_OBJ_NODE {
			return fmt.Errorf("invalid object kind")
		}
		if len(o.Fanins) != o.declared {
			return fmt.Errorf("object %d: got %d fanins, want %d", o.ID, len(o.Fanins), o.declared)
		}
		if o.Type == BBL_OBJ_NODE {
			if err := validateSOP(o.SOP, o.declared); err != nil {
				return fmt.Errorf("node %d: %w", o.ID, err)
			}
		} else if o.SOP != "" {
			return fmt.Errorf("non-node with SOP")
		}
		if o.Type == BBL_OBJ_CI && o.declared != 0 || o.Type == BBL_OBJ_CO && o.declared != 1 {
			return fmt.Errorf("invalid CI/CO arity")
		}
		for _, f := range o.Fanins {
			if !members[f] || f.Type == BBL_OBJ_CO {
				return fmt.Errorf("invalid fanin of %d", o.ID)
			}
		}
	}
	return nil
}
func Bbl_ManCheck(p *Bbl_Man_t) int {
	if p.Check() != nil {
		return 0
	}
	return 1
}
func Bbl_ObjIsInput(o *Bbl_Obj_t) int                  { return boolean(o.Type == BBL_OBJ_CI) }
func Bbl_ObjIsOutput(o *Bbl_Obj_t) int                 { return boolean(o.Type == BBL_OBJ_CO) }
func Bbl_ObjIsLut(o *Bbl_Obj_t) int                    { return boolean(o.Type == BBL_OBJ_NODE) }
func Bbl_ObjId(o *Bbl_Obj_t) int                       { return o.ID }
func Bbl_ObjIdOriginal(_ *Bbl_Man_t, o *Bbl_Obj_t) int { return o.OriginalID }
func Bbl_ObjFaninNumber(o *Bbl_Obj_t) int              { return o.declared }
func Bbl_ObjSop(_ *Bbl_Man_t, o *Bbl_Obj_t) string     { return o.SOP }
func Bbl_ObjIsMarked(o *Bbl_Obj_t) int                 { return boolean(o.Marked) }
func Bbl_ObjMark(o *Bbl_Obj_t)                         { o.Marked = true }
func Bbl_ObjFncHandle(o *Bbl_Obj_t) int                { return o.function }
func Bbl_ManFncSize(p *Bbl_Man_t) int                  { return p.functionBytes }
func Bbl_ManObjFirst(p *Bbl_Man_t) *Bbl_Obj_t {
	if len(p.Objects) == 0 {
		return nil
	}
	return p.Objects[0]
}
func Bbl_ManObjNext(p *Bbl_Man_t, o *Bbl_Obj_t) *Bbl_Obj_t {
	if o.index+1 >= len(p.Objects) {
		return nil
	}
	return p.Objects[o.index+1]
}
func Bbl_ObjFaninFirst(o *Bbl_Obj_t) *Bbl_Obj_t {
	if len(o.Fanins) == 0 {
		return nil
	}
	return o.Fanins[0]
}
func Bbl_ObjFaninNext(o, previous *Bbl_Obj_t) *Bbl_Obj_t {
	for i, f := range o.Fanins {
		if f == previous {
			if i+1 < len(o.Fanins) {
				return o.Fanins[i+1]
			}
			return nil
		}
	}
	return nil
}
func boolean(v bool) int {
	if v {
		return 1
	}
	return 0
}

func validateSOP(s string, n int) error {
	if n < 0 || n > 0xfffffff || s == "" || len(s)%(n+3) != 0 {
		return fmt.Errorf("invalid SOP length")
	}
	polarity := byte(0)
	for start := 0; start < len(s); start += n + 3 {
		cube := s[start : start+n+3]
		for _, c := range []byte(cube[:n]) {
			if c != '0' && c != '1' && c != '-' {
				return fmt.Errorf("invalid cube literal")
			}
		}
		if cube[n] != ' ' || cube[n+2] != '\n' || cube[n+1] != '0' && cube[n+1] != '1' {
			return fmt.Errorf("invalid cube ending")
		}
		if polarity != 0 && polarity != cube[n+1] {
			return fmt.Errorf("mixed SOP polarities")
		}
		polarity = cube[n+1]
	}
	return nil
}
func Bbl_ManSortCubes(cubes []string, nCubes, nVars int) {
	sort.SliceStable(cubes[:nCubes], func(i, j int) bool { return cubes[i][:nVars] < cubes[j][:nVars] })
}
func Bbl_ManSortSop(s string, n int) string {
	if validateSOP(s, n) != nil {
		return s
	}
	cubes := make([]string, 0, len(s)/(n+3))
	for i := 0; i < len(s); i += n + 3 {
		cubes = append(cubes, s[i:i+n+3])
	}
	if len(cubes) < 300 {
		Bbl_ManSortCubes(cubes, len(cubes), n)
	}
	return strings.Join(cubes, "")
}
func Bbl_ManSaveSop(p *Bbl_Man_t, s string, n int) int {
	s = Bbl_ManSortSop(s, n)
	if h, ok := p.handles[s]; ok {
		return h
	}
	h := p.functionBytes
	p.functionBytes += 4 + ((len(s)+1+3)/4)*4
	p.handles[s] = h
	p.functions = append(p.functions, s)
	return h
}

// WriteBinary emits the original 32-bit little-endian BBLIF wire layout. No Go
// object layout, pointer value or machine-sized integer is written to disk.
func (p *Bbl_Man_t) WriteBinary(w io.Writer) error {
	if err := p.Check(); err != nil {
		return err
	}
	nameSize := 4 * ((len(p.Name)+1)/4 + 1)
	if nameSize < 16 {
		nameSize = 16
	}
	name := make([]byte, nameSize)
	copy(name, p.Name)
	offsets := map[*Bbl_Obj_t]int{}
	size := 0
	for _, o := range p.Objects {
		offsets[o] = size
		size += 12 + 4*len(o.Fanins)
	}
	objects := new(bytes.Buffer)
	for _, o := range p.Objects {
		flags := uint32(len(o.Fanins)) << 4
		if o.Type == BBL_OBJ_CI {
			flags |= 1
		}
		if o.Type == BBL_OBJ_CO {
			flags |= 2
		}
		if o.Box {
			flags |= 4
		}
		if o.Marked {
			flags |= 8
		}
		for _, v := range []uint32{uint32(o.OriginalID), uint32(o.function), flags} {
			if err := binary.Write(objects, binary.LittleEndian, v); err != nil {
				return err
			}
		}
		for _, f := range o.Fanins {
			binary.Write(objects, binary.LittleEndian, int32(offsets[o]-offsets[f]))
		}
	}
	functions := new(bytes.Buffer)
	for _, s := range p.functions {
		words := (len(s) + 4) / 4
		binary.Write(functions, binary.LittleEndian, int32(words))
		data := make([]byte, words*4)
		copy(data, s)
		functions.Write(data)
	}
	for _, pool := range [][]byte{name, objects.Bytes(), functions.Bytes()} {
		if len(pool) > mathMaxInt32 {
			return fmt.Errorf("BBLIF pool too large")
		}
		if err := binary.Write(w, binary.LittleEndian, int32(len(pool))); err != nil {
			return err
		}
		if err := writeAll(w, pool); err != nil {
			return err
		}
	}
	return nil
}

const mathMaxInt32 = 1<<31 - 1

func writeAll(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err == nil && n != len(b) {
		return io.ErrShortWrite
	}
	return err
}
func Bbl_ManDumpBinaryBlif(p *Bbl_Man_t, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = p.WriteBinary(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Bbl_ManReadBinaryBlif(path string) (*Bbl_Man_t, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBinary(data)
}
func ParseBinary(data []byte) (*Bbl_Man_t, error) {
	pools := make([][]byte, 3)
	cursor := 0
	for i := range pools {
		if len(data)-cursor < 4 {
			return nil, io.ErrUnexpectedEOF
		}
		n := int(int32(binary.LittleEndian.Uint32(data[cursor:])))
		cursor += 4
		if n < 0 || n > len(data)-cursor {
			return nil, fmt.Errorf("invalid BBLIF pool length")
		}
		pools[i] = data[cursor : cursor+n]
		cursor += n
	}
	if cursor != len(data) {
		return nil, fmt.Errorf("trailing BBLIF bytes")
	}
	name := pools[0]
	if i := bytes.IndexByte(name, 0); i >= 0 {
		name = name[:i]
	} else if len(name) > 0 {
		return nil, fmt.Errorf("unterminated design name")
	}
	p := Bbl_ManStart(string(name))
	functions := map[int]string{}
	for off := 0; off < len(pools[2]); {
		pool := pools[2]
		if len(pool)-off < 4 {
			return nil, io.ErrUnexpectedEOF
		}
		words := int(int32(binary.LittleEndian.Uint32(pool[off:])))
		if words <= 0 || words > (len(pool)-off-4)/4 {
			return nil, fmt.Errorf("invalid function size")
		}
		raw := pool[off+4 : off+4+words*4]
		end := bytes.IndexByte(raw, 0)
		if end < 0 {
			return nil, fmt.Errorf("unterminated SOP")
		}
		s := string(raw[:end])
		functions[off] = s
		p.handles[s] = off
		p.functions = append(p.functions, s)
		off += 4 + words*4
	}
	p.functionBytes = len(pools[2])
	objects := pools[1]
	offsets := map[int]*Bbl_Obj_t{}
	faninOffsets := map[*Bbl_Obj_t][]int{}
	for off := 0; off < len(objects); {
		if len(objects)-off < 12 {
			return nil, io.ErrUnexpectedEOF
		}
		original := int(int32(binary.LittleEndian.Uint32(objects[off:])))
		fn := int(int32(binary.LittleEndian.Uint32(objects[off+4:])))
		flags := binary.LittleEndian.Uint32(objects[off+8:])
		n := int(flags >> 4)
		if original < 0 || n > (len(objects)-off-12)/4 || flags&3 == 3 {
			return nil, fmt.Errorf("invalid object record")
		}
		if _, exists := p.byID[original]; exists {
			return nil, fmt.Errorf("duplicate original id")
		}
		kind := BBL_OBJ_NODE
		if flags&1 != 0 {
			kind = BBL_OBJ_CI
		}
		if flags&2 != 0 {
			kind = BBL_OBJ_CO
		}
		sop := ""
		if fn >= 0 {
			var ok bool
			sop, ok = functions[fn]
			if !ok {
				return nil, fmt.Errorf("invalid function handle")
			}
		} else if fn != -1 {
			return nil, fmt.Errorf("invalid function handle")
		}
		o := &Bbl_Obj_t{ID: len(p.Objects), OriginalID: original, Type: kind, SOP: sop, declared: n, index: len(p.Objects), function: fn, Box: flags&4 != 0, Marked: flags&8 != 0}
		p.Objects = append(p.Objects, o)
		p.byID[original] = o
		offsets[off] = o
		for i := 0; i < n; i++ {
			relative := int(int32(binary.LittleEndian.Uint32(objects[off+12+4*i:])))
			faninOffsets[o] = append(faninOffsets[o], off-relative)
		}
		off += 12 + 4*n
	}
	for o, refs := range faninOffsets {
		for _, off := range refs {
			f := offsets[off]
			if f == nil {
				return nil, fmt.Errorf("fanin offset is not an object")
			}
			o.Fanins = append(o.Fanins, f)
		}
	}
	if err := p.Check(); err != nil {
		return nil, err
	}
	return p, nil
}

func Bbl_ManTruthToSop(truth []uint32, n int) string {
	if n < 0 || n > 16 || len(truth) < ((1<<n)+31)/32 {
		return ""
	}
	count := 0
	for b := 0; b < 1<<n; b++ {
		if truth[b/32]>>(b%32)&1 != 0 {
			count++
		}
	}
	if count == 0 || count == 1<<n {
		return strings.Repeat("-", n) + fmt.Sprintf(" %d\n", boolean(count != 0))
	}
	var out strings.Builder
	for b := 0; b < 1<<n; b++ {
		if truth[b/32]>>(b%32)&1 == 0 {
			continue
		}
		for v := 0; v < n; v++ {
			out.WriteByte(byte('0' + (b>>v)&1))
		}
		out.WriteString(" 1\n")
	}
	return out.String()
}
func Bbl_ManSopToTruth(s string, n int) []uint32 {
	if n < 0 || n > 16 || validateSOP(s, n) != nil {
		return nil
	}
	count := 1 << n
	bits := count
	if bits < 32 {
		bits = 32
	}
	truth := make([]uint32, bits/32)
	complement := s[n+1] == '0'
	for c := 0; c < len(s); c += n + 3 {
		base, free := 0, 0
		for v := 0; v < n; v++ {
			switch s[c+v] {
			case '1':
				base |= 1 << v
			case '-':
				free |= 1 << v
			}
		}
		for subset := free; ; subset = (subset - 1) & free {
			assignment := base | subset
			for bit := assignment; bit < bits; bit += count {
				truth[bit/32] |= 1 << uint(bit%32)
			}
			if subset == 0 {
				break
			}
		}
	}
	if complement {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	return truth
}
func Bbl_ManTestTruth(s string, n int) bool {
	return equalTruth(Bbl_ManSopToTruth(s, n), Bbl_ManSopToTruth(Bbl_ManTruthToSop(Bbl_ManSopToTruth(s, n), n), n))
}
func equalTruth(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func (p *Bbl_Man_t) WriteBLIF(w io.Writer) error {
	if err := p.Check(); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Test file written by Bbl_ManDumpBlif() in ABC.\n.model %s\n", p.Name)
	for _, o := range p.Objects {
		switch o.Type {
		case BBL_OBJ_CI:
			fmt.Fprintf(&b, ".inputs %d\n", o.ID)
		case BBL_OBJ_CO:
			fmt.Fprintf(&b, ".outputs %d\n", o.ID)
		case BBL_OBJ_NODE:
			b.WriteString(".names")
			for _, f := range o.Fanins {
				fmt.Fprintf(&b, " %d", f.ID)
			}
			fmt.Fprintf(&b, " %d\n%s", o.ID, o.SOP)
		}
	}
	for _, o := range p.Objects {
		if o.Type == BBL_OBJ_CO {
			fmt.Fprintf(&b, ".names %d %d\n1 1\n", o.Fanins[0].ID, o.ID)
		}
	}
	b.WriteString(".end\n")
	return writeAll(w, []byte(b.String()))
}
func Bbl_ManDumpBlif(p *Bbl_Man_t, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = p.WriteBLIF(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Bbl_ManPrintStats(p *Bbl_Man_t) {
	nodes := 0
	for _, o := range p.Objects {
		if o.Type == BBL_OBJ_NODE {
			nodes++
		}
	}
	fmt.Printf("Total objects = %7d.  Total nodes = %7d. Unique functions = %7d.\n", len(p.Objects), nodes, len(p.functions))
}
func Bbl_ManFileSize(path string) (int64, error) {
	s, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return s.Size(), nil
}
func Bbl_ManFileRead(path string) ([]byte, error) { return os.ReadFile(path) }
func Bbl_ManSimpleDemo() *Bbl_Man_t {
	p := Bbl_ManStart("hadder")
	for _, id := range []int{1, 2, 3} {
		Bbl_ManCreateObject(p, BBL_OBJ_CI, id, 0, "")
	}
	Bbl_ManCreateObject(p, BBL_OBJ_CO, 4, 1, "")
	Bbl_ManCreateObject(p, BBL_OBJ_CO, 5, 1, "")
	Bbl_ManCreateObject(p, BBL_OBJ_NODE, 6, 3, "100 1\n010 1\n001 1\n111 1\n")
	Bbl_ManCreateObject(p, BBL_OBJ_NODE, 7, 3, "-11 1\n1-1 1\n11- 1\n")
	Bbl_ManAddFanin(p, 4, 6)
	Bbl_ManAddFanin(p, 5, 7)
	for _, id := range []int{6, 7} {
		for _, f := range []int{1, 2, 3} {
			Bbl_ManAddFanin(p, id, f)
		}
	}
	return p
}
