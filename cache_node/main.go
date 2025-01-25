package main

import (
    "encoding/json"
    "net/http"
    "sync"
)

type Cache struct {
    mu    sync.RWMutex
    store map[string]string
}

func NewCache() *Cache {
    return &Cache{
        store: make(map[string]string),
    }
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.store[key] = value
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    value, ok := c.store[key]
    return value, ok
}

func main() {
    cache := NewCache()
    http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
        var kv map[string]string
        json.NewDecoder(r.Body).Decode(&kv)
        for k, v := range kv {
            cache.Set(k, v)
        }
        w.WriteHeader(http.StatusOK)
    })

    http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
        key := r.URL.Query().Get("key")
        value, ok := cache.Get(key)
        if !ok {
            http.NotFound(w, r)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{key: value})
    })

    http.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    http.ListenAndServe(":8080", nil)
}