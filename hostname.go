package main

import (
	"net"
	"os"
	"strings"
)

/* short and long hostname global variables */
var ServerShortHostname string
var ServerHostname string

/* GetHostname returns server short hostname */
func GetShortHostname() string {
	if len(ServerShortHostname) == 0 {
		var err error
		if ServerShortHostname, err = os.Hostname(); err != nil {
			ServerShortHostname = "unknown"
		}
	}
	return ServerShortHostname
}

/* GetHostname attempts to determine what is the fully
   qualified domain name of the current server */
func GetHostname() string {
	if len(ServerHostname) == 0 {
		// get list of ipv4 addresses associated with hostname
		AddrList, err := net.LookupIP(GetShortHostname())
		if err != nil {
			ServerHostname = GetShortHostname()
		}
		// resolve ipv4 addresses
		for _, addr := range AddrList {
			ipv4 := addr.To4()
			if ipv4 == nil {
				// not a valid ipv4, skip
				continue
			}
			// convert to ipv4 text representation
			ip, err := ipv4.MarshalText()
			if err != nil {
				ServerHostname = GetShortHostname()
				break
			}
			// forward resolve ipv4
			hosts, err := net.LookupAddr(string(ip))
			if err != nil {
				ServerHostname = GetShortHostname()
				break
			}
			// return server fqdn
			ServerHostname = strings.TrimSuffix(hosts[0], ".")
			break
		}
	}
	return ServerHostname
}
