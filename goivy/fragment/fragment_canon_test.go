package fragment

import (
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	uf "github.com/glycerine/ivy/goivy/unionfind"
)

// testFixtures holds shared test data for fragment canon tests.
type testFixtures struct {
	S   *lg.UninterpretedSort
	X   *lg.Variable
	Y   *lg.Variable
	eq  *lg.Eq
	fs  *lg.FunctionSort
	sym *lg.Const
	n0  *uf.UFNode
	n1  *uf.UFNode
	n2  *uf.UFNode
}

func makeTestFixtures(t *testing.T) testFixtures {
	t.Helper()
	uf.ResetCounter()

	S := &lg.UninterpretedSort{Name: "S"}
	X, err := lg.NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	Y, err := lg.NewVariable("Y", S)
	if err != nil {
		t.Fatal(err)
	}
	eq, err := lg.NewEq(X, Y)
	if err != nil {
		t.Fatal(err)
	}
	fs, err := lg.NewFunctionSort(S, S, lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	sym := lg.NewConst("f", fs)

	n0 := uf.NewUFNode() // id 0
	n1 := uf.NewUFNode() // id 1
	n2 := uf.NewUFNode() // id 2

	return testFixtures{S: S, X: X, Y: Y, eq: eq, fs: fs, sym: sym, n0: n0, n1: n1, n2: n2}
}

// buildAllFragmentSexps constructs all fragment canon vectors and returns id -> sexp.
func buildAllFragmentSexps(t *testing.T) map[string]string {
	t.Helper()
	f := makeTestFixtures(t)

	sortEqSym := lg.NewConst("=", f.S)

	m := map[string]string{
		// UFNode
		"ufnode_basic":     ufNodeSexp(f.n0),
		"ufnode_nil":       ufNodeSexp(nil),
		"ufnode_set_empty": ufNodeSetSexp(map[*uf.UFNode]bool{}),
		"ufnode_set_three": ufNodeSetSexp(map[*uf.UFNode]bool{f.n0: true, f.n1: true, f.n2: true}),
		"ufnode_set_nil":   ufNodeSetSexp(nil),

		// varID
		"varID_simple": string(varID{name: "X", sort: "S"}.Sexp()),
		"varID_other":  string(varID{name: "Y", sort: "S"}.Sexp()),

		// strat_key (Go varKey/appKey/eqExprKey)
		"strat_key_var":  string(varKey(f.X)),
		"strat_key_app":  string(appKey(f.sym, 2)),
		"strat_key_expr": string(eqExprKey(f.X)),

		// stratEntry
		"strat_entry_var": string((&stratEntry{v: f.X}).Sexp()),
		"strat_entry_app": string((&stratEntry{sym: f.sym, idx: 1}).Sexp()),
		"strat_entry_sort": string((&stratEntry{
			sym: sortEqSym, isSort: true,
		}).Sexp()),

		// arc
		"arc_with_idx": string((&arc{
			from: f.n0, to: f.n1, fmla: f.eq, lineno: 42, argIdx: 1, hasIdx: true,
		}).Sexp()),
		"arc_no_idx": string((&arc{
			from: f.n0, to: f.n1, fmla: f.eq, lineno: 42, argIdx: -1, hasIdx: false,
		}).Sexp()),
		"arc_nil_fmla": string((&arc{
			from: f.n0, to: f.n1, fmla: nil, lineno: 10, argIdx: -1, hasIdx: false,
		}).Sexp()),
		"arc_nil_nodes": string((&arc{
			from: nil, to: nil, fmla: f.eq, lineno: 5, argIdx: -1, hasIdx: false,
		}).Sexp()),

		// mapFmlaRes
		"map_fmla_res_basic": string((&mapFmlaRes{
			node: f.n0, uvs: map[*uf.UFNode]bool{f.n1: true, f.n2: true},
		}).Sexp()),
		"map_fmla_res_nil": string((&mapFmlaRes{
			node: nil, uvs: nil,
		}).Sexp()),
		"map_fmla_res_empty_uvs": string((&mapFmlaRes{
			node: f.n1, uvs: map[*uf.UFNode]bool{},
		}).Sexp()),

		// skolemEntry
		"skolem_entry_basic": string((&skolemEntry{
			fmla: f.eq, ast: f.X,
		}).Sexp()),
		"skolem_entry_nil": string((&skolemEntry{}).Sexp()),

		// fmlaPair
		"fmla_pair_basic": string((&fmlaPair{
			fmla: f.eq, source: f.X, lineno: 15,
		}).Sexp()),
		"fmla_pair_nil_source": string((&fmlaPair{
			fmla: f.eq, source: nil, lineno: 7,
		}).Sexp()),
		"fmla_pair_nil_fmla": string((&fmlaPair{}).Sexp()),

		// FragmentError
		"fragment_error": string((&FragmentError{
			Message: "test error message",
		}).Sexp()),
	}
	return m
}

// --- Test A: Canon() == Canonical(Sexp()) ---

func TestFragmentCanonEqualsSexp(t *testing.T) {
	f := makeTestFixtures(t)

	tests := []struct {
		name  string
		sexp  lg.NodeKey
		canon iu.Canonical
	}{
		{"FragmentError", (&FragmentError{Message: "x"}).Sexp(), (&FragmentError{Message: "x"}).Canon()},
		{"stratEntry", (&stratEntry{v: f.X}).Sexp(), (&stratEntry{v: f.X}).Canon()},
		{"arc", (&arc{from: f.n0, to: f.n1, fmla: f.eq, lineno: 1}).Sexp(),
			(&arc{from: f.n0, to: f.n1, fmla: f.eq, lineno: 1}).Canon()},
		{"varID", varID{name: "X", sort: "S"}.Sexp(), varID{name: "X", sort: "S"}.Canon()},
		{"macroDef", (&macroDef{}).Sexp(), (&macroDef{}).Canon()},
		{"mapFmlaRes", (&mapFmlaRes{node: f.n0}).Sexp(), (&mapFmlaRes{node: f.n0}).Canon()},
		{"skolemEntry", (&skolemEntry{fmla: f.eq}).Sexp(), (&skolemEntry{fmla: f.eq}).Canon()},
		{"fmlaPair", (&fmlaPair{fmla: f.eq, lineno: 5}).Sexp(), (&fmlaPair{fmla: f.eq, lineno: 5}).Canon()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if iu.Canonical(tc.sexp) != tc.canon {
				t.Errorf("Canon != Canonical(Sexp) for %s:\n  sexp:  %s\n  canon: %s",
					tc.name, tc.sexp, tc.canon)
			}
		})
	}
}

