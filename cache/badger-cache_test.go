package cache

import (
	"testing"
	"time"
)

// TestBadgerCache_Set Validates that the value stored in
// the cache matches the expected value.
func TestBadgerCache_Set(t *testing.T) {
	data := "value"

	err := testBadgerCache.Set("foo", data, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// Verify that the key exists and has the correct value
	result, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrievedData, ok := result.(string)
	if !ok {
		t.Errorf("Expected EntryCache type, got %T", result)
	}

	if retrievedData != "value" {
		t.Errorf("Expected %v, got %v", data, retrievedData)
	}

	err = testBadgerCache.Delete("foo")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Get  Confirms the retrieved value matches the expected value.
func TestBadgerCache_Get(t *testing.T) {
	data := "school-girl"
	err := testBadgerCache.Set("myKey", data, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result, err := testBadgerCache.Get("myKey")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrievedData, ok := result.(string)
	if !ok {
		t.Errorf("Expected string type, got %T", result)
	}
	if retrievedData != data {
		t.Errorf("Expected %v, got %v", data, retrievedData)
	}
	err = testBadgerCache.Delete("myKey")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Delete2 Checks that the key is deleted and cannot be retrieved.
func TestBadgerCache_Delete2(t *testing.T) {
	data := 6754

	err := testBadgerCache.Set("myKey", data, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = testBadgerCache.Delete("myKey")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = testBadgerCache.Get("myKey")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// TestEmptyByMatch tests the EmptyByMatch method.
func TestEmptyByMatchBadger(t *testing.T) {
	data1 := 4
	data2 := 65

	err := testBadgerCache.Set("key1", data1, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = testBadgerCache.Set("key2", data2, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = testBadgerCache.EmptyByMatch("key*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	keys, err := testBadgerCache.Keys("key*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %v", keys)
	}

	err = testBadgerCache.Delete("key1")
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Delete("key2")
	if err != nil {
		t.Error(err)
	}
}

// TestEmpty tests the Empty method.
func TestEmptyBadger(t *testing.T) {
	data1 := 23
	data2 := 12

	err := testBadgerCache.Set("keyed1", data1, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = testBadgerCache.Set("keyed2", data2, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = testBadgerCache.Empty()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	keys, err := testBadgerCache.Keys()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %v", keys)
	}
}

// TestExists tests Verifies the existence and non-existence of keys.
func TestBadgerCache_Exists(t *testing.T) {
	// Check existence of a key after setting it
	data := "fanout"
	err := testBadgerCache.Set("myKey", data, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	exists, err := testBadgerCache.Exists("myKey")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !exists {
		t.Errorf("Expected key to exist")
	}

	// Check non-existence of a key after deleting it
	err = testBadgerCache.Delete("myKey")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	exists, err = testBadgerCache.Exists("myKey")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if exists {
		t.Errorf("Expected key to not exist")
	}
	// Check non-existence of a key that was never set
	exists, err = testBadgerCache.Exists("nonExistentKey")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if exists {
		t.Errorf("Expected key to not exist")
	}
}

// TestBadgerCache_Keys Ensures the correct keys are returned for various
// patterns and specific key requests.
func TestBadgerCache_Keys(t *testing.T) {
	data1 := 67
	data2 := "berries"

	err := testBadgerCache.Set("key1", data1, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	err = testBadgerCache.Set("key2", data2, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test retrieving all keys
	keys, err := testBadgerCache.Keys()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedKeys := []string{"test-gudu:key1", "test-gudu:key2"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}
	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key %v in result, but it was not found", key)
		}
	}

	// Test retrieving keys matching a pattern
	keys, err = testBadgerCache.Keys("key*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}
	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key %v in result, but it was not found", key)
		}
	}

	// Test retrieving a specific key
	keys, err = testBadgerCache.Keys("key1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(keys) != 1 || keys[0] != "test-gudu:key1" {
		t.Errorf("Expected key 'test-gudu:key1', got %v", keys)
	}

	// Test retrieving multiple specific keys
	keys, err = testBadgerCache.Keys("key1", "key2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}
	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key %v in result, but it was not found", key)
		}
	}

	// Clean up
	err = testBadgerCache.Delete("key1")
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Delete("key2")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Update Ensures the updated value is correctly stored and retrieved.
func TestBadgerCache_Update(t *testing.T) {
	data := 30

	err := testBadgerCache.Set("foo", data, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	updatedData := "john"
	err = testBadgerCache.Update("foo", updatedData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify that the key exists and has the updated value
	result, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	retrievedData, ok := result.(string)
	if !ok {
		t.Errorf("Expected string type, got %T", result)
	}

	if retrievedData != updatedData {
		t.Errorf("Expected %v, got %v", updatedData, retrievedData)
	}

}
