package rando

import (
	"net"
	"strconv"
	"strings"
)

func Port() int {
	l, _ := net.Listen("tcp", ":0")
	defer l.Close()
	addrParts := strings.Split(l.Addr().String(), ":")
	port, _ := strconv.Atoi(addrParts[len(addrParts)-1])
	return port
}