// --- Test B: Sexp() is deterministic ---

func TestFragmentSexpDeterministic(t *testing.T) {
	f := makeTestFixtures(t)

	objects := []struct {
		name string
		fn   func() string
	}{
		{"FragmentError", func() string { return string((&FragmentError{Message: "m"}).Sexp()) }},
		{"stratEntry", func() string { return string((&stratEntry{sym: f.sym, idx: 3, v: f.X}).Sexp()) }},
		{"arc", func() string {
			return string((&arc{from: f.n0, to: f.n1, fmla: f.eq, lineno: 1, argIdx: 2, hasIdx: true}).Sexp())
		}},
		{"varID", func() string { return string(varID{name: "X", sort: "S"}.Sexp()) }},
		{"mapFmlaRes", func() string {
			return string((&mapFmlaRes{
				node: f.n0,
				uvs:  map[*uf.UFNode]bool{f.n1: true, f.n2: true},
			}).Sexp())
		}},
		{"skolemEntry", func() string { return string((&skolemEntry{fmla: f.eq, ast: f.X}).Sexp()) }},
		{"fmlaPair", func() string { return string((&fmlaPair{fmla: f.eq, source: f.X, lineno: 7}).Sexp()) }},
	}

	for _, tc := range objects {
		t.Run(tc.name, func(t *testing.T) {
			s1 := tc.fn()
			s2 := tc.fn()
			if s1 != s2 {
				t.Errorf("Non-deterministic Sexp for %s:\n  first:  %s\n  second: %s", tc.name, s1, s2)
			}
		})
	}
}

