package ss

import (
	"crypto/cipher"
	"crypto/des" // nolint
	"crypto/md5" // nolint
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/howeyc/gopass"
	"github.com/spf13/viper"
)

// Pbe configs the passphrase.
type Pbe struct {
	Passphrase string
}

// Encrypt encrypts p by PBEWithMD5AndDES with 19 iterations.
// it will prompt password if viper get none.
func (c Pbe) Encrypt(p string) (string, error) {
	pwd := c.Passphrase
	if pwd == "" {
		pwd = GetPbePwd()
	}

	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	encrypt, err := PbeEncrypt(p, pwd, iterations)
	if err != nil {
		return "", err
	}

	return pbePrefix + encrypt, nil
}

// Decrypt decrypts p by PBEWithMD5AndDES with 19 iterations.
func (c Pbe) Decrypt(p string) (string, error) {
	if !strings.HasPrefix(p, pbePrefix) {
		return p, nil
	}

	pwd := c.Passphrase
	if pwd == "" {
		pwd = GetPbePwd()
	}

	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	return PbeDecrypt(p[len(pbePrefix):], pwd, iterations)
}

var pbeRe = regexp.MustCompile(`\{PBE\}[\w_-]+`) // nolint

// ChangePbe changes the {PBE}xxx to {PBE} yyy with a new passphase
func (c Pbe) Change(s, newPassphrase string) (string, error) {
	var err error

	m := make(map[string]string)
	f := func(old string) string {
		if v, ok := m[old]; ok {
			return v
		}

		raw := ""
		raw, err = c.Decrypt(old)

		if err != nil {
			return ""
		}

		newPwd := ""
		newPwd, err = Pbe{Passphrase: newPassphrase}.Encrypt(raw)

		if err != nil {
			return ""
		}

		m[old] = newPwd
		return newPwd
	}

	return pbeRe.ReplaceAllStringFunc(s, f), err
}

// Decode free the {PBE}xxx to yyy with a  passphrase
func (c Pbe) Decode(s string) (string, error) {
	var err error

	m := make(map[string]string)

	f := func(old string) string {
		if v, ok := m[old]; ok {
			return v
		}

		raw := ""
		raw, err = c.Decrypt(old)

		if err != nil {
			return ""
		}

		m[old] = raw
		return raw
	}

	return pbeRe.ReplaceAllStringFunc(s, f), err
}

// Encode will PBE encrypt the passwords in the text
// passwords should be as any of following format and its converted pattern
// 1. {PWD:clear} -> {PBE:cyphered}
// 2. [PWD:clear] -> {PBE:cyphered}
// 3. (PWD:clear) -> {PBE:cyphered}
// 4. "PWD:clear" -> "{PBE:cyphered}"
// 5.  PWD:clear  ->  {PBE:cyphered}
func (c Pbe) Encode(s string) (string, error) {
	m := make(map[string]string)

	pbed := ""
	src := s

	var err error

	for {
		pos := strings.Index(src, "PWD:")
		if pos <= 0 {
			pbed += src
			break
		}

		left := src[pos-1]
		if Alphanumeric(left) {
			pbed += "PWD:"
			src = src[4:]

			continue
		}

		expectRight := string(getRight(left))
		rpos := strings.Index(src[pos+4:], expectRight)

		if rpos < 0 {
			pbed += src[0:4]
			src = src[4:]

			continue
		}

		raw := src[pos+4 : pos+rpos+4]

		pwd := ""
		if v, ok := m[raw]; ok {
			pwd = v
		} else {
			if pwd, err = c.Encrypt(raw); err != nil {
				return "", err
			}

			m[raw] = pwd
		}

		switch left {
		case '(', '{', '[':
		default:
			pwd = string(left) + pwd + expectRight
		}

		pbed += src[0:pos-1] + pwd
		src = src[pos+rpos+5:]
	}

	return pbed, nil
}

func getRight(left uint8) uint8 {
	switch left {
	case '(':
		return ')'
	case '{':
		return '}'
	case '[':
		return ']'
	default:
		return left
	}
}

// Alphanumeric tells u is letter or digit char.
func Alphanumeric(u uint8) bool {
	return u >= '0' && u <= '9' || u >= 'a' && u <= 'z' || u >= 'A' && u <= 'Z'
}

var pbePwdOnce sync.Once // nolint
var pbePwd string        // nolint

// GetPbePwd read pbe password from viper, or from stdin.
func GetPbePwd() string {
	pbePwdOnce.Do(readInternal)

	return pbePwd
}

// PbePwd defines the keyword for client flag.
const PbePwd = "pbepwd"

func readInternal() {
	pbePwd = viper.GetString(PbePwd)
	if pbePwd != "" {
		return
	}

	fmt.Printf("PBE Password: ")

	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetPasswd error %v", err)
		os.Exit(1) // nolint gomnd
	}

	pbePwd = string(pass)
}

const iterations = 19
const pbePrefix = `{PBE}`

// PbeEncode encrypts p by PBEWithMD5AndDES with 19 iterations.
// it will prompt password if viper get none.
func PbeEncode(p string) (string, error) {
	pwd := GetPbePwd()
	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	encrypt, err := PbeEncrypt(p, pwd, iterations)
	if err != nil {
		return "", err
	}

	return pbePrefix + encrypt, nil
}

// PbeDecode decrypts p by PBEWithMD5AndDES with 19 iterations.
func PbeDecode(p string) (string, error) {
	if !strings.HasPrefix(p, pbePrefix) {
		return p, nil
	}

	pwd := GetPbePwd()
	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	return PbeDecrypt(p[len(pbePrefix):], pwd, iterations)
}

