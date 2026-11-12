package bsat

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [satTrace.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Records the trace of SAT solving in the CNF form.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satTrace.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

// Trace APIs retain the source's explicit-write contract. They report I/O errors
// rather than silently producing a truncated proof trace.
func Sat_SolverTraceStart(s *sat_solver, name string) error {
	if s.pFile != nil {
		return fmt.Errorf("solver trace already open")
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(file, "%128s\n", ""); err != nil {
		file.Close()
		return err
	}
	s.pFile, s.nClauses, s.nRoots = file, 0, 0
	return nil
}
func Sat_SolverTraceStop(s *sat_solver) error {
	if s.pFile == nil {
		return nil
	}
	file := s.pFile
	s.pFile = nil
	_, err := file.Seek(0, 0)
	if err == nil {
		_, err = fmt.Fprintf(file, "p %d %d %d", s.size, s.nClauses, s.nRoots)
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
func Sat_SolverTraceWrite(s *sat_solver, literals []lit, root int32) error {
	if s.pFile == nil {
		return nil
	}
	for _, l := range literals {
		if _, err := fmt.Fprintf(s.pFile, " %d", lit_print(l)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(s.pFile, " 0"); err != nil {
		return err
	}
	s.nClauses++
	if root != 0 {
		s.nRoots++
	}
	return nil
}
