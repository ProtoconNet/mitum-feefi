package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact *PoolPolicyUpdaterFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	token []byte,
	bsender base.AddressDecoder,
	btarget base.AddressDecoder,
	bfe []byte,
	pi string,
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

	err = encoder.Decode(bfe, enc, &fact.fee)
	if err != nil {
		return err
	}

	fact.h = h
	fact.token = token
	fact.sender = sender
	fact.target = target
	fact.poolID = extensioncurrency.ContractID(pi)
	fact.currency = currency.CurrencyID(cr)

	return nil
}
