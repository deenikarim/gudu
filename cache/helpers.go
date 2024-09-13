package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// encodeValue encodes a value into gob format.
func encodeValue(item EntryCache) ([]byte, error) {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(item)
	if err != nil {
		return nil, fmt.Errorf("failed to encode value: %w", err)
	}
	return buff.Bytes(), nil
}

// decodeValue decodes gob data into an EntryCache.
func decodeValue(data []byte) (EntryCache, error) {
	var item EntryCache
	buff := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buff)
	err := decoder.Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}
	return item, nil
}

// contains Helper function to check if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
