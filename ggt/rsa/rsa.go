package rsa

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bingoohuang/ngg/ggt/encrypt"
	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/rsa"
	"github.com/spf13/cobra"
)

// https://github.com/deatil/go-cryptobin/blob/main/docs/rsa.md

func init() {
	c := &cobra.Command{
		Use: "rsa",
	}
	root.AddCommand(c, nil)

	root.CreateSubCmd(c, "newkey", "生成公钥私钥", &newKeyCmd{})
	root.CreateSubCmd(c, "sign", "签名", &signCmd{})
	root.CreateSubCmd(c, "verify", "验签", &verifyCmd{})
	root.CreateSubCmd(c, "encrypt", "公钥加密/私钥解密, 私钥加密/公钥解密", &encryptCmd{})
	root.CreateSubCmd(c, "keypair", "公检测私钥公钥是否匹配", &checkKeyPairCmd{})
}

type checkKeyPairCmd struct {
	Pri  string `short:"p" help:"private key pem file"`
	Pass string `help:"private key password"`

	Pub string `short:"P" help:"public key pem file"`
}

func (f *checkKeyPairCmd) Run(cmd *cobra.Command, args []string) error {
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

func (f *encryptCmd) Run(cmd *cobra.Command, args []string) error {
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
		}

		if err := obj.Error(); err != nil {
			return err
		}
		if err := encrypt.WriteDataFile(f.Out, obj.ToBytes(), !f.Decrypt); err != nil {
			return err
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
		} else {
			// 私钥加密
			obj = obj.PrivateKeyEncrypt()
		}

		if err := obj.Error(); err != nil {
			return err
		}

		if err := encrypt.WriteDataFile(f.Out, obj.ToBytes(), !f.Decrypt); err != nil {
			return err
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

func (f *verifyCmd) Run(cmd *cobra.Command, args []string) error {
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

func (f *signCmd) Run(cmd *cobra.Command, args []string) error {
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

type newKeyCmd struct {
	Bits  int    `help:"keys bits" default:"2048" enum:"512,1024,2048,4096"`
	Pass  string `short:"p" help:"privatekey password"`
	Dir   string `help:"output dir"`
	Pkcs8 bool   `help:"pkcs8"`
}

func (r *newKeyCmd) Run(cmd *cobra.Command, args []string) error {
	obj := rsa.New().GenerateKey(r.Bits)

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

	encrypt.WriteKeyFile(pri.ToKeyBytes(), r.Dir, "rsa_pri.pem")

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
	pub = ss.If(r.Pkcs8, obj.CreatePKCS8PublicKey, obj.CreatePKCS1PublicKey)()
	// CreatePKCS8PublicKey().
	// CreateXMLPublicKey().

	encrypt.WriteKeyFile(pub.ToKeyBytes(), r.Dir, "rsa_pub.pem")

	return nil
}
