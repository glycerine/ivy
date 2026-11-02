package abc

// Hand-checked skeleton of src/sat/glucose2/Options2.cpp (namespace Gluco2).

// Gluco2ParseOptions is Gluco2::parseOptions (returns void in this version); consumed options are removed from argv and argc.
// C: void Gluco2::parseOptions(int& argc, char** argv, bool strict)
func Gluco2ParseOptions(argc *int, argv []string, strict bool) {
	panic("TODO, stubbed")
}

// Gluco2SetUsageHelp is Gluco2::setUsageHelp.
// C: void Gluco2::setUsageHelp(const char* str)
func Gluco2SetUsageHelp(str string) {
	panic("TODO, stubbed")
}

// Gluco2SetHelpPrefixStr is Gluco2::setHelpPrefixStr.
// C: void Gluco2::setHelpPrefixStr(const char* str)
func Gluco2SetHelpPrefixStr(str string) {
	panic("TODO, stubbed")
}

// Gluco2PrintUsageAndExit is Gluco2::printUsageAndExit (returns void in this version; verbose defaults to false).
// C: void Gluco2::printUsageAndExit(int argc, char** argv, bool verbose)
func Gluco2PrintUsageAndExit(argc int, argv []string, verbose bool) {
	panic("TODO, stubbed")
}
