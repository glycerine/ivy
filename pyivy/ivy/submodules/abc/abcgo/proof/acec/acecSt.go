package acec

// Three-variable NPN classes from acecSt.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
)

var Npn3Table = [256][2]int32{
	{0x00, 0},
	{0x01, 1},
	{0x01, 1},
	{0x03, 2},
	{0x01, 1},
	{0x03, 2},
	{0x06, 3},
	{0x07, 4},
	{0x01, 1},
	{0x06, 3},
	{0x03, 2},
	{0x07, 4},
	{0x03, 2},
	{0x07, 4},
	{0x07, 4},
	{0x0F, 5},
	{0x01, 1},
	{0x03, 2},
	{0x06, 3},
	{0x07, 4},
	{0x06, 3},
	{0x07, 4},
	{0x16, 6},
	{0x17, 7},
	{0x18, 8},
	{0x19, 9},
	{0x19, 9},
	{0x1B, 10},
	{0x19, 9},
	{0x1B, 10},
	{0x1E, 11},
	{0x07, 4},
	{0x01, 1},
	{0x06, 3},
	{0x03, 2},
	{0x07, 4},
	{0x18, 8},
	{0x19, 9},
	{0x19, 9},
	{0x1B, 10},
	{0x06, 3},
	{0x16, 6},
	{0x07, 4},
	{0x17, 7},
	{0x19, 9},
	{0x1E, 11},
	{0x1B, 10},
	{0x07, 4},
	{0x03, 2},
	{0x07, 4},
	{0x07, 4},
	{0x0F, 5},
	{0x19, 9},
	{0x1B, 10},
	{0x1E, 11},
	{0x07, 4},
	{0x19, 9},
	{0x1E, 11},
	{0x1B, 10},
	{0x07, 4},
	{0x3C, 12},
	{0x19, 9},
	{0x19, 9},
	{0x03, 2},
	{0x01, 1},
	{0x06, 3},
	{0x18, 8},
	{0x19, 9},
	{0x03, 2},
	{0x07, 4},
	{0x19, 9},
	{0x1B, 10},
	{0x06, 3},
	{0x16, 6},
	{0x19, 9},
	{0x1E, 11},
	{0x07, 4},
	{0x17, 7},
	{0x1B, 10},
	{0x07, 4},
	{0x03, 2},
	{0x07, 4},
	{0x19, 9},
	{0x1B, 10},
	{0x07, 4},
	{0x0F, 5},
	{0x1E, 11},
	{0x07, 4},
	{0x19, 9},
	{0x1E, 11},
	{0x3C, 12},
	{0x19, 9},
	{0x1B, 10},
	{0x07, 4},
	{0x19, 9},
	{0x03, 2},
	{0x06, 3},
	{0x16, 6},
	{0x19, 9},
	{0x1E, 11},
	{0x19, 9},
	{0x1E, 11},
	{0x3C, 12},
	{0x19, 9},
	{0x16, 6},
	{0x69, 13},
	{0x1E, 11},
	{0x16, 6},
	{0x1E, 11},
	{0x16, 6},
	{0x19, 9},
	{0x06, 3},
	{0x07, 4},
	{0x17, 7},
	{0x1B, 10},
	{0x07, 4},
	{0x1B, 10},
	{0x07, 4},
	{0x19, 9},
	{0x03, 2},
	{0x1E, 11},
	{0x16, 6},
	{0x19, 9},
	{0x06, 3},
	{0x19, 9},
	{0x06, 3},
	{0x18, 8},
	{0x01, 1},
	{0x01, 1},
	{0x18, 8},
	{0x06, 3},
	{0x19, 9},
	{0x06, 3},
	{0x19, 9},
	{0x16, 6},
	{0x1E, 11},
	{0x03, 2},
	{0x19, 9},
	{0x07, 4},
	{0x1B, 10},
	{0x07, 4},
	{0x1B, 10},
	{0x17, 7},
	{0x07, 4},
	{0x06, 3},
	{0x19, 9},
	{0x16, 6},
	{0x1E, 11},
	{0x16, 6},
	{0x1E, 11},
	{0x69, 13},
	{0x16, 6},
	{0x19, 9},
	{0x3C, 12},
	{0x1E, 11},
	{0x19, 9},
	{0x1E, 11},
	{0x19, 9},
	{0x16, 6},
	{0x06, 3},
	{0x03, 2},
	{0x19, 9},
	{0x07, 4},
	{0x1B, 10},
	{0x19, 9},
	{0x3C, 12},
	{0x1E, 11},
	{0x19, 9},
	{0x07, 4},
	{0x1E, 11},
	{0x0F, 5},
	{0x07, 4},
	{0x1B, 10},
	{0x19, 9},
	{0x07, 4},
	{0x03, 2},
	{0x07, 4},
	{0x1B, 10},
	{0x17, 7},
	{0x07, 4},
	{0x1E, 11},
	{0x19, 9},
	{0x16, 6},
	{0x06, 3},
	{0x1B, 10},
	{0x19, 9},
	{0x07, 4},
	{0x03, 2},
	{0x19, 9},
	{0x18, 8},
	{0x06, 3},
	{0x01, 1},
	{0x03, 2},
	{0x19, 9},
	{0x19, 9},
	{0x3C, 12},
	{0x07, 4},
	{0x1B, 10},
	{0x1E, 11},
	{0x19, 9},
	{0x07, 4},
	{0x1E, 11},
	{0x1B, 10},
	{0x19, 9},
	{0x0F, 5},
	{0x07, 4},
	{0x07, 4},
	{0x03, 2},
	{0x07, 4},
	{0x1B, 10},
	{0x1E, 11},
	{0x19, 9},
	{0x17, 7},
	{0x07, 4},
	{0x16, 6},
	{0x06, 3},
	{0x1B, 10},
	{0x19, 9},
	{0x19, 9},
	{0x18, 8},
	{0x07, 4},
	{0x03, 2},
	{0x06, 3},
	{0x01, 1},
	{0x07, 4},
	{0x1E, 11},
	{0x1B, 10},
	{0x19, 9},
	{0x1B, 10},
	{0x19, 9},
	{0x19, 9},
	{0x18, 8},
	{0x17, 7},
	{0x16, 6},
	{0x07, 4},
	{0x06, 3},
	{0x07, 4},
	{0x06, 3},
	{0x03, 2},
	{0x01, 1},
	{0x0F, 5},
	{0x07, 4},
	{0x07, 4},
	{0x03, 2},
	{0x07, 4},
	{0x03, 2},
	{0x06, 3},
	{0x01, 1},
	{0x07, 4},
	{0x06, 3},
	{0x03, 2},
	{0x01, 1},
	{0x03, 2},
	{0x01, 1},
	{0x01, 1},
	{0x00, 0},
}

