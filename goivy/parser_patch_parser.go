//go:build ignore

// patch_parser.go patches goyacc-generated full-file parsers to always
// read the lookahead token before performing default reductions. This
// ensures trace ordering matches Python PLY's behavior, where the parser
// always reads the next token before deciding to reduce.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	patchParser("parser_grammar_v16.go", "parser16")
	patchParser("parser_grammar_v17.go", "parser17")
}

func patchParser(filename, prefix string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	src := string(data)

	// The goyacc default reduction section looks like:
	//
	//   <prefix>default:
	//       /* default state action */
	//       <prefix>n = int(<prefix>Def[<prefix>state])
	//       if <prefix>n == -2 {
	//           if <prefix>rcvr.char < 0 {
	//               <prefix>rcvr.char, <prefix>token = <prefix>lex1(<prefix>lex, &<prefix>rcvr.lval)
	//           }
	//
	// We patch it to always read the lookahead BEFORE checking the default action:
	//
	//   <prefix>default:
	//       /* default state action */
	//       <prefix>n = int(<prefix>Def[<prefix>state])
	//       // Always read lookahead before reducing, to match PLY trace order.
	//       if <prefix>rcvr.char < 0 {
	//           <prefix>rcvr.char, <prefix>token = <prefix>lex1(<prefix>lex, &<prefix>rcvr.lval)
	//       }
	//       if <prefix>n == -2 {

	old := fmt.Sprintf(`%[1]sn = int(%[1]sDef[%[1]sstate])
	if %[1]sn == -2 {
		if %[1]srcvr.char < 0 {
			%[1]srcvr.char, %[1]stoken = %[1]slex1(%[1]slex, &%[1]srcvr.lval)
		}`, prefix)

	new := fmt.Sprintf(`%[1]sn = int(%[1]sDef[%[1]sstate])
	// Always read lookahead before reducing, to match PLY trace order.
	if %[1]srcvr.char < 0 {
		%[1]srcvr.char, %[1]stoken = %[1]slex1(%[1]slex, &%[1]srcvr.lval)
	}
	if %[1]sn == -2 {`, prefix)

	if !strings.Contains(src, old) {
		// Already patched or format changed
		return
	}

	src = strings.Replace(src, old, new, 1)
	err = os.WriteFile(filename, []byte(src), 0644)
	if err != nil {
		panic(err)
	}
}
