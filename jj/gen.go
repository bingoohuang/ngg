package jj

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/bingoohuang/ngg/jj/reggen"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/bingoohuang/ngg/tsid"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
)

type SubstituteFn struct {
	Fn   any
	Demo string
}

var DefaultSubstituteFns = map[string]SubstituteFn{
	"ip": {Fn: RandomIP,
		Demo: "随机IP: e.g. @ip @ip(192.0.2.0/24) @ip(v6)"},
	"random": {Fn: Random,
		Demo: "随机值: e.g. @random(男,女) @random(10) @random(5-10) @random(1,2,3)"},
	"random_int": {Fn: RandomInt,
		Demo: "随机整数: e.g. @random_int @random_int(100-999) @random_int(1,2,3)"},
	"random_bool": {Fn: func(_ string) any { return ss.Rand().Bool() },
		Demo: "随机布尔值: e.g. @random_bool"},
	"random_time": {Fn: RandomTime,
		Demo: "随机时间: e.g. @random_time @random_time(yyyy-MM-dd) @random_time(now, yyyy-MM-dd) @random_time(yyyy-MM-dd,1990-01-01,2021-06-06) @random_time(sep=# yyyy-MM-dd#1990-01-01#2021-06-06)"},
	"random_image": {Fn: RandomImage,
		Demo: "随机图片: e.g. @random_image(format=jpg size=640x320)"},
	"objectId": {Fn: func(string) any { return NewObjectID().Hex() },
		Demo: "随机对象ID: e.g. @objectId"},
	"regex": {Fn: Regex,
		Demo: "正则表达式: e.g. @regex([a-z]{10})"},
	"uuid": {Fn: func(version string) any { return NewUUID(version).String() },
		Demo: "随机UUID: e.g. @uuid @uuid(v4) @uuid(v5)"},
	"base64": {Fn: RandomBase64,
		Demo: "随机Base64: e.g. @base64(size=1000 std raw file=dir/f.png)"},
	"name": {Fn: func(_ string) any { return randomdata.SillyName() },
		Demo: "随机姓名: e.g. @name"},
	"ksuid": {Fn: func(_ string) any { v, _ := ksuid.NewRandom(); return v.String() },
		Demo: "随机Ksuid: e.g. @ksuid"},
	"tsid": {Fn: func(format string) any {
		id := tsid.Fast()
		switch format {
		case "number":
			return id.ToNumber()
		case "lower":
			return id.ToLower()
		case "bytes":
			return id.ToBytes()
		default:
			return id.ToString()
		}
	},
		Demo: "随机TSID: e.g. @tsid @tsid(number) @tsid(lower) @tsid(bytes)"},
	"汉字": {Fn: randomChinese,
		Demo: "随机汉字: e.g. @汉字 @汉字(1-10) @@汉字(3)"},
	"emoji": {Fn: randomEmoji,
		Demo: "随机emoji: e.g. @emoji @emoji(1-10) @@emoji(3)"},
	"姓名": {Fn: func(_ string) any { return ss.Rand().ChineseName() },
		Demo: "随机姓名: e.g. @姓名"},
	"性别": {Fn: func(_ string) any { return ss.Rand().Sex() },
		Demo: "随机性别, e.g. @性别"},
	"地址": {Fn: func(_ string) any { return ss.Rand().Address() },
		Demo: "随机地址: e.g. @地址"},
	"手机": {Fn: func(_ string) any { return ss.Rand().Mobile() },
		Demo: "随机手机: e.g. @手机"},
	"身份证": {Fn: func(_ string) any { return ss.Rand().ChinaID() },
		Demo: "随机身份证: e.g. @身份证"},
	"发证机关": {Fn: func(_ string) any { return ss.Rand().IssueOrg() },
		Demo: "随机发证机关: e.g. @发证机关"},
	"邮箱": {Fn: func(_ string) any { return ss.Rand().Email() },
		Demo: "随机邮箱: e.g. @邮箱"},
	"银行卡": {Fn: func(_ string) any { return ss.Rand().BankNo() },
		Demo: "随机银行卡: e.g. @银行卡"},
	"env": {Fn: func(name string) any { return os.Getenv(name) },
		Demo: "环境变量: e.g. @env(PATH)"},
	"file": {Fn: atFile,
		Demo: "读取文件: e.g. @file(path/file) @file(path/file, :bytes) @file(path/file, :hex) @file(path/file, :base64) @file(path/file, :datauri)"},
	"seq": {Fn: SeqGenerator,
		Demo: "序列生成器: e.g. @seq @seq(100)"},
	"gofakeit": {Fn: Gofakeit,
		Demo: "gofakeit.Template: e.g. @gofakeit(Dear {{LastName}}) https://github.com/brianvoe/gofakeit#templates"},
}

