package lomocoin

import (
	"encoding/binary"
	"errors"
	"sync"
	"github.com/syndtr/goleveldb/leveldb"
)

type Block struct {
	Raw               []byte
	Hash              *Uint256
	Txs               []*Tx
	TxCount, TxOffset int

	NoWitnessSize int
	BlockWeight   uint
}

type BlockExtraInfo struct {
	VerifyFlags    uint32
	Height         uint32
	SigopsCost     uint32
	MedianPastTime uint32
}

func NewBlock(data []byte) (bl *Block, er error) {
	if data == nil {
		er = errors.New("nil pointer")
		return
	}
	bl = new(Block)
	bl.Hash = NewSha2Hash(data[:80])
	er = bl.UpdateContent(data)
	return
}

func (bl *Block) UpdateContent(data []byte) error {
	if len(data) < 81 {
		return errors.New("Block too short")
	}
	bl.Raw = data
	bl.TxCount, bl.TxOffset = VLen(data[80:])
	if bl.TxOffset == 0 {
		return errors.New("Block's txn_count field corrupt - RPC_Result:bad-blk-length")
	}
	bl.TxOffset += 80
	return nil
}

func (bl *Block) BuildTxList() (e error) {
	if bl.TxCount == 0 {
		bl.TxCount, bl.TxOffset = VLen(bl.Raw[80:])
		if bl.TxCount == 0 || bl.TxOffset == 0 {
			e = errors.New("Block's txn_count field corrupt - RPC_Result:bad-blk-length")
			return
		}
		bl.TxOffset += 80
	}
	bl.Txs = make([]*Tx, bl.TxCount)

	offs := bl.TxOffset

	var wg sync.WaitGroup
	var data2hash, witness2hash []byte

	bl.NoWitnessSize = 80 + VLenSize(uint64(bl.TxCount))
	bl.BlockWeight = 4 * uint(bl.NoWitnessSize)

	for i := 0; i < bl.TxCount; i++ {
		var n int
		bl.Txs[i], n = NewTx(bl.Raw[offs:])
		if bl.Txs[i] == nil || n == 0 {
			e = errors.New("NewTx failed")
			break
		}
		bl.Txs[i].Raw = bl.Raw[offs:offs+n]
		bl.Txs[i].Size = uint32(n)
		if i == 0 {
			for _, ou := range bl.Txs[0].TxOut {
				ou.WasCoinbase = true
			}
		}

		data2hash = bl.Txs[i].Raw
		bl.Txs[i].NoWitSize = bl.Txs[i].Size
		witness2hash = nil

		bl.BlockWeight += uint(3*bl.Txs[i].NoWitSize + bl.Txs[i].Size)
		bl.NoWitnessSize += len(data2hash)
		wg.Add(1)
		go func(tx *Tx, b, w []byte) {
			tx.Hash.Calc(b)
			if w != nil {
				tx.wTxID.Calc(w)
			}
			wg.Done()
		}(bl.Txs[i], data2hash, witness2hash)
		offs += n
	}

	wg.Wait()

	return
}

func (bl *Block) LeveldbFindBlockHeightWhereHash(db *leveldb.DB, blockHash [32]byte) (result int) {
	query := make([]byte, 43)

	blockQueryHeader := []byte{0x0A, 0x62, 0x6C, 0x6F, 0x63, 0x6B, 0x69, 0x6E, 0x64, 0x65, 0x78}

	copy(query[:11], blockQueryHeader)
	copy(query[11:], blockHash[:])

	dbResult, error := db.Get(query, nil)

	if error != nil {
		panic(error)
	}

	result = int(binary.LittleEndian.Uint32(dbResult[44:48]))

	return
}
