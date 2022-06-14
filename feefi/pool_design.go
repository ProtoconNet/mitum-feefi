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
	PoolDesignType   = hint.Type("mitum-feefi-pool-design")
	PoolDesignHint   = hint.NewHint(PoolDesignType, "v0.0.1")
	PoolDesignHinter = PoolDesign{BaseHinter: hint.NewBaseHinter(PoolDesignHint)}
)

// PoolDesign is the design for FeeFiPool
// field fee is incoming fee amount
type PoolDesign struct {
	hint.BaseHinter
	policy  PoolPolicy
	address base.Address
}

func NewPoolDesign(fee currency.Amount, address base.Address) PoolDesign {
	policy := NewPoolPolicy(fee)
	feefiPoolDesign := PoolDesign{
		BaseHinter: hint.NewBaseHinter(PoolDesignHint),
		policy:     policy,
		address:    address,
	}
	return feefiPoolDesign
}

func MustNewPoolDesign(fee currency.Amount, address base.Address) PoolDesign {
	design := NewPoolDesign(fee, address)
	if err := design.IsValid(nil); err != nil {
		panic(err)
	}
	return design
}

func (dg PoolDesign) Policy() PoolPolicy {
	return dg.policy
}

func (dg PoolDesign) Address() base.Address {
	return dg.address
}

func (dg PoolDesign) Bytes() []byte {
	return util.ConcatBytesSlice(dg.policy.Bytes(), dg.address.Bytes())
}

func (dg PoolDesign) Hash() valuehash.Hash {
	return dg.GenerateHash()
}

func (dg PoolDesign) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(dg.Bytes())
}

func (dg PoolDesign) IsValid([]byte) error { // nolint:revive
	err := isvalid.Check(nil, false, dg.policy, dg.address)
	if err != nil {
		return err
	}
	return nil
}

func (dg PoolDesign) Equal(b PoolDesign) bool {
	if !dg.address.Equal(b.address) {
		return false
	}
	return dg.policy.Equal(b.policy)
}
