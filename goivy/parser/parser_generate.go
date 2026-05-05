package parser

//go:generate goyacc -o parser_grammar_v17.go -p parser17 grammar_v17.y
//go:generate go run parser_patch_parser.go
