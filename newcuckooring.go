package cuckoo

import "sync"

type CuckooRing_ struct {
	pool         []Filter
	slotPosition int
	slotCapacity uint
	mutex        sync.RWMutex
}

func NewCuckooRing1(capacity uint) *CuckooRing_ {
	ring := &CuckooRing_{
		pool: make([]Filter, 2),
	}
	ring.slotCapacity = capacity / 2
	for i := 0; i < 2; i++ {
		ring.pool[i] = *NewFilter(ring.slotCapacity)
	}
	return ring
}

func (c *CuckooRing_) Add(b []byte) {
	slot := c.pool[c.slotPosition]
	if slot.Count() > c.slotCapacity {
		// Move to next slot and reset
		c.slotPosition = (c.slotPosition + 1) % 2
		slot = c.pool[c.slotPosition]
		slot.Reset()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	slot.Insert(b)
}

func (c *CuckooRing_) Test(b []byte) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.pool[0].Lookup(b) || c.pool[1].Lookup(b)
}
