package abc

// Abc_NtkAddPoBuffers inserts a node on each primary-output edge.
func Abc_NtkAddPoBuffers(n *Network) {
	if Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkAddPoBuffers requires a strashed network")
	}
	for _, output := range n.ObjectsOfKind(ObjectPO) {
		input := output.fanins[0]
		buffer := n.NewObject(ObjectNode)
		if err := n.AddFanin(buffer, input); err != nil {
			panic(err)
		}
		if err := n.PatchFanin(output, 0, buffer.Signal()); err != nil {
			panic(err)
		}
	}
}
