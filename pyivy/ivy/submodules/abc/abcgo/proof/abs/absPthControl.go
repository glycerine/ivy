package abs

import "sync"

// absProverControl tracks the asynchronous prover state used by ABC.
// Like the C pthread code, cancellation does not discard a published proof.
type absProverControl struct {
	mu     sync.Mutex
	runID  int32
	proved bool
}

func (c *absProverControl) start(solve func(int32) int32) int32 {
	c.mu.Lock()
	c.runID++
	id := c.runID
	c.proved = false
	c.mu.Unlock()
	go func() {
		result := solve(id)
		c.mu.Lock()
		if result == 1 {
			c.proved = true
		}
		c.mu.Unlock()
	}()
	return id
}

func (c *absProverControl) cancel() {
	c.mu.Lock()
	c.runID++
	c.mu.Unlock()
}

func (c *absProverControl) stopped(id int32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if id > c.runID {
		panic("abs: prover run ID exceeds current ID")
	}
	return id < c.runID
}

func (c *absProverControl) check() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.proved {
		return 0
	}
	c.proved = false
	return 1
}
