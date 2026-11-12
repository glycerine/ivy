package amap

import (
	"fmt"
	"io"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/map/amap/formula"
)

// Amap_ParseFormulaToHop is the native Go entry point for parsing an Amap
// gate equation and constructing its Hop graph. Pin names are supplied in the
// order used by the library cell; the returned Hop variables use that order.
func Amap_ParseFormulaToHop(output io.Writer, text string, pinNames []string, manager *hop.Hop_Man_t, gateName string) (*hop.Hop_Obj_t, error) {
	if manager == nil {
		return nil, fmt.Errorf("Amap_ParseFormula(): gate %q: nil Hop manager", gateName)
	}
	expr, err := formula.Parse(text)
	if err != nil {
		wrapped := fmt.Errorf("Amap_ParseFormula(): gate %q: %w", gateName, err)
		if output != nil {
			_, _ = fmt.Fprintln(output, wrapped)
		}
		return nil, wrapped
	}
	vars := make(map[string]*hop.Hop_Obj_t, len(pinNames))
	for i, name := range pinNames {
		if _, exists := vars[name]; exists {
			return nil, fmt.Errorf("Amap_ParseFormula(): gate %q: duplicate pin %q", gateName, name)
		}
		vars[name] = hop.Hop_IthVar(manager, int32(i))
	}
	root, err := expr.ToHop(manager, vars)
	if err != nil {
		wrapped := fmt.Errorf("Amap_ParseFormula(): gate %q: %w", gateName, err)
		if output != nil {
			_, _ = fmt.Fprintln(output, wrapped)
		}
		return nil, wrapped
	}
	return root, nil
}
