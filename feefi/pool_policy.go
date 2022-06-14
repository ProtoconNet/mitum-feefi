package feefi

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/isvalid"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	PoolPolicyType   = hint.Type("mitum-feefi-pool-policy")
	PoolPolicyHint   = hint.NewHint(PoolPolicyType, "v0.0.1")
	PoolPolicyHinter = PoolPolicy{BaseHinter: hint.NewBaseHinter(PoolPolicyHint)}
)

// PoolPolicy is the policy for FeeFiPool
// field fee is incoming fee amount
type PoolPolicy struct {
	hint.BaseHinter
	fee currency.Amount
}

func NewPoolPolicy(fee currency.Amount) PoolPolicy {
	feefiPoolPolicy := PoolPolicy{
		BaseHinter: hint.NewBaseHinter(PoolPolicyHint),
		fee:        fee,
	}
	return feefiPoolPolicy
}

func MustNewPoolPolicy(fee currency.Amount) PoolPolicy {
	design := NewPoolPolicy(fee)
	if err := design.IsValid(nil); err != nil {
		panic(err)
	}
	return design
}

func (dg PoolPolicy) Fee() currency.Amount {
	return dg.fee
}

func (dg PoolPolicy) Bytes() []byte {
	return util.ConcatBytesSlice(dg.fee.Bytes())
}

func (dg PoolPolicy) Hash() valuehash.Hash {
	return dg.GenerateHash()
}

func (dg PoolPolicy) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(dg.Bytes())
}

func (dg PoolPolicy) IsValid([]byte) error { // nolint:revive
	err := isvalid.Check(nil, false, dg.fee)
	if err != nil {
		return err
	}
	return nil
}

func (dg PoolPolicy) Equal(b PoolPolicy) bool {
	return dg.fee.Equal(b.fee)
}
