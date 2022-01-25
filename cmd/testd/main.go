package main

import (
	"fmt"

	"inet.af/netaddr"
)

func main() {
	ip := netaddr.IPPort{}
	fmt.Println(ip.String())
}
