package lomocoin

import (
	"bytes"
	"fmt"
	"math/big"
)

const Uint256IdxLen = 8 // The bigger it is, the more memory is needed, but lower chance of a collision

type Uint256 struct {
	Hash [32]byte
}

func NewSha2Hash(data []byte) (res *Uint256) {
	res = new(Uint256)
	ShaHash(data, res.Hash[:])
	return
}

func (u *Uint256) Bytes() []byte {
	return u.Hash[:]
}

func (u *Uint256) String() (s string) {
	for i := 0; i < 32; i++ {
		s += fmt.Sprintf("%02x", u.Hash[31-i])
	}
	return
}

func (u *Uint256) Equal(o *Uint256) bool {
	return bytes.Equal(u.Hash[:], o.Hash[:])
}

func (u *Uint256) Calc(data []byte) {
	ShaHash(data, u.Hash[:])
}

func BIdx(hash []byte) (o [Uint256IdxLen]byte) {
	copy(o[:], hash[:Uint256IdxLen])
	return
}

func (u *Uint256) BIdx() (o [Uint256IdxLen]byte) {
	o = BIdx(u.Hash[:])
	return
}

func (u *Uint256) BigInt() *big.Int {
	var buf [32]byte
	for i := range buf {
		buf[i] = u.Hash[31-i]
	}
	return new(big.Int).SetBytes(buf[:])
}
