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
	doneCh := make(chan bool, r.slotCount)
	for _, s := range r.slots {
		go func() {
			r.mutex.RLock()
			defer r.mutex.RUnlock()
			select {
			case doneCh <- s.Lookup(b):
			default:
			}
		}()
	}
	for {
		select {
		case ret := <-doneCh:
			if ret {
				return true
			}
		default:
			if len(doneCh) == 0 {
				break
			}
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
