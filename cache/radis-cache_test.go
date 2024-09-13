package cache

import (
	"testing"
	"time"
)

// TestRedisCache_Set tests the Set Entry method.
func TestRedisCache_Set(t *testing.T) {
	data := []string{"beta", "roads"}

	err := testRedisCache.Set("foo", data, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	exists, err := testRedisCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	// Check directly in miniredis
	if !exists {
		t.Errorf("Expected key 'test-gudu:myKey' to exist but did not")
	}

	err = testRedisCache.Delete("foo")
	if err != nil {
		t.Errorf("Error deleting cache: %v", err)
	}
}

func TestRedisCache_Exists(t *testing.T) {
	err := testRedisCache.Delete("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache when it shouldn't")
	}

	data := []string{"beta", "roads"}

	err = testRedisCache.Set("foo", data, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	inCache, err = testRedisCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found in cache when it should be there")
	}

	err = testRedisCache.Delete("foo")
	if err != nil {
		t.Errorf("Error deleting cache: %v", err)
	}

}

// TestRedisCache_Keys tests the Keys method.
func TestRedisCache_Keys(t *testing.T) {

	data1 := []string{"beta", "roads"}

	data2 := []string{"beta", "roads"}

	data3 := 3

	err := testRedisCache.Set("ping1", data1, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("ping2", data2, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("ping3", data3, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// todo Test retrieving all keys with the prefix
	keys, err := testRedisCache.Keys()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedKeys := []string{"test-gudu:ping1", "test-gudu:ping2", "test-gudu:ping3"}

	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}

	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key %v in result, but it was not found", key)
		}
	}

	// todo: 	// Test retrieving keys matching a pattern
	keys, err = testRedisCache.Keys("ping*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedKeys = []string{"test-gudu:ping1", "test-gudu:ping2", "test-gudu:ping3"}

	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}

	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key %v in result, but it was not found", key)
		}
	}

	// todo	// Test retrieving a specific key
	keys, err = testRedisCache.Keys("ping1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(keys) != 1 || keys[0] != "test-gudu:ping1" {
		t.Errorf("Expected key 'test-gudu:ping1', got %v", keys)
	}

	// todo	// Test retrieving multiple specific keys
	keys, err = testRedisCache.Keys("ping1", "ping2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedKeys = []string{"test-gudu:ping1", "test-gudu:ping2"}

	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}

	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key %v in result, but it was not found", key)
		}
	}

	// Clean up
	err = testRedisCache.Delete("ping1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = testRedisCache.Delete("ping2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = testRedisCache.Delete("ping3")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

}

func TestRedisCache_Get(t *testing.T) {
	err := testRedisCache.Set("gettest", "bar")
	if err != nil {
		t.Error(err)
	}

	x, err := testRedisCache.Get("gettest")
	if err != nil {
		t.Error(err)
	}

	if x != "bar" {
		t.Error("could not get the correct value from the cache")
	}

}

func TestRedisCache_Delete(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Delete("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in the cache when it should not be there")
	}
}

func TestRedisCache_EmptyByMatch(t *testing.T) {
	err := testRedisCache.Set("alpha", "foo")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("alpha2", "come")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("beta", "back")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.EmptyByMatch("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("found alpha but it should not be there")
	}

	inCache, err = testRedisCache.Exists("alpha2")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("found alpha2 but it should not be there")
	}

	inCache, err = testRedisCache.Exists("beta")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("beta not found in cache but it should be there")
	}

}

// TestRedisCache_Expire tests the Expire method.
func TestRedisCache_Expire(t *testing.T) {
	data := 12
	err := testRedisCache.Set("ex", data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = testRedisCache.Expire("ex", 1*time.Second)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	time.Sleep(2 * time.Second)

	exists, err := testRedisCache.Exists("ex")
	if err != nil {
		t.Error(err)
	}

	if exists {
		t.Errorf("Expected false because key should expired by then, got true")
	}

}

// TestTTL tests the TTL method.
func TestRedisCache_TTL(t *testing.T) {
	data := 12
	err := testRedisCache.Set("ex", data, 30*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	ttl, err := testRedisCache.TTL("ex")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if ttl <= 0 {
		t.Errorf("Expected ttl > 0, got %v", ttl)
	}

}

func TestDecodeEncode(t *testing.T) {
	entry := EntryCache{}
	entry["foo"] = "bar"

	bytes, err := encodeValue(entry)
	if err != nil {
		t.Error(err)
	}

	_, err = decodeValue(bytes)
	if err != nil {
		t.Error(err)
	}
}

func TestRedisCache_Empty(t *testing.T) {
	err := testRedisCache.Set("yell", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("my", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("ky", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("rome", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Empty()
	if err != nil {
		t.Error(err)
	}

	keys, err := testRedisCache.Keys("*")
	if err != nil {
		t.Error(err)
	}

	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %v", keys)
	}
}
