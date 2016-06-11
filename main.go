package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/gopacket/pcap"
)

var (
	// debugOut redirects debug logs to os.Stdout
	debugOut io.Writer = os.Stdout
	// infoOut redirects debug logs to os.Stdout
	infoOut = os.Stdout
	// warningOut redirects debug logs to os.Stdout
	warningOut = os.Stdout
	// errorOut redirects debug logs to os.Stdout
	errorOut = os.Stderr
	// iface is the active net interface
	iface string
	// snapshotLen is the maximum size that will be read from any packet
	snapshotLen int32 = 1024
	// promisc decides whether or not to put the interface into promiscuous mode
	promisc = true
	// timeout duration for the packet cutoff
	timeout = 30 * time.Second
	// handle provides an interface to the pcap handle
	handle *pcap.Handle
)

func main() {
	InitLog(debugOut, infoOut, warningOut, errorOut)
	Info.Println("QuantiFi starting...")
	var err error
	iface, err = FindActiveInterface()
	if err != nil {
		Error.Println(err.Error())
		os.Exit(1)
	}
	Info.Println(iface)
}

// GetInterfaces prints the status of all the current network interfaces
func GetInterfaces() error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}
	Info.Println("Devices found:")
	for _, device := range devices {
		Info.Println("---")
		Info.Println("Name: ", device.Name)
		Info.Println("Description: ", device.Description)
		Info.Println("Devices addresses: ", device.Description)
		for _, address := range device.Addresses {
			Info.Println("- IP address: ", address.IP)
			Info.Println("- Subnet mask: ", address.Netmask)
		}
	}
	return nil
}

// FindActiveInterface uses the platform native route command to determine
// which network interface will be used for WAN connections.
func FindActiveInterface() (string, error) {
	Debug.Println(runtime.GOOS)
	switch runtime.GOOS {
	case "linux":
		cmdOut, cmdErr := exec.Command("/sbin/ip", "route", "get", "8.8.8.8").Output()
		if cmdErr != nil {
			return "", cmdErr
		}
		interfaceName := strings.Split(string(cmdOut), " ")[4]
		return interfaceName, nil
	case "darwin":
		cmdOut, cmdErr := exec.Command("/sbin/route", "get", "8.8.8.8").Output()
		if cmdErr != nil {
			return "", cmdErr
		}
		tempOut := string(cmdOut)
		if tempOut == "route: writing to routing socket: not in table" {
			return "", errors.New(tempOut)
		}
		var ifaceName = ""
		for _, line := range strings.Split(tempOut, "\n") {
			if strings.Contains(string(line), "interface: ") {
				ifaceName = strings.Split(string(line), "interface: ")[1]
			}
		}
		if ifaceName == "" {
			return "", errors.New("quantifi: could not find active interface")
		}
		return ifaceName, nil
	default:
		return "", errors.New("quantifi: operating system " + runtime.GOOS + " is not supported")
	}
}
