package redis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bingoohuang/ngg/ggt/root"
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

	Key   string        `short:"k"`
	Field string        `short:"f" help:"hash field"`
	Val   string        `short:"v" help:"set/hset value for the key"`
	Exp   time.Duration `help:"set expiry time for the key"`
}

func (f *subCmd) run(cmd *cobra.Command, args []string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     f.Server,
		Password: f.Password, // no password set
		DB:       f.Db,       // use default DB
	})

	ctx := context.Background()

	if f.Val == "" {
		typ, err := rdb.Type(ctx, f.Key).Result()
		if err != nil {
			log.Printf("redis type %s err: %v", f.Key, err)
			return nil
		} else {
			log.Printf("redis type %s", typ)
		}

		if f.Field == "" {
			val, err := rdb.Get(ctx, f.Key).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					log.Printf("Get key: %s does not exist", f.Key)
					return nil
				}
				log.Printf("Get key: %s error: %v", f.Key, err)
			} else {
				log.Printf("Get key: %s value: %s", f.Key, val)
			}
		} else {
			val, err := rdb.HGet(ctx, f.Key, f.Field).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					log.Printf("HGET key: %s does not exist", f.Key)
					return nil
				}
				log.Printf("HGET key: %s field: %s error: %v", f.Key, f.Field, err)
				return err
			}
			log.Printf("HGET key: %s field: %s value: %s", f.Key, f.Field, val)
		}
	} else {
		var err error
		if f.Field == "" {
			err = rdb.Set(ctx, f.Key, f.Val, f.Exp).Err()
		} else {
			err = rdb.HSet(ctx, f.Key, f.Field, f.Val).Err()
		}
		return err
	}

	return nil
}
