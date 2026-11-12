package exor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func writeText(w io.Writer, s string) error {
	n, err := io.WriteString(w, s)
	if err == nil && n != len(s) {
		return io.ErrShortWrite
	}
	return err
}
func WriteTableIntoFile(m *Manager, w io.Writer) error {
	for _, c := range m.Cubes {
		var line strings.Builder
		for i := 0; i < m.Inputs; i++ {
			switch GetVar(c, i) {
			case VAR_NEG:
				line.WriteByte('0')
			case VAR_POS:
				line.WriteByte('1')
			case VAR_ABS:
				line.WriteByte('-')
			default:
				return fmt.Errorf("exor: invalid literal")
			}
		}
		line.WriteByte(' ')
		for i := 0; i < m.Outputs; i++ {
			if c.Output(i) {
				line.WriteByte('1')
			} else {
				line.WriteByte('0')
			}
		}
		line.WriteByte('\n')
		if err := writeText(w, line.String()); err != nil {
			return err
		}
	}
	return nil
}
func (m *Manager) WritePLA(w io.Writer) error {
	if err := writeText(w, fmt.Sprintf("# EXORCISM Go ESOP cover\n.i %d\n.o %d\n.p %d\n.type esop\n", m.Inputs, m.Outputs, len(m.Cubes))); err != nil {
		return err
	}
	if err := WriteTableIntoFile(m, w); err != nil {
		return err
	}
	return writeText(w, ".e\n")
}
func WriteResultIntoFile(m *Manager, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	writeErr := m.WritePLA(f)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
func ReadPLA(r io.Reader) (*Manager, error) {
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 4096), 16<<20)
	ni, no, declared := -1, -1, -1
	var rows []string
	ended := false
	esop := false
	line := 0
	for scan.Scan() {
		line++
		s := strings.TrimSpace(strings.SplitN(scan.Text(), "#", 2)[0])
		if s == "" {
			continue
		}
		fields := strings.Fields(s)
		if fields[0][0] != '.' {
			if ended {
				return nil, fmt.Errorf("exor: data after .e at line %d", line)
			}
			rows = append(rows, s)
			continue
		}
		switch fields[0] {
		case ".i", ".o", ".p":
			if len(fields) != 2 {
				return nil, fmt.Errorf("exor: invalid directive at line %d", line)
			}
			n, err := strconv.Atoi(fields[1])
			if err != nil || n < 0 {
				return nil, fmt.Errorf("exor: invalid count at line %d", line)
			}
			switch fields[0] {
			case ".i":
				ni = n
			case ".o":
				no = n
			case ".p":
				declared = n
			}
		case ".type":
			esop = true
			if len(fields) != 2 || fields[1] != "esop" {
				return nil, fmt.Errorf("exor: expected .type esop")
			}
		case ".e", ".end":
			ended = true
		default:
			return nil, fmt.Errorf("exor: unsupported PLA directive %s", fields[0])
		}
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	if !esop {
		return nil, fmt.Errorf("exor: explicit .type esop is required")
	}
	m, err := NewManager(ni, no, Options{})
	if err != nil {
		return nil, err
	}
	if declared >= 0 && declared != len(rows) {
		return nil, fmt.Errorf("exor: declared cube count does not match rows")
	}
	for _, row := range rows {
		f := strings.Fields(row)
		if ni == 0 && len(f) == 1 {
			f = append([]string{""}, f...)
		}
		if len(f) != 2 || len(f[0]) != ni || len(f[1]) != no {
			return nil, fmt.Errorf("exor: incorrect PLA row dimensions")
		}
		inputs := make([]VarValue, ni)
		outputs := make([]bool, no)
		for i, ch := range f[0] {
			switch ch {
			case '0':
				inputs[i] = VAR_NEG
			case '1':
				inputs[i] = VAR_POS
			case '-':
				inputs[i] = VAR_ABS
			default:
				return nil, fmt.Errorf("exor: invalid PLA input")
			}
		}
		for i, ch := range f[1] {
			if ch != '0' && ch != '1' {
				return nil, fmt.Errorf("exor: invalid PLA output")
			}
			outputs[i] = ch == '1'
		}
		if err = m.Add(NewCube(inputs, outputs)); err != nil {
			return nil, err
		}
	}
	return m, nil
}
func (m *Manager) Evaluate(inputs []bool) ([]bool, error) {
	if len(inputs) != m.Inputs {
		return nil, fmt.Errorf("exor: wrong input count")
	}
	out := make([]bool, m.Outputs)
	for _, c := range m.Cubes {
		active := true
		for i, v := range inputs {
			lit := GetVar(c, i)
			if lit == VAR_NEG && v || lit == VAR_POS && !v {
				active = false
				break
			}
		}
		if active {
			for i := range out {
				out[i] = out[i] != c.Output(i)
			}
		}
	}
	return out, nil
}
