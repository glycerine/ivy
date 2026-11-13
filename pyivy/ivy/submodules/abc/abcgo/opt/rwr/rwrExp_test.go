package rwr

import (
	"github.com/glycerine/abcgo/misc/extra/npn"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceExploreCounts(t *testing.T) {
	Rwt_Man4ExploreStart()
	Rwt_Man4ExploreCount(0xaaaa)
	Rwt_Man4ExploreCount(0x5555)
	canonical, _, _, _ := npn.Extra_Truth4VarNPN()
	if got := s_pManRwrExp4.pnCounts[canonical[0xaaaa]]; got != 2 {
		t.Fatalf("canonical four-var count %d", got)
	}
	Rwt_Man5ExploreStart()
	Rwt_Man5ExploreCount(0x12345678)
	Rwt_Man5ExploreCount(0x12345678)
	Rwt_Man5ExploreCount(0x87654321)
	if len(s_pManRwrExp5.tTableNN) != 2 || s_pManRwrExp5.tTableNN[0x12345678] != 2 {
		t.Fatalf("five-var counts %v", s_pManRwrExp5.tTableNN)
	}
}

func TestSourceTruthCountFileFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "counts")
	rwrWriteTruthCounts(path, []uint32{0x1234abcd}, map[uint32]int{0x1234abcd: 7})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "0x1234ABCD          7\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
