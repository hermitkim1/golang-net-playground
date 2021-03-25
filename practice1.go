package main

import (
	"encoding/binary"
	"fmt"
	"github.com/cloud-barista/cb-log"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"path/filepath"
)

// CBLogger represents a logger to show execution processes according to the logging level.
var CBLogger *logrus.Logger

func init() {
	// cblog is a global variable.
	configPath := filepath.Join("configs", "log_conf.yaml")
	CBLogger = cblog.GetLoggerWithConfigPath("cb-network", configPath)
}

func main() {
	networkCIDRBlock := "192.168.10.0/23"

	// Get IPNet struct from string
	_, ipv4Net, err := net.ParseCIDR(networkCIDRBlock)
	if err != nil {
		log.Fatal(err)
	}
	// Get NetworkAddress(uint32) (The first IP address of this network)
	firstIP := binary.BigEndian.Uint32(ipv4Net.IP)
	CBLogger.Tracef("The first IP address of the network: %v", firstIP)

	// Get Subnet Mask(uint32) from IPNet struct
	subnetMask := binary.BigEndian.Uint32(ipv4Net.Mask)
	CBLogger.Tracef("The subnet mask of the network: %v", subnetMask)

	// Get BroadcastAddress(uint32) (The last IP address of this network)
	lastIP := (firstIP & subnetMask) | (subnetMask ^ 0xffffffff)
	CBLogger.Tracef("The last IP address of the network: %v", lastIP)

	// Create IP address of type net.IP. IPv4 is 4 bytes, IPv6 is 16 bytes.
	var ip = make(net.IP, 4)
	targetIp := firstIP+3
	if targetIp < lastIP-1{
		binary.BigEndian.PutUint32(ip, targetIp)
	}

	// Get CIDR Prefix
	cidrPrefix, _ := ipv4Net.Mask.Size()
	// Create Host IP CIDR Block
	ipCIDRBlock := fmt.Sprint(ip, "/", cidrPrefix)
	// To string IP Address
	ipAddress := fmt.Sprint(ip)

	CBLogger.Tracef("CIDR prefix: %v", cidrPrefix)
	CBLogger.Tracef("IP CIDR block: %v", ipCIDRBlock)
	CBLogger.Tracef("IP address: %v", ipAddress)

}
