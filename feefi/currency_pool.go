package feefi

import (
	"sync"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/state"
)

type CurrencyPool struct {
	sync.RWMutex
	demap  map[currency.CurrencyID]extensioncurrency.CurrencyDesign
	stsmap map[currency.CurrencyID]state.State
	cids   []currency.CurrencyID
}

func NewCurrencyPool() *CurrencyPool {
	return &CurrencyPool{
		demap:  map[currency.CurrencyID]extensioncurrency.CurrencyDesign{},
		stsmap: map[currency.CurrencyID]state.State{},
	}
}

func (cp *CurrencyPool) Clear() {
	cp.Lock()
	defer cp.Unlock()

	cp.demap = nil
	cp.stsmap = nil
	cp.cids = nil
}

func (cp *CurrencyPool) Set(st state.State) error {
	cp.Lock()
	defer cp.Unlock()

	de, err := StateCurrencyDesignValue(st)
	if err != nil {
		return err
	}

	cp.demap[de.Currency()] = de
	cp.stsmap[de.Currency()] = st
	cp.cids = append(cp.cids, de.Currency())

	return nil
}

func (cp *CurrencyPool) CIDs() []currency.CurrencyID {
	cp.RLock()
	defer cp.RUnlock()

	return cp.cids
}

func (cp *CurrencyPool) Designs() map[currency.CurrencyID]extensioncurrency.CurrencyDesign {
	cp.RLock()
	defer cp.RUnlock()

	m := map[currency.CurrencyID]extensioncurrency.CurrencyDesign{}
	for k := range cp.demap {
		m[k] = cp.demap[k]
	}

	return m
}

func (cp *CurrencyPool) States() map[currency.CurrencyID]state.State {
	cp.RLock()
	defer cp.RUnlock()

	m := map[currency.CurrencyID]state.State{}
	for k := range cp.stsmap {
		m[k] = cp.stsmap[k]
	}

	return m
}

func (cp *CurrencyPool) Policy(cid currency.CurrencyID) (extensioncurrency.CurrencyPolicy, bool) {
	i, found := cp.Get(cid)
	if !found {
		return extensioncurrency.CurrencyPolicy{}, false
	}
	return i.Policy(), true
}

func (cp *CurrencyPool) Feeer(cid currency.CurrencyID) (extensioncurrency.Feeer, bool) {
	i, found := cp.Get(cid)
	if !found {
		return nil, false
	}
	return i.Policy().Feeer(), true
}

func (cp *CurrencyPool) State(cid currency.CurrencyID) (state.State, bool) {
	i, found := cp.stsmap[cid]
	return i, found
}

func (cp *CurrencyPool) TraverseDesign(callback func(cid currency.CurrencyID, de extensioncurrency.CurrencyDesign) bool) {
	cp.RLock()
	defer cp.RUnlock()

	for k := range cp.demap {
		if !callback(k, cp.demap[k]) {
			break
		}
	}
}

func (cp *CurrencyPool) TraverseState(callback func(cid currency.CurrencyID, de state.State) bool) {
	cp.RLock()
	defer cp.RUnlock()

	for k := range cp.stsmap {
		if !callback(k, cp.stsmap[k]) {
			break
		}
	}
}

func (cp *CurrencyPool) Exists(cid currency.CurrencyID) bool {
	cp.RLock()
	defer cp.RUnlock()

	_, found := cp.demap[cid]

	return found
}

func (cp *CurrencyPool) Get(cid currency.CurrencyID) (extensioncurrency.CurrencyDesign, bool) {
	cp.RLock()
	defer cp.RUnlock()

	i, found := cp.demap[cid]
	return i, found
}
