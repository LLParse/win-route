package winroute

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

type NetRoute struct {
	iphlpapi             *syscall.DLL
	getIpInterfaceEntry  *syscall.Proc
	getIpForwardTable    *syscall.Proc
	createIpForwardEntry *syscall.Proc
	setIpForwardEntry    *syscall.Proc
	deleteIpForwardEntry *syscall.Proc
	systemError          map[uint32]string
}

func NewNetRoute() *NetRoute {
	iphlpapi := syscall.MustLoadDLL("iphlpapi.dll")
	return &NetRoute{
		iphlpapi:             iphlpapi,
		getIpInterfaceEntry:  iphlpapi.MustFindProc("GetIpInterfaceEntry"),
		getIpForwardTable:    iphlpapi.MustFindProc("GetIpForwardTable"),
		createIpForwardEntry: iphlpapi.MustFindProc("CreateIpForwardEntry"),
		setIpForwardEntry:    iphlpapi.MustFindProc("SetIpForwardEntry"),
		deleteIpForwardEntry: iphlpapi.MustFindProc("DeleteIpForwardEntry"),

		// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681382(v=vs.85).aspx
		systemError: map[uint32]string{
			0:    "ERROR_SUCCESS",
			2:    "ERROR_FILE_NOT_FOUND",
			5:    "ERROR_ACCESS_DENIED",
			50:   "ERROR_NOT_SUPPORTED",
			87:   "ERROR_INVALID_PARAMETER",
			122:  "ERROR_INSUFFICIENT_BUFFER",
			1168: "ERROR_NOT_FOUND",
		},
	}
}

func (r *NetRoute) Close() {
	r.iphlpapi.Release()
}

func (r *NetRoute) GetInterfaceByIndex(index uint32) (*IPInterfaceEntry, error) {
	ie := &IPInterfaceEntry{
		Family:         2, // AF_INET (IPv4)
		InterfaceIndex: index,
	}

	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa814417(v=vs.85).aspx
	r1, r2, err := r.getIpInterfaceEntry.Call(uintptr(unsafe.Pointer(ie)))
	log.Debugf("%+v", ie)
	return ie, r.buildError(r1, r2, err)
}

func (r *NetRoute) GetRoutes() ([]IPForwardRow, error) {
	// 112 bytes per route, assume 256 routes by default
	bufSize := uint32(112 * 256)
	buf := newDynamicMemory(bufSize)

	r1, r2, err := r.getIpForwardTable.Call(
		buf.Address(),
		uintptr(unsafe.Pointer(&bufSize)),
		0,
	)

	for r1 == 122 {
		log.WithFields(log.Fields{
			"cur_bufsize": len(buf.mem),
			"req_bufsize": bufSize,
		}).Warn("Insufficient Buffer")

		buf = newDynamicMemory(bufSize)

		r1, r2, err = r.getIpForwardTable.Call(
			buf.Address(),
			uintptr(unsafe.Pointer(&bufSize)),
			0,
		)
	}

	if r1 != 0 {
		return nil, r.buildError(r1, r2, err)
	}

	num := *(*uint32)(unsafe.Pointer(buf.Address()))

	rows := []IPForwardRow{}
	sh_rows := (*SliceHeader)(unsafe.Pointer(&rows))
	sh_rows.Addr = buf.Address() + 4
	sh_rows.Len = int(num)
	sh_rows.Cap = int(num)
	return rows, nil
}

// AddRoute adds an IPForwardRow to the routing table.
func (r *NetRoute) AddRoute(route *IPForwardRow) error {
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365860(v=vs.85).aspx
	r1, r2, err := r.createIpForwardEntry.Call(uintptr(unsafe.Pointer(route)))
	return r.buildError(r1, r2, err)
}

func (r *NetRoute) UpdateRoute(route *IPForwardRow) error {
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa366363(v=vs.85).aspx
	r1, r2, err := r.setIpForwardEntry.Call(uintptr(unsafe.Pointer(route)))
	return r.buildError(r1, r2, err)
}

func (r *NetRoute) DeleteRoute(route *IPForwardRow) error {
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365878(v=vs.85).aspx
	r1, r2, err := r.deleteIpForwardEntry.Call(uintptr(unsafe.Pointer(route)))
	return r.buildError(r1, r2, err)
}

func (r *NetRoute) buildError(r1 uintptr, r2 uintptr, err error) error {
	log.Debugf("r1=%v r2=%v err=%+v", r1, r2, err)
	if r1 == 0 {
		return nil
	}
	if m, ok := r.systemError[uint32(r1)]; ok {
		return errors.New(m)
	}
	return errors.New(fmt.Sprintf("ERROR CODE %d", r1))
}
