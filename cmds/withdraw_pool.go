package cmds

import (
	extensioncmds "github.com/ProtoconNet/mitum-currency-extension/cmds"
	"github.com/ProtoconNet/mitum-feefi/feefi"
	"github.com/pkg/errors"

	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	currency "github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	mitumcmds "github.com/spikeekips/mitum/launch/cmds"
	"github.com/spikeekips/mitum/util"
)

type WithdrawPoolCommand struct {
	*BaseCommand
	OperationFlags
	Sender  AddressFlag                       `arg:"" name:"sender" help:"sender address" required:"true"`
	Pool    AddressFlag                       `arg:"" name:"pool" help:"pool address" required:"true"`
	PoolID  extensioncmds.ContractIDFlag      `arg:"" name:"pool-id" help:"pool currency id" required:"true"`
	Seal    mitumcmds.FileLoad                `help:"seal" optional:""`
	Amounts []currencycmds.CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	sender  base.Address
	pool    base.Address
}

func NewWithdrawPoolCommand() WithdrawPoolCommand {
	return WithdrawPoolCommand{
		BaseCommand: NewBaseCommand("withdraw-operation"),
	}
}

func (cmd *WithdrawPoolCommand) Run(version util.Version) error {
	if err := cmd.Initialize(cmd, version); err != nil {
		return errors.Wrap(err, "failed to initialize command")
	}

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	sl, err := LoadSealAndAddOperation(
		cmd.Seal.Bytes(),
		cmd.Privatekey,
		cmd.NetworkID.NetworkID(),
		op,
	)
	if err != nil {
		return err
	}
	currencycmds.PrettyPrint(cmd.Out, cmd.Pretty, sl)

	return nil
}

func (cmd *WithdrawPoolCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Amounts) < 1 {
		return errors.Errorf("empty currency-amount, must be given at least one")
	}

	if sender, err := cmd.Sender.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else if receiver, err := cmd.Pool.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else {
		cmd.sender = sender
		cmd.pool = receiver
	}

	return nil
}

func (cmd *WithdrawPoolCommand) createOperation() (operation.Operation, error) { // nolint:dupl
	ams := make([]currency.Amount, len(cmd.Amounts))
	for i := range cmd.Amounts {
		a := cmd.Amounts[i]
		am := currency.NewAmount(a.Big, a.CID)
		if err := am.IsValid(nil); err != nil {
			return nil, err
		}

		ams[i] = am
	}

	fact := feefi.NewPoolWithdrawsFact([]byte(cmd.Token), cmd.sender, cmd.pool, cmd.PoolID.ID, ams)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(cmd.Privatekey, fact, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.Privatekey.Publickey(), sig))

	op, err := feefi.NewPoolWithdraws(fact, fs, cmd.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create withdraws operation")
	}
	return op, nil
}
