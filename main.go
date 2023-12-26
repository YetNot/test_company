package main

import (
  "fmt"
  "sync"
  "time"
)

type KeyValueStorage struct {
  data     map[string]interface{}
  expiry   map[string]time.Time
  mutex    sync.RWMutex
  cleanup  chan string
  shutdown chan struct{}
}

func NewKeyValueStorage() *KeyValueStorage {
  storage := &KeyValueStorage{
    data:     make(map[string]interface{}),
    expiry:   make(map[string]time.Time),
    cleanup:  make(chan string),
    shutdown: make(chan struct{}),
  }

  go storage.startCleanupRoutine()

  return storage
}

func (s *KeyValueStorage) Set(key string, value interface{}, ttl time.Duration) {
  s.mutex.Lock()
  defer s.mutex.Unlock()

  s.data[key] = value
  if ttl > 0 {
    s.expiry[key] = time.Now().Add(ttl)
  }
}

func (s *KeyValueStorage) Get(key string) interface{} {
  s.mutex.RLock()
  defer s.mutex.RUnlock()

  value, ok := s.data[key]
  if !ok {
    return nil
  }

  expiry, ok := s.expiry[key]
  if ok && expiry.Before(time.Now()) {
    s.cleanup <- key
    return nil
  }

  return value
}

func (s *KeyValueStorage) Delete(key string) {
  s.mutex.Lock()
  defer s.mutex.Unlock()

  delete(s.data, key)
  delete(s.expiry, key)
}

func (s *KeyValueStorage) startCleanupRoutine() {
  for {
    select {
    case key := <-s.cleanup:
      s.Delete(key)
    case <-s.shutdown:
      return
    }
  }
}

func (s *KeyValueStorage) Shutdown() {
  s.shutdown <- struct{}{}
}

func main() {
  storage := NewKeyValueStorage()

  storage.Set("key1", "value1", time.Second*5)
  storage.Set("key2", "value2", time.Second*10)

  fmt.Println(storage.Get("key1")) // Output: value1
  fmt.Println(storage.Get("key2")) // Output: value2

  time.Sleep(time.Second * 6)

  fmt.Println(storage.Get("key1")) // Output: nil

  storage.Shutdown()
}