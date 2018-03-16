package events

import (
	"encoding/hex"
	"github.com/lomocoin/blockchain-parsing/lomocoin"
	"github.com/therecipe/qt/widgets"
	"fmt"
)

func Put(bl *lomocoin.Block, console *widgets.QTextEdit) {
	console.InsertPlainText(fmt.Sprintf("Block hash: %v \n", bl.Hash))
	console.InsertPlainText(fmt.Sprintf("Block size: %v \n", len(bl.Raw)))

	for index, tx := range bl.Txs {
		console.InsertPlainText(fmt.Sprintf("==================TxIndex %v=================== \n", index))
		console.InsertPlainText(fmt.Sprintf("TxId: %v\n", tx.Hash.String()))
		console.InsertPlainText(fmt.Sprintf("Tx Size: %v\n", tx.Size))
		console.InsertPlainText(fmt.Sprintf("Tx Lock time: %v\n", tx.Lock_time))
		console.InsertPlainText(fmt.Sprintf("Tx Version: %v\n", tx.Version))
		console.InsertPlainText(fmt.Sprintf("Tx Time: %v\n", tx.TxTime))

		console.InsertPlainText(fmt.Sprintf("TxIns:\n"))

		if tx.IsCoinBase() {
			console.InsertPlainText(fmt.Sprintf("TxIn coinbase, newly generated coins\n"))
		} else {
			for txin_index, txin := range tx.TxIn {
				console.InsertPlainText(fmt.Sprintf("TxIn index: %v\n", txin_index))
				console.InsertPlainText(fmt.Sprintf("TxIn Txid: %v\n", txin.Input.HashString()))
				console.InsertPlainText(fmt.Sprintf("TxIn Vout: %v\n", txin.Input.VoutString()))
				console.InsertPlainText(fmt.Sprintf("TxIn ScriptSig hex: %v\n", hex.EncodeToString(txin.ScriptSig)))
				console.InsertPlainText(fmt.Sprintf("TxIn Sequence: %v\n", txin.Sequence))
			}
		}

		console.InsertPlainText(fmt.Sprintf("TxOuts:\n"))

		for txo_index, txout := range tx.TxOut {
			console.InsertPlainText(fmt.Sprintf("TxOut index: %v\n", txo_index))
			console.InsertPlainText(fmt.Sprintf("TxOut value: %v\n", txout.Value))
			console.InsertPlainText(fmt.Sprintf("TxOut scriptPubKey hex: %s\n", hex.EncodeToString(txout.Pk_script)))
			console.InsertPlainText(fmt.Sprintf("TxOut scriptPubKey type: %s\n", lomocoin.ScriptPubKeyType(txout)))

			txout_addr := lomocoin.NewAddrFromPkScript(txout.Pk_script, false)
			if txout_addr != nil {
				console.InsertPlainText(fmt.Sprintf("TxOut address: %v\n", txout_addr.String()))
			} else {
				console.InsertPlainText(fmt.Sprintf("TxOut address: can't decode address\n"))
			}
		}
	}
}