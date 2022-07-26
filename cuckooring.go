package cuckoo

import "sync"

type CuckooRing struct {
	slotCapacity uint
	slotPosition uint
	slotCount    uint
	entryCounter uint
	slots        []*Filter
	mutex        sync.RWMutex
}

func NewCuckooRing(slot, capacity uint) *CuckooRing {
	r := &CuckooRing{
		slotCapacity: capacity / slot,
		slotCount:    slot,
		slots:        make([]*Filter, slot),
	}
	for i := 0; i < int(slot); i++ {
		r.slots[i] = NewFilter(r.slotCapacity)
	}
	return r
}
func (r *CuckooRing) Add(b []byte) {
	if r == nil {
		return
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.add(b)
}

func (r *CuckooRing) add(b []byte) {
	slot := r.slots[r.slotPosition]
	if r.entryCounter > r.slotCapacity {
		// Move to next slot and reset
		r.slotPosition = (r.slotPosition + 1) % r.slotCount
		slot = r.slots[r.slotPosition]
		slot.Reset()
		r.entryCounter = 0
	}
	r.entryCounter++
	slot.Insert(b)
}

func (r *CuckooRing) Test(b []byte) bool {
	if r == nil {
		return false
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	test := r.test(b)
	return test
}

func (r *CuckooRing) test(b []byte) bool {
	for _, s := range r.slots {
		if s.Lookup(b) {
			return true
		}
	}
	return false
}

func (r *CuckooRing) Check(b []byte) bool {
	if r.Test(b) {
		return true
	}
	r.Add(b)
	return false
}
