package encrypt

import (
	"crypto/aes"
	"fmt"
	"log"
	"os"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/crypto"
	"github.com/deatil/go-cryptobin/cryptobin/crypto/mode/wrap"
	"github.com/spf13/cobra"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:   "encrypt",
		Short: "aes/sm4 encryption/decryption",
		RunE:  fc.run,
	}

	root.AddCommand(c, fc)
}

type subCmd struct {
	Input   string `short:"i" help:"Input string, or filename"`
	Key     string `short:"k" help:"public key pem file" env:"auto"`
	IV      string `help:"IV" env:"auto"`
	Out     string `short:"o" help:"output file name"`
	Decrypt bool   `short:"d" help:"decrypt"`
	Cbc     bool   `help:"CBC mode"`
	Gcm     bool   `help:"GCM mode" default:"true"`
	Ccm     bool   `help:"CCM mode"`
	Ctr     bool   `help:"CTR mode"`
	Ecb     bool   `help:"ECB mode"`
	CFB     bool   `help:"CFB mode"`
	OFB     bool   `help:"OFB mode"`
	Wrap    bool   `help:"WRAP mode"`

	Additional string `help:"additional in GCM/CCM"`

	SM4     bool `help:"sm4"`
	Verbose bool `short:"v" help:"verbose"`

	Base64 bool `help:"base64 encrypted output"`
	Hex    bool `help:"hex encrypted output"`
}

func (f *subCmd) run(_ *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	if f.Verbose && r.SourceType == gterm.SourceRandom {
		log.Printf("random input: %s:base64", ss.Base64().EncodeBytes(data).V1.Bytes())
	}

	if f.Key == "" {
		f.Key = string(ss.Rand().Bytes(16)) // 128位
		if f.Verbose {
			log.Printf("rand --key %x:hex", f.Key)
		}
	} else {
		key, err := gterm.DecodeByTailTag(f.Key)
		if err != nil {
			log.Printf("decode key error: %v", err)
			return nil
		}
		f.Key = string(key)
	}
	if f.IV == "" {
		f.IV = string(ss.Rand().Bytes(aes.BlockSize))
		if f.Verbose {
			log.Printf("rand --iv %x:hex", f.IV)
		}
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

	if f.Gcm {
		if f.Additional != "" {
			obj = obj.GCM([]byte(f.Additional)).NoPadding()
		} else {
			obj = obj.GCM().NoPadding()
		}

		action += "/GCM/NoPadding"
	} else if f.Ccm {
		if f.Additional != "" {
			obj = obj.CCM([]byte(f.Additional)).NoPadding()
		} else {
			obj = obj.CCM().NoPadding()
		}
		action += "/CCM/NoPadding"
	} else if f.Ctr {
		obj = obj.CTR().NoPadding()
		action += "/CTR/NoPadding"
	} else if f.Ecb {
		obj = obj.ECB().PKCS7Padding()
		action += "/ECB/PKCS7Padding"
	} else if f.CFB {
		obj = obj.CFB().NoPadding()
		action += "/CFB/NoPadding"
	} else if f.OFB {
		obj = obj.OFB().NoPadding()
		action += "/OFB/NoPadding"
	} else if f.Wrap {
		obj = obj.ModeBy(wrap.Wrap)
		action += "/WRAP/PKCS7Padding"
	} else {
		obj = obj.CBC().PKCS7Padding()
		action += "/CBC/PKCS7Padding"
	}

	if f.Decrypt {
		obj = obj.Decrypt()
		action += " Decrypt"
	} else {
		obj = obj.Encrypt()
		action += " Encrypt"
	}

	if err := obj.Error(); err != nil {
		log.Printf("%s error: %v", action, err)
		return nil
	}

	var result []byte
	if f.Base64 {
		result = []byte(obj.ToBase64String())
	} else if f.Hex {
		result = []byte(obj.ToHexString())
	} else {
		result = obj.ToBytes()
	}

	if f.Out != "" {
		if err := os.WriteFile(f.Out, result, os.ModePerm); err != nil {
			log.Printf("write file %s failed: %v", f.Out, err)
		} else {
			log.Printf("%s file %s", action, f.Out)
		}
	} else {
		if f.Verbose {
			log.Printf("%s: %s", action, result)
		} else {
			fmt.Printf("%s", result)
		}
	}

	return nil
}
