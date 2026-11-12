package pla

/**CFile****************************************************************

  FileName    [plaWrite.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaWrite.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func (p *Pla_Man_t) Write(w io.Writer) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# SOP %q written via PLA package in ABC\n\n", p.Name)
	types := []string{"fd", "f", "fr", "fdr", "???"}
	if p.Type >= PLA_FILE_NONE {
		return fmt.Errorf("invalid PLA type")
	}
	if p.Type != PLA_FILE_FD {
		fmt.Fprintf(&b, ".type %s\n", types[p.Type])
	}
	fmt.Fprintf(&b, ".i %d\n.o %d\n.p %d\n", p.Ins, p.Outs, len(p.Cubes))
	for _, d := range p.Directives {
		b.WriteString(d)
		b.WriteByte('\n')
	}
	const literals = "-01?"
	for _, cube := range p.Cubes {
		if len(cube.In) != words(p.Ins) || len(cube.Out) != words(p.Outs) {
			return fmt.Errorf("cube storage width mismatch")
		}
		for i := int32(0); i < p.Ins; i++ {
			b.WriteByte(literals[Pla_CubeGetLit(cube.In, i)])
		}
		b.WriteByte(' ')
		for i := int32(0); i < p.Outs; i++ {
			b.WriteByte(literals[Pla_CubeGetLit(cube.Out, i)])
		}
		b.WriteByte('\n')
	}
	b.WriteString(".e\n\n")
	_, err := io.WriteString(w, b.String())
	return err
}
func Pla_WritePlaInt(p *Pla_Man_t) string {
	var b strings.Builder
	if err := p.Write(&b); err != nil {
		panic(err)
	}
	return b.String()
}
func Pla_WritePla(p *Pla_Man_t, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	err = p.Write(f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
