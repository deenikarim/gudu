package cache

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"time"
)

// BadgerCache struct holds the Badger database instance and key prefix.
type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

// prefixedKey returns the key with the specified prefix.
func (b *BadgerCache) prefixedKey(key string) string {
	return fmt.Sprintf("%s:%s", b.Prefix, key)
	// return rc.Prefix + key
}

// Exists checks if a key exists in the Badger cache.
func (b *BadgerCache) Exists(keyStr string) (bool, error) {
	prefixedKey := b.prefixedKey(keyStr)

	// Start a read-only transaction
	txn := b.Conn.NewTransaction(false)
	defer txn.Discard()

	// Use txn.Get to check if the key exists
	_, err := txn.Get([]byte(prefixedKey))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return true, nil
}

// Get retrieves the value for a given prefixed key from the Badger cache
// and decodes it into an EntryCache.
func (b *BadgerCache) Get(keyStr string) (interface{}, error) {
	var result []byte

	prefixedKey := b.prefixedKey(keyStr)

	err := b.Conn.View(func(txn *badger.Txn) error {
		// get the key from the cache
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			return err
		}

		if err = item.Value(func(val []byte) error {
			result = append(result[:0], val...)
			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	decoded, err := decodeValue(result)
	if err != nil {
		return nil, err
	}

	item, exists := decoded[prefixedKey]
	if !exists {
		return nil, fmt.Errorf("key %s not found in decoded value", prefixedKey)
	}

	return item, nil
}

// Keys retrieves all keys matching a certain pattern, a specific key, or
// a list of keys.
func (b *BadgerCache) Keys(patternOrKey ...string) ([]string, error) {
	var keys []string

	if err := b.Conn.View(func(txn *badger.Txn) error {
		// If no argument is provided, scan all keys with the prefix
		if len(patternOrKey) == 0 {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			prefixedPattern := b.prefixedKey("")

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				if bytes.HasPrefix(item.Key(), []byte(prefixedPattern)) {
					keys = append(keys, string(item.Key()))
				}
			}
		} else if len(patternOrKey) == 1 {
			// If a single pattern or key is provided, use it as is
			prefixedPatternOrKey := b.prefixedKey(patternOrKey[0])
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Seek([]byte(prefixedPatternOrKey)); it.ValidForPrefix([]byte(prefixedPatternOrKey)); it.Next() {
				item := it.Item()
				keys = append(keys, string(item.Key()))
			}
		} else {
			// If multiple specific keys are provided, get each key individually
			for _, k := range patternOrKey {
				prefixedKey := b.prefixedKey(k)
				_, err := txn.Get([]byte(prefixedKey))
				if err == nil {
					keys = append(keys, prefixedKey)
				} else if !errors.Is(err, badger.ErrKeyNotFound) {
					return err
				}
			}
		}
		return nil

	}); err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	return keys, nil
}

// Set adds a key-value pair to the Badger cache with a prefixed key.
// It handles optional expiration time.
func (b *BadgerCache) Set(keyStr string, value interface{}, expires ...time.Duration) error {
	prefixedKey := b.prefixedKey(keyStr)

	entry := EntryCache{}
	entry[prefixedKey] = value

	encoded, err := encodeValue(entry)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	return b.Conn.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(prefixedKey), encoded)
		if len(expires) > 0 {
			e.WithTTL(expires[0])
		}
		return txn.SetEntry(e)
	})
}

// Update updates an existing key-value pair in the Badger cache.
func (b *BadgerCache) Update(keyStr string, value interface{}) error {
	prefixedKey := b.prefixedKey(keyStr)

	// Check if the key exists
	exist, err := b.Exists(prefixedKey)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("key %s does not exist", prefixedKey)
	}

	// Encode the value
	entry := EntryCache{}
	entry[prefixedKey] = value

	encoded, err := encodeValue(entry)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	// Update the value in the cache
	return b.Conn.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(prefixedKey), encoded)
		return txn.SetEntry(e)
	})
}

// Delete removes a key-value pair with a prefixed key from the Badger cache.
func (b *BadgerCache) Delete(keyStr string) error {
	prefixedKey := b.prefixedKey(keyStr)
	return b.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(prefixedKey))
		if err != nil {
			return fmt.Errorf("failed to delete key %s: %w", prefixedKey, err)
		}
		return nil
	})
}

// Expire sets a timeout on a key.
func (b *BadgerCache) Expire(keyStr string, expiration time.Duration) error {
	return b.Conn.Update(func(txn *badger.Txn) error {
		prefixedKey := b.prefixedKey(keyStr)
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			e := badger.NewEntry([]byte(prefixedKey), val).WithTTL(expiration)
			return txn.SetEntry(e)
		})
	})
}

// TTL retrieves the time-to-live of a key.
func (b *BadgerCache) TTL(keyStr string) (time.Duration, error) {
	prefixedKey := b.prefixedKey(keyStr)
	var ttl time.Duration

	if err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			return err
		}
		// set the time to live for the key
		ttl = time.Duration(item.ExpiresAt())
		return nil
	}); err != nil {
		return 0, fmt.Errorf("failed to retrieve TTL: %w", err)
	}

	return ttl, nil
}

// EmptyByMatch deletes all keys matching a specific pattern
func (b *BadgerCache) EmptyByMatch(pattern string) error {
	prefixedPattern := b.prefixedKey(pattern)
	for {
		err := b.Conn.Update(func(txn *badger.Txn) error {
			deleted, err := b.deleteKeysMatchingPattern(txn, prefixedPattern)
			if err != nil {
				return err
			}

			if deleted == 0 {
				return nil
			}
			return nil
		})
		// if the error occurred during the transaction
		if err != nil {
			return err
		}

		if err == nil {
			break
		}
	}
	return nil
}

// Empty deletes all keys with the specific prefix using a pipeline.
func (b *BadgerCache) Empty() error {
	prefixedPattern := fmt.Sprintf("%s:", b.Prefix)
	for {
		err := b.Conn.Update(func(txn *badger.Txn) error {
			deleted, err := b.deleteKeysMatchingPattern(txn, prefixedPattern)
			if err != nil {
				return err
			}

			if deleted == 0 {
				return nil
			}
			return nil
		})
		// if the error occurred during the transaction
		if err != nil {
			return err
		}

		if err == nil {
			break
		}
	}
	return nil
}

// ============================ utility functions ============
// deleteKeysMatchingPattern deletes keys in batches of 10,000 that match
// the specified pattern.
func (b *BadgerCache) deleteKeysMatchingPattern(txn *badger.Txn, pattern string) (int, error) {
	opts := badger.DefaultIteratorOptions
	opts.AllVersions = false
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	deleted := 0

	for it.Seek([]byte(pattern)); it.ValidForPrefix([]byte(pattern)); it.Next() {
		item := it.Item()
		if err := txn.Delete(item.Key()); err != nil {
			return deleted, fmt.Errorf("failed to delete key %s: %w", item.Key(), err)
		}
		deleted++

		if deleted >= 10000 {
			break
		}

	}
	return deleted, nil
}
