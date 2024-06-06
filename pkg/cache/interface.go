package cache

import "time"

type ObjectCache interface {
	Get(key string, value interface{}) bool
	Set(key string, value interface{}) error
	Delete(key string) error
	Clean(maxAge time.Duration) error
}
