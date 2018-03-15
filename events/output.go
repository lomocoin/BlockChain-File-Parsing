package events

import (
	"log"
	"encoding/hex"
	"github.com/lomocoin/blockchain-parsing/lomocoin"
)

func Put(bl *lomocoin.Block) {
	log.Printf("Block hash: %v", bl.Hash)
	log.Printf("Block size: %v", len(bl.Raw))

	for _, tx := range bl.Txs {
		log.Printf("======================================")
		log.Printf("TxId: %v", tx.Hash.String())
		log.Printf("Tx Size: %v", tx.Size)
		log.Printf("Tx Lock time: %v", tx.Lock_time)
		log.Printf("Tx Version: %v", tx.Version)
		log.Printf("Tx Time: %v", tx.TxTime)

		log.Println("TxIns:")

		if tx.IsCoinBase() {
			log.Printf("TxIn coinbase, newly generated coins")
		} else {
			for txin_index, txin := range tx.TxIn {
				log.Printf("TxIn index: %v", txin_index)
				log.Printf("TxIn Txid: %v", txin.Input.HashString())
				log.Printf("TxIn Vout: %v", txin.Input.VoutString())
				log.Printf("TxIn ScriptSig hex: %v", hex.EncodeToString(txin.ScriptSig))
				log.Printf("TxIn Sequence: %v", txin.Sequence)
			}
		}

		log.Println("TxOuts:")

		for txo_index, txout := range tx.TxOut {
			log.Printf("TxOut index: %v", txo_index)
			log.Printf("TxOut value: %v", txout.Value)
			log.Printf("TxOut value: %v", txout.Value)
			log.Printf("TxOut scriptPubKey hex: %s", hex.EncodeToString(txout.Pk_script))
			log.Printf("TxOut scriptPubKey type: %s", lomocoin.ScriptPubKeyType(txout))

			txout_addr := lomocoin.NewAddrFromPkScript(txout.Pk_script, false)
			if txout_addr != nil {
				log.Printf("TxOut address: %v", txout_addr.String())
			} else {
				log.Printf("TxOut address: can't decode address")
			}
		}
	}
	log.Println()
}