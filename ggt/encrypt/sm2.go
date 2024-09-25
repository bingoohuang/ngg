package encrypt

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/deatil/go-cryptobin/cryptobin/sm2"
	"github.com/spf13/cobra"
)

func init() {
	fc := &sm2Cmd{}
	c := &cobra.Command{
		Use:   "sm2",
		Short: "sm2 公钥私钥生成、签名验签、加解密",
		Run:   fc.run,
	}

	root.AddCommand(c, fc)
	root.CreateSubCmd(c, "key", "生成公钥私钥", &sm2KeyCmd{})
}

type sm2Cmd struct{}

func (f *sm2Cmd) run(_ *cobra.Command, args []string) {}

type sm2KeyCmd struct {
	Pass string `help:"private key password"`
	Dir  string `help:"output dir"`
}

func (f *sm2KeyCmd) Run(_ *cobra.Command, args []string) error {
	obj := sm2.New().GenerateKey()
	if f.Pass != "" {
		obj = obj.CreatePrivateKeyWithPassword(f.Pass)
	} else {
		obj = obj.CreatePrivateKey()
	}

	if f.Dir != "" {
		keyFile := filepath.Join(ss.ExpandHome(f.Dir), "sm2_pri.pem")
		if err := os.WriteFile(keyFile, obj.ToKeyBytes(), os.ModePerm); err != nil {
			return err
		}
		log.Printf("key file %s created!", keyFile)
	} else {
		log.Printf("private key: %s", obj.ToKeyString())
	}

	obj = obj.CreatePublicKey()
	if f.Dir != "" {
		keyFile := filepath.Join(ss.ExpandHome(f.Dir), "sm2_pub.pem")
		if err := os.WriteFile(keyFile, obj.ToKeyBytes(), os.ModePerm); err != nil {
			return err
		}
		log.Printf("key file %s created!", keyFile)
	} else {
		pubKeyPem := obj.ToKeyString()
		log.Printf("public key: %s", pubKeyPem)
	}
	return nil
}