func RegisterSubstituteFn(name string, f SubstituteFn) {
	DefaultSubstituteFns[name] = f
}

func Gofakeit(args string) (any, error) {
	value, err := gofakeit.Template(args, nil)
	return value, err
}

// atFile
// @file(/path/file) read /path/file content as string
// @file(/path/file, :bytes) read /path/file content as []byte
// @file(/path/file, :hex) read /path/file content as hex string
// @file(/path/file, :base64) read /path/file content as base64 string
// @file(/path/file, :datauri) read /path/file content as datauri base64 string
func atFile(args string) (any, error) {
	fileArgs := strings.Split(args, ",")
	name := fileArgs[0]
	d, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", name, err)
	}

	useBytes := false
	useBase64 := false
	useHex := false
	datauri := false
	for i := 1; i < len(fileArgs); i++ {
		switch option := strings.ToLower(fileArgs[i]); option {
		case ":bytes":
			useBytes = true
		case ":hex":
			useHex = true
		case ":base64":
			useBase64 = true
		case ":datauri":
			datauri = true
		}
	}

	switch {
	case useBase64:
		return base64.StdEncoding.EncodeToString(d), nil
	case useHex:
		return hex.EncodeToString(d), nil
	case useBytes:
		return d, nil
	case datauri:
		mime := http.DetectContentType(d)
		data := base64.StdEncoding.EncodeToString(d)
		return fmt.Sprintf("data:%s;base64,%s", mime, data), nil
	default:
		return string(d), nil
	}
}

// RandomImage creates a random image.
// checked on https://codebeautify.org/base64-to-image-converter
func RandomImage(conf string) any {
	arg := struct {
		Format string
		Size   string
		Way    int
	}{}

	ParseConf(conf, &arg)

	imgFormat := ""
	switch strings.ToLower(arg.Format) {
	case ".jpg", "jpg", ".jpeg", "jpeg":
		imgFormat = ".jpg"
	default:
		imgFormat = ".png"
	}

	width, height := parseImageSize(arg.Size)
	c := ss.RandImgConfig{
		Width:      width,
		Height:     height,
		RandomText: fmt.Sprintf("%d", ss.Rand().Int()),
		FastMode:   false,
		PixelSize:  40,
	}

	var data []byte
	result := ""
	if imgFormat == ".png" {
		result += "data:image/png;base64,"
		if arg.Way == 1 {
			data = gofakeit.ImagePng(c.Width, c.Height)
		}
	} else {
		result += "data:image/jpeg;base64,"
		if arg.Way == 1 {
			data = gofakeit.ImageJpeg(c.Width, c.Height)
		}
	}

	if len(data) == 0 {
		data, _ = c.Gen(imgFormat)
	}
	result += base64.StdEncoding.EncodeToString(data)
	return result
}

func parseImageSize(val string) (width, height int) {
	width, height = 640, 320
	if val != "" {
		val = strings.ToLower(val)
		parts := strings.SplitN(val, "x", 2)
		if len(parts) == 2 {
			if v, _ := ss.Parse[int](parts[0]); v > 0 {
				width = v
			}
			if v, _ := ss.Parse[int](parts[1]); v > 0 {
				height = v
			}
		}
	}
	return width, height
}

type Substituter struct {
	raw     map[string]SubstituteFn
	gen     map[string]SubstitutionErrorFn
	genLock sync.RWMutex
}

func NewSubstituter(m map[string]SubstituteFn) *Substituter {
	return &Substituter{
		raw: m,
		gen: map[string]SubstitutionErrorFn{},
	}
}

func (r *Substituter) UsageDemos() []string {
	demo := make([]string, 0, len(r.raw))
	for _, v := range r.raw {
		demo = append(demo, v.Demo)
	}
	return demo
}

