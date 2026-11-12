package bsat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

/**CFile****************************************************************

  FileName    [satChecker.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Resolution proof checker.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satChecker.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

// ProofSummary describes ABC's numbered binary-resolution proof stream.
type ProofSummary struct {
	Roots, Steps, Clauses int
	EmptyDerived          bool
}

// ResolveProofClauses resolves clauses written as signed DIMACS literals. Exactly
// one complementary pair is required; parent clauses must not contain duplicates.
func ResolveProofClauses(a, b []int32) ([]int32, error) {
	seen := map[int32]bool{}
	for _, clause := range [][]int32{a, b} {
		clear(seen)
		for _, l := range clause {
			if l == 0 || l == (-1<<31) {
				return nil, fmt.Errorf("invalid proof literal %d", l)
			}
			if seen[-l] {
				return nil, fmt.Errorf("tautological parent clause at literal %d", l)
			}
			if seen[l] {
				return nil, fmt.Errorf("duplicate proof literal %d", l)
			}
			seen[l] = true
		}
	}
	pivot, count := int32(0), 0
	for _, x := range a {
		if seen[-x] {
			pivot = x
			count++
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("resolution requires exactly one pivot, got %d", count)
	}
	clear(seen)
	var result []int32
	for _, clause := range [][]int32{a, b} {
		for _, l := range clause {
			if l == pivot || l == -pivot || seen[l] {
				continue
			}
			seen[l] = true
			result = append(result, l)
		}
	}
	return result, nil
}

// CheckResolutionProof accepts root records "id literals... 0 0" and derived
// records "id * parent1 parent2 0". IDs start at one and refer only backwards.
func CheckResolutionProof(input io.Reader) (ProofSummary, error) {
	reader := bufio.NewReader(input)
	clauses := [][]int32{nil}
	var summary ProofSummary
	for {
		var id int
		_, err := fmt.Fscan(reader, &id)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return summary, fmt.Errorf("read clause ID: %w", err)
		}
		if id != len(clauses) {
			return summary, fmt.Errorf("clause ID %d: expected %d", id, len(clauses))
		}
		var token string
		if _, err := fmt.Fscan(reader, &token); err != nil {
			return summary, fmt.Errorf("clause %d: %w", id, err)
		}
		var result []int32
		if token == "*" {
			var a, b, zero int
			if _, err := fmt.Fscan(reader, &a, &b, &zero); err != nil {
				return summary, fmt.Errorf("resolution %d: %w", id, err)
			}
			if a <= 0 || a >= id || b <= 0 || b >= id || zero != 0 {
				return summary, fmt.Errorf("resolution %d: invalid parents or terminator", id)
			}
			result, err = ResolveProofClauses(clauses[a], clauses[b])
			if err != nil {
				return summary, fmt.Errorf("resolution %d: %w", id, err)
			}
			summary.Steps++
		} else {
			seen := map[int32]bool{}
			for {
				n, err := strconv.ParseInt(token, 10, 32)
				if err != nil || n == (-1<<31) {
					return summary, fmt.Errorf("clause %d: invalid literal %q", id, token)
				}
				if n == 0 {
					break
				}
				if seen[-int32(n)] {
					return summary, fmt.Errorf("clause %d: tautological root", id)
				}
				if seen[int32(n)] {
					return summary, fmt.Errorf("clause %d: duplicate literal %d", id, n)
				}
				seen[int32(n)] = true
				result = append(result, int32(n))
				if _, err := fmt.Fscan(reader, &token); err != nil {
					return summary, fmt.Errorf("clause %d: missing root terminator: %w", id, err)
				}
			}
			var zero int
			if _, err := fmt.Fscan(reader, &zero); err != nil || zero != 0 {
				return summary, fmt.Errorf("clause %d: missing second zero", id)
			}
			summary.Roots++
		}
		clauses = append(clauses, result)
		summary.Clauses++
	}
	if summary.Clauses == 0 {
		return summary, errors.New("empty proof stream")
	}
	summary.EmptyDerived = len(clauses[len(clauses)-1]) == 0
	return summary, nil
}

// Compatibility entry points use Go slices in place of C Vec_Vec_t storage.
func Sat_PrintClause(clauses *[][]int32, id int32) {
	fmt.Printf("Clause %d:  {", id)
	for _, l := range (*clauses)[id] {
		fmt.Printf(" %d", l)
	}
	fmt.Println(" }")
}
func Sat_ProofResolve(clauses *[][]int32, result, a, b int32) int32 {
	if result <= 0 || int(result) >= len(*clauses) || a <= 0 || b <= 0 || a >= result || b >= result || len((*clauses)[result]) != 0 {
		return 0
	}
	resolved, err := ResolveProofClauses((*clauses)[a], (*clauses)[b])
	if err != nil {
		fmt.Printf("Error: Clause %d = Resolve(%d, %d): %v\n", result, a, b, err)
		return 0
	}
	(*clauses)[result] = resolved
	return 1
}
func Sat_ProofChecker(name string) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	summary, err := CheckResolutionProof(file)
	if err != nil {
		fmt.Printf("Error detected in the resolution proof: %v\n", err)
		return
	}
	fmt.Printf("The proof contains %d roots and %d resolution steps.\n", summary.Roots, summary.Steps)
	if summary.EmptyDerived {
		fmt.Println("The empty clause is derived.")
	} else {
		fmt.Println("The last clause is not empty.")
	}
}
