package main

import (
	"encoding/json"
	"net/http"
	"time"
)

var (
	// Port that the API will listen and serve on
	Port = "1337"
	// Response for the lastest json response on "/usage"
	Response *Usage
)

// Device represents a per-device json response
type Device struct {
	Name    string    `json:"name"`
	History []History `json:"history"`
	Total   int       `json:"total"`
}

// History represents an item in the device history
type History struct {
	Time  int64 `json:"time"`
	Total int   `json:"total"`
}

// Usage represents the core usage json response
type Usage struct {
	Devices map[string]*Device `json:"devices"`
	Total   int                `json:"total"`
}

// StartServer starts the http server for the API
func StartServer() {
	Info.Println("Starting HTTP server...")
	Info.Println("Building response...")
	Response = BuildUsage()
	go UpdateJSON()
	http.HandleFunc("/usage", totalUsage)
	Error.Panicln(http.ListenAndServe(":"+Port, nil))
}

// BuildUsage builds the initial Usage item
func BuildUsage() *Usage {
	usage := &Usage{
		Devices: make(map[string]*Device),
		Total:   activePcapManager.byteTotal,
	}
	for key, val := range activePcapManager.peerList {
		usage.Devices[key] = &Device{
			Name: key,
			History: []History{History{
				Time:  time.Now().Unix(),
				Total: val,
			}},
			Total: val,
		}
	}
	return usage
}

// UpdateJSON using the interval constant
func UpdateJSON() {
	for {
		for key, val := range Response.Devices {
			newHistory := History{
				Time:  time.Now().Unix(),
				Total: activePcapManager.peerList[key],
			}
			val.History = append(val.History, newHistory)
		}
		time.Sleep(30 * time.Second)
	}
}

func totalUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	for key, val := range Response.Devices {
		val.Total = activePcapManager.peerList[key]
	}
	Response.Total = activePcapManager.byteTotal
	json, jsonErr := json.Marshal(Response)
	if jsonErr != nil {
		Error.Println(jsonErr.Error())
		w.WriteHeader(500)
	}
	w.Write(json)
}
