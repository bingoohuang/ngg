package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gum"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/xo/dburl"
)

func main() {
	var cobraCmd = &cobra.Command{
		Use:  "redis",
		Long: "redis client",
	}

	root.CreateCmd(cobraCmd, "scan", "scan keys for list/del", &scanCmd{})
	root.CreateCmd(cobraCmd, "set", "set string key", &setCmd{})
	root.CreateCmd(cobraCmd, "get", "get key", &getCmd{})
	root.CreateCmd(cobraCmd, "export", "export keys", &exportCmd{})
	root.CreateCmd(cobraCmd, "import", "import keys", &importCmd{})

	if err := cobraCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type Basic struct {
	Server   string `short:"s" help:"redis server" default:":6379" persist:"1" env:"auto"`
	Password string `short:"p" persist:"1"  env:"auto"`
	Db       int    `help:"default redis DB index" persist:"1"`

	Timeout time.Duration `help:"timeout"`
}

func (f Basic) NewRds() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         FulfileHostPort(f.Server, "127.0.0.1", 6379),
		Password:     f.Password,
		DB:           f.Db,
		DialTimeout:  f.Timeout,
		ReadTimeout:  f.Timeout,
		WriteTimeout: f.Timeout,
	})
}

type getCmd struct {
	Basic    `squash:"true"`
	Key      string   `short:"k" help:"key"`
	Field    []string `short:"f" help:"hash field"`
	Raw      bool     `short:"r" help:"use raw json format"`
	OnlyKeys bool     `help:"only list hash fields "`
}

// FulfileHostPort 补全 Host 和 Port
// e.g.
// FulfileHostPort("192.168.56.110", "127.0.0.1", 6379)  => 192.168.56.110:6379
// FulfileHostPort(":8080", "127.0.0.1", 6379)           => 127.0.0.1:6379
// FulfileHostPort(":", "127.0.0.1", 6379)               => 127.0.0.1:6379
func FulfileHostPort(expr, host string, port int) string {
	// 检查 expr 是否仅仅是 ":"
	if expr == ":" {
		return fmt.Sprintf("%s:%d", host, port)
	}

	// 检查 expr 是否是 IP 地址或主机名（不包含端口）
	if net.ParseIP(expr) != nil {
		return fmt.Sprintf("%s:%d", expr, port)
	}

	// 检查 expr 是否只包含端口
	if strings.HasPrefix(expr, ":") {
		return fmt.Sprintf("%s:%s", host, expr[1:])
	}

	// 默认返回 expr
	return expr
}

func (f *getCmd) Run(cmd *cobra.Command, args []string) error {
	rdb := f.NewRds()
	defer rdb.Close()

	typ, err := rdb.Type(cmd.Context(), f.Key).Result()
	if err != nil {
		return err
	}

	var val any
	switch typ {
	case "string":
		val, err = rdb.Get(cmd.Context(), f.Key).Result()
	case "hash":
		if len(f.Field) > 0 {
			val, err = rdb.HMGet(cmd.Context(), f.Key, f.Field...).Result()
		} else {
			if f.OnlyKeys {
				val, err = rdb.HKeys(cmd.Context(), f.Key).Result()
				sort.Strings(val.([]string))
			} else {
				val, err = rdb.HGetAll(cmd.Context(), f.Key).Result()
			}
		}
	default:
	}

	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Printf("key: %s does not exist", f.Key)
		} else {
			log.Printf("get key: %s field: %s error: %v", f.Key, f.Field, err)
		}
		return nil
	}

	switch typ {
	case "string":
		if !f.Raw && jj.Valid(val.(string)) {
			log.Printf("%s key: %s value: %s", typ, f.Key, jj.Pretty([]byte(val.(string))))
		} else {
			log.Printf("%s key: %s value: %v", typ, f.Key, val)
		}
	case "hash":
		value := ss.Json(val)
		if !f.Raw {
			value = jj.Pretty(jj.FreeInnerJSON(value))
		}

		log.Printf("%s key: %s field: %v value: %s", typ, f.Key, f.Field, value)
	}

	return nil
}

type setCmd struct {
	Basic `squash:"true"`
	Key   string        `short:"k" help:"key"`
	Field string        `short:"f" help:"hash field"`
	Val   string        `short:"v" help:"value"`
	Tmpl  bool          `help:"support go template in val, like '{{unixTime}}'"`
	Exp   time.Duration `help:"set expiry time for the key"`
}

// 定义一个 FuncMap，将我们自定义的函数注册到模板中
var funcMap = template.FuncMap{
	"unixTime": func() string {
		return strconv.FormatInt(time.Now().Unix(), 10)
	},
}

