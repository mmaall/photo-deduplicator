package helper

import (
	"sync"
)

// Threadsafe Map
type SafeMap struct {
	_map    map[string]string
	mapLock *sync.RWMutex
}

// Create a new safe map
func NewSafeMap() *SafeMap {
	return &SafeMap{
		_map:    make(map[string]string),
		mapLock: &sync.RWMutex{},
	}
}

func (sm *SafeMap) Read(key string) string {
	sm.mapLock.RLock()
	defer sm.mapLock.RUnlock()
	return sm._map[key]
}

func (sm *SafeMap) Write(key, value string) {
	sm.mapLock.Lock()
	defer sm.mapLock.Unlock()
	sm._map[key] = value
}
