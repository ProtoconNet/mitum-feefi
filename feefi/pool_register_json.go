package feefi // nolint: dupl

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/valuehash"
)

type PoolRegisterFactJSONPacker struct {
	jsonenc.HintedHead
	H  valuehash.Hash      `json:"hash"`
	TK []byte              `json:"token"`
	SD base.Address        `json:"sender"`
	TG base.Address        `json:"target"`
	FE currency.Amount     `json:"initialfee"`
	FI currency.CurrencyID `json:"incomingcid"`
	FO currency.CurrencyID `json:"outgoingcid"`
	CR currency.CurrencyID `json:"currency"`
}

func (fact PoolRegisterFact) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(PoolRegisterFactJSONPacker{
		HintedHead: jsonenc.NewHintedHead(fact.Hint()),
		H:          fact.h,
		TK:         fact.token,
		SD:         fact.sender,
		TG:         fact.target,
		FE:         fact.initialFee,
		FI:         fact.incomeCID,
		FO:         fact.outlayCID,
		CR:         fact.currency,
	})
}

type PoolRegisterFactJSONUnpacker struct {
	H  valuehash.Bytes     `json:"hash"`
	TK []byte              `json:"token"`
	SD base.AddressDecoder `json:"sender"`
	TG base.AddressDecoder `json:"target"`
	FE json.RawMessage     `json:"initialfee"`
	FI string              `json:"incomingcid"`
	FO string              `json:"outgoingcid"`
	CR string              `json:"currency"`
}

func (fact *PoolRegisterFact) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufact PoolRegisterFactJSONUnpacker
	if err := enc.Unmarshal(b, &ufact); err != nil {
		return err
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, ufact.TG, ufact.FE, ufact.FI, ufact.FO, ufact.CR)
}

func (op *PoolRegister) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo currency.BaseOperation
	if err := ubo.UnpackJSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
