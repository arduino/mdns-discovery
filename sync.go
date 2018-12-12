//
// This file is part of serial-discovery.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/oleksandr/bonjour"
)

type syncOutputJSON struct {
	EventType string         `json:"eventType"`
	Port      *boardPortJSON `json:"port"`
}

func outputSyncMessage(message *syncOutputJSON) {
	d, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		outputError(err)
	} else {
		syncronizedPrintLn(string(d))
	}
}

func startSync() (chan<- bool, error) {

	closeChan := make(chan bool)

	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return nil, err
	}

	results := make(chan *bonjour.ServiceEntry)

	go func(results chan *bonjour.ServiceEntry, exitCh chan<- bool) {
		for {
			for e := range results {
				log.Printf("%+v", e)
				if e.AddrIPv4 != nil {
					fmt.Println(e)
				}
			}
		}
	}(results, resolver.Exit)

	// Sample output
	//2018/12/12 18:05:14 &{ServiceRecord:{Instance:Arduino Service:_arduino._tcp Domain:local serviceName: serviceInstanceName: serviceTypeName:} HostName:Arduino.local. Port:65280 Text:[ssh_upload=no tcp_check=no auth_upload=yes board=uno2018] TTL:120 AddrIPv4:10.130.22.247 AddrIPv6:<nil>}
	//&{{Arduino _arduino._tcp local   } Arduino.local. 65280 [ssh_upload=no tcp_check=no auth_upload=yes board=uno2018] 120 10.130.22.247 <nil>}

	err = resolver.Browse("_arduino._tcp", "", results)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return nil, err
	}

	return closeChan, nil
}