func (r *Substituter) Register(fn string, f SubstituteFn) { r.raw[fn] = f }

type Substitute interface {
	ss.Valuer
	Register(fn string, f SubstituteFn)
	UsageDemos() []string
}

type GenRun struct {
	Src           string
	Out           string
	Opens         int
	repeater      *Repeater
	BreakRepeater bool

	*GenContext
	repeaterWait bool
	Err          error
}

type GenContext struct {
	MockTimes int
	Substitute
}

func NewGenContext(s Substitute) *GenContext { return &GenContext{Substitute: s} }

func NewGen() *GenContext { return NewGenContext(NewSubstituter(DefaultSubstituteFns)) }

func (r *GenRun) walk(start, end, info int) int {
	element := r.Src[start:end]

	switch {
	case IsToken(info, TokOpen):
		r.Opens++
		if r.repeater != nil {
			src := r.Src[start:]
			r.repeater.Repeat(func(last bool) {
				p, _, err := r.GenContext.Process(src)
				if err != nil {
					r.Err = err
					return
				}
				if r.Out += p.Out; !last {
					r.Out += ","
				}
			})
			if r.Err != nil {
				return 0
			}
			r.repeater = nil
			r.BreakRepeater = true
			return -1
		}
	case IsToken(info, TokClose):
		r.Opens--
		if r.BreakRepeater {
			r.BreakRepeater = false
			return ss.If(r.Opens > 0, 1, 0)
		}
	case IsToken(info, TokString):
		s := element[1 : len(element)-1]
		switch {
		case r.repeater == nil:
			r.repeater = r.parseRepeat(s)
			if r.repeaterWait = r.repeater != nil && r.repeater.Key == ""; r.repeaterWait {
				return 1
			}

			if r.repeater != nil && r.repeater.Key != "" {
				r.Out += strconv.Quote(r.repeater.Key)
				return 1
			}

			fallthrough
		case IsToken(info, TokValue):
			if subs := ss.ParseExpr(s); subs.CountVars() > 0 {
				if r.repeater == nil {
					ret, err := r.Eval(subs, true)
					if err != nil {
						r.Err = err
						return 0
					}
					r.Out += ret
					return 1
				}

				repeatedValue := ""
				r.repeater.Repeat(func(last bool) {
					if r.repeater.Key == "" {
						ret, err := r.Eval(subs, true)
						if err != nil {
							r.Err = err
							return
						}
						if r.Out += ret; !last {
							r.Out += ","
						}
					} else {
						ret, err := r.Eval(subs, false)
						r.Err = err
						repeatedValue += ret
					}
				})
				if r.Err != nil {
					return 0
				}
				if r.repeater.Key != "" {
					r.Out += strconv.Quote(repeatedValue)
				}

				r.repeater = nil
				return -1
			} else if r.repeater != nil {
				r.repeatStr(element)
				return 1
			}
		}
	case IsToken(info, TokValue) && r.repeater != nil:
		r.repeater.Repeat(func(last bool) {
			if r.Out += element; !last {
				r.Out += ","
			}
		})
		r.repeater = nil
		return 1
	}

	if r.repeater == nil || !r.repeaterWait {
		r.Out += element
	}
	return ss.If(r.Opens > 0, 1, 0)
}

func (r *GenRun) Eval(subs ss.Subs, quote bool) (s string, err error) {
	ret, err := subs.Eval(r.Substitute)
	if err != nil {
		return "", err
	}
	if v, ok := ret.(string); ok {
		if quote {
			return strconv.Quote(v), nil
		}

		return v, nil
	}

	return ToString(ret, quote), nil
}

func ToString(value any, quote bool) string {
	var vvv string
	switch vv := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", vv)
	case float32, float64:
		return fmt.Sprintf("%f", vv)
	case bool:
		return fmt.Sprintf("%t", vv)
	case string:
		return vv
	case []byte:
		vvv = base64.StdEncoding.EncodeToString(vv)
	default:
		vvv = fmt.Sprintf("%v", value)
	}
	if quote {
		return strconv.Quote(vvv)
	}

	return vvv
}

