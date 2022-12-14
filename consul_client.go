package consulclient

import "time"

type ConsulClient interface {
	Get(key string) (any, error)
	GetOrSet(key string, period time.Duration) (any, error)
}
