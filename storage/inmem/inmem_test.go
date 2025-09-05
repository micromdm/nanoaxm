package inmem

import (
	"context"
	"testing"

	"github.com/micromdm/nanoaxm/storage/test"
)

func TestInMem(t *testing.T) {
	s := New()
	test.TestStorage(t, context.Background(), s, s.RetrieveClientAssertion)
}
