package cmds

import (
	"github.com/pkg/errors"
	"github.com/protoconNet/mitum-account-extension/extension"

	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	currency "github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	mitumcmds "github.com/spikeekips/mitum/launch/cmds"
	"github.com/spikeekips/mitum/util"
)

type WithdrawCommand struct {
	*BaseCommand
	OperationFlags
	Sender  AddressFlag                       `arg:"" name:"sender" help:"sender address" required:"true"`
	Target  AddressFlag                       `arg:"" name:"target" help:"target contract account address" required:"true"`
	Seal    mitumcmds.FileLoad                `help:"seal" optional:""`
	Amounts []currencycmds.CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	sender  base.Address
	target  base.Address
}

func NewWithdrawCommand() WithdrawCommand {
	return WithdrawCommand{
		BaseCommand: NewBaseCommand("withdraw-operation"),
	}
}

func (cmd *WithdrawCommand) Run(version util.Version) error {
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

func (cmd *WithdrawCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Amounts) < 1 {
		return errors.Errorf("empty currency-amount, must be given at least one")
	}

	if sender, err := cmd.Sender.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else if receiver, err := cmd.Target.Encode(jenc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else {
		cmd.sender = sender
		cmd.target = receiver
	}

	return nil
}

func (cmd *WithdrawCommand) createOperation() (operation.Operation, error) { // nolint:dupl
	i, err := loadOperations(cmd.Seal.Bytes(), cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}

	var items []extension.WithdrawsItem
	for j := range i {
		if t, ok := i[j].(currency.Transfers); ok {
			items = t.Fact().(extension.WithdrawsFact).Items()
		}
	}

	ams := make([]currency.Amount, len(cmd.Amounts))
	for i := range cmd.Amounts {
		a := cmd.Amounts[i]
		am := currency.NewAmount(a.Big, a.CID)
		if err = am.IsValid(nil); err != nil {
			return nil, err
		}

		ams[i] = am
	}

	item := extension.NewWithdrawsItemMultiAmounts(cmd.target, ams)
	if err = item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := extension.NewWithdrawsFact([]byte(cmd.Token), cmd.sender, items)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(cmd.Privatekey, fact, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.Privatekey.Publickey(), sig))

	op, err := extension.NewWithdraws(fact, fs, cmd.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create withdraws operation")
	}
	return op, nil
}
