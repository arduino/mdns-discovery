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
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	properties "github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"github.com/hashicorp/mdns"
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

const mdnsServiceName = "_arduino._tcp"

// MDNSDiscovery is the implementation of the network pluggable-discovery
type MDNSDiscovery struct {
	cancelFunc  func()
	entriesChan chan *mdns.ServiceEntry
}

// Hello handles the pluggable-discovery HELLO command
func (d *MDNSDiscovery) Hello(userAgent string, protocolVersion int) error {
	// The mdns library used has some logs statement that we must disable
	log.SetOutput(ioutil.Discard)
	return nil
}

// Stop handles the pluggable-discovery STOP command
func (d *MDNSDiscovery) Stop() error {
	if d.cancelFunc != nil {
		d.cancelFunc()
		d.cancelFunc = nil
	}
	if d.entriesChan != nil {
		close(d.entriesChan)
		d.entriesChan = nil
	}
	return nil
}

// Quit handles the pluggable-discovery QUIT command
func (d *MDNSDiscovery) Quit() {
	close(d.entriesChan)
}

// StartSync handles the pluggable-discovery START_SYNC command
func (d *MDNSDiscovery) StartSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) error {
	if d.entriesChan != nil {
		return fmt.Errorf("already syncing")
	}

	d.entriesChan = make(chan *mdns.ServiceEntry, 4)
	var receiver <-chan *mdns.ServiceEntry = d.entriesChan
	var sender chan<- *mdns.ServiceEntry = d.entriesChan

	go func() {
		for entry := range receiver {
			eventCB("add", toDiscoveryPort(entry))
		}
	}()

	params := &mdns.QueryParam{
		Service:             mdnsServiceName,
		Domain:              "local",
		Timeout:             time.Second * 15,
		Entries:             sender,
		WantUnicastResponse: false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			if err := mdns.Query(params); err != nil {
				errorCB("mdns lookup error: " + err.Error())
			}
			select {
			default:
			case <-ctx.Done():
				return
			}
		}
	}()
	d.cancelFunc = cancel
	return nil
}

func toDiscoveryPort(entry *mdns.ServiceEntry) *discovery.Port {
	ip := ""
	if len(entry.AddrV4) > 0 {
		ip = entry.AddrV4.String()
	} else if len(entry.AddrV6) > 0 {
		ip = entry.AddrV6.String()
	}

	props := properties.NewMap()
	props.Set("hostname", entry.Host)
	props.Set("port", strconv.Itoa(entry.Port))

	for _, field := range entry.InfoFields {
		split := strings.Split(field, "=")
		if len(split) != 2 {
			continue
		}
		key, value := split[0], split[1]
		props.Set(key, value)
		if key == "board" {
			// duplicate for backwards compatibility
			props.Set(".", value)
		}
	}

	return &discovery.Port{
		Address:       ip,
		AddressLabel:  fmt.Sprintf("%s at %s", entry.Name, ip),
		Protocol:      "network",
		ProtocolLabel: "Network Port",
		Properties:    props,
	}
}