// --- Test C: Expected string values ---

func TestFragmentSexpExpectedStrings(t *testing.T) {
	sexps := buildAllFragmentSexps(t)

	// Spot-check a few simple ones against hardcoded expectations.
	checks := map[string]string{
		"ufnode_basic":     "(ufNode id:0)",
		"ufnode_nil":       "nil",
		"ufnode_set_empty": "[]",
		"ufnode_set_three": "[0 1 2]",
		"ufnode_set_nil":   "nil",
		"varID_simple":     `(varID name:"X" sort:"S")`,
		"fragment_error":   `(fragmentError message:"test error message")`,
	}

	for id, expected := range checks {
		got, ok := sexps[id]
		if !ok {
			t.Errorf("Missing vector %q", id)
			continue
		}
		if got != expected {
			t.Errorf("Vector %s:\n  expected: %s\n  got:      %s", id, expected, got)
		}
	}
}

// --- Test D: Cross-language comparison ---

func TestFragmentFragmentSexpCrossLanguage(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}

	goSexps := buildAllFragmentSexps(t)

	_, thisFile, _, _ := runtime.Caller(0)
	pyScript := filepath.Join(filepath.Dir(thisFile), "..", "pytesthelper", "emit_fragment_sexp.py")

	cmd := exec.Command("python3", pyScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Python subprocess failed: %v\nOutput: %s", err, out)
	}

	pyMap := parseTSV(string(out))

	for id, goSexp := range goSexps {
		pySexp, ok := pyMap[id]
		if !ok {
			t.Errorf("Python missing vector %q", id)
			continue
		}
		if goSexp != pySexp {
			t.Errorf("Cross-language mismatch for %s:\n  Go: %s\n  Py: %s", id, goSexp, pySexp)
		}
	}

	for id := range pyMap {
		if _, ok := goSexps[id]; !ok {
			t.Errorf("Python has vector %q but Go does not", id)
		}
	}
}

func parseTSV(data string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		m[parts[0]] = parts[1]
	}
	return m
}

// --- Fuzz tests ---

// randomSort generates a random Sort from seed.
func randomSort(rng *rand.Rand) lg.Sort {
	switch rng.Intn(3) {
	case 0:
		return &lg.UninterpretedSort{Name: "S"}
	case 1:
		return lg.Boolean
	default:
		return lg.TopS
	}
}

// randomVariable generates a random Variable.
func randomVariable(rng *rand.Rand) *lg.Variable {
	names := []string{"X", "Y", "Z", "W", "V"}
	name := names[rng.Intn(len(names))]
	v, _ := lg.NewVariable(name, randomSort(rng))
	return v
}

// randomConst generates a random Const.
func randomConst(rng *rand.Rand) *lg.Const {
	names := []string{"f", "g", "h", "p", "q"}
	return lg.NewConst(names[rng.Intn(len(names))], randomSort(rng))
}

// randomExpr generates a random Expr (variable, const, or eq).
func fragmentRandomExpr(rng *rand.Rand) lg.Expr {
	switch rng.Intn(4) {
	case 0:
		return randomVariable(rng)
	case 1:
		return randomConst(rng)
	case 2:
		v1 := randomVariable(rng)
		v2, _ := lg.NewVariable(v1.Name, v1.VSort) // same sort for Eq
		e, err := lg.NewEq(v1, v2)
		if err != nil {
			return v1
		}
		return e
	default:
		return nil
	}
}

// randomUFNode creates a UFNode or returns nil.
func randomUFNode(rng *rand.Rand) *uf.UFNode {
	if rng.Intn(4) == 0 {
		return nil
	}
	return uf.NewUFNode()
}

type sexpable interface {
	Sexp() lg.NodeKey
	Canon() iu.Canonical
}

