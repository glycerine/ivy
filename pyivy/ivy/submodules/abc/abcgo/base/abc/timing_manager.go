package abc

// NetworkTimingManager returns the source-compatible timing state attached to n.
func NetworkTimingManager(n *Network) any {
	if n == nil {
		return nil
	}
	return n.TimingManager
}

// NetworkSetTimingManager attaches source-compatible timing state to n.
func NetworkSetTimingManager(n *Network, manager any) {
	if n != nil {
		n.TimingManager = manager
	}
}
