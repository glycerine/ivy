package lalr_logicparser

//go:generate goyacc -o lalr_logicparser_grammar_v17.go -p lalr17 grammar_v17.y
//go:generate goyacc -o v16/v16_grammar_v16.go -p lalr16 v16/grammar_v16.y
//go:generate goyacc -o v12/v12_grammar_v12.go -p lalr12 v12/grammar_v12.y
