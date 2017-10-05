# win-route
Manipulate Windows routing table via system calls

## Example

```
package main

import (
  "flag"
  "net"

  log "github.com/Sirupsen/logrus"
)

func main() {
  gateway := flag.String("gateway", "", "interface (IPv4) address serving as a gateway")
  debug := flag.Bool("debug", false, "enable debug logging")
  flag.Parse()

  var gatewayAddress net.IP
  if *gateway != "" {
    gatewayAddress = net.ParseIP(*gateway)
    if gatewayAddress == nil {
      log.WithField("address", *gateway).Warn("Invalid gateway address specified")
    }
  }

  if *debug {
    log.SetLevel(log.DebugLevel)
  }

  intf := MustResolveInterface(net.ParseIP(*gateway))

  r := NewNetRoute()
  defer r.Close()

  // Make system call to get interface metric
  i, err := r.GetInterfaceByIndex(uint32(intf.Index))
  if err != nil {
    log.Fatal(err)
  }

  printRoutes(r)

  r1 := &IPForwardRow{
    ForwardDest:    Inet_aton("192.168.1.0", false),
    ForwardMask:    Inet_aton("255.255.255.0", false),
    ForwardNextHop: Inet_aton("172.22.101.121", false),
    ForwardIfIndex: i.InterfaceIndex,
    ForwardType:    3,
    ForwardProto:   3,
    ForwardMetric1: i.Metric, // route metric is 0 (+ interface metric)
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
      "dest":    dest,
      "mask":    mask,
      "gate":    gate,
      "metric":  route.ForwardMetric1,
      "ifIndex": route.ForwardIfIndex,
    }).Info("Route")
  }
}
```

## Credits

Special thanks to `github.com/yijunjun/route-table`

