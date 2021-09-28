package main

import (
	"fmt"
	"net"
	"path/filepath"

	cblog "github.com/cloud-barista/cb-log"
	"github.com/sirupsen/logrus"
)

// CBLogger represents a logger to show execution processes according to the logging level.
var CBLogger *logrus.Logger

func init() {
	// cblog is a global variable.
	configPath := filepath.Join("..", "configs", "log_conf.yaml")
	CBLogger = cblog.GetLoggerWithConfigPath("cb-network", configPath)
}

// initMap initializes a map with an integer key starting at 1
func initMap(keyFrom int, keyTo int, initValue bool) map[int]bool {
	m := make(map[int]bool)
	for i := keyFrom; i <= keyTo; i++ {
		m[i] = initValue
	}
	return m
}

// GetAvailableCIDRBlocks represents a function to check and return available CIDR blocks
func GetAvailableCIDRBlocks(ips []string) {
	CBLogger.Debug("Start.........")

	CBLogger.Tracef("IPs: %v", ips)

	ip10, ipnet10, _ := net.ParseCIDR("10.0.0.0/8")
	prefixMap10 := initMap(8, 32, true)
	ip172, ipnet172, _ := net.ParseCIDR("172.16.0.0/12")
	prefixMap172 := initMap(12, 32, true)
	ip192, ipnet192, _ := net.ParseCIDR("192.168.0.0/16")
	prefixMap192 := initMap(16, 32, true)

	for _, ipStr := range ips {
		// CBLogger.Tracef("i: %v", i)
		// CBLogger.Tracef("IP: %v", ipStr)

		ip, ipnet, _ := net.ParseCIDR(ipStr)
		// Get CIDR Prefix
		cidrPrefix, _ := ipnet.Mask.Size()

		if ipnet10.Contains(ip) {
			CBLogger.Tracef("'%s' contains '%s/%v'", ipnet10, ip, cidrPrefix)
			prefixMap10[cidrPrefix] = false

		} else if ipnet172.Contains(ip) {
			CBLogger.Tracef("'%s' contains '%s/%v'", ipnet172, ip, cidrPrefix)
			prefixMap172[cidrPrefix] = false

		} else if ipnet192.Contains(ip) {
			CBLogger.Tracef("'%s' contains '%s/%v'", ipnet192, ip, cidrPrefix)
			prefixMap192[cidrPrefix] = false

		} else {
			CBLogger.Tracef("Nothing contains '%s/%v'", ip, cidrPrefix)
		}
	}

	// net10
	availableIPNet10 := make([]string, 32)
	j := 0
	for cidrPrefix, isTrue := range prefixMap10 {
		if isTrue {
			ipNet := fmt.Sprint(ip10, "/", cidrPrefix)
			CBLogger.Tracef("'%s' is possible for a virtual network.", ipNet)
			availableIPNet10[j] = ipNet
			j++
		}
	}

	// net172
	availableIPNet172 := make([]string, 32)
	j = 0
	for cidrPrefix, isTrue := range prefixMap172 {
		if isTrue {
			ipNet := fmt.Sprint(ip172, "/", cidrPrefix)
			CBLogger.Tracef("'%s' is possible for a virtual network.", ipNet)
			availableIPNet172[j] = ipNet
			j++
		}
	}

	// net192
	availableIPNet192 := make([]string, 32)
	j = 0
	for cidrPrefix, isTrue := range prefixMap192 {
		if isTrue {
			ipNet := fmt.Sprint(ip192, "/", cidrPrefix)
			CBLogger.Tracef("'%s' is possible for a virtual network.", ipNet)
			availableIPNet192[j] = ipNet
			j++
		}
	}

	CBLogger.Tracef("Available IPNets in 10.0.0.0/8 : %v", availableIPNet10)
	CBLogger.Tracef("Available IPNets in 172.16.0.0/12 : %v", availableIPNet172)
	CBLogger.Tracef("Available IPNets in 192.168.0.0/16 : %v", availableIPNet192)

	CBLogger.Debug("End.........")
}

func main() {

	var dummyIPs = []string{
		"10.0.2.2/8",
		"10.0.2.3/10",
		"10.0.2.4/12",
		"10.0.2.5/18",
		"10.0.2.6/20",
		"172.16.10.12/12",
		"172.16.10.13/14",
		"172.16.10.14/16",
		"172.16.10.15/18",
		"172.16.10.16/24",
		"192.168.2.22/16",
		"192.168.2.23/18",
		"192.168.2.24/20",
		"192.168.2.25/22",
		"192.168.2.26/24"}

	GetAvailableCIDRBlocks(dummyIPs)

}
