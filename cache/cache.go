package cache

import (
	"time"
)

type Cache interface {
	Exists(keyStr string) (bool, error)
	Get(keyStr string) (interface{}, error)
	Set(keyStr string, value interface{}, expires ...time.Duration) error
	Delete(keyStr string) error
	EmptyByMatch(keyStr string) error
	Empty() error
	Keys(patternOrKey ...string) ([]string, error)
	Expire(keyStr string, expiration time.Duration) error
	TTL(keyStr string) (time.Duration, error)
	Update(keyStr string, value interface{}) error
}

// EntryCache is a type alias for a map used to store entries.
type EntryCache map[string]interface{}
