package ggtrand

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	rd "github.com/Pallinder/go-randomdata"
	"github.com/aidarkhanov/nanoid/v2"
	"github.com/bingoohuang/ngg/ggtrand/pkg/art"
	"github.com/bingoohuang/ngg/ggtrand/pkg/c7a"
	"github.com/bingoohuang/ngg/ggtrand/pkg/cid"
	"github.com/bingoohuang/ngg/ggtrand/pkg/colors"
	"github.com/bingoohuang/ngg/ggtrand/pkg/genpw"
	"github.com/bingoohuang/ngg/ggtrand/pkg/hash"
	"github.com/bingoohuang/ngg/ggtrand/pkg/img"
	"github.com/bingoohuang/ngg/ggtrand/pkg/ksid"
	"github.com/bingoohuang/ngg/ggtrand/pkg/objectid"
	"github.com/bingoohuang/ngg/ggtrand/pkg/snow2"
	"github.com/bingoohuang/ngg/ggtrand/pkg/str"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/bwmarrin/snowflake"
	"github.com/chilts/sid"
	oid "github.com/coolbed/mgo-oid"
	"github.com/dromara/dongle/base58"
	"github.com/google/uuid"
	"github.com/jxskiss/base62"
	"github.com/kjk/betterguid"
	pwe "github.com/kuking/go-pwentropy"
	"github.com/lithammer/shortuuid"
	"github.com/manifoldco/promptui"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	guuid "github.com/satori/go.uuid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
	"github.com/spaolacci/murmur3" // Fast, fully fledged murmur3 in Go 版本. https://github.com/twmb/murmur3
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/sqids/sqids-go"
	"github.com/vishal-bihani/go-tsid"
)

func (f *Cmd) initFlags(p *pflag.FlagSet) {
	p.StringVarP(&dir, "dir", "d", "", "")
	p.StringVarP(&tag, "tag", "t", "", "tag, use help to see all tags")
	p.StringVarP(&input, "input", "i", "", "")
	p.IntVarP(&num, "num", "n", 1, "")
	p.CountVarP(&verbose, "verbose", "v", "")
	p.IntVarP(&argLen, "len", "l", 100, "")
	p.BoolVarP(&raw, "raw", "r", false, "raw format(like use int64 instead of string)")
}

func Run() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

var cmd = func() *cobra.Command {
	r := &cobra.Command{
		Use:   "ggtrand",
		Short: "generate randoms",
	}

	r.Version = "version"
	r.SetVersionTemplate(ver.Version() + "\n")

	fc := &Cmd{}
	r.Run = func(cmd *cobra.Command, args []string) {
		if err := fc.run(); err != nil {
			fmt.Println(err)
		}
	}
	fc.initFlags(r.Flags())

	return r
}()

type Cmd struct {
}

func (f *Cmd) run() error {
	p := createPrinter()
	runRandoms(p)

	if strings.EqualFold(tag, "HELP") {
		for i, t := range allTags {
			fmt.Printf("%d: %s\n", i+1, t)
		}
		return nil
	}

	log.SetOutput(os.Stdout)

	if dir != "" {
		img.Dir = dir
	}

	if tag != "" {
		return nil
	}

	if p := prompt(); p != nil {
		runRandoms(p)
	}

	return nil
}

