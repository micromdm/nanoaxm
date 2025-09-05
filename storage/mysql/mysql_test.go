package mysql

import (
	"context"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micromdm/nanoaxm/storage"
	"github.com/micromdm/nanoaxm/storage/test"
)

func TestMySQLStorage(t *testing.T) {
	testDSN := os.Getenv("NANOAXM_MYSQL_STORAGE_TEST_DSN")
	if testDSN == "" {
		t.Skip("NANOAXM_MYSQL_STORAGE_TEST_DSN is empty")
	}

	s, err := New(WithDSN(testDSN))
	if err != nil {
		t.Fatal(err)
	}

	gca := func(ctx context.Context, axmName string) (storage.ClientAssertion, error) {
		dbca, err := s.q.RetrieveClientAssertion(ctx, axmName)
		return dbcaToCA(dbca), err
	}

	test.TestStorage(t, context.Background(), s, gca)
}
