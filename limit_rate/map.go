package limit_rate

import "sync"

type StringIntMap struct {
	coreMap map[string]int
	mutex   sync.RWMutex
}

func NewStringIntMap() *StringIntMap {
	return &StringIntMap{
		coreMap: make(map[string]int),
	}
}

func NewStringIntMapFromMap(data map[string]int) *StringIntMap {
	newMap := make(map[string]int)
	for key, value := range data {
		newMap[key] = value
	}
	return &StringIntMap{
		coreMap: newMap,
	}
}

func (mapObject *StringIntMap) Increment(key string, value int) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] += value
}

func (mapObject *StringIntMap) Set(key string, value int) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	mapObject.coreMap[key] = value
}

func (mapObject *StringIntMap) Get(key string) int {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	return mapObject.coreMap[key]
}

func (mapObject *StringIntMap) Has(key string) bool {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	_, ok := mapObject.coreMap[key]
	return ok
}

func (mapObject *StringIntMap) RLock() {
	mapObject.mutex.RLock()
}

func (mapObject *StringIntMap) RUnlock() {
	mapObject.mutex.RUnlock()
}

func (mapObject *StringIntMap) Delete(key string) {
	mapObject.mutex.Lock()
	defer mapObject.mutex.Unlock()
	delete(mapObject.coreMap, key)
}

func (mapObject *StringIntMap) Len() int {
	return len(mapObject.coreMap)
}

func (mapObject *StringIntMap) Copy() map[string]int {
	mapObject.mutex.RLock()
	defer mapObject.mutex.RUnlock()
	newMap := make(map[string]int)
	for key, value := range mapObject.coreMap {
		newMap[key] = value
	}
	return newMap
}

// =========================================
