package cut

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"io"
	"math/bits"
	"os"
	"strings"
)

const CUT_CELL_MVAR int32 = 9

type Cut_Cell_t struct {
	NextVar, Parent      *Cut_Cell_t
	Used                 int32
	Box                  [4]byte
	NVars                int32
	CrossBar0, CrossBar1 int32
	CrossBarPhase        byte
	CanonPhase           uint32
	CanonPerm            [12]byte
	Store                [18]int32
	Truth                [16]uint32
}
type Cut_CMan_t struct {
	table                                map[string]*Cut_Cell_t
	sameVar                              [10]*Cut_Cell_t
	Total, Good, Found, NotFound         int64
	VarCounts, SymGroups, SymGroupsEqual [10]int64
}

var s_pCMan *Cut_CMan_t

func Cut_CManStart() *Cut_CMan_t {
	p := &Cut_CMan_t{table: map[string]*Cut_Cell_t{}}
	s_pCMan = p
	return p
}
func Cut_CManStop(p *Cut_CMan_t) {
	if p != nil {
		p.table = nil
		p.sameVar = [10]*Cut_Cell_t{}
	}
	if s_pCMan == p {
		s_pCMan = nil
	}
}
func Cut_CellIsRunning() int32 {
	if s_pCMan != nil {
		return 1
	}
	return 0
}
func cellKey(c *Cut_Cell_t) string {
	buffer := make([]byte, 1+4*int(Cut_TruthWords(c.NVars)))
	buffer[0] = byte(c.NVars)
	for i, w := range c.Truth[:Cut_TruthWords(c.NVars)] {
		binary.LittleEndian.PutUint32(buffer[1+4*i:], w)
	}
	return string(buffer)
}
func Cut_CellTableLookup(p *Cut_CMan_t, c *Cut_Cell_t) int32 {
	key := cellKey(c)
	if p.table[key] != nil {
		return 1
	}
	p.table[key] = c
	c.NextVar = p.sameVar[c.NVars]
	p.sameVar[c.NVars] = c
	return 0
}
func Cut_CellSuppMin(c *Cut_Cell_t) {
	n := c.NVars
	if n < 0 || n > CUT_CELL_MVAR {
		panic("cut: cell support out of bounds")
	}
	support := kit.Kit_TruthSupport(c.Truth[:], n)
	count := int32(bits.OnesCount32(support))
	if count == n {
		return
	}
	out := make([]uint32, Cut_TruthWords(n))
	kit.Kit_TruthShrink(out, c.Truth[:], count, n, support, 0)
	copy(c.Truth[:], out)
	var permutation [12]byte
	k := 0
	for v := int32(0); v < n; v++ {
		if support&(1<<uint(v)) != 0 {
			permutation[k] = c.CanonPerm[v]
			k++
		}
	}
	copy(c.CanonPerm[:], permutation[:])
	c.NVars = count
}

// Cell semi-canonicalization follows Extra's ascending cofactor-count ordering,
// which differs from Kit's descending convention. This preserves library files.
func Cut_CellCanonicize(p *Cut_CMan_t, c *Cut_Cell_t) {
	n := c.NVars
	words := Cut_TruthWords(n)
	truth := c.Truth[:words]
	var phase uint32
	ones := kit.Kit_TruthCountOnes(truth, n)
	if ones > words*16 || ones == words*16 && truth[0]&1 != 0 {
		phase |= 1 << uint(n)
		kit.Kit_TruthNot(truth, truth, n)
	}
	store := c.Store[:2*n]
	kit.Kit_TruthCountOnesInCofs(truth, n, store)
	for v := int32(0); v < n; v++ {
		if store[2*v] > store[2*v+1] {
			phase |= 1 << uint(v)
			store[2*v], store[2*v+1] = store[2*v+1], store[2*v]
			kit.Kit_TruthChangePhase(truth, n, v)
		}
	}
	aux := make([]uint32, words)
	for changed := true; changed; {
		changed = false
		for v := int32(0); v < n-1; v++ {
			if store[2*v] <= store[2*(v+1)] {
				continue
			}
			changed = true
			c.CanonPerm[v], c.CanonPerm[v+1] = c.CanonPerm[v+1], c.CanonPerm[v]
			store[2*v], store[2*v+2] = store[2*v+2], store[2*v]
			store[2*v+1], store[2*v+3] = store[2*v+3], store[2*v+1]
			kit.Kit_TruthSwapAdjacentVars(aux, truth, n, v)
			copy(truth, aux)
		}
	}
	c.CanonPhase = phase
}
func Cut_CellCrossBar(c *Cut_Cell_t) {
	if c.CrossBar0 < 0 || c.CrossBar1 <= c.CrossBar0 || c.CrossBar1 >= c.NVars || c.CrossBarPhase > 2 {
		panic("cut: invalid crossbar")
	}
	input := c.Truth
	clear(c.Truth[:])
	for assignment := 0; assignment < int(Cut_TruthWords(c.NVars))*32; assignment++ {
		value := assignment >> uint(c.CrossBar0) & 1
		x, y := value, value
		if c.CrossBarPhase == 1 {
			x ^= 1
		}
		if c.CrossBarPhase == 2 {
			y ^= 1
		}
		index := assignment&^(1<<uint(c.CrossBar0)|1<<uint(c.CrossBar1)) | x<<uint(c.CrossBar0) | y<<uint(c.CrossBar1)
		if input[index>>5]&(1<<uint(index&31)) != 0 {
			c.Truth[assignment>>5] |= 1 << uint(assignment&31)
		}
	}
}