func (r *GenRun) repeatStr(element string) {
	s := element[1 : len(element)-1]
	repeatedValue := ""
	r.repeater.Repeat(func(last bool) {
		if r.repeater.Key == "" {
			if r.Out += element; !last {
				r.Out += ","
			}
		} else {
			repeatedValue += s
		}
	})

	if r.repeater.Key != "" {
		r.Out += strconv.Quote(repeatedValue)
	}

	r.repeater = nil
}

func (r *Substituter) Value(name, params, expr string) (any, error) {
	r.genLock.RLock()
	f, ok := r.gen[name]
	r.genLock.RUnlock()

	if ok {
		return f(params)
	}

	r.genLock.Lock()
	defer r.genLock.Unlock()

	fullname := name

	if g, ok := r.raw[name]; ok {
		if gt, ok := g.Fn.(SubstitutionFnGen); ok {
			gtf := gt(params)
			f := func(args string) (any, error) {
				return gtf(args), nil
			}

			r.gen[fullname] = f
			return f(params)
		}

		if gt, ok := g.Fn.(SubstitutionErrorFnGen); ok {
			f := gt(params)
			r.gen[fullname] = f
			return f(params)
		}
		if gt, ok := g.Fn.(func(args string) func(args string) any); ok {
			gtf := gt(params)
			f := func(args string) (any, error) {
				return gtf(args), nil
			}
			r.gen[fullname] = f
			return f(params)
		}
		if gt, ok := g.Fn.(func(args string) func(args string) (any, error)); ok {
			f := gt(params)
			r.gen[fullname] = f
			return f(params)
		}
		if gt, ok := g.Fn.(SubstitutionFn); ok {
			f := func(args string) (any, error) {
				return gt(args), nil
			}
			r.gen[fullname] = f
			return f(params)
		}
		if gt, ok := g.Fn.(SubstitutionErrorFn); ok {
			r.gen[fullname] = gt
			return f(params)
		}
		if gt, ok := g.Fn.(func(args string) any); ok {
			f := func(args string) (any, error) {
				return gt(args), nil
			}
			r.gen[fullname] = f
			return f(params)
		}
		if gt, ok := g.Fn.(func(args string) (any, error)); ok {
			r.gen[fullname] = gt
			return gt(params)
		}
	}

	f = func(args string) (any, error) {
		return expr, nil
	}
	r.gen[fullname] = f
	return f(params)
}

type Repeater struct {
	Key   string
	Times int
}

func (r Repeater) Repeat(f func(last bool)) {
	for i := 0; i < r.Times; i++ {
		f(i == r.Times-1)
	}
}

func (r *GenContext) parseRepeat(s string) *Repeater {
	p := strings.Index(s, "|")
	if p < 0 {
		return nil
	}

	key, s := s[:p], s[p+1:]
	_, _, _, _, times, err := parseRandSize(s)
	if err != nil {
		return nil
	}

	n := ss.If(r.MockTimes > 0, r.MockTimes, int(times))
	return &Repeater{Key: key, Times: n}
}

func parseRandSize(s string) (ranged bool, paddingSize int, from, to, time int64, err error) {
	p := strings.Index(s, "-")
	times := int64(0)
	if p < 0 {
		if s != "0" && strings.HasPrefix(s, "0") {
			paddingSize = len(s)
		}
		if times, err = strconv.ParseInt(strings.TrimLeft(s, "0"), 10, 64); err != nil {
			return ranged, 0, 0, 0, 0, err
		}
		return ranged, paddingSize, times, times, times, nil
	}

	ranged = true

	s1 := s[:p]
	fromExpr := s1
	if s1 != "0" && strings.HasPrefix(s1, "0") {
		paddingSize = len(s1)
		fromExpr = strings.TrimLeft(s1, "0")
	}

	from, err1 := strconv.ParseInt(fromExpr, 10, 64)
	if err1 != nil {
		return ranged, 0, 0, 0, 0, err1
	}

	to, err2 := strconv.ParseInt(s[p+1:], 10, 64)
	if err2 != nil {
		return ranged, 0, 0, 0, 0, err2
	}
	times = ss.Rand().Int64Between(from, to)
	return ranged, paddingSize, from, to, times, nil
}

