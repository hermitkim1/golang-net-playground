package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	cblog "github.com/cloud-barista/cb-log"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

const (
	ExitSetupSuccess = 0
	ExitSetupFailed  = 1
)

const (
	ENV_WG_TUN_FD  = "WG_TUN_FD"
	ENV_WG_UAPI_FD = "WG_UAPI_FD"
)

// const (
// 	tunDevice = "/dev/net/tun"
// 	ifReqSize = unix.IFNAMSIZ + 64
// )

// func checkErr(err error) {
// 	if err == nil {
// 		return
// 	}

// 	panic(err)
// }

// func createTUN() {
// 	nfd, err := unix.Open(tunDevice, os.O_RDWR, 0)
// 	checkErr(err)

// 	var ifr [ifReqSize]byte
// 	var flags uint16 = unix.IFF_TUN | unix.IFF_NO_PI
// 	name := []byte("wg1")
// 	copy(ifr[:], name)
// 	*(*uint16)(unsafe.Pointer(&ifr[unix.IFNAMSIZ])) = flags
// 	fmt.Println(string(ifr[:unix.IFNAMSIZ]))

// 	_, _, errno := unix.Syscall(
// 		unix.SYS_IOCTL,
// 		uintptr(nfd),
// 		uintptr(unix.TUNSETIFF),
// 		uintptr(unsafe.Pointer(&ifr[0])),
// 	)
// 	if errno != 0 {
// 		checkErr(fmt.Errorf("ioctl errno: %d", errno))
// 	}

// 	err = unix.SetNonblock(nfd, true)

// 	fd := os.NewFile(uintptr(nfd), tunDevice)
// 	checkErr(err)

// 	for {
// 		buf := make([]byte, 1500)
// 		_, err := fd.Read(buf)
// 		if err != nil {
// 			fmt.Printf("read error: %v\n", err)
// 			continue
// 		}

// 		fmt.Println("received packet")

// 		switch buf[0] & 0xF0 {
// 		case 0x40:
// 			fmt.Println("received ipv4")
// 			fmt.Printf("Length: %d\n", binary.BigEndian.Uint16(buf[2:4]))
// 			fmt.Printf("Protocol: %d (1=ICMP, 6=TCP, 17=UDP)\n", buf[9])
// 			fmt.Printf("Source IP: %s\n", net.IP(buf[12:16]))
// 			fmt.Printf("Destination IP: %s\n", net.IP(buf[16:20]))
// 		case 0x60:
// 			fmt.Println("received ipv6")
// 			fmt.Printf("Length: %d\n", binary.BigEndian.Uint16(buf[4:6]))
// 			fmt.Printf("Protocol: %d (1=ICMP, 6=TCP, 17=UDP)\n", buf[7])
// 			fmt.Printf("Source IP: %s\n", net.IP(buf[8:24]))
// 			fmt.Printf("Destination IP: %s\n", net.IP(buf[24:40]))
// 		}
// 	}
// }

// CBLogger represents a logger to show execution processes according to the logging level.
var CBLogger *logrus.Logger

func init() {
	// cblog is a global variable.
	configPath := filepath.Join("configs", "log_conf.yaml")
	CBLogger = cblog.GetLoggerWithConfigPath("cb-network", configPath)
}

