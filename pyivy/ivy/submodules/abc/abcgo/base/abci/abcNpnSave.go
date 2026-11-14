package abci

import (
	"bufio"
	"fmt"
	"math/bits"
	"os"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// word is ABC's 64-bit truth-table word.
type word = uint64

type Npn_Obj_t = Npn_Obj_t_
type Npn_Man_t = Npn_Man_t_
type Npn_Obj_t_ struct {
	uTruth       word
	Count, iNext int32
}
type Npn_Man_t_ struct {
	pBuffer                      []Npn_Obj_t
	pBins                        []int32
	nBins, nBufferSize, nEntries int32
}

var Truth = [8]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000, 0, ^word(0)}
var pNpnMan *Npn_Man_t
var spStatic_Npn_TruthPermute_rec_count int32

func Npn_ManObj(p *Npn_Man_t, i int32) *Npn_Obj_t {
	if p == nil || i <= 0 || i >= int32(len(p.pBuffer)) {
		return nil
	}
	return &p.pBuffer[i]
}
func Npn_ManObjNum(p *Npn_Man_t, obj *Npn_Obj_t) int32 {
	if p == nil || obj == nil {
		return 0
	}
	for i := 1; i < len(p.pBuffer); i++ {
		if &p.pBuffer[i] == obj {
			return int32(i)
		}
	}
	return 0
}

func Npn_TruthPermute_rec(s string, mid, end int32) {
	b := []byte(s)
	var rec func(int)
	rec = func(k int) {
		if k == int(end) {
			fmt.Printf("%03d: %s\n", spStatic_Npn_TruthPermute_rec_count, string(b))
			spStatic_Npn_TruthPermute_rec_count++
			return
		}
		for i := k; i <= int(end); i++ {
			b[k], b[i] = b[i], b[k]
			rec(k + 1)
			b[k], b[i] = b[i], b[k]
		}
	}
	if mid < 0 || end < mid || int(end) >= len(b) {
		return
	}
	rec(int(mid))
}
func Npn_TruthHasVar(t word, v int32) int32 {
	if v < 0 || v >= 6 {
		return 0
	}
	shift := uint(1 << uint(v))
	return abcBool32(((t & Truth[v]) >> shift) != (t &^ Truth[v]))
}
func Npn_TruthSupport(t word) int32 {
	var s int32
	for v := int32(0); v < 6; v++ {
		if Npn_TruthHasVar(t, v) != 0 {
			s |= 1 << uint(v)
		}
	}
	return s
}
func Npn_TruthSuppSize(t word, nVars int32) int32 {
	if nVars > 6 {
		nVars = 6
	}
	var n int32
	for v := int32(0); v < nVars; v++ {
		n += Npn_TruthHasVar(t, v)
	}
	return n
}
func Npn_TruthIsMinBase(t word) int32 { s := Npn_TruthSupport(t); return abcBool32(s&(s+1) == 0) }
func Npn_TruthPadWord(t word, nVars int32) word {
	if nVars < 0 {
		nVars = 0
	}
	if nVars >= 6 {
		return t
	}
	period := uint(1) << uint(nVars)
	mask := word(1)<<period - 1
	base := t & mask
	var out word
	for i := uint(0); i < 64; i++ {
		out |= ((base >> (i % period)) & 1) << i
	}
	return out
}
func Npn_TruthCountOnes(t word) int32 { return int32(bits.OnesCount64(t)) }
func Npn_TruthChangePhase(t word, v int32) word {
	if v < 0 || v >= 6 {
		return t
	}
	step := uint(1) << uint(v)
	m := Truth[v]
	return ((t & m) >> step) | ((t &^ m) << step)
}
func Npn_TruthSwapAdjacentVars(t word, v int32) word {
	if v < 0 || v >= 5 {
		return t
	}
	step := uint(1) << uint(v)
	var out word
	for i := uint(0); i < 64; i++ {
		j := i
		a := (i >> uint(v)) & 1
		b := (i >> uint(v+1)) & 1
		if a != b {
			j ^= step | (step << 1)
		}
		out |= ((t >> i) & 1) << j
	}
	return out
}
func Npn_TruthCanon(t word, nVars int32, pPhase *int32) word {
	if nVars < 0 || nVars > 6 {
		return t
	}
	phase := int32(0)
	if Npn_TruthCountOnes(t) > 32 {
		t = ^t
		phase = 1 << 6
	}
	if pPhase != nil {
		*pPhase = phase
	}
	return t
}

