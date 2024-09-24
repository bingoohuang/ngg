package rsa

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/rsa"
	"github.com/spf13/cobra"
)

// https://github.com/deatil/go-cryptobin/blob/main/docs/rsa.md

func init() {
	rsaCobra := &cobra.Command{
		Use: "rsa",
	}
	root.AddCommand(rsaCobra, nil)

	keyCobra := &cobra.Command{
		Use:   "key",
		Short: "生成公钥私钥",
	}
	key := &keyCmd{}
	keyCobra.RunE = key.run
	ss.PanicErr(root.InitFlags(key, keyCobra.Flags()))
	rsaCobra.AddCommand(keyCobra)

	signCobra := &cobra.Command{
		Use:   "sign",
		Short: "签名",
	}
	sign := &signCmd{}
	signCobra.RunE = sign.run
	ss.PanicErr(root.InitFlags(sign, signCobra.Flags()))
	rsaCobra.AddCommand(signCobra)

	verifyCobra := &cobra.Command{
		Use:   "verify",
		Short: "验签",
	}
	verify := &verifyCmd{}
	verifyCobra.RunE = verify.run
	ss.PanicErr(root.InitFlags(verify, verifyCobra.Flags()))
	rsaCobra.AddCommand(verifyCobra)

	encryptCobra := &cobra.Command{
		Use:   "encrypt",
		Short: "公钥加密/私钥解密 / Encrypt with public key, 或者  私钥加密/公钥解密 / Encrypt with private key",
	}
	encrypt := &encryptCmd{}
	encryptCobra.RunE = encrypt.run
	ss.PanicErr(root.InitFlags(encrypt, encryptCobra.Flags()))
	rsaCobra.AddCommand(encryptCobra)

	checkKeyPairCobra := &cobra.Command{
		Use:   "check-keypair",
		Short: "检测私钥公钥是否匹配 / Check KeyPair",
	}
	checkKeyPair := &checkKeyPairCmd{}
	checkKeyPairCobra.RunE = checkKeyPair.run
	ss.PanicErr(root.InitFlags(checkKeyPair, checkKeyPairCobra.Flags()))
	rsaCobra.AddCommand(checkKeyPairCobra)

}

type checkKeyPairCmd struct {
	Pri  string `short:"p" help:"private key pem file"`
	Pass string `help:"private key password"`

	Pub string `short:"P" help:"public key pem file"`
}

func (f *checkKeyPairCmd) run(cmd *cobra.Command, args []string) error {
	pubKeyPem, err := os.ReadFile(ss.ExpandHome(f.Pub))
	if err != nil {
		return err
	}

	priKeyPem, err := os.ReadFile(ss.ExpandHome(f.Pri))
	if err != nil {
		return err
	}

	obj := rsa.New()
	if f.Pass != "" {
		obj = obj.FromPrivateKeyWithPassword(priKeyPem, f.Pass)
	} else {
		obj = obj.FromPrivateKey(priKeyPem)
	}

	// FromPrivateKey([]byte(prikeyPem)).
	// FromPrivateKeyWithPassword([]byte(prikeyPem), psssword).
	// FromPKCS1PrivateKey([]byte(prikeyPem)).
	// FromPKCS1PrivateKeyWithPassword([]byte(prikeyPem), psssword).
	// FromPKCS8PrivateKey([]byte(prikeyPem)).
	// FromPKCS8PrivateKeyWithPassword([]byte(prikeyPem), psssword).
	// FromPublicKey([]byte(pubkeyPem)).
	// FromPKCS1PublicKey([]byte(pubkeyPem)).
	obj = obj.FromPublicKey(pubKeyPem)
	log.Printf("check key pair result: %v", obj.CheckKeyPair())
	return nil
}

type encryptCmd struct {
	Pri       string `short:"p" help:"private key pem file"`
	Pub       string `short:"P" help:"public key pem file"`
	Pass      string `help:"private key password"`
	Input     string `short:"i" help:"Input string, or filename"`
	Hash      string `help:"sign hash" default:"SHA256"`
	Out       string `short:"o" help:"output file name"`
	Decrypt   bool   `short:"d" help:"decrypt"`
	Oaep      bool   `help:"填充模式OAEP(Optimal Asymmetric Encryption Padding), 只在公钥加密私钥解密时有用"`
	OaepHash  string `help:"OAEP hash" default:"SHA256"`
	OaepLabel string `help:"OAEP label"`
}