type (
	SubstitutionFn         func(args string) any
	SubstitutionErrorFn    func(args string) (any, error)
	SubstitutionFnGen      func(args string) func(args string) any
	SubstitutionErrorFnGen func(args string) func(args string) (any, error)
)

func (r *GenContext) RegisterFn(fn string, f SubstituteFn) { r.Substitute.Register(fn, f) }

var DefaultGen = NewGen()

func Gen(src string) (string, error) { return DefaultGen.Gen(src) }

func (r *GenContext) Gen(src string) (string, error) {
	p, _, err := r.Process(src)
	if err != nil {
		return "", err
	}

	return p.Out, err
}

func (r *GenContext) Process(src string) (*GenRun, int, error) {
	gr := &GenRun{Src: src, GenContext: r}
	ret := StreamParse([]byte(src), gr.walk)
	return gr, ret, gr.Err
}

func ParseParams(params string) []string {
	params = strings.TrimSpace(params)
	sep := ","
	if strings.HasPrefix(params, "sep=") {
		if idx := strings.Index(params, " "); idx > 0 {
			sep = params[4:idx]
			params = params[idx+1:]
		}
	}

	return ss.Split(params, sep)
}

func RandomTime(args string) any {
	t := ss.Rand().Time()
	if args == "" {
		return t.Format(time.RFC3339Nano)
	}

	pp := ParseParams(args)
	if v, found := filter(pp, "now"); found {
		t = time.Now()
		pp = v
	}

	layout := tick.ToLayout(pp[0])
	if len(pp) == 1 {
		return t.Format(layout)
	}

	if len(pp) == 3 {
		from, err := time.ParseInLocation(layout, pp[1], time.Local)
		if err != nil {
			log.Printf("failed to parse %s by layout %s, error:%v", pp[1], pp[0], err)
			return t.Format(time.RFC3339Nano)
		}
		to, err := time.ParseInLocation(layout, pp[2], time.Local)
		if err != nil {
			log.Printf("failed to parse %s by layout %s, error:%v", pp[2], pp[0], err)
			return t.Format(time.RFC3339Nano)
		}

		fromUnix := from.Unix()
		toUnix := to.Unix()
		r := ss.Rand().Int64Between(fromUnix, toUnix)
		return time.Unix(r, 0).Format(layout)
	}

	return t.Format(time.RFC3339Nano)
}

func filter(pp []string, s string) (filtered []string, found bool) {
	filtered = make([]string, 0, len(pp))
	for _, p := range pp {
		if p == s {
			found = true
		} else {
			filtered = append(filtered, p)
		}
	}
	return
}

var SeqStart = ss.Pick1(ss.Getenv[uint64]("SEQ", 0))

func SeqGenerator(args string) func(args string) any {
	if args == "" {
		return func(args string) any {
			return atomic.AddUint64(&SeqStart, 1)
		}
	}

	fields := strings.Split(args, ",")

	if i, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
		if len(fields) == 1 {
			return func(args string) any {
				return atomic.AddUint64(&i, 1) - 1
			}
		} else {
			return func(args string) any {
				return fmt.Sprintf(fields[1], atomic.AddUint64(&i, 1)-1)
			}
		}
	}

	log.Printf("bad argument %s for @seq, should use int like @seq(1000)", args)
	return func(args string) any {
		return 0
	}
}

func RandomIP(args string) any {
	if args == "" || args == "v4" {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, rand.Uint32())
		return net.IP(buf).String()
	} else if args == "v6" {
		buf := make([]byte, 16)
		binary.LittleEndian.PutUint64(buf, rand.Uint64())
		binary.LittleEndian.PutUint64(buf[8:], rand.Uint64())
		return net.IP(buf).To16().String()
	}

	if _, ipNet, err := net.ParseCIDR(args); err == nil {
		// The number of leading 1s in the mask
		ones, _ := ipNet.Mask.Size()
		quotient := ones / 8
		remainder := ones % 8

		// create random 4-byte byte slice
		r := make([]byte, 4)
		_, _ = crand.Read(r)

		for i := 0; i <= quotient; i++ {
			if i < quotient {
				r[i] = ipNet.IP[i]
			} else {
				shifted := r[i] >> remainder
				r[i] = ^ipNet.IP[i] & shifted
			}
		}
		return net.IPv4(r[0], r[1], r[2], r[3]).String()
	}

	return "127.0.0.1"
}