// randomFragmentStruct generates a random fragment struct from a seed.
func randomFragmentStruct(seed uint64) sexpable {
	rng := rand.New(rand.NewSource(int64(seed)))
	switch rng.Intn(7) {
	case 0: // FragmentError
		msgs := []string{"err", "test", "fragment fail", "some message with special chars: <>&"}
		return &FragmentError{Message: msgs[rng.Intn(len(msgs))]}
	case 1: // varID
		v := varID{
			name: []string{"X", "Y", "Z"}[rng.Intn(3)],
			sort: []string{"S", "T", "Boolean"}[rng.Intn(3)],
		}
		return &v
	case 2: // stratEntry
		se := &stratEntry{
			idx:    rng.Intn(5),
			isSort: rng.Intn(2) == 0,
		}
		if rng.Intn(2) == 0 {
			se.sym = randomConst(rng)
		}
		if rng.Intn(2) == 0 {
			se.v = randomVariable(rng)
		}
		return se
	case 3: // arc
		return &arc{
			from:   randomUFNode(rng),
			to:     randomUFNode(rng),
			fmla:   fragmentRandomExpr(rng),
			lineno: rng.Intn(1000),
			argIdx: rng.Intn(10) - 1,
			hasIdx: rng.Intn(2) == 0,
		}
	case 4: // mapFmlaRes
		r := &mapFmlaRes{node: randomUFNode(rng)}
		if rng.Intn(2) == 0 {
			r.uvs = make(map[*uf.UFNode]bool)
			for i := 0; i < rng.Intn(4); i++ {
				n := uf.NewUFNode()
				r.uvs[n] = true
			}
		}
		return r
	case 5: // skolemEntry
		return &skolemEntry{
			fmla: fragmentRandomExpr(rng),
			ast:  fragmentRandomExpr(rng),
		}
	default: // fmlaPair
		fp := &fmlaPair{
			fmla:   fragmentRandomExpr(rng),
			lineno: rng.Intn(500),
		}
		if rng.Intn(2) == 0 {
			fp.source = randomVariable(rng)
		}
		return fp
	}
}

func FuzzFragmentSexpDeterministic(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(1))
	f.Add(uint64(42))
	f.Add(uint64(12345))
	f.Add(uint64(99999))

	f.Fuzz(func(t *testing.T, seed uint64) {
		uf.ResetCounter()
		obj1 := randomFragmentStruct(seed)
		s1 := string(obj1.Sexp())

		uf.ResetCounter()
		obj2 := randomFragmentStruct(seed)
		s2 := string(obj2.Sexp())

		if s1 != s2 {
			t.Errorf("Non-deterministic Sexp for seed %d:\n  first:  %s\n  second: %s", seed, s1, s2)
		}
	})
}

func FuzzFragmentSexpBalancedParens(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(1))
	f.Add(uint64(42))
	f.Add(uint64(12345))

	f.Fuzz(func(t *testing.T, seed uint64) {
		uf.ResetCounter()
		obj := randomFragmentStruct(seed)
		s := string(obj.Sexp())

		depth := 0
		bracketDepth := 0
		for _, ch := range s {
			switch ch {
			case '(':
				depth++
			case ')':
				depth--
			case '[':
				bracketDepth++
			case ']':
				bracketDepth--
			}
			if depth < 0 {
				t.Fatalf("Unbalanced parens (depth<0) in: %s", s)
			}
			if bracketDepth < 0 {
				t.Fatalf("Unbalanced brackets (depth<0) in: %s", s)
			}
		}
		if depth != 0 {
			t.Fatalf("Unbalanced parens (final depth %d) in: %s", depth, s)
		}
		if bracketDepth != 0 {
			t.Fatalf("Unbalanced brackets (final depth %d) in: %s", bracketDepth, s)
		}
	})
}

func FuzzFragmentSexpCanonEquality(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(1))
	f.Add(uint64(42))
	f.Add(uint64(12345))

	f.Fuzz(func(t *testing.T, seed uint64) {
		uf.ResetCounter()
		obj := randomFragmentStruct(seed)
		sexp := obj.Sexp()
		canon := obj.Canon()
		if iu.Canonical(sexp) != canon {
			t.Errorf("Canon != Canonical(Sexp) for seed %d:\n  sexp:  %s\n  canon: %s",
				seed, sexp, canon)
		}
	})
}
