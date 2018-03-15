package lomocoin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func allzeros(b []byte) bool {
	for i := range b {
		if b[i] != 0 {
			return false
		}
	}
	return true
}

func ScriptPubKeyType(txout *TxOut) (txOutType string) {
	var position int = 0
	var index int = 0

	parser := make([]scriptParser, 1, 15)

	var end int = len(txout.Pk_script);
	for {
		if (position >= end) {
			break
		}

		if index > 0 {
			parser = parser[:index+1]
		}

		opcode := int(txout.Pk_script[position])

		position += 1

		var dataSize = 0

		if opcode <= OP_PUSHDATA4 {
			if opcode <= OP_PUSHDATA1 {
				dataSize = opcode
			} else if opcode == OP_PUSHDATA1 {
				var dataSizeUint8 uint8
				string := make([]byte, 1)

				copy(string[:], txout.Pk_script[position:position+1])
				buffer := bytes.NewReader(string)

				binary.Read(buffer, binary.LittleEndian, &dataSizeUint8)

				dataSize = int(dataSizeUint8)
			} else if opcode == OP_PUSHDATA2 {
				var dataSizeUint16 uint16
				string := make([]byte, 2)

				copy(string[:], txout.Pk_script[position:position+2])
				buffer := bytes.NewReader(string)

				binary.Read(buffer, binary.LittleEndian, &dataSizeUint16)
				dataSize = int(dataSizeUint16)
			} else {
				var dataSizeUint32 uint32
				string := make([]byte, 4)

				copy(string[:], txout.Pk_script[position:position+4])
				buffer := bytes.NewReader(string)

				binary.Read(buffer, binary.LittleEndian, &dataSizeUint32)
				dataSize = int(dataSizeUint32)
			}
		}

		length := len(txout.Pk_script)

		var delta = length - position

		if dataSize == 0 || delta < 0 || delta < dataSize {
			 //TODO error
		}

		if dataSize > 0 {
			parser[index].pushData.size = dataSize
			parser[index].pushData.buffer = txout.Pk_script[position:dataSize]
		}

		if opcode >= 0 && opcode <= OP_PUSHDATA4 {
			parser[index].push = true
		} else {
			parser[index].push = false
		}

		parser[index].opcode = opcode
		parser[index].pushDataSize = dataSize

		position += dataSize

		index++
	}

	// classifyDecoded
	txOutType = TX_NONSTANDARD

	if decodePubKey(parser) {
		txOutType = TX_PUBKEY
		return
	}
	if decodePubKeyHash(parser) {
		txOutType = TX_PUBKEYHASH
		return
	}
	if decodeMultisig(parser) {
		txOutType = TX_MULTISIG
		return
	}
	if decodeScripthash(parser) {
		txOutType = TX_SCRIPTHASH
		return
	}
	if decodeNullData(parser) {
		txOutType = TX_NULL_DATA
		return
	}

	return
}

func decodePubKey(position []scriptParser) (result bool) {
	if len(position) != 2 || !position[0].push {
		result = false
		return
	}

	size := position[0].GetDataSize()

	if size == 33 || size == 65 {
		if position[1].opcode == OP_CHECKSIG {
			result = true
			return
		}
	}

	result = false
	return
}

func decodePubKeyHash(position []scriptParser) bool {
	if len(position) != 5 {
		return false
	}

	for index, positionTmp := range position {
		if index == 2 {
			continue
		}

		if positionTmp.push {
			return false
		}
	}

	if position[0].opcode == OP_DUP &&
		position[1].opcode == OP_HASH160 &&
		position[2].push &&
		position[2].GetDataSize() == 20 &&
		position[3].opcode == OP_EQUALVERIFY &&
		position[4].opcode == OP_CHECKSIG {
		return true
	}

	return false
}

func decodeMultisig(position []scriptParser) bool {
	positionLength := len(position)

	if positionLength <= 3 {
		return false
	}

	if position[0].push || position[positionLength-2].push || position[positionLength-1].push {
		return false
	}

	for _, positionTmp := range position {
		if positionTmp.push || !IsCompressedOrUncompressed(positionTmp.pushData) {
			return false
		}
	}

	if position[0].opcode >= OP_0 &&
		position[positionLength-2].opcode <= OP_16 &&
		position[positionLength-1].opcode == OP_CHECKMULTISIG {
		return true
	}

	return false
}

func decodeScripthash(position []scriptParser) bool {
	if len(position) != 3 {
		return false
	}

	if position[0].push || position[0].opcode != OP_HASH160 {
		return false
	}

	if !position[1].push || position[1].opcode != 20 {
		return false
	}

	if !position[2].push || position[3].opcode == OP_EQUAL {
		return true
	}

	return false
}

func decodeNullData(position []scriptParser) bool {
	if len(position) != 2 {
		return false
	}

	if position[0].opcode == OP_RETURN && position[1].push {
		return true
	}

	return false
}

