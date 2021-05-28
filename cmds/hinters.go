package cmds

import (
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util/hint"

	"github.com/protoconNet/mitum-account-extension/digest"
	"github.com/protoconNet/mitum-account-extension/extension"
	"github.com/spikeekips/mitum-currency/currency"
)

var (
	Hinters []hint.Hinter
	Types   []hint.Type
)

var types = []hint.Type{
	currency.AccountType,
	currency.AddressType,
	currency.AmountType,
	currency.CreateAccountsFactType,
	currency.CreateAccountsItemMultiAmountsType,
	currency.CreateAccountsItemSingleAmountType,
	currency.CreateAccountsType,
	currency.CurrencyDesignType,
	currency.CurrencyPolicyType,
	currency.CurrencyPolicyUpdaterFactType,
	currency.CurrencyPolicyUpdaterType,
	currency.CurrencyRegisterFactType,
	currency.CurrencyRegisterType,
	currency.FeeOperationFactType,
	currency.FeeOperationType,
	currency.FixedFeeerType,
	currency.GenesisCurrenciesFactType,
	currency.GenesisCurrenciesType,
	currency.AccountKeyType,
	currency.KeyUpdaterFactType,
	currency.KeyUpdaterType,
	currency.AccountKeysType,
	currency.NilFeeerType,
	currency.RatioFeeerType,
	currency.SuffrageInflationFactType,
	currency.SuffrageInflationType,
	currency.TransfersFactType,
	currency.TransfersItemMultiAmountsType,
	currency.TransfersItemSingleAmountType,
	currency.TransfersType,
	extension.ContractAccountKeysType,
	extension.ContractAccountStatusType,
	extension.CreateContractAccountsFactType,
	extension.CreateContractAccountsType,
	extension.CreateContractAccountsItemMultiAmountsType,
	extension.CreateContractAccountsItemSingleAmountType,
	extension.DeactivateFactType,
	extension.DeactivateType,
	extension.WithdrawsFactType,
	extension.WithdrawsType,
	extension.WithdrawsItemMultiAmountsType,
	extension.WithdrawsItemSingleAmountType,
	digest.ProblemType,
	digest.NodeInfoType,
	digest.BaseHalType,
	digest.AccountValueType,
	digest.OperationValueType,
}

var hinters = []hint.Hinter{
	currency.AccountHinter,
	currency.AddressHinter,
	currency.AmountHinter,
	currency.CreateAccountsFactHinter,
	currency.CreateAccountsItemMultiAmountsHinter,
	currency.CreateAccountsItemSingleAmountHinter,
	currency.CreateAccountsHinter,
	currency.CurrencyDesignHinter,
	currency.CurrencyPolicyUpdaterFactHinter,
	currency.CurrencyPolicyUpdaterHinter,
	currency.CurrencyPolicyHinter,
	currency.CurrencyRegisterFactHinter,
	currency.CurrencyRegisterHinter,
	currency.FeeOperationFactHinter,
	currency.FeeOperationHinter,
	currency.FixedFeeerHinter,
	currency.GenesisCurrenciesFactHinter,
	currency.GenesisCurrenciesHinter,
	currency.KeyUpdaterFactHinter,
	currency.KeyUpdaterHinter,
	currency.AccountKeysHinter,
	currency.AccountKeyHinter,
	currency.NilFeeerHinter,
	currency.RatioFeeerHinter,
	currency.SuffrageInflationFactHinter,
	currency.SuffrageInflationHinter,
	currency.TransfersFactHinter,
	currency.TransfersItemMultiAmountsHinter,
	currency.TransfersItemSingleAmountHinter,
	currency.TransfersHinter,
	extension.ContractAccountKeysHinter,
	extension.ContractAccountStatusHinter,
	extension.CreateContractAccountsFactHinter,
	extension.CreateContractAccountsHinter,
	extension.CreateContractAccountsItemMultiAmountsHinter,
	extension.CreateContractAccountsItemSingleAmountHinter,
	extension.DeactivateFactHinter,
	extension.DeactivateHinter,
	extension.WithdrawsFactHinter,
	extension.WithdrawsHinter,
	extension.WithdrawsItemMultiAmountsHinter,
	extension.WithdrawsItemSingleAmountHinter,
	digest.AccountValue{},
	digest.BaseHal{},
	digest.NodeInfo{},
	digest.OperationValue{},
	digest.Problem{},
}

func init() {
	Hinters = make([]hint.Hinter, len(launch.EncoderHinters)+len(hinters))
	copy(Hinters, launch.EncoderHinters)
	copy(Hinters[len(launch.EncoderHinters):], hinters)

	Types = make([]hint.Type, len(launch.EncoderTypes)+len(types))
	copy(Types, launch.EncoderTypes)
	copy(Types[len(launch.EncoderTypes):], types)
}
