package feefi

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type DepositsItemJSONPacker struct {
	jsonenc.HintedHead
	RC base.Address        `json:"pool"`
	PI currency.CurrencyID `json:"poolcid"`
	AM []currency.Amount   `json:"amounts"`
}

func (it BaseDepositsItem) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(DepositsItemJSONPacker{
		HintedHead: jsonenc.NewHintedHead(it.Hint()),
		RC:         it.pool,
		PI:         it.poolcid,
		AM:         it.amounts,
	})
}

type BaseDepositsItemJSONUnpacker struct {
	RC base.AddressDecoder `json:"pool"`
	PI string              `json:"poolcid"`
	AM json.RawMessage     `json:"amounts"`
}

func (it *BaseDepositsItem) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var uit BaseDepositsItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return err
	}

	return it.unpack(enc, uit.RC, uit.PI, uit.AM)
}
