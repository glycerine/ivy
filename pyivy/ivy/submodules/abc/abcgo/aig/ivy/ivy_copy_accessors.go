package ivy

// Ivy_ManReadCopy returns the manager's borrowed temporary pCopy payload.
func Ivy_ManReadCopy(manager *Ivy_Man_t) any { return manager.pCopy }

// Ivy_ManSetCopy stores the temporary pCopy payload without changing its ownership.
func Ivy_ManSetCopy(manager *Ivy_Man_t, copy any) { manager.pCopy = copy }
