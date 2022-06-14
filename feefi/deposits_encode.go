package feefi

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact *DepositFact) unpack(
	enc encoder.Encoder,
	h valuehash.Hash,
	token []byte,
	bSender base.AddressDecoder,
	bpool base.AddressDecoder,
	spi string,
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
	fact.poolID = extensioncurrency.ContractID(spi)
	fact.amount = am

	return nil
}
