package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/google/gopacket/pcap"
)

var (
	// DebugOut redirects debug logs to os.Stdout
	debugOut io.Writer = os.Stdout
	// InfoOut redirects debug logs to os.Stdout
	infoOut = os.Stdout
	// WarningOut redirects debug logs to os.Stdout
	warningOut = os.Stdout
	// ErrorOut redirects debug logs to os.Stdout
	errorOut = os.Stderr
)

func main() {
	InitLog(debugOut, infoOut, warningOut, errorOut)
	Info.Println("QuantiFi starting...")
	iface, err := FindActiveInterface()
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
