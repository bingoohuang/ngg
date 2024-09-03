package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/muesli/cache2go"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CountCacheValue struct {
	Error error
	Imgs  []Img
}

var ksuidCache = func() *cache2go.CacheTable {
	table := cache2go.Cache("count")
	table.SetDataLoader(func(key any, args ...any) *cache2go.CacheItem {
		// Apply some clever loading logic here, e.g. read values for
		// this key from database, network or file.
		dbName := key.(string)

		db, err := dbCache.Value(dbName)
		ccv := &CountCacheValue{
			Error: err,
		}
		if err != nil {
			return cache2go.NewCacheItem(key, 5*time.Minute, ccv)
		}
		dcv := db.Data().(*DbCacheValue)
		db2 := dcv.DB.Select("id").Find(&ccv.Imgs)
		ccv.Error = db2.Error
		// This helper method creates the cached item for us. Yay!
		return cache2go.NewCacheItem(key, 5*time.Minute, ccv)
	})

	return table
}()

type DbCacheValue struct {
	Error error
	DB    *gorm.DB
	Name  string
}

var dbCache = func() *cache2go.CacheTable {
	table := cache2go.Cache("db")
	// The data loader gets called automatically whenever something
	// tries to retrieve a non-existing key from the cache.
	table.SetDataLoader(dbLoader)
	// This callback will be triggered every time an item
	// is about to be removed from the cache.
	table.SetAboutToDeleteItemCallback(closeCacheDB)

	return table
}()

var ErrClosed = errors.New("db closed")

var LogLevel = func() logger.LogLevel {
	env := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(env) {
	case "info":
		return logger.Info
	case "silent":
		return logger.Silent
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}()

func dbLoader(key any, args ...any) *cache2go.CacheItem {
	// Apply some clever loading logic here, e.g. read values for
	// this key from database, network or file.
	dbName := key.(string)

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(LogLevel),
	})

	if err == nil && db != nil {
		// 迁移 schema
		if err = db.AutoMigrate(&Img{}); err != nil {
			log.Printf("AutoMigrate error: %v", err)
		}
	}

	// This helper method creates the cached item for us. Yay!
	item := cache2go.NewCacheItem(key, 5*time.Minute, &DbCacheValue{
		Name:  dbName,
		DB:    db,
		Error: err,
	})
	return item
}

func closeCacheDB(entry *cache2go.CacheItem) {
	dbValue := entry.Data().(*DbCacheValue)
	if db := dbValue.DB; db != nil {
		if db1, _ := db.DB(); db1 != nil {
			if err := db1.Close(); err != nil {
				log.Printf("close %v error: %v", entry.Key(), err)
			}
		}
		dbValue.DB = nil
		dbValue.Error = ErrClosed
	}
}
