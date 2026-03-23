//go:build ignore

// patch_parser.go patches the goyacc-generated grammar_v17.go to always
// read the lookahead token before performing default reductions. This
// ensures trace ordering matches Python PLY's behavior, where the parser
// always reads the next token before deciding to reduce.
package main

import (
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("grammar_v17.go")
	if err != nil {
		panic(err)
	}
	src := string(data)

	// The goyacc default reduction section looks like:
	//
	//   v17default:
	//       /* default state action */
	//       v17n = int(v17Def[v17state])
	//       if v17n == -2 {
	//           if v17rcvr.char < 0 {
	//               v17rcvr.char, v17token = v17lex1(v17lex, &v17rcvr.lval)
	//           }
	//
	// We patch it to always read the lookahead BEFORE checking the default action:
	//
	//   v17default:
	//       /* default state action */
	//       v17n = int(v17Def[v17state])
	//       // Always read lookahead before reducing, to match PLY trace order.
	//       if v17rcvr.char < 0 {
	//           v17rcvr.char, v17token = v17lex1(v17lex, &v17rcvr.lval)
	//       }
	//       if v17n == -2 {

	old := `v17n = int(v17Def[v17state])
	if v17n == -2 {
		if v17rcvr.char < 0 {
			v17rcvr.char, v17token = v17lex1(v17lex, &v17rcvr.lval)
		}`

	new := `v17n = int(v17Def[v17state])
	// Always read lookahead before reducing, to match PLY trace order.
	if v17rcvr.char < 0 {
		v17rcvr.char, v17token = v17lex1(v17lex, &v17rcvr.lval)
	}
	if v17n == -2 {`

	if !strings.Contains(src, old) {
		// Already patched or format changed
		return
	}

	src = strings.Replace(src, old, new, 1)
	err = os.WriteFile("grammar_v17.go", []byte(src), 0644)
	if err != nil {
		panic(err)
	}
}
