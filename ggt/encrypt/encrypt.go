package encrypt

import (
	"crypto/aes"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/crypto"
	"github.com/deatil/go-cryptobin/cryptobin/crypto/mode/wrap"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/pbkdf2"
)

func init() {
	c := &cobra.Command{
		Use:     "encrypt",
		Aliases: []string{"aes", "enc"},
		Short:   "aes/sm4 encryption/decryption",
	}

	root.AddCommand(c, &subCmd{})
}

type subCmd struct {
	Input   string `short:"i" help:"Input string, or filename"`
	Key     string `short:"k" help:"public key pem file" env:"auto"`
	Pass    string `short:"p" help:"password to generate key"`
	IV      string `help:"IV, Nonce for GCM" env:"auto"`
	Out     string `short:"o" help:"output file name"`
	Decrypt bool   `short:"d" help:"decrypt"`

	Mode       string `help:"mode" default:"GCM" enum:"CBC,GCM,CCM,CTR,ECB,CFB,OFB,WRAP" env:"auto"`
	Additional string `help:"additional in GCM/CCM"`
	Salt       string `default:"" help:"salt for pbkdf2"`
	KeyLen     int    `default:"16" enum:"16,24,32" help:"key length"`

	SM4     bool `help:"sm4"`
	Verbose bool `short:"v" help:"verbose"`
}

func (f *subCmd) Run(*cobra.Command, []string) error {
	r, err := gterm.Option{Random: true, TryDecode: true}.Open(f.Input)
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
		if f.Pass != "" {
			salt := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
			if f.Salt != "" {
				if s, err := gterm.DecodeByTailTag(f.Salt); err != nil {
					return err
				} else {
					salt = s
				}
			}
			f.Key = string(pbkdf2.Key([]byte(f.Pass), salt, 10000, f.KeyLen, sha256.New))
			log.Printf("pbkdf2 --key %x:hex, salt: %x:hex", f.Key, salt)
		} else {
			f.Key = string(ss.Rand().Bytes(16)) // 128位
			log.Printf("rand --key %x:hex", f.Key)
		}
	} else {
		key, err := gterm.DecodeByTailTag(f.Key, f.KeyLen)
		if err != nil {
			log.Printf("decode key error: %v", err)
			return nil
		}
		f.Key = string(key)
	}
	if f.IV == "" {
		ivLen := ss.If(strings.EqualFold(f.Mode, "GCM"), 12, aes.BlockSize)
		f.IV = string(ss.Rand().Bytes(ivLen))
		log.Printf("rand --iv %x:hex", f.IV)
	} else {
		iv, err := gterm.DecodeByTailTag(f.IV, 12, 16)
		if err != nil {
			log.Printf("decode iv error: %v", err)
			return nil
		}
		f.IV = string(iv)
	}

	obj := crypto.FromBytes(data).SetKey(f.Key).SetIv(f.IV)
	obj = ss.If(f.SM4, obj.SM4, obj.Aes)()
	action := ss.If(f.SM4, "SM4", "AES")

	switch strings.ToUpper(f.Mode) {
	case "GCM":
		if f.Additional != "" {
			obj = obj.GCM([]byte(f.Additional))
		} else {
			obj = obj.GCM()
		}
		obj = obj.NoPadding()
		action += "/GCM/NoPadding"
	case "CCM":
		if f.Additional != "" {
			obj = obj.CCM([]byte(f.Additional))
		} else {
			obj = obj.CCM()
		}
		obj = obj.NoPadding()
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

	return WriteDataFile(action, f.Out, obj.ToBytes(), !f.Decrypt)
}
