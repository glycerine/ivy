package startingpoint

// Retained for existing callers; the catalog-driven engine owns library rewrites.
func rewriteLibraryCalls(source string, addImport bool) (string, error) {
	draft, _, err := rewriteStandardLibrary(source, addImport)
	return draft, err
}
