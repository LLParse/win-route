package winroute

import (
	"encoding/binary"
	"errors"
	"net"

	log "github.com/Sirupsen/logrus"
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

func MustResolveInterface(gatewayAddress net.IP) net.Interface {
	i, err := ResolveInterface(gatewayAddress)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func ResolveInterface(gatewayAddress net.IP) (net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}

	var viableInterfaces []net.Interface

	for _, intf := range interfaces {
		// Skip down and loopback interfaces
		if (intf.Flags&net.FlagUp == 0) || (intf.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		var viableGatewayAddress net.IP

		for _, addr := range addrs {
			ipAddr, ok := addr.(*net.IPNet)

			// Skip loopback and link-local addresses
			if !ok || !ipAddr.IP.IsGlobalUnicast() {
				continue
			}

			// Skip IPv6 addresses
			if ipAddr.IP.To4() == nil {
				continue
			}

			// Skip addresses not matching target gateway, if specified
			if gatewayAddress != nil && !ipAddr.IP.Equal(gatewayAddress) {
				continue
			}

			viableGatewayAddress = ipAddr.IP
		}

		// Skip interfaces without a viable gateway address
		if viableGatewayAddress == nil {
			continue
		}

		viableInterfaces = append(viableInterfaces, intf)
	}

	for _, i := range viableInterfaces {
		log.WithFields(log.Fields{
			"index": i.Index,
			"name":  i.Name,
		}).Debug("Found viable interface")
	}

	switch len(viableInterfaces) {
	case 1:
		return viableInterfaces[0], nil
	case 0:
		return net.Interface{}, errors.New("No viable interface detected!")
	default:
		return net.Interface{}, errors.New("Multiple viable interfaces detected! Please specify a gateway address.")
	}
}
