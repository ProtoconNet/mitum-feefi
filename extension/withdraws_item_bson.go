package extension // nolint:dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
)

func (it BaseWithdrawsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"target":  it.target,
				"amounts": it.amounts,
			}),
	)
}

type BaseWithdrawsItemBSONUnpacker struct {
	TG base.AddressDecoder `bson:"target"`
	AM bson.Raw            `bson:"amounts"`
}

func (it *BaseWithdrawsItem) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit BaseWithdrawsItemBSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return err
	}

	return it.unpack(enc, uit.TG, uit.AM)
}