func prompt() func(name string, f func(int) any) {
	prompt := promptui.Select{
		Label: "Select One of the Randoms",
		Items: allTags,
		Searcher: func(input string, index int) bool {
			return ss.ContainsFold(allTags[index], input)
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil
	}

	return printRandom(func(name string) bool { return name == result })
}

func runRandoms(p func(name string, f func(int) any)) {
	p("file", func(int) any {
		file, err := GenerateRandomTextFile(argLen)
		if err != nil {
			return err.Error()
		}
		return file
	})
	p("blake3hash-zeebo", createFileFunc(hash.Blake3Zeebo))
	p("blake3hash-luke", createFileFunc(hash.Blake3Luke))
	p("xxhash", createFileFunc(hash.XXH64File))
	p("md5-hash", createFileFunc(hash.MD5HashFile))
	p("sha256-hash", createFileFunc(func(f string) ([]byte, error) { return hash.HashFile(f, sha256.New()) }))
	p("murmur3-32-hash", createFileFunc(func(f string) ([]byte, error) { return hash.HashFile(f, murmur3.New32()) }))
	p("murmur3-64-hash", createFileFunc(func(f string) ([]byte, error) { return hash.HashFile(f, murmur3.New64()) }))
	p("murmur3-128-hash", createFileFunc(func(f string) ([]byte, error) { return hash.HashFile(f, murmur3.New128()) }))
	p("imo-hash", createFileFunc(hash.IMOHashFile))

	p("Base64Std", func(int) any { return base64.StdEncoding.EncodeToString(randToken()) })
	p("Base64RawStd", func(int) any { return base64.RawStdEncoding.EncodeToString(randToken()) })
	p("Base64RawURL", func(int) any { return base64.RawURLEncoding.EncodeToString(randToken()) })
	p("Base64URL", func(int) any { return base64.URLEncoding.EncodeToString(randToken()) })
	p("SillyName", wrap(rd.SillyName))
	p("Email", wrap(rd.Email))
	p("IP v4", wrap(rd.IpV4Address))
	p("IP v6", wrap(rd.IpV6Address))
	p("UserAgent", wrap(rd.UserAgentString))
	p("Pwd", func(int) any {
		return string(genpw.Gen(genpw.WithMinCountOfNumbers(1), genpw.WithMinCountOfSymbols(1)))
	})
	p("Password", func(int) any { return pwe.PwGen(pwe.FormatComplex, pwe.Strength96) })
	p("Password Easy", func(int) any { return pwe.PwGen(pwe.FormatEasy, pwe.Strength96) })
	p("FirstName", func(int) any { return gofakeit.FirstName() })
	p("Name", func(int) any { return gofakeit.Name() })
	p("Numbers", func(int) any { return gofakeit.DigitN(uint(argLen)) })
	p("Letters", func(int) any { return gofakeit.LetterN(uint(argLen)) })
	jj.DefaultRandOptions.Pretty = false
	p("JSON", func(int) any { return string(jj.Rand()) })
	p("String", func(int) any { return str.RandStr(argLen) })

	p("Captcha Direct", func(int) any {
		cp := c7a.NewCaptcha(150, 40, 5)
		cp.SetMode(c7a.DirectString)
		code, pImg := cp.OutPut()
		return code + " " + img.ToPng(pImg, false)
	})

	p("Captcha SimpleMathFormula", func(int) any {
		cp := c7a.NewCaptcha(150, 40, 5)
		cp.SetMode(c7a.SimpleMathFormula)
		code, pImg := cp.OutPut()
		return code + " " + img.ToPng(pImg, false)
	})

	p("sony/sonyflake", func(int) any {
		flake := sonyflake.NewSonyflake(sonyflake.Settings{})
		v, err := flake.NextID()
		if err != nil {
			log.Fatalf("flake.NextID() failed with %s\n", err)
		}
		// Note: this is base16, could shorten by encoding as base62 string
		return base62.EncodeToString(base62.FormatUint(v))
	})
	p("max", func(int) any {
		return []string{
			fmt.Sprintf("\n int64: %d (len: %d), uint64: %d (len: %d)",
				math.MaxInt64, len(fmt.Sprintf("%+v", math.MaxInt64)),
				uint64(math.MaxUint64), len(fmt.Sprintf("%+v", uint64(math.MaxUint64)))),

			fmt.Sprintf("\n int32: %d (len: %d), uint32: %d (len: %d)",
				math.MaxInt32, len(fmt.Sprintf("%+v", math.MaxInt32)),
				math.MaxUint32, len(fmt.Sprintf("%+v", math.MaxUint32))),

			fmt.Sprintf("\n int16: %d (len: %d), uint16: %d (len: %d)",
				math.MaxInt16, len(fmt.Sprintf("%+v", math.MaxInt16)),
				math.MaxUint16, len(fmt.Sprintf("%+v", math.MaxUint16))),

			fmt.Sprintf("\n int8: %d (len: %d), uint8: %d (len: %d)",
				math.MaxInt8, len(fmt.Sprintf("%+v", math.MaxInt8)),
				math.MaxUint8, len(fmt.Sprintf("%+v", math.MaxUint8))),

			fmt.Sprintf("\n float64: %f, float32 %f", math.MaxFloat64, math.MaxFloat32),
		}
	})
	p("oklog/ulid", func(int) any {
		v := ulid.MustNew(ulid.Now(), rand.Reader)
		return []string{v.String(), "base32固定26位，48位时间(ms)+80位随机"}
	})
	p("chilts/sid", func(int) any { return []string{sid.IdBase64(), "32位时间(ns)+64位随机"} })
	p("kjk/betterguid", func(int) any { return []string{betterguid.New(), "32位时间(ms)+72位随机"} })
	p("segmentio/ksuid", func(int) any {
		return []string{ksuid.New().String(), "32位时间(s)+128位随机，20字节，base62固定27位，优选"}
	})

	p("ksid base64", func(int) any {
		t := time.Now()
		a := ksid.New(ksid.WithValue(ksid.Nil), ksid.WithTime(t)).String()
		b := ksid.New(ksid.WithTime(t)).String()
		c := ksid.New(ksid.WithValue(ksid.Max), ksid.WithTime(t)).String()
		return []string{a, b, c, fmt.Sprintf("a<=b: %v", a <= b), fmt.Sprintf("b<=c: %v", b <= c)}
	})
	p("google/uuid v4", func(int) any { return []string{uuid.New().String(), "128位随机"} })
	p("lithammer/shortuuid v4 base57", func(int) any {
		return []string{shortuuid.New(), "concise, unambiguous, URL-safe UUID"}
	})
	p("lithammer/shortuuid v5 base57", func(int) any {
		return []string{shortuuid.NewWithNamespace("http://example.com"), "shortuuid UUID v5"}
	})
	p("lithammer/shortuuid v4 ", func(int) any {
		enc := base58Encoder{}
		return []string{shortuuid.NewWithEncoder(enc), "shortuuid UUID base58"}
	})
	p("satori/go.uuid v4", func(int) any {
		v4 := guuid.NewV4()
		return []string{v4.String(), "UUIDv4 from RFC 4112 for comparison"}
	})
	p("aidarkhanov/nanoid/v2", func(int) any { return PickStr(nanoid.New()) })  // "i25_rX9zwDdDn7Sg-ZoaH"
	p("matoous/go-nanoid/v2", func(int) any { return PickStr(gonanoid.New()) }) // "i25_rX9zwDdDn7Sg-ZoaH"
	p("coolbed/mgo-oid Mongodb Object ID", func(int) any {
		v := oid.NewOID()
		return []string{v.String(), fmt.Sprintf("Timestamp: %d", v.Timestamp())}
	})
	p("rs/xid Mongo Object ID", func(int) any {
		v := xid.New()
		m := base64.RawURLEncoding.EncodeToString(v.Machine())
		return []string{v.String(), fmt.Sprintf(
			"32 位Time: %s, 24位Machine: %s, Pid: %d, , Counter: %d 4B time(s) + 3B machine id + 2B pid + 3Brandom",
			v.Time(), m, v.Pid(), v.Counter())}
	})
	p("BSON Object ID", func(int) any { return objectid.NewObjectId().String() })
	n, _ := snowflake.NewNode(1)
	p("snowflake ID", func(int) any {
		v := n.Generate()
		return []string{v.String(), fmt.Sprintf("41位 Time: %d, 10位 Node: %d, 12位 Step:%d", v.Time(), v.Node(), v.Step())}
	})

	p("Random ID with fixed length 12", func(int) any { return fmt.Sprintf("%d", cid.Cid12()) })
	p("customized snowflake ID with fixed length 12", func(int) any {
		return fmt.Sprintf("%d", snow2.Node12.Next())
	})
	p("customized snowflake ID with uint32", func(int) any {
		return fmt.Sprintf("%d", snow2.NodeUint32.Next())
	})
	p("vishal-bihani/go-tsid", func(int) any {
		ts := tsid.Fast()
		result := ""
		if raw {
			result = fmt.Sprintf("%d", ts.ToNumber())
		} else {
			result = ts.ToString()
		}
		return []string{result, "按生成时间排序;64位整数/13位字符串; URL安全,不区分大小写,没有连字符; 比UUID、ULID 和 KSUID 短"}
	})

	p("姓名", wrap(ss.Rand().ChineseName))
	p("性别", wrap(ss.Rand().Sex))
	p("地址", wrap(ss.Rand().Address))
	p("手机", wrap(ss.Rand().Mobile))
	p("身份证", wrap(ss.Rand().ChinaID))
	p("有效期", wrap(ss.Rand().ValidPeriod))
	p("发证机关", wrap(ss.Rand().IssueOrg))
	p("邮箱", wrap(ss.Rand().Email))
	p("银行卡", wrap(ss.Rand().BankNo))
	p("日期", func(int) any { return RandDate().Format("2006年01月02日") })

	arts(p)
	p("Image", img.RandomImage)

	p("Random Colors", colors.RandomColors)
	p("Random Palettes", colors.RandomPalettes)
	p("Color Blend", colors.ColorBlend)

	p("pbe", pbeEncrptDealer)
	p("ebp", pbeDecrptDealer)
	p("Sqids", sqidsDealer)
}

// RandDate 返回随机时间，时间区间从 1970 年 ~ 2020 年
func RandDate() time.Time {
	begin, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	end, _ := time.Parse("2006-01-02 15:04:05", "2020-01-01 00:00:00")
	return RandDateRange(begin, end)
}

// RandDateRange 返回随机时间，时间区间从 1970 年 ~ 2020 年
func RandDateRange(from, to time.Time) time.Time {
	return time.Unix(ss.Rand().Int64Between(from.Unix(), to.Unix()), 0)
}

type base58Encoder struct{}

func (enc base58Encoder) Encode(u uuid.UUID) string {
	return string(base58.Encode(u[:]))
}

func (enc base58Encoder) Decode(s string) (uuid.UUID, error) {
	return uuid.FromBytes(base58.Decode([]byte(s)))
}

func pbeEncrptDealer(int) any {
	if tag == "ALL" {
		return nil
	}

	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}

	plain := promptui.Prompt{
		Label:    "Plain",
		Validate: validate,
		Default:  pwe.PwGen(pwe.FormatComplex, pwe.Strength96),
	}

	plainResult, err := plain.Run()
	if err != nil {
		return err.Error()
	}

	passwd := promptui.Prompt{
		Label:    "Password",
		Validate: validate,
		Mask:     '*',
	}
	passwdResult, err := passwd.Run()
	if err != nil {
		return err.Error()
	}

	result, err := ss.PbeEncrypt(plainResult, passwdResult, 19)
	if err != nil {
		return err.Error()
	}

	return result
}

