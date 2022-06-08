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

type DepositCommand struct {
	*BaseCommand
	OperationFlags
	Sender AddressFlag                     `arg:"" name:"sender" help:"sender address" required:"true"`
	Pool   AddressFlag                     `arg:"" name:"pool-address" help:"feefi pool address" required:"true"`
	PoolID extensioncmds.ContractIDFlag    `arg:"" name:"pool-id" help:"feefi pool currency id" required:"true"`
	Seal   mitumcmds.FileLoad              `help:"seal" optional:""`
	Amount currencycmds.CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	sender base.Address
	pool   base.Address
	cid    currency.CurrencyID
}

func NewDepositCommand() DepositCommand {
	return DepositCommand{
		BaseCommand: NewBaseCommand("deposit-operation"),
	}
}

func (cmd *DepositCommand) Run(version util.Version) error {
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
	PrettyPrint(cmd.Out, cmd.Pretty, sl)

	return nil
}

func (cmd *DepositCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
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

func (cmd *DepositCommand) createOperation() (operation.Operation, error) { // nolint:dupl
	am := currency.NewAmount(cmd.Amount.Big, cmd.Amount.CID)
	fact := feefi.NewDepositFact([]byte(cmd.Token), cmd.sender, cmd.pool, cmd.PoolID.ID, am)

	var fs []base.FactSign
	sig, err := base.NewFactSignature(cmd.Privatekey, fact, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, err
	}
	fs = append(fs, base.NewBaseFactSign(cmd.Privatekey.Publickey(), sig))

	op, err := feefi.NewDeposit(fact, fs, cmd.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deposits operation")
	}
	return op, nil
}
