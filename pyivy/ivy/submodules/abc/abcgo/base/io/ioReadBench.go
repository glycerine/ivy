package io

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/extra"
)

func ParseBench(name, text string, check bool) (*abc.Network, error) {
	lines, err := blifLines(name, text)
	if err != nil {
		return nil, err
	}
	m := &blifModel{network: abc.NewNetwork(name), nets: map[string]*abc.Object{}, domains: map[string]*abc.MVVariable{}}
	// Abc_NtkStartRead: the name drops the extension, the spec keeps the file
	m.network.Name, m.network.Spec = extra.Extra_FileNameGeneric(name), name
	lutsPresent := false
	m.network.Type = uint8(abc.ABC_NTK_NETLIST)
	m.network.Function = uint8(abc.ABC_FUNC_SOP)
	for _, line := range lines {
		s := strings.TrimSpace(line.text)
		upper := strings.ToUpper(s)
		for _, directive := range []string{"INPUT", "OUTPUT"} {
			if strings.HasPrefix(upper, directive+"(") && strings.HasSuffix(s, ")") {
				signal := strings.TrimSpace(s[len(directive)+1 : len(s)-1])
				if directive == "INPUT" {
					err = m.input(signal)
				} else {
					err = m.output(signal)
				}
				if err != nil {
					return nil, blifError(name, line, "%v", err)
				}
				s = ""
				break
			}
		}
		if s == "" {
			continue
		}
		assignment := strings.SplitN(s, "=", 2)
		if len(assignment) != 2 {
			return nil, blifError(name, line, "expected assignment")
		}
		output := strings.TrimSpace(assignment[0])
		rhs := strings.TrimSpace(assignment[1])
		if strings.HasPrefix(rhs, "gnd") || strings.HasPrefix(rhs, "vdd") {
			// gate type gnd/vdd with no fanins: a constant node
			cover := " 0\n"
			if strings.HasPrefix(rhs, "vdd") {
				cover = " 1\n"
			}
			if _, err = m.node([]string{output}, cover); err != nil {
				return nil, blifError(name, line, "%v", err)
			}
			continue
		}
		at := strings.IndexByte(rhs, '(')
		if at < 0 || !strings.HasSuffix(rhs, ")") {
			return nil, blifError(name, line, "expected gate(inputs)")
		}
		gate := strings.ToUpper(strings.TrimSpace(rhs[:at]))
		inputs := strings.Split(rhs[at+1:len(rhs)-1], ",")
		for i := range inputs {
			inputs[i] = strings.TrimSpace(inputs[i])
		}
		if len(inputs) == 1 && inputs[0] == "" {
			inputs = nil
		}

		if strings.HasPrefix(gate, "DFF") {
			if len(inputs) == 4 {
				if err = m.input(output); err != nil {
					return nil, err
				}
				for _, input := range inputs {
					target := output + "_" + input
					if _, err = m.node([]string{input, target}, "1 1\n"); err != nil {
						return nil, err
					}
					if err = m.output(target); err != nil {
						return nil, err
					}
				}
				m.network.Constraints++
				continue
			}
			if len(inputs) == 0 {
				return nil, blifError(name, line, "DFF requires data input")
			}
			init := "2"
			if gate == "DFF0" {
				init = "0"
			} else if gate == "DFF1" {
				init = "1"
			}
			err = blifLatch(m, []string{inputs[0], output, init})
			if err != nil {
				return nil, blifError(name, line, "%v", err)
			}
			continue
		}
		if (gate == "XOR" || gate == "XNOR") && len(inputs) > 2 {
			signal := inputs[0]
			for i, input := range inputs[1:] {
				target := output
				last := i == len(inputs)-2
				if !last {
					target = m.uniqueNetName("__bench_xor")
				}
				cover := "01 1\n10 1\n"
				if last && gate == "XNOR" {
					cover = "00 1\n11 1\n"
				}
				if _, err = m.node([]string{signal, input, target}, cover); err != nil {
					return nil, blifError(name, line, "%v", err)
				}
				signal = target
			}
			continue
		}
		var cover string
		if fields := strings.Fields(gate); len(fields) == 2 && fields[0] == "LUT" {
			lutsPresent = true
			if len(inputs) > 8 {
				fmt.Printf("%s: Currently cannot read truth tables with more than 8 inputs (%d).\n", name, len(inputs))
				return nil, blifError(name, line, "LUT with more than 8 inputs")
			}
			if !strings.HasPrefix(fields[1], "0X") {
				return nil, blifError(name, line, "LUT requires hexadecimal signature")
			}
			hex := fields[1][2:]
			cover, err = benchHexCover(hex, len(inputs))
			if err == nil {
				switch {
				case !strings.Contains(cover, " 1\n"): // constant 0: no fanins
					cover, inputs = " 0\n", nil
				case len(inputs) > 0 && strings.Count(cover, "\n") == 1<<uint(len(inputs)):
					cover, inputs = " 1\n", nil // constant 1
				case len(inputs) == 1 && strings.HasPrefix(hex, "2"):
					cover = "1 1\n"
				case len(inputs) == 1 && strings.HasPrefix(hex, "1"):
					cover = "0 1\n"
				case len(inputs) == 1:
					fmt.Printf("%s: Reading truth table (%s) of single-input node has failed.\n", name, hex)
					return nil, blifError(name, line, "single-input LUT %s", hex)
				}
			}
		} else {
			cover, err = benchGateCover(gate, len(inputs))
		}
		if err != nil {
			return nil, blifError(name, line, "%v", err)
		}

		if _, err = m.node(append(inputs, output), cover); err != nil {
			return nil, blifError(name, line, "%v", err)
		}
	}
	// undriven constant nets, in Io_ReadBenchNetwork's order
	for _, c := range []struct{ signal, value, note string }{
		{"gnd", " 0\n", ""},
		{"1", " 0\n", "Io_ReadBenchNetwork(): Adding constant 0 fanin to non-driven net \"1\".\n"},
		{"vdd", " 1\n", ""},
		{"2", " 1\n", "Io_ReadBenchNetwork(): Adding constant 1 fanin to non-driven net \"2\".\n"},
	} {
		if net := m.nets[c.signal]; net != nil && net.FaninNum() == 0 {
			if c.note != "" {
				fmt.Print(c.note)
			}
			if _, err = m.node([]string{c.signal}, c.value); err != nil {
				return nil, err
			}
		}
	}
	// if LUTs are present, collapse the truth tables into cubes
	if lutsPresent {
		if abc.Abc_NtkToBdd(m.network) == 0 {
			fmt.Printf("Io_ReadBenchNetwork(): Converting to BDD has failed.\n")
			return nil, fmt.Errorf("%s: converting to BDD has failed", name)
		}
		if abc.Abc_NtkToSop(m.network, -1, util.ABC_INFINITY) == 0 {
			fmt.Printf("Io_ReadBenchNetwork(): Converting to SOP has failed.\n")
			return nil, fmt.Errorf("%s: converting to SOP has failed", name)
		}
	}
	if check {
		for signal, net := range m.nets {
			if net.FaninNum() == 0 {
				return nil, fmt.Errorf("%s: undriven signal %s", name, signal)
			}
		}
		if err = m.network.Check(); err != nil {
			return nil, err
		}
	}
	return m.network, nil
}
func benchGateCover(gate string, n int) (string, error) {
	switch gate {
	case "BUFF", "BUF":
		if n == 1 {
			return "1 1\n", nil
		}
	case "NOT":
		if n == 1 {
			return "0 1\n", nil
		}
	case "AND", "NAND":
		if n > 0 {
			value := "1"
			if gate == "NAND" {
				value = "0"
			}
			return strings.Repeat("1", n) + " " + value + "\n", nil
		}
	case "OR", "NOR":
		if n > 0 {
			value := "0"
			if gate == "NOR" {
				value = "1"
			}
			return strings.Repeat("0", n) + " " + value + "\n", nil
		}
	case "XOR", "XNOR":
		if n > 0 && n <= 16 {
			var out strings.Builder
			for word := 0; word < 1<<uint(n); word++ {
				parity := 0
				pattern := make([]byte, n)
				for i := range pattern {
					bit := (word >> uint(i)) & 1
					parity ^= bit
					pattern[i] = byte('0' + bit)
				}
				if (parity == 1) == (gate == "XOR") {
					fmt.Fprintf(&out, "%s 1\n", pattern)
				}
			}
			return out.String(), nil
		}
	}
	return "", fmt.Errorf("unsupported gate %s with %d inputs", gate, n)
}
func ReadBench(filename string, check bool) (*abc.Network, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseBench(filename, string(data), check)
}
func Io_ReadBench(filename string, check int32) *abc.Network {
	n, err := ReadBench(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}

func benchHexCover(hex string, n int) (string, error) {
	if n < 0 || n > 20 {
		return "", fmt.Errorf("LUT supports at most20 inputs")
	}
	if hex == "" {
		return "", fmt.Errorf("empty LUT signature")
	}
	for _, digit := range hex {
		if !strings.ContainsRune("0123456789abcdefABCDEF", digit) {
			return "", fmt.Errorf("invalid LUT signature")
		}
	}
	var out strings.Builder
	for word := 0; word < 1<<uint(n); word++ {
		index := len(hex) - 1 - word/4
		if index < 0 {
			continue
		}
		value, err := strconv.ParseUint(hex[index:index+1], 16, 4)
		if err != nil {
			return "", err
		}
		if value&(1<<uint(word%4)) == 0 {
			continue
		}
		pattern := make([]byte, n)
		for i := range pattern {
			pattern[i] = '0' + byte(word>>uint(i)&1)
		}
		fmt.Fprintf(&out, "%s 1\n", pattern)
	}
	if out.Len() == 0 {
		return strings.Repeat("-", n) + " 0\n", nil
	}
	return out.String(), nil
}
