package cmds

import (
	extensioncurrency "github.com/ProtoconNet/mitum-currency-extension/currency"
	"github.com/ProtoconNet/mitum-feefi/digest"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util/hint"
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
	currency.AccountKeyType,
	currency.KeyUpdaterFactType,
	currency.KeyUpdaterType,
	currency.AccountKeysType,
	currency.TransfersFactType,
	currency.TransfersItemMultiAmountsType,
	currency.TransfersItemSingleAmountType,
	currency.TransfersType,
	extensioncurrency.AmountValueType,
	extensioncurrency.CurrencyDesignType,
	extensioncurrency.CurrencyPolicyType,
	extensioncurrency.CurrencyPolicyUpdaterFactType,
	extensioncurrency.CurrencyPolicyUpdaterType,
	extensioncurrency.CurrencyRegisterFactType,
	extensioncurrency.CurrencyRegisterType,
	feefi.GenesisCurrenciesFactType,
	feefi.GenesisCurrenciesType,
	extensioncurrency.SuffrageInflationFactType,
	extensioncurrency.SuffrageInflationType,
	extensioncurrency.FixedFeeerType,
	extensioncurrency.NilFeeerType,
	extensioncurrency.RatioFeeerType,
	extensioncurrency.ContractAccountKeysType,
	extensioncurrency.ContractAccountType,
	extensioncurrency.CreateContractAccountsFactType,
	extensioncurrency.CreateContractAccountsType,
	extensioncurrency.CreateContractAccountsItemMultiAmountsType,
	extensioncurrency.CreateContractAccountsItemSingleAmountType,
	extensioncurrency.WithdrawsFactType,
	extensioncurrency.WithdrawsType,
	extensioncurrency.WithdrawsItemMultiAmountsType,
	extensioncurrency.WithdrawsItemSingleAmountType,
	feefi.DepositFactType,
	feefi.DepositType,
	feefi.FeeOperationFactType,
	feefi.FeeOperationType,
	feefi.PoolPolicyUpdaterFactType,
	feefi.PoolPolicyUpdaterType,
	feefi.PoolRegisterFactType,
	feefi.PoolRegisterType,
	feefi.PoolWithdrawsFactType,
	feefi.PoolWithdrawsType,
	feefi.FeefiFeeerType,
	feefi.PoolType,
	feefi.PoolUserBalanceType,
	feefi.PoolDesignType,
	feefi.PoolPolicyType,
	digest.ProblemType,
	digest.NodeInfoType,
	digest.BaseHalType,
	digest.AccountValueType,
	digest.FeefiPoolValueType,
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
	currency.KeyUpdaterFactHinter,
	currency.KeyUpdaterHinter,
	currency.AccountKeysHinter,
	currency.AccountKeyHinter,
	currency.TransfersFactHinter,
	currency.TransfersItemMultiAmountsHinter,
	currency.TransfersItemSingleAmountHinter,
	currency.TransfersHinter,
	extensioncurrency.AmountValueHinter,
	extensioncurrency.CurrencyDesignHinter,
	extensioncurrency.CurrencyPolicyUpdaterFactHinter,
	extensioncurrency.CurrencyPolicyUpdaterHinter,
	extensioncurrency.CurrencyPolicyHinter,
	extensioncurrency.CurrencyRegisterFactHinter,
	extensioncurrency.CurrencyRegisterHinter,
	feefi.GenesisCurrenciesFactHinter,
	feefi.GenesisCurrenciesHinter,
	extensioncurrency.SuffrageInflationFactHinter,
	extensioncurrency.SuffrageInflationHinter,
	extensioncurrency.FixedFeeerHinter,
	extensioncurrency.NilFeeerHinter,
	extensioncurrency.RatioFeeerHinter,
	extensioncurrency.ContractAccountKeysHinter,
	extensioncurrency.ContractAccountHinter,
	extensioncurrency.CreateContractAccountsFactHinter,
	extensioncurrency.CreateContractAccountsHinter,
	extensioncurrency.CreateContractAccountsItemMultiAmountsHinter,
	extensioncurrency.CreateContractAccountsItemSingleAmountHinter,
	extensioncurrency.WithdrawsFactHinter,
	extensioncurrency.WithdrawsHinter,
	extensioncurrency.WithdrawsItemMultiAmountsHinter,
	extensioncurrency.WithdrawsItemSingleAmountHinter,
	feefi.DepositFactHinter,
	feefi.DepositHinter,
	feefi.FeeOperationFactHinter,
	feefi.FeeOperationHinter,
	feefi.PoolPolicyUpdaterFactHinter,
	feefi.PoolPolicyUpdaterHinter,
	feefi.PoolRegisterFactHinter,
	feefi.PoolRegisterHinter,
	feefi.PoolWithdrawsFactHinter,
	feefi.PoolWithdrawsHinter,
	feefi.FeefiFeeerHinter,
	feefi.PoolHinter,
	feefi.PoolUserBalanceHinter,
	feefi.PoolDesignHinter,
	feefi.PoolPolicyHinter,
	digest.AccountValue{},
	digest.FeefiPoolValue{},
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
