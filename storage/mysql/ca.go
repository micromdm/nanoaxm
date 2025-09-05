package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/micromdm/nanoaxm/storage"
	"github.com/micromdm/nanoaxm/storage/mysql/sqlc"
)

func dbcaToCA(dbca sqlc.RetrieveClientAssertionRow) storage.ClientAssertion {
	return storage.ClientAssertion{
		Token:    dbca.CaToken.String,
		Validity: time.Duration(dbca.CaValiditySec.Int32) * time.Second,
		Expiry:   time.Unix(int64(dbca.CaExpiryUnix.Int32), 0),
		ClientID: dbca.ClientID,
	}
}

// GetOrRefreshClientAssertion refreshes the OAuth 2 client assertion for axmName and stores it.
func (s *MySQLStorage) GetOrRefreshClientAssertion(ctx context.Context, axmName string, refreshFunc func(ctx context.Context, ac storage.AuthCredentials) (storage.ClientAssertion, error), refresh bool) (token storage.ClientAssertion, err error) {
	if axmName == "" {
		return token, fmt.Errorf("%w: empty name", storage.ErrInvalidAXMName)
	}
	if refreshFunc == nil {
		return token, errors.New("nil refresher")
	}

	return token, tx(ctx, s.db, s.q, func(ctx context.Context, tx *sql.Tx, qtx *sqlc.Queries) error {
		if !refresh {
			dbca, err := qtx.RetrieveClientAssertion(ctx, axmName)
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%v: %w", err, storage.ErrInvalidAXMName)
			} else if err != nil {
				return err
			}

			token = dbcaToCA(dbca)

			if token.Valid() {
				return nil
			}
			fmt.Println("invalid!")
		}

		ac, err := s.RetrieveAuthCredentials(ctx, axmName)
		if err != nil {
			return fmt.Errorf("retrieving auth creds: %w", err)
		}

		token, err = refreshFunc(ctx, ac)
		if err != nil {
			return fmt.Errorf("refreshing client assertion: %w", err)
		}

		if err = token.ValidError(); err != nil {
			return fmt.Errorf("refreshed token invalid: %w", err)
		}

		err = qtx.UpdateClientAssertion(ctx, sqlc.UpdateClientAssertionParams{
			CaToken:       sql.NullString{String: token.Token, Valid: true},
			CaValiditySec: sql.NullInt32{Int32: int32(token.Validity.Seconds()), Valid: true},
			CaExpiryUnix:  sql.NullInt32{Int32: int32(token.Expiry.Unix()), Valid: true},
			Name:          axmName,
		})
		if err != nil {
			return fmt.Errorf("storing client assertion: %w", err)
		}

		return nil
	})
}
