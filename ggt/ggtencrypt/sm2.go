package ggtencrypt

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/sm2"
	gmsm2 "github.com/deatil/go-cryptobin/gm/sm2"
	"github.com/spf13/cobra"
)

func initSm2(rootCmd *cobra.Command) {
	c := &cobra.Command{
		Use:   "sm2",
		Short: "sm2 公钥私钥生成、签名验签、加解密",
	}

	rootCmd.AddCommand(c)
	root.CreateCmd(c, "newkey", "生成公钥私钥", &sm2KeyCmd{})
	root.CreateCmd(c, "sign", "签名", &sm2SignCmd{})
	root.CreateCmd(c, "verify", "验签", &sm2VerifyCmd{})
	root.CreateCmd(c, "encrypt", "加密", &sm2EncryptCmd{})
	root.CreateCmd(c, "decrypt", "解密", &sm2DecryptCmd{})
	root.CreateCmd(c, "inspect", "私钥公钥证书解析 / Parse PrivateKey or PublicKey/获取 x, y, d 16进制数据", &sm2InspectCmd{})
	root.CreateCmd(c, "recover", "SM2 用 x, y 生成公钥，用 d 生成私钥 / use x,y to make public key and use d to make private key", &sm2RecoverCmd{})
	root.CreateCmd(c, "convert", "私钥证书编码格式转换 / Change PrivateKey type", &sm2ConvertCmd{})
}

type sm2KeyCmd struct {
	Pass string `help:"private key password"`
	Dir  string `help:"output dir"`
}

func (f *sm2KeyCmd) Run(_ *cobra.Command, _ []string) error {
	obj := sm2.New().GenerateKey()

	pri := gmsm2.PrivateKeyTo(obj.GetPrivateKey())
	pub := gmsm2.PublicKeyTo(obj.GetPublicKey())

	log.Printf("private key (Base64): %s", ss.Base64().EncodeBytes(pri).V1.Bytes())
	log.Printf("public key  (Base64): %s", ss.Base64().EncodeBytes(pub).V1.Bytes())
	log.Printf("private key (Hex): %s", hex.EncodeToString(pri))
	log.Printf("public key  (Hex): %s", hex.EncodeToString(pub))

	if f.Pass != "" {
		obj = obj.CreatePrivateKeyWithPassword(f.Pass)
	} else {
		obj = obj.CreatePrivateKey()
	}

	if err := WriteKeyFile(obj.ToKeyBytes(), f.Dir, "sm2_pri.pem"); err != nil {
		return nil
	}

	obj = obj.CreatePublicKey()
	if err := WriteKeyFile(obj.ToKeyBytes(), f.Dir, "sm2_pub.pem"); err != nil {
		return nil
	}
	return nil
}

type sm2SignCmd struct {
	Input string `short:"i" help:"Input string, or filename"`
	Key   string `short:"k" help:"private key pem file"`
	Pass  string `help:"private key password"`
	Uid   string `default:"1234567812345678" help:"uid data"`
	Hash  string `default:"SM3" help:"sign hash"`
}

func ParsePublic(obj sm2.SM2, key string) (sm2.SM2, error) {
	keyResult := ss.Base64().Decode(key)
	if keyResult.V2 == nil {
		obj = obj.FromPublicKeyBytes(keyResult.V1.Bytes())
		return obj, obj.Error()
	}

	keyPem, err := os.ReadFile(ss.ExpandHome(key))
	if err != nil {
		return obj, err
	}

	obj = obj.FromPublicKey(keyPem)
	return obj, obj.Error()
}

func ParsePrivate(obj sm2.SM2, key, pass string) (sm2.SM2, error) {
	keyResult := ss.Base64().Decode(key)
	if keyResult.V2 == nil {
		obj = obj.FromPrivateKeyBytes(keyResult.V1.Bytes())
		return obj, obj.Error()
	}

	keyPem, err := os.ReadFile(ss.ExpandHome(key))
	if err != nil {
		return obj, err
	}

	if pass != "" {
		obj = obj.FromPrivateKeyWithPassword(keyPem, pass)
	} else {
		obj = obj.FromPrivateKey(keyPem)
	}

	return obj, obj.Error()
}

func (f *sm2SignCmd) Run(_ *cobra.Command, _ []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()
	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	obj := sm2.New()

	// 私钥签名
	// private key sign data
	// 比如: SM2withSM3 => ... SetSignHash("SM3").Sign() ...

	obj = obj.FromBytes(data)
	obj, err = ParsePrivate(obj, f.Key, f.Pass)
	if err != nil {
		return err
	}

	if f.Uid != "" {
		uid, err := gterm.DecodeBySchema(f.Uid)
		if err != nil {
			return err
		}
		obj = obj.WithUID(uid)
	}

	// FromPrivateKeyWithPassword([]byte(priKeyPem), psssword).
	// FromPKCS1PrivateKey([]byte(priKeyPem)).
	// FromPKCS1PrivateKeyWithPassword([]byte(priKeyPem), psssword).
	// FromPKCS8PrivateKey([]byte(priKeyPem)).
	// FromPKCS8PrivateKeyWithPassword([]byte(priKeyPem), psssword).
	// WithUID(uid).
	// SetSignHash("SM3").
	// WithSignHash(hash).
	// Sign().
	// SignASN1().
	// SignBytes().
	// ToBase64String()

	obj = obj.Sign() // SM2 签名时，默认的 Hash 就是 SM3
	if err := obj.Error(); err != nil {
		return err
	}

	fmt.Printf("%s\n", obj.ToBase64String())
	return nil
}