func (f *setCmd) Run(cmd *cobra.Command, args []string) error {
	r, err := gterm.Option{Random: true}.Open(f.Val)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer r.Close()

	data, err := r.ToBytes()
	if err != nil {
		return err
	}

	val := string(data)
	if f.Tmpl {
		// 解析模板并将 FuncMap 应用到模板中
		t := template.Must(template.New("val").Funcs(funcMap).Parse(val))
		var buf bytes.Buffer
		if err := t.Execute(&buf, nil); err != nil {
			return err
		}
		val1 := buf.String()
		if val1 != val {
			log.Printf("val interpreted to %s", val1)
			val = val1
		}
	}

	rdb := f.NewRds()
	defer rdb.Close()

	if f.Field == "" {
		if err := rdb.Set(cmd.Context(), f.Key, val, f.Exp).Err(); err != nil {
			log.Printf("redis set err: %v", err)
		}
	} else {
		if err := rdb.HSet(cmd.Context(), f.Key, f.Field, val).Err(); err != nil {
			log.Printf("redis hset err: %v", err)
		}
	}

	return nil
}

type importCmd struct {
	Basic `squash:"true"`
	File  string `help:"exported file, e.g. redis.json"`
}

func (f *importCmd) Run(cmd *cobra.Command, args []string) error {
	rdb := f.NewRds()
	defer rdb.Close()

	file, err := os.Open(f.File)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	itemIndex := 0

	for {
		var item ImportKeyItem
		if err := decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch item.Type {
		case "string":
			var str string
			if err := json.Unmarshal(item.Value, &str); err != nil {
				log.Printf("unmarshal value err: %v", err)
				continue
			}
			if err := rdb.Set(cmd.Context(), item.Key, str, 0).Err(); err != nil {
				log.Printf("redis set err: %v", err)
			} else {
				itemIndex++
				log.Printf("%d: string key %s imported", itemIndex, item.Key)
			}
		case "hash":
			var val map[string]string
			if err := json.Unmarshal(item.Value, &val); err != nil {
				return err
			}

			for k, v := range val {
				if err := rdb.HSet(cmd.Context(), item.Key, k, v).Err(); err != nil {
					return err
				}
			}
			itemIndex++
			log.Printf("%d: hash key %s imported", itemIndex, item.Key)
		}
	}

	return nil
}

type RawMessage json.RawMessage

func (r RawMessage) String() string {
	if len(r) == 0 {
		return ""
	}

	if len(r) >= 2 && r[0] == '"' && r[len(r)-1] == '"' {
		return string(r[1 : len(r)-1])
	}

	return string(r)
}

