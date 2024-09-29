package encrypt

import (
	"crypto/aes"
	"fmt"
	"log"
	"strings"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/crypto"
	"github.com/deatil/go-cryptobin/cryptobin/crypto/mode/wrap"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "encrypt",
		Short: "aes/sm4 encryption/decryption",
	}

	root.AddCommand(c, &subCmd{})
}

type subCmd struct {
	Input   string `short:"i" help:"Input string, or filename"`
	Key     string `short:"k" help:"public key pem file" env:"auto"`
	IV      string `help:"IV, Nonce for GCM" env:"auto"`
	Out     string `short:"o" help:"output file name"`
	Decrypt bool   `short:"d" help:"decrypt"`

	Mode       string `help:"mode" default:"GCM" enum:"CBC,GCM,CCM,CTR,ECB,CFB,OFB,WRAP" env:"auto"`
	Additional string `help:"additional in GCM/CCM"`

	SM4     bool `help:"sm4"`
	Verbose bool `short:"v" help:"verbose"`

	Base64 bool `help:"base64 encrypted output"`
	Hex    bool `help:"hex encrypted output"`
}

func (f *subCmd) Run(_ *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	if r.SourceType == gterm.SourceRandom {
		log.Printf("random input: %s:base64", ss.Base64().EncodeBytes(data).V1.Bytes())
	}

	if f.Key == "" {
		f.Key = string(ss.Rand().Bytes(16)) // 128位
		log.Printf("rand --key %x:hex", f.Key)
	} else {
		key, err := gterm.DecodeByTailTag(f.Key)
		if err != nil {
			log.Printf("decode key error: %v", err)
			return nil
		}
		f.Key = string(key)
	}
	if f.IV == "" {
		f.IV = string(ss.Rand().Bytes(ss.If(strings.EqualFold(f.Mode, "GCM"), 12, aes.BlockSize)))
		log.Printf("rand --iv %x:hex", f.IV)
	} else {
		iv, err := gterm.DecodeByTailTag(f.IV)
		if err != nil {
			log.Printf("decode iv error: %v", err)
			return nil
		}
		f.IV = string(iv)
	}

	obj := crypto.
		FromBytes(data).
		SetKey(f.Key).
		SetIv(f.IV)
	action := ""
	if f.SM4 {
		obj = obj.SM4()
		action += "SM4"
	} else {
		obj = obj.Aes()
		action += "AES"
	}

	switch strings.ToUpper(f.Mode) {
	case "GCM":
		if f.Additional != "" {
			obj = obj.GCM([]byte(f.Additional)).NoPadding()
		} else {
			obj = obj.GCM().NoPadding()
		}

		action += "/GCM/NoPadding"
	case "CCM":
		if f.Additional != "" {
			obj = obj.CCM([]byte(f.Additional)).NoPadding()
		} else {
			obj = obj.CCM().NoPadding()
		}
		action += "/CCM/NoPadding"
	case "CTR":
		obj = obj.CTR().NoPadding()
		action += "/CTR/NoPadding"
	case "ECB":
		obj = obj.ECB().PKCS7Padding()
		action += "/ECB/PKCS7Padding"
	case "CFB":
		obj = obj.CFB().NoPadding()
		action += "/CFB/NoPadding"
	case "OFB":
		obj = obj.OFB().NoPadding()
		action += "/OFB/NoPadding"
	case "WRAP":
		obj = obj.ModeBy(wrap.Wrap)
		action += "/WRAP/PKCS7Padding"
	default:
		obj = obj.CBC().PKCS7Padding()
		action += "/CBC/PKCS7Padding"
	}

	obj = ss.If(f.Decrypt, obj.Decrypt, obj.Encrypt)()
	action += ss.If(f.Decrypt, " Decrypt", " Encrypt")

	if err := obj.Error(); err != nil {
		log.Printf("%s error: %v", action, err)
		return err
	}

	result := lo.
		If(f.Base64, obj.ToBase64String).
		ElseIf(f.Hex, obj.ToHexString).
		Else(obj.ToString)()

	return WriteDataFile(f.Out, []byte(result), !f.Decrypt)
}
