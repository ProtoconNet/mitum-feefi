package digest

import (
	"net/http"
	"strings"
	"time"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/util"
)

func (hd *Handlers) handleFeefi(w http.ResponseWriter, r *http.Request) {
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	var fid string
	s, found := mux.Vars(r)["feefipoolid"]
	if !found {
		HTTP2ProblemWithError(w, errors.Errorf("empty feefipool id"), http.StatusNotFound)

		return
	}

	s = strings.TrimSpace(s)
	if len(s) < 1 {
		HTTP2ProblemWithError(w, errors.Errorf("empty feefipool id"), http.StatusBadRequest)

		return
	}
	fid = s

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return hd.handleFeefiInGroup(fid)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cachekey, time.Second*3)
		}
	}
}

func (hd *Handlers) handleFeefiInGroup(fid string) (interface{}, error) {
	switch va, found, err := hd.database.FeefiPool(fid); {
	case err != nil:
		return nil, err
	case !found:
		return nil, util.NotFoundError
	default:
		hal, err := hd.buildFeefiPoolHal(va)
		if err != nil {
			return nil, err
		}

		return hd.enc.Marshal(hal)
	}
}

func (hd *Handlers) buildFeefiPoolHal(va FeefiPoolValue) (Hal, error) {
	hinted := extensioncurrency.ContractID(va.PrevIncomeBalance().Currency().String()).String()
	h, err := hd.combineURL(HandlerPathFeefi, "feefipoolid", hinted)
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(va, NewHalLink(h, nil))
	/*
		hal = hal.AddLink("currency:{currencyid}", NewHalLink(HandlerPathCurrency, nil).SetTemplated())
		h, err = hd.combineURL(HandlerPathAccountOperations, "address", hinted)
		if err != nil {
			return nil, err
		}
		hal = hal.
			AddLink("operations", NewHalLink(h, nil)).
			AddLink("operations:{offset}", NewHalLink(h+"?offset={offset}", nil).SetTemplated()).
			AddLink("operations:{offset,reverse}", NewHalLink(h+"?offset={offset}&reverse=1", nil).SetTemplated())

		h, err = hd.combineURL(HandlerPathBlockByHeight, "height", va.Height().String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("block", NewHalLink(h, nil))

		if va.PreviousHeight() > base.PreGenesisHeight {
			h, err = hd.combineURL(HandlerPathBlockByHeight, "height", va.PreviousHeight().String())
			if err != nil {
				return nil, err
			}
			hal = hal.AddLink("previous_block", NewHalLink(h, nil))
		}
	*/
	return hal, nil
}

func (hd *Handlers) buildFeefiPoolUserBalanceHal(va feefi.PoolUserBalance) (Hal, error) {
	hinted := va.Income().ID().String()

	h, err := hd.combineURL(HandlerPathFeefi, "feefipoolid", hinted)
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(va, NewHalLink(h, nil))
	/*
		hal = hal.AddLink("currency:{currencyid}", NewHalLink(HandlerPathCurrency, nil).SetTemplated())
		h, err = hd.combineURL(HandlerPathAccountOperations, "address", hinted)
		if err != nil {
			return nil, err
		}
		hal = hal.
			AddLink("operations", NewHalLink(h, nil)).
			AddLink("operations:{offset}", NewHalLink(h+"?offset={offset}", nil).SetTemplated()).
			AddLink("operations:{offset,reverse}", NewHalLink(h+"?offset={offset}&reverse=1", nil).SetTemplated())

		h, err = hd.combineURL(HandlerPathBlockByHeight, "height", va.Height().String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("block", NewHalLink(h, nil))

		if va.PreviousHeight() > base.PreGenesisHeight {
			h, err = hd.combineURL(HandlerPathBlockByHeight, "height", va.PreviousHeight().String())
			if err != nil {
				return nil, err
			}
			hal = hal.AddLink("previous_block", NewHalLink(h, nil))
		}
	*/
	return hal, nil
}
