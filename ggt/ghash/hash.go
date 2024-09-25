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
	"hash"
	"io"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/cespare/xxhash/v2"
	"github.com/deatil/go-cryptobin/hash/sm3"
	"github.com/spf13/cobra"
	"github.com/zeebo/blake3"
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

	// register(ripemd160.New, "ripemd160") // 160bit，20字节
	// SHA-256算法输⼊报⽂的最⼤⻓度不超过2^64 bit，输⼊按512bit分组进⾏处理，产⽣的输出是⼀个256bit的报⽂摘要。
	register(sha256.New, "sha256") // 256bit，32字节
	register(sha512.New, "sha512") // 512bit，64字节
	register(sha1.New, "sha1")     // 160bit，20字节
	register(md5.New, "md5")       // 128bit，16字节
	register(sm3.New, "sm3")       // 128bit，16字节

	register(func() hash.Hash { return blake3.New() }, "blake3")
	register(func() hash.Hash { return xxhash.New() }, "xxhash")
}

func register(hasher func() hash.Hash, name string) {
	fc := &subCmd{Hasher: hasher, name: name}
	c := &cobra.Command{
		Use:   name,
		Short: name + " file or input string",
		RunE:  fc.run,
	}
	root.AddCommand(c, fc)
}

func (f *subCmd) run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Input)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	h := f.Hasher()
	if f.Key != "" {
		h = hmac.New(f.Hasher, []byte(f.Key))
	}
	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	out := h.Sum(nil)

	if f.Port {
		port := binary.BigEndian.Uint16(out[:2])
		fmt.Printf("Port: %d\n", port)
	}

	var s string
	if f.Base64 {
		s = base64.StdEncoding.EncodeToString(out)
	} else {
		s = fmt.Sprintf("%x", out)
	}

	if f.Key != "" {
		fmt.Printf("hmac-")
	}
	fmt.Printf("%s %s => %s len:%d\n", f.name, r.SourceTitle, s, len(s))
	return nil
}

type subCmd struct {
	Hasher func() hash.Hash

	Key    string `help:"Hmac Key (enable hmac), or $KEY" env:"auto"`
	Input  string `short:"i" help:"Input string, or filename"`
	name   string
	Base64 bool `short:"b" help:"Base64 encode the output"`
	Port   bool `short:"p" help:"First 2 byte as port number"`
}
