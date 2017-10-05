package winroute

import "unsafe"

type DynamicMemory struct {
	mem []byte
}

func newDynamicMemory(bytes uint32) *DynamicMemory {
	return &DynamicMemory{
		mem: make([]byte, bytes, bytes),
	}
}

func (d *DynamicMemory) Len() uint32 {
	return uint32(len(d.mem))
}

func (d *DynamicMemory) Address() uintptr {
	return (*SliceHeader)(unsafe.Pointer(&d.mem)).Addr
}
