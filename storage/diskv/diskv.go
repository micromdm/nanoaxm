// Package diskv implements a storage backend using the diskv key-value store.
package diskv

import (
	"path/filepath"

	"github.com/micromdm/nanoaxm/storage/kv"

	"github.com/micromdm/nanolib/storage/kv/kvdiskv"
	"github.com/micromdm/nanolib/storage/kv/kvtxn"
	"github.com/peterbourgon/diskv/v3"
)

// Diskv is a storage backend that uses diskv.
type Diskv struct {
	*kv.KV
}

// New creates a new storage backend that uses diskv.
func New(path string) *Diskv {
	return &Diskv{KV: kv.New(
		kvtxn.New(kvdiskv.New(diskv.New(diskv.Options{
			BasePath:     filepath.Join(path, "axm_names"),
			Transform:    kvdiskv.FlatTransform,
			CacheSizeMax: 1024 * 1024,
		}))),
	)}
}
