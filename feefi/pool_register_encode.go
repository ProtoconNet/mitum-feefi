package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact *PoolRegisterFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	token []byte,
	bsender base.AddressDecoder,
	btarget base.AddressDecoder,
	bfe []byte,
	fi string,
	fo string,
	cr string,
) error {
	sender, err := bsender.Encode(enc)
	if err != nil {
		return err
	}

	target, err := btarget.Encode(enc)
	if err != nil {
		return err
	}

	err = encoder.Decode(bfe, enc, &fact.initialFee)
	if err != nil {
		return err
	}

	fact.h = h
	fact.token = token
	fact.sender = sender
	fact.target = target
	fact.incomeCID = currency.CurrencyID(fi)
	fact.outlayCID = currency.CurrencyID(fo)
	fact.currency = currency.CurrencyID(cr)

	return nil
}
