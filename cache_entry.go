package consulclient

import "time"

type CacheEntry struct {
	Expired time.Time
	Value   any
}
