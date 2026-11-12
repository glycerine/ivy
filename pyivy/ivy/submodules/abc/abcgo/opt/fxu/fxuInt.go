package fxu

// Literal columns retain FXU's co-singleton encoding: old nodes use even=true,
// odd=false, whereas extracted nodes use odd=true, even=false.
type Fxu_Var struct {
	Index    int32
	Literals []*Fxu_Lit
}
type Fxu_Cube struct {
	Index, Owner int32
	Literals     []*Fxu_Lit
	deleted      bool
}
type Fxu_Lit struct {
	Cube *Fxu_Cube
	Var  *Fxu_Var
}
type Fxu_Pair struct {
	Cube1, Cube2        *Fxu_Cube
	Base, First, Second []int32
	Divisor             *Fxu_Double
}
type Fxu_Single struct {
	Num, Weight int32
	Var1, Var2  *Fxu_Var
	heapIndex   int
}
type Fxu_Double struct {
	Num, Weight   int32
	First, Second []int32
	Pairs         []*Fxu_Pair
	heapIndex     int
}
type Fxu_Matrix struct {
	Vars                                         []*Fxu_Var
	Cubes                                        []*Fxu_Cube
	Singles                                      []*Fxu_Single
	Doubles                                      []*Fxu_Double
	OldNodes                                     int32
	covers                                       map[int32][]*Fxu_Cube
	phases                                       map[int32]bool
	changed                                      map[int32]bool
	singleCatalog                                map[[2]int32]*Fxu_Single
	doubleCatalog                                map[string]*Fxu_Double
	singleHeap                                   *Fxu_HeapSingle
	doubleHeap                                   *Fxu_HeapDouble
	params                                       *Fxu_Data_t
	weightLimit                                  int32
	singleSerial, doubleSerial                   int32
	SinglesTaken, DoublesTaken, ComplementsTaken int32
	cubeRing                                     []*Fxu_Cube
	varRing                                      []*Fxu_Var
}

func (p *Fxu_Matrix) EntryCount() int32 {
	n := int32(0)
	for _, c := range p.Cubes {
		if !c.deleted {
			n += int32(len(c.Literals))
		}
	}
	return n
}
func columnPositive(p *Fxu_Matrix, column int32) bool {
	return (column&1 == 0) != (column/2 >= p.OldNodes)
}
func (c *Fxu_Cube) columns() []int32 {
	result := make([]int32, len(c.Literals))
	for i, l := range c.Literals {
		result[i] = l.Var.Index
	}
	return result
}
