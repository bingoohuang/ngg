package cipher_test

import (
	"bytes"
	"crypto/cipher"
	"testing"

	"github.com/emmansun/gmsm/internal/cryptotest"
	"github.com/emmansun/gmsm/sm4"
)

var ctrSM4Tests = []struct {
	name string
	key  []byte
	iv   []byte
	in   []byte
	out  []byte
}{
	{
		"2 blocks",
		[]byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		[]byte{
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51},
		[]byte{
			0xbc, 0x71, 0x0d, 0x76, 0x2d, 0x07, 0x0b, 0x26, 0x36, 0x1d, 0xa8, 0x2b, 0x54, 0x56, 0x5e, 0x46,
			0xb0, 0x2b, 0x3d, 0xbd, 0xdd, 0x50, 0xd5, 0xb4, 0x58, 0xae, 0xcc, 0xb2, 0x5d, 0xa1, 0x05, 0xe1},
	},
	{
		"4 blocks",
		[]byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		[]byte{
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
		},
		[]byte{
			0xbc, 0x71, 0x0d, 0x76, 0x2d, 0x07, 0x0b, 0x26, 0x36, 0x1d, 0xa8, 0x2b, 0x54, 0x56, 0x5e, 0x46,
			0xb0, 0x2b, 0x3d, 0xbd, 0xdd, 0x50, 0xd5, 0xb4, 0x58, 0xae, 0xcc, 0xb2, 0x5d, 0xa1, 0x05, 0xe1,
			0x6a, 0xd7, 0x0b, 0xc0, 0x11, 0x75, 0xad, 0x43, 0xb0, 0x80, 0x6a, 0x2e, 0x7b, 0x9c, 0xa5, 0x45,
			0x60, 0x24, 0x59, 0xa0, 0x6b, 0x7d, 0x13, 0x0d, 0xde, 0x42, 0xa3, 0xe0, 0x47, 0x68, 0x18, 0xd2,
		},
	},
	{
		"6 blocks",
		[]byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		[]byte{
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
		},
		[]byte{
			0xbc, 0x71, 0x0d, 0x76, 0x2d, 0x07, 0x0b, 0x26, 0x36, 0x1d, 0xa8, 0x2b, 0x54, 0x56, 0x5e, 0x46,
			0xb0, 0x2b, 0x3d, 0xbd, 0xdd, 0x50, 0xd5, 0xb4, 0x58, 0xae, 0xcc, 0xb2, 0x5d, 0xa1, 0x05, 0xe1,
			0x6a, 0xd7, 0x0b, 0xc0, 0x11, 0x75, 0xad, 0x43, 0xb0, 0x80, 0x6a, 0x2e, 0x7b, 0x9c, 0xa5, 0x45,
			0x60, 0x24, 0x59, 0xa0, 0x6b, 0x7d, 0x13, 0x0d, 0xde, 0x42, 0xa3, 0xe0, 0x47, 0x68, 0x18, 0xd2,
			0x00, 0xb8, 0x33, 0x1a, 0x66, 0x57, 0xd6, 0xbe, 0xb8, 0x5b, 0x72, 0x4f, 0x55, 0x0c, 0xd5, 0x2d,
			0x96, 0xf3, 0xe4, 0x12, 0x37, 0xa2, 0x07, 0x44, 0x43, 0xa5, 0x43, 0x3a, 0x41, 0x33, 0x0d, 0xca,
		},
	},
	{
		"8 blocks",
		[]byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		[]byte{
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
		},
		[]byte{
			0xbc, 0x71, 0x0d, 0x76, 0x2d, 0x07, 0x0b, 0x26, 0x36, 0x1d, 0xa8, 0x2b, 0x54, 0x56, 0x5e, 0x46,
			0xb0, 0x2b, 0x3d, 0xbd, 0xdd, 0x50, 0xd5, 0xb4, 0x58, 0xae, 0xcc, 0xb2, 0x5d, 0xa1, 0x05, 0xe1,
			0x6a, 0xd7, 0x0b, 0xc0, 0x11, 0x75, 0xad, 0x43, 0xb0, 0x80, 0x6a, 0x2e, 0x7b, 0x9c, 0xa5, 0x45,
			0x60, 0x24, 0x59, 0xa0, 0x6b, 0x7d, 0x13, 0x0d, 0xde, 0x42, 0xa3, 0xe0, 0x47, 0x68, 0x18, 0xd2,
			0x00, 0xb8, 0x33, 0x1a, 0x66, 0x57, 0xd6, 0xbe, 0xb8, 0x5b, 0x72, 0x4f, 0x55, 0x0c, 0xd5, 0x2d,
			0x96, 0xf3, 0xe4, 0x12, 0x37, 0xa2, 0x07, 0x44, 0x43, 0xa5, 0x43, 0x3a, 0x41, 0x33, 0x0d, 0xca,
			0x41, 0xb7, 0x3a, 0x57, 0x17, 0x65, 0xef, 0x3d, 0xbb, 0xd1, 0x04, 0x5d, 0xb7, 0xf8, 0x71, 0x39,
			0xff, 0x82, 0x01, 0x19, 0x75, 0xaf, 0x8a, 0x01, 0x8a, 0x0a, 0x26, 0x93, 0x85, 0xfd, 0x04, 0x5f,
		},
	},
	{
		"16 blocks",
		[]byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		[]byte{
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
			0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
			0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
			0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
			0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
		},
		[]byte{
			0xbc, 0x71, 0x0d, 0x76, 0x2d, 0x07, 0x0b, 0x26, 0x36, 0x1d, 0xa8, 0x2b, 0x54, 0x56, 0x5e, 0x46,
			0xb0, 0x2b, 0x3d, 0xbd, 0xdd, 0x50, 0xd5, 0xb4, 0x58, 0xae, 0xcc, 0xb2, 0x5d, 0xa1, 0x05, 0xe1,
			0x6a, 0xd7, 0x0b, 0xc0, 0x11, 0x75, 0xad, 0x43, 0xb0, 0x80, 0x6a, 0x2e, 0x7b, 0x9c, 0xa5, 0x45,
			0x60, 0x24, 0x59, 0xa0, 0x6b, 0x7d, 0x13, 0x0d, 0xde, 0x42, 0xa3, 0xe0, 0x47, 0x68, 0x18, 0xd2,
			0x00, 0xb8, 0x33, 0x1a, 0x66, 0x57, 0xd6, 0xbe, 0xb8, 0x5b, 0x72, 0x4f, 0x55, 0x0c, 0xd5, 0x2d,
			0x96, 0xf3, 0xe4, 0x12, 0x37, 0xa2, 0x07, 0x44, 0x43, 0xa5, 0x43, 0x3a, 0x41, 0x33, 0x0d, 0xca,
			0x41, 0xb7, 0x3a, 0x57, 0x17, 0x65, 0xef, 0x3d, 0xbb, 0xd1, 0x04, 0x5d, 0xb7, 0xf8, 0x71, 0x39,
			0xff, 0x82, 0x01, 0x19, 0x75, 0xaf, 0x8a, 0x01, 0x8a, 0x0a, 0x26, 0x93, 0x85, 0xfd, 0x04, 0x5f,
			0x73, 0xd4, 0xc6, 0x3a, 0x81, 0x0a, 0x91, 0xe3, 0xb9, 0x17, 0x89, 0xdf, 0x4c, 0xcd, 0xe8, 0xe3,
			0x4e, 0xe7, 0x8d, 0x52, 0x89, 0x93, 0xb9, 0xef, 0x42, 0xe7, 0x5d, 0x67, 0xa8, 0x25, 0xad, 0xf0,
			0xe2, 0x45, 0x9d, 0x8c, 0x30, 0x61, 0x8a, 0x26, 0x90, 0x4f, 0x52, 0x61, 0xa0, 0x61, 0x62, 0xfb,
			0x36, 0xc8, 0x95, 0xe2, 0x8d, 0x75, 0x86, 0xf5, 0xbf, 0x22, 0x1c, 0xdd, 0xc9, 0x52, 0x71, 0x5a,
			0x7e, 0xb0, 0x56, 0xd6, 0x8a, 0x7e, 0xfa, 0x4f, 0xda, 0x6b, 0x97, 0x95, 0x23, 0xa7, 0xa8, 0x39,
			0x76, 0x31, 0x10, 0x79, 0x47, 0x98, 0x5b, 0x71, 0xbf, 0xc9, 0x4c, 0xce, 0xb7, 0xd4, 0x19, 0x86,
			0x04, 0x87, 0xc0, 0xba, 0xe8, 0xa5, 0x4c, 0xc8, 0x48, 0x9c, 0x28, 0xd3, 0x4b, 0x4d, 0xfc, 0x3f,
			0x9b, 0xbc, 0xf3, 0xd1, 0x9d, 0x25, 0x43, 0x47, 0x37, 0xea, 0xb7, 0x5e, 0x0d, 0xdb, 0x58, 0xf1,
		},
	},
}