func isFilenameArg(args []string) (string, bool) {
	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		filename := args[0][1:]
		filename = ExpandHome(filename)

		if stat, err := os.Stat(filename); err == nil && !stat.IsDir() {
			return filename, true
		}
	}

	return "", false
}

// PbePrintEncrypt prints the PBE encryption.
func PbePrintEncrypt(passStr string, plains ...string) {
	if filename, yes := isFilenameArg(plains); yes {
		PbeEncryptFile(filename, passStr)

		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "#\tPlain\tEncrypted")

	for i, p := range plains {
		pbed, err := PbeEncrypt(p, passStr, iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pbe.Encrypt error %v", err)
			os.Exit(1) // nolint gomnd
		}
		fmt.Fprintf(w, "%d\t%q\t%q\n", i+1, p, pbePrefix+pbed)
	}

	w.Flush()
}

// PbePrintDecrypt prints the PBE decryption.
func PbePrintDecrypt(passStr string, cipherText ...string) {
	if filename, yes := isFilenameArg(cipherText); yes {
		PbeEncryptFile(filename, passStr)

		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "#\tEncrypted\tPlain")

	for i, ebp := range cipherText {
		ebpx := strings.TrimPrefix(ebp, pbePrefix)

		p, err := PbeDecrypt(ebpx, passStr, iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pbe.Decrypt error %v", err)
			os.Exit(1) // nolint gomnd
		}

		fmt.Fprintf(w, "%d\t%q\t%q\n", i+1, ebp, p)
	}

	w.Flush()
}

func PbeEncryptFile(filename, passStr string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	text, err := Pbe{Passphrase: passStr}.Encode(string(file))
	if err != nil {
		panic(err)
	}

	ft, _ := os.Stat(filename)

	if err := os.WriteFile(filename, []byte(text), ft.Mode()); err != nil {
		panic(err)
	}
}

func PbeEncryptFileUpdate(filename, passStr, pbenew string) {
	filename = ExpandHome(filename)

	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	text, err := Pbe{Passphrase: passStr}.Change(string(file), pbenew)
	if err != nil {
		panic(err)
	}

	ft, _ := os.Stat(filename)

	if err := os.WriteFile(filename, []byte(text), ft.Mode()); err != nil {
		panic(err)
	}
}

func PbeDecryptFile(filename, passStr string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	text, err := Pbe{Passphrase: passStr}.Decode(string(file))
	if err != nil {
		panic(err)
	}

	ft, _ := os.Stat(filename)

	if err := os.WriteFile(filename, []byte(text), ft.Mode()); err != nil {
		panic(err)
	}
}

// PbeEncrypt PrintEncrypt the plainText based on password and iterations with random salt.
// The result contains the first 8 bytes salt before BASE64.
func PbeEncrypt(plainText, password string, iterations int) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	encText, err := pbeDoEncrypt(plainText, password, salt, iterations)
	if err != nil {
		return "", err
	}

	return Base64().EncodeBytes(append(salt, encText...), Url).V1.String(), nil
}

// PbeDecrypt PrintDecrypt the cipherText(result of Encrypt) based on password and iterations.
func PbeDecrypt(cipherText, password string, iterations int) (string, error) {
	p := Base64().Decode(cipherText)
	if p.V2 != nil {
		return "", p.V2
	}

	msgData := p.V1.Bytes()
	salt := msgData[:8]
	encText := msgData[8:]

	return pbeDoDecrypt(encText, password, salt, iterations)
}

// PbeEncryptSalt PrintEncrypt the plainText based on password and iterations with fixed salt.
func PbeEncryptSalt(plainText, password, fixedSalt string, iterations int) (string, error) {
	salt := make([]byte, 8)
	copy(salt, fixedSalt)

	encText, err := pbeDoEncrypt(plainText, password, salt, iterations)
	if err != nil {
		return "", err
	}

	return Base64().EncodeBytes(encText, Url).V1.String(), nil
}

// PbeDecryptSalt PrintDecrypt the cipherText(result of EncryptSalt) based on password and iterations.
func PbeDecryptSalt(cipherText, password, fixedSalt string, iterations int) (string, error) {
	p := Base64().Decode(cipherText)
	if p.V2 != nil {
		return "", p.V2
	}

	salt := make([]byte, 8)
	copy(salt, fixedSalt)

	return pbeDoDecrypt(p.V1.Bytes(), password, salt, iterations)
}

func pbeDoEncrypt(plainText, password string, salt []byte, iterations int) ([]byte, error) {
	padNum := byte(8 - len(plainText)%8) // nolint gomnd
	for i := byte(0); i < padNum; i++ {
		plainText += string(padNum)
	}

	dk, iv := bpeGetDerivedKey(password, string(salt), iterations)
	block, err := des.NewCipher(dk) // nolint

	if err != nil {
		return nil, err
	}

	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plainText))
	encrypter.CryptBlocks(encrypted, []byte(plainText))

	return encrypted, nil
}

func pbeDoDecrypt(encText []byte, password string, salt []byte, iterations int) (string, error) {
	dk, iv := bpeGetDerivedKey(password, string(salt), iterations)
	block, err := des.NewCipher(dk) // nolint

	if err != nil {
		return "", err
	}

	decrypter := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encText))
	decrypter.CryptBlocks(decrypted, encText)

	decryptedString := strings.TrimRight(string(decrypted), "\x01\x02\x03\x04\x05\x06\x07\x08")
	return decryptedString, nil
}

func bpeGetDerivedKey(password, salt string, iterations int) ([]byte, []byte) {
	key := md5.Sum([]byte(password + salt)) // nolint

	for i := 0; i < iterations-1; i++ {
		key = md5.Sum(key[:]) // nolint
	}

	return key[:8], key[8:]
}