type ImportKeyItem struct {
	Key   string          `json:"key"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type exportCmd struct {
	Basic `squash:"true"`

	Type     string   `help:"scan type" enum:"string,hash,list,set,zset"`
	Key      []string `short:"k" help:"keys scan pattern" default:"*"`
	Excludes []string `help:"exclude keys"`
	Max      int      `help:"scan max keys" default:"30"`
	File     string   `help:"export file, e.g. redis.json"`
	Pretty   bool     `help:"use pretty json format"`

	exportItems int

	Rdb       string `help:"relational database URL for exporting, e.g. mysql://root:pass@127.0.0.1:3306/mydb"`
	StringSQL string `help:"insert sql for export string values"`
	HashSQL   string `help:"insert sql for export hash values"`

	db            *sql.DB
	stringSQlStmt *sql.Stmt
	hashSQlStmt   *sql.Stmt
	keyIndex      int
}

func (f *exportCmd) Run(cmd *cobra.Command, args []string) error {
	rdb := f.NewRds()
	defer rdb.Close()

	var err error
	if f.Rdb != "" {
		if f.db, err = dburl.Open(f.Rdb); err != nil {
			return err
		}
		if err := f.db.Ping(); err != nil {
			return err
		}

		defer func() {
			if f.stringSQlStmt != nil {
				f.stringSQlStmt.Close()
			}
			if f.hashSQlStmt != nil {
				f.hashSQlStmt.Close()
			}
			f.db.Close()
		}()
	}

	var exportFile = os.Stdout
	if f.File != "" {
		exportFile, err = os.Create(f.File)
		if err != nil {
			return err
		}
		defer ss.Close(exportFile)
	}

	encoder := json.NewEncoder(exportFile)
	defer func() {
		log.Printf("total %d keys exported", f.exportItems)
	}()

	excluded := createKeyExcludes(f.Excludes)
	if len(f.Key) == 0 {
		f.Key = []string{"*"}
	}

	for _, key := range f.Key {
		if err := f.exportPattern(cmd.Context(), rdb, key, excluded, encoder); err != nil {
			return err
		}

		if f.Max > 0 && f.keyIndex >= f.Max {
			break
		}
	}

	return nil
}

func (f *exportCmd) exportPattern(ctx context.Context, rdb *redis.Client,
	pattern string, excluded func(string) bool, encoder *json.Encoder,
) error {
	var cursor uint64
	scanCount := ss.If(f.Max > 0, f.Max, 0)

	for {
		keys, cursor2, err := rdb.ScanType(ctx, cursor, pattern, int64(scanCount), f.Type).Result()
		if err != nil {
			log.Printf("scan redis error: %v", err)
			return err
		}
		for _, key := range keys {
			if excluded(key) {
				continue
			}

			f.keyIndex++
			typ := f.Type
			if f.Type == "" {
				typ, err = rdb.Type(ctx, key).Result()
				if err != nil {
					return err
				}
			}

			if err := f.exportKeys(ctx, rdb, encoder, key, typ); err != nil {
				log.Printf("export %s error: %v", key, err)
			}
		}

		cursor = cursor2
		if cursor == 0 || f.Max > 0 && f.keyIndex >= f.Max { // no more keys
			return nil
		}
	}

}

type scanCmd struct {
	Basic `squash:"true"`

	Type     string   `help:"scan type" enum:"string,hash,list,set,zset"`
	Key      string   `short:"k" help:"keys scan pattern" default:"*"`
	Excludes []string `help:"exclude keys"`
	Max      int      `help:"scan max keys" default:"30"`
	Del      bool     `help:"delete the keys"`
	Force    bool     `short:"f" help:"force delete without confirm"`
}

func (f *scanCmd) Run(cmd *cobra.Command, args []string) error {
	rdb := f.NewRds()
	defer rdb.Close()

	excluded := createKeyExcludes(f.Excludes)

	var cursor uint64
	keyIndex := 0
	scanCount := ss.If(f.Max > 0, f.Max, 0)
	for {
		var keys []string
		var err error
		keys, cursor, err = rdb.ScanType(cmd.Context(), cursor, f.Key, int64(scanCount), f.Type).Result()
		if err != nil {
			log.Printf("scan redis error: %v", err)
			return nil
		}
		for _, key := range keys {
			if excluded(key) {
				continue
			}

			keyIndex++
			typ := f.Type
			if f.Type == "" {
				typ, err = rdb.Type(cmd.Context(), key).Result()
				if err != nil {
					return err
				}
			}

			log.Printf("#%d key: %s, type: %s", keyIndex, key, typ)
			if f.Del {
				confirm := true
				if !f.Force {
					confirm, _ = gum.Confirm("Sure to delete " + strconv.Quote(key))
				}

				if confirm {
					deletedResult, err := rdb.Del(cmd.Context(), key).Result()
					if err != nil {
						log.Printf("%s deleted error: %v", key, err)
					} else {
						log.Printf("%s deleted reuslt: %d", key, deletedResult)
					}
				}
			}
		}

		if cursor == 0 || f.Max > 0 && keyIndex >= f.Max { // no more keys
			break
		}
	}

	return nil
}

func createKeyExcludes(excludesKeys []string) func(string) bool {
	if len(excludesKeys) == 0 {
		return func(string) bool { return false }
	}

	return func(s string) bool {
		for _, exclude := range excludesKeys {
			if yes, _ := filepath.Match(exclude, s); yes {
				return true
			}
		}
		return false
	}
}

type KeyItem struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value any    `json:"value"`
	Error error  `json:"error,omitempty"`
}

func (f *exportCmd) exportKeys(ctx context.Context, rdb *redis.Client, encoder *json.Encoder, key, typ string) error {
	var item KeyItem
	item.Key = key
	item.Type = typ
	switch typ {
	case "string":
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			item.Error = err
		} else {
			if f.db != nil && f.StringSQL != "" {
				if f.stringSQlStmt == nil {
					f.stringSQlStmt, err = f.db.PrepareContext(ctx, f.StringSQL)
					if err != nil {
						return err
					}
				}
				if _, err := f.stringSQlStmt.ExecContext(ctx, key, val); err != nil {
					return err
				}
			}

			if f.Pretty && jj.Valid(val) {
				item.Value = json.RawMessage(jj.FreeInnerJSON([]byte(val)))
			} else {
				item.Value = val
			}
		}

	case "hash":
		val, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			item.Error = err
		} else {
			if f.db != nil && f.HashSQL != "" {
				if f.hashSQlStmt == nil {
					f.hashSQlStmt, err = f.db.PrepareContext(ctx, f.HashSQL)
					if err != nil {
						return err
					}
				}

				for k, v := range val {
					if _, err := f.hashSQlStmt.ExecContext(ctx, key, k, v); err != nil {
						return err
					}
				}
			}

			if f.Pretty {
				hashValues := make(map[string]any)
				for k, v := range val {
					if jj.Valid(v) {
						hashValues[k] = json.RawMessage(jj.FreeInnerJSON([]byte(v)))
					} else {
						hashValues[k] = v
					}
				}
				item.Value = hashValues
			} else {
				item.Value = val
			}
		}
	default:
	}

	if err := encoder.Encode(item); err != nil {
		return err
	}
	f.exportItems++

	return nil
}
