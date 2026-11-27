package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
)

// ParseEqn reads EQN text as Io_ReadEqn does: statements end at ';', '#'
// starts a comment, and each "name = formula" becomes one node of an AIG
// netlist whose fanins are the formula's distinct names in order of
// appearance (Io_ReadEqnNetwork, Parse_FormulaParserEqn).
func ParseEqn(name, text string, check bool) (*abc.Network, error) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_AIG, 1)
	n.Name, n.Spec = extra.Extra_FileNameGeneric(name), name
	if n.Name == "" {
		n.Name = "unknown"
	}
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	for _, statement := range eqnStatements(text) {
		tokens := eqnTokens(statement)
		if len(tokens) == 0 {
			continue
		}
		left := Io_ReadEqnStrCompact(tokens[0])
		if left == "" {
			break
		}
		if len(tokens) != 2 {
			fmt.Printf("%s: Wrong input file format.\n", name)
			return nil, fmt.Errorf("%s: wrong input file format", name)
		}
		switch {
		case strings.HasPrefix(left, "INORDER"):
			for _, input := range Io_ReadEqnStrCutAt(tokens[1], " \n\r\t", false) {
				Io_ReadCreatePi(n, input)
			}
		case strings.HasPrefix(left, "OUTORDER"):
			for _, output := range Io_ReadEqnStrCutAt(tokens[1], " \n\r\t", false) {
				Io_ReadCreatePo(n, output)
			}
		default:
			formula := Io_ReadEqnStrCompact(tokens[1])
			var vars []string
			if !(len(formula) == 1 && (formula[0] == '0' || formula[0] == '1')) {
				vars = Io_ReadEqnStrCutAt(formula, "!*+()", true)
			}
			node := Io_ReadCreateNode(n, left, vars, int32(len(vars)))
			function := hop.ParseEquation(manager, os.Stdout, formula, vars)
			if function == nil {
				return nil, fmt.Errorf("%s: cannot parse the formula of %s", name, left)
			}
			node.Data = function
		}
	}
	abc.Abc_NtkFinalizeRead(n)
	if check && abc.Abc_NtkCheckRead(n) == 0 {
		fmt.Printf("Io_ReadEqn: The network check has failed.\n")
		return nil, fmt.Errorf("%s: the network check has failed", name)
	}
	return n, nil
}

// eqnStatements splits the text at ';' after dropping '#' comments, as
// Extra_FileReader does for Io_ReadEqn.
func eqnStatements(text string) []string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(text, "\n") {
		if at := strings.IndexByte(line, '#'); at >= 0 {
			line = line[:at] + "\n"
		}
		b.WriteString(line)
	}
	return strings.Split(b.String(), ";")
}

// eqnTokens splits a statement at '=' and drops empty tokens.
func eqnTokens(statement string) []string {
	var tokens []string
	for _, t := range strings.Split(statement, "=") {
		if strings.TrimSpace(t) != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

// Io_ReadEqnStrCompact removes white space.
func Io_ReadEqnStrCompact(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, s)
}

// Io_ReadEqnStrCutAt splits s at any of the stop characters (strtok),
// keeping only first occurrences when unique is set.
func Io_ReadEqnStrCutAt(s, stops string, unique bool) []string {
	var tokens []string
	seen := map[string]bool{}
	for _, t := range strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(stops, r) }) {
		if unique && seen[t] {
			continue
		}
		seen[t] = true
		tokens = append(tokens, t)
	}
	return tokens
}

func ReadEqn(filename string, check bool) (*abc.Network, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseEqn(filename, string(data), check)
}
func Io_ReadEqn(filename string, check int32) *abc.Network {
	n, err := ReadEqn(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
func init() { RegisterFormat(IO_FILE_EQN, ReadEqn, nil) }
