// Package mvc implements ABC's multi-valued cube and cover operations using
// Go-owned word slices and linked cube lists.
package mvc

const BITS_PER_WORD = 32
const BITS_PER_WORD_MINUS = 31
const BITS_PER_WORD_LOG = 5
const BITS_DISJOINT uint32 = 0x55555555
const BITS_FULL uint32 = 0xffffffff

type Mvc_CubeWord_t = uint32
type Mvc_Cube_t struct {
	Next         *Mvc_Cube_t
	Unused, Ones int
	Prime, Essen bool
	Data         []uint32
}
type Mvc_List_t struct {
	Head, Tail *Mvc_Cube_t
	Items      int
}
type Mvc_Cover_t struct {
	NBits int
	Cubes Mvc_List_t
	Array []*Mvc_Cube_t
	Lits  []int
	Mask  *Mvc_Cube_t
	Mem   *Mvc_Manager_t
}
type Mvc_Manager_t struct{}
type Mvc_Data_t struct {
	Man     *Mvc_Manager_t
	BinVars int
	MaskBin *Mvc_Cube_t
	Masks   []*Mvc_Cube_t
	Temp    [3]*Mvc_Cube_t
}
