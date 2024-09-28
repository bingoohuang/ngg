package codec

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/golang-module/dongle"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use: "codec",
	}
	root.AddCommand(c, &codec{})
}

type codec struct {
	From string `short:"f" help:"from" enum:"string,hex,base32,base45,base58,base62,base64,base85,base91,base100,safeURL"`
	To   string `short:"t" help:"to" enum:"string,hex,base32,base45,base58,base62,base64,base85,base91,base100,safeURL,md2,md4,md5,sha1,sha3-224,sha3-256,sha3-384,sha3-512,sha224,sha256,sha384,sha512,sha512-224,sha512-256,shake128-256,shake128-512,shake256,ripemd160,blake2b-256,blake2b-384,blake2b-512,blake2s-256"`

	Input string `short:"i" help:"Input string, or filename"`
	Raw   bool   `short:"r" help:"print raw bytes"`
}

func (f codec) Run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	var d []byte
	switch strings.ToLower(f.From) {
	case "string", "":
		d = data
	case "hex":
		d = dongle.Decode.FromBytes(data).ByHex().ToBytes()
	case "base32":
		d = dongle.Decode.FromBytes(data).ByBase32().ToBytes()
	case "base45":
		d = dongle.Decode.FromBytes(data).ByBase45().ToBytes()
	case "base58":
		d = dongle.Decode.FromBytes(data).ByBase58().ToBytes()
	case "base85":
		d = dongle.Decode.FromBytes(data).ByBase85().ToBytes()
	case "base64":
		d = dongle.Decode.FromBytes(convertBase64ToRawStd(data)).ByBase64().ToBytes()
	case "base91":
		d = dongle.Decode.FromBytes(data).ByBase91().ToBytes()
	case "base100":
		d = dongle.Decode.FromBytes(data).ByBase100().ToBytes()
	case "safeURL":
		d = dongle.Decode.FromBytes(data).BySafeURL().ToBytes()
	}

	var rawString, hexString, base64String string
	switch strings.ToLower(f.To) {
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
	}

	if f.Raw {
		fmt.Println([]byte(rawString))
		fmt.Printf("%08b\n", []byte(rawString))
	} else {
		fmt.Println("raw:", rawString)
	}
	if hexString != "" {
		fmt.Println("hex:", hexString)
	}
	if base64String != "" {
		fmt.Println("base64:", base64String)
	}
	return err
}

func convertBase64ToRawStd(s []byte) []byte {
	s = bytes.TrimSpace(s)
	s = bytes.TrimRight(s, "=")
	// // the standard encoding with - and _ substituted for + and /.
	s = bytes.ReplaceAll(s, []byte("-"), []byte("+"))
	return bytes.ReplaceAll(s, []byte("_"), []byte("/"))
}
