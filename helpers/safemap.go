package helpers

import (
	"sync"
)

// Thread safe map structure
type SafeMap struct {
	_map    map[string]string
	mapLock *sync.RWMutex
}

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

// Write only succeeds if the key has never been seen before
// Returns collided value in case of collision. Otherwise empty string returned
func (sm *SafeMap) WriteUnique(key, value string) string {
	sm.mapLock.Lock()
	defer sm.mapLock.Unlock()
	existingValue := sm._map[key]
	// If default value, we're good to write
	if existingValue == "" {
		sm._map[key] = value
		return ""
	} else {
		// Found something, write fails
		return sm._map[key]
	}
}
