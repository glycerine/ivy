package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"strings"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManDumpTests writes one compact digit sequence per test.
func Gia_ManDumpTests(tests *vec.Vec_Int_t, iterations int32, filename string) error {
	vals := vec.Vec_IntArray(tests)
	if iterations <= 0 || len(vals)%int(iterations) != 0 {
		return fmt.Errorf("bmc: invalid test dimensions")
	}
	width := len(vals) / int(iterations)
	var out strings.Builder
	for i, v := range vals {
		fmt.Fprint(&out, v)
		if (i+1)%width == 0 {
			out.WriteByte('\n')
		}
	}
	if width == 0 {
		for i := int32(0); i < iterations; i++ {
			out.WriteByte('\n')
		}
	}
	return os.WriteFile(filename, []byte(out.String()), 0666)
}

// Gia_ManDumpTestsSimulate replaces the input vector with next-state values.
func Gia_ManDumpTestsSimulate(p *gia.Gia_Man_t, values *vec.Vec_Int_t) {
	vals := vec.Vec_IntArray(values)
	inputs := make([]bool, len(vals))
	for i, v := range vals {
		if v != 0 && v != 1 {
			panic("bmc: non-binary simulation input")
		}
		inputs[i] = v != 0
	}
	outputs, err := p.Evaluate(inputs)
	if err != nil {
		panic(err)
	}
	vec.Vec_IntClear(values)
	for _, v := range outputs[gia.Gia_ManPoNum(p):] {
		if v {
			vec.Vec_IntPush(values, 1)
		} else {
			vec.Vec_IntPush(values, 0)
		}
	}
}

// Gia_ManDumpTestsDelay writes first-frame and second-frame input/state rows.
func Gia_ManDumpTestsDelay(tests *vec.Vec_Int_t, iterations int32, filename string, p *gia.Gia_Man_t) error {
	vals := vec.Vec_IntArray(tests)
	if p == nil || iterations <= 0 || len(vals)%int(iterations) != 0 {
		return fmt.Errorf("bmc: invalid test dimensions")
	}
	regs, pis := int(gia.Gia_ManRegNum(p)), int(gia.Gia_ManPiNum(p))
	width := 2*pis + regs
	if len(vals) != int(iterations)*width {
		return fmt.Errorf("bmc: invalid delay test width")
	}
	var out strings.Builder
	for i := 0; i < int(iterations); i++ {
		row := vals[i*width : (i+1)*width]
		state := vec.Vec_IntAlloc(pis + regs)
		for _, v := range row[regs : regs+pis] {
			fmt.Fprint(&out, v)
			vec.Vec_IntPush(state, v)
		}
		for _, v := range row[:regs] {
			fmt.Fprint(&out, v)
			vec.Vec_IntPush(state, v)
		}
		out.WriteByte('\n')
		Gia_ManDumpTestsSimulate(p, state)
		for _, v := range row[regs+pis:] {
			fmt.Fprint(&out, v)
		}
		for _, v := range vec.Vec_IntArray(state) {
			fmt.Fprint(&out, v)
		}
		out.WriteByte('\n')
	}
	return os.WriteFile(filename, []byte(out.String()), 0666)
}

// Gia_ManGetTestPatterns reads binary digits separated by ASCII whitespace.
func Gia_ManGetTestPatterns(filename string) *vec.Vec_Int_t {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Cannot open input file %q.\n", filename)
		return nil
	}
	result := vec.Vec_IntAlloc(len(data))
	for _, c := range data {
		switch c {
		case ' ', '\t', '\r', '\n':
		case '0', '1':
			vec.Vec_IntPush(result, int32(c-'0'))
		default:
			fmt.Printf("Wrong symbol (%c) in the input file.\n", c)
			return nil
		}
	}
	return result
}

// Gia_ManDeriveDup duplicates the graph and appends additional inputs.
func Gia_ManDeriveDup(p *gia.Gia_Man_t, extra int32) *gia.Gia_Man_t {
	if p == nil {
		return nil
	}
	if extra < 0 {
		panic("bmc: negative input count")
	}
	q := gia.Gia_ManDup(p)
	for i := int32(0); i < extra; i++ {
		gia.Gia_ManAppendCi(q)
	}
	return q
}

// Gia_ManPrintResults appends a fault-analysis summary to fault_stats.txt.
func Gia_ManPrintResults(p *gia.Gia_Man_t, solver *bsat.Solver, iterations int32, elapsed time.Duration) error {
	f, err := os.OpenFile("fault_stats.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	stats := solver.Statistics()
	_, writeErr := fmt.Fprintf(f, "%s %d %d %d %d %d %d %d %d %.2f\n", gia.Gia_ManName(p), gia.Gia_ManPiNum(p), gia.Gia_ManPoNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManAndNum(p), stats.Variables, stats.Clauses, stats.Conflicts, iterations, elapsed.Seconds())
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