type sm2VerifyCmd struct {
	Input string `short:"i" help:"Input string, or filename"`
	Key   string `short:"K" help:"public key pem file"`
	Pass  string `help:"private key password"`
	Sign  string `help:"signed base64 string to be verified"`
	Uid   string `default:"1234567812345678" help:"uid data"`
	Hash  string `default:"SM3" help:"sign hash"`
}

func (f *sm2VerifyCmd) Run(_ *cobra.Command, _ []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	// 公钥验证
	// public key verify signed data
	obj := sm2.New().FromBase64String(f.Sign)
	obj, err = ParsePublic(obj, f.Key)
	if err != nil {
		return err
	}

	if f.Uid != "" {
		uid, err := gterm.DecodeBySchema(f.Uid)
		if err != nil {
			return err
		}
		obj = obj.WithUID(uid)
	}

	obj = obj.SetSignHash(f.Hash)

	// WithUID(uid).
	// SetSignHash("SM3").
	// WithSignHash(hash).
	obj = obj.Verify(data)

	if err := obj.Error(); err != nil {
		return err
	}

	// VerifyASN1([]byte(data)).
	// VerifyBytes([]byte(data)).
	fmt.Printf("%t\n", obj.ToVerify())

	return nil
}

type sm2EncryptCmd struct {
	Input string `short:"i" help:"Input string, or filename"`
	Key   string `short:"K" help:"public key pem file"`
	Mode  string `help:"mode C1C3C2/C1C2C3" default:"C1C3C2"`
	Out   string `short:"o" help:"output file name"`
	Uid   string `default:"1234567812345678" help:"uid data"`
}

func (f *sm2EncryptCmd) Run(_ *cobra.Command, _ []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()
	// 待加密数据
	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	// 加密解密 - 公钥加密/私钥解密 / Encrypt with public key
	// https://github.com/deatil/go-cryptobin/blob/main/docs/sm2.md

	// 公钥加密
	// public key Encrypt data
	obj := sm2.New().
		FromBytes(data).
		WithUID([]byte(f.Uid))

	obj, err = ParsePublic(obj, f.Key)
	if err != nil {
		return err
	}

	// SetMode 为可选，默认为 C1C3C2
	obj = obj.SetMode(f.Mode). // C1C3C2 | C1C2C3
					Encrypt()

	if err := obj.Error(); err != nil {
		return err
	}

	if err := WriteDataFile("SM2 encrypt", f.Out, obj.ToBytes(), true); err != nil {
		return err
	}

	return nil
}

type sm2DecryptCmd struct {
	Input string `short:"i" help:"Input string, or filename"`
	Key   string `short:"k" help:"private key pem file"`
	Pass  string `help:"private key password"`
	Mode  string `help:"mode C1C3C2/C1C2C3" default:"C1C3C2"`
	Uid   string `default:"1234567812345678" help:"uid data"`
	Out   string `short:"o" help:"output file name"`
}

func (f *sm2DecryptCmd) Run(_ *cobra.Command, _ []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()
	// 待加密数据
	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	// 私钥解密

	obj := sm2.New().
		FromBytes(data).
		WithUID([]byte(f.Uid))

	obj, err = ParsePrivate(obj, f.Key, f.Pass)
	if err != nil {
		return err
	}

	obj = obj.SetMode(f.Mode) // C1C3C2 | C1C2C3
	// FromPrivateKeyWithPassword([]byte(priKeyPem), psssword).
	// FromPKCS1PrivateKey([]byte(priKeyPem)).
	// FromPKCS1PrivateKeyWithPassword([]byte(priKeyPem), psssword).
	// FromPKCS8PrivateKey([]byte(priKeyPem)).
	// FromPKCS8PrivateKeyWithPassword([]byte(priKeyPem), psssword).
	// SetMode 为可选，默认为 C1C3C2
	// SetMode("C1C3C2"). // C1C3C2 | C1C2C3
	obj = obj.Decrypt()

	if err := obj.Error(); err != nil {
		return err
	}

	if err := WriteDataFile("SM2 decrypt", f.Out, obj.ToBytes(), false); err != nil {
		return err
	}

	return nil
}

type sm2InspectCmd struct {
	Pri  string `short:"k" help:"private key pem file"`
	Pass string `help:"private key password"`

	Pub       string `short:"K" help:"private key pem file"`
	CheckPair bool   `short:"c" help:"check keypair 检测私钥公钥是否匹配"`
}

