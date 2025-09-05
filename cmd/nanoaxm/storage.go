package main

import (
	"errors"
	"fmt"

	"github.com/micromdm/nanoaxm/storage"
	"github.com/micromdm/nanoaxm/storage/diskv"
	"github.com/micromdm/nanoaxm/storage/inmem"
	"github.com/micromdm/nanoaxm/storage/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micromdm/nanolib/log"
)

var errOptionsNotSupported = errors.New("options not supported")

func newStore(storage, dsn, options string, _ log.Logger) (storage.AllStorage, error) {
	switch storage {
	case "file", "filekv":
		if options != "" {
			return nil, errOptionsNotSupported
		}
		if dsn == "" {
			dsn = "db"
		}
		return diskv.New(dsn), nil
	case "inmem":
		if options != "" {
			return nil, errOptionsNotSupported
		}
		return inmem.New(), nil
	case "mysql":
		if options != "" {
			return nil, errOptionsNotSupported
		}
		return mysql.New(mysql.WithDSN(dsn))
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storage)
	}
}
