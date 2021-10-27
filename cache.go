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
	"sync"
	"time"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
)

// cacheItem stores TTL of discovered ports and its timer to handle TTL.
type cacheItem struct {
	timerTTL   *time.Timer
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// portsCached is a cache to store discovered ports with a TTL.
// Ports that reach their TTL are automatically deleted and
// main discovery process is notified.
// All operations on the internal data are thread safe.
type portsCache struct {
	data             map[string]*cacheItem
	dataMutex        sync.Mutex
	itemsTTL         time.Duration
	deletionCallback func(port *discovery.Port)
}

// newCache creates a new portsCache and returns it.
// itemsTTL is the TTL of a single item, when it's reached
// the stored item is deleted.
func newCache(itemsTTL time.Duration, deletionCallback func(port *discovery.Port)) *portsCache {
	return &portsCache{
		itemsTTL:         itemsTTL,
		data:             make(map[string]*cacheItem),
		deletionCallback: deletionCallback,
	}
}

// storeOrUpdate stores a new port and sets its TTL or
// updates the TTL if already stored.
// Return true if the port TTL has been updated, false otherwise
// storeOrUpdate is thread safe.
func (c *portsCache) storeOrUpdate(port *discovery.Port) bool {
	key := fmt.Sprintf("%s:%s %s", port.Address, port.Properties.Get("port"), port.Properties.Get("board"))
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()
	// We need a cancellable context to avoid leaving
	// goroutines hanging if an item's timer TTL is stopped.
	ctx, cancelFunc := context.WithCancel(context.Background())
	item, ok := c.data[key]
	timerTTL := time.NewTimer(c.itemsTTL)
	if ok {
		item.timerTTL.Stop()
		// If we stop the timer the goroutine that waits
		// for it to go off would just hang forever if we
		// don't cancel the item's context
		item.cancelFunc()
		item.timerTTL = timerTTL
		item.ctx = ctx
		item.cancelFunc = cancelFunc
	} else {
		item = &cacheItem{
			timerTTL:   timerTTL,
			ctx:        ctx,
			cancelFunc: cancelFunc,
		}
		c.data[key] = item
	}

	go func(key string, item *cacheItem, port *discovery.Port) {
		select {
		case <-item.timerTTL.C:
			c.dataMutex.Lock()
			defer c.dataMutex.Unlock()
			c.deletionCallback(port)
			delete(c.data, key)
			return
		case <-item.ctx.Done():
			// The TTL has been renewed, we also stop the timer
			// that keeps track of the TTL.
			// The channel of a stopped timer will never fire so
			// if keep waiting for it in this goroutine it will
			// hang forever.
			// Using a context we handle this gracefully.
			return
		}
	}(key, item, port)

	return ok
}

// clear removes all the stored items and stops their TTL timers.
// clear is thread safe.
func (c *portsCache) clear() {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()
	for key, item := range c.data {
		item.timerTTL.Stop()
		item.cancelFunc()
		delete(c.data, key)
	}
}
