package abc

// Hand-checked skeleton of src/sat/glucose/Options.cpp (namespace Gluco).

// GlucoParseOptions is Gluco::parseOptions (returns void in this version); consumed options are removed from argv and argc.
// C: void Gluco::parseOptions(int& argc, char** argv, bool strict)
func GlucoParseOptions(argc *int, argv []string, strict bool) {
	panic("TODO, stubbed")
}

// GlucoSetUsageHelp is Gluco::setUsageHelp.
// C: void Gluco::setUsageHelp(const char* str)
func GlucoSetUsageHelp(str string) {
	panic("TODO, stubbed")
}

// GlucoSetHelpPrefixStr is Gluco::setHelpPrefixStr.
// C: void Gluco::setHelpPrefixStr(const char* str)
func GlucoSetHelpPrefixStr(str string) {
	panic("TODO, stubbed")
}

// GlucoPrintUsageAndExit is Gluco::printUsageAndExit (returns void in this version; verbose defaults to false).
// C: void Gluco::printUsageAndExit(int argc, char** argv, bool verbose)
func GlucoPrintUsageAndExit(argc int, argv []string, verbose bool) {
	panic("TODO, stubbed")
}
