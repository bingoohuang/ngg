package metric

import (
	"github.com/bingoohuang/ngg/metrics/pkg/ks"
	"github.com/bingoohuang/ngg/metrics/pkg/util"
	"github.com/bingoohuang/ngg/ss"
	"github.com/sirupsen/logrus"
)

// Key defines a slice of keys.
type Key struct {
	ks      *ks.Ks
	Keys    []string
	Checked bool
}

// NewKey create Keys.
func NewKey(keys []string) Key {
	if len(keys) == 0 || len(keys) > 3 {
		panic("only at least 1 key or max 3 keys allowed")
	}
	k := Key{Keys: keys, Checked: false}
	k.Check()

	return k
}

// Check checks the validation of keys.
func (k *Key) Check() {
	k.Checked = true

	if len(k.Keys) == 0 {
		k.Checked = false
		logrus.Warn("Keys required")
		return
	}

	for i, key := range k.Keys {
		if !k.validateKey(i, key) {
			k.Checked = false
		}
	}
}

const strippedChars = `" .,|#\` + "\t\r\n"

func (k *Key) validateKey(i int, key string) bool {
	if key == "" {
		logrus.Warn("Key can not be empty")
		return false
	}

	key = util.StripAny(key, strippedChars)
	if key == "" {
		logrus.Warnf("invalid Key %s", key)
		return false
	}

	k.Keys[i] = ss.Abbreviate(key, 100, ss.DefaultEllipse)
	return true
}