func Cut_CellTruthElem(inA, inB, inC, out []uint32, n, kind int32) {
	if kind < 0 || kind >= 22 {
		panic("cut: invalid cell type")
	}
	for i := int32(0); i < Cut_TruthWords(n); i++ {
		a, b, c := inA[i], inB[i], inC[i]
		_, _, _ = a, b, c
		switch kind {
		case 0:
			out[i] = 0
		case 1:
			out[i] = 0xFFFFFFFF
		case 2:
			out[i] = a
		case 3:
			out[i] = a & b
		case 4:
			out[i] = ^(a & b)
		case 5:
			out[i] = a ^ b
		case 6:
			out[i] = a & b & c
		case 7:
			out[i] = ^(a & b & c)
		case 8:
			out[i] = a & (b | c)
		case 9:
			out[i] = ^(a & (b | c))
		case 10:
			out[i] = a ^ b ^ c
		case 11:
			out[i] = a ^ (b & c)
		case 12:
			out[i] = a & (b ^ c)
		case 13:
			out[i] = ^(a & (b ^ c))
		case 14:
			out[i] = (a & b) | (b & c) | (a & c)
		case 15:
			out[i] = (a & b & c) | (^a & ^b & ^c)
		case 16:
			out[i] = ^((a & b & c) | (^a & ^b & ^c))
		case 17:
			out[i] = (a & b) | (b & c) | (^a & c)
		case 18:
			out[i] = (^a & b & c) | (a & ^b & c) | (a & b & ^c)
		case 19:
			out[i] = ^((^a & b & c) | (a & ^b & c) | (a & b & ^c))
		case 20:
			out[i] = (a & ^b & ^c) | (b & c)
		case 21:
			out[i] = ^((a & ^b & ^c) | (b & c))
		}
	}
}

