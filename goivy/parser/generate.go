package parser

//go:generate goyacc -o grammar_v17.go -p v17 grammar_v17.y
//go:generate go run patch_parser.go
