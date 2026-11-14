package abci

import (
	"testing"

	baseabc "github.com/glycerine/abcgo/base/abc"
)

func TestAbcNtkMapCopiesCiArrivalTimes(t *testing.T) {
	network := baseabc.NewNetwork("map-arrivals")
	baseabc.Abc_NtkCreatePi(network)
	baseabc.Abc_NtkCreatePi(network)
	times := []Abc_Time_t{{Rise: 1.25, Fall: 2.5}, {Rise: 4, Fall: 3}}
	got := Abc_NtkMapCopyCiArrival(network, &times[0])
	if len(got) != 2 {
		t.Fatalf("arrival count = %d, want 2", len(got))
	}
	want := []Map_Time_t{{Rise: 1.25, Fall: 2.5, Worst: 2.5}, {Rise: 4, Fall: 3, Worst: 4}}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arrival[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
