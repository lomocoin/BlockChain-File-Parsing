package lomocoin

import (
	"encoding/hex"
	"errors"
	"github.com/lomocoin/blockchain-parsing/lib/secp256k1"
)

type PublicKey struct {
	secp256k1.XY
}

type Signature struct {
	secp256k1.Signature
	HashType byte
}

func NewPublicKey(buf []byte) (res *PublicKey, e error) {
	res = new(PublicKey)
	if !res.XY.ParsePubkey(buf) {
		e = errors.New("NewPublicKey: Unknown format: " + hex.EncodeToString(buf[:]))
		res = nil
	}
	return
}

func NewSignature(buf []byte) (*Signature, error) {
	sig := new(Signature)
	le := sig.ParseBytes(buf)
	if le < 0 {
		return nil, errors.New("NewSignature: ParseBytes error")
	}
	if le < len(buf) {
		sig.HashType = buf[len(buf)-1]
	}
	return sig, nil
}

// Recoved public key form a signature
func (sig *Signature) RecoverPublicKey(msg []byte, recid int) (key *PublicKey) {
	key = new(PublicKey)
	if !secp256k1.RecoverPublicKey(sig.R.Bytes(), sig.S.Bytes(), msg, recid, &key.XY) {
		key = nil
	}
	return
}

func (sig *Signature) IsLowS() bool {
	return sig.S.Cmp(&secp256k1.TheCurve.HalfOrder.Int) < 1
}

// Returns serialized canoncal signature followed by a hash type
func (sig *Signature) Bytes() []byte {
	return append(sig.Signature.Bytes(), sig.HashType)
}

func IsCompressedOrUncompressed (pushData PushData) bool {

	 buffer := pushData.GetPushDataBuffer()

	if pushData.size < 33 {
		return false
	}

	if buffer[0] == 0x04 {
		if pushData.size != 65 {
			return false
		}
	} else if buffer[0] == 0x02 || buffer[0] == 0x03 {
		if pushData.size != 33 {
			return false
		}
	} else {
		return false
	}

	return true
}
