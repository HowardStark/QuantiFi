package main

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	// PcapDefaultPromisc is true by default, since you should rarely
	// need to not be in promiscuous mode.
	PcapDefaultPromisc = true
	// PcapDefaultTimeout is set to 30 seconds, solely because it
	// seemed like a sensible default.
	PcapDefaultTimeout = 30 * time.Second
	// PcapDefaultSnapLen is set to 1024 for testing purposes, but
	// will probably need to be increased since we are measuring the
	// length of the packets.
	PcapDefaultSnapLen int32 = 1024
)

// PcapManager is a struct that encompasses all the functions needed
// for a single network interface.
type PcapManager struct {
	// interfaceName is the name of the active network interface
	interfaceName string
	// snapshotLen is the cutoff size for the packets we intercept
	snapshotLen int32
	// promiscuousMode sets whether or not to switch the active
	// network interface to promiscuous mode.
	promiscuousMode bool
	// timeoutPacket is the amount of time we wait before giving
	// up on a packet.
	timeoutPacket time.Duration
	// pcapHandle provides an interface to the pcap handle.
	pcapHandle *pcap.Handle
}

// NewPcapManager builds a PcapManager from the given arguments
func NewPcapManager(interfaceName string, snapshotLen int32, promiscuousMode bool, timeoutPacket time.Duration) *PcapManager {
	pcapManager := &PcapManager{
		interfaceName:   interfaceName,
		snapshotLen:     snapshotLen,
		promiscuousMode: promiscuousMode,
		timeoutPacket:   timeoutPacket,
	}
	return pcapManager
}

// StartMonitor puts the active network interface into promiscuous mode
// and begins to capture and print outgoing packets.
func (pcapManager *PcapManager) StartMonitor() {
	Info.Println("Starting to monitor interface \"" + pcapManager.interfaceName + "\"...")
	var pcapErr error
	pcapManager.pcapHandle, pcapErr = pcap.OpenLive(pcapManager.interfaceName, pcapManager.snapshotLen, pcapManager.promiscuousMode, pcapManager.timeoutPacket)
	if pcapErr != nil {
		Error.Println(pcapErr.Error())
		return
	}
	Info.Println("Successfully opened pcap handle.")
	defer pcapManager.pcapHandle.Close()
	Info.Println("Starting to parse packets...")
	packetSource := gopacket.NewPacketSource(pcapManager.pcapHandle, pcapManager.pcapHandle.LinkType())
	for packet := range packetSource.Packets() {
		Debug.Println(packet)
	}
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

// GetInterfaces prints the status of all the current network interfaces
func (pcapManager *PcapManager) GetInterfaces() error {
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
