package cache

import "time"

type Option func(*item)

func WithTTL(ttl time.Duration) func(*item) {
	return func(c *item) {
		c.ttl = ttl
	}
}