func (f *encryptCmd) run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	if f.Pub != "" {
		pubKeyPem, err := os.ReadFile(ss.ExpandHome(f.Pub))
		if err != nil {
			return err
		}

		obj := rsa.New().
			FromBytes(data).
			FromPublicKey([]byte(pubKeyPem))
		// FromPKCS1PublicKey([]byte(pubKeyPem)).
		// FromPKCS8PublicKey([]byte(pubKeyPem)).
		// FromXMLPublicKey([]byte(pubKeyXML)).

		if f.Decrypt {
			// 公钥解密
			obj = obj.PublicKeyDecrypt()
			if err := obj.Error(); err != nil {
				return err
			}

			if f.Out != "" {
				if err := os.WriteFile(ss.ExpandHome(f.Out), obj.ToBytes(), os.ModePerm); err != nil {
					return err
				} else {
					log.Printf("decrypted result written to file %s", f.Out)
				}
			} else {
				log.Printf("decrypted: %s", obj.ToString())
			}
		} else {
			// 公钥加密
			if f.Oaep {
				if f.OaepHash != "" {
					// SetOAEPHash("SHA256"). // OAEP 可选
					obj = obj.SetOAEPHash(f.OaepHash)
				}
				if f.OaepLabel != "" {
					// SetOAEPLabel("test-label"). // OAEP 可选
					obj = obj.SetOAEPLabel(f.OaepLabel)
				}
				obj = obj.EncryptOAEP()
			} else {
				obj = obj.Encrypt()
			}

			if err := obj.Error(); err != nil {
				return err
			}

			if f.Out != "" {
				if err := os.WriteFile(ss.ExpandHome(f.Out), obj.ToBytes(), os.ModePerm); err != nil {
					return err
				} else {
					log.Printf("encrypted result written to file %s", f.Out)
				}
			} else {
				log.Printf("encrypted: %s", obj.ToBase64String())
			}
		}
	} else if f.Pri != "" {
		priKeyPem, err := os.ReadFile(ss.ExpandHome(f.Pri))
		if err != nil {
			return err
		}
		obj := rsa.New().FromBytes(data).FromPrivateKey(priKeyPem)
		// FromPrivateKeyWithPassword([]byte(priKeyPem), psssword).
		// FromPKCS1PrivateKey([]byte(priKeyPem)).
		// FromPKCS1PrivateKeyWithPassword([]byte(priKeyPem), psssword).
		// FromPKCS8PrivateKey([]byte(priKeyPem)).
		// FromPKCS8PrivateKeyWithPassword([]byte(priKeyPem), psssword).
		// FromXMLPrivateKey([]byte(priKeyXML))

		if f.Decrypt {
			// 私钥解密

			if f.Oaep {
				if f.OaepHash != "" {
					// SetOAEPHash("SHA256"). // OAEP 可选
					obj = obj.SetOAEPHash(f.OaepHash)
				}
				if f.OaepLabel != "" {
					// SetOAEPLabel("test-label"). // OAEP 可选
					obj = obj.SetOAEPLabel(f.OaepLabel)
				}
				obj = obj.DecryptOAEP()
			} else {

				obj = obj.Decrypt()
			}

			if err := obj.Error(); err != nil {
				return err
			}

			if f.Out != "" {
				if err := os.WriteFile(ss.ExpandHome(f.Out), obj.ToBytes(), os.ModePerm); err != nil {
					return err
				} else {
					log.Printf("decrypted result written to file %s", f.Out)
				}
			} else {
				log.Printf("decrypted: %s", obj.ToBytes())
			}
		} else {
			// 私钥加密
			obj = obj.PrivateKeyEncrypt()

			if err := obj.Error(); err != nil {
				return err
			}

			if f.Out != "" {
				if err := os.WriteFile(ss.ExpandHome(f.Out), obj.ToBytes(), os.ModePerm); err != nil {
					return err
				} else {
					log.Printf("encrypted result written to file %s", f.Out)
				}
			} else {
				log.Printf("encrypted: %s", obj.ToBase64String())
			}
		}
	} else {
		return fmt.Errorf("必须指定公钥或者私钥，用于加密或者解密")
	}

	return nil
}

type verifyCmd struct {
	Key   string `short:"k" help:"public key pem file"`
	Input string `short:"i" help:"Input string, or filename"`

	Sign string `help:"signed base64 string to be verified"`
	Hash string `help:"sign hash" default:"SHA256"`
	Pss  bool   `help:"填充模式PSS(Probabilistic Signature Scheme)"`
}

