package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

func Io_ReadDsdFindEnd(text string) string {
	if text == "" || text[0] != '(' {
		return ""
	}
	depth := 0
	for i, c := range text {
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
			if depth == 0 {
				return text[i:]
			}
		}
	}
	return ""
}
func dsdParts(text string) ([]string, byte, error) {
	var parts []string
	depth, start := 0, 0
	var operator byte
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return nil, 0, fmt.Errorf("unbalanced DSD parentheses")
			}
		case '*', '+', ',':
			if depth == 0 {
				if operator != 0 && operator != text[i] {
					return nil, 0, fmt.Errorf("mixed DSD separators")
				}
				operator = text[i]
				if i == start {
					return nil, 0, fmt.Errorf("empty DSD argument")
				}
				parts = append(parts, text[start:i])
				start = i + 1
			}
		}
	}
	if depth != 0 || start == len(text) {
		return nil, 0, fmt.Errorf("unclosed or empty DSD argument")
	}
	parts = append(parts, text[start:])
	return parts, operator, nil
}
func Io_ReadDsdStrSplit(text string, parts []string, xor *int32) int32 {
	values, operator, err := dsdParts(text)
	if err != nil || len(values) > len(parts) {
		return 0
	}
	copy(parts, values)
	if xor != nil {
		*xor = 0
		if operator == '+' {
			*xor = 1
		}
	}
	return int32(len(values))
}
func dsdNode(m *blifModel, text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("empty DSD expression")
	}
	if text[0] == '!' {
		input, err := dsdNode(m, text[1:])
		if err != nil {
			return "", err
		}
		output := m.uniqueNetName("__dsd_not")
		_, err = m.node([]string{input, output}, "0 1\n")
		return output, err
	}
	if len(text) == 1 && text[0] >= 'a' && text[0] <= 'z' {
		return text, nil
	}
	at := strings.IndexByte(text, '(')
	if at < 0 || text[len(text)-1] != ')' {
		return "", fmt.Errorf("expected DSD group in %q", text)
	}
	prefix := text[:at]
	parts, operator, err := dsdParts(text[at+1 : len(text)-1])
	if err != nil {
		return "", err
	}
	inputs := make([]string, len(parts))
	for i, part := range parts {
		inputs[i], err = dsdNode(m, part)
		if err != nil {
			return "", err
		}
	}
	output := m.uniqueNetName("__dsd_gate")
	var cover string
	if prefix != "" {
		cover, err = benchHexCover(prefix, len(inputs))
		if err != nil {
			return "", err
		}

	} else if operator == '+' {
		signal := inputs[0]
		for _, input := range inputs[1:] {
			next := m.uniqueNetName("__dsd_xor")
			if _, err = m.node([]string{signal, input, next}, "01 1\n10 1\n"); err != nil {
				return "", err
			}
			signal = next
		}
		_, err = m.node([]string{signal, output}, "1 1\n")
		return output, err
	} else {
		cover, err = benchGateCover("AND", len(inputs))
		if err != nil {
			return "", err
		}
	}
	_, err = m.node(append(inputs, output), cover)
	return output, err
}
func ParseDSD(expression string) (*abc.Network, error) {
	expression = strings.Join(strings.Fields(expression), "")
	m := &blifModel{network: abc.NewNetwork("dsd"), nets: map[string]*abc.Object{}, domains: map[string]*abc.MVVariable{}}
	m.network.Type = uint8(abc.ABC_NTK_NETLIST)
	m.network.Function = uint8(abc.ABC_FUNC_SOP)
	highest := byte(0)
	for i := range expression {
		if expression[i] >= 'a' && expression[i] <= 'z' && expression[i] > highest {
			highest = expression[i]
		}
	}
	if highest != 0 {
		for c := byte('a'); c <= highest; c++ {
			if err := m.input(string(c)); err != nil {
				return nil, err
			}
		}
	}
	signal, err := dsdNode(m, expression)
	if err != nil {
		return nil, err
	}
	if err = m.output(signal); err != nil {
		return nil, err
	}
	return m.network, nil
}
func Io_ReadDsd(expression string) *abc.Network {
	n, err := ParseDSD(expression)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
