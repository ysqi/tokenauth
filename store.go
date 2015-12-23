// Copyright 2016 Author YuShuangqi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tokenauth

import (
	"fmt"
	"runtime"
	"time"
)

//Token store interface.
type TokenStore interface {

	// Init store
	// Returns error if init fail.
	Open(config string) error

	// Close store
	Close() error

	// Save audience into store.
	// Returns error if error occured during execution.
	SaveAudience(audience *Audience) error

	// Delete audience and  all tokens of audience.
	DeleteAudience(clientID string) error

	// Get audience info or returns error.
	GetAudience(clientID string) (*Audience, error)

	// Save token to token.
	// Returns error if save token fail.
	SaveToken(token *Token) error

	// Delete token info from store.
	// Returns error if error occured during execution
	DeleteToken(tokenString string) error

	// Get token info from store.
	// Returns nil if not found token.
	// Returns error if get token fail.
	GetToken(tokenString string) (*Token, error)

	DeleteExpired()
}

// Janitor contains TokenStore and Janitor instance.
type janitorTaget struct {
	store   TokenStore
	janitor *janitor
}

//Janitor implement method to period exec store method.
type janitor struct {
	Interval time.Duration
	stop     chan bool
}

// Period Run
func (j *janitor) Run(taget *janitorTaget) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			taget.store.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopJanitor(taget *janitorTaget) {
	taget.janitor.stop <- true
}

func runJanitor(taget *janitorTaget, ci time.Duration) {
	j := &janitor{
		Interval: ci,
	}
	taget.janitor = j
	go j.Run(taget)
}

var adapters = make(map[string]TokenStore)

// Resister one store provider.
// If name is empty,will panic.
// If same name has registerd ,will panic.
func RegStore(name string, adapter TokenStore) {

	if adapter == nil {
		panic("tokenStore: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("tokenStore: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

// New regiesterd store
func NewStore(adapterName, config string) (TokenStore, error) {

	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("tokenStore: unknown adapter name %q (forgot registration ?)", adapterName)
	}
	if err := adapter.Open(config); err != nil {
		return nil, err
	} else {
		s := &janitorTaget{store: adapter}
		runJanitor(s, time.Minute*5)
		runtime.SetFinalizer(s, stopJanitor)
		return adapter, nil
	}

}