func (f *verifyCmd) run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r.Reader)
	if err != nil {
		return err
	}

	keyPem, err := os.ReadFile(ss.ExpandHome(f.Key))
	if err != nil {
		return err
	}

	obj := rsa.New().
		FromBase64String(f.Sign).
		FromPublicKey(keyPem).
		// FromPKCS1PublicKey([]byte(pubKeyPem)).
		// FromPKCS8PublicKey([]byte(pubKeyPem)).
		// FromXMLPublicKey([]byte(pubKeyXML)).
		SetSignHash(f.Hash)
	if f.Pss {
		obj = obj.VerifyPSS(data)
	} else {
		obj = obj.Verify(data)
	}

	var verfied bool = obj.ToVerify()

	log.Printf("verfied: %v", verfied)
	return nil
}

type signCmd struct {
	Key   string `short:"k" help:"private key pem file"`
	Pass  string `help:"privatekey password"`
	Input string `short:"i" help:"Input string, or filename"`
	Hash  string `help:"sign hash" default:"SHA256"`
	Pss   bool   `help:"填充模式PSS(Probabilistic Signature Scheme)"`
}

func (f *signCmd) run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r.Reader)
	if err != nil {
		return err
	}

	keyPem, err := os.ReadFile(ss.ExpandHome(f.Key))
	if err != nil {
		return err
	}

	obj := rsa.New().FromBytes(data)
	if f.Pass != "" {
		obj = obj.FromPrivateKeyWithPassword(keyPem, f.Pass)
	} else {
		obj = obj.FromPrivateKey(keyPem)
	}
	obj = obj.SetSignHash(f.Hash)
	if f.Pss {
		obj = obj.SignPSS()
	} else {
		obj = obj.Sign()
	}
	signed := obj.ToBase64String()
	log.Printf("signed: %s", signed)

	return nil
}

type keyCmd struct {
	Bits  int    `help:"keys bits, 512 | 1024 | 2048 | 4096" default:"2048"`
	Pass  string `short:"p" help:"privatekey password"`
	Dir   string `help:"output dir"`
	Pkcs8 bool   `help:"pkcs8"`
}

func (r *keyCmd) run(cmd *cobra.Command, args []string) error {
	obj := rsa.New().GenerateKey(2048)

	// 生成私钥
	// create private key
	var pri rsa.RSA

	if r.Pkcs8 {
		if r.Pass != "" {
			pri = obj.CreatePKCS8PrivateKeyWithPassword(r.Pass)
		} else {
			pri = obj.CreatePKCS8PrivateKey()
		}
	} else {
		if r.Pass != "" {
			pri = obj.CreatePrivateKeyWithPassword(r.Pass)
		} else {
			pri = obj.CreatePKCS1PrivateKey()
		}

		obj.GetPublicKey()
	}

	// CreatePrivateKeyWithPassword(psssword, "AES256CBC").
	// CreatePKCS1PrivateKey().
	// CreatePKCS1PrivateKeyWithPassword(psssword, "AES256CBC").
	// CreatePKCS8PrivateKey().
	// CreatePKCS8PrivateKeyWithPassword(psssword, "AES256CBC", "SHA256").
	// CreateXMLPrivateKey().
	if r.Dir != "" {
		keyFile := filepath.Join(ss.ExpandHome(r.Dir), "pri.key")
		if err := os.WriteFile(keyFile, pri.ToKeyBytes(), os.ModePerm); err != nil {
			return err
		}
		log.Printf("key file %s created!", keyFile)
	} else {
		log.Printf("private key pem: \n%s", pri.ToKeyString())
	}
	// 自定义私钥加密类型
	// use custom encrypt options
	// var PriKeyPem string = obj.
	//     CreatePKCS8PrivateKeyWithPassword(psssword, rsa.Opts{
	//         Cipher:  rsa.GetCipherFromName("AES256CBC"),
	//         KDFOpts: rsa.ScryptOpts{
	//             CostParameter:            1 << 15,
	//             BlockSize:                8,
	//             ParallelizationParameter: 1,
	//             SaltSize:                 8,
	//         },
	//     }).
	//     ToKeyString()

	var pub rsa.RSA
	// 生成公钥
	// create public key
	if r.Pkcs8 {
		pub = obj.CreatePKCS8PublicKey()
	} else {
		pub = obj.CreatePKCS1PublicKey()
	}
	// CreatePKCS8PublicKey().
	// CreateXMLPublicKey().

	if r.Dir != "" {
		keyFile := filepath.Join(ss.ExpandHome(r.Dir), "pub.key")
		if err := os.WriteFile(keyFile, pub.ToKeyBytes(), os.ModePerm); err != nil {
			return err
		}
		log.Printf("key file %s created!", keyFile)
	} else {
		log.Printf("public key pem: \n%s", pub.ToKeyString())
	}

	return nil
}
