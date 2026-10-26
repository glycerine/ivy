package abc

// ANSI color diagnostic rendering from src/misc/util/utilColor.c.

import (
	"fmt"
	"strings"
)

// ColorTestANSI returns the exact non-Windows output of Abc_ColorTest without
// mutating process-global stdout or terminal state.
func ColorTestANSI() string {
	var output strings.Builder
	output.WriteString("Background       ")
	for background := 0; background < 8; background++ {
		fmt.Fprintf(&output, "  [1;4%dm", background)
	}
	output.WriteByte('\n')
	for foreground := 0; foreground < 2; foreground++ {
		fmt.Fprintf(&output, "Foreground [%dm   ", foreground)
		for style := 0; style < 8; style++ {
			fmt.Fprintf(&output, "  \x1b[%d;3%dm\x1b[%dm  Hi  \x1b[0m", foreground&1, foreground>>1, style)
		}
		output.WriteByte('\n')
	}
	for foreground := 0; foreground < 16; foreground++ {
		fmt.Fprintf(&output, "Foreground [%d;3%dm", foreground&1, foreground>>1)
		for background := 0; background < 8; background++ {
			fmt.Fprintf(&output, "  \x1b[%d;3%dm\x1b[1;4%dm  Hi  \x1b[0m", foreground&1, foreground>>1, background)
		}
		output.WriteByte('\n')
	}
	output.WriteString("\x1b[4mUnderlined\x1b[0m\n")
	output.WriteString("\x1b[5mBlinking  \x1b[0m\n")
	output.WriteString("\x1b[7mInverted  \x1b[0m\n")
	output.WriteString("\x1b[8mConcealed \x1b[0m\n")
	return output.String()
}
