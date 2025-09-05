package kv

import (
	"context"
	"errors"
	"fmt"

	"github.com/micromdm/nanoaxm/storage"
	"github.com/micromdm/nanolib/storage/kv"
)

const (
	// going for a "auth.<name>.key" format

	keyPfxAC = "auth"

	keySfxClientID = "cid"
	keySfxKeyID    = "kid"
	keySfxPrivKey  = "key"
)

// RetrieveAuthCredential retrieves the auth crendetials from storage for axmName.
// An error will be returned if axmName is invalid.
func (s *KV) RetrieveAuthCredentials(ctx context.Context, axmName string) (storage.AuthCredentials, error) {
	if axmName == "" {
		return storage.AuthCredentials{}, fmt.Errorf("%w: empty name", storage.ErrInvalidAXMName)
	}

	retMap, err := kv.GetMap(ctx, s.b, []string{
		join(keyPfxAC, axmName, keySfxClientID),
		join(keyPfxAC, axmName, keySfxKeyID),
		join(keyPfxAC, axmName, keySfxPrivKey),
	})
	if errors.Is(err, kv.ErrKeyNotFound) {
		return storage.AuthCredentials{}, fmt.Errorf("%w: %v", storage.ErrInvalidAXMName, err)
	} else if err != nil {
		return storage.AuthCredentials{}, err
	}

	return storage.AuthCredentials{
		ClientID:      string(retMap[join(keyPfxAC, axmName, keySfxClientID)]),
		KeyID:         string(retMap[join(keyPfxAC, axmName, keySfxKeyID)]),
		PrivateKeyPEM: retMap[join(keyPfxAC, axmName, keySfxPrivKey)],
	}, nil
}

// StoreAuthCredentials stores the auth credentials to storage for axmName.
// An error will be returned if axmName is invalid.
func (s *KV) StoreAuthCredentials(ctx context.Context, axmName string, ac storage.AuthCredentials) error {
	if axmName == "" {
		return fmt.Errorf("%w: empty name", storage.ErrInvalidAXMName)
	}
	if err := ac.ValidError(); err != nil {
		return fmt.Errorf("auth creds invalid: %s: %w", axmName, err)
	}

	return kv.PerformCRUDBucketTxn(ctx, s.b, func(ctx context.Context, b kv.CRUDBucket) error {
		return kv.SetMap(ctx, b, map[string][]byte{
			join(keyPfxAC, axmName, keySfxClientID): []byte(ac.ClientID),
			join(keyPfxAC, axmName, keySfxKeyID):    []byte(ac.KeyID),
			join(keyPfxAC, axmName, keySfxPrivKey):  []byte(ac.PrivateKeyPEM),
			join(keyPfxName, axmName):               []byte(valOne),
		})
	})
}
