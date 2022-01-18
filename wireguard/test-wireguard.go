package main

import (
	"C"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/net/ipv4"
)

// const (
// 	ExitSetupSuccess = 0
// 	ExitSetupFailed  = 1
// )

// const (
// 	ENV_WG_TUN_FD             = "WG_TUN_FD"
// 	ENV_WG_UAPI_FD            = "WG_UAPI_FD"
// 	ENV_WG_PROCESS_FOREGROUND = "WG_PROCESS_FOREGROUND"
// )

const (
	cIFFTUN  = 0x0001
	cIFFTAP  = 0x0002
	cIFFNOPI = 0x1000
	IFNAMSIZ = 0x10
)

const (
	// TUNSETIFF is from the syscall package which is the standard lib
	TUNSETIFF = 0x400454ca
)

const (
	SYS_IOCTL = 16
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func main() {

	ifName := "mymy0"

	fd, err := syscall.Open("/dev/net/tun", os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		log.Fatal(err)
	}

	fdInt := uintptr(fd)

	// Setup Fd

	var flags uint16 = syscall.IFF_NO_PI
	flags |= syscall.IFF_TUN

	// Create an interface
	var req ifReq

	req.Flags = flags
	copy(req.Name[:], ifName)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fdInt, uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		return
	}

	createdIFName := strings.Trim(string(req.Name[:]), "\x00")

	fmt.Printf("createdIFName: %s\n", createdIFName)

	// if errno != 0 {
	// 	log.Fatal(errno)
	// }
	// unix.SetNonblock(tunfd, true)

	// fd := os.NewFile(uintptr(tunfd), "/dev/net/tun")

	tunFd := os.NewFile(fdInt, "tun")

	wait := sync.WaitGroup{}
	wait.Add(1)

	go func() {
		var err error

		// tunName := "cbnet0"
		runIP("link", "set", "dev", ifName, "mtu", "1420")
		runIP("addr", "add", "192.168.10.1/26", "dev", ifName)
		runIP("link", "set", "dev", ifName, "up")

		// c := exec.Command("sh", "-c", "ip link set up cheese && ip a a 192.168.9.2/24 dev cheese")
		// c.Start()
		// c.Wait()
		// exec.Command("sh", "-c", "ping -c 4 -f 192.168.9.1; ip link set down cheese; ip a f dev cheese").Start()
		time.Sleep(2 * time.Second)
		b := [2000]byte{}
		for {
			n, err := tunFd.Read(b[:])
			if err != nil {
				break
			}
			fmt.Printf("Read %d bytes\n", n)
			// b2 := b[:n]

			// fmt.Printf("Str: %s\n", string(b2[20:]))

			// fmt.Printf("Source: %s\n", net.IPv4(b2[12], b2[13], b2[14], b2[15]))
			// fmt.Printf("Destination: %s\n", net.IPv4(b2[16], b2[17], b2[18], b2[19]))

			header, _ := ipv4.ParseHeader(b[:n])
			fmt.Printf("SRC: %+v\n", header.Src)
			fmt.Printf("Des: %+v\n", header.Dst)

			log.Printf("Header %+v\n", header)
		}
		log.Print("Read errored: ", err)
		wait.Done()
	}()
	time.Sleep(time.Second * 15)
	log.Print("Closing")
	err = tunFd.Close()
	if err != nil {
		log.Print("Close errored: ", err)
	}
	wait.Wait()
	log.Print("Exiting")

}

func runIP(args ...string) {
	fmt.Println("Start.........")

	fmt.Println(args)

	cmd := exec.Command("/sbin/ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		fmt.Println("Error running /sbin/ip:", err)
	}

	fmt.Println("End.........")
}
