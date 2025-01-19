package main

import (
    "fmt"
    "net/http"
    "sync"

    "github.com/stathat/consistent"
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
    fmt.Println("Cache service started")
    cache := NewCache()
    hash := consistent.New()
    hash.Add("cache1")
    hash.Add("cache2")
    hash.Add("cache3")

    http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
        key := r.URL.Query().Get("key")
        value := r.URL.Query().Get("value")
        node, err := hash.Get(key)
        if err != nil {
            http.Error(w, "Error finding node", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "Set key=%s value=%s on node=%s\n", key, value, node)
        cache.Set(key, value)
    })

    http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
        key := r.URL.Query().Get("key")
        node, err := hash.Get(key)
        if err != nil {
            http.Error(w, "Error finding node", http.StatusInternalServerError)
            return
        }
        value, ok := cache.Get(key)
        if !ok {
            http.Error(w, "Key not found", http.StatusNotFound)
            return
        }
        fmt.Fprintf(w, "Get key=%s value=%s from node=%s\n", key, value, node)
    })

    http.ListenAndServe(":8080", nil)
}