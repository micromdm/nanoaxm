package diskv

import (
	"context"
	"testing"

	"github.com/micromdm/nanoaxm/storage/test"
)

func TestDiskv(t *testing.T) {
	s := New(t.TempDir())
	test.TestStorage(t, context.Background(), s, s.RetrieveClientAssertion)
}
