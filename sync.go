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
	"context"
	"encoding/json"

	"github.com/brutella/dnssd"
)

var service = "_arduino._tcp.local."

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

	addFn := func(srv dnssd.Service) {
		outputSyncMessage(&syncOutputJSON{
			EventType: "add",
			Port:      newBoardPortJSON(&srv),
		})
	}

	remFn := func(srv dnssd.Service) {
		outputSyncMessage(&syncOutputJSON{
			EventType: "remove",
			Port:      newBoardPortJSON(&srv),
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := dnssd.LookupType(ctx, service, addFn, remFn); err != nil {
		return nil, err
	}

	return closeChan, nil
}