func Npn_ManHash(p *Npn_Man_t, t word) int32 {
	if p == nil || p.nBins <= 0 {
		return 0
	}
	k := (t * 101) ^ (t * 733) ^ (t * 1777)
	return int32(k % uint64(p.nBins))
}
func Npn_ManResize(p *Npn_Man_t) {
	if p == nil {
		return
	}
	old := p.pBins
	p.nBins = primeAtLeast(3 * max32(1, p.nBins))
	p.pBins = make([]int32, p.nBins)
	for i := int32(1); i < p.nEntries; i++ {
		o := &p.pBuffer[i]
		k := Npn_ManHash(p, o.uTruth)
		o.iNext = p.pBins[k]
		p.pBins[k] = int32(i)
	}
	_ = old
}
func Npn_ManAdd(p *Npn_Man_t, t word) *Npn_Obj_t {
	if p == nil {
		return nil
	}
	if p.nBins == 0 {
		p.nBins = 101
		p.pBins = make([]int32, p.nBins)
	}
	k := Npn_ManHash(p, t)
	for id := p.pBins[k]; id != 0; id = p.pBuffer[id].iNext {
		o := &p.pBuffer[id]
		if o.uTruth == t {
			o.Count++
			return o
		}
	}
	if p.nEntries >= int32(len(p.pBuffer)) {
		p.pBuffer = append(p.pBuffer, make([]Npn_Obj_t, max32(1, int32(len(p.pBuffer))))...)
	}
	id := p.nEntries
	if id == 0 {
		id = 1
		p.nEntries = 1
	}
	o := &p.pBuffer[id]
	o.uTruth = t
	o.Count = 1
	o.iNext = p.pBins[k]
	p.pBins[k] = id
	p.nEntries = id + 1
	p.nBufferSize = int32(len(p.pBuffer))
	if p.nEntries > 3*p.nBins {
		Npn_ManResize(p)
	}
	return o
}
func Npn_ManRead(p *Npn_Man_t, name string) {
	if p == nil {
		return
	}
	f, e := os.Open(name)
	if e != nil {
		Abc_Print(-1, "Cannot open NPN function file %q.\n", name)
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024), 1<<20)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if len(fields[0]) != 16 {
			Abc_Print(0, "Skipping token %s that does not look like a 16-digit hex number.\n", fields[0])
			continue
		}
		u, e := strconv.ParseUint(fields[0], 16, 64)
		if e != nil {
			continue
		}
		o := Npn_ManAdd(p, word(u))
		if len(fields) > 1 {
			if n, err := strconv.Atoi(fields[1]); err == nil {
				o.Count = int32(n)
			}
		}
	}
}
func Npn_ManCompareEntries(a, b *Npn_Obj_t) int32 {
	if a.Count > b.Count {
		return -1
	}
	if a.Count < b.Count {
		return 1
	}
	return 0
}
func Npn_ManWrite(p *Npn_Man_t, name string) {
	if p == nil {
		return
	}
	f, e := os.Create(name)
	if e != nil {
		Abc_Print(-1, "Cannot open NPN function file %q.\n", name)
		return
	}
	defer f.Close()
	entries := make([]*Npn_Obj_t, 0, p.nEntries)
	for i := int32(1); i < p.nEntries; i++ {
		entries = append(entries, &p.pBuffer[i])
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Count > entries[j].Count })
	for _, o := range entries {
		fmt.Fprintf(f, "%016x %d %d\n", o.uTruth, o.Count, Npn_TruthSuppSize(o.uTruth, 6))
	}
}
func Npn_ManStart(name string) *Npn_Man_t {
	p := &Npn_Man_t{nBufferSize: 100, nEntries: 1, nBins: 53}
	p.pBuffer = make([]Npn_Obj_t, p.nBufferSize)
	p.pBins = make([]int32, p.nBins)
	if name != "" {
		if _, e := os.Stat(name); e != nil {
			Abc_Print(-1, "Cannot open NPN function file %q.\n", name)
			return nil
		}
		Npn_ManRead(p, name)
	}
	return p
}
func Npn_ManStop(p *Npn_Man_t) {
	if p != nil {
		p.pBuffer = nil
		p.pBins = nil
		p.nEntries = 0
	}
}
func Npn_ManClean() { Npn_ManStop(pNpnMan); pNpnMan = nil }
func Npn_ManLoad(name string) {
	if pNpnMan != nil {
		Abc_Print(1, "Removing old table with %d entries.\n", pNpnMan.nEntries)
		Npn_ManStop(pNpnMan)
	}
	pNpnMan = Npn_ManStart(name)
	if pNpnMan != nil {
		Abc_Print(1, "Created new table with %d entries from file %q.\n", pNpnMan.nEntries, name)
	}
}
func Npn_ManSave(name string) {
	if pNpnMan == nil {
		Abc_Print(1, "There is no table with entries.\n")
		return
	}
	Npn_ManWrite(pNpnMan, name)
	Abc_Print(1, "Dumped table with %d entries from file %q.\n", pNpnMan.nEntries, name)
}
func Npn_ManSaveOne(truth *uint32, nVars int32) {
	if truth == nil || nVars < 0 || nVars > 6 {
		return
	}
	words := unsafe.Slice(truth, 2)
	t := word(words[0]) | word(words[1])<<32
	if pNpnMan == nil {
		Abc_Print(1, "Creating new table with 0 entries.\n")
		pNpnMan = Npn_ManStart("")
	}
	if Npn_TruthIsMinBase(t) == 0 {
		return
	}
	t = Npn_TruthPadWord(t, nVars)
	t = Npn_TruthCanon(t, 6, nil)
	Npn_ManAdd(pNpnMan, t)
}
func abcBool32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func primeAtLeast(n int32) int32 {
	for x := n; x > 1; x++ {
		prime := true
		for d := int32(2); d*d <= x; d++ {
			if x%d == 0 {
				prime = false
				break
			}
		}
		if prime {
			return x
		}
	}
	return 2
}
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
