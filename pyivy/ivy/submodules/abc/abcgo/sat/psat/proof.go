package psat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

type Chain struct {
	Clauses []int
	Pivots  []int
}

// Proof IDs number roots first, then chains in order. Every chain resolves its
// first parent successively with each next parent using the corresponding pivot.
type Proof struct {
	Version   int
	BaseRoots int
	Roots     [][]int
	Chains    []Chain
	Final     int
}

func (p *Proof) Clone() *Proof {
	if p == nil {
		return nil
	}
	q := *p
	q.Roots = make([][]int, len(p.Roots))
	for i, root := range p.Roots {
		q.Roots[i] = append([]int(nil), root...)
	}
	q.Chains = make([]Chain, len(p.Chains))
	for i, chain := range p.Chains {
		q.Chains[i] = Chain{append([]int(nil), chain.Clauses...), append([]int(nil), chain.Pivots...)}
	}
	return &q
}
func normalize(lits []int) ([]int, bool) {
	out := append([]int(nil), lits...)
	sort.Ints(out)
	kept := out[:0]
	tautology := false
	for _, lit := range out {
		if len(kept) > 0 {
			last := kept[len(kept)-1]
			if lit == last {
				continue
			}
			if lit == last^1 {
				tautology = true
			}
		}
		kept = append(kept, lit)
	}
	return kept, tautology
}

// Resolve checks both opposite pivot occurrences. It never drops a variable
// absent from one parent, which would manufacture an unsound "proof".
func Resolve(left, right []int, pivot int) ([]int, error) {
	if pivot < 0 {
		return nil, fmt.Errorf("negative pivot")
	}
	a, b := -1, -1
	out := make([]int, 0, len(left)+len(right))
	for side, clause := range [][]int{left, right} {
		for _, lit := range clause {
			if lit < 0 {
				return nil, fmt.Errorf("negative literal")
			}
			if lit>>1 == pivot {
				sign := lit & 1
				if side == 0 {
					if a >= 0 && a != sign {
						return nil, fmt.Errorf("ambiguous left pivot")
					}
					a = sign
				} else {
					if b >= 0 && b != sign {
						return nil, fmt.Errorf("ambiguous right pivot")
					}
					b = sign
				}
			} else {
				out = append(out, lit)
			}
		}
	}
	if a < 0 || b < 0 || a == b {
		return nil, fmt.Errorf("parents do not contain complementary pivot %d", pivot)
	}
	out, _ = normalize(out)
	return out, nil
}
func (p *Proof) Validate() error {
	if p == nil || p.Version != 1 {
		return fmt.Errorf("unsupported proof version")
	}
	if p.BaseRoots < 0 || p.BaseRoots > len(p.Roots) {
		return fmt.Errorf("invalid base-root count")
	}
	clauses := make([][]int, 0, len(p.Roots)+len(p.Chains))
	for _, root := range p.Roots {
		for _, lit := range root {
			if lit < 0 {
				return fmt.Errorf("negative root literal")
			}
		}
		canonical, _ := normalize(root)
		clauses = append(clauses, canonical)
	}
	for i, chain := range p.Chains {
		if len(chain.Pivots) == 0 || len(chain.Clauses) != len(chain.Pivots)+1 {
			return fmt.Errorf("invalid chain %d arity", i)
		}
		for _, id := range chain.Clauses {
			if id < 0 || id >= len(clauses) {
				return fmt.Errorf("invalid or forward parent in chain %d", i)
			}
		}
		current := clauses[chain.Clauses[0]]
		for k, v := range chain.Pivots {
			var err error
			current, err = Resolve(current, clauses[chain.Clauses[k+1]], v)
			if err != nil {
				return fmt.Errorf("chain %d: %w", i, err)
			}
		}
		clauses = append(clauses, current)
	}
	if p.Final < 0 || p.Final >= len(clauses) || len(clauses[p.Final]) != 0 {
		return fmt.Errorf("proof does not end at an empty clause")
	}
	return nil
}
func (p *Proof) Write(w io.Writer) error {
	if err := p.Validate(); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}
func ReadProof(r io.Reader) (*Proof, error) {
	var p Proof
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	var trailing json.RawMessage
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, fmt.Errorf("trailing proof data: %w", err)
		}
		return nil, fmt.Errorf("unexpected data after proof")
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// ValidateAgainst anchors every root to the actual CNF and assumptions, rather
// than merely checking a valid resolution derivation for arbitrary input roots.
func (p *Proof) ValidateAgainst(clauses [][]int, assumptions []int) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.BaseRoots != len(clauses) || len(p.Roots) != len(clauses)+len(assumptions) {
		return fmt.Errorf("proof roots do not match input counts")
	}
	for i := range p.Roots {
		expected := []int(nil)
		if i < len(clauses) {
			expected = clauses[i]
		} else {
			expected = []int{assumptions[i-len(clauses)]}
		}
		for _, lit := range expected {
			if lit < 0 {
				return fmt.Errorf("negative input literal")
			}
		}
		want, _ := normalize(expected)
		got, _ := normalize(p.Roots[i])
		if len(want) != len(got) {
			return fmt.Errorf("proof root %d differs from input", i)
		}
		for j := range want {
			if want[j] != got[j] {
				return fmt.Errorf("proof root %d differs from input", i)
			}
		}
	}
	return nil
}
