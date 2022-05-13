package feefi // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact PoolRegisterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"hash":       fact.h,
				"token":      fact.token,
				"sender":     fact.sender,
				"target":     fact.target,
				"initialfee": fact.initialFee,
				"incomecid":  fact.incomeCID,
				"outlaycid":  fact.outlayCID,
				"currency":   fact.currency,
			}))
}

type PoolRegisterFactBSONUnpacker struct {
	H  valuehash.Bytes     `bson:"hash"`
	TK []byte              `bson:"token"`
	SD base.AddressDecoder `bson:"sender"`
	TG base.AddressDecoder `bson:"target"`
	FE bson.Raw            `bson:"initialfee"`
	FI string              `bson:"incomecid"`
	FO string              `bson:"outlaycid"`
	CR string              `bson:"currency"`
}

func (fact *PoolRegisterFact) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var ufact PoolRegisterFactBSONUnpacker
	if err := bson.Unmarshal(b, &ufact); err != nil {
		return err
	}

	return fact.unpack(enc, ufact.H, ufact.TK, ufact.SD, ufact.TG, ufact.FE, ufact.FI, ufact.FO, ufact.CR)
}

func (op *PoolRegister) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo currency.BaseOperation
	if err := ubo.UnpackBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
