package extension // nolint: dupl, revive

import (
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (cs ContractAccountStatus) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(cs.Hint()),
		bson.M{
			"isactive": cs.isActive,
			"owner":    cs.owner,
		}),
	)
}

type ContractAccountStatusBSONUnpacker struct {
	IA bool                `bson:"isactive"`
	OW base.AddressDecoder `bson:"owner"`
}

func (cs *ContractAccountStatus) UnpackBSON(b []byte, enc *bsonenc.Encoder) error {
	var cub ContractAccountStatusBSONUnpacker
	if err := bsonenc.Unmarshal(b, &cub); err != nil {
		return err
	}

	return cs.unpack(enc, cub.IA, cub.OW)
}
