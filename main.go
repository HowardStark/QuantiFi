package main

import (
	"io"
	"os"
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
	// activePcapManager is the PcapManager for the active network
	// interface.
	activePcapManager *PcapManager
)

func main() {
	InitLog(debugOut, infoOut, warningOut, errorOut)
	Info.Println("QuantiFi starting...")
	ifaceName, ifaceErr := FindActiveInterface()
	if ifaceErr != nil {
		Error.Println(ifaceErr.Error())
		os.Exit(1)
	}
	var pcapManagerErr error
	activePcapManager, pcapManagerErr = NewPcapManager(ifaceName, PcapDefaultSnapLen, PcapDefaultPromisc, PcapDefaultTimeout)
	if pcapManagerErr != nil {
		Error.Println(pcapManagerErr.Error())
		os.Exit(1)
	}
	Debug.Println(activePcapManager.interfaceName)
	activePcapManager.StartMonitor()
}