var epochTime, _ = time.Parse("2006-01-02 15:04:05", "2023-09-22 14:01:59")

func sqidsDealer(int) any {
	var values []uint64
	if input != "" {
		nums := strings.Split(input, ",")
		for _, n := range nums {
			if n == "" {
				continue
			}
			p, err := strconv.ParseUint(n, 10, 64)
			if err != nil {
				log.Fatalf("%s is not a valid uint", n)
			}
			values = append(values, p)
		}
	} else {
		values = []uint64{uint64(time.Since(epochTime).Milliseconds()), RandUint64()}
	}

	s, _ := sqids.New()
	id, _ := s.Encode(values)
	if verbose > 0 {
		originalValues := s.Decode(id) // [1, 2, 3]
		if !slices.Equal(values, originalValues) {
			log.Printf("before %v", values)
			log.Printf("decode %v", originalValues)
		}
	}

	return id
}

func RandUint64() uint64 {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf) // Always succeeds, no need to check error
	return binary.LittleEndian.Uint64(buf)
}

func pbeDecrptDealer(int) any {
	if tag == "ALL" {
		return nil
	}

	validate := func(input string) error {
		if len(input) < 6 {
			return errors.New("password must have more than 6 characters")
		}
		return nil
	}

	plain := promptui.Prompt{
		Label:    "PBE Encrypted TEXT",
		Validate: validate,
		Default:  pwe.PwGen(pwe.FormatComplex, pwe.Strength96),
	}

	plainResult, err := plain.Run()
	if err != nil {
		return err.Error()
	}

	passwd := promptui.Prompt{
		Label:    "Password",
		Validate: validate,
		Mask:     '*',
	}
	passwdResult, err := passwd.Run()
	if err != nil {
		return err.Error()
	}

	result, err := ss.PbeDecrypt(plainResult, passwdResult, 19)
	if err != nil {
		return err.Error()
	}

	return result
}

