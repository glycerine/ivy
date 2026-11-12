package io

import (
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

func WriteBench(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	used := map[string]bool{}
	for _, o := range n.Objects() {
		name := blifObjectName(o)
		if strings.ContainsAny(name, "(),= \t\r\n") {
			return fmt.Errorf("signal name %q cannot be represented in BENCH", name)
		}
		used[name] = true
	}
	temporary := func() string {
		for i := 0; ; i++ {
			name := fmt.Sprintf("__bench_%d", i)
			if !used[name] {
				used[name] = true
				return name
			}
		}
	}
	var out strings.Builder
	constants := map[bool]string{}
	constant := func(value bool) string {
		if signal := constants[value]; signal != "" {
			return signal
		}
		signal := temporary()
		digit := 0
		if value {
			digit = 1
		}
		fmt.Fprintf(&out, "%s = LUT 0x%d()\n", signal, digit)
		constants[value] = signal
		return signal
	}
	fmt.Fprintf(&out, "# %s\n", n.Name)
	for _, port := range n.ObjectsOfKind(abc.ObjectPI) {
		fmt.Fprintf(&out, "INPUT(%s)\n", blifPortName(port, true))
	}
	for _, port := range n.ObjectsOfKind(abc.ObjectPO) {
		fmt.Fprintf(&out, "OUTPUT(%s)\n", blifPortName(port, false))
	}
	for _, box := range n.Boxes() {
		if box.Kind != abc.ObjectLatch {
			return fmt.Errorf("BENCH output requires flattened hierarchy")
		}
		fans, outs := box.Fanins(), box.Fanouts()
		if len(fans) != 1 || len(outs) != 1 {
			return fmt.Errorf("malformed latch")
		}
		input, output := blifPortNet(fans[0].Object, false), blifPortNet(outs[0], true)
		if input == nil || output == nil {
			return fmt.Errorf("latch has no net terminals")
		}
		fmt.Fprintf(&out, "%s = DFF(%s)\n", blifObjectName(output), blifObjectName(input))
	}
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		sop, ok := node.Data.(string)
		if !ok {
			return fmt.Errorf("BENCH requires SOP functions")
		}
		outputs := node.Fanouts()
		if len(outputs) != 1 {
			return fmt.Errorf("SOP node must drive one net")
		}
		output := blifObjectName(outputs[0])
		inputs := node.Fanins()
		var cubes []string
		polarity := ""
		for _, line := range strings.Split(strings.TrimSpace(sop), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			fields := strings.Fields(line)
			pattern, value := "", ""
			if len(fields) == 1 {
				value = fields[0]
			} else if len(fields) == 2 {
				pattern, value = fields[0], fields[1]
			} else {
				return fmt.Errorf("invalid SOP row")
			}
			if value != "0" && value != "1" {
				return fmt.Errorf("invalid SOP output phase")
			}
			if polarity != "" && polarity != value {
				return fmt.Errorf("mixed SOP phases")
			}
			polarity = value
			var literals []string
			if pattern != "" && len(pattern) != len(inputs) {
				return fmt.Errorf("SOP width mismatch")
			}
			for i, c := range pattern {
				if c == '-' {
					continue
				}
				if c != '0' && c != '1' {
					return fmt.Errorf("invalid SOP literal")
				}
				signal := blifObjectName(inputs[i].Object)
				inverted := inputs[i].Complement != (c == '0')
				if inverted {
					tmp := temporary()
					fmt.Fprintf(&out, "%s = NOT(%s)\n", tmp, signal)
					signal = tmp
				}
				literals = append(literals, signal)
			}
			if len(literals) == 0 {
				cubes = append(cubes, constant(true))
			} else if len(literals) == 1 {
				cubes = append(cubes, literals[0])
			} else {
				tmp := temporary()
				fmt.Fprintf(&out, "%s = AND(%s)\n", tmp, strings.Join(literals, ", "))
				cubes = append(cubes, tmp)
			}
		}
		if len(cubes) == 0 {
			fmt.Fprintf(&out, "%s = BUFF(%s)\n", output, constant(false))
		} else if len(cubes) == 1 {
			gate := "BUFF"
			if polarity == "0" {
				gate = "NOT"
			}
			fmt.Fprintf(&out, "%s = %s(%s)\n", output, gate, cubes[0])
		} else {
			gate := "OR"
			if polarity == "0" {
				gate = "NOR"
			}
			fmt.Fprintf(&out, "%s = %s(%s)\n", output, gate, strings.Join(cubes, ", "))
		}
	}
	return writeText(w, out.String())
}
func WriteBenchFile(n *abc.Network, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = WriteBench(file, n)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteBench(n *abc.Network, filename string) int32 {
	if err := WriteBenchFile(n, filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0
	}
	return 1
}
func Io_WriteBenchOne(w stdio.Writer, n *abc.Network) int32 {
	if err := WriteBench(w, n); err != nil {
		return 0
	}
	return 1
}
func Io_WriteBenchCheckNames(n *abc.Network) int32 {
	for _, o := range n.Objects() {
		if strings.ContainsAny(blifObjectName(o), "(),= \t\r\n") {
			return 0
		}
	}
	return 1
}
func init() { RegisterFormat(IO_FILE_BENCH, nil, WriteBenchFile) }
