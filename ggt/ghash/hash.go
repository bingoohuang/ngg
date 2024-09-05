package ghash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/cespare/xxhash/v2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zeebo/blake3"

	"hash"
	"io"
	"os"
)

/*

# sm3
$ ggt sm3 -i lqlq666lqlq946
sm3 lqlq666lql... => E64FD76F4078E51DCA428323D3FADBD5D52723BBF1379184650DA5CE6002B2BF len:64

# sm3hmac
$ KEY=123 ggt sm3 -i lqlq666lqlq946
sm3 lqlq666lql... => FBB67FC936777011AA70336F0F0B6305D529A97A87D8ECA8880472CD2C30A721 len:64

*/

func init() {
	// crypto.MD4,         // import golang.org/x/crypto/md4
	// crypto.MD5,         // import crypto/md5
	// crypto.SHA1,        // import crypto/sha1
	// crypto.SHA224,      // import crypto/sha256
	// crypto.SHA256,      // import crypto/sha256
	// crypto.SHA384,      // import crypto/sha512
	// crypto.SHA512,      // import crypto/sha512
	// crypto.MD5SHA1,     // no implementation; MD5+SHA1 used for TLS RSA
	// crypto.RIPEMD160,   // import golang.org/x/crypto/ripemd160
	// crypto.SHA3_224,    // import golang.org/x/crypto/sha3
	// crypto.SHA3_256,    // import golang.org/x/crypto/sha3
	// crypto.SHA3_384,    // import golang.org/x/crypto/sha3
	// crypto.SHA3_512,    // import golang.org/x/crypto/sha3
	// crypto.SHA512_224,  // import crypto/sha512
	// crypto.SHA512_256,  // import crypto/sha512
	// crypto.BLAKE2s_256, // import golang.org/x/crypto/blake2s
	// crypto.BLAKE2b_256, // import golang.org/x/crypto/blake2b
	// crypto.BLAKE2b_384, // import golang.org/x/crypto/blake2b
	// crypto.BLAKE2b_512, // import golang.org/x/crypto/blake2b

	// Register(root.Cmd, ripemd160.New, "ripemd160") // 160bit，20字节
	// SHA-256算法输⼊报⽂的最⼤⻓度不超过2^64 bit，输⼊按512bit分组进⾏处理，产⽣的输出是⼀个256bit的报⽂摘要。
	Register(root.Cmd, sha256.New, "sha256") // 256bit，32字节
	Register(root.Cmd, sha512.New, "sha512") // 512bit，64字节
	Register(root.Cmd, sha1.New, "sha1")     // 160bit，20字节
	Register(root.Cmd, md5.New, "md5")       // 128bit，16字节
	Register(root.Cmd, sm3.New, "sm3")       // 128bit，16字节

	Register(root.Cmd, func() hash.Hash { return blake3.New() }, "blake3")
	Register(root.Cmd, func() hash.Hash { return xxhash.New() }, "xxhash")
}

func (f *Cmd) run() error {
	r, err := gterm.Option{Random: true}.Open(f.input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	h := f.Hasher()
	if f.hmacKey != "" {
		h = hmac.New(f.Hasher, []byte(f.hmacKey))
	}
	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	out := h.Sum(nil)

	if f.port {
		port := binary.BigEndian.Uint16(out[:2])
		fmt.Printf("Port: %d\n", port)
	}

	var s string
	if f.base64 {
		s = base64.StdEncoding.EncodeToString(out)
	} else {
		s = fmt.Sprintf("%X", out)
	}

	if f.hmacKey != "" {
		fmt.Printf("hmac-")
	}
	fmt.Printf("%s %s => %s len:%d\n", f.Name, r.SourceTitle, s, len(s))
	return nil
}

func (f *Cmd) initFlags(p *pflag.FlagSet) {
	p.StringVarP(&f.hmacKey, "key", "k", os.Getenv("KEY"), "Hmac Key (enable hmac), or $KEY")
	p.StringVarP(&f.input, "input", "i", "", "Input string, or filename")
	p.BoolVarP(&f.base64, "base64", "b", false, "Base64 encode the output")
	p.BoolVarP(&f.port, "port", "p", false, "First 2 byte as port number")
}

type Cmd struct {
	*root.RootCmd
	Hasher  func() hash.Hash
	hmacKey string
	input   string
	Name    string
	base64  bool
	port    bool
}

func Register(rootCmd *root.RootCmd, hasher func() hash.Hash, name string) {
	c := &cobra.Command{
		Use:   name,
		Short: name + " file or input string",
	}

	fc := &Cmd{RootCmd: rootCmd, Hasher: hasher, Name: name}
	c.Run = func(cmd *cobra.Command, args []string) {
		if err := fc.run(); err != nil {
			fmt.Println(err)
		}
	}
	fc.initFlags(c.Flags())
	rootCmd.AddCommand(c)
}
