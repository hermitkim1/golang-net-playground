package main

import (
	"fmt"
	"net"
)

func main() {

	for _, CIDRBlock := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		ip, _, err := net.ParseCIDR(CIDRBlock)
		if err != nil {
			panic(fmt.Errorf("parse error on %v: %v", CIDRBlock, err))
		}

		if ip.IsPrivate() {
			fmt.Printf("%+v - IsPrivate\n", ip)
		} else if ip.IsGlobalUnicast() {
			fmt.Printf("%+v - IsGlobalUnicast\n", ip)
		} else if ip.IsInterfaceLocalMulticast() {
			fmt.Printf("%+v - IsInterfaceLocalMulticast\n", ip)
		} else if ip.IsLinkLocalMulticast() {
			fmt.Printf("%+v - IsLinkLocalMulticast\n", ip)
		} else if ip.IsLinkLocalUnicast() {
			fmt.Printf("%+v - IsLinkLocalUnicast\n", ip)
		} else if ip.IsLoopback() {
			fmt.Printf("%+v - IsLoopback\n", ip)
		} else if ip.IsMulticast() {
			fmt.Printf("%+v - IsMulticast\n", ip)
		} else if ip.IsUnspecified() {
			fmt.Printf("%+v - IsUnspecified\n", ip)
		} else {
			fmt.Printf("%+v - ???????\n", ip)
		}
	}

}
