//go:build ignore

// patch_parser.go patches the goyacc-generated parser_grammar_v17.go to always
// read the lookahead token before performing default reductions. This
// ensures trace ordering matches Python PLY's behavior, where the parser
// always reads the next token before deciding to reduce.
package main

import (
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("parser_grammar_v17.go")
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

	old := `parser17n = int(parser17Def[parser17state])
	if parser17n == -2 {
		if parser17rcvr.char < 0 {
			parser17rcvr.char, parser17token = parser17lex1(parser17lex, &parser17rcvr.lval)
		}`

	new := `parser17n = int(parser17Def[parser17state])
	// Always read lookahead before reducing, to match PLY trace order.
	if parser17rcvr.char < 0 {
		parser17rcvr.char, parser17token = parser17lex1(parser17lex, &parser17rcvr.lval)
	}
	if parser17n == -2 {`

	if !strings.Contains(src, old) {
		// Already patched or format changed
		return
	}

	src = strings.Replace(src, old, new, 1)
	err = os.WriteFile("parser_grammar_v17.go", []byte(src), 0644)
	if err != nil {
		panic(err)
	}
}
