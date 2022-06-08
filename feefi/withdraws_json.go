package feefi

import (
	"encoding/json"

	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/valuehash"
)

type WithdrawsFactJSONPacker struct {
	jsonenc.HintedHead
	H  valuehash.Hash               `json:"hash"`
	TK []byte                       `json:"token"`
	SD base.Address                 `json:"sender"`
	PL base.Address                 `json:"pool"`
	PI extensioncurrency.ContractID `json:"poolid"`
	AM []currency.Amount            `json:"amounts"`
}

func (fact WithdrawsFact) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(WithdrawsFactJSONPacker{
		HintedHead: jsonenc.NewHintedHead(fact.Hint()),
		H:          fact.h,
		TK:         fact.token,
		SD:         fact.sender,
		PL:         fact.pool,
		PI:         fact.poolID,
		AM:         fact.amounts,
	})
}

func (fact *WithdrawsFact) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufact struct {
		H  valuehash.Bytes     `json:"hash"`
		TK []byte              `json:"token"`
		SD base.AddressDecoder `json:"sender"`
		PL base.AddressDecoder `json:"pool"`
		PI string              `json:"poolid"`
		AM json.RawMessage     `json:"amounts"`
	}
	if err := jsonenc.Unmarshal(b, &ufact); err != nil {
		return err
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, ufact.PL, ufact.PI, ufact.AM)
}

func (op *Withdraws) UnpackJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo currency.BaseOperation
	if err := ubo.UnpackJSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
