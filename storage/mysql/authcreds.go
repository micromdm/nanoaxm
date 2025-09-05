package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/micromdm/nanoaxm/storage"
)

// RetrieveAuthCredential retrieves the auth crendetials from storage for axmName.
// An error will be returned if axmName is invalid.
func (s *MySQLStorage) RetrieveAuthCredentials(ctx context.Context, axmName string) (storage.AuthCredentials, error) {
	if axmName == "" {
		return storage.AuthCredentials{}, fmt.Errorf("%w: empty name", storage.ErrInvalidAXMName)
	}

	var ac storage.AuthCredentials

	dbac, err := s.q.RetrieveAuthCredentials(ctx, axmName)
	if errors.Is(err, sql.ErrNoRows) {
		return ac, fmt.Errorf("%v: %w", err, storage.ErrInvalidAXMName)
	} else if err != nil {
		return ac, err
	}

	ac.ClientID = dbac.ClientID
	ac.KeyID = dbac.KeyID
	ac.PrivateKeyPEM = dbac.PrivKeyPem

	return ac, nil
}

// StoreAuthCredentials stores the auth credentials to storage for axmName.
// An error will be returned if axmName is invalid.
func (s *MySQLStorage) StoreAuthCredentials(ctx context.Context, axmName string, ac storage.AuthCredentials) error {
	if axmName == "" {
		return fmt.Errorf("%w: empty name", storage.ErrInvalidAXMName)
	}
	if err := ac.ValidError(); err != nil {
		return fmt.Errorf("auth creds invalid: %s: %w", axmName, err)
	}

	// raw SQL (vs. sqlc) due to https://github.com/sqlc-dev/sqlc/issues/2789
	_, err := s.db.ExecContext(
		ctx, `
INSERT INTO axm_names 
	(name, client_id, key_id, priv_key_pem)
VALUES 
	(?, ?, ?, ?) as new
ON DUPLICATE KEY UPDATE 
	client_id = new.client_id,
	key_id = new.key_id,
	priv_key_pem = new.priv_key_pem;`,
		axmName,
		ac.ClientID,
		ac.KeyID,
		ac.PrivateKeyPEM,
	)
	return err
}
