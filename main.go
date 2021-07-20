//
// This file is part of mdns-discovery.
//
// Copyright 2018-2021 ARDUINO SA (http://www.arduino.cc/)
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
	"fmt"
	"os"
	"strconv"

	properties "github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"github.com/brutella/dnssd"
)

func main() {
	parseArgs()
	mdnsDiscovery := &MDNSDiscovery{}
	disc := discovery.NewDiscoveryServer(mdnsDiscovery)
	if err := disc.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

}

const mdnsServiceName = "_arduino._tcp.local."

type MDNSDiscovery struct {
	cancelFunc func()
}

func (d *MDNSDiscovery) Hello(userAgent string, protocolVersion int) error {
	return nil
}

func (d *MDNSDiscovery) Start() error {
	return nil
}

func (d *MDNSDiscovery) Stop() error {
	if d.cancelFunc != nil {
		d.cancelFunc()
		d.cancelFunc = nil
	}
	return nil
}

func (d *MDNSDiscovery) Quit() {
}

func (d *MDNSDiscovery) List() ([]*discovery.Port, error) {
	return []*discovery.Port{}, nil
}

func (d *MDNSDiscovery) StartSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) error {
	addFn := func(srv dnssd.Service) {
		eventCB("add", newBoardPortJSON(&srv))
	}
	remFn := func(srv dnssd.Service) {
		eventCB("remove", newBoardPortJSON(&srv))
	}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := dnssd.LookupType(ctx, mdnsServiceName, addFn, remFn); err != nil {
			errorCB("mdns lookup error: " + err.Error())
		}
	}()
	d.cancelFunc = cancel
	return nil
}

func newBoardPortJSON(port *dnssd.Service) *discovery.Port {
	ip := "127.0.0.1"
	if len(port.IPs) > 0 {
		ip = port.IPs[0].String()
	}

	props := properties.NewMap()
	props.Set("ttl", strconv.Itoa(int(port.TTL.Seconds())))
	props.Set("hostname", port.Hostname())
	props.Set("port", strconv.Itoa(port.Port))
	for key, value := range port.Text {
		props.Set(key, value)
		if key == "board" {
			// duplicate for backwards compatibility
			props.Set(".", value)
		}
	}
	return &discovery.Port{
		Address:       ip,
		AddressLabel:  port.Name + " at " + ip,
		Protocol:      "network",
		ProtocolLabel: "Network Port",
		Properties:    props,
	}
}
