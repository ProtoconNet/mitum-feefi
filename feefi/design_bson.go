package feefi

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (dg Design) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(dg.Hint()),
		bson.M{
			"fee":     dg.fee,
			"address": dg.address,
		}),
	)
}

type DesignBSONUnpacker struct {
	HT hint.Hint           `bson:"_hint"`
	FE bson.Raw            `bson:"fee"`
	AD base.AddressDecoder `bson:"address"`
}

func (dg *Design) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uds DesignBSONUnpacker
	if err := enc.Unmarshal(b, &uds); err != nil {
		return err
	}

	return dg.unpack(enc, uds.HT, uds.FE, uds.AD)
}
