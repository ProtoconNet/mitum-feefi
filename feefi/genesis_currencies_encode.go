package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base/key"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact *GenesisCurrenciesFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	tk []byte,
	genesisNodeKey key.PublickeyDecoder,
	bks []byte,
	bcs []byte,
) error {
	gkey, err := genesisNodeKey.Encode(enc)
	if err != nil {
		return err
	}

	var keys currency.AccountKeys
	hinter, err := enc.Decode(bks)
	if err != nil {
		return err
	} else if k, ok := hinter.(currency.AccountKeys); !ok {
		return errors.Errorf("not Keys: %T", hinter)
	} else {
		keys = k
	}

	fact.h = h
	fact.token = tk
	fact.genesisNodeKey = gkey
	fact.keys = keys

	hcs, err := enc.DecodeSlice(bcs)
	if err != nil {
		return err
	}

	fact.cs = make([]extensioncurrency.CurrencyDesign, len(hcs))
	for i := range hcs {
		j, ok := hcs[i].(extensioncurrency.CurrencyDesign)
		if !ok {
			return util.WrongTypeError.Errorf("expected CurrencyDesign, not %T", hcs[i])
		}

		fact.cs[i] = j
	}

	return nil
}
