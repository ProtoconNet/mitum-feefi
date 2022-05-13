package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	DesignType   = hint.Type("mitum-feefi-design")
	DesignHint   = hint.NewHint(DesignType, "v0.0.1")
	DesignHinter = Design{BaseHinter: hint.NewBaseHinter(DesignHint)}
)

// Design is the design for FeeFiPool
// field fee is incoming fee amount
type Design struct {
	hint.BaseHinter
	fee     currency.Amount
	address base.Address
}

func NewDesign(fee currency.Amount, address base.Address) Design {
	feefiPoolDesign := Design{
		BaseHinter: hint.NewBaseHinter(DesignHint),
		fee:        fee,
		address:    address,
	}
	return feefiPoolDesign
}

func MustNewDesign(fee currency.Amount, address base.Address) Design {
	design := NewDesign(fee, address)
	if err := design.IsValid(nil); err != nil {
		panic(err)
	}
	return design
}

func (dg Design) Fee() currency.Amount {
	return dg.fee
}

func (dg Design) Address() base.Address {
	return dg.address
}

func (dg Design) Bytes() []byte {
	return util.ConcatBytesSlice(dg.fee.Bytes(), dg.address.Bytes())
}

func (dg Design) Hash() valuehash.Hash {
	return dg.GenerateHash()
}

func (dg Design) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(dg.fee.Bytes())
}

func (dg Design) IsValid([]byte) error { // nolint:revive
	err := isvalid.Check(nil, false, dg.fee)
	if err != nil {
		return err
	}
	return nil
}

func (dg Design) Equal(b Design) bool {
	if !dg.address.Equal(b.address) {
		return false
	}
	return dg.fee.Equal(b.fee)
}
