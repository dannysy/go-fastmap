package fastmap

import (
	"sync/atomic"
	"unsafe"
)

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type skiplist struct {
	_     noCopy
	head  uintptr
	items []*item // used only to prevent GC collection
}

type item struct {
	_     noCopy
	next  uintptr
	value uintptr
	e     *entry // used only to prevent GC collection
}

func (l *skiplist) Len() (out uint64) {
	curr := atomic.LoadUintptr(&l.head)
	for curr != 0 {
		i := (*item)(unsafe.Pointer(curr))
		if atomic.LoadUintptr(&i.value) != 0 {
			out++
		}
		curr = atomic.LoadUintptr(&i.next)
	}
	return out
}

func (l *skiplist) Any() (value entry, ok bool) {
	curr := atomic.LoadUintptr(&l.head)
	for curr != 0 {
		i := (*item)(unsafe.Pointer(curr))
		if atomic.LoadUintptr(&i.value) != 0 {
			e := (*entry)(unsafe.Pointer(i.value))
			return *e, true
		}
		curr = atomic.LoadUintptr(&i.next)
	}
	return value, false
}

func (l *skiplist) Get(key string) (value entry, ok bool) {
	curr := atomic.LoadUintptr(&l.head)
	for curr != 0 {
		i := (*item)(unsafe.Pointer(curr))
		if atomic.LoadUintptr(&i.value) != 0 {
			e := (*entry)(unsafe.Pointer(i.value))
			if e.key == key {
				return *e, true
			}
		}
		curr = atomic.LoadUintptr(&i.next)
	}
	return value, false
}

func (l *skiplist) Delete(key string) (ok bool) {
	curr := atomic.LoadUintptr(&l.head)
	for curr != 0 {
		i := (*item)(unsafe.Pointer(curr))
		if atomic.LoadUintptr(&i.value) != 0 {
			e := (*entry)(unsafe.Pointer(i.value))
			if e.key == key {
				ov := atomic.SwapUintptr(&i.value, 0)
				if ov == 0 {
					//fmt.Printf("del key %s ov %v routine id %v\n", key, ov, internal.GoroutineId())
					return false
				}
				return true
			}
		}
		curr = atomic.LoadUintptr(&i.next)
	}
	return false
}

func (l *skiplist) CompareAndSwap(value *entry) (ok bool) {
	curr := atomic.LoadUintptr(&l.head)
	for curr != 0 {
		i := (*item)(unsafe.Pointer(curr))
		if atomic.LoadUintptr(&i.value) != 0 {
			e := (*entry)(unsafe.Pointer(i.value))
			if e.key == value.key {
				atomic.SwapUintptr(&i.value, uintptr(unsafe.Pointer(value)))
				return true
			}
		}
		curr = atomic.LoadUintptr(&i.next)
	}
	return false
}

func (l *skiplist) Add(value *entry) {
	i := item{
		next:  0,
		value: uintptr(unsafe.Pointer(value)),
		e:     value,
	}
	l.items = append(l.items, &i)
	if atomic.LoadUintptr(&l.head) == 0 {
		atomic.StoreUintptr(&l.head, uintptr(unsafe.Pointer(&i)))
		return
	}
	i.next = atomic.SwapUintptr(&l.head, uintptr(unsafe.Pointer(&i)))
}
