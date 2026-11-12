package io

import (
	"fmt"
	stdio "io"
	"os"
	"time"

	"github.com/glycerine/abcgo/base/abc"
)

type plaProduct struct {
	inputs  string
	outputs string
}

func writePLARegistered(network *abc.Network, filename string) error {
	if network == nil {
		return fmt.Errorf("nil network")
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = fmt.Fprintf(file, "# Benchmark %q written by ABC on %s\n", network.Name, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}
	return writePLA(file, network, false, -1, 20)
}

func writeMOPLARegistered(network *abc.Network, filename string) error {
	if network == nil {
		return fmt.Errorf("nil network")
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = fmt.Fprintf(file, "# Benchmark %q written by ABC on %s\n", network.Name, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}
	return writePLA(file, network, true, -1, 16)
}

// Io_WritePlaOne writes a one-hot-output PLA body to an already-open file.
func Io_WritePlaOne(file *os.File, network *Abc_Ntk_t) int32 {
	if file == nil || network == nil || abc.Abc_NtkIsSopNetlist(network) == 0 {
		return 0
	}
	if err := writePLA(file, network, false, -1, 20); err != nil {
		fmt.Fprintf(os.Stderr, "Io_WritePlaOne: %v\n", err)
		return 0
	}
	return 1
}

func Io_WritePla(network *Abc_Ntk_t, filename string) int32 {
	if network == nil || abc.Abc_NtkIsSopNetlist(network) == 0 {
		return 0
	}
	if err := writePLARegistered(network, filename); err != nil {
		fmt.Fprintf(os.Stderr, "Io_WritePla: %v\n", err)
		return 0
	}
	return 1
}

// Io_WriteMoPlaOne writes a multi-output PLA body. It uses complete minterms,
// so it works for SOP and AIG networks without a CUDD dependency.
func Io_WriteMoPlaOne(file *os.File, network *Abc_Ntk_t) int32 {
	if file == nil || network == nil {
		return 0
	}
	if err := writePLA(file, network, true, -1, 16); err != nil {
		fmt.Fprintf(os.Stderr, "Io_WriteMoPlaOne: %v\n", err)
		return 0
	}
	return 1
}

func Io_WriteMoPla(network *Abc_Ntk_t, filename string) int32 {
	if err := writeMOPLARegistered(network, filename); err != nil {
		fmt.Fprintf(os.Stderr, "Io_WriteMoPla: %v\n", err)
		return 0
	}
	return 1
}

// Io_WriteMoPlaM writes up to nMints on-set minterms for the first output.
func Io_WriteMoPlaM(network *Abc_Ntk_t, filename string, nMints int32) int32 {
	if network == nil || nMints < 0 {
		return 0
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Io_WriteMoPlaM: %v\n", err)
		return 0
	}
	defer file.Close()
	if _, err = fmt.Fprintf(file, "# Benchmark %q written by ABC on %s\n", network.Name, time.Now().Format(time.RFC3339)); err != nil {
		return 0
	}
	if err = writePLA(file, network, true, int(nMints), 16); err != nil {
		fmt.Fprintf(os.Stderr, "Io_WriteMoPlaM: %v\n", err)
		return 0
	}
	return 1
}

func writePLA(dst stdio.Writer, network *abc.Network, multiOutput bool, mintermLimit, maxInputs int) error {
	inputs := network.ObjectsOfKind(abc.ObjectPI)
	outputs := network.ObjectsOfKind(abc.ObjectPO)
	if len(inputs) > maxInputs {
		return fmt.Errorf("PLA writer supports at most %d inputs (got %d)", maxInputs, len(inputs))
	}
	if len(outputs) == 0 {
		return fmt.Errorf("network has no primary outputs")
	}
	selectedOutputs := outputs
	if mintermLimit >= 0 {
		selectedOutputs = outputs[:1]
		multiOutput = true
	}
	if _, err := fmt.Fprintf(dst, ".i %d\n.o %d\n.ilb", len(inputs), len(selectedOutputs)); err != nil {
		return err
	}
	for _, input := range inputs {
		if _, err := fmt.Fprintf(dst, " %s", plaPortName(input, "x")); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(dst, "\n.ob"); err != nil {
		return err
	}
	for _, output := range selectedOutputs {
		if _, err := fmt.Fprintf(dst, " %s", plaPortName(output, "z")); err != nil {
			return err
		}
	}

	limit := 1 << uint(len(inputs))
	products := make([]plaProduct, 0)
	for assignment := 0; assignment < limit; assignment++ {
		values := make(map[*abc.Object]bool, len(inputs))
		pattern := make([]byte, len(inputs))
		for i, input := range inputs {
			value := assignment&(1<<uint(i)) != 0
			values[input] = value
			pattern[i] = '0'
			if value {
				pattern[i] = '1'
			}
		}
		outputBits := make([]byte, len(selectedOutputs))
		anyOn := false
		for i, output := range selectedOutputs {
			value, err := network.Evaluate(output.Signal(), values)
			if err != nil {
				return fmt.Errorf("evaluate output %q at assignment %d: %w", output.Name(), assignment, err)
			}
			outputBits[i] = '0'
			if value {
				outputBits[i] = '1'
				anyOn = true
			}
		}
		if !anyOn {
			continue
		}
		if mintermLimit >= 0 {
			if len(products) >= mintermLimit {
				break
			}
			products = append(products, plaProduct{inputs: string(pattern), outputs: "1"})
		} else if multiOutput {
			products = append(products, plaProduct{inputs: string(pattern), outputs: string(outputBits)})
		} else {
			for i, bit := range outputBits {
				if bit == '1' {
					oneHot := make([]byte, len(outputBits))
					for j := range oneHot {
						oneHot[j] = '0'
					}
					oneHot[i] = '1'
					products = append(products, plaProduct{inputs: string(pattern), outputs: string(oneHot)})
				}
			}
		}
	}
	if _, err := fmt.Fprintf(dst, "\n.p %d\n", len(products)); err != nil {
		return err
	}
	for _, product := range products {
		if _, err := fmt.Fprintf(dst, "%s %s\n", product.inputs, product.outputs); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(dst, ".e")
	return err
}

func plaPortName(object *abc.Object, prefix string) string {
	if object.Name() != "" {
		return object.Name()
	}
	return fmt.Sprintf("%s%d", prefix, object.ID)
}
