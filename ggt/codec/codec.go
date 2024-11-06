package main

import (
	"bytes"
	"crypto/hmac"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/cespare/xxhash/v2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/golang-module/dongle"
	"github.com/spf13/cobra"
	"github.com/zeebo/blake3"
)

func main() {
	c := root.CreateCmd(nil, "codec", "hash, baseXx, and etc.", &codec{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type codec struct {
	From string   `short:"f" help:"from" enum:"string,hex,base32,base45,base58,base62,base64,base85,base91,base100,safeURL"`
	To   []string `short:"t" help:"to" enum:"string,hex,base32,base45,base58,base62,base64,base85,base91,base100,safeURL,md2,md4,md5,sha1,sha3-224,sha3-256,sha3-384,sha3-512,sha224,sha256,sha384,sha512,sha512-224,sha512-256,shake128-256,shake128-512,shake256,ripemd160,blake2b-256,blake2b-384,blake2b-512,blake2s-256,blake3,sm3,xxhash"`

	Input string `short:"i" help:"Input string, or filename"`
	Key   string `short:"k" env:"auto" help:"HMAC key"`
	Raw   bool   `short:"r" help:"print raw bytes"`
}

func (f *codec) Run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true, TryDecode: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	key, err := gterm.DecodeByTailTag(f.Key)
	if err != nil {
		log.Printf("decode key error: %v", err)
		return nil
	}
	f.Key = string(key)

	var d []byte
	var dd dongle.Decoder
	switch algo := strings.ToLower(f.From); algo {
	case "string", "":
		d = data
	case "hex":
		dd = dongle.Decode.FromBytes(data).ByHex()
		d = dd.ToBytes()
	case "base32":
		dd = dongle.Decode.FromBytes(data).ByBase32()
		d = dd.ToBytes()
	case "base45":
		dd = dongle.Decode.FromBytes(data).ByBase45()
		d = dd.ToBytes()
	case "base58":
		dd = dongle.Decode.FromBytes(data).ByBase58()
		d = dd.ToBytes()
	case "base85":
		dd = dongle.Decode.FromBytes(data).ByBase85()
		d = dd.ToBytes()
	case "base64":
		p := ss.Base64().Decode(string(data))
		dd.Error = p.V2
		if p.V2 == nil {
			d = p.V1.Bytes()
		}
	case "base91":
		dd = dongle.Decode.FromBytes(data).ByBase91()
		d = dd.ToBytes()
	case "base100":
		dd = dongle.Decode.FromBytes(data).ByBase100()
		d = dd.ToBytes()
	case "safeURL":
		dd = dongle.Decode.FromBytes(data).BySafeURL()
		d = dd.ToBytes()
	}

	if dd.Error != nil {
		return dd.Error
	}

	if len(f.To) == 0 {
		rawString := string(d)
		if f.Raw {
			fmt.Printf("%s", []byte(rawString))
		} else {
			log.Printf("raw: %s (len: %d)", rawString, len(rawString))
		}
	}
	for _, to := range f.To {
		var rawString, hexString, base64String string
		algo := strings.ToLower(to)
		switch algo {
		case "string", "":
			rawString = string(d)
		case "hex":
			rawString = dongle.Encode.FromBytes(d).ByHex().String()
		case "base32":
			rawString = dongle.Encode.FromBytes(d).ByBase32().String()
		case "base45":
			rawString = dongle.Encode.FromBytes(d).ByBase45().String()
		case "base58":
			rawString = dongle.Encode.FromBytes(d).ByBase58().String()
		case "base64":
			rawString = dongle.Encode.FromBytes(d).ByBase64().String()
		case "base64url":
			rawString = dongle.Encode.FromBytes(d).ByBase64URL().String()
		case "base91":
			rawString = dongle.Encode.FromBytes(d).ByBase91().String()
		case "base100":
			rawString = dongle.Encode.FromBytes(d).ByBase100().String()
		case "safeURL":
			rawString = dongle.Encode.FromBytes(d).BySafeURL().String()
		case "md2":
			e := dongle.Encrypt.FromBytes(d).ByMd2()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "md4":
			e := dongle.Encrypt.FromBytes(d).ByMd4()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "md5":
			e := dongle.Encrypt.FromBytes(d).ByMd5()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha1":
			e := dongle.Encrypt.FromBytes(d).BySha1()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha3-224":
			e := dongle.Encrypt.FromBytes(d).BySha3(224)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha3-256":
			e := dongle.Encrypt.FromBytes(d).BySha3(256)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha3-384":
			e := dongle.Encrypt.FromBytes(d).BySha3(384)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha3-512":
			e := dongle.Encrypt.FromBytes(d).BySha3(512)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha224":
			e := dongle.Encrypt.FromBytes(d).BySha224()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha256":
			e := dongle.Encrypt.FromBytes(d).BySha256()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha384":
			e := dongle.Encrypt.FromBytes(d).BySha384()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha512":
			e := dongle.Encrypt.FromBytes(d).BySha512()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha512-224":
			e := dongle.Encrypt.FromBytes(d).BySha512(224)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "sha512-256":
			e := dongle.Encrypt.FromBytes(d).BySha512(256)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "shake128-256":
			e := dongle.Encrypt.FromBytes(d).ByShake128(256)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "shake128-512":
			e := dongle.Encrypt.FromBytes(d).ByShake128(512)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "ripemd160":
			e := dongle.Encrypt.FromBytes(d).ByRipemd160()
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "blake2b-256":
			e := dongle.Encrypt.FromBytes(d).ByBlake2b(256)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "blake2b-384":
			e := dongle.Encrypt.FromBytes(d).ByBlake2b(384)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "blake2b-512":
			e := dongle.Encrypt.FromBytes(d).ByBlake2b(512)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()
		case "blake2s-256":
			e := dongle.Encrypt.FromBytes(d).ByBlake2s(256)
			rawString, hexString, base64String = e.String(), e.ToHexString(), e.ToBase64String()

		case "blake3", "xxhash", "sm3":
			hf := func() hash.Hash {
				switch algo {
				case "blake3":
					return blake3.New()
				case "xxhash":
					return xxhash.New()
				case "sm3":
					return sm3.New()
				}
				return nil
			}
			var h hash.Hash

			if f.Key != "" {
				h = hmac.New(hf, []byte(f.Key))
			} else {
				h = hf()
			}
			if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
				return fmt.Errorf("copy: %w", err)
			}

			if sum64, ok := h.(interface {
				Sum64() uint64
			}); ok {
				s := sum64.Sum64()
				log.Printf("sum64: %d, hex: %s", s, strconv.FormatUint(s, 16))
			}
			e := h.Sum(nil)
			rawString, hexString, base64String = string(e), fmt.Sprintf("%x", e), ss.Base64().EncodeBytes(e).V1.String()
		}

		if f.Raw {
			fmt.Printf("%s", []byte(rawString))
		} else {
			log.Printf("%s raw: %s (len: %d)", algo, rawString, len(rawString))
			if hexString != "" {
				log.Printf("%s hex: %s (len: %d)", algo, hexString, len(hexString))
			}
			if base64String != "" {
				log.Printf("%s base64: %s (len: %d)", algo, base64String, len(base64String))
			}
		}
	}
	return err
}

func ToBase64RawStd(s []byte) []byte {
	s = bytes.TrimSpace(s)
	s = bytes.TrimRight(s, "=")
	// // the standard encoding with - and _ substituted for + and /.
	s = bytes.ReplaceAll(s, []byte("-"), []byte("+"))
	return bytes.ReplaceAll(s, []byte("_"), []byte("/"))
}
