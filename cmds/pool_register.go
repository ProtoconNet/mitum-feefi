package cmds

import (
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"
	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/util"
)

type PoolRegisterCommand struct {
	*BaseCommand
	OperationFlags
	Sender         AddressFlag                     `arg:"" name:"sender" help:"sender address" required:"true"`
	Pool           currencycmds.AddressFlag        `arg:"" name:"pool" help:"pool address" required:"true"`
	IncomeCurrency currencycmds.CurrencyIDFlag     `arg:"" name:"income-fee-currency-id" help:"income fee currency id" required:"true"`
	OutlayCurrency currencycmds.CurrencyIDFlag     `arg:"" name:"outlay-fee-currency-id" help:"outlay fee currency id" required:"true"`
	InitialFee     currencycmds.CurrencyAmountFlag `arg:"" name:"initial-fee" help:"initial fee amount" required:"true"`
	Currency       currencycmds.CurrencyIDFlag     `arg:"" name:"currency-id" help:"currency id" required:"true"`
	sender         base.Address
	pool           base.Address
}

func NewPoolRegisterCommand() PoolRegisterCommand {
	return PoolRegisterCommand{
		BaseCommand: NewBaseCommand("pool-register-operation"),
	}
}

func (cmd *PoolRegisterCommand) Run(version util.Version) error { // nolint:dupl
	if err := cmd.Initialize(cmd, version); err != nil {
		return errors.Wrap(err, "failed to initialize command")
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	var op operation.Operation
	if i, err := cmd.createOperation(); err != nil {
		return errors.Wrap(err, "failed to create pool-register operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid currency-register operation")
	} else {
		cmd.Log().Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	i, err := operation.NewBaseSeal(
		cmd.OperationFlags.Privatekey,
		[]operation.Operation{op},
		[]byte(cmd.OperationFlags.NetworkID),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create operation.Seal")
	}
	cmd.Log().Debug().Interface("seal", i).Msg("seal loaded")

	PrettyPrint(cmd.Out, cmd.Pretty, i)

	return nil
}

func (cmd *PoolRegisterCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}
	if sender, err := cmd.Sender.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else if target, err := cmd.Pool.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid pool address format, %q", cmd.Pool.String())
	} else {
		cmd.sender = sender
		cmd.pool = target
	}

	return nil
}

func (cmd *PoolRegisterCommand) createOperation() (feefi.PoolRegister, error) {
	am := currency.NewAmount(cmd.InitialFee.Big, cmd.InitialFee.CID)
	fact := feefi.NewPoolRegisterFact([]byte(cmd.Token), cmd.sender, cmd.pool, am, cmd.IncomeCurrency.CID, cmd.OutlayCurrency.CID, cmd.Currency.CID)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(
		cmd.OperationFlags.Privatekey,
		fact,
		[]byte(cmd.OperationFlags.NetworkID),
	)
	if err != nil {
		return feefi.PoolRegister{}, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.OperationFlags.Privatekey.Publickey(), sig))

	return feefi.NewPoolRegister(fact, fs, cmd.OperationFlags.Memo)
}
