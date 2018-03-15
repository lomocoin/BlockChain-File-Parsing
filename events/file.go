package events

import (
	"fmt"
	"bytes"
	"encoding/hex"
	"github.com/lomocoin/blockchain-parsing/lomocoin"
	"encoding/csv"
)

func WriteCSV(bl *lomocoin.Block, blockIndex int, writeObj *csv.Writer) {
	for _, tx := range bl.Txs {
		data := make([]string, 6)
		data[0] = fmt.Sprintf("%v", blockIndex)
		data[1] = fmt.Sprintf("%v", bl.Hash)
		data[2] = fmt.Sprintf("%v", tx.TxTime)
		data[3] = fmt.Sprintf("%v", tx.Hash.String())

		if tx.IsCoinBase() {
			data[4] = "[{\"index\":0, \"tx_id\":\"\", \"vout\":0, \"coinbase\":1}]"
		} else {
			var vin bytes.Buffer

			for txin_index, txin := range tx.TxIn {
				if txin_index == 0 {
					vin.WriteString("[")
				}

				if txin_index > 0 {
					vin.WriteString(",")
				}

				vin.WriteString("{\"index\":")
				vin.WriteString(fmt.Sprintf("%v", txin_index))
				vin.WriteString(",\"tx_id\":\"")
				vin.WriteString(fmt.Sprintf("%v", txin.Input.HashString()))
				vin.WriteString("\",\"vout\":")
				vin.WriteString(fmt.Sprintf("%v", txin.Input.Vout))
				vin.WriteString(",\"coinbase\":0}")

				if txin_index == (len(tx.TxIn) - 1) {
					vin.WriteString("]")
				}


			}

			data[4] = fmt.Sprintf("%v", vin.String())
		}

		var vout bytes.Buffer

		for txo_index, txout := range tx.TxOut {
			if txo_index == 0 {
				vout.WriteString("[")
			}

			if txo_index > 0 {
				vout.WriteString(",")
			}

			vout.WriteString("{\"index\":")
			vout.WriteString(fmt.Sprintf("%v", txo_index))
			vout.WriteString(",\"value\":")
			vout.WriteString(fmt.Sprintf("%v", txout.Value))
			vout.WriteString(",\"script\":\"")
			vout.WriteString(fmt.Sprintf("%s", hex.EncodeToString(txout.Pk_script)))
			vout.WriteString("\",\"script_type\":\"")
			vout.WriteString(fmt.Sprintf("%s", lomocoin.ScriptPubKeyType(txout)))
			vout.WriteString("\",\"address\":\"")

			txout_addr := lomocoin.NewAddrFromPkScript(txout.Pk_script, false)
			if txout_addr != nil {
				vout.WriteString(fmt.Sprintf("%v", txout_addr.String()))
			} else {
				vout.WriteString("can_not_decode")
			}

			vout.WriteString("\"}")

			if txo_index == (len(tx.TxOut) - 1) {
				vout.WriteString("]")
			}
		}

		data[5] = fmt.Sprintf("%v", vout.String())

		writeObj.Write(data)
	}

	writeObj.Flush()
}