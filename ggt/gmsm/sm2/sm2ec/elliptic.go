package sm2ec

import (
	"crypto/elliptic"
	"math/big"
	"sync"
)

var initonce sync.Once

var sm2Params = &elliptic.CurveParams{
	Name:    "sm2p256v1",
	BitSize: 256,
	P:       bigFromHex("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF"),
	N:       bigFromHex("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123"),
	B:       bigFromHex("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93"),
	Gx:      bigFromHex("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7"),
	Gy:      bigFromHex("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0"),
}

func bigFromHex(s string) *big.Int {
	b, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("sm2/elliptic: internal error: invalid encoding")
	}
	return b
}

func initAll() {
	initSM2P256()
}

func P256() elliptic.Curve {
	initonce.Do(initAll)
	return sm2p256
}

// Since golang 1.19
// unmarshaler is implemented by curves with their own constant-time Unmarshal.
// There isn't an equivalent interface for Marshal/MarshalCompressed because
// that doesn't involve any mathematical operations, only FillBytes and Bit.
type unmarshaler interface {
	Unmarshal([]byte) (x, y *big.Int)
	UnmarshalCompressed([]byte) (x, y *big.Int)
}

func Unmarshal(curve elliptic.Curve, data []byte) (x, y *big.Int) {
	if c, ok := curve.(unmarshaler); ok {
		return c.Unmarshal(data)
	}
	return elliptic.Unmarshal(curve, data)
}

// UnmarshalCompressed converts a point, serialized by MarshalCompressed, into
// an x, y pair. It is an error if the point is not in compressed form, is not
// on the curve, or is the point at infinity. On error, x = nil.
func UnmarshalCompressed(curve elliptic.Curve, data []byte) (x, y *big.Int) {
	if c, ok := curve.(unmarshaler); ok {
		return c.UnmarshalCompressed(data)
	}
	return elliptic.UnmarshalCompressed(curve, data)
}
