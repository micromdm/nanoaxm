// Package kv implements a NanoAxM storage backend using a key-value store.
package kv

import (
	"strconv"
	"strings"
	"time"

	"github.com/micromdm/nanolib/storage/kv"
)

const (
	keySep = "."
	valOne = "1"

	keyPfxName = "name"
)

// KV is a NanoAxM storage backend that uses a key-value store.
type KV struct {
	b kv.TxnCRUDBucket
}

func New(b kv.TxnCRUDBucket) *KV {
	return &KV{b: b}
}

// join concatenates s together by placing [keySep] in-between.
func join(s ...string) string {
	return strings.Join(s, keySep)
}

// durationToBytes converts a [time.Duration] to a byte slice representation.
func durationToBytes(d time.Duration) []byte {
	return []byte(strconv.Itoa(int(d.Seconds())))
}

// durationFromBytes converts a byte slice representation to a [time.Duration].
func durationFromBytes(b []byte) (time.Duration, error) {
	sec, err := strconv.Atoi(string(b))
	return time.Duration(sec) * time.Second, err
}

// timeToBytes converts a [time.Time] to a byte slice representation.
func timeToBytes(t time.Time) []byte {
	return []byte(strconv.FormatInt(t.Unix(), 10))
}

// timeFromBytes converts a byte slice representation to a [time.Time].
func timeFromBytes(b []byte) (time.Time, error) {
	micro, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(micro, 0), err
}
