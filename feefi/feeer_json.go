package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type FeefiFeeerJSONPacker struct {
	jsonenc.HintedHead
	RC base.Address        `json:"receiver"`
	AM currency.Big        `json:"amount"`
	EX bool                `json:"exchangeable"`
	EC currency.CurrencyID `json:"exchangecid"`
	FR base.Address        `json:"feefier"`
	EM currency.Big        `json:"exchangeminamount"`
}

func (fa FeefiFeeer) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(FeefiFeeerJSONPacker{
		HintedHead: jsonenc.NewHintedHead(fa.Hint()),
		RC:         fa.receiver,
		AM:         fa.amount,
		EX:         fa.exchangeable,
		EC:         fa.exchangeCID,
		FR:         fa.feefier,
		EM:         fa.exchangeMinAmount,
	})
}

type FeefiFeeerJSONUnpacker struct {
	HT hint.Hint           `json:"_hint"`
	RC base.AddressDecoder `json:"receiver"`
	AM currency.Big        `json:"amount"`
	EX bool                `json:"exchangeable"`
	EC string              `json:"exchangecid"`
	FR base.AddressDecoder `json:"feefier"`
	EM currency.Big        `json:"exchangeminamount"`
}

func (fa *FeefiFeeer) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufa FeefiFeeerJSONUnpacker
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return err
	}

	return fa.unpack(enc, ufa.HT, ufa.RC, ufa.AM, ufa.EX, ufa.EC, ufa.FR, ufa.EM)
}