// Acec_GenerateNpnTable prints the canonical three-input truth classes.
func Acec_GenerateNpnTable() {
	classes := make(map[uint32]int)
	for truth := uint32(0); truth < 256; truth++ {
		canonical := extra.Extra_TruthCanonNPN(truth, 3)
		class, ok := classes[canonical]
		if !ok {
			class = len(classes)
			classes[canonical] = class
		}
		fmt.Printf("{0x%02X, %2d}, // 0x%02X = %3d\n", canonical, class, truth, truth)
	}
}

// acecStatsMatrix counts mapped LUT fanin/fanout class pairs and output
// drivers. Unmapped inputs intentionally use class zero as in the C source.
func acecStatsMatrix(graph *gia.Gia_Man_t) [256]int32 {
	if graph == nil || gia.Gia_ManHasMapping(graph) == 0 || gia.Gia_ManLutSizeMax(graph) >= 4 {
		panic("ACEC statistics require a mapped 3-LUT network")
	}
	var matrix [256]int32
	count := int(gia.Gia_ManObjNum(graph))
	truths := vec.Vec_WrdStart(count)
	temp := vec.Vec_WrdStart(count)
	for id := int32(1); id < int32(count); id++ {
		if gia.Gia_ObjIsLut(graph, id) != 0 {
			vec.Vec_WrdWriteEntry(truths, int(id), gia.Gia_ObjComputeTruthTable6Lut(graph, id, temp))
		}
	}
	for id := int32(1); id < int32(count); id++ {
		if gia.Gia_ObjIsLut(graph, id) == 0 {
			continue
		}
		out := Npn3Table[vec.Vec_WrdEntry(truths, int(id))&0xff][1]
		for _, fanin := range gia.Gia_ObjLutFanins(graph, id) {
			in := Npn3Table[vec.Vec_WrdEntry(truths, int(fanin))&0xff][1]
			matrix[(in<<4)+out]++
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(graph); i++ {
		co := gia.Gia_ManCo(graph, i)
		id := gia.Gia_ObjId(graph, co)
		fanin := gia.Gia_ObjFaninId0(co, id)
		in := Npn3Table[vec.Vec_WrdEntry(truths, int(fanin))&0xff][1]
		matrix[in<<4]++
	}
	return matrix
}

func Acec_StatsCollect(graph *gia.Gia_Man_t, verbose int32) {
	matrix := acecStatsMatrix(graph)
	fmt.Print("fi / fo")
	for i := 0; i < 14; i++ {
		fmt.Printf("%6d ", i)
	}
	fmt.Println()
	for i := 0; i < 14; i++ {
		fmt.Printf("%6d ", i)
		for j := 0; j < 14; j++ {
			value := matrix[i*16+j]
			if value != 0 {
				fmt.Printf("%6d ", value)
			} else {
				fmt.Printf("%6s ", ".")
			}
		}
		fmt.Println()
	}
	fmt.Println("\nClasses:")
	names := kit.Kit_DsdNpn4ClassNames()
	for class := int32(0); class < 14; class++ {
		for _, entry := range Npn3Table {
			if entry[1] == class {
				index := dar.Dar_LibReturnClass(uint32(entry[0]<<8 | entry[0]))
				fmt.Printf("%2d : %s\n", class, names[index])
				break
			}
		}
	}
}
