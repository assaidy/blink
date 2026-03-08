package cache

import (
	"fmt"

	"github.com/valkey-io/valkey-go"
)

var client valkey.Client

func GetValkeyClient(valkeyAddr string) valkey.Client {
	if client != nil {
		return client
	}

	var err error
	if client, err = valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{valkeyAddr},
	}); err != nil {
		panic(fmt.Sprintf("error connecting to valkey server: %v", err))
	}

	return client
}