func (f *sm2InspectCmd) Run(_ *cobra.Command, _ []string) error {
	if f.CheckPair { // 检测私钥公钥是否匹配
		obj := sm2.New()

		var err error
		obj, err = ParsePrivate(obj, f.Pri, f.Pass)
		if err != nil {
			return err
		}

		obj, err = ParsePublic(obj, f.Pub)
		if err != nil {
			return err
		}

		log.Printf("pair checked: %t", obj.CheckKeyPair())
		return nil
	}

	// 私钥解析

	if f.Pri != "" {
		// 私钥密码
		// privatekey password
		obj := sm2.New()

		var err error
		obj, err = ParsePrivate(obj, f.Pri, f.Pass)
		if err != nil {
			return err
		}

		// var parsedPrivateKey *gmsm2.PrivateKey = obj.
		// FromPrivateKeyWithPassword(priKeyPem, psssword).
		// FromPKCS1PrivateKey(priKeyPem).
		// FromPKCS1PrivateKeyWithPassword(priKeyPem, psssword).
		// FromPKCS8PrivateKey(priKeyPem).
		// FromPKCS8PrivateKeyWithPassword(priKeyPem, psssword).
		// GetPrivateKey()

		log.Printf("private key D data: %s", obj.GetPrivateKeyDString())

		pub := obj.MakePublicKey()
		x := pub.GetPublicKeyXString()
		y := pub.GetPublicKeyYString()

		log.Printf("private key X data: %s", x)
		log.Printf("private key Y data: %s", y)
	}

	if f.Pub != "" {
		// 公钥解析
		// Parse PublicKey
		var err error
		obj := sm2.New()
		obj, err = ParsePublic(obj, f.Pub)
		if err != nil {
			return err
		}

		x := obj.GetPublicKeyXString()
		y := obj.GetPublicKeyYString()
		log.Printf("public key X data: %s", x)
		log.Printf("public key Y data: %s", y)
	}
	return nil
}

type sm2RecoverCmd struct {
	X string `short:"x" help:"公钥 X HEX"`
	Y string `short:"y" help:"公钥 Y HEX"`
	D string `short:"d" help:"私钥 D HEX"`

	Dir string `help:"output dir"`
}

func WriteKeyFile(keyBytes []byte, dir, keyFileName string) error {
	if dir != "" {
		keyFile := filepath.Join(ss.ExpandHome(dir), keyFileName)
		if err := os.WriteFile(keyFile, keyBytes, os.ModePerm); err != nil {
			return err
		}
		log.Printf("key file %s created!", keyFile)
	} else {
		log.Printf("key:\n%s", keyBytes)
	}

	return nil
}

func WriteDataFile(action, out string, data []byte, base64Console bool) error {
	if out != "" {
		if err := os.WriteFile(ss.ExpandHome(out), data, os.ModePerm); err != nil {
			return err
		}

		log.Printf("%s result written to file %s", action, out)
		return nil
	}

	if base64Console {
		log.Printf("%s result: base64://%s", action, ss.Base64().EncodeBytes(data).V1.Bytes())
	} else {
		log.Printf("%s result: %s", action, data)
	}
	return nil
}

func (f *sm2RecoverCmd) Run(_ *cobra.Command, _ []string) error {
	if f.X != "" && f.Y != "" {
		obj := sm2.New().
			FromPublicKeyXYString(f.X, f.Y)

		log.Printf("public key: base64://%s", ss.Base64().EncodeBytes(gmsm2.PublicKeyTo(obj.GetPublicKey())).V1.Bytes())

		obj = obj.CreatePublicKey()
		if err := WriteKeyFile(obj.ToKeyBytes(), f.Dir, "sm2_pub.pem"); err != nil {
			return nil
		}
	}

	if f.D != "" {
		obj := sm2.New().
			FromPrivateKeyString(f.D)

		log.Printf("private key: base64://%s", ss.Base64().EncodeBytes(gmsm2.PrivateKeyTo(obj.GetPrivateKey())).V1.Bytes())

		obj = obj.CreatePrivateKey()
		if err := WriteKeyFile(obj.ToKeyBytes(), f.Dir, "sm2_pri.pem"); err != nil {
			return nil
		}
	}

	return nil
}

type sm2ConvertCmd struct {
	Pri  string `short:"k" help:"private key pem file"`
	Pass string `help:"private key password"`

	Pkcs8     bool   `help:"convert private to PKCS8"`
	Pkcs8Pass string `help:"private PCKS8 key password"`

	Dir string `help:"output dir"`
}

func (f *sm2ConvertCmd) Run(_ *cobra.Command, _ []string) error {
	priKeyPem, err := os.ReadFile(ss.ExpandHome(f.Pri))
	if err != nil {
		return err
	}

	obj := sm2.New()

	if f.Pass != "" {
		obj = obj.FromPrivateKeyWithPassword(priKeyPem, f.Pass)
	} else {
		obj = obj.FromPrivateKey(priKeyPem)
	}

	if f.Pkcs8 {
		if f.Pkcs8Pass != "" {
			obj = obj.CreatePKCS8PrivateKeyWithPassword(f.Pkcs8Pass)
		} else {
			obj = obj.CreatePKCS8PrivateKey()
		}

		if err := WriteKeyFile(obj.ToKeyBytes(), f.Dir, "sm2_pri.pem"); err != nil {
			return nil
		}
	}

	return nil
}
