package feefi // nolint: dupl

import (
	"encoding/json"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/valuehash"
)

type PoolPolicyUpdaterFactJSONPacker struct {
	jsonenc.HintedHead
	H  valuehash.Hash               `json:"hash"`
	TK []byte                       `json:"token"`
	SD base.Address                 `json:"sender"`
	TG base.Address                 `json:"target"`
	FE currency.Amount              `json:"fee"`
	PI extensioncurrency.ContractID `json:"poolid"`
	CR currency.CurrencyID          `json:"currency"`
}

func (fact PoolPolicyUpdaterFact) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolPolicyUpdaterFactJSONPacker{
		HintedHead: jsonenc.NewHintedHead(fact.Hint()),
		H:          fact.h,
		TK:         fact.token,
		SD:         fact.sender,
		TG:         fact.target,
		FE:         fact.fee,
		PI:         fact.poolID,
		CR:         fact.currency,
	})
}

type PoolPolicyUpdaterFactJSONUnpacker struct {
	H  valuehash.Bytes     `json:"hash"`
	TK []byte              `json:"token"`
	SD base.AddressDecoder `json:"sender"`
	TG base.AddressDecoder `json:"target"`
	FE json.RawMessage     `json:"fee"`
	PI string              `json:"poolid"`
	CR string              `json:"currency"`
}

func (fact *PoolPolicyUpdaterFact) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufact PoolPolicyUpdaterFactJSONUnpacker
	if err := enc.Unmarshal(b, &ufact); err != nil {
		return err
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, ufact.TG, ufact.FE, ufact.PI, ufact.CR)
}

func (op *PoolPolicyUpdater) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo currency.BaseOperation
	if err := ubo.UnpackJSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
