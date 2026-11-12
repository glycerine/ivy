package io

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/amap/formula"
)

// addEquation lowers the actual expression DAG into small SOP nodes, avoiding
// exponential truth-table enumeration for equations with many inputs.
func addEquation(m *blifModel, output, expression string) error {
	expr, err := formula.Parse(expression)
	if err != nil {
		return err
	}
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	variables := map[string]*hop.Hop_Obj_t{}
	names := map[*hop.Hop_Obj_t]string{}
	for i, name := range expr.Variables() {
		v := hop.Hop_IthVar(manager, int32(i))
		variables[name] = v
		names[v] = name
		m.net(name)
	}
	root, err := expr.ToHop(manager, variables)
	if err != nil {
		return err
	}
	var lower func(*hop.Hop_Obj_t) (string, error)
	lower = func(edge *hop.Hop_Obj_t) (string, error) {
		if name := names[edge]; name != "" {
			return name, nil
		}
		regular := hop.Hop_Regular(edge)
		var name string
		if hop.Hop_IsComplement(edge) != 0 {
			input, err := lower(regular)
			if err != nil {
				return "", err
			}
			name = m.uniqueNetName("__eqn_inv")
			if _, err = m.node([]string{input, name}, "0 1\n"); err != nil {
				return "", err
			}
		} else if hop.Hop_ObjIsConst1(regular) != 0 {
			name = m.uniqueNetName("__eqn_one")
			if _, err = m.node([]string{name}, " 1\n"); err != nil {
				return "", err
			}
		} else if hop.Hop_ObjIsNode(regular) != 0 {
			left, err := lower(hop.Hop_NotCond(hop.Hop_ObjFanin0(regular), hop.Hop_ObjFaninC0(regular)))
			if err != nil {
				return "", err
			}
			right, err := lower(hop.Hop_NotCond(hop.Hop_ObjFanin1(regular), hop.Hop_ObjFaninC1(regular)))
			if err != nil {
				return "", err
			}
			name = m.uniqueNetName("__eqn_gate")
			cover := "11 1\n"
			if hop.Hop_ObjIsExor(regular) != 0 {
				cover = "01 1\n10 1\n"
			}
			if _, err = m.node([]string{left, right, name}, cover); err != nil {
				return "", err
			}
		} else {
			return "", fmt.Errorf("unbound expression input")
		}
		names[edge] = name
		return name, nil
	}
	signal, err := lower(root)
	if err != nil {
		return err
	}
	_, err = m.node([]string{signal, output}, "1 1\n")
	return err
}
func ParseEqn(name, text string, check bool) (*abc.Network, error) {
	lines, err := blifLines(name, text)
	if err != nil {
		return nil, err
	}
	var source strings.Builder
	for _, line := range lines {
		source.WriteString(line.text)
		source.WriteByte('\n')
	}
	m := &blifModel{network: abc.NewNetwork(name), nets: map[string]*abc.Object{}, domains: map[string]*abc.MVVariable{}}
	m.network.Spec = name
	m.network.Type = uint8(abc.ABC_NTK_NETLIST)
	m.network.Function = uint8(abc.ABC_FUNC_SOP)
	statements := strings.Split(source.String(), ";")
	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if i == len(statements)-1 {
			return nil, fmt.Errorf("%s: missing equation semicolon", name)
		}
		parts := strings.SplitN(statement, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s: expected equation assignment", name)
		}
		left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch strings.ToUpper(left) {
		case "INORDER":
			for _, input := range strings.Fields(right) {
				if err = m.input(input); err != nil {
					return nil, err
				}
			}
		case "OUTORDER":
			for _, output := range strings.Fields(right) {
				if err = m.output(output); err != nil {
					return nil, err
				}
			}
		default:
			if len(strings.Fields(left)) != 1 {
				return nil, fmt.Errorf("%s: invalid equation output %s", name, left)
			}
			if err = addEquation(m, left, right); err != nil {
				return nil, fmt.Errorf("%s: equation %s: %w", name, left, err)
			}
		}
	}
	if check {
		for signal, net := range m.nets {
			if len(net.Fanins()) == 0 {
				return nil, fmt.Errorf("%s: undriven signal %s", name, signal)
			}
		}
		if err = m.network.Check(); err != nil {
			return nil, err
		}
		if _, err = m.network.Topological(m.network.ObjectsOfKind(abc.ObjectPO)); err != nil {
			return nil, err
		}
	}
	return m.network, nil
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