func TestCTR_SM4(t *testing.T) {
	for _, tt := range ctrSM4Tests {
		test := tt.name

		c, err := sm4.NewCipher(tt.key)
		if err != nil {
			t.Errorf("%s: NewCipher(%d bytes) = %s", test, len(tt.key), err)
			continue
		}

		for j := 0; j <= 5; j += 5 {
			in := tt.in[0 : len(tt.in)-j]
			ctr := cipher.NewCTR(c, tt.iv)
			encrypted := make([]byte, len(in))
			ctr.XORKeyStream(encrypted, in)
			if out := tt.out[0:len(in)]; !bytes.Equal(out, encrypted) {
				t.Errorf("%s/%d: CTR\ninpt %x\nhave %x\nwant %x", test, len(in), in, encrypted, out)
			}
		}

		for j := 0; j <= 7; j += 7 {
			in := tt.out[0 : len(tt.out)-j]
			ctr := cipher.NewCTR(c, tt.iv)
			plain := make([]byte, len(in))
			ctr.XORKeyStream(plain, in)
			if out := tt.in[0:len(in)]; !bytes.Equal(out, plain) {
				t.Errorf("%s/%d: CTRReader\nhave %x\nwant %x", test, len(out), plain, out)
			}
		}

		if t.Failed() {
			break
		}
	}
}

func TestCTRStream(t *testing.T) {
	t.Run("SM4", func(t *testing.T) {
		rng := newRandReader(t)

		key := make([]byte, 16)
		rng.Read(key)

		block, err := sm4.NewCipher(key)
		if err != nil {
			panic(err)
		}

		cryptotest.TestStreamFromBlock(t, block, cipher.NewCTR)
	})
}
