package feefi // nolint:dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
)

func (it BaseDepositsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"pool":    it.pool,
				"poolcid": it.poolcid,
				"amounts": it.amounts,
			}),
	)
}

type BaseDepositsItemBSONUnpacker struct {
	RC base.AddressDecoder `bson:"pool"`
	PI string              `bson:"poolcid"`
	AM bson.Raw            `bson:"amounts"`
}

func (it *BaseDepositsItem) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit BaseDepositsItemBSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return err
	}

	return it.unpack(enc, uit.RC, uit.PI, uit.AM)
}