func RandomInt(args string) any {
	if args == "" {
		return ss.Rand().Int64()
	}

	if ranged, paddingSize, from, to, _, err := parseRandSize(args); err == nil {
		var n int64
		if from < to || ranged {
			n = ss.Rand().Int64Between(from, to)
		} else {
			n = ss.Rand().Int64n(to)
		}

		if paddingSize <= 0 {
			return n
		}
		return fmt.Sprintf("%0*d", paddingSize, n)
	}

	var err error
	vv := int64(0)
	count := 0
	for _, el := range strings.Split(args, ",") {
		v := strings.TrimSpace(el)
		if v == "" {
			continue
		} else if !ss.Rand().Bool() {
			continue
		}

		if vv, err = strconv.ParseInt(v, 10, 64); err == nil {
			return vv
		}
		count++
	}

	if count > 0 {
		return vv
	}

	return ss.Rand().Int64()
}

var argRegexp = regexp.MustCompile(`([^\s=]+)\s*(?:=\s*(\S+))?`)

func ParseConf(args string, v any) {
	MapToConf(ParseArguments(args), v)
}

func ParseArguments(args string) map[string][]string {
	result := make(map[string][]string)
	subs := argRegexp.FindAllStringSubmatch(args, -1)
	for _, sub := range subs {
		k, v := sub[1], sub[2]
		result[k] = append(result[k], v)
	}

	return result
}

func MapToConf(source map[string][]string, v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		panic("v should be pointer to struct ")
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		panic("v should be pointer to struct ")
	}
	mm := make(map[string][]string)
	for k, vv := range source {
		kk := strings.ToLower(k)
		for _, v := range vv {
			mm[kk] = append(mm[kk], v)
		}
	}

	t := elem.Type()
	for i := 0; i < t.NumField(); i++ {
		ti := t.Field(i)
		fi := elem.Field(i)

		switch ti.Type.Kind() {
		case reflect.Map:
			if prefix := ti.Tag.Get("prefix"); prefix != "" {
				m := make(map[string]string)
				for mk, mv := range source {
					if strings.HasPrefix(mk, prefix) {
						delete(m, mk)
						m[strings.TrimPrefix(mk, prefix)] = mv[0]
					}
				}
				fi.Set(reflect.ValueOf(m))
				continue
			}
		}

		name := strings.ToLower(ti.Name)
		if mv, ok := mm[name]; ok {
			delete(mm, name)
			bv := mv[0]

			switch ti.Type.Kind() {
			case reflect.Slice:
				switch ti.Type.Elem().Kind() {
				case reflect.String:
					fi.Set(reflect.ValueOf(mv))
				}
			case reflect.String:
				fi.Set(reflect.ValueOf(bv))
			case reflect.Bool:
				b := bv == "" || bv == "true" || bv == "yes" || bv == "1"
				fi.Set(reflect.ValueOf(b))
			case reflect.Int:
				b, _ := strconv.Atoi(bv)
				fi.Set(reflect.ValueOf(b))
			}
		}
	}
}

func RandomBase64(args string) any {
	arg := struct {
		Size string
		Std  bool
		URL  bool
		Raw  bool
		File string
	}{}

	ParseConf(args, &arg)

	var token []byte
	if arg.File != "" {
		if r, err := os.ReadFile(arg.File); err == nil {
			token = r
		} else {
			log.Printf("read file %s failed: %v", arg.File, err)
		}
	} else if size, _ := ss.ParseBytes(arg.Size); size > 0 {
		token = make([]byte, size)
		rand.New(rand.NewSource(time.Now().UnixNano())).Read(token)
	}

	encoding := base64.StdEncoding
	if arg.URL {
		if arg.Raw {
			encoding = base64.RawURLEncoding
		} else {
			encoding = base64.URLEncoding
		}
	} else {
		if arg.Raw {
			encoding = base64.RawStdEncoding
		}
	}

	return encoding.EncodeToString(token)
}

