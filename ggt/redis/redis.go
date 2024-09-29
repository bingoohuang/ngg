package redis

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/xo/dburl"
)

func init() {
	c := &cobra.Command{
		Use:  "redis",
		Long: "redis client",
	}
	root.AddCommand(c, &subCmd{})
}

type subCmd struct {
	Server   string `short:"s" help:"redis server" default:"127.0.0.1:6379"`
	Password string `short:"p"`
	Db       int    `help:"default redis DB index"`

	Key      []string      `short:"k"`
	Pattern  string        `short:"P" help:"keys scan pattern" default:"*"`
	MaxKeys  int           `help:"scan max keys" default:"10"`
	Type     string        `help:"scan type, string/hash/list/set/zset"`
	Field    []string      `short:"f" help:"hash field"`
	Val      string        `short:"v" help:"set/hset value for the key"`
	Exp      time.Duration `help:"set expiry time for the key"`
	Raw      bool          `short:"r" help:"use raw json format"`
	Del      bool          `help:"delete keys"`
	Export   string        `help:"export file, e.g. redis.json"`
	Excludes []string      `help:"exclude keys"`

	exportItems int

	Rdb       string `help:"relational database URL for exporting, e.g. mysql://root:pass@127.0.0.1:3306/mydb"`
	StringSQL string `help:"insert sql for export string values"`
	HashSQL   string `help:"insert sql for export hash values"`

	db            *sql.DB
	stringSQlStmt *sql.Stmt
	hashSQlStmt   *sql.Stmt
}

type KeyItem struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value any    `json:"value"`
	Error error  `json:"error,omitempty"`
}

type RedisData struct {
	Keys []any `json:"keys"`
}

func (f *subCmd) Run(cmd *cobra.Command, args []string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     f.Server,
		Password: f.Password, // no password set
		DB:       f.Db,       // use default DB
	})

	ctx := context.Background()

	excluded := func(string) bool { return false }
	if len(f.Excludes) > 0 {
		excluded = func(s string) bool {
			for _, exclude := range f.Excludes {
				if yes, _ := filepath.Match(exclude, s); yes {
					return true
				}
			}
			return false
		}
	}

	var err error
	var exportJsonEncoder *json.Encoder
	var exportFile *os.File
	if f.Export != "" {
		exportFile, err = os.Create(f.Export)
		if err != nil {
			return err
		}
		exportJsonEncoder = json.NewEncoder(exportFile)
		defer func() {
			ss.Close(exportFile)
			log.Printf("total %d keys exported to %s", f.exportItems, f.Export)
		}()
	}

	if f.Rdb != "" {
		f.db, err = dburl.Open(f.Rdb)
		if err != nil {
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

	if len(f.Key) == 0 {
		var cursor uint64
		keyIndex := 0
		for {
			var keys []string
			var err error
			keys, cursor, err = rdb.ScanType(ctx, cursor, f.Pattern, 0, f.Type).Result()
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
					typ, err = rdb.Type(ctx, key).Result()
					if err != nil {
						return err
					}
				}

				if exportFile != nil {
					if err := f.exportKeys(rdb, ctx, exportJsonEncoder, key, typ); err != nil {
						log.Printf("export %s error: %v", key, err)
					}
				} else {
					log.Printf("#%d key: %s, type: %s", keyIndex, key, typ)
				}
			}

			if cursor == 0 || f.MaxKeys > 0 && keyIndex >= f.MaxKeys { // no more keys
				break
			}
		}
		return nil
	}

	if f.Del {
		result, err := rdb.Del(ctx, f.Key...).Result()
		if err != nil {
			log.Printf("del error: %v", err)
		} else {
			log.Printf("del result: %v", result)
		}
		return nil
	}

	if f.Val == "" {
		for _, key := range f.Key {
			if excluded(key) {
				continue
			}

			typ, err := rdb.Type(ctx, key).Result()
			if err != nil {
				log.Printf("redis type %s err: %v", f.Key, err)
				continue
			}

			if exportFile != nil {
				if err := f.exportKeys(rdb, ctx, exportJsonEncoder, key, typ); err != nil {
					log.Printf("export %s error: %v", key, err)
				}
				continue
			}

			var val any
			switch typ {
			case "string":
				val, err = rdb.Get(ctx, key).Result()
			case "hash":
				if len(f.Field) > 0 {
					val, err = rdb.HMGet(ctx, key, f.Field...).Result()
				} else {
					val, err = rdb.HGetAll(ctx, key).Result()
				}
			default:
			}

			if err != nil {
				if errors.Is(err, redis.Nil) {
					log.Printf("key: %s does not exist", f.Key)

				} else {
					log.Printf("HGET key: %s field: %s error: %v", f.Key, f.Field, err)
				}
				continue
			}

			switch typ {
			case "string":
				if !f.Raw && jj.Valid(val.(string)) {
					log.Printf("%s key: %s value: %s", typ, key, jj.Pretty([]byte(val.(string))))
				} else {
					log.Printf("%s key: %s value: %v", typ, key, val)
				}
			case "hash":
				value := ss.Json(val)
				if !f.Raw {
					value = jj.Pretty(jj.FreeInnerJSON(value))
				}

				log.Printf("%s key: %s field: %v value: %s", typ, f.Key, f.Field, value)
			}
		}

		return nil
	}

	for _, key := range f.Key {
		if len(f.Field) > 0 {
			for _, field := range f.Field {
				if err := rdb.HSet(ctx, key, field, f.Val).Err(); err != nil {
					log.Printf("redis hset err: %v", err)
				}
			}
		} else {
			if err := rdb.Set(ctx, key, f.Val, f.Exp).Err(); err != nil {
				log.Printf("redis set err: %v", err)
			}
		}
	}

	return nil
}

func (f *subCmd) exportKeys(rdb *redis.Client, ctx context.Context, encoder *json.Encoder, key, typ string) error {
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

			if !f.Raw && jj.Valid(val) {
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

			if !f.Raw {
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
