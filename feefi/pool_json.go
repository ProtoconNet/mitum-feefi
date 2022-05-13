package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

type PoolJSONPacker struct {
	jsonenc.HintedHead
	US map[string]PoolUserBalance `json:"users"`
	PI currency.Amount            `json:"previncomefeebalance"`
	PO currency.Amount            `json:"prevoutlayfeebalance"`
}

func (pl Pool) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolJSONPacker{
		HintedHead: jsonenc.NewHintedHead(pl.Hint()),
		US:         pl.users,
		PI:         pl.prevIncomeBalance,
		PO:         pl.prevOutlayBalance,
	})
}

type PoolJSONUnpacker struct {
	HT hint.Hint `json:"_hint"`
	US bson.Raw  `json:"users"`
	PI bson.Raw  `json:"previncomefeebalance"`
	PO bson.Raw  `json:"prevoutlayfeebalance"`
}

func (pl Pool) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var upl PoolJSONUnpacker
	if err := enc.Unmarshal(b, &upl); err != nil {
		return err
	}

	return pl.unpack(enc, upl.HT, upl.US, upl.PI, upl.PO)
}

type PoolUserBalanceJSONPacker struct {
	jsonenc.HintedHead
	PI currency.Amount `json:"incomeamount"`
	PO currency.Amount `json:"outlayamount"`
}

func (pl PoolUserBalance) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolUserBalanceJSONPacker{
		HintedHead: jsonenc.NewHintedHead(pl.Hint()),
		PI:         pl.incomeAmount,
		PO:         pl.outlayAmount,
	})
}

type PoolUserBalanceJSONUnpacker struct {
	HT hint.Hint `json:"_hint"`
	PI bson.Raw  `json:"incomeamount"`
	PO bson.Raw  `json:"outlayamount"`
}

func (pl PoolUserBalance) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var upl PoolJSONUnpacker
	if err := enc.Unmarshal(b, &upl); err != nil {
		return err
	}

	return pl.unpack(enc, upl.HT, upl.PI, upl.PO)
}
