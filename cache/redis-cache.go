package cache

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

// RedisCache struct holds the Redis connection pool and key prefix.
type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

// Close closes the Redis connection pool.
func (rc *RedisCache) Close() error {
	return rc.Conn.Close()
}

// prefixedKey returns the key with the specified prefix.
func (rc *RedisCache) prefixedKey(key string) string {
	return fmt.Sprintf("%s:%s", rc.Prefix, key)
	// return rc.Prefix + key
}

// Exists checks if a key exists in the Redis cache.
func (rc *RedisCache) Exists(keyStr string) (bool, error) {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedKey := rc.prefixedKey(keyStr)

	exists, err := redis.Bool(conn.Do("EXISTS", prefixedKey))
	if err != nil {
		log.Printf("Error checking existence of key %s: %v", keyStr, err)
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists, nil
}

// Get retrieves the value for a given prefixed key from the Redis cache
// and decodes it into an EntryCache.
func (rc *RedisCache) Get(keyStr string) (interface{}, error) {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedKey := rc.prefixedKey(keyStr)

	cacheRetrieved, err := redis.Bytes(conn.Do("GET", prefixedKey))
	if errors.Is(err, redis.ErrNil) {
		return nil, nil // Cache miss
	} else if err != nil {
		log.Printf("Error getting cache for key %s: %v", keyStr, err)
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	// un-serialize cache
	result, err := decodeValue(cacheRetrieved)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}

	// get the item from the map cache
	item := result[prefixedKey]

	return item, nil
}

// Keys retrieves all keys matching a certain pattern, a specific key, or a list of keys.
func (rc *RedisCache) Keys(patternOrKey ...string) ([]string, error) {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	var keys []string
	var err error

	// If no argument is provided, scan all keys with the prefix
	if len(patternOrKey) == 0 {
		prefixedPattern := fmt.Sprintf("%s*", rc.Prefix)
		keys, err = rc.getKeys(prefixedPattern)
		if err != nil {
			return nil, err
		}
	} else if len(patternOrKey) == 1 {
		// If a single pattern or key is provided, use KEYS command
		prefixedPatternOrKey := rc.prefixedKey(patternOrKey[0])
		keys, err = redis.Strings(conn.Do("KEYS", prefixedPatternOrKey))
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve keys: %w", err)
		}
	} else {
		// If multiple specific keys are provided, get each key individually
		for _, key := range patternOrKey {
			prefixedKey := rc.prefixedKey(key)
			exists, err := redis.Bool(conn.Do("EXISTS", prefixedKey))
			if err != nil {
				return nil, fmt.Errorf("failed to check existence of key %s: %w", prefixedKey, err)
			}

			if exists {
				keys = append(keys, prefixedKey)
			}
		}
	}
	return keys, nil
}

// Set adds a key-value pair to the Redis cache with a prefixed key.
// It handles optional expiration time.
func (rc *RedisCache) Set(keyStr string, value interface{}, expires ...time.Duration) error {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedKey := rc.prefixedKey(keyStr)

	// create an instance of EntryCache
	entryCache := EntryCache{}
	entryCache[prefixedKey] = value

	// serialize the data to be entered to the cache
	encodedData, err := encodeValue(entryCache)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	// if expiration time is set
	if len(expires) > 0 {
		_, err = conn.Do("SETEX", prefixedKey, int(expires[0].Minutes()), encodedData)
	} else {
		_, err = conn.Do("SET", prefixedKey, encodedData)
	}

	if err != nil {
		log.Printf("Error setting cache for key %s: %v", keyStr, err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key-value pair with a prefixed key from the Redis cache.
func (rc *RedisCache) Delete(keyStr string) error {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedKey := rc.prefixedKey(keyStr)

	// delete something from the cache
	_, err := conn.Do("DEL", prefixedKey)
	if err != nil {
		log.Printf("Error deleting cache for key %s: %v", keyStr, err)
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Expire sets a timeout on a key.
func (rc *RedisCache) Expire(keyStr string, expiration time.Duration) error {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedKey := rc.prefixedKey(keyStr)

	// set expiration time settings
	_, err := conn.Do("EXPIRE", prefixedKey, int(expiration.Minutes()))
	if err != nil {
		log.Printf("Error setting expiration for key %s: %v", keyStr, err)
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// TTL retrieves the time-to-live of a key.
func (rc *RedisCache) TTL(keyStr string) (time.Duration, error) {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedKey := rc.prefixedKey(keyStr)

	// set expiration time settings
	ttl, err := redis.Int(conn.Do("TTL", prefixedKey))
	if err != nil {
		log.Printf("Error retrieving TTL for key %s: %v", keyStr, err)
		return 0, fmt.Errorf("failed to retrieve TTL: %w", err)
	}

	return time.Duration(ttl) * time.Minute, nil
}

// EmptyByMatch deletes all keys matching a specific pattern using a pipeline.
func (rc *RedisCache) EmptyByMatch(pattern string) error {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedPattern := rc.prefixedKey(pattern)

	keys, err := rc.getKeys(prefixedPattern)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")

	// delete the keys that match the pattern
	for _, k := range keys {
		_, err := conn.Do("DEL", k)
		if err != nil {
			return err
		}
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		return fmt.Errorf("failed to execute pipeline for deletion: %w", err)
	}

	return nil
}

// Empty deletes all keys with the specific prefix using a pipeline.
func (rc *RedisCache) Empty() error {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	prefixedPattern := fmt.Sprintf("%s*", rc.Prefix)

	keys, err := rc.getKeys(prefixedPattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		for _, key := range keys {
			_, err = conn.Do("DEL", key)
			if err != nil {
				return fmt.Errorf("failed to delete key %s: %w", key, err)
			}
		}
	}

	return nil

}

// Update updates an existing key-value pair in the Badger cache.
func (rc *RedisCache) Update(keyStr string, value interface{}) error {
	return nil
}

// ============================ utility functions ============
// getKeys retrieves all keys matching a specific pattern using SCAN.
func (rc *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	iter := 0
	var keys []string

	for {
		arrays, scanErr := redis.Values(conn.Do("SCAN", iter,
			"MATCH", pattern+"*"))
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", scanErr)
		}

		iter, _ = redis.Int(arrays[0], nil)
		scannedKeys, _ := redis.Strings(arrays[1], nil)
		keys = append(keys, scannedKeys...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}
