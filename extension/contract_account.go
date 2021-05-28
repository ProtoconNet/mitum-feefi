package extension // nolint: dupl, revive

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/state"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	ContractAccountStatusType   = hint.Type("mitum-currency-contract-account-status")
	ContractAccountStatusHint   = hint.NewHint(ContractAccountStatusType, "v0.0.1")
	ContractAccountStatusHinter = ContractAccountStatus{BaseHinter: hint.NewBaseHinter(ContractAccountStatusHint)}
)

type ContractAccountStatus struct {
	hint.BaseHinter
	owner    base.Address
	isActive bool
}

func NewContractAccountStatus(owner base.Address, isActive bool) ContractAccountStatus {
	us := ContractAccountStatus{
		BaseHinter: hint.NewBaseHinter(ContractAccountStatusHint),
		owner:      owner,
		isActive:   isActive,
	}
	return us
}

func (cs ContractAccountStatus) Bytes() []byte {
	var v int8
	if cs.isActive {
		v = 1
	}

	return util.ConcatBytesSlice(cs.owner.Bytes(), []byte{byte(v)})
}

func (cs ContractAccountStatus) Hash() valuehash.Hash {
	return cs.GenerateHash()
}

func (cs ContractAccountStatus) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(cs.Bytes())
}

func (cs ContractAccountStatus) IsValid([]byte) error { // nolint:revive
	return nil
}

func (cs ContractAccountStatus) Owner() base.Address { // nolint:revive
	return cs.owner
}

func (cs ContractAccountStatus) SetOwner(a base.Address) (ContractAccountStatus, error) { // nolint:revive
	err := a.IsValid(nil)
	if err != nil {
		return ContractAccountStatus{}, err
	}

	cs.owner = a

	return cs, nil
}

func (cs ContractAccountStatus) IsActive() bool { // nolint:revive
	return cs.isActive
}

func (cs ContractAccountStatus) SetIsActive(b bool) ContractAccountStatus { // nolint:revive
	cs.isActive = b
	return cs
}

func (cs ContractAccountStatus) Equal(b ContractAccountStatus) bool {
	if cs.isActive != b.isActive {
		return false
	}
	if !cs.owner.Equal(b.owner) {
		return false
	}

	return true
}

type Config interface {
	ID() string            // id of state set in contract model
	ConfigType() hint.Type // config type in contract model
	Hint() hint.Hint
	Bytes() []byte
	Hash() valuehash.Hash
	GenerateHash() valuehash.Hash
	IsValid([]byte) error
	Address() base.Address // contract account address
	SetStateValue(st state.State) (state.State, error)
}

var (
	BaseConfigDataType   = hint.Type("mitum-currency-contract-account-configdata")
	BaseConfigDataHint   = hint.NewHint(BaseConfigDataType, "v0.0.1")
	BaseConfigDataHinter = BaseConfigData{BaseHinter: hint.NewBaseHinter(BaseConfigDataHint)}
)

type BaseConfigData struct {
	hint.BaseHinter
	config Config
}

func NewBaseConfigData(cfg Config) (BaseConfigData, error) {
	err := cfg.IsValid(nil)
	if err != nil {
		return BaseConfigData{}, err
	}
	bcfg := BaseConfigData{
		BaseHinter: hint.NewBaseHinter(BaseConfigDataHint),
		config:     cfg,
	}
	return bcfg, nil
}

func (cfd BaseConfigData) Config() Config {
	return cfd.config
}

func (cfd BaseConfigData) SetConfig(cfg Config) (BaseConfigData, error) {
	err := cfg.IsValid(nil)
	if err != nil {
		return BaseConfigData{}, err
	}

	cfd.config = cfg
	return cfd, nil
}

func (cfd BaseConfigData) Hash() valuehash.Hash {
	return cfd.GenerateHash()
}

func (cfd BaseConfigData) GenerateHash() valuehash.Hash {
	return valuehash.NewSHA256(cfd.Bytes())
}

func (cfd BaseConfigData) Bytes() []byte {
	return cfd.config.Bytes()
}

func (cfd BaseConfigData) IsValid([]byte) error {
	return cfd.config.IsValid(nil)
}

func (cfd BaseConfigData) Equal(b BaseConfigData) bool {
	return cfd.Equal(b)
}
