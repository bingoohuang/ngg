package redis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:  "redis",
		Long: "redis client",
		RunE: fc.run,
	}
	root.AddCommand(c, fc)
}

type subCmd struct {
	Server   string `short:"s" help:"redis server" default:"127.0.0.1:6379"`
	Password string `short:"p"`
	Db       int    `help:"default DB"`

	Key     []string      `short:"k"`
	Pattern string        `short:"P" help:"keys scan pattern" default:"*"`
	MaxKeys int           `help:"scan max keys" default:"10"`
	Type    string        `help:"scan type, string/hash/list/set/zset"`
	Field   []string      `short:"f" help:"hash field"`
	Val     string        `short:"v" help:"set/hset value for the key"`
	Exp     time.Duration `help:"set expiry time for the key"`
	Raw     bool          `short:"r" help:"use raw json format"`
	Del     bool          `short:"d" help:"delete keys"`
}

func (f *subCmd) run(cmd *cobra.Command, args []string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     f.Server,
		Password: f.Password, // no password set
		DB:       f.Db,       // use default DB
	})

	ctx := context.Background()

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
				keyIndex++
				typ := f.Type
				if f.Type == "" {
					typ, _ = rdb.Type(ctx, key).Result()
				}
				log.Printf("#%d key: %s, type: %s", keyIndex, key, typ)
			}

			if cursor == 0 || keyIndex >= f.MaxKeys { // no more keys
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
			typ, err := rdb.Type(ctx, key).Result()
			if err != nil {
				log.Printf("redis type %s err: %v", f.Key, err)
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
