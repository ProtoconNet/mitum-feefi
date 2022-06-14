package feefi

import (
	"encoding/json"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type PoolJSONPacker struct {
	jsonenc.HintedHead
	US map[string]PoolUserBalance `json:"users"`
	PI currency.Amount            `json:"previncomebalance"`
	PO currency.Amount            `json:"prevoutlaybalance"`
}

func (pl Pool) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolJSONPacker{
		HintedHead: jsonenc.NewHintedHead(pl.Hint()),
		US:         pl.users,
		PI:         pl.prevIncomeAmount,
		PO:         pl.prevOutlayAmount,
	})
}

type PoolJSONUnpacker struct {
	HT hint.Hint       `json:"_hint"`
	US json.RawMessage `json:"users"`
	PI json.RawMessage `json:"previncomebalance"`
	PO json.RawMessage `json:"prevoutlaybalance"`
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
	PI extensioncurrency.AmountValue `json:"incomeamount"`
	PO extensioncurrency.AmountValue `json:"outlayamount"`
}

func (pl PoolUserBalance) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolUserBalanceJSONPacker{
		HintedHead: jsonenc.NewHintedHead(pl.Hint()),
		PI:         pl.income,
		PO:         pl.outlay,
	})
}

type PoolUserBalanceJSONUnpacker struct {
	HT hint.Hint       `json:"_hint"`
	PI json.RawMessage `json:"incomeamount"`
	PO json.RawMessage `json:"outlayamount"`
}

func (pl PoolUserBalance) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var upl PoolUserBalanceJSONUnpacker
	if err := enc.Unmarshal(b, &upl); err != nil {
		return err
	}

	return pl.unpack(enc, upl.HT, upl.PI, upl.PO)
}
