# to document all three test suites, from fastest to slowest: (see oracle_test.go)
test:
	XTRACE_OFF=1 go test -v -timeout=0 -count=1
	SLOWTEST=1 XTRACE_OFF=1 go test -v -timeout=0 -count=1
	ORACLE_TEST_TARGET_TEST=1 SLOWTEST=1 XTRACE_OFF=1 go test -v -timeout=0 -count=1
	ORACLE_TEST=1 SLOWTEST=1 XTRACE_OFF=1 go test -v -timeout=0 -count=1
