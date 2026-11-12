package fxu

import (
	"fmt"
	"io"
)

func matrixPrint(w io.Writer, p *Fxu_Matrix) error {
	if _, err := fmt.Fprintf(w, "FXU matrix: %d columns, %d cubes, %d literals\n", len(p.Vars), len(p.Cubes), p.EntryCount()); err != nil {
		return err
	}
	for _, c := range p.Cubes {
		if c.deleted {
			continue
		}
		if _, err := fmt.Fprintf(w, "node %d cube %d: %v\n", c.Owner, c.Index, c.columns()); err != nil {
			return err
		}
	}
	return nil
}
func Fxu_MatrixPrint(w io.Writer, p *Fxu_Matrix) error { return matrixPrint(w, p) }
func Fxu_MatrixPrintDivisor(w io.Writer, d *Fxu_Double) error {
	_, err := fmt.Fprintf(w, "D%d weight=%d %v + %v pairs=%d\n", d.Num, d.Weight, d.First, d.Second, len(d.Pairs))
	return err
}
func Fxu_MatrixPrintSingle(w io.Writer, s *Fxu_Single) error {
	_, err := fmt.Fprintf(w, "S%d weight=%d %d %d\n", s.Num, s.Weight, s.Var1.Index, s.Var2.Index)
	return err
}
func Fxu_MatrixPrintCube(w io.Writer, c *Fxu_Cube) error {
	_, err := fmt.Fprintf(w, "node %d cube %d %v\n", c.Owner, c.Index, c.columns())
	return err
}
