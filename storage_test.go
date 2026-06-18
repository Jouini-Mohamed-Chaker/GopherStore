package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"testing"
)

type Data struct {
	key   string
	value []byte
}

func getTestsAndFilledStore() ([]Data, *Store) {
	var tests = []Data{
		{"key", []byte("value")},
		{"a", []byte("hello")},
		{"abc", []byte("b12a")},
		{"123", []byte("test_@123")},
	}

	var testStore = NewStore()
	for _, tt := range tests {
		testStore.Set(tt.key, tt.value)
	}
	return tests, &testStore
}

func TestStore_Set(t *testing.T) {
	var tests, testStore = getTestsAndFilledStore()

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%d", tt.key, tt.value)
		t.Run(testname, func(t *testing.T) {
			testStore.Set(tt.key, tt.value)
			val, ok := testStore.Get(tt.key)
			if !ok || !bytes.Equal(val, tt.value) {
				t.Errorf("got %s, want %s", val, tt.value)
			}
		})
	}
}

func TestStore_Get(t *testing.T) {
	var tests, testStore = getTestsAndFilledStore()

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%s", tt.key, tt.value)
		t.Run(testname, func(t *testing.T) {
			val, ok := testStore.Get(tt.key)
			if !ok || !bytes.Equal(val, tt.value) {
				t.Errorf("got %s, want %s", val, tt.value)
			}
		})
	}
}

func TestStore_Delete(t *testing.T) {
	var tests, testStore = getTestsAndFilledStore()

	for _, tt := range tests {
		testname := fmt.Sprintf("%s,%s", tt.key, tt.value)
		t.Run(testname, func(t *testing.T) {
			ok := testStore.Delete(tt.key)
			_, valueExists := testStore.Get(tt.key)
			if !ok {
				t.Errorf("Delete(%q) returned false, want true (key didn't exist to begin with)", tt.key)
			}

			if valueExists {
				t.Errorf("Get(%q) returned valueExists=true after deletion, want false", tt.key)
			}
		})
	}
}

func BenchmarkStore_Set(b *testing.B) {
	store := NewStore()

	// Precompute keys and values — not part of what you're benchmarking
	keys := make([]string, b.N)
	vals := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(i)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(i))
		vals[i] = buf
	}

	b.ResetTimer() // <-- only measure from here
	for i := 0; i < b.N; i++ {
		store.Set(keys[i], vals[i])
	}
}

func BenchmarkStore_Get(b *testing.B) {
	store := NewStore()
	for i := 0; i < 1000; i++ {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(i))
		store.Set(strconv.Itoa(i), buf)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(strconv.Itoa(i % 1000))
	}
}

func BenchmarkStore_Delete(b *testing.B) {
	store := NewStore()
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(i)
	}
	dummyData := []byte{1, 2, 3, 4}
	for i := 0; i < b.N; i++ {
		store.Set(keys[i], dummyData)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Delete(keys[i])
	}
}
