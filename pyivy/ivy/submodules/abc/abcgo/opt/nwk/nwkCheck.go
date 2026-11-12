package nwk

import "fmt"

func (p *Nwk_Man_t) Check() error {
	for i, o := range p.Objects {
		if o == nil {
			continue
		}
		if o.Manager != p || o.ID != int32(i) {
			return fmt.Errorf("nwk: object %d has invalid identity", i)
		}
		if o.Type == NWK_OBJ_CI && len(o.Fanins) != 0 {
			return fmt.Errorf("nwk: input %d has fanins", i)
		}
		if o.Type == NWK_OBJ_CO && len(o.Fanins) != 1 {
			return fmt.Errorf("nwk: output %d needs one fanin", i)
		}
		for _, n := range o.Fanins {
			if n == nil || n.Manager != p || findNode(n.Fanouts, o) < 0 {
				return fmt.Errorf("nwk: asymmetric fanin at %d", i)
			}
		}
		for _, n := range o.Fanouts {
			if n == nil || n.Manager != p || findNode(n.Fanins, o) < 0 {
				return fmt.Errorf("nwk: asymmetric fanout at %d", i)
			}
		}
	}
	for i, o := range p.Inputs {
		if o == nil || o.Type != NWK_OBJ_CI || o.PioID != int32(i) {
			return fmt.Errorf("nwk: invalid input interface")
		}
	}
	for i, o := range p.Outputs {
		if o == nil || o.Type != NWK_OBJ_CO || o.PioID != int32(i) {
			return fmt.Errorf("nwk: invalid output interface")
		}
	}
	return nil
}
func Nwk_ManCheck(p *Nwk_Man_t) int32 {
	if e := p.Check(); e != nil {
		fmt.Fprintln(writer(p), e)
		return 0
	}
	return 1
}
