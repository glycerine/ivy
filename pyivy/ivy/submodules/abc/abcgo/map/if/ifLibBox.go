// Combinational/sequential box libraries from ifLibBox.c.
package if_

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

const BoxNoArc int32 = -1000000000

func If_BoxStart(name string, id, inputs, outputs, seq, black, outer int32) *If_Box_t {
	if name == "" || strings.ContainsAny(name, " \t\r\n#") || id < 0 || inputs < 0 || outputs < 0 || int64(inputs)*int64(outputs) > math.MaxInt32 {
		return nil
	}
	flag := func(v int32) byte {
		if v != 0 {
			return 1
		}
		return 0
	}
	return &If_Box_t{pName: name, Id: id, nPis: inputs, nPos: outputs, fSeq: flag(seq), fBlack: flag(black), fOuter: flag(outer), pDelays: make([]int32, int(inputs)*int(outputs))}
}
func If_BoxDup(p *If_Box_t) *If_Box_t {
	if p == nil {
		return nil
	}
	copy := *p
	copy.pDelays = append([]int32(nil), p.pDelays...)
	return &copy
}
func If_BoxFree(p *If_Box_t) {
	if p != nil {
		*p = If_Box_t{}
	}
}
func (p *If_Box_t) Name() string   { return p.pName }
func (p *If_Box_t) ID() int32      { return p.Id }
func (p *If_Box_t) Inputs() int32  { return p.nPis }
func (p *If_Box_t) Outputs() int32 { return p.nPos }
func (p *If_Box_t) Delay(output, input int32) int32 {
	if output < 0 || output >= p.nPos || input < 0 || input >= p.nPis {
		panic("box pin outside delay matrix")
	}
	return p.pDelays[output*p.nPis+input]
}
func (p *If_Box_t) SetDelay(output, input, delay int32) {
	if output < 0 || output >= p.nPos || input < 0 || input >= p.nPis {
		panic("box pin outside delay matrix")
	}
	p.pDelays[output*p.nPis+input] = delay
}
func If_LibBoxStart() *If_LibBox_t { return &If_LibBox_t{vBoxes: vec.Vec_PtrAlloc(100)} }
func (p *If_LibBox_t_) Boxes() []*If_Box_t {
	if p == nil || p.vBoxes == nil {
		return nil
	}
	var result []*If_Box_t
	for _, value := range vec.Vec_PtrArray(p.vBoxes) {
		if value != nil {
			result = append(result, value.(*If_Box_t))
		}
	}
	return result
}
func If_LibBoxDup(p *If_LibBox_t) *If_LibBox_t {
	if p == nil {
		return nil
	}
	copy := If_LibBoxStart()
	for _, box := range p.Boxes() {
		if err := If_LibBoxAdd(copy, If_BoxDup(box)); err != nil {
			return nil
		}
	}
	return copy
}
func If_LibBoxFree(p *If_LibBox_t) {
	if p == nil {
		return
	}
	for _, box := range p.Boxes() {
		If_BoxFree(box)
	}
	*p = If_LibBox_t{}
}
func (p *If_LibBox_t_) Close() error { If_LibBoxFree(p); return nil }
func If_LibBoxReadBox(p *If_LibBox_t, id int32) *If_Box_t {
	if p == nil || p.vBoxes == nil || id < 0 || int(id) >= vec.Vec_PtrSize(p.vBoxes) {
		return nil
	}
	box, _ := vec.Vec_PtrEntry(p.vBoxes, int(id)).(*If_Box_t)
	return box
}
func If_LibBoxFindBox(p *If_LibBox_t, name string) *If_Box_t {
	for _, box := range p.Boxes() {
		if box.pName == name {
			return box
		}
	}
	return nil
}
func If_LibBoxAdd(p *If_LibBox_t, box *If_Box_t) error {
	if p == nil || box == nil || box.Id < 0 || box.nPis < 0 || box.nPos < 0 || int64(box.nPis)*int64(box.nPos) != int64(len(box.pDelays)) {
		return fmt.Errorf("invalid box or delay matrix")
	}
	if p.vBoxes == nil {
		p.vBoxes = vec.Vec_PtrAlloc(0)
	}
	if If_LibBoxReadBox(p, box.Id) != nil || If_LibBoxFindBox(p, box.pName) != nil {
		return fmt.Errorf("duplicate box ID or name: %d/%s", box.Id, box.pName)
	}
	vec.Vec_PtrFillExtra(p.vBoxes, int(box.Id)+1, nil)
	vec.Vec_PtrWriteEntry(p.vBoxes, int(box.Id), box)
	p.nBoxes++
	return nil
}
func If_LibBoxNum(p *If_LibBox_t) int32 {
	if p == nil {
		return 0
	}
	return p.nBoxes
}

type boxLibraryLine struct {
	fields []string
	line   int
}

