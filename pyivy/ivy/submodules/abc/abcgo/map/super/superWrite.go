package super

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func compareFloat(a, b float32) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Super_WriteCompare(a, b *Super_Gate_t) int {
	if a.Truth < b.Truth {
		return -1
	}
	if a.Truth > b.Truth {
		return 1
	}
	return 0
}
func Super_DelayCompare(a, b *Super_Gate_t) int { return compareFloat(a.MaxDelay, b.MaxDelay) }
func Super_AreaCompare(a, b *Super_Gate_t) int  { return compareFloat(a.Area, b.Area) }
func (p *Super_Man_t) selected() []*Super_Gate_t {
	var gates []*Super_Gate_t
	for _, group := range p.table {
		for _, g := range group {
			if g.Root == nil {
				continue
			}
			gap, valid := false, true
			for _, delay := range g.Delays {
				if delay < SUPER_NO_VAR+SUPER_EPSILON {
					gap = true
				} else if gap {
					valid = false
					break
				}
			}
			if valid {
				gates = append(gates, g)
			}
		}
	}
	sort.SliceStable(gates, func(i, j int) bool {
		a, b := gates[i], gates[j]
		if a.Truth != b.Truth {
			return a.Truth < b.Truth
		}
		if a.Area != b.Area {
			return a.Area < b.Area
		}
		if a.MaxDelay != b.MaxDelay {
			return a.MaxDelay < b.MaxDelay
		}
		return a.serial < b.serial
	})
	return gates
}
func writeHeader(p *Super_Man_t, w io.Writer, gates []*Super_Gate_t) error {
	functions := 0
	var previous uint64
	for index, g := range gates {
		if index == 0 || g.Truth != previous {
			functions++
		}
		previous = g.Truth
	}
	name := p.Name
	if name == "" {
		name = "unnamed.genlib"
	}
	if strings.ContainsAny(name, "\r\n") {
		return fmt.Errorf("super: library name contains newline")
	}
	_, err := fmt.Fprintf(w, "# Supergate library derived for %q.\n# Inputs=%d Levels=%d Attempts=%d Functions=%d\n%s\n%d\n%d\n", name, p.Options.Inputs, p.Levels, p.Stats.Tried, functions, name, p.Options.Inputs, len(gates))
	return err
}
func Super_WriteFileHeader(p *Super_Man_t, w io.Writer) error { return writeHeader(p, w, p.selected()) }
func Super_WriteFileHeaderStr(p *Super_Man_t, v *Vec_Str_t) error {
	var b bytes.Buffer
	if err := Super_WriteFileHeader(p, &b); err != nil {
		return err
	}
	vec.Vec_StrPushBuffer(v, b.Bytes(), b.Len())
	return nil
}

// MarshalTree numbers fanins before users and writes each shared subgate once.
// Numbering/marking are local, so repeated serialization cannot omit old nodes.
func (p *Super_Man_t) MarshalTree() ([]byte, error) {
	if p == nil || p.Options.Inputs < 1 || p.Options.Inputs > 6 {
		return nil, fmt.Errorf("super: invalid writer manager")
	}
	gates := p.selected()
	selected := make(map[*Super_Gate_t]bool)
	for _, g := range gates {
		selected[g] = true
	}
	numbers := make(map[*Super_Gate_t]int)
	active := make(map[*Super_Gate_t]bool)
	counter := p.Options.Inputs
	var body bytes.Buffer
	var write func(*Super_Gate_t) error
	write = func(g *Super_Gate_t) error {
		if g == nil {
			return fmt.Errorf("super: nil gate in tree")
		}
		if _, ok := numbers[g]; ok {
			return nil
		}
		if g.Root == nil {
			if g.Variable < 0 || g.Variable >= p.Options.Inputs {
				return fmt.Errorf("super: invalid variable")
			}
			numbers[g] = g.Variable
			return nil
		}
		if active[g] {
			return fmt.Errorf("super: cycle in serialization graph")
		}
		active[g] = true
		if len(g.Fanins) != int(mio.Mio_GateReadPinNum(g.Root)) {
			return fmt.Errorf("super: inconsistent gate arity")
		}
		for _, child := range g.Fanins {
			if err := write(child); err != nil {
				return err
			}
		}
		numbers[g] = counter
		counter++
		if selected[g] {
			body.WriteString("* ")
		}
		body.WriteString(mio.Mio_GateReadName(g.Root))
		for _, child := range g.Fanins {
			fmt.Fprintf(&body, " %d", numbers[child])
		}
		body.WriteByte('\n')
		delete(active, g)
		return nil
	}
	for _, g := range gates {
		if err := write(g); err != nil {
			return nil, err
		}
	}
	var result bytes.Buffer
	if err := writeHeader(p, &result, gates); err != nil {
		return nil, err
	}
	fmt.Fprintf(&result, "%d\n", counter)
	result.Write(body.Bytes())
	return result.Bytes(), nil
}
func Super_Write(p *Super_Man_t) *Vec_Str_t {
	data, err := p.MarshalTree()
	if err != nil {
		return nil
	}
	p.Gates = p.selected()
	return vec.Vec_StrAllocArrayCopy(append(data, 0), len(data)+1)
}
func Super_WriteLibraryTreeStr(p *Super_Man_t) *Vec_Str_t { return Super_Write(p) }
func outputBase(name string) string {
	if name == "" {
		return "unnamed"
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}
func Super_WriteLibraryTreeFile(p *Super_Man_t) error {
	data, err := p.MarshalTree()
	if err != nil {
		return err
	}
	return os.WriteFile(outputBase(p.Name)+".super", data, 0644)
}
func Super_WriteLibraryTree(p *Super_Man_t) error { return Super_WriteLibraryTreeFile(p) }
func Super_WriteLibraryGateName(g *Super_Gate_t) string {
	if g.Root == nil {
		return string(rune('a' + g.Variable))
	}
	children := make([]string, len(g.Fanins))
	for i, child := range g.Fanins {
		children[i] = Super_WriteLibraryGateName(child)
	}
	return mio.Mio_GateReadName(g.Root) + "(" + strings.Join(children, ",") + ")"
}
func Super_WriteLibraryGateName_rec(g *Super_Gate_t, b *strings.Builder) {
	b.WriteString(Super_WriteLibraryGateName(g))
}
func Super_WriteLibraryGate(w io.Writer, p *Super_Man_t, g *Super_Gate_t, number int32) error {
	if _, err := fmt.Fprintf(w, "%04d  %0*b   %5.2f  ", number, 1<<uint(p.Options.Inputs), g.Truth, g.MaxDelay); err != nil {
		return err
	}
	for _, delay := range g.Delays {
		if delay == SUPER_NO_VAR {
			delay = 0
		}
		if _, err := fmt.Fprintf(w, " %5.2f", delay); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "   %5.2f   %s\n", g.Area, Super_WriteLibraryGateName(g))
	return err
}
func (p *Super_Man_t) WriteOld(w io.Writer) error {
	gates := p.selected()
	if err := writeHeader(p, w, gates); err != nil {
		return err
	}
	for index, g := range gates {
		if err := Super_WriteLibraryGate(w, p, g, int32(index+1)); err != nil {
			return err
		}
	}
	return nil
}
func Super_WriteLibrary(p *Super_Man_t) error {
	var data bytes.Buffer
	if err := p.WriteOld(&data); err != nil {
		return err
	}
	return os.WriteFile(outputBase(p.Name)+".super_old", data.Bytes(), 0644)
}
