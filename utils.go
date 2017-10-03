package main

import (
	"encoding/binary"
	"net"
)

func Inet_ntoa(ipnr uint32, isBig bool) string {
	ip := net.IPv4(0, 0, 0, 0)
	var bo binary.ByteOrder
	if isBig {
		bo = binary.BigEndian
	} else {
		bo = binary.LittleEndian
	}
	bo.PutUint32([]byte(ip.To4()), ipnr)
	return ip.String()
}

func Inet_aton(ip string, isBig bool) uint32 {
	var bo binary.ByteOrder
	if isBig {
		bo = binary.BigEndian
	} else {
		bo = binary.LittleEndian
	}
	return bo.Uint32(
		[]byte(net.ParseIP(ip).To4()),
	)
}
