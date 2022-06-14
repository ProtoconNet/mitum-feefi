package feefi

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (dg PoolDesign) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(dg.Hint()),
		bson.M{
			"policy":  dg.policy,
			"address": dg.address,
		}),
	)
}

type PoolDesignBSONUnpacker struct {
	HT hint.Hint           `bson:"_hint"`
	PO bson.Raw            `bson:"policy"`
	AD base.AddressDecoder `bson:"address"`
}

func (dg *PoolDesign) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uds PoolDesignBSONUnpacker
	if err := enc.Unmarshal(b, &uds); err != nil {
		return err
	}

	return dg.unpack(enc, uds.HT, uds.PO, uds.AD)
}
