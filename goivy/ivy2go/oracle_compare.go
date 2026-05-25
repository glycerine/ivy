package ivy2go

// oracle_compare.go is a placeholder for the post-MVP behavioural
// oracle harness described in ARCHITECTURE_TODO.md §3.8 Tier 4.
//
// The harness will: build the ivy2cpp C++ output and the ivy2go Go
// output for the same fixture, run both with the same .test.args /
// .in inputs, and diff stdout. The Stub interface here is the seam
// the harness will plug into.
//
// M10 ships this file so the name and shape exist; M11+ implements
// the comparison.

// CompareGoVsCpp is the public entry point the future oracle harness
// will call. The current stub returns "not implemented" so any caller
// that mistakes the placeholder for a finished harness fails loudly.
func CompareGoVsCpp(goRunDir, cppRunDir, inputs string) error {
	_ = goRunDir
	_ = cppRunDir
	_ = inputs
	return errOracleNotYetImplemented
}

// errOracleNotYetImplemented is the canonical error returned by the
// stub. Promoted to a typed sentinel so callers can errors.Is-match
// it once the real implementation lands.
var errOracleNotYetImplemented = oracleStubError("ivy2go: behavioural oracle harness deferred to M11+")

type oracleStubError string

func (e oracleStubError) Error() string { return string(e) }
