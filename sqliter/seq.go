package sqliter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/bingoohuang/ngg/ss"
	bolt "go.etcd.io/bbolt"
)

type BoltSeq struct {
	DB     *bolt.DB
	Bucket []byte
}

func NewBoltSeq(name, bucket string) (*BoltSeq, error) {
	// Open the data file in your current directory.
	// It will be created if it doesn't exist.

	// "invalid argument" when opening db https://github.com/boltdb/bolt/issues/272
	// Basically you can't run a boltdb file out of a shared folder in Virtualbox, especially if your on a mac.
	// The workaround here is to run the boltdb file out of a non-shared folder.
	// 基本上，您不能从 Virtualbox 中的共享文件夹运行 boltdb 文件，尤其是在 Mac 上时。 此处的解决方法是在非共享文件夹中运行 boltdb 文件。
	db, err := bolt.Open(name, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}

	if err := createBucketIfNotExists(db, []byte(bucket)); err != nil {
		return nil, fmt.Errorf("create bucket %s: %w", bucket, err)
	}

	return &BoltSeq{
		DB:     db,
		Bucket: []byte(bucket),
	}, nil
}

func (c *BoltSeq) Close() error {
	return c.DB.Close()
}

func (c *BoltSeq) Next(key string) (seq uint64, err error) {
	seq, err = c.Get(key)
	if err != nil && errors.Is(err, ErrNotFound) {
		return c.set(key)
	}

	return seq, nil
}

func (c *BoltSeq) set(key string) (seq uint64, err error) {
	k := []byte(key)
	err = c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.Bucket)
		if b == nil {
			return fmt.Errorf("no bucket %s", c.Bucket)
		}

		value := b.Get(k)
		if len(value) > 0 {
			seq = binary.LittleEndian.Uint64(value)
			return nil
		}

		if seq, err = b.NextSequence(); err != nil {
			return fmt.Errorf("next sequence: %w", err)
		}

		value = make([]byte, 8)
		binary.LittleEndian.PutUint64(value, seq)
		if err := b.Put(k, value); err != nil {
			return fmt.Errorf("put key %s: %w", key, err)
		}

		if err := b.Put(value, k); err != nil {
			return fmt.Errorf("put key %s: %w", key, err)
		}

		return nil
	})

	return
}

var (
	// ErrNotFound is returned when an entry is not found.
	ErrNotFound = errors.New("not found")
)

func (c *BoltSeq) WalkSeq(walker func(seq uint64, str string)) (err error) {
	return c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.Bucket)
		if b == nil {
			return fmt.Errorf("no bucket %s", c.Bucket)
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(k) == 8 {
				seq := binary.LittleEndian.Uint64(k)
				walker(seq, string(v))
			}
		}

		return nil
	})
}

func (c *BoltSeq) Find(seq string) (key string, err error) {
	intSeq, err := ss.Parse[uint64](seq)
	if err != nil {
		return "", err
	}

	if intSeq == 0 {
		return seq, nil
	}

	k := make([]byte, 8)
	binary.LittleEndian.PutUint64(k, intSeq)
	err = c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.Bucket)
		if b == nil {
			return fmt.Errorf("no bucket %s", c.Bucket)
		}

		if key = string(b.Get(k)); key == "" {
			key = seq
		}
		return nil
	})

	return
}

func (c *BoltSeq) Get(key string) (seq uint64, err error) {
	k := []byte(key)
	err = c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.Bucket)
		if b == nil {
			return fmt.Errorf("no bucket %s", c.Bucket)
		}

		value := b.Get(k)
		if len(value) == 0 {
			return fmt.Errorf("key not found %s: %w", key, ErrNotFound)
		}

		seq = binary.LittleEndian.Uint64(value)
		return nil
	})

	return
}

func createBucketIfNotExists(db *bolt.DB, bucketName []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if errors.Is(err, bolt.ErrBucketExists) {
			return nil
		}
		return err
	})
}
