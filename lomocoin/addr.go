package lomocoin

import (
	"math/big"
)

type LomocoinAddr struct {
	Version  byte
	Hash160  [20]byte
	Checksum []byte
	Pubkey   []byte
	Enc58str string
}

func NewAddrFromHash160(in []byte, ver byte) (a *LomocoinAddr) {
	a = new(LomocoinAddr)
	a.Version = ver
	copy(a.Hash160[:], in[:])
	return
}

func NewAddrFromPubkey(in []byte, ver byte) (a *LomocoinAddr) {
	a = new(LomocoinAddr)
	a.Pubkey = make([]byte, len(in))
	copy(a.Pubkey[:], in[:])
	a.Version = ver
	RimpHash(in, a.Hash160[:])
	return
}

func AddrVerPubkey(testnet bool) byte {
	if testnet {
		return 111
	} else {
		return 48
	}
}

func AddrVerScript(testnet bool) byte {
	if testnet {
		return 196
	} else {
		return 125
	}
}

func NewAddrFromPkScript(scr []byte, testnet bool) *LomocoinAddr {
	if len(scr) == 0 {
		return nil
	}

	if len(scr) == 25 && scr[0] == 0x76 && scr[1] == 0xa9 && scr[2] == 0x14 && scr[23] == 0x88 && scr[24] == 0xac {
		return NewAddrFromHash160(scr[3:23], AddrVerPubkey(testnet))
	} else if len(scr) == 67 && scr[0] == 0x41 && scr[66] == 0xac {
		return NewAddrFromPubkey(scr[1:66], AddrVerPubkey(testnet))
	} else if len(scr) == 35 && scr[0] == 0x21 && scr[34] == 0xac {
		return NewAddrFromPubkey(scr[1:34], AddrVerPubkey(testnet))
	} else if len(scr) == 23 && scr[0] == 0xa9 && scr[1] == 0x14 && scr[22] == 0x87 {
		return NewAddrFromHash160(scr[2:22], AddrVerScript(testnet))
	}

	return nil
}

func (a *LomocoinAddr) String() string {
	if a.Enc58str == "" {
		var ad [25]byte
		ad[0] = a.Version // PUBKEY_ADDRESS
		copy(ad[1:21], a.Hash160[:])
		if a.Checksum == nil {
			sh := Sha2Sum(ad[0:21])
			a.Checksum = make([]byte, 4)
			copy(a.Checksum, sh[:4])
		}
		copy(ad[21:25], a.Checksum[:])
		a.Enc58str = Encodeb58(ad[:])
	}
	return a.Enc58str
}

var b58set []byte = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

var bn0 *big.Int = big.NewInt(0)
var bn58 *big.Int = big.NewInt(58)

func Encodeb58(a []byte) (s string) {
	idx := len(a)*138/100 + 1
	buf := make([]byte, idx)
	bn := new(big.Int).SetBytes(a)
	var mo *big.Int
	for bn.Cmp(bn0) != 0 {
		bn, mo = bn.DivMod(bn, bn58, new(big.Int))
		idx--
		buf[idx] = b58set[mo.Int64()]
	}
	for i := range a {
		if a[i] != 0 {
			break
		}
		idx--
		buf[idx] = b58set[0]
	}

	s = string(buf[idx:])

	return
}
