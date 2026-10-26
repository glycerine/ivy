package abc

// Go-only tests for the remaining src/proof/acec arithmetic semantics.

import (
	"errors"
	"math/big"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func acecMakeXOR(t testing.TB, sop bool, invert bool) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	var root GiaLit
	if sop {
		x, e := g.AppendAnd(a, b.Not())
		if e != nil {
			t.Fatal(e)
		}
		y, e := g.AppendAnd(a.Not(), b)
		if e != nil {
			t.Fatal(e)
		}
		bothZero, e := g.AppendAnd(x.Not(), y.Not())
		if e != nil {
			t.Fatal(e)
		}
		root = bothZero.Not()
	} else {
		root, err = g.AppendXorReal(a, b)
		if err != nil {
			t.Fatal(err)
		}
	}
	if invert {
		root = root.Not()
	}
	if _, err = g.AppendCO(root); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestAcecPortInventoryCompleteAndOwned(t *testing.T) {
	got := AcecPortInventory()
	want := []string{"acec2Mult.c", "acecBo.c", "acecCl.c", "acecCo.c", "acecCore.c", "acecCover.c", "acecFadds.c", "acecMult.c", "acecPa.c", "acecPo.c", "acecPool.c", "acecRe.c", "acecSt.c", "acecStruct.c", "acecTree.c", "acecXor.c", "module.make"}
	if len(got) != len(want) {
		t.Fatalf("inventory length %d", len(got))
	}
	for i := range want {
		if got[i].Source != want[i] || got[i].Definitions == "" || got[i].Disposition == "" {
			t.Fatalf("unit %d: %+v", i, got[i])
		}
	}
	got[0].Source = "changed"
	if AcecPortInventory()[0].Source != want[0] {
		t.Fatal("inventory aliases caller storage")
	}
}

func TestAcecArithmeticEquivalentDifferentStructure(t *testing.T) {
	left, right := acecMakeXOR(t, false, false), acecMakeXOR(t, true, false)
	r, err := CheckAcecArithmeticEquivalent(left, right, AcecEquivalenceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Equivalent || !r.Proven || r.Witness != nil || r.Left.Digest != r.Right.Digest {
		t.Fatalf("result %+v", r)
	}
	if len(r.Left.Terms) != 3 {
		t.Fatalf("xor terms: %+v", r.Left.Terms)
	}
	clone := r.Left.Clone()
	clone.Terms[0].Coefficient.SetInt64(99)
	clone.Terms[0].Inputs[0] = 99
	if r.Left.Terms[0].Coefficient.Int64() == 99 || r.Left.Terms[0].Inputs[0] == 99 {
		t.Fatal("signature clone aliases")
	}
}

func TestAcecArithmeticCounterexampleAndSignedValues(t *testing.T) {
	left, right := acecMakeXOR(t, false, false), acecMakeXOR(t, false, true)
	r, err := CheckAcecArithmeticEquivalent(left, right, AcecEquivalenceOptions{Signed: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.Equivalent || !r.Proven || !slices.Equal(r.Witness, []bool{false, false}) {
		t.Fatalf("counterexample %+v", r)
	}
	if r.LeftValue.Cmp(big.NewInt(0)) != 0 || r.RightValue.Cmp(big.NewInt(-1)) != 0 {
		t.Fatalf("signed values %v %v", r.LeftValue, r.RightValue)
	}
	r.Witness[0] = true
	again, _ := CheckAcecArithmeticEquivalent(left, right, AcecEquivalenceOptions{Signed: true})
	if again.Witness[0] {
		t.Fatal("witness aliases prior result")
	}
}

func TestAcecArithmeticExhaustiveSmallCircuits(t *testing.T) {
	// Compare native and decomposed XOR for every input inversion and then
	// independently compare normalized values with direct bit-vector values.
	for mask := 0; mask < 4; mask++ {
		g1 := NewGiaManager()
		in1 := acecAppendInputs(t, g1, 2)
		a, b := in1[0], in1[1]
		if mask&1 != 0 {
			a = a.Not()
		}
		if mask&2 != 0 {
			b = b.Not()
		}
		x, e := g1.AppendXorReal(a, b)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = g1.AppendCO(x); e != nil {
			t.Fatal(e)
		}
		g2 := NewGiaManager()
		in2 := acecAppendInputs(t, g2, 2)
		a, b = in2[0], in2[1]
		if mask&1 != 0 {
			a = a.Not()
		}
		if mask&2 != 0 {
			b = b.Not()
		}
		p, e := g2.AppendAnd(a, b.Not())
		if e != nil {
			t.Fatal(e)
		}
		q, e := g2.AppendAnd(a.Not(), b)
		if e != nil {
			t.Fatal(e)
		}
		n, e := g2.AppendAnd(p.Not(), q.Not())
		if e != nil {
			t.Fatal(e)
		}
		if _, e = g2.AppendCO(n.Not()); e != nil {
			t.Fatal(e)
		}
		r, e := CheckAcecArithmeticEquivalent(g1, g2, AcecEquivalenceOptions{})
		if e != nil || !r.Equivalent {
			t.Fatalf("mask %d: %+v %v", mask, r, e)
		}
		for assignment := 0; assignment < 4; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0}
			out, e := g1.Evaluate(bits)
			if e != nil {
				t.Fatal(e)
			}
			values := make([]bool, g1.ObjectCount())
			for i, id := range g1.cis {
				values[id] = bits[i]
			}
			// Polynomial evaluation is independently checked through the public
			// normal form using input positions.
			v := new(big.Int)
			for _, term := range r.Left.Terms {
				present := true
				for _, i := range term.Inputs {
					present = present && bits[i]
				}
				if present {
					v.Add(v, term.Coefficient)
				}
			}
			if (v.Sign() != 0) != out[0] {
				t.Fatalf("mask %d assignment %d polynomial=%v output=%v", mask, assignment, v, out)
			}
		}
	}
}

func TestAcecArithmeticLimitsCancellationMalformedLifecycle(t *testing.T) {
	g := acecMakeXOR(t, false, false)
	cases := []AcecEquivalenceOptions{{MaxObjects: 1}, {MaxTerms: 1}, {MaxCoefficientBits: 1}, {MaxBytes: 1}, {MaxWork: 1}, {Stop: func() bool { return true }}}
	for i, o := range cases {
		if s, e := NormalizeAcecArithmetic(g, o); len(s.Terms) != 0 || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatalf("case %d: %+v %v", i, s, e)
		}
	}
	deep := acecMakeXOR(t, true, false)
	if s, e := NormalizeAcecArithmetic(deep, AcecEquivalenceOptions{MaxDepth: 1}); len(s.Terms) != 0 || !errors.Is(e, ErrLimit) {
		t.Fatalf("depth: %+v %v", s, e)
	}
	if _, e := NormalizeAcecArithmetic(nil, AcecEquivalenceOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatalf("nil: %v", e)
	}
	if r, e := CheckAcecArithmeticEquivalent(g, g, AcecEquivalenceOptions{MaxTerms: 3}); !errors.Is(e, ErrLimit) || r.Proven {
		t.Fatalf("cumulative terms: %+v %v", r, e)
	}
	bad := g.Clone()
	bad.objects[bad.cos[0]].word0 = GiaNone
	if s, e := NormalizeAcecArithmetic(bad, AcecEquivalenceOptions{}); len(s.Terms) != 0 || e == nil {
		t.Fatalf("malformed: %+v %v", s, e)
	}
	checker, e := NewAcecEquivalenceChecker(AcecEquivalenceOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if r, e := checker.Check(g, g); e != nil || !r.Equivalent {
		t.Fatalf("check: %+v %v", r, e)
	}
	checker.Close()
	checker.Close()
	if r, e := checker.Check(g, g); e == nil || r.Equivalent {
		t.Fatalf("closed: %+v %v", r, e)
	}
	if _, e = NewAcecEquivalenceChecker(AcecEquivalenceOptions{MaxTerms: -1}); !errors.Is(e, ErrDimension) {
		t.Fatalf("negative: %v", e)
	}
}

func TestAcecArithmeticBoundedUnknownAndDimensions(t *testing.T) {
	a, b := acecMakeXOR(t, false, false), acecMakeXOR(t, false, true)
	r, e := CheckAcecArithmeticEquivalent(a, b, AcecEquivalenceOptions{MaxInputBits: 1})
	if e != nil {
		t.Fatal(e)
	}
	if r.Proven || r.Equivalent || r.Witness != nil {
		t.Fatalf("expected honest unknown: %+v", r)
	}
	c := NewGiaManager()
	x, _ := c.AppendCI()
	c.AppendCO(x)
	if _, e = CheckAcecArithmeticEquivalent(a, c, AcecEquivalenceOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatalf("dimension: %v", e)
	}
}

func TestAcecArithmeticCIPositionSignedWidthAndOwnership(t *testing.T) {
	left := NewGiaManager()
	leftInputs := acecAppendInputs(t, left, 2)
	if _, err := left.AppendCO(leftInputs[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := left.AppendCO(leftInputs[1]); err != nil {
		t.Fatal(err)
	}

	// The unused node shifts the second CI's object ID.  Canonical signatures
	// must nevertheless name CIs by their positions at the API boundary.
	right := NewGiaManager()
	a, err := right.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = right.AppendBuffer(a); err != nil {
		t.Fatal(err)
	}
	b, err := right.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = right.AppendCO(a); err != nil {
		t.Fatal(err)
	}
	if _, err = right.AppendCO(b); err != nil {
		t.Fatal(err)
	}

	unsigned, err := CheckAcecArithmeticEquivalent(left, right, AcecEquivalenceOptions{})
	if err != nil || !unsigned.Equivalent || !unsigned.Proven {
		t.Fatalf("unsigned: %+v %v", unsigned, err)
	}
	if len(unsigned.Left.Terms) != 2 || unsigned.Left.Terms[0].Coefficient.Cmp(big.NewInt(1)) != 0 || unsigned.Left.Terms[1].Coefficient.Cmp(big.NewInt(2)) != 0 || !slices.Equal(unsigned.Left.Terms[0].Inputs, []int{0}) || !slices.Equal(unsigned.Left.Terms[1].Inputs, []int{1}) {
		t.Fatalf("unsigned fixed-width signature: %+v", unsigned.Left.Terms)
	}
	signed, err := NormalizeAcecArithmetic(left, AcecEquivalenceOptions{Signed: true})
	if err != nil || len(signed.Terms) != 2 || signed.Terms[0].Coefficient.Cmp(big.NewInt(1)) != 0 || signed.Terms[1].Coefficient.Cmp(big.NewInt(-2)) != 0 {
		t.Fatalf("two's-complement signature: %+v %v", signed.Terms, err)
	}
	unsigned.Left.Terms[0].Coefficient.SetInt64(77)
	unsigned.Left.Terms[0].Inputs[0] = 77
	again, err := NormalizeAcecArithmetic(left, AcecEquivalenceOptions{})
	if err != nil || again.Terms[0].Coefficient.Cmp(big.NewInt(1)) != 0 || again.Terms[0].Inputs[0] != 0 {
		t.Fatalf("caller mutation escaped ownership: %+v %v", again.Terms, err)
	}
}

func TestAcecArithmeticCumulativeLimitsAndNoMutation(t *testing.T) {
	g := acecMakeXOR(t, false, false)
	before := g.Clone()
	combined := 2 * g.ObjectCount()
	if r, err := CheckAcecArithmeticEquivalent(g, g, AcecEquivalenceOptions{MaxObjects: combined - 1}); !errors.Is(err, ErrLimit) || r.Proven {
		t.Fatalf("cumulative objects: %+v %v", r, err)
	}
	if !reflect.DeepEqual(g, before) {
		t.Fatal("limit check mutated graph")
	}
	if r, err := CheckAcecArithmeticEquivalent(g, g, AcecEquivalenceOptions{MaxObjects: combined}); err != nil || !r.Equivalent {
		t.Fatalf("exact object budget: %+v %v", r, err)
	}
	if !reflect.DeepEqual(g, before) {
		t.Fatal("equivalence check mutated graph")
	}
	and := NewGiaManager()
	inputs := acecAppendInputs(t, and, 2)
	root, err := and.AppendAnd(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = and.AppendCO(root); err != nil {
		t.Fatal(err)
	}
	if _, err = NormalizeAcecArithmetic(and, AcecEquivalenceOptions{MaxDegree: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("degree limit: %v", err)
	}
	fourBits := NewGiaManager()
	x, err := fourBits.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		output := GiaConst0
		if i == 3 {
			output = x
		}
		if _, err = fourBits.AppendCO(output); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = NormalizeAcecArithmetic(fourBits, AcecEquivalenceOptions{Signed: true, MaxCoefficientBits: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("coefficient-bit limit: %v", err)
	}
}

func TestAcecCheckerConcurrentLifecycleAndReentrantStop(t *testing.T) {
	g := acecMakeXOR(t, false, false)
	var checker *AcecEquivalenceChecker
	checker, err := NewAcecEquivalenceChecker(AcecEquivalenceOptions{Stop: func() bool {
		if checker == nil {
			return false
		}
		checker.Close()
		return true
	}})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, checkErr := checker.Check(g, g)
		done <- checkErr
	}()
	select {
	case err = <-done:
		if !errors.Is(err, ErrCancelled) {
			t.Fatalf("reentrant stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop callback deadlocked with Close")
	}
	if _, err = checker.Check(g, g); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed checker: %v", err)
	}

	concurrent, err := NewAcecEquivalenceChecker(AcecEquivalenceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	errorsSeen := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, checkErr := concurrent.Check(g, g)
			if checkErr != nil || !result.Equivalent || !result.Proven {
				errorsSeen <- checkErr
			}
		}()
	}
	wait.Wait()
	close(errorsSeen)
	for checkErr := range errorsSeen {
		t.Fatalf("concurrent check: %v", checkErr)
	}
	concurrent.Close()
}

func TestAcecArithmeticDeterministic5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0xacec17))
	for trial := 0; trial < 5000; trial++ {
		g := NewGiaManager()
		inputs := acecAppendInputs(t, g, 3)
		nodes := slices.Clone(inputs)
		for step := 0; step < 4; step++ {
			a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			var lit GiaLit
			var err error
			if rng.Intn(2) == 0 {
				lit, err = g.AppendAndSimplified(a, b)
			} else {
				lit, err = g.AppendXorRealSimplified(a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			nodes = append(nodes, lit)
		}
		root := nodes[rng.Intn(len(nodes))]
		if rng.Intn(2) == 0 {
			root = root.Not()
		}
		if _, e := g.AppendCO(root); e != nil {
			t.Fatal(e)
		}
		s1, e := NormalizeAcecArithmetic(g, AcecEquivalenceOptions{MaxTerms: 256, MaxWork: 4096})
		if e != nil {
			t.Fatalf("trial %d: %v", trial, e)
		}
		s2, e := NormalizeAcecArithmetic(g, AcecEquivalenceOptions{MaxTerms: 256, MaxWork: 4096})
		if e != nil || s1.Digest != s2.Digest {
			t.Fatalf("trial %d nondeterministic: %v", trial, e)
		}
		bits := []bool{rng.Intn(2) != 0, rng.Intn(2) != 0, rng.Intn(2) != 0}
		out, e := g.Evaluate(bits)
		if e != nil {
			t.Fatal(e)
		}
		value := new(big.Int)
		for _, term := range s1.Terms {
			present := true
			for _, i := range term.Inputs {
				present = present && bits[i]
			}
			if present {
				value.Add(value, term.Coefficient)
			}
		}
		if (value.Sign() != 0) != out[0] {
			t.Fatalf("trial %d bits %v polynomial %v output %v terms %v", trial, bits, value, out, strings.TrimSpace(value.String()))
		}
	}
}

func FuzzAcecArithmeticEquivalence(f *testing.F) {
	for i := 0; i < 64; i++ {
		f.Add(uint64(i) * 0x9e3779b97f4a7c15)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		rng := rand.New(rand.NewSource(int64(seed)))
		g := NewGiaManager()
		inputs := acecAppendInputs(t, g, 3)
		nodes := slices.Clone(inputs)
		for i := 0; i < 5; i++ {
			a, b := nodes[rng.Intn(len(nodes))], nodes[rng.Intn(len(nodes))]
			var n GiaLit
			var e error
			if rng.Intn(2) == 0 {
				n, e = g.AppendAndSimplified(a, b)
			} else {
				n, e = g.AppendXorRealSimplified(a, b)
			}
			if e != nil {
				t.Fatal(e)
			}
			nodes = append(nodes, n)
		}
		if _, e := g.AppendCO(nodes[len(nodes)-1]); e != nil {
			t.Fatal(e)
		}
		r, e := CheckAcecArithmeticEquivalent(g, g, AcecEquivalenceOptions{MaxTerms: 1024, MaxWork: 1 << 16})
		if e != nil || !r.Proven || !r.Equivalent {
			t.Fatalf("seed %d: %+v %v", seed, r, e)
		}
	})
}
