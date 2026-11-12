// Bookshelf database writers from place_io.c.
package place

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func bookshelfCellName(c *ConcreteCell) string {
	if c.m_label != "" && len(strings.Fields(c.m_label)) == 1 && !strings.ContainsAny(c.m_label, "#\r\n") {
		return c.m_label
	}
	return fmt.Sprintf("CELL%d", c.m_id)
}
func writePlacementFile(filename string, write func(io.Writer) error) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	b := bufio.NewWriter(f)
	err = write(b)
	if flushErr := b.Flush(); err == nil {
		err = flushErr
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}
func writeBookshelfNodes(filename string) error {
	return writePlacementFile(filename, func(w io.Writer) error {
		nodes, terms := 0, 0
		for _, c := range g_place_concreteCells {
			if c != nil {
				nodes++
				if c.m_parent.m_pad {
					terms++
				}
			}
		}
		if _, err := fmt.Fprintf(w, "UCLA nodes 1.0\nNumNodes : %d\nNumTerminals : %d\n", nodes, terms); err != nil {
			return err
		}
		for _, c := range g_place_concreteCells {
			if c == nil {
				continue
			}
			terminal := ""
			if c.m_parent.m_pad {
				terminal = " terminal"
			}
			if _, err := fmt.Fprintf(w, "%s %g %g%s\n", bookshelfCellName(c), c.m_parent.m_width, c.m_parent.m_height, terminal); err != nil {
				return err
			}
		}
		return nil
	})
}
func writeBookshelfPl(filename string) error {
	return writePlacementFile(filename, func(w io.Writer) error {
		if _, err := fmt.Fprintln(w, "UCLA pl 1.0"); err != nil {
			return err
		}
		for _, c := range g_place_concreteCells {
			if c == nil {
				continue
			}
			fixed := ""
			if c.m_fixed {
				fixed = " /FIXED"
			}
			if _, err := fmt.Fprintf(w, "%s %g %g : N%s\n", bookshelfCellName(c), c.m_x, c.m_y, fixed); err != nil {
				return err
			}
		}
		return nil
	})
}
func writeBookshelfNets(filename string) error {
	return writePlacementFile(filename, func(w io.Writer) error {
		nets, pins := 0, 0
		for _, n := range g_place_concreteNets {
			if n != nil {
				nets++
				pins += len(n.m_terms)
			}
		}
		if _, err := fmt.Fprintf(w, "UCLA nets 1.0\nNumNets : %d\nNumPins : %d\n", nets, pins); err != nil {
			return err
		}
		for _, n := range g_place_concreteNets {
			if n == nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "NetDegree : %d NET%d\n", len(n.m_terms), n.m_id); err != nil {
				return err
			}
			for _, c := range n.m_terms {
				if _, err := fmt.Fprintf(w, "%s B\n", bookshelfCellName(c)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func writeBookshelf(prefix string) error {
	if _, err := bookshelfNames(g_place_concreteCells); err != nil {
		return err
	}
	if err := validatePlacementGeometry(); err != nil {
		return err
	}
	if err := writeBookshelfNodes(prefix + ".nodes"); err != nil {
		return err
	}
	if err := writeBookshelfNets(prefix + ".nets"); err != nil {
		return err
	}
	return writeBookshelfPl(prefix + ".pl")
}
func WriteBookshelf(prefix string) error            { return writeBookshelf(prefix) }
func WriteBookshelfPlacement(filename string) error { return writeBookshelfPl(filename) }
