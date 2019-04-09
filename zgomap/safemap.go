package zgomap

import "sync"

/*
@Time : 2019-03-15 11:29
@Author : rubinus.chu
@File : safeMap
@project: zgo
*/

var Map Maper

type Maper interface {
	New() *safeMap
	Get(k interface{}) interface{}
	Set(k interface{}, v interface{}) bool
	IsExists(k interface{}) bool
	IsEmpty() bool
	Delete(k interface{})
	Size() int
	Range() chan *Sma
}

// safeMap is concurrent security map
type safeMap struct {
	lock *sync.RWMutex
	sm   map[interface{}]interface{}
}

// NewsafeMap get a new concurrent security map
func GetMap() *safeMap {
	return &safeMap{
		lock: new(sync.RWMutex),
		sm:   make(map[interface{}]interface{}),
	}
}

func (m *safeMap) New() *safeMap {
	return GetMap()
}

// Get used to get a value by key
func (m *safeMap) Get(k interface{}) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.sm[k]; ok {
		return val
	}
	return nil
}

// Set used to set value with key
func (m *safeMap) Set(k interface{}, v interface{}) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.sm[k]; !ok {
		m.sm[k] = v
	} else if val != v {
		m.sm[k] = v
	} else {
		return false
	}
	return true
}

// IsExists determine whether k exists
func (m *safeMap) IsExists(k interface{}) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.sm[k]; !ok {
		return false
	}
	return true
}

// Delete used to delete a key
func (m *safeMap) Delete(k interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.sm, k)
}

// Len长度
func (m *safeMap) Size() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.sm)
}

// IsEmpty
func (m *safeMap) IsEmpty() bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.sm) > 0
}

type Sma struct {
	Key interface{}
	Val interface{}
}

func (m *safeMap) Range() chan *Sma {
	m.lock.RLock()
	defer m.lock.RUnlock()
	out := make(chan *Sma)
	go func() {
		for k, v := range m.sm {
			c := &Sma{
				Key: k,
				Val: v,
			}
			out <- c
		}
		close(out)
	}()
	return out
}
