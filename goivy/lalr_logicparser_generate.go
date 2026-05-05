package goivy

//go:generate goyacc -o lalr_logicparser_grammar_v17.go -p lalr17 lalr_logicparser/grammar_v17.y
//go:generate goyacc -o v16_grammar_v16.go -p lalr16 lalr_logicparser/v16/grammar_v16.y
//go:generate goyacc -o v12_grammar_v12.go -p lalr12 lalr_logicparser/v12/grammar_v12.y
