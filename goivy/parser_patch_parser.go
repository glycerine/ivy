//go:build ignore

// patch_parser.go patches the goyacc-generated full-file parsers to always
// read the lookahead token before performing default reductions. This
// ensures trace ordering matches Python PLY's behavior, where the parser
// always reads the next token before deciding to reduce.
package main

import (
	"os"
	"strings"
)

func main() {
	patch("parser_grammar_v17.go", "parser17")
	patch("parser_grammar_v16.go", "parser16")
}

func patch(path, prefix string) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	src := string(data)

	// The goyacc default reduction section looks like:
	//
	//   parser17default:
	//       /* default state action */
	//       parser17n = int(parser17Def[parser17state])
	//       if parser17n == -2 {
	//           if parser17rcvr.char < 0 {
	//               parser17rcvr.char, parser17token = parser17lex1(parser17lex, &parser17rcvr.lval)
	//           }
	//
	// We patch it to always read the lookahead BEFORE checking the default action:
	//
	//   parser17default:
	//       /* default state action */
	//       parser17n = int(parser17Def[parser17state])
	//       // Always read lookahead before reducing, to match PLY trace order.
	//       if parser17rcvr.char < 0 {
	//           parser17rcvr.char, parser17token = parser17lex1(parser17lex, &parser17rcvr.lval)
	//       }
	//       if parser17n == -2 {

	old := prefix + `n = int(` + prefix + `Def[` + prefix + `state])
	if ` + prefix + `n == -2 {
		if ` + prefix + `rcvr.char < 0 {
			` + prefix + `rcvr.char, ` + prefix + `token = ` + prefix + `lex1(` + prefix + `lex, &` + prefix + `rcvr.lval)
		}`

	new := prefix + `n = int(` + prefix + `Def[` + prefix + `state])
	// Always read lookahead before reducing, to match PLY trace order.
	if ` + prefix + `rcvr.char < 0 {
		` + prefix + `rcvr.char, ` + prefix + `token = ` + prefix + `lex1(` + prefix + `lex, &` + prefix + `rcvr.lval)
	}
	if ` + prefix + `n == -2 {`

	if !strings.Contains(src, old) {
		// Already patched or format changed
		return
	}

	src = strings.Replace(src, old, new, 1)
	err = os.WriteFile(path, []byte(src), 0644)
	if err != nil {
		panic(err)
	}
}
