package main

import (
	"errors"
	"os/exec"
	"regexp"
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
	PcapDefaultSnapLen = 1048576
	// PcapFilterIncoming filters all packets whose destination is not
	// an IP in one of the 3 IPV4 private networks.
	PcapFilterIncoming = "(dst net 10.0.0.0 mask 255.0.0.0 or dst net 192.168.0.0 mask 255.255.0.0 or dst net 172.16.0.0 mask 255.240.0.0) and not (src net 10.0.0.0 mask 255.0.0.0 or src net 192.168.0.0 mask 255.255.0.0 or src net 172.16.0.0 mask 255.240.0.0)"
)

// PcapManager is a struct that encompasses all the functions needed
// for a single network interface.
type PcapManager struct {
	// interfaceName is the name of the active network interface
	interfaceName string
	// snapshotLen is the cutoff size for the packets we intercept
	snapshotLen int
	// promiscuousMode sets whether or not to switch the active
	// network interface to promiscuous mode.
	promiscuousMode bool
	// timeoutPacket is the amount of time we wait before giving
	// up on a packet.
	timeoutPacket time.Duration
	// pcapHandle provides an interface to the pcap handle.
	pcapHandle *pcap.Handle
	// addressRegex is the compiled regex to find the address
	// of the incoming recipient.
	addressRegex *regexp.Regexp
	// peerList contains the hwids of the peers on the
	// current network.
	peerList map[string]int
	// byteTotal for incoming packets
	byteTotal int
}

// NewPcapManager builds a PcapManager from the given arguments
func NewPcapManager(interfaceName string, snapshotLen int, promiscuousMode bool, timeoutPacket time.Duration) (*PcapManager, error) {
	regComp, regErr := regexp.Compile(`Address1=([a-zA-Z0-9:]+)`)
	if regErr != nil {
		return nil, regErr
	}
	pcapManager := &PcapManager{
		interfaceName:   interfaceName,
		snapshotLen:     snapshotLen,
		promiscuousMode: promiscuousMode,
		timeoutPacket:   timeoutPacket,
		addressRegex:    regComp,
		byteTotal:       0,
	}
	peerList, peerErr := pcapManager.GetPeerHwids()
	if peerErr != nil {
		return nil, peerErr
	}
	pcapManager.peerList = peerList
	return pcapManager, nil
}

// BuildHandle constructs a pcap handle interface for the current
// PcapManager.
func (pcapManager *PcapManager) BuildHandle() error {
	inactive, inactiveErr := pcap.NewInactiveHandle(pcapManager.interfaceName)
	defer inactive.CleanUp()
	if inactiveErr != nil {
		return inactiveErr
	}

	if monErr := inactive.SetRFMon(true); monErr != nil {
		return monErr
	}
	if snapErr := inactive.SetSnapLen(pcapManager.snapshotLen); snapErr != nil {
		return snapErr
	}
	if timeoutErr := inactive.SetTimeout(pcapManager.timeoutPacket); timeoutErr != nil {
		return timeoutErr
	}
	active, activeErr := inactive.Activate()
	if activeErr != nil {
		return activeErr
	}
	pcapManager.pcapHandle = active
	return nil
}

// StartMonitor puts the active network interface into promiscuous mode
// and begins to capture and print outgoing packets.
func (pcapManager *PcapManager) StartMonitor() {
	Info.Println("Starting to monitor interface \"" + pcapManager.interfaceName + "\"...")
	Info.Println("Building pcap handle interface...")
	pcapErr := pcapManager.BuildHandle()
	if pcapErr != nil {
		Error.Println(pcapErr.Error())
		return
	}
	Info.Println("Successfully opened pcap handle.")
	defer pcapManager.pcapHandle.Close()
	Info.Println("Starting to parse packets...")
	packetSource := gopacket.NewPacketSource(pcapManager.pcapHandle, pcapManager.pcapHandle.LinkType())
	for packet := range packetSource.Packets() {
		pcapManager.parsePacket(packet)
	}
}

func (pcapManager *PcapManager) parsePacket(packetInc gopacket.Packet) {
	data := packetInc.String()
	if strings.Contains(data, "Type=Data") {
		result := pcapManager.addressRegex.FindAllStringSubmatch(data, 1)[0][1]
		if _, ok := pcapManager.peerList[result]; ok {
			byteLen := len(packetInc.Data())
			Debug.Println("Received packet for " + result)
			Debug.Println("Size: ", byteLen)
			pcapManager.byteTotal = pcapManager.byteTotal + byteLen
			pcapManager.peerList[result] = pcapManager.peerList[result] + byteLen
		}
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

// GetPeerHwids builds a string array of the mac addresses of all devices
// that are with you on the network.
func (pcapManager *PcapManager) GetPeerHwids() (map[string]int, error) {
	Info.Println("Finding peer hwids...")
	hwids := make(map[string]int)
	cmdOut, cmdErr := exec.Command("/usr/sbin/arp", "-a").Output()
	if cmdErr != nil {
		return nil, cmdErr
	}
	cmdOutStr := string(cmdOut)
	for _, line := range strings.Split(cmdOutStr, "\n") {
		if line == "" {
			continue
		}
		parsed := strings.Split(line, " ")[3]
		if parsed == "ff:ff:ff:ff:ff:ff" {
			continue
		}
		hwids[parsed] = 0
	}
	return hwids, nil
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