func cellRecord(p *Cut_CMan_t, c *Cut_Cell_t) {
	p.Total++
	if Cut_CellTableLookup(p, c) != 0 {
		return
	}
	p.Good++
	p.VarCounts[c.NVars]++
	for i := int32(0); i < c.NVars-1; i++ {
		j := i + 1
		for j < c.NVars && c.Store[2*i] == c.Store[2*j] {
			j++
		}
		if j-i > 1 {
			if c.Store[2*i] == c.Store[2*i+1] {
				p.SymGroupsEqual[j-i]++
			} else {
				p.SymGroups[j-i]++
			}
		}
		i = j - 1
	}
}
func PrecomputeCellLibrary(rootTypes, bottomTypes []int32, crossbars bool) *Cut_CMan_t {
	p := Cut_CManStart()
	inputs := make([][]uint32, 9)
	for i := range inputs {
		inputs[i] = make([]uint32, 16)
		kit.Kit_TruthIthVar(inputs[i], 9, int32(i))
	}
	bottom := make([][3][]uint32, len(bottomTypes))
	for i, kind := range bottomTypes {
		for group := 0; group < 3; group++ {
			bottom[i][group] = make([]uint32, 16)
			Cut_CellTruthElem(inputs[3*group], inputs[3*group+1], inputs[3*group+2], bottom[i][group], 9, kind)
		}
	}
	for _, root := range rootTypes {
		for i, a := range bottom {
			for j, b := range bottom {
				for k, c := range bottom {
					cell := &Cut_Cell_t{NVars: 9, Box: [4]byte{byte(root), byte(bottomTypes[i]), byte(bottomTypes[j]), byte(bottomTypes[k])}}
					for v := range cell.CanonPerm {
						cell.CanonPerm[v] = byte(v)
					}
					Cut_CellTruthElem(a[0], b[1], c[2], cell.Truth[:], 9, root)
					Cut_CellSuppMin(cell)
					Cut_CellCanonicize(p, cell)
					cellRecord(p, cell)
				}
			}
		}
	}
	if crossbars {
		for n := int32(9); n > 3; n-- {
			for parent := p.sameVar[n]; parent != nil; parent = parent.NextVar {
				for a := int32(0); a < n; a++ {
					for b := a + 1; b < n; b++ {
						for phase := byte(0); phase < 3; phase++ {
							cell := &Cut_Cell_t{NVars: n, Truth: parent.Truth, Parent: parent, CrossBar0: a, CrossBar1: b, CrossBarPhase: phase}
							for v := range cell.CanonPerm {
								cell.CanonPerm[v] = byte(v)
							}
							Cut_CellCrossBar(cell)
							Cut_CellSuppMin(cell)
							Cut_CellCanonicize(p, cell)
							cellRecord(p, cell)
						}
					}
				}
			}
		}
	}
	return p
}
func Cut_CellPrecompute() {
	PrecomputeCellLibrary([]int32{6, 8, 10, 11, 12, 14, 15, 17, 18, 20}, []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, true)
}
func LoadCellLibrary(reader io.Reader) (*Cut_CMan_t, error) {
	p := Cut_CManStart()
	scanner := bufio.NewScanner(reader)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		nBits := len(text) * 4
		n := int32(bits.Len(uint(nBits - 1)))
		if n > 9 || len(text) > 128 {
			Cut_CManStop(p)
			return nil, fmt.Errorf("cut: cell line %d exceeds nine variables", line)
		}
		cell := &Cut_Cell_t{NVars: n, Used: 1}
		for i := range cell.CanonPerm {
			cell.CanonPerm[i] = byte(i)
		}
		for i, c := range []byte(text) {
			var digit byte
			switch {
			case c >= '0' && c <= '9':
				digit = c - '0'
			case c >= 'a' && c <= 'f':
				digit = c - 'a' + 10
			case c >= 'A' && c <= 'F':
				digit = c - 'A' + 10
			default:
				Cut_CManStop(p)
				return nil, fmt.Errorf("cut: invalid hex at cell line %d", line)
			}
			pos := len(text) - 1 - i
			cell.Truth[pos/8] |= uint32(digit) << uint(4*(pos%8))
		}
		if n < 5 {
			for shift := uint(1) << uint(n); shift < 32; shift *= 2 {
				cell.Truth[0] |= cell.Truth[0] << shift
			}
		}
		Cut_CellSuppMin(cell)
		Cut_CellCanonicize(p, cell)
		cellRecord(p, cell)
	}
	if err := scanner.Err(); err != nil {
		Cut_CManStop(p)
		return nil, err
	}
	return p, nil
}
func Cut_CellLoad() {
	file, err := os.Open("cells22_daomap_iwls.txt")
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		return
	}
	defer file.Close()
	if _, err := LoadCellLibrary(file); err != nil {
		fmt.Fprintln(os.Stdout, err)
	}
}
func DumpCellLibrary(p *Cut_CMan_t, writer io.Writer) error {
	if p == nil {
		return fmt.Errorf("cut: cell manager not defined")
	}
	for n := int32(0); n <= 9; n++ {
		for cell := p.sameVar[n]; cell != nil; cell = cell.NextVar {
			if cell.Used == 0 {
				continue
			}
			digits := int32(8)
			if n > 5 {
				digits = 1 << (n - 2)
			}
			for i := digits - 1; i >= 0; i-- {
				if _, err := fmt.Fprintf(writer, "%x", cell.Truth[i/8]>>uint(4*(i%8))&15); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(writer); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
	}
	return nil
}
func Cut_CellDumpToFile() {
	file, err := os.Create("celllib22.txt")
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		return
	}
	err = DumpCellLibrary(s_pCMan, file)
	closeErr := file.Close()
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
	} else if closeErr != nil {
		fmt.Fprintln(os.Stdout, closeErr)
	}
}
func Cut_CellTruthLookup(truth []uint32, n int32) int32 {
	p := s_pCMan
	if p == nil {
		return 0
	}
	if n < 0 || n > 9 {
		return 0
	}
	cell := &Cut_Cell_t{NVars: n}
	copy(cell.Truth[:], truth[:Cut_TruthWords(n)])
	for v := range cell.CanonPerm {
		cell.CanonPerm[v] = byte(v)
	}
	Cut_CellSuppMin(cell)
	Cut_CellCanonicize(p, cell)
	if old := p.table[cellKey(cell)]; old != nil {
		old.Used++
		p.Found++
		return 1
	}
	p.NotFound++
	return 0
}
