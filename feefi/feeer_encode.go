package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (fa *FeefiFeeer) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	brc base.AddressDecoder,
	am currency.Big,
	ex bool,
	sec string,
	bfr base.AddressDecoder,
	em currency.Big,
) error {
	fa.BaseHinter = hint.NewBaseHinter(ht)

	ra, err := brc.Encode(enc)
	if err != nil {
		return err
	}
	fa.receiver = ra

	fa.amount = am

	fa.exchangeable = ex
	fa.exchangeCID = currency.CurrencyID(sec)

	fr, err := bfr.Encode(enc)
	if err != nil {
		return err
	}

	fa.feefier = fr
	fa.exchangeMinAmount = em

	return nil
}
