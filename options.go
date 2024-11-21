package cache

import "time"

type Option func(*Item)

func WithTTL(ttl time.Duration) Option {
	return func(c *Item) {
		c.ttl = ttl
	}
}
