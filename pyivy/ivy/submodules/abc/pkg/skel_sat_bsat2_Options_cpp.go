package abc

// Hand-checked skeleton of src/sat/bsat2/Options.cpp (namespace Minisat).

// MinisatParseOptions is Minisat::parseOptions; consumed options are removed from argv and argc.
// C: int Minisat::parseOptions(int& argc, char** argv, bool strict)
func MinisatParseOptions(argc *int, argv []string, strict bool) int {
	panic("TODO, stubbed")
}

// MinisatSetUsageHelp is Minisat::setUsageHelp.
// C: void Minisat::setUsageHelp(const char* str)
func MinisatSetUsageHelp(str string) {
	panic("TODO, stubbed")
}

// MinisatSetHelpPrefixStr is Minisat::setHelpPrefixStr.
// C: void Minisat::setHelpPrefixStr(const char* str)
func MinisatSetHelpPrefixStr(str string) {
	panic("TODO, stubbed")
}

// MinisatPrintUsageAndExit is Minisat::printUsageAndExit (verbose defaults to false).
// C: int Minisat::printUsageAndExit(int argc, char** argv, bool verbose)
func MinisatPrintUsageAndExit(argc int, argv []string, verbose bool) int {
	panic("TODO, stubbed")
}
