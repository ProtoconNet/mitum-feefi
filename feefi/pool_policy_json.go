package feefi

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type PoolPolicyJSONPacker struct {
	jsonenc.HintedHead
	FE currency.Amount `json:"fee"`
}

func (dg PoolPolicy) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolPolicyJSONPacker{
		HintedHead: jsonenc.NewHintedHead(dg.Hint()),
		FE:         dg.fee,
	})
}

type PoolPolicyJSONUnpacker struct {
	HT hint.Hint       `json:"_hint"`
	FE json.RawMessage `json:"fee"`
}

func (dg *PoolPolicy) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uds PoolPolicyJSONUnpacker
	if err := enc.Unmarshal(b, &uds); err != nil {
		return err
	}

	return dg.unpack(enc, uds.HT, uds.FE)
}
