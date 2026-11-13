package rwr_test

import (
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/opt/rwr"
)

type apiNetwork struct{}

func (apiNetwork) Lookup(id int32) (any, bool)                          { return id, true }
func (apiNetwork) Complement(v any) any                                 { return ^v.(int32) }
func (apiNetwork) FanoutCount(any) int                                  { return 2 }
func (apiNetwork) MarkBoundary([]any, int)                              {}
func (apiNetwork) LabelMFFC(any) int32                                  { return 3 }
func (apiNetwork) CountGraph(any, *dec.Dec_Graph_t, int32, int32) int32 { return 1 }

// This external-package test verifies the rewrite entry point and adapter
// contract remain callable by clients that cannot access opt/rwr internals.
func TestExportedRewriteAPI(t *testing.T) {
	manager := rwr.Rwr_ManStart(0)
	gain := rwr.Rwr_NodeRewrite(manager, apiNetwork{}, nil,
		[]rwr.Rwr_Cut{{Truth: 0x6996, Leaves: []int32{0, 1, 2, 3}}}, 12, 0, 0)
	if gain != 2 || rwr.Rwr_ManReadDecs(manager) == nil || len(rwr.Rwr_ManReadLeaves(manager)) != 4 {
		t.Fatalf("gain=%d graph=%p leaves=%v", gain, rwr.Rwr_ManReadDecs(manager), rwr.Rwr_ManReadLeaves(manager))
	}
}
