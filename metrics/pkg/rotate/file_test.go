package rotate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/bingoohuang/ngg/metrics/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestDay(t *testing.T) {
	dir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("TempDir:", dir)

	defer os.RemoveAll(dir)

	option, err := NewFile(filepath.Join(dir, "abc.log"), 3)
	assert.Nil(t, err)

	day, _ := util.ParseTime("2020-02-10", yyyyMMdd)
	rotated, backups := option.detectRotate(day)
	assert.Zero(t, rotated)
	assert.Zero(t, backups)

	assert.Nil(t, option.doRotate(rotated, backups))

	day, _ = util.ParseTime("2020-02-11", yyyyMMdd)
	rotated, backups = option.detectRotate(day)
	assert.Equal(t, rotated, filepath.Join(dir, "abc.log.2020-02-10"))
	assert.Zero(t, backups)

	assert.Nil(t, option.doRotate(rotated, backups))

	day, _ = util.ParseTime("2020-02-12", yyyyMMdd)
	rotated, backups = option.detectRotate(day)
	assert.Equal(t, rotated, filepath.Join(dir, "abc.log.2020-02-11"))
	assert.Zero(t, backups)

	assert.Nil(t, option.doRotate(rotated, backups))

	day, _ = util.ParseTime("2020-02-13", yyyyMMdd)
	rotated, backups = option.detectRotate(day)
	assert.Equal(t, rotated, filepath.Join(dir, "abc.log.2020-02-12"))
	assert.Zero(t, backups)

	assert.Nil(t, option.doRotate(rotated, backups))

	day, _ = util.ParseTime("2020-02-14", yyyyMMdd)
	rotated, backups = option.detectRotate(day)
	assert.Equal(t, rotated, filepath.Join(dir, "abc.log.2020-02-13"))
	assert.Equal(t, backups, []string{filepath.Join(dir, "abc.log.2020-02-10")})

	assert.Nil(t, option.doRotate(rotated, backups))
}