func randToken() []byte {
	token := make([]byte, argLen)
	n, err := rand.Read(token)
	if err != nil {
		panic(err)
	}

	return token[:n]
}

func PickStr(s string, _ any) string {
	return s
}

func wrap(f func() string) func(int) any { return func(int) any { return f() } }

var allTags []string

func createPrinter() func(name string, f func(int) any) {
	tag = strings.ToUpper(tag)
	if tag == "" || tag == "HELP" {
		allTags = []string{}
		return func(name string, f func(int) any) {
			allTags = append(allTags, name)
		}
	}

	var okFn func(string) bool

	if tag == "ALL" {
		okFn = func(string) bool { return true }
	} else {
		okFn = func(name string) bool { return strings.Contains(strings.ToUpper(name), tag) }
	}

	return printRandom(okFn)
}

func printRandom(okFn func(string) bool) func(name string, f func(int) any) {
	return func(name string, f func(int) any) {
		if okFn(name) {
			start0 := time.Now()
			for i := 0; i < num; i++ {
				start := time.Now()
				v := f(i)
				if verbose == 0 {
					if v2, ok := v.([]string); ok && len(v2) > 0 {
						fmt.Printf("%s\n", v2[0])
					} else if v1, ok := v.(string); ok {
						fmt.Printf("%s\n", v1)
					} else {
						fmt.Printf("%v\n", v)
					}
					continue
				}

				if v2, ok := v.([]string); ok && len(v2) > 0 {
					log.Printf("%s: %s (len: %d) %s, cost %s", name, v2[0], len(v2[0]), v2[1:], time.Since(start))
				} else if v1, ok := v.(string); ok {
					log.Printf("%s: %s (len: %d), cost %s", name, v1, len(v1), time.Since(start))
				} else {
					log.Printf("%s: %v , cost %s", name, v, time.Since(start))
				}
			}

			if verbose > 0 {
				log.Printf("Completed, cost %s", time.Since(start0))
			}
		}
	}
}