func GetNetworkInformation() {
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
	targetIp := firstIP + 3
	if targetIp < lastIP-1 {
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

func availableInterfaces() {

	interfaces, err := net.Interfaces()

	if err != nil {
		fmt.Print(err)
		os.Exit(0)
	}

	fmt.Println("Available network interfaces on this machine : ")
	for _, i := range interfaces {
		//fmt.Printf("Name : %v \n", i.Name)
		fmt.Printf("Full object : %+v \n", i)
	}
}

func createAndUpTUN2() {
	interfaceName := "cbnet0"

	// get log level (default: info)

	logLevel := func() int {
		switch os.Getenv("LOG_LEVEL") {
		case "verbose", "debug":
			return device.LogLevelVerbose
		case "error":
			return device.LogLevelError
		case "silent":
			return device.LogLevelSilent
		}
		return device.LogLevelError
	}()

	// open TUN device (or use supplied fd)
	tun, err := func() (tun.Device, error) {
		tunFdStr := os.Getenv(ENV_WG_TUN_FD)
		if tunFdStr == "" {
			return tun.CreateTUN(interfaceName, device.DefaultMTU)
		}

		// // construct tun device from supplied fd

		// fd, err := strconv.ParseUint(tunFdStr, 10, 32)
		// if err != nil {
		// 	return nil, err
		// }

		// err = syscall.SetNonblock(int(fd), true)
		// if err != nil {
		// 	return nil, err
		// }

		// file := os.NewFile(uintptr(fd), "")
		// return tun.CreateTUNFromFile(file, device.DefaultMTU)
		return nil, fmt.Errorf("errrrrrr")
	}()

	if err == nil {
		realInterfaceName, err2 := tun.Name()
		if err2 == nil {
			interfaceName = realInterfaceName
		}
	}

	logger := device.NewLogger(
		logLevel,
		fmt.Sprintf("(%s) ", interfaceName),
	)

	if err != nil {
		logger.Errorf("Failed to create TUN device: %v", err)
		os.Exit(ExitSetupFailed)
	}

	// open UAPI file (or use supplied fd)

	// fileUAPI, err := func() (*os.File, error) {
	// 	uapiFdStr := os.Getenv(ENV_WG_UAPI_FD)
	// 	if uapiFdStr == "" {
	// 		return ipc.UAPIOpen(interfaceName)
	// 	}

	// 	// use supplied fd

	// 	fd, err := strconv.ParseUint(uapiFdStr, 10, 32)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return os.NewFile(uintptr(fd), ""), nil
	// }()

	MTU := device.DefaultMTU

	if err != nil {
		logger.Errorf("UAPI listen error: %v", err)
		os.Exit(ExitSetupFailed)
		return
	}

	// device := device.NewDevice(tun, conn.NewDefaultBind(), logger)

	logger.Verbosef("Device started")

	localIP := flag.String("local", "192.168.10.2/24", "Local tun interface IP/MASK like 192.168.3.3⁄24")

	// Set interface parameters
	runIP("link", "set", "dev", interfaceName, "mtu", strconv.Itoa(MTU))
	runIP("addr", "add", *localIP, "dev", interfaceName)
	runIP("link", "set", "dev", interfaceName, "up")

	errs := make(chan error)
	term := make(chan os.Signal, 1)

	// uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	// if err != nil {
	// 	logger.Errorf("Failed to listen on uapi socket: %v", err)
	// 	os.Exit(ExitSetupFailed)
	// }

	// go func() {
	// 	for {
	// 		conn, err := uapi.Accept()
	// 		if err != nil {
	// 			errs <- err
	// 			return
	// 		}
	// 		go device.IpcHandle(conn)
	// 	}
	// }()

	logger.Verbosef("UAPI listener started")

	// wait for program to terminate

	// signal.Notify(term, syscall.SIGTERM)
	// signal.Notify(term, os.Interrupt)

	select {
	case <-term:
	case <-errs:
		// case <-device.Wait():
	}

	// clean up

	// uapi.Close()
	// device.Close()

	logger.Verbosef("Shutting down")
}

func createAndUpTUN() {
	interfaceName := "cbnet0"

	// get log level (default: info)

	logLevel := func() int {
		switch os.Getenv("LOG_LEVEL") {
		case "verbose", "debug":
			return device.LogLevelVerbose
		case "error":
			return device.LogLevelError
		case "silent":
			return device.LogLevelSilent
		}
		return device.LogLevelError
	}()

	// open TUN device (or use supplied fd)
	tun, err := func() (tun.Device, error) {
		tunFdStr := os.Getenv(ENV_WG_TUN_FD)
		if tunFdStr == "" {
			return tun.CreateTUN(interfaceName, device.DefaultMTU)
		}

		// // construct tun device from supplied fd

		// fd, err := strconv.ParseUint(tunFdStr, 10, 32)
		// if err != nil {
		// 	return nil, err
		// }

		// err = syscall.SetNonblock(int(fd), true)
		// if err != nil {
		// 	return nil, err
		// }

		// file := os.NewFile(uintptr(fd), "")
		// return tun.CreateTUNFromFile(file, device.DefaultMTU)
		return nil, fmt.Errorf("errrrrrr")
	}()

	if err == nil {
		realInterfaceName, err2 := tun.Name()
		if err2 == nil {
			interfaceName = realInterfaceName
		}
	}

	logger := device.NewLogger(
		logLevel,
		fmt.Sprintf("(%s) ", interfaceName),
	)

	if err != nil {
		logger.Errorf("Failed to create TUN device: %v", err)
		os.Exit(ExitSetupFailed)
	}

	// open UAPI file (or use supplied fd)

	// fileUAPI, err := func() (*os.File, error) {
	// 	uapiFdStr := os.Getenv(ENV_WG_UAPI_FD)
	// 	if uapiFdStr == "" {
	// 		return ipc.UAPIOpen(interfaceName)
	// 	}

	// 	// use supplied fd

	// 	fd, err := strconv.ParseUint(uapiFdStr, 10, 32)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return os.NewFile(uintptr(fd), ""), nil
	// }()

	MTU := device.DefaultMTU

	if err != nil {
		logger.Errorf("UAPI listen error: %v", err)
		os.Exit(ExitSetupFailed)
		return
	}

	// device := device.NewDevice(tun, conn.NewDefaultBind(), logger)

	logger.Verbosef("Device started")

	localIP := flag.String("local", "192.168.10.2/24", "Local tun interface IP/MASK like 192.168.3.3⁄24")

	// Set interface parameters
	runIP("link", "set", "dev", interfaceName, "mtu", strconv.Itoa(MTU))
	runIP("addr", "add", *localIP, "dev", interfaceName)
	runIP("link", "set", "dev", interfaceName, "up")

	errs := make(chan error)
	term := make(chan os.Signal, 1)

	// uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	// if err != nil {
	// 	logger.Errorf("Failed to listen on uapi socket: %v", err)
	// 	os.Exit(ExitSetupFailed)
	// }

	// go func() {
	// 	for {
	// 		conn, err := uapi.Accept()
	// 		if err != nil {
	// 			errs <- err
	// 			return
	// 		}
	// 		go device.IpcHandle(conn)
	// 	}
	// }()

	// logger.Verbosef("UAPI listener started")

	// // wait for program to terminate

	// signal.Notify(term, syscall.SIGTERM)
	// signal.Notify(term, os.Interrupt)

	select {
	case <-term:
	case <-errs:
		// case <-device.Wait():
	}

	// clean up

	// uapi.Close()
	// device.Close()

	logger.Verbosef("Shutting down")
}

func runIP(args ...string) {
	CBLogger.Debug("Start.........")

	CBLogger.Trace(args)

	cmd := exec.Command("/sbin/ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		CBLogger.Fatal("Error running /sbin/ip:", err)
	}

	CBLogger.Debug("End.........")
}

func main() {
	// GetNetworkInformation()
	// availableInterfaces()
	// createTUN()
	// createAndUpTUN()
	createAndUpTUN2()

}