func randomEmoji(args string) any {
	if ranged, _, from, to, _, err := parseRandSize(args); err == nil {
		if from < to || ranged {
			return GenerateTimes(gofakeit.Emoji, from, to)
		}
		return GenerateTimes(gofakeit.Emoji, to, to)
	}

	return gofakeit.Emoji()
}

func GenerateTimes(f func() string, from, to int64) string {
	ret := ""
	end := int(ss.Rand().Int64Between(from, to))
	for i := 0; i < end; i++ {
		ret += f()
	}
	return ret
}

func randomChinese(args string) any {
	if ranged, _, from, to, _, err := parseRandSize(args); err == nil {
		if from < to || ranged {
			return ss.Rand().Chinese(int(from), int(to))
		}

		return ss.Rand().Chinese(int(to), int(to))
	}

	return ss.Rand().Chinese(2, 3)
}

func Random(args string) any {
	if args == "" {
		return ss.Rand().String(10)
	}
	if i, err := strconv.Atoi(args); err == nil {
		return ss.Rand().String(i)
	}

	if size, err := ss.ParseBytes(args); err == nil {
		b := make([]byte, size*3/4)
		n, _ := crand.Read(b)
		return base64.RawURLEncoding.EncodeToString(b[:n])
	}

	lastEl := ""
	for _, el := range strings.Split(args, ",") {
		if lastEl = strings.TrimSpace(el); lastEl == "" {
			continue
		}

		if ss.Rand().Bool() {
			return el
		}
	}

	if lastEl != "" {
		return lastEl
	}

	return ss.Rand().String(10)
}

func Regex(args string) any {
	g, err := reggen.Generate(args, 100)
	if err != nil {
		log.Printf("bad regex: %s, err: %v", args, err)
	}
	return g
}

// ObjectID is the BSON ObjectID type.
type ObjectID [12]byte

var (
	objectIDCounter = readRandomUint32()
	processUnique   = processUniqueBytes()
)

// NewObjectID generates a new ObjectID.
func NewObjectID() ObjectID {
	return NewObjectIDFromTimestamp(time.Now())
}

// NewObjectIDFromTimestamp generates a new ObjectID based on the given time.
func NewObjectIDFromTimestamp(timestamp time.Time) ObjectID {
	var b [12]byte

	binary.BigEndian.PutUint32(b[0:4], uint32(timestamp.Unix()))
	copy(b[4:9], processUnique[:])
	putUint24(b[9:12], atomic.AddUint32(&objectIDCounter, 1))

	return b
}

// Timestamp extracts the time part of the ObjectId.
func (id ObjectID) Timestamp() time.Time {
	unixSecs := binary.BigEndian.Uint32(id[0:4])
	return time.Unix(int64(unixSecs), 0).UTC()
}

// Hex returns the hex encoding of the ObjectID as a string.
func (id ObjectID) Hex() string {
	return hex.EncodeToString(id[:])
}

func processUniqueBytes() [5]byte {
	var b [5]byte
	_, err := io.ReadFull(rander, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return b
}

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rander, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func putUint24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

// NewUUID creates a new random UUID or panics.
func NewUUID(version string) uuid.UUID {
	return MustNewUUID(NewRandomUUID(version))
}

// MustNewUUID returns uuid if err is nil and panics otherwise.
func MustNewUUID(uuid uuid.UUID, err error) uuid.UUID {
	if err != nil {
		panic(err)
	}
	return uuid
}

var (
	rander = crand.Reader // random function
)

// NewRandomUUID returns a Random (Version 4) UUID.
//
// The strength of the UUIDs is based on the strength of the crypto/rand
// package.
//
// A note about uniqueness derived from the UUID Wikipedia entry:
//
//	Randomly generated UUIDs have 122 random bits.  One's annual risk of being
//	hit by a meteorite is estimated to be one chance in 17 billion, that
//	means the probability is about 0.00000000006 (6 × 10−11),
//	equivalent to the odds of creating a few tens of trillions of UUIDs in a
//	year and having one duplicate.
func NewRandomUUID(version string) (uuid.UUID, error) {
	switch version {
	case "v1", "V1", "1":
		return uuid.NewUUID()
	case "v6", "V6", "6":
		return uuid.NewV6()
	case "v7", "V7", "7":
		return uuid.NewV7()
	}
	return uuid.New(), nil
}