func arts(p1 func(name string, f func(int) any)) {
	p1("Generative art", func(i int) any {
		item := artMaps[i%(len(artMaps))]
		result := item.Fn()
		return item.Name + ": " + result
	})
}

var (
	tag     string
	input   string
	num     int
	verbose int
	argLen  int
	dir     string
	raw     bool
)

type artMap struct {
	Fn   func() string
	Name string
}

var artMaps = []artMap{
	{Name: "Junas", Fn: art.Junas},
	{Name: "Random Shapes", Fn: art.RandomShapes},
	{Name: "Color Circle2", Fn: art.ColorCircle2},
	{Name: "Circle Grid", Fn: art.CircleGrid},
	{Name: "Circle Composes Circle", Fn: art.CircleComposesCircle},
	{Name: "Pixel Hole", Fn: art.PixelHole},
	{Name: "Dots Wave", Fn: art.DotsWave},
	{Name: "Contour Line", Fn: art.ContourLine},
	{Name: "Noise Line", Fn: art.NoiseLine},
	{Name: "Dot Line", Fn: art.DotLine},
	{Name: "Ocean Fish", Fn: art.OceanFish},
	{Name: "Circle Loop", Fn: art.CircleLoop},
	{Name: "Domain Warp", Fn: art.DomainWarp},
	{Name: "Circle Noise", Fn: art.CircleNoise},
	{Name: "Perlin Perls", Fn: art.PerlinPerls},
	{Name: "Color Canve", Fn: art.ColorCanve},
	{Name: "Julia Set", Fn: art.JuliaSet},
	{Name: "Black Hole", Fn: art.BlackHole},
	{Name: "Silk Sky", Fn: art.SilkSky},
	{Name: "Circle Move", Fn: art.CircleMove},
	{Name: "Random Circle", Fn: art.RandomCircle},
}

func printInspect(id ksuid.KSUID) string {
	return fmt.Sprintf(`(Time: %v, Timestamp: %v, Payload: %v) `,
		id.Time(), id.Timestamp(), strings.ToUpper(hex.EncodeToString(id.Payload())))
}

func createFileFunc(f func(string) ([]byte, error)) func(int) any {
	return func(int) any {
		file := input
		stat, err := os.Stat(file)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("failed to stat %s error: %v", file, err)
				return nil
			}

			if temp, err := os.CreateTemp("", ""); err != nil {
				log.Printf("failed to create temporary file: %v", err)
				return nil
			} else {
				_, _ = io.CopyN(temp, rand.Reader, 10*1024*1024)
				_ = temp.Close()
				file = temp.Name()
			}
		} else if stat.IsDir() {
			log.Printf("%s is not allowed to be a directory", file)
			return nil
		}

		d, err := f(file)
		if err != nil {
			log.Printf("error creating hash: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(d)
	}
}
