package pla

/**CFile****************************************************************

  FileName    [plaRead.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaRead.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func Pla_ReadFile(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{'\n'}, data...), '\n', 0), nil
}
func Pla_ReadPlaRemoveComments(data []byte) {
	comment := false
	for i, c := range data {
		if c == '\n' {
			comment = false
		}
		if c == '#' {
			comment = true
		}
		if comment {
			data[i] = ' '
		}
	}
}
func parseType(s string) (Pla_File_t, error) {
	switch s {
	case "fd":
		return PLA_FILE_FD, nil
	case "f":
		return PLA_FILE_F, nil
	case "fr":
		return PLA_FILE_FR, nil
	case "fdr":
		return PLA_FILE_FDR, nil
	}
	return PLA_FILE_NONE, fmt.Errorf("unsupported PLA type %q", s)
}
func parseLiteral(c byte, kind Pla_File_t) (Pla_Lit_t, error) {
	switch c {
	case '-', '2':
		return PLA_LIT_DASH, nil
	case '0':
		return PLA_LIT_ZERO, nil
	case '1':
		return PLA_LIT_ONE, nil
	case '?':
		return PLA_LIT_FULL, nil
	case '~':
		switch kind {
		case PLA_FILE_F, PLA_FILE_FD:
			return PLA_LIT_ZERO, nil
		case PLA_FILE_FR:
			return PLA_LIT_DASH, nil
		case PLA_FILE_FDR:
			return PLA_LIT_FULL, nil
		}
	}
	return 0, fmt.Errorf("invalid PLA literal %q", c)
}
func Parse(r io.Reader, name string) (*Pla_Man_t, error) {
	ins, outs, declared := int32(-1), int32(-1), int32(-1)
	kind := PLA_FILE_FD
	var rows []string
	var directives []string
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		text = strings.TrimRight(text, "\x00")
		if text == "" {
			continue
		}
		fields := strings.Fields(text)
		if fields[0][0] == '.' {
			key := fields[0]
			if key == ".e" || key == ".end" {
				break
			}
			switch key {
			case ".i", ".o", ".p":
				if len(fields) != 2 {
					return nil, fmt.Errorf("line %d: %s requires one count", line, key)
				}
				n, err := strconv.ParseInt(fields[1], 10, 32)
				if err != nil || n < 0 {
					return nil, fmt.Errorf("line %d: invalid %s count", line, key)
				}
				switch key {
				case ".i":
					ins = int32(n)
				case ".o":
					outs = int32(n)
				case ".p":
					declared = int32(n)
				}
			case ".type":
				if len(fields) != 2 {
					return nil, fmt.Errorf("line %d: missing PLA type", line)
				}
				var err error
				kind, err = parseType(fields[1])
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", line, err)
				}
			case ".mv":
				return nil, fmt.Errorf("line %d: multi-valued PLA is not a binary PLA manager input", line)
			default:
				directives = append(directives, text)
			}
			continue
		}
		if ins <= 0 || outs <= 0 {
			return nil, fmt.Errorf("line %d: positive .i and .o must precede cubes", line)
		}
		if len(fields) != 2 || len(fields[0]) != int(ins) || len(fields[1]) != int(outs) {
			return nil, fmt.Errorf("line %d: cube width does not match .i %d .o %d", line, ins, outs)
		}
		rows = append(rows, fields[0]+fields[1])
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if ins <= 0 || outs <= 0 {
		return nil, fmt.Errorf("positive .i and .o declarations are required")
	}
	p := Pla_ManAlloc(name, ins, outs, int32(len(rows)))
	p.Type = kind
	p.Directives = directives
	if declared >= 0 && int(declared) != len(rows) {
		p.Warnings = append(p.Warnings, fmt.Sprintf("declared number of cubes (%d) differs from actual (%d)", declared, len(rows)))
	}
	for i, row := range rows {
		for j, c := range []byte(row) {
			lit, err := parseLiteral(c, kind)
			if err != nil {
				return nil, fmt.Errorf("cube %d literal %d: %w", i+1, j+1, err)
			}
			if j < int(ins) {
				Pla_CubeSetLit(p.Cubes[i].In, int32(j), lit)
			} else {
				Pla_CubeSetLit(p.Cubes[i].Out, int32(j)-ins, lit)
			}
		}
	}
	return p, nil
}
func ReadFile(name string) (*Pla_Man_t, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f, name)
}
func Pla_ReadPla(name string) *Pla_Man_t { p, _ := ReadFile(name); return p }

// Pla_ReadPlaHeader exposes the source's header-only pass without interpreting
// cube text. Counts absent from the header retain the original -1 sentinel.
func Pla_ReadPlaHeader(data []byte) (ins, outs, cubes int32, kind Pla_File_t, err error) {
	ins, outs, cubes = -1, -1, -1
	kind = PLA_FILE_FD
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		text := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == ".e" || fields[0] == ".end" {
			break
		}
		switch fields[0] {
		case ".i", ".o", ".p":
			if len(fields) != 2 {
				err = fmt.Errorf("invalid %s declaration", fields[0])
				return
			}
			n, e := strconv.ParseInt(fields[1], 10, 32)
			if e != nil {
				err = e
				return
			}
			switch fields[0] {
			case ".i":
				ins = int32(n)
			case ".o":
				outs = int32(n)
			case ".p":
				cubes = int32(n)
			}
		case ".type":
			if len(fields) != 2 {
				err = fmt.Errorf("missing PLA type")
				return
			}
			kind, err = parseType(fields[1])
			if err != nil {
				return
			}
		}
	}
	if scanner.Err() != nil {
		err = scanner.Err()
		return
	}
	if ins <= 0 || outs <= 0 {
		err = fmt.Errorf("positive .i and .o declarations are required")
	}
	return
}
func Pla_ReadPlaBody(data []byte, kind Pla_File_t) ([]Pla_Lit_t, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var result []Pla_Lit_t
	for scanner.Scan() {
		text := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if strings.HasPrefix(text, ".e") {
			break
		}
		if strings.HasPrefix(text, ".") {
			continue
		}
		for _, c := range []byte(text) {
			if c == ' ' || c == '\t' || c == '\r' || c == 0 {
				continue
			}
			lit, err := parseLiteral(c, kind)
			if err != nil {
				return nil, err
			}
			result = append(result, lit)
		}
	}
	return result, scanner.Err()
}
func Pla_ReadAddBody(p *Pla_Man_t, literals []Pla_Lit_t) error {
	width := int(p.Ins + p.Outs)
	if width <= 0 || len(literals)%width != 0 {
		return fmt.Errorf("literal count is not a multiple of cube width")
	}
	next := Pla_ManAlloc(p.Spec, p.Ins, p.Outs, int32(len(literals)/width))
	for i, lit := range literals {
		cube, position := i/width, i%width
		if lit > PLA_LIT_FULL {
			return fmt.Errorf("invalid literal code")
		}
		if position < int(p.Ins) {
			Pla_CubeSetLit(next.Cubes[cube].In, int32(position), lit)
		} else {
			Pla_CubeSetLit(next.Cubes[cube].Out, int32(position)-p.Ins, lit)
		}
	}
	p.Cubes = next.Cubes
	p.CubeLits = nil
	p.Occurs = nil
	return nil
}
