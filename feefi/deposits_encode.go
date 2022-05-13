package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (it *BaseDepositsItem) unpack(
	enc encoder.Encoder,
	bReceiver base.AddressDecoder,
	spi string,
	bam []byte,
) error {
	a, err := bReceiver.Encode(enc)
	if err != nil {
		return err
	}
	it.pool = a

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return err
	}

	it.poolcid = currency.CurrencyID(spi)

	am := make([]currency.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(currency.Amount)
		if !ok {
			return util.WrongTypeError.Errorf("expected Amount, not %T", ham[i])
		}

		am[i] = j
	}

	it.amounts = am

	return nil
}

func (fact *DepositFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	token []byte,
	bSender base.AddressDecoder,
	bpool base.AddressDecoder,
	sci string,
	bam []byte,
) error {
	sender, err := bSender.Encode(enc)
	if err != nil {
		return err
	}

	pool, err := bpool.Encode(enc)
	if err != nil {
		return err
	}

	a, err := enc.Decode(bam)
	if err != nil {
		return err
	}
	am, ok := a.(currency.Amount)
	if !ok {
		return util.WrongTypeError.Errorf("expected Amount, not %T", a)
	}

	fact.h = h
	fact.token = token
	fact.sender = sender
	fact.pool = pool
	fact.poolCID = currency.CurrencyID(sci)
	fact.amount = am

	return nil
}
