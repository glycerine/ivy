package aig

// FanoutIndex preserves ABC's circular per-pin fanout lists. Each node owns
// five slots: first consumer, then previous/next links for each consumer pin.
// Handles encode consumer nodeID*2+pin, independently of edge complementation.
type FanoutIndex struct{ data []int32 }

func Aig_FanoutCreate(id, pin int32) int32 {
	if id <= 0 || pin < 0 || pin > 1 {
		panic("aig: invalid fanout handle")
	}
	return id<<1 | pin
}
func Aig_FanoutObj(data []int32, id int32) []int32  { return data[id*5:] }
func Aig_FanoutPrev(data []int32, fan int32) *int32 { return &data[5*(fan>>1)+1+(fan&1)] }
func Aig_FanoutNext(data []int32, fan int32) *int32 { return &data[5*(fan>>1)+3+(fan&1)] }
func (f *FanoutIndex) ensure(id int) {
	size := len(f.data) / 5
	if id < size {
		return
	}
	next := 2 * (id + 1)
	if next < 4096 {
		next = 4096
	}
	f.data = append(f.data, make([]int32, 5*(next-size))...)
}
func (f *FanoutIndex) add(driver int, consumer int, pin int32) {
	f.ensure(driver)
	f.ensure(consumer)
	fan := Aig_FanoutCreate(int32(consumer), pin)
	prev, next := Aig_FanoutPrev(f.data, fan), Aig_FanoutNext(f.data, fan)
	if *prev != 0 || *next != 0 {
		panic("aig: fanout pin already linked")
	}
	first := &f.data[5*driver]
	if *first == 0 {
		*first = fan
		*prev = fan
		*next = fan
		return
	}
	tail := *Aig_FanoutPrev(f.data, *first)
	*prev = tail
	*next = *first
	*Aig_FanoutPrev(f.data, *first) = fan
	*Aig_FanoutNext(f.data, tail) = fan
}
func (f *FanoutIndex) remove(driver, consumer int, pin int32) {
	fan := Aig_FanoutCreate(int32(consumer), pin)
	if driver >= len(f.data)/5 || consumer >= len(f.data)/5 {
		panic("aig: fanout not allocated")
	}
	prev, next := Aig_FanoutPrev(f.data, fan), Aig_FanoutNext(f.data, fan)
	if *prev == 0 || *next == 0 || *Aig_FanoutNext(f.data, *prev) != fan || *Aig_FanoutPrev(f.data, *next) != fan {
		panic("aig: missing or corrupt fanout")
	}
	first := &f.data[5*driver]
	if *first == 0 {
		panic("aig: missing fanout list")
	}
	if *first == fan {
		if *next == fan {
			*first = 0
			*prev = 0
			*next = 0
			return
		}
		*first = *next
	}
	*Aig_FanoutNext(f.data, *prev) = *next
	*Aig_FanoutPrev(f.data, *next) = *prev
	*prev = 0
	*next = 0
}
func fanoutPins(n Node) int {
	switch n.Kind {
	case And, ExclusiveOr:
		return 2
	case Buffer, Output:
		return 1
	}
	return 0
}
func (f *FanoutIndex) Rebuild(p *Graph) {
	clear(f.data)
	f.ensure(len(p.nodes) - 1)
	for id, n := range p.nodes {
		if p.deleted[id] {
			continue
		}
		for pin := 0; pin < fanoutPins(n); pin++ {
			driver := n.Fanins[pin]
			if !p.valid(driver) {
				panic("aig: invalid fanout driver")
			}
			f.add(driver.NodeID(), id, int32(pin))
		}
	}
}
func Aig_ManFanoutStart(p *Graph) {
	if p.fanouts != nil {
		panic("aig: fanout index already active")
	}
	p.fanouts = new(FanoutIndex)
	p.fanouts.Rebuild(p)
}
func Aig_ManFanoutStop(p *Graph) {
	if p.fanouts == nil {
		panic("aig: fanout index not active")
	}
	p.fanouts = nil
}
func Aig_ManFanoutRebuild(p *Graph) {
	if p.fanouts != nil {
		p.fanouts.Rebuild(p)
	}
}
func Aig_ObjAddFanoutPin(p *Graph, driver, consumer Edge, pin int32) {
	if p.fanouts == nil || driver.Complemented() || consumer.Complemented() || !p.valid(driver) || !p.valid(consumer) {
		panic("aig: invalid fanout objects")
	}
	if pin < 0 || int(pin) >= fanoutPins(p.nodes[consumer.NodeID()]) || p.nodes[consumer.NodeID()].Fanins[pin].NodeID() != driver.NodeID() {
		panic("aig: consumer pin does not use driver")
	}
	p.fanouts.add(driver.NodeID(), consumer.NodeID(), pin)
}
func Aig_ObjRemoveFanoutPin(p *Graph, driver, consumer Edge, pin int32) {
	if p.fanouts == nil || driver.Complemented() || consumer.Complemented() || !p.valid(driver) || !p.valid(consumer) {
		panic("aig: invalid fanout objects")
	}
	if pin < 0 || int(pin) >= fanoutPins(p.nodes[consumer.NodeID()]) || p.nodes[consumer.NodeID()].Fanins[pin].NodeID() != driver.NodeID() {
		panic("aig: consumer pin does not use driver")
	}
	p.fanouts.remove(driver.NodeID(), consumer.NodeID(), pin)
}
func Aig_ObjAddFanout(p *Graph, driver, consumer Edge) {
	if p.fanouts == nil || !p.valid(consumer) {
		panic("aig: fanout index not active")
	}
	p.fanouts.ensure(consumer.NodeID())
	for pin := 0; pin < fanoutPins(p.nodes[consumer.NodeID()]); pin++ {
		fan := Aig_FanoutCreate(int32(consumer.NodeID()), int32(pin))
		if p.nodes[consumer.NodeID()].Fanins[pin].NodeID() == driver.NodeID() && *Aig_FanoutPrev(p.fanouts.data, fan) == 0 {
			Aig_ObjAddFanoutPin(p, driver, consumer, int32(pin))
			return
		}
	}
	panic("aig: no unlinked matching fanout pin")
}
func Aig_ObjRemoveFanout(p *Graph, driver, consumer Edge) {
	if p.fanouts == nil || !p.valid(consumer) {
		panic("aig: fanout index not active")
	}
	for pin := 0; pin < fanoutPins(p.nodes[consumer.NodeID()]); pin++ {
		fan := Aig_FanoutCreate(int32(consumer.NodeID()), int32(pin))
		if p.nodes[consumer.NodeID()].Fanins[pin].NodeID() == driver.NodeID() && *Aig_FanoutPrev(p.fanouts.data, fan) != 0 {
			Aig_ObjRemoveFanoutPin(p, driver, consumer, int32(pin))
			return
		}
	}
	panic("aig: no linked matching fanout pin")
}
func (f *FanoutIndex) Handles(driver Edge) []int32 {
	if driver.NodeID() >= len(f.data)/5 {
		return nil
	}
	first := f.data[5*driver.NodeID()]
	if first == 0 {
		return nil
	}
	var result []int32
	fan := first
	for {
		result = append(result, fan)
		fan = *Aig_FanoutNext(f.data, fan)
		if fan == first {
			return result
		}
		if fan == 0 || len(result) > len(f.data) {
			panic("aig: corrupt circular fanout list")
		}
	}
}
