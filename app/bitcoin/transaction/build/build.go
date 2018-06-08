package build

import (
	"fmt"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/memo/app/bitcoin/memo"
	"github.com/memocash/memo/app/bitcoin/transaction"
	"github.com/memocash/memo/app/bitcoin/wallet"
	"github.com/memocash/memo/app/db"
)

func Build(spendOutputs []transaction.SpendOutput, privateKey *wallet.PrivateKey) (*wire.MsgTx, error) {
	var minInput = int64(memo.BaseTxFee + memo.InputFeeP2PKH + memo.OutputFeeP2PKH + transaction.DustMinimumOutput)

	for _, spendOutput := range spendOutputs {
		switch spendOutput.Type {
		case transaction.SpendOutputTypeP2PK:
			minInput += memo.OutputFeeP2PKH + spendOutput.Amount
		default:
			outputFee, err := getMemoOutputFee(spendOutput)
			if err != nil {
				return nil, jerr.Get("error getting memo output fee", err)
			}
			minInput += outputFee
		}
	}

	txOuts, err := db.GetSpendableTxOuts(privateKey.GetPublicKey().GetAddress().GetScriptAddress(), minInput)
	if err != nil {
		return nil, jerr.Get("error getting spendable tx out", err)
	}

	var totalInputs int64
	for _, txOut := range txOuts {
		totalInputs += txOut.Value
	}

	var fee = int64(memo.BaseTxFee + len(txOuts)*memo.InputFeeP2PKH) + memo.OutputFeeP2PKH

	var totalOutputs int64
	for _, spendOutput := range spendOutputs {
		totalOutputs += spendOutput.Amount
		switch spendOutput.Type {
		case transaction.SpendOutputTypeP2PK:
			fee += memo.OutputFeeP2PKH
		default:
			outputFee, err := getMemoOutputFee(spendOutput)
			if err != nil {
				return nil, jerr.Get("error getting memo output fee", err)
			}
			fee += outputFee
		}
	}

	var change = totalInputs - fee - totalOutputs
	fmt.Printf("totalInputs: %d, fee: %d, totalOutputs: %d, change: %d)", totalInputs, fee, totalOutputs, change)
	if change < transaction.DustMinimumOutput {
		return nil, jerr.New("not enough funds")
	}
	spendOutputs = append(spendOutputs, transaction.SpendOutput{
		Type:    transaction.SpendOutputTypeP2PK,
		Address: privateKey.GetPublicKey().GetAddress(),
		Amount:  change,
	})

	var tx *wire.MsgTx
	tx, err = transaction.Create(txOuts, privateKey, spendOutputs)
	if err != nil {
		return nil, jerr.Get("error creating tx", err)
	}
	return tx, nil
}

func getMemoOutputFee(spendOutput transaction.SpendOutput) (int64, error) {
	switch spendOutput.Type {
	case transaction.SpendOutputTypeMemoLike:
		return int64(memo.OutputFeeOpReturn + len(spendOutput.Data)), nil
	case transaction.SpendOutputTypeMemoSetProfilePic:
		return int64(memo.OutputFeeOpReturn + len(spendOutput.Data)), nil
	}
	return 0, jerr.New("unable to get fee for output type")
}