func boxLibraryLines(text string) ([]boxLibraryLine, error) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 4096), 4<<20)
	var lines []boxLibraryLine
	number := 0
	for scanner.Scan() {
		number++
		line, _, _ := strings.Cut(scanner.Text(), "#")
		if fields := strings.Fields(line); len(fields) > 0 {
			lines = append(lines, boxLibraryLine{fields, number})
		}
	}
	return lines, scanner.Err()
}
func boxNonnegative(text string) (int32, error) {
	value, err := strconv.ParseInt(text, 10, 32)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("invalid nonnegative integer %q", text)
	}
	return int32(value), nil
}
func boxDelay(text string, extended bool) (int32, error) {
	missing := BoxNoArc
	if extended {
		missing = -1
	}
	if text == "-" {
		return missing, nil
	}
	value, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid delay %q", text)
	}
	if value < 0 {
		return missing, nil
	}
	return int32(value), nil
}

// ParseBoxLibrary reads classic name/id/type/input/output records or .box
// records with comb/seq, white/black and inner/outer attributes.
func ParseBoxLibrary(name, text string, extended bool) (*If_LibBox_t, error) {
	lines, err := boxLibraryLines(text)
	if err != nil {
		return nil, err
	}
	lib := If_LibBoxStart()
	fail := func(line int, err error) (*If_LibBox_t, error) { return nil, fmt.Errorf("%s:%d: %w", name, line, err) }
	if !extended {
		var tokens []string
		var numbers []int
		for _, line := range lines {
			tokens = append(tokens, line.fields...)
			for range line.fields {
				numbers = append(numbers, line.line)
			}
		}
		for index := 0; index < len(tokens); {
			start := index
			line := numbers[index]
			if len(tokens)-index < 5 {
				return fail(line, fmt.Errorf("incomplete box header"))
			}
			header := tokens[index : index+5]
			index += 5
			if strings.HasPrefix(header[0], ".") {
				return fail(line, fmt.Errorf("extended box format requires extended reader"))
			}
			id, e := boxNonnegative(header[1])
			if e != nil {
				return fail(line, e)
			}
			white, e := boxNonnegative(header[2])
			if e != nil || white > 1 {
				return fail(line, fmt.Errorf("box type must be 0 or 1"))
			}
			inputs, e := boxNonnegative(header[3])
			if e != nil {
				return fail(line, e)
			}
			outputs, e := boxNonnegative(header[4])
			if e != nil {
				return fail(line, e)
			}
			entries := int64(inputs) * int64(outputs)
			if entries > int64(len(tokens)-index) {
				return fail(line, fmt.Errorf("box %s needs %d delays", header[0], entries))
			}
			box := If_BoxStart(header[0], id, inputs, outputs, 0, 1-white, 0)
			if box == nil {
				return fail(line, fmt.Errorf("invalid box dimensions"))
			}
			for i := range box.pDelays {
				box.pDelays[i], e = boxDelay(tokens[index], false)
				if e != nil {
					return fail(numbers[index], e)
				}
				index++
			}
			if e = If_LibBoxAdd(lib, box); e != nil {
				return fail(numbers[start], e)
			}
		}
	} else {
		suffix := make([]int, len(lines)+1)
		for i := len(lines) - 1; i >= 0; i-- {
			suffix[i] = suffix[i+1] + len(lines[i].fields)
		}
		var current *If_Box_t
		filled := 0
		for lineIndex, line := range lines {
			fields := line.fields
			if fields[0] == ".box" {
				if current != nil && filled < len(current.pDelays) {
					return fail(line.line, fmt.Errorf("box %s table is incomplete", current.pName))
				}
				if len(fields) < 5 {
					return fail(line.line, fmt.Errorf("incomplete .box header"))
				}
				id, e := boxNonnegative(fields[1])
				if e != nil {
					return fail(line.line, e)
				}
				inputs, e := boxNonnegative(fields[3])
				if e != nil {
					return fail(line.line, e)
				}
				outputs, e := boxNonnegative(fields[4])
				if e != nil {
					return fail(line.line, e)
				}
				// Bound allocation by the actual input token count before constructing.
				entries := int64(inputs) * int64(outputs)
				available := suffix[lineIndex+1]
				if entries > int64(available) {
					return fail(line.line, fmt.Errorf("box %s table is incomplete", fields[2]))
				}
				seq, black, outer := int32(0), int32(0), int32(0)
				seen := make(map[string]bool)
				for _, attribute := range fields[5:] {
					var group string
					switch attribute {
					case "seq", "comb":
						group = "kind"
					case "black", "white":
						group = "color"
					case "outer", "inner":
						group = "location"
					default:
						return fail(line.line, fmt.Errorf("unknown box attribute %q", attribute))
					}
					if seen[group] {
						return fail(line.line, fmt.Errorf("repeated/conflicting box attribute %s", attribute))
					}
					seen[group] = true
					switch attribute {
					case "seq":
						seq = 1
					case "black":
						black = 1
					case "outer":
						outer = 1
					}
				}
				current = If_BoxStart(fields[2], id, inputs, outputs, seq, black, outer)
				if current == nil {
					return fail(line.line, fmt.Errorf("invalid box dimensions"))
				}
				if e = If_LibBoxAdd(lib, current); e != nil {
					return fail(line.line, e)
				}
				filled = 0
				continue
			}
			if current == nil {
				return fail(line.line, fmt.Errorf("delay table appears before .box header"))
			}
			for _, token := range fields {
				if filled >= len(current.pDelays) {
					return fail(line.line, fmt.Errorf("too many box delays"))
				}
				value, e := boxDelay(token, true)
				if e != nil {
					return fail(line.line, e)
				}
				current.pDelays[filled] = value
				filled++
			}
		}
		if current != nil && filled < len(current.pDelays) {
			return fail(len(strings.Split(text, "\n")), fmt.Errorf("box %s table is incomplete", current.pName))
		}
	}
	if lib.nBoxes == 0 {
		return nil, fmt.Errorf("%s: box library is empty", name)
	}
	return lib, nil
}
func ReadBoxLibrary(filename string, extended bool) (*If_LibBox_t, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseBoxLibrary(filename, string(data), extended)
}
func If_LibBoxRead(filename string) *If_LibBox_t {
	p, err := ReadBoxLibrary(filename, false)
	if err != nil {
		return nil
	}
	return p
}
func If_LibBoxRead2(filename string) *If_LibBox_t {
	p, err := ReadBoxLibrary(filename, true)
	if err != nil {
		return nil
	}
	return p
}
func If_LibBoxGetToken(r io.Reader) string {
	var token strings.Builder
	var b [1]byte
	comment := false
	emptyReads := 0
	for {
		n, err := r.Read(b[:])
		if n > 0 {
			emptyReads = 0
			c := b[0]
			if comment {
				if c == '\n' {
					comment = false
				}
				if token.Len() > 0 && c == '\n' {
					return token.String()
				}
				continue
			}
			if c == '#' {
				comment = true
				continue
			}
			if strings.ContainsRune(" \t\r\n", rune(c)) {
				if token.Len() > 0 {
					return token.String()
				}
				continue
			}
			token.WriteByte(c)
		}
		if err != nil {
			return token.String()
		}
		if n == 0 {
			emptyReads++
			if emptyReads >= 100 {
				return token.String()
			}
		}
	}
}
func WriteBoxLibrary(w io.Writer, lib *If_LibBox_t, extended bool) error {
	if lib == nil {
		return fmt.Errorf("nil box library")
	}
	if _, err := fmt.Fprintln(w, "# Box delay library"); err != nil {
		return err
	}
	for _, box := range lib.Boxes() {
		if int64(box.nPis)*int64(box.nPos) != int64(len(box.pDelays)) {
			return fmt.Errorf("box %s delay dimensions differ", box.pName)
		}
		if extended {
			kind, color, location := "comb", "white", "inner"
			if box.fSeq != 0 {
				kind = "seq"
			}
			if box.fBlack != 0 {
				color = "black"
			}
			if box.fOuter != 0 {
				location = "outer"
			}
			if _, err := fmt.Fprintf(w, ".box %d %s %d %d %s %s %s\n", box.Id, box.pName, box.nPis, box.nPos, kind, color, location); err != nil {
				return err
			}
		} else {
			if box.fSeq != 0 || box.fOuter != 0 {
				return fmt.Errorf("box %s attributes require extended output", box.pName)
			}
			white := 1
			if box.fBlack != 0 {
				white = 0
			}
			if _, err := fmt.Fprintf(w, "%s %d %d %d %d\n", box.pName, box.Id, white, box.nPis, box.nPos); err != nil {
				return err
			}
		}
		for output := int32(0); output < box.nPos; output++ {
			for input := int32(0); input < box.nPis; input++ {
				value := box.pDelays[output*box.nPis+input]
				text := strconv.FormatInt(int64(value), 10)
				if value < 0 {
					text = "-"
				}
				if _, err := fmt.Fprintf(w, "%s ", text); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
	}
	return nil
}
func If_LibBoxPrint(w io.Writer, lib *If_LibBox_t) error {
	extended := false
	for _, box := range lib.Boxes() {
		extended = extended || box.fSeq != 0 || box.fOuter != 0
	}
	return WriteBoxLibrary(w, lib, extended)
}
func If_LibBoxWrite(filename string, lib *If_LibBox_t) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = If_LibBoxPrint(file, lib)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
func If_LibBoxLoad(filename string) int32 {
	companion := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".cdl"
	lib, err := ReadBoxLibrary(companion, true)
	if err != nil {
		return 0
	}
	abcmain.Abc_FrameSetLibBox(lib)
	return 1
}

func (p *If_Box_t) Sequential() bool                { return p.fSeq != 0 }
func (p *If_Box_t) BlackBox() bool                  { return p.fBlack != 0 }
func (p *If_Box_t) Outer() bool                     { return p.fOuter != 0 }
func (p *If_Box_t) HasArc(output, input int32) bool { return p.Delay(output, input) >= 0 }
