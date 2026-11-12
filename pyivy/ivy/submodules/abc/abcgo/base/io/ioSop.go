package io

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"strings"
)

func sopExpression(node *abc.Object, signalName func(*abc.Object) string, and, or, one, zero string) (string, error) {
	text, ok := node.Data.(string)
	if !ok {
		return "", fmt.Errorf("equation output requires SOP node functions")
	}
	inputs := node.Fanins()
	var cubes []string
	phase := ""
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pattern, value := "", ""
		if len(fields) == 1 {
			value = fields[0]
		} else if len(fields) == 2 {
			pattern, value = fields[0], fields[1]
		} else {
			return "", fmt.Errorf("invalid SOP row")
		}
		if value != "0" && value != "1" {
			return "", fmt.Errorf("invalid SOP phase")
		}
		if phase != "" && phase != value {
			return "", fmt.Errorf("mixed SOP phases")
		}
		phase = value
		if pattern != "" && len(pattern) != len(inputs) {
			return "", fmt.Errorf("SOP width mismatch")
		}
		var literals []string
		for i, c := range pattern {
			if c == '-' {
				continue
			}
			if c != '0' && c != '1' {
				return "", fmt.Errorf("invalid SOP literal")
			}
			signal := signalName(inputs[i].Object)
			if inputs[i].Complement != (c == '0') {
				signal = "!" + signal
			}
			literals = append(literals, signal)
		}
		cube := one
		if len(literals) > 0 {
			cube = "(" + strings.Join(literals, and) + ")"
		}
		cubes = append(cubes, cube)
	}
	if len(cubes) == 0 {
		return zero, nil
	}
	result := "(" + strings.Join(cubes, or) + ")"
	if phase == "0" {
		result = "!" + result
	}
	return result, nil
}

func sopEquation(node *abc.Object) (string, error) {
	return sopExpression(node, blifObjectName, "*", "+", "1", "0")
}
