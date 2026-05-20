package goivy

//go:generate goyacc -o parser_grammar_v17.go -p parser17 parser_grammar_v17.y
//go:generate goyacc -o parser_grammar_v16.go -p parser16 parser_grammar_v16.y
//go:generate go run parser_patch_parser.go
