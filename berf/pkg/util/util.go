package util

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"go.uber.org/multierr"
)

type Float64 float64

func (f Float64) MarshalJSON() ([]byte, error) {
	b := []byte(strconv.FormatFloat(float64(f), 'f', 1, 64))
	i := len(b) - 1
	for ; i >= 0; i-- {
		if b[i] != '0' {
			if b[i] != '.' {
				i++
			}
			break
		}
	}

	return b[:i], nil
}

type SizeUnit int

const (
	KILO SizeUnit = 1000
	MEGA          = 1000 * KILO
	GIGA          = 1000 * MEGA
)

func BytesToGiga(bytes uint64) Float64 {
	return Float64(float64(bytes) / float64(GIGA))
}

func BytesToMEGA(bytes uint64) Float64 {
	return Float64(float64(bytes) / float64(MEGA))
}

func BytesToBPS(bytes uint64, d time.Duration) Float64 {
	return Float64(float64(bytes*8) / float64(MEGA) / d.Seconds())
}

func BytesToMBS(bytes uint64, d time.Duration) Float64 {
	return Float64(float64(bytes) / float64(MEGA) / d.Seconds())
}

func NumberToRate(num uint64, d time.Duration) Float64 {
	return Float64(float64(num) / d.Seconds())
}

type JSONLogFile struct {
	F *os.File
	*sync.Mutex
	Name    string
	Dry     bool
	Closed  bool
	HasRows bool
}

const (
	DrySuffix = ":dry"
	GzSuffix  = ".gz"
)

func IsDrySuffix(file string) bool {
	return strings.HasSuffix(file, DrySuffix) || strings.HasSuffix(file, GzSuffix)
}

func TrimDrySuffix(file string) string {
	return strings.TrimSuffix(file, DrySuffix)
}

type JSONLogger interface {
	io.Closer
	IsDry() bool
	ReadAll() []byte
	WriteJSON(data []byte) error
}

type jsonLoggerNoop struct{}

func (jsonLoggerNoop) Close() error           { return nil }
func (jsonLoggerNoop) IsDry() bool            { return false }
func (jsonLoggerNoop) ReadAll() []byte        { return nil }
func (jsonLoggerNoop) WriteJSON([]byte) error { return nil }

func NewJsonLogFile(file string) JSONLogger {
	if file == "" {
		return &jsonLoggerNoop{}
	}

	dry := IsDrySuffix(file)
	if dry {
		file = TrimDrySuffix(file)
	} else {
		file = "berf_" + time.Now().Format(`200601021504`) + ".log"
	}
	logFile := &JSONLogFile{Name: file, Mutex: &sync.Mutex{}, Dry: dry}

	if dry {
		return logFile
	}

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		log.Printf("E! Fail to open log file %s error: %v", file, err)
	}
	logFile.F = f
	if n, err := f.Seek(0, io.SeekEnd); err != nil {
		log.Printf("E! fail to seek file %s error: %v", file, err)
	} else if n == 0 {
		_, _ = f.WriteString("[]\n")
	} else {
		logFile.HasRows = true
	}
	return logFile
}

// LogErr1 logs an error.
func LogErr1(err error) {
	if err != nil {
		log.Printf("failed %v", err)
	}
}

func LogErr2[T any](t T, err error) T {
	if err != nil {
		log.Printf("failed %v", err)
	}
	return t
}

func (f JSONLogFile) ReadAll() []byte {
	f.Lock()
	defer f.Unlock()

	if f.F == nil || f.Closed {
		return ss.Must(ss.ReadFile(f.Name))
	}

	_, _ = f.F.Seek(0, io.SeekStart)
	defer LogErr2(f.F.Seek(0, io.SeekEnd))

	data, err := io.ReadAll(f.F)
	if err != nil {
		log.Printf("E! fail to read log file %s, error: %v", f.F.Name(), err)
	}
	return data
}

func (f *JSONLogFile) WriteJSON(data []byte) error {
	if f.F == nil {
		return nil
	}

	f.Lock()
	defer f.Unlock()

	_, _ = f.F.Seek(-2, io.SeekEnd) // \n]
	var err0 error

	if !f.HasRows {
		f.HasRows = true
		_, err0 = f.F.WriteString("\n")
	} else {
		_, err0 = f.F.WriteString(",\n")
	}
	_, err1 := f.F.Write(data)
	_, err2 := f.F.WriteString("\n]")
	return multierr.Combine(err0, err1, err2)
}

func (f JSONLogFile) IsDry() bool { return f.Dry }

func (f *JSONLogFile) Close() error {
	if f.F == nil {
		return nil
	}

	f.Lock()
	defer f.Unlock()
	f.Closed = true

	compress := false
	if stat, err := f.F.Stat(); err == nil && stat.Size() > 3 {
		compress = true
	}

	ss.Close(f.F)

	if compress {
		_ = gzipFile(f.Name)
	}

	return nil
}

func gzipFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(f)

	if f, err = os.Create(name + ".gz"); err != nil {
		return err
	}
	w := gzip.NewWriter(f)
	_, err = io.Copy(w, reader)
	if err := w.Close(); err == nil {
		_ = f.Close()
		_ = os.Remove(name)
	}

	return err
}

func NewFeatures(features ...string) Features {
	m := make(Features)
	m.Setup(features)
	return m
}

// Features defines a feature map.
type Features map[string]string

