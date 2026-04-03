package cache

import (
	"fmt"
	"sync"

	"github.com/assaidy/blink/app/config"
	"github.com/valkey-io/valkey-go"
)

var (
	client valkey.Client
	once   sync.Once
)

func Get() valkey.Client {
	once.Do(func() {
		var err error
		if client, err = valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{config.Get().ValkeyAddr},
		}); err != nil {
			panic(fmt.Sprintf("error connecting to valkey server: %v", err))
		}
	})

	return client
}
