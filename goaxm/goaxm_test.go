package goaxm_test

import (
	"context"
	"log"
	"net/http"

	"github.com/micromdm/nanoaxm/goaxm"
	"github.com/micromdm/nanoaxm/storage/inmem"
)

func Example() {
	ctx := context.Background()

	store := inmem.New()

	client := goaxm.NewClient(store)

	var output struct{ Item string }

	err := client.Do(ctx, "test-axm-name", http.MethodGet, "https://api-business.apple.com/v1/mdmServers", nil, output)
	if err != nil {
		log.Fatal(err)
	}
}
