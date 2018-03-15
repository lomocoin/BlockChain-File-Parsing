package main

import (
	"log"
	"github.com/lomocoin/blockchain-parsing/lomocoin"
	"github.com/lomocoin/blockchain-parsing/lib/others/blockdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/lomocoin/blockchain-parsing/events"
	"encoding/csv"
	"os"
)

func main() {
	Magic := [4]byte{0xa6, 0xb8, 0xc9, 0xd5}

	BlockDatabase := blockdb.NewBlockDB("/Users/viest/Desktop/LoMoCoin", Magic)

	db, leveldbErr := leveldb.OpenFile("/Users/viest/Desktop/LoMoCoin/blockindex",nil)
	if leveldbErr != nil {
		panic(leveldbErr)
	}

	start_block := 1
	end_block := 1

	var state string = "output"
	var writeObj *csv.Writer

	if state == "csv" {
		file, err := os.Create("/Users/viest/Desktop/block33.csv")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		file.WriteString("\xEF\xBB\xBF")

		writeObj = csv.NewWriter(file)
		writeObj.Write([]string{"block_index", "block_hash", "tx_time", "tx_id", "tx_in", "tx_out"})
	}

	// 0 block has an exception and must be skipped by the file pointer.
	for blockIndex := 0; blockIndex <= end_block; blockIndex++ {
		dat, er := BlockDatabase.FetchNextBlock()
		if dat == nil || er != nil {
			log.Println("END of DB file")
			break
		}
		bl, er := lomocoin.NewBlock(dat[:])

		if er != nil {
			println("Block inconsistent:", er.Error())
			break
		}

		if blockIndex < start_block {
			continue
		}

		dbBlockHeight := bl.LeveldbFindBlockHeightWhereHash(db, bl.Hash.Hash)

		// Fork block
		if blockIndex != dbBlockHeight {
			blockIndex = dbBlockHeight
		}

		log.Printf("Current block height: %v", blockIndex)

		bl.BuildTxList()

		if state == "csv" {
			events.WriteCSV(bl, blockIndex, writeObj)
		}

		if state != "csv" {
			events.Put(bl)
		}
	}
}
