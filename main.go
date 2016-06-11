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
	ifaceName, err := FindActiveInterface()
	if err != nil {
		Error.Println(err.Error())
		os.Exit(1)
	}
	activePcapManager = NewPcapManager(ifaceName, PcapDefaultSnapLen, PcapDefaultPromisc, PcapDefaultTimeout)
	Info.Println(activePcapManager.interfaceName)
	activePcapManager.StartMonitor()
}