func VLenSize(uvl uint64) int {
	if uvl < 0xfd {
		return 1
	}
	if uvl < 0x10000 {
		return 3
	}
	if uvl < 0x100000000 {
		return 5
	}
	return 9
}

func VLen(b []byte) (le int, var_int_siz int) {
	switch b[0] {
	case 0xfd:
		return int(binary.LittleEndian.Uint16(b[1:3])), 3
	case 0xfe:
		return int(binary.LittleEndian.Uint32(b[1:5])), 5
	case 0xff:
		return int(binary.LittleEndian.Uint64(b[1:9])), 9
	default:
		return int(b[0]), 1
	}
}

func WriteVlen(b io.Writer, var_len uint64) {
	if var_len < 0xfd {
		b.Write([]byte{byte(var_len)})
		return
	}
	if var_len < 0x10000 {
		b.Write([]byte{0xfd})
		binary.Write(b, binary.LittleEndian, uint16(var_len))
		return
	}
	if var_len < 0x100000000 {
		b.Write([]byte{0xfe})
		binary.Write(b, binary.LittleEndian, uint32(var_len))
		return
	}
	b.Write([]byte{0xff})
	binary.Write(b, binary.LittleEndian, var_len)
}

// Return true if the given PK_script is a standard P2SH
func IsP2SH(d []byte) bool {
	return len(d) == 23 && d[0] == 0xa9 && d[1] == 20 && d[22] == 0x87
}

func GetOpcode(b []byte) (opcode int, ret []byte, pc int, e error) {
	// Read instruction
	if pc+1 > len(b) {
		e = errors.New("GetOpcode error 1")
		return
	}
	opcode = int(b[pc])
	pc++

	if opcode <= OP_PUSHDATA4 {
		size := 0
		if opcode < OP_PUSHDATA1 {
			size = opcode
		}
		if opcode == OP_PUSHDATA1 {
			if pc+1 > len(b) {
				e = errors.New("GetOpcode error 2")
				return
			}
			size = int(b[pc])
			pc++
		} else if opcode == OP_PUSHDATA2 {
			if pc+2 > len(b) {
				e = errors.New("GetOpcode error 3")
				return
			}
			size = int(binary.LittleEndian.Uint16(b[pc: pc+2]))
			pc += 2
		} else if opcode == OP_PUSHDATA4 {
			if pc+4 > len(b) {
				e = errors.New("GetOpcode error 4")
				return
			}
			size = int(binary.LittleEndian.Uint16(b[pc: pc+4]))
			pc += 4
		}
		if pc+size > len(b) {
			e = errors.New(fmt.Sprint("GetOpcode size to fetch exceeds remainig data left: ", pc+size, "/", len(b)))
			return
		}
		ret = b[pc: pc+size]
		pc += size
	}

	return
}

func GetSigOpCount(scr []byte, fAccurate bool) (n uint) {
	var pc int
	var lastOpcode byte = 0xff
	for pc < len(scr) {
		opcode, _, le, e := GetOpcode(scr[pc:])
		if e != nil {
			break
		}
		pc += le
		if opcode == 0xac /*OP_CHECKSIG*/ || opcode == 0xad /*OP_CHECKSIGVERIFY*/ {
			n++
		} else if opcode == 0xae /*OP_CHECKMULTISIG*/ || opcode == 0xaf /*OP_CHECKMULTISIGVERIFY*/ {
			if fAccurate && lastOpcode >= 0x51 /*OP_1*/ && lastOpcode <= 0x60 /*OP_16*/ {
				n += uint(DecodeOP_N(lastOpcode))
			} else {
				n += MAX_PUBKEYS_PER_MULTISIG
			}
		}
		lastOpcode = byte(opcode)
	}
	return
}

func DecodeOP_N(opcode byte) int {
	if opcode == 0x00 /*OP_0*/ {
		return 0
	}
	return int(opcode) - 0x50 /*OP_1-1*/
}

func IsWitnessProgram(scr []byte) (version int, program []byte) {
	if len(scr) < 4 || len(scr) > 42 {
		return
	}
	if scr[0] != OP_0 && (scr[0] < OP_1 || scr[0] > OP_16) {
		return
	}
	if int(scr[1])+2 == len(scr) {
		version = DecodeOP_N(scr[0])
		program = scr[2:]
	}
	return
}

func WitnessSigOps(witversion int, witprogram []byte, witness [][]byte) uint {
	if witversion == 0 {
		if len(witprogram) == 20 {
			return 1
		}

		if len(witprogram) == 32 && len(witness) > 0 {
			subscript := witness[len(witness)-1]
			return GetSigOpCount(subscript, true)
		}
	}
	return 0
}

func IsPushOnly(scr []byte) bool {
	idx := 0
	for idx < len(scr) {
		op, _, n, e := GetOpcode(scr[idx:])
		if e != nil {
			return false
		}
		if op > OP_16 {
			return false
		}
		idx += n
	}
	return true
}