// Setup sets up a feature map by features string, which separates feature names by comma.
func (f *Features) Setup(featuresArr []string) {
	for _, features := range featuresArr {
		for k, v := range ss.SplitToMap(features, ",", "=") {
			(*f)[strings.ToLower(k)] = v
		}
	}
}

func (f *Features) IsNop() bool { return f.Has("nop") }

// GetOr gets the feature value or default.
func (f *Features) GetOr(feature, defaultValue string) string {
	s := (*f)[strings.ToLower(feature)]
	return ss.Or(s, defaultValue)
}

// GetInt gets the feature value as int.
func (f *Features) GetInt(feature string, defaultValue int) int {
	s := (*f)[strings.ToLower(feature)]
	if s == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("%s's value %s is not an int", feature, s)
		return defaultValue
	}

	return val
}

// GetFloat gets the feature value as int.
func (f *Features) GetFloat(feature string, defaultValue float64) float64 {
	s := (*f)[strings.ToLower(feature)]
	if s == "" {
		return defaultValue
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("%s's value %s is not a float", feature, s)
		return defaultValue
	}

	return val
}

// Get gets the feature map contains a features.
func (f *Features) Get(feature string) string {
	s := f.get(feature)
	return tryReadFile(s)
}

func tryReadFile(s string) string {
	if v, _, err := readFile(s); err != nil {
		// log.Fatalf("Read File %s failed: %v", s, err)
		return s
	} else {
		return string(v)
	}
}

func readFile(s string) (data []byte, fn string, e error) {
	if !strings.HasPrefix(s, "@") {
		return []byte(s), "", nil
	}

	s = strings.TrimPrefix(s, "@")
	f, err := os.Open(s)
	if err != nil {
		return nil, s, err
	}
	defer ss.Close(f)
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, s, err
	}
	return content, s, nil
}

func (f *Features) get(feature string) string {
	if s := (*f)[strings.ToLower(feature)]; s != "" {
		return s
	}

	return os.Getenv("FEATURE_" + strings.ToUpper(feature))
}

// Has tells the feature map contains a features.
func (f *Features) Has(feature string) bool {
	if _, ok := (*f)[strings.ToLower(feature)]; ok {
		return ok
	}

	return ss.Must(ss.GetenvBool("FEATURE_"+strings.ToUpper(feature), false))
}

// HasAny tells the feature map contains any of the features.
func (f *Features) HasAny(features ...string) bool {
	for _, feature := range features {
		if f.Has(feature) {
			return true
		}
	}

	return false
}

type WidthHeight struct {
	W, H int
}

func (h WidthHeight) WidthPx() string  { return fmt.Sprintf("%dpx", h.W) }
func (h WidthHeight) HeightPx() string { return fmt.Sprintf("%dpx", h.H) }

func ParseWidthHeight(val string, defaultWidth, defaultHeight int) WidthHeight {
	wh := WidthHeight{
		W: defaultWidth,
		H: defaultHeight,
	}
	if val != "" {
		val = strings.ToLower(val)
		parts := strings.SplitN(val, "x", 2)
		if len(parts) == 2 {
			if v, _ := ss.Parse[int](parts[0]); v > 0 {
				wh.W = v
			}
			if v, _ := ss.Parse[int](parts[1]); v > 0 {
				wh.H = v
			}
		}
	}
	return wh
}

type GoroutineIncr struct {
	Up   int
	Dur  time.Duration
	Down int
}

func (i GoroutineIncr) Modifier() string {
	return ss.If(i.Up > 0, "max ", "")
}

func (i GoroutineIncr) IsEmpty() bool {
	return i.Up <= 0 && i.Down <= 0
}

// ParseGoIncr parse a GoIncr expressions like:
// 1. (empty) => GoroutineIncr{}
// 2. 0       => GoroutineIncr{}
// 3. 1       => GoroutineIncr{Up: 1}
// 4. 1:10s   => GoroutineIncr{Up: 1, Dur:10s}
// 5. 1:10s:1 => GoroutineIncr{Up: 1, Dur:10s, Down:1}
func ParseGoIncr(s string) GoroutineIncr {
	s = strings.TrimSpace(s)
	if s == "" {
		return GoroutineIncr{Up: 0, Dur: 0}
	}

	var err error
	parts := ss.Split(s, ":")
	v, _ := ss.Parse[int](parts[0])
	ret := GoroutineIncr{Up: v, Dur: 0}
	if len(parts) >= 2 {
		ret.Dur, err = time.ParseDuration(parts[1])
		if err != nil {
			log.Printf("W! %s is invalid", s)
		}
	}
	if len(parts) >= 3 {
		ret.Down, _ = ss.Parse[int](parts[2])
	}

	return ret
}

func MergeCodes(codes []string) string {
	n := 0
	last := ""
	merged := ""
	for _, code := range codes {
		if code != last {
			if last != "" {
				merged = mergeCodes(merged, n, last)
			}
			last = code
			n = 1
		} else {
			n++
		}
	}

	if n > 0 {
		merged = mergeCodes(merged, n, last)
	}

	return merged
}

func mergeCodes(merged string, n int, last string) string {
	if merged != "" {
		merged += ","
	}
	if n > 1 {
		merged += fmt.Sprintf("%sx%d", last, n)
	} else {
		merged += last
	}
	return merged
}

func SplitTail(s *string, tail string) bool {
	if strings.HasSuffix(*s, tail) {
		*s = strings.TrimSuffix(*s, tail)
		return true
	}
	return false
}
