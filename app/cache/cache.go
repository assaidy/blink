package cache

import (
	"fmt"

	"github.com/assaidy/blink/app/env"
	"github.com/valkey-io/valkey-go"
)

var client valkey.Client

func GetClient() valkey.Client {
	if client != nil {
		return client
	}

	var err error
	if client, err = valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{env.ValkeyAddr},
	}); err != nil {
		panic(fmt.Sprintf("error connecting to valkey server: %v", err))
	}

	return client
}
