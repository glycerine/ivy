package llb

import (
	"fmt"
	"os"
)

func Llb_ManGetDummyName(prefix string, number, digits int32) string {
	return fmt.Sprintf("%s%0*d", prefix, digits, number)
}

// Llb_ManDumpReached writes a BLIF truth-table representation for reached
// states. Truth-table export is bounded because BLIF's explicit cube format is
// exponential in the function support.
func Llb_ManDumpReached(dd *DdManager, reached *DdNode, model, filename string) {
	if dd == nil || reached == nil {
		return
	}
	variables := dd.Support(reached)
	if len(variables) > 24 {
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, ".model %s\n.inputs", model)
	for _, variable := range variables {
		fmt.Fprintf(file, " %s", Llb_ManGetDummyName("ff", int32(variable), 1))
	}
	fmt.Fprintln(file, "\n.outputs Reached")
	fmt.Fprint(file, ".names")
	for _, variable := range variables {
		fmt.Fprintf(file, " %s", Llb_ManGetDummyName("ff", int32(variable), 1))
	}
	fmt.Fprintln(file, " Reached")
	assignment := make([]bool, len(dd.Vars))
	for mask := 0; mask < 1<<len(variables); mask++ {
		for bit, variable := range variables {
			assignment[variable] = mask&(1<<bit) != 0
		}
		value, evalErr := dd.Evaluate(reached, assignment)
		if evalErr != nil || !value {
			continue
		}
		for bit := range variables {
			if assignment[variables[bit]] {
				fmt.Fprint(file, "1")
			} else {
				fmt.Fprint(file, "0")
			}
		}
		fmt.Fprintln(file, " 1")
	}
	fmt.Fprintln(file, ".end")
}
