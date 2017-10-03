package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
)

func init() {
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	r := NewNetRoute()
	defer r.Close()

	ifIndex := uint32(10)
	i, err := r.GetInterfaceByIndex(ifIndex)
	if err != nil {
		log.Fatal(err)
	}

	printRoutes(r)

	r1 := &IPForwardRow{
		ForwardDest:    Inet_aton("192.168.1.0", false),
		ForwardMask:    Inet_aton("255.255.255.0", false),
		ForwardNextHop: Inet_aton("172.22.101.121", false),
		ForwardIfIndex: ifIndex,
		ForwardType:    3,
		ForwardProto:   3,
		ForwardMetric1: i.Metric,
	}

	if err = r.AddRoute(r1); err == nil {
		log.Info("Added route")
		printRoutes(r)

		if err = r.DeleteRoute(r1); err == nil {
			log.Info("Deleted route")
			printRoutes(r)
		} else {
			log.Warn("Error deleting route")
		}
	} else {
		log.Warn("Error adding route")
	}
}

func printRoutes(r *NetRoute) {
	routes, err := r.GetRoutes()
	if err != nil {
		log.Error("Error getting routes")
	}
	for _, route := range routes {
		dest := Inet_ntoa(route.ForwardDest, false)
		mask := Inet_ntoa(route.ForwardMask, false)
		gate := Inet_ntoa(route.ForwardNextHop, false)
		log.WithFields(log.Fields{
			"dest": dest,
			"mask": mask,
			"gate": gate,
		}).Info("Route")
	}
}
